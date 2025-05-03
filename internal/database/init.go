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
				Type:      domain.UserTypeUser,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			query := `
				INSERT INTO users (id, email, name, type, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = db.Exec(query,
				rootUser.ID,
				rootUser.Email,
				rootUser.Name,
				rootUser.Type,
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
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS lists (
			id VARCHAR(32) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			is_double_optin BOOLEAN NOT NULL DEFAULT FALSE,
			is_public BOOLEAN NOT NULL DEFAULT FALSE,
			description TEXT,
			total_active INTEGER NOT NULL DEFAULT 0,
			total_pending INTEGER NOT NULL DEFAULT 0,
			total_unsubscribed INTEGER NOT NULL DEFAULT 0,
			total_bounced INTEGER NOT NULL DEFAULT 0,
			total_complained INTEGER NOT NULL DEFAULT 0,
			double_optin_template JSONB,
			welcome_template JSONB,
			unsubscribe_template JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS contact_lists (
			email VARCHAR(255) NOT NULL,
			list_id VARCHAR(32) NOT NULL,
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			PRIMARY KEY (email, list_id)
		)`,
		`CREATE TABLE IF NOT EXISTS templates (
			id VARCHAR(32) NOT NULL,
			name VARCHAR(255) NOT NULL,
			version INTEGER NOT NULL,
			channel VARCHAR(20) NOT NULL,
			email JSONB NOT NULL,
			category VARCHAR(20) NOT NULL,
			template_macro_id VARCHAR(32),
			utm_source VARCHAR(255),
			utm_medium VARCHAR(255),
			utm_campaign VARCHAR(255),
			test_data JSONB,
			settings JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			PRIMARY KEY (id, version)
		)`,
		`CREATE TABLE IF NOT EXISTS broadcasts (
			id VARCHAR(255) NOT NULL,
			workspace_id VARCHAR(32) NOT NULL,
			name VARCHAR(255) NOT NULL,
			status VARCHAR(20) NOT NULL,
			audience JSONB NOT NULL,
			schedule JSONB NOT NULL,
			test_settings JSONB NOT NULL,
			tracking_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			utm_parameters JSONB,
			metadata JSONB,
			total_sent INTEGER NOT NULL DEFAULT 0,
			total_delivered INTEGER NOT NULL DEFAULT 0,
			total_bounced INTEGER NOT NULL DEFAULT 0,
			total_complained INTEGER NOT NULL DEFAULT 0,
			total_failed INTEGER NOT NULL DEFAULT 0,
			total_opens INTEGER NOT NULL DEFAULT 0,
			total_clicks INTEGER NOT NULL DEFAULT 0,
			winning_variation VARCHAR(32),
			test_sent_at TIMESTAMP,
			winner_sent_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			cancelled_at TIMESTAMP,
			paused_at TIMESTAMP,
			PRIMARY KEY (id)
		)`,
		`CREATE TABLE IF NOT EXISTS message_history (
			id VARCHAR(255) NOT NULL PRIMARY KEY,
			contact_id VARCHAR(255) NOT NULL,
			broadcast_id VARCHAR(255),
			template_id VARCHAR(32) NOT NULL,
			template_version INTEGER NOT NULL,
			channel VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL,
			error VARCHAR(255),
			message_data JSONB NOT NULL,
			sent_at TIMESTAMP NOT NULL,
			delivered_at TIMESTAMP,
			failed_at TIMESTAMP,
			opened_at TIMESTAMP,
			clicked_at TIMESTAMP,
			bounced_at TIMESTAMP,
			complained_at TIMESTAMP,
			unsubscribed_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (contact_id) REFERENCES contacts(email) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS transactional_notifications (
			id VARCHAR(32) NOT NULL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			channels JSONB NOT NULL,
			is_public BOOLEAN NOT NULL DEFAULT FALSE,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_message_history_contact_id ON message_history(contact_id)`,
		`CREATE INDEX IF NOT EXISTS idx_message_history_broadcast_id ON message_history(broadcast_id)`,
		`CREATE INDEX IF NOT EXISTS idx_message_history_template_id ON message_history(template_id, template_version)`,
		`CREATE INDEX IF NOT EXISTS idx_message_history_status ON message_history(status)`,
		`CREATE INDEX IF NOT EXISTS idx_message_history_sent_at ON message_history(sent_at)`,
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
