package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"aidanwoods.dev/go-paseto"
	_ "github.com/lib/pq"

	"notifuse/server/config"
	"notifuse/server/internal/database"
	httpHandler "notifuse/server/internal/http"
	"notifuse/server/internal/http/middleware"
	"notifuse/server/internal/repository"
	"notifuse/server/internal/service"
	"notifuse/server/pkg/logger"
)

type emailSender struct{}

func (s *emailSender) SendMagicCode(email, code string) error {
	// TODO: Implement email sending using SMTP
	log.Printf("Sending magic code to %s: %s", email, code)
	return nil
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

	// Connect to system database
	systemDB, err := sql.Open("postgres", database.GetSystemDSN(&cfg.Database))
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to connect to system database")
	}
	defer systemDB.Close()

	// Test database connection
	if err := systemDB.Ping(); err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to ping system database")
	}

	// Initialize database schema if needed
	if err := database.InitializeDatabase(systemDB, cfg.RootEmail); err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to initialize database schema")
	}

	// Set connection pool settings
	systemDB.SetMaxOpenConns(25)
	systemDB.SetMaxIdleConns(25)
	systemDB.SetConnMaxLifetime(5 * time.Minute)

	// Initialize components
	userRepo := repository.NewUserRepository(systemDB)
	workspaceRepo := repository.NewWorkspaceRepository(systemDB, &cfg.Database)
	authRepo := repository.NewSQLAuthRepository(systemDB, appLogger)
	emailSender := &emailSender{}

	userService, err := service.NewUserService(service.UserServiceConfig{
		Repository:    userRepo,
		PrivateKey:    []byte(cfg.Security.PasetoPrivateKey),
		PublicKey:     []byte(cfg.Security.PasetoPublicKey),
		EmailSender:   emailSender,
		SessionExpiry: 15 * 24 * time.Hour, // 15 days
		Logger:        appLogger,
	})
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to create user service")
	}

	// Create auth service
	authService := service.NewAuthService(authRepo, appLogger)

	// Create workspace service
	workspaceService := service.NewWorkspaceService(workspaceRepo, appLogger)

	// Parse public key for PASETO
	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromBytes([]byte(cfg.Security.PasetoPublicKey))
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to parse PASETO public key")
	}

	userHandler := httpHandler.NewUserHandler(userService, workspaceService, cfg, publicKey)
	rootHandler := httpHandler.NewRootHandler()
	workspaceHandler := httpHandler.NewWorkspaceHandler(workspaceService, authService, publicKey)
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
	}
}
