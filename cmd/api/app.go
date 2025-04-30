package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/database"
	"github.com/Notifuse/notifuse/internal/domain"
	httpHandler "github.com/Notifuse/notifuse/internal/http"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/internal/repository"
	"github.com/Notifuse/notifuse/internal/service"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mailer"
)

// App encapsulates the application dependencies and configuration
type App struct {
	config *config.Config
	logger logger.Logger
	db     *sql.DB
	mailer mailer.Mailer

	// Repositories
	userRepo        domain.UserRepository
	workspaceRepo   domain.WorkspaceRepository
	authRepo        domain.AuthRepository
	contactRepo     domain.ContactRepository
	listRepo        domain.ListRepository
	contactListRepo domain.ContactListRepository
	templateRepo    domain.TemplateRepository
	broadcastRepo   domain.BroadcastRepository
	taskRepo        domain.TaskRepository
	// Services
	authService        *service.AuthService
	userService        *service.UserService
	workspaceService   *service.WorkspaceService
	contactService     *service.ContactService
	listService        *service.ListService
	contactListService *service.ContactListService
	templateService    *service.TemplateService
	emailService       *service.EmailService
	broadcastService   *service.BroadcastService
	taskService        *service.TaskService
	eventBus           domain.EventBus

	// HTTP handlers
	mux    *http.ServeMux
	server *http.Server

	// Server synchronization
	serverMu      sync.RWMutex
	serverStarted chan struct{}
}

// AppOption defines a functional option for configuring the App
type AppOption func(*App)

// WithMockDB configures the app to use a mock database
func WithMockDB(db *sql.DB) AppOption {
	return func(a *App) {
		a.db = db
	}
}

// WithMockMailer configures the app to use a mock mailer
func WithMockMailer(m mailer.Mailer) AppOption {
	return func(a *App) {
		a.mailer = m
	}
}

// WithLogger sets a custom logger
func WithLogger(logger logger.Logger) AppOption {
	return func(a *App) {
		a.logger = logger
	}
}

// NewRealApp creates a new application instance
func NewRealApp(cfg *config.Config, opts ...AppOption) AppInterface {
	app := &App{
		config:        cfg,
		logger:        logger.NewLogger(), // Default logger
		mux:           http.NewServeMux(),
		serverStarted: make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		opt(app)
	}

	return app
}

