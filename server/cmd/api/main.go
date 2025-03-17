package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"notifuse/server/config"
	"notifuse/server/internal/database"
	httpHandler "notifuse/server/internal/http"
	"notifuse/server/internal/repository"
	"notifuse/server/internal/service"
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

	// Connect to system database
	systemDB, err := sql.Open("postgres", database.GetSystemDSN(&cfg.Database))
	if err != nil {
		log.Fatalf("Failed to connect to system database: %v", err)
	}
	defer systemDB.Close()

	// Test database connection
	if err := systemDB.Ping(); err != nil {
		log.Fatalf("Failed to ping system database: %v", err)
	}

	// Initialize database schema if needed
	if err := database.InitializeDatabase(systemDB, cfg.RootEmail); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Set connection pool settings
	systemDB.SetMaxOpenConns(25)
	systemDB.SetMaxIdleConns(25)
	systemDB.SetConnMaxLifetime(5 * time.Minute)

	// Initialize components
	userRepo := repository.NewUserRepository(systemDB)
	emailSender := &emailSender{}

	userService, err := service.NewUserService(service.UserServiceConfig{
		Repository:    userRepo,
		PrivateKey:    []byte(cfg.Security.PasetoPrivateKey),
		PublicKey:     []byte(cfg.Security.PasetoPublicKey),
		EmailSender:   emailSender,
		SessionExpiry: 15 * 24 * time.Hour, // 15 days
	})
	if err != nil {
		log.Fatalf("Failed to create user service: %v", err)
	}

	userHandler := httpHandler.NewUserHandler(userService)

	// Set up routes
	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
