package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"notifuse/server/config"
	"notifuse/server/internal/database"
)

// setupTestDB connects to a real PostgreSQL database for testing
// This should be used when full database integration is needed
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
	if err := database.InitializeDatabase(db, cfg.RootEmail); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return db
}

// setupMockTestDB creates a mock SQL database for testing
// This should be used for unit tests that don't require a real database
func setupMockTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	// Use regexp matcher instead of exact match for more flexibility with SQL queries
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err, "Failed to create mock database")
	return db, mock
}
