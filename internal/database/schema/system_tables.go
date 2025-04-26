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
}

// TableNames returns a list of all table names in creation order
var TableNames = []string{
	"users",
	"user_sessions",
	"workspaces",
	"user_workspaces",
	"workspace_invitations",
	"contacts",
	"broadcasts",
}
