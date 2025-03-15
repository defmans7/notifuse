package repository

import (
	"database/sql"
	"fmt"
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

	// Test connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Clean up database
	if err := cleanDatabase(db); err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func cleanDatabase(db *sql.DB) error {
	tables := []string{"user_sessions", "users", "user_workspaces", "workspaces"}
	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	return nil
}

func runMigrations(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE user_sessions (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE workspaces (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			website_url VARCHAR(255),
			logo_url VARCHAR(255),
			timezone VARCHAR(50),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}
