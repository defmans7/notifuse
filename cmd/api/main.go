package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

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

type emailSender struct{}

func (s *emailSender) SendMagicCode(email, code string) error {
	// TODO: Implement email sending using SMTP
	log.Printf("Sending magic code to %s: %s", email, code)
	return nil
}

// osExit is a variable to allow mocking os.Exit in tests
var osExit = os.Exit

// authServiceMiddlewareAdapter adapts AuthService to implement middleware.AuthServiceInterface
type authServiceMiddlewareAdapter struct {
	authService *service.AuthService
}

func (a *authServiceMiddlewareAdapter) VerifyUserSession(ctx context.Context, userID, sessionID string) (*domain.User, error) {
	return a.authService.VerifyUserSession(ctx, userID, sessionID)
}

func (a *authServiceMiddlewareAdapter) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return a.authService.GetUserByID(ctx, userID)
}

// userServiceAdapter adapts AuthService to implement httpHandler.UserServiceInterface
type userServiceAdapter struct {
	authService *service.AuthService
	userService *service.UserService
}

func (a *userServiceAdapter) SignIn(ctx context.Context, input service.SignInInput) (string, error) {
	return a.userService.SignIn(ctx, input)
}

func (a *userServiceAdapter) VerifyCode(ctx context.Context, input service.VerifyCodeInput) (*service.AuthResponse, error) {
	return a.userService.VerifyCode(ctx, input)
}

func (a *userServiceAdapter) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error) {
	return a.authService.VerifyUserSession(ctx, userID, sessionID)
}

func (a *userServiceAdapter) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return a.authService.GetUserByID(ctx, userID)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.NewLogger()
	appLogger.Info("Starting API server")

	// Ensure system database exists
	if err := database.EnsureSystemDatabaseExists(&cfg.Database); err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to ensure system database exists")
		osExit(1)
		return
	}
	appLogger.Info("System database check completed")

	// Connect to system database
	systemDB, err := sql.Open("postgres", database.GetSystemDSN(&cfg.Database))
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to connect to system database")
		osExit(1)
		return
	}
	defer systemDB.Close()

	// Test database connection
	if err := systemDB.Ping(); err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to ping system database")
		osExit(1)
		return
	}

	// Initialize database schema if needed
	if err := database.InitializeDatabase(systemDB, cfg.RootEmail); err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to initialize database schema")
		osExit(1)
		return
	}

	// Set connection pool settings
	systemDB.SetMaxOpenConns(25)
	systemDB.SetMaxIdleConns(25)
	systemDB.SetConnMaxLifetime(5 * time.Minute)

	// Initialize repositories
	userRepo := repository.NewUserRepository(systemDB)
	workspaceRepo := repository.NewWorkspaceRepository(systemDB, &cfg.Database)
	authRepo := repository.NewSQLAuthRepository(systemDB, appLogger)
	emailSender := &emailSender{}

	// Initialize mailer
	var mailService mailer.Mailer
	if cfg.IsDevelopment() {
		// Use console mailer in development
		mailService = mailer.NewConsoleMailer()
		appLogger.Info("Using console mailer for development")
	} else {
		// Use SMTP mailer in production
		mailService = mailer.NewSMTPMailer(&mailer.Config{
			SMTPHost:     os.Getenv("SMTP_HOST"), // Get from environment variables or config
			SMTPPort:     587,                    // Default SMTP port
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
			FromEmail:    os.Getenv("FROM_EMAIL"),
			FromName:     "Notifuse",
			BaseURL:      os.Getenv("BASE_URL"),
		})
		appLogger.Info("Using SMTP mailer for production")
	}

	// Create auth service first
	authService, err := service.NewAuthService(service.AuthServiceConfig{
		Repository: authRepo,
		PrivateKey: cfg.Security.PasetoPrivateKeyBytes,
		PublicKey:  cfg.Security.PasetoPublicKeyBytes,
		Logger:     appLogger,
	})
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to create auth service")
		osExit(1)
		return
	}

	// Create adapter for AuthService to implement middleware.AuthServiceInterface
	authServiceAdapter := &authServiceMiddlewareAdapter{authService: authService}

	// Then create user service with auth service as dependency
	userService, err := service.NewUserService(service.UserServiceConfig{
		Repository:    userRepo,
		AuthService:   authService,
		EmailSender:   emailSender,
		SessionExpiry: 15 * 24 * time.Hour, // 15 days
		Logger:        appLogger,
		IsDevelopment: cfg.IsDevelopment(),
	})
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to create user service")
		osExit(1)
		return
	}

	// Create adapter for UserService
	userServiceAdapter := &userServiceAdapter{
		authService: authService,
		userService: userService,
	}

	// Create workspace service with mailer
	workspaceService := service.NewWorkspaceService(
		workspaceRepo,
		appLogger,
		userService,
		authService,
		mailService,
		cfg)

	// Use the already parsed PASETO public key
	userHandler := httpHandler.NewUserHandler(
		userServiceAdapter,
		workspaceService,
		cfg,
		cfg.Security.PasetoPublicKey,
		appLogger)
	rootHandler := httpHandler.NewRootHandler()
	workspaceHandler := httpHandler.NewWorkspaceHandler(
		workspaceService,
		authServiceAdapter,
		cfg.Security.PasetoPublicKey,
		appLogger)
	faviconHandler := httpHandler.NewFaviconHandler()

	// Set up routes
	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)
	workspaceHandler.RegisterRoutes(mux)
	rootHandler.RegisterRoutes(mux)
	mux.HandleFunc("/api/detect-favicon", faviconHandler.DetectFavicon)

	// Wrap mux with CORS middleware
	handler := middleware.CORSMiddleware(mux)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	appLogger.WithField("address", addr).Info("Server starting")

	if cfg.Server.SSL.Enabled {
		appLogger.WithField("cert_file", cfg.Server.SSL.CertFile).Info("SSL enabled")
		err = http.ListenAndServeTLS(addr, cfg.Server.SSL.CertFile, cfg.Server.SSL.KeyFile, handler)
	} else {
		err = http.ListenAndServe(addr, handler)
	}

	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Server failed to start")
		osExit(1)
		return
	}
}
