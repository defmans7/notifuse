package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Notifuse/notifuse/internal/database/schema"
	"github.com/Notifuse/notifuse/internal/domain"
)

// InitializeDatabase creates all necessary database tables if they don't exist
func InitializeDatabase(db *sql.DB, rootEmail string) error {
	// Run all table creation queries
	for _, query := range schema.TableDefinitions {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create root user if it doesn't exist
	if rootEmail != "" {
		// Check if root user exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", rootEmail).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check root user existence: %w", err)
		}

		if !exists {
			// Create root user
			rootUser := &domain.User{
				ID:        uuid.New().String(),
				Email:     rootEmail,
				Name:      "Root User",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			query := `
				INSERT INTO users (id, email, name, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
			`
			_, err = db.Exec(query,
				rootUser.ID,
				rootUser.Email,
				rootUser.Name,
				rootUser.CreatedAt,
				rootUser.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to create root user: %w", err)
			}
		}
	}

	return nil
}

// InitializeWorkspaceDatabase creates the necessary tables for a workspace database
func InitializeWorkspaceDatabase(db *sql.DB) error {
	// Create workspace tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS contacts (
			email VARCHAR(255) PRIMARY KEY,
			external_id VARCHAR(255),
			timezone VARCHAR(50),
			language VARCHAR(50),
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			phone VARCHAR(50),
			address_line_1 VARCHAR(255),
			address_line_2 VARCHAR(255),
			country VARCHAR(100),
			postcode VARCHAR(20),
			state VARCHAR(100),
			job_title VARCHAR(255),
			lifetime_value DECIMAL,
			orders_count INTEGER,
			last_order_at TIMESTAMP,
			custom_string_1 VARCHAR(255),
			custom_string_2 VARCHAR(255),
			custom_string_3 VARCHAR(255),
			custom_string_4 VARCHAR(255),
			custom_string_5 VARCHAR(255),
			custom_number_1 DECIMAL,
			custom_number_2 DECIMAL,
			custom_number_3 DECIMAL,
			custom_number_4 DECIMAL,
			custom_number_5 DECIMAL,
			custom_datetime_1 TIMESTAMP,
			custom_datetime_2 TIMESTAMP,
			custom_datetime_3 TIMESTAMP,
			custom_datetime_4 TIMESTAMP,
			custom_datetime_5 TIMESTAMP,
			custom_json_1 JSONB,
			custom_json_2 JSONB,
			custom_json_3 JSONB,
			custom_json_4 JSONB,
			custom_json_5 JSONB,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS lists (
			id VARCHAR(20) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			is_double_optin BOOLEAN NOT NULL DEFAULT FALSE,
			is_public BOOLEAN NOT NULL DEFAULT FALSE,
			description TEXT,
			total_active INTEGER NOT NULL DEFAULT 0,
			total_pending INTEGER NOT NULL DEFAULT 0,
			total_unsubscribed INTEGER NOT NULL DEFAULT 0,
			total_bounced INTEGER NOT NULL DEFAULT 0,
			total_complained INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS contact_lists (
			email VARCHAR(255) NOT NULL,
			list_id VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			PRIMARY KEY (email, list_id)
		)`,
	}

	// Run all table creation queries
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create workspace table: %w", err)
		}
	}

	return nil
}

// CleanDatabase drops all tables in reverse order
func CleanDatabase(db *sql.DB) error {
	// Drop tables in reverse order to handle dependencies
	for i := len(schema.TableNames) - 1; i >= 0; i-- {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", schema.TableNames[i])
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", schema.TableNames[i], err)
		}
	}
	return nil
}
