package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Notifuse/notifuse/config"
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
