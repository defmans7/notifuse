package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Notifuse/notifuse/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// GetSystemDSN returns the DSN for the system database
func GetSystemDSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
}

// GetPostgresDSN returns the DSN for connecting to PostgreSQL server without specifying a database
func GetPostgresDSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)
}

// GetWorkspaceDSN returns the DSN for a workspace database
func GetWorkspaceDSN(cfg *config.DatabaseConfig, workspaceID string) string {
	// Replace hyphens with underscores for PostgreSQL compatibility
	safeID := strings.ReplaceAll(workspaceID, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", cfg.Prefix, safeID)
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		dbName,
	)
}

// ConnectToWorkspace creates a new database connection for a workspace
func ConnectToWorkspace(cfg *config.DatabaseConfig, workspaceID string) (*sql.DB, error) {
	// Ensure the workspace database exists
	if err := EnsureWorkspaceDatabaseExists(cfg, workspaceID); err != nil {
		return nil, fmt.Errorf("failed to ensure workspace database exists: %w", err)
	}

	dsn := GetWorkspaceDSN(cfg, workspaceID)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to workspace database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping workspace database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}

// EnsureWorkspaceDatabaseExists creates the workspace database if it doesn't exist
func EnsureWorkspaceDatabaseExists(cfg *config.DatabaseConfig, workspaceID string) error {
	// Replace hyphens with underscores for PostgreSQL compatibility
	safeID := strings.ReplaceAll(workspaceID, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", cfg.Prefix, safeID)

	// Connect to PostgreSQL server without specifying a database
	pgDSN := GetPostgresDSN(cfg)
	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL server: %w", err)
	}

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		// Use fmt.Sprintf for proper quoting of identifiers in SQL
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s",
			// Proper quoting to prevent SQL injection
			strings.ReplaceAll(dbName, `"`, `""`))

		_, err = db.Exec(createDBQuery)
		if err != nil {
			return fmt.Errorf("failed to create workspace database: %w", err)
		}

		// Connect to the new database to initialize schema
		wsDB, err := sql.Open("postgres", GetWorkspaceDSN(cfg, workspaceID))
		if err != nil {
			return fmt.Errorf("failed to connect to new workspace database: %w", err)
		}
		defer wsDB.Close()

		// Test the connection
		if err := wsDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping new workspace database: %w", err)
		}

		// Initialize the workspace database schema
		if err := InitializeWorkspaceDatabase(wsDB); err != nil {
			return fmt.Errorf("failed to initialize workspace database schema: %w", err)
		}
	}

	return nil
}

// EnsureSystemDatabaseExists creates the system database if it doesn't exist
func EnsureSystemDatabaseExists(cfg *config.DatabaseConfig) error {
	// Connect to PostgreSQL server without specifying a database
	pgDSN := GetPostgresDSN(cfg)
	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL server: %w", err)
	}

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, cfg.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		// Use fmt.Sprintf for proper quoting of identifiers in SQL
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s",
			// Proper quoting to prevent SQL injection
			strings.ReplaceAll(cfg.DBName, `"`, `""`))

		_, err = db.Exec(createDBQuery)
		if err != nil {
			return fmt.Errorf("failed to create system database: %w", err)
		}
	}

	return nil
}
