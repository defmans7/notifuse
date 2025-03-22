// Package schema defines the database schema for development.
//
// DEVELOPMENT USE ONLY
// This file contains the current database schema and is used for development and testing.
// Before deploying to production, these table definitions should be converted to proper migrations.
package schema

// TableDefinitions contains all the SQL statements to create the database tables
// Don't put REFERENCES and don't put CHECK constraints in the CREATE TABLE statements
var TableDefinitions = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		name VARCHAR(255),
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS user_sessions (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		magic_code VARCHAR(255),
		magic_code_expires_at TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS workspaces (
		id VARCHAR(20) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		settings JSONB NOT NULL DEFAULT '{"timezone": "UTC"}',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS user_workspaces (
		user_id UUID NOT NULL,
		workspace_id VARCHAR(20) NOT NULL,
		role VARCHAR(20) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		PRIMARY KEY (user_id, workspace_id)
	)`,
	`CREATE TABLE IF NOT EXISTS workspace_invitations (
		id UUID PRIMARY KEY,
		workspace_id VARCHAR(20) NOT NULL,
		inviter_id UUID NOT NULL,
		email VARCHAR(255) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS contacts (
		uuid UUID PRIMARY KEY,
		external_id VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		timezone VARCHAR(50) NOT NULL,
		first_name VARCHAR(255),
		last_name VARCHAR(255),
		phone VARCHAR(50),
		address_line_1 VARCHAR(255),
		address_line_2 VARCHAR(255),
		country VARCHAR(50),
		postcode VARCHAR(20),
		state VARCHAR(50),
		job_title VARCHAR(100),
		lifetime_value DECIMAL(15, 2),
		orders_count INTEGER,
		last_order_at TIMESTAMP,
		custom_string_1 VARCHAR(255),
		custom_string_2 VARCHAR(255),
		custom_string_3 VARCHAR(255),
		custom_string_4 VARCHAR(255),
		custom_string_5 VARCHAR(255),
		custom_number_1 DECIMAL(15, 2),
		custom_number_2 DECIMAL(15, 2),
		custom_number_3 DECIMAL(15, 2),
		custom_number_4 DECIMAL(15, 2),
		custom_number_5 DECIMAL(15, 2),
		custom_datetime_1 TIMESTAMP,
		custom_datetime_2 TIMESTAMP,
		custom_datetime_3 TIMESTAMP,
		custom_datetime_4 TIMESTAMP,
		custom_datetime_5 TIMESTAMP,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS lists (
		id VARCHAR(20) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(20) NOT NULL,
		is_double_optin BOOLEAN NOT NULL DEFAULT FALSE,
		description TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS contact_lists (
		contact_id UUID NOT NULL,
		list_id VARCHAR(20) NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		PRIMARY KEY (contact_id, list_id)
	)`,
}

// TableNames returns a list of all table names in creation order
var TableNames = []string{
	"users",
	"user_sessions",
	"workspaces",
	"user_workspaces",
	"workspace_invitations",
	"contacts",
	"lists",
	"contact_lists",
}
