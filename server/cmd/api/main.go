package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
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

// osExit is a variable to allow mocking os.Exit in tests
var osExit = os.Exit

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

	// Create auth service first
	authService, err := service.NewAuthService(service.AuthServiceConfig{
		Repository: authRepo,
		PrivateKey: cfg.Security.PasetoPrivateKey,
		PublicKey:  cfg.Security.PasetoPublicKey,
		Logger:     appLogger,
	})
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to create auth service")
		osExit(1)
		return
	}

	// Then create user service with auth service as dependency
	userService, err := service.NewUserService(service.UserServiceConfig{
		Repository:    userRepo,
		AuthService:   authService,
		EmailSender:   emailSender,
		SessionExpiry: 15 * 24 * time.Hour, // 15 days
		Logger:        appLogger,
	})
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to create user service")
		osExit(1)
		return
	}

	// Create workspace service
	workspaceService := service.NewWorkspaceService(workspaceRepo, appLogger)

	// Parse public key for PASETO
	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromBytes(cfg.Security.PasetoPublicKey)
	if err != nil {
		appLogger.WithField("error", err.Error()).Fatal("Failed to parse PASETO public key")
		osExit(1)
		return
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
		osExit(1)
		return
	}
}