// InitDB initializes the database connection
func (a *App) InitDB() error {
	// Ensure system database exists
	if err := database.EnsureSystemDatabaseExists(&a.config.Database); err != nil {
		return fmt.Errorf("failed to ensure system database exists: %w", err)
	}
	a.logger.Info("System database check completed")

	// Connect to system database
	db, err := sql.Open("postgres", database.GetSystemDSN(&a.config.Database))
	if err != nil {
		return fmt.Errorf("failed to connect to system database: %w", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping system database: %w", err)
	}

	// Initialize database schema if needed
	if err := database.InitializeDatabase(db, a.config.RootEmail); err != nil {
		db.Close()
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	a.db = db
	return nil
}

// InitMailer initializes the mailer service
func (a *App) InitMailer() error {
	// Skip if mailer already set (e.g., by mock)
	if a.mailer != nil {
		return nil
	}

	if a.config.IsDevelopment() {
		// Use console mailer in development
		a.mailer = mailer.NewConsoleMailer()
		a.logger.Info("Using console mailer for development")
	} else {
		// Use SMTP mailer in production
		a.mailer = mailer.NewSMTPMailer(&mailer.Config{
			SMTPHost:     os.Getenv("SMTP_HOST"),
			SMTPPort:     587,
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
			FromEmail:    os.Getenv("FROM_EMAIL"),
			FromName:     "Notifuse",
			BaseURL:      os.Getenv("BASE_URL"),
		})
		a.logger.Info("Using SMTP mailer for production")
	}

	return nil
}

// InitRepositories initializes all repositories
func (a *App) InitRepositories() error {
	if a.db == nil {
		return fmt.Errorf("database must be initialized before repositories")
	}

	a.userRepo = repository.NewUserRepository(a.db)
	a.workspaceRepo = repository.NewWorkspaceRepository(a.db, &a.config.Database, a.config.Security.SecretKey)
	a.authRepo = repository.NewSQLAuthRepository(a.db, a.logger)
	a.contactRepo = repository.NewContactRepository(a.workspaceRepo)
	a.listRepo = repository.NewListRepository(a.workspaceRepo)
	a.contactListRepo = repository.NewContactListRepository(a.workspaceRepo)
	a.templateRepo = repository.NewTemplateRepository(a.workspaceRepo)
	a.broadcastRepo = repository.NewBroadcastRepository(a.workspaceRepo, a.logger)
	a.taskRepo = repository.NewTaskRepository(a.db)

	return nil
}

// InitServices initializes all application services
func (a *App) InitServices() error {
	// Initialize event bus first
	a.eventBus = domain.NewInMemoryEventBus()

	// Initialize auth service
	authServiceConfig := service.AuthServiceConfig{
		Repository:          a.authRepo,
		WorkspaceRepository: a.workspaceRepo,
		PrivateKey:          a.config.Security.PasetoPrivateKeyBytes,
		PublicKey:           a.config.Security.PasetoPublicKeyBytes,
		Logger:              a.logger,
	}

	var err error
	a.authService, err = service.NewAuthService(authServiceConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize auth service: %w", err)
	}

	// Initialize user service
	userServiceConfig := service.UserServiceConfig{
		Repository:    a.userRepo,
		AuthService:   a.authService,
		EmailSender:   a.mailer,
		SessionExpiry: 30 * 24 * time.Hour, // 30 days
		Logger:        a.logger,
		IsDevelopment: a.config.IsDevelopment(),
	}

	a.userService, err = service.NewUserService(userServiceConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize user service: %w", err)
	}

	// Initialize template service
	a.templateService = service.NewTemplateService(
		a.templateRepo,
		a.authService,
		a.logger,
	)

	// Initialize contact service
	a.contactService = service.NewContactService(
		a.contactRepo,
		a.workspaceRepo,
		a.authService,
		a.logger,
	)

	// Initialize list service
	a.listService = service.NewListService(
		a.listRepo,
		a.authService,
		a.logger,
	)

	// Initialize contact list service
	a.contactListService = service.NewContactListService(
		a.contactListRepo,
		a.authService,
		a.contactRepo,
		a.listRepo,
		a.logger,
	)

	// Initialize email service
	a.emailService = service.NewEmailService(
		a.logger,
		a.authService,
		a.config.Security.SecretKey,
		a.workspaceRepo,
		a.templateRepo,
		a.templateService,
	)

	// Initialize workspace service
	a.workspaceService = service.NewWorkspaceService(
		a.workspaceRepo,
		a.logger,
		a.userService,
		a.authService,
		a.mailer,
		a.config,
		a.contactService,
		a.listService,
		a.contactListService,
		a.templateService,
		a.config.Security.SecretKey,
	)

	// Initialize task service
	a.taskService = service.NewTaskService(a.taskRepo, a.logger, a.authService, a.config.APIEndpoint)

	// Initialize broadcast service
	a.broadcastService = service.NewBroadcastService(
		a.logger,
		a.broadcastRepo,
		a.emailService,
		a.contactRepo,
		a.templateService,
		nil, // No taskService yet
		a.authService,
		a.eventBus, // Pass the event bus
	)

	// Register the broadcast processor with the task service
	a.taskService.RegisterDefaultProcessors(a.broadcastService)

	// Register task service to listen for broadcast events
	a.taskService.SubscribeToBroadcastEvents(a.eventBus)

	// Set the task service on the broadcast service
	a.broadcastService.SetTaskService(a.taskService)

	return nil
}

// InitHandlers initializes all HTTP handlers and routes
func (a *App) InitHandlers() error {

	// Initialize handlers
	userHandler := httpHandler.NewUserHandler(
		a.userService,
		a.workspaceService,
		a.config,
		a.config.Security.PasetoPublicKey,
		a.logger)
	rootHandler := httpHandler.NewRootHandler()
	workspaceHandler := httpHandler.NewWorkspaceHandler(
		a.workspaceService,
		a.config.Security.PasetoPublicKey,
		a.logger,
		a.config.Security.SecretKey,
	)
	faviconHandler := httpHandler.NewFaviconHandler()
	contactHandler := httpHandler.NewContactHandler(a.contactService, a.config.Security.PasetoPublicKey, a.logger)
	listHandler := httpHandler.NewListHandler(a.listService, a.config.Security.PasetoPublicKey, a.logger)
	contactListHandler := httpHandler.NewContactListHandler(a.contactListService, a.config.Security.PasetoPublicKey, a.logger)
	templateHandler := httpHandler.NewTemplateHandler(a.templateService, a.config.Security.PasetoPublicKey, a.logger)
	emailHandler := httpHandler.NewEmailHandler(a.emailService, a.config.Security.PasetoPublicKey, a.logger, a.config.Security.SecretKey)
	broadcastHandler := httpHandler.NewBroadcastHandler(a.broadcastService, a.templateService, a.config.Security.PasetoPublicKey, a.logger)
	taskHandler := httpHandler.NewTaskHandler(
		a.taskService,
		a.config.Security.PasetoPublicKey,
		a.logger,
		a.config.Security.SecretKey,
	)

	// Register routes
	userHandler.RegisterRoutes(a.mux)
	workspaceHandler.RegisterRoutes(a.mux)
	rootHandler.RegisterRoutes(a.mux)
	contactHandler.RegisterRoutes(a.mux)
	listHandler.RegisterRoutes(a.mux)
	contactListHandler.RegisterRoutes(a.mux)
	templateHandler.RegisterRoutes(a.mux)
	emailHandler.RegisterRoutes(a.mux)
	broadcastHandler.RegisterRoutes(a.mux)
	taskHandler.RegisterRoutes(a.mux)
	a.mux.HandleFunc("/api/detect-favicon", faviconHandler.DetectFavicon)

	return nil
}

// Start starts the HTTP server
func (a *App) Start() error {
	// Create server with wrapped handler for CORS
	handler := middleware.CORSMiddleware(a.mux)

	addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	a.logger.WithField("address", addr).Info("Server starting")

	// Create a fresh notification channel and update the server
	a.serverMu.Lock()
	// Close the existing channel if it exists
	if a.serverStarted != nil {
		close(a.serverStarted)
	}
	a.serverStarted = make(chan struct{})

	// Create the server
	a.server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Get a reference to the channel before unlocking
	serverStarted := a.serverStarted
	a.serverMu.Unlock()

	// Signal that the server has been created and is about to start
	close(serverStarted)

	// Start the server based on SSL configuration
	if a.config.Server.SSL.Enabled {
		a.logger.WithField("cert_file", a.config.Server.SSL.CertFile).Info("SSL enabled")
		return a.server.ListenAndServeTLS(a.config.Server.SSL.CertFile, a.config.Server.SSL.KeyFile)
	}

	return a.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (a *App) Shutdown(ctx context.Context) error {
	a.serverMu.RLock()
	server := a.server
	a.serverMu.RUnlock()

	if server != nil {
		return server.Shutdown(ctx)
	}

	// Close database connection if it exists
	if a.db != nil {
		a.db.Close()
	}

	return nil
}

// IsServerCreated safely checks if the server has been created
func (a *App) IsServerCreated() bool {
	a.serverMu.RLock()
	defer a.serverMu.RUnlock()
	return a.server != nil
}

// WaitForServerStart waits for the server to be created and initialized
// Returns true if the server started successfully, false if context expired
func (a *App) WaitForServerStart(ctx context.Context) bool {
	// Get the current channel under lock
	a.serverMu.RLock()
	started := a.serverStarted
	a.serverMu.RUnlock()

	// If the channel is nil, that's a logic error - just wait on the context
	if started == nil {
		a.logger.Error("serverStarted channel is nil - server initialization error")
		select {
		case <-ctx.Done():
			return false
		}
	}

	// Wait for signal or timeout
	select {
	case <-started:
		return a.IsServerCreated() // Double-check server was created
	case <-ctx.Done():
		return false
	}
}

// Initialize sets up all components of the application
func (a *App) Initialize() error {
	if err := a.InitDB(); err != nil {
		return err
	}

	if err := a.InitMailer(); err != nil {
		return err
	}

	if err := a.InitRepositories(); err != nil {
		return err
	}

	if err := a.InitServices(); err != nil {
		return err
	}

	if err := a.InitHandlers(); err != nil {
		return err
	}

	return nil
}

// GetConfig returns the app's configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}

// GetLogger returns the app's logger
func (a *App) GetLogger() logger.Logger {
	return a.logger
}

// GetMux returns the app's HTTP multiplexer
func (a *App) GetMux() *http.ServeMux {
	return a.mux
}

// GetDB returns the app's database connection
func (a *App) GetDB() *sql.DB {
	return a.db
}

// GetMailer returns the app's mailer
func (a *App) GetMailer() mailer.Mailer {
	return a.mailer
}

// SetHandler allows setting a custom HTTP handler
func (a *App) SetHandler(handler http.Handler) {
	a.mux = handler.(*http.ServeMux)
}

// Ensure App implements AppInterface
var _ AppInterface = (*App)(nil)
