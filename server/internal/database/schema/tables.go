// Package schema defines the database schema for development.
//
// DEVELOPMENT USE ONLY
// This file contains the current database schema and is used for development and testing.
// Before deploying to production, these table definitions should be converted to proper migrations.
package schema

// TableDefinitions contains all the SQL statements to create the database tables
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
		user_id UUID NOT NULL REFERENCES users(id),
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		magic_code VARCHAR(255),
		magic_code_expires_at TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS workspaces (
		id VARCHAR(20) PRIMARY KEY CHECK (id ~ '^[a-zA-Z0-9]+$'),
		name VARCHAR(255) NOT NULL,
		website_url VARCHAR(255),
		logo_url VARCHAR(255),
		timezone VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`,
}

// TableNames returns a list of all table names in creation order
var TableNames = []string{
	"users",
	"user_sessions",
	"workspaces",
}
