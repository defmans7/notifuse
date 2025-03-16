package repository

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"

	"notifuse/server/config"
	"notifuse/server/internal/database"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Connect to database using configuration
	db, err := sql.Open("postgres", database.GetSystemDSN(&cfg.Database))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Clean up database
	if err := database.CleanDatabase(db); err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Initialize database with tables
	if err := database.InitializeDatabase(db); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return db
}
