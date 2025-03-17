package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"notifuse/server/config"
	"notifuse/server/internal/database"
	"notifuse/server/internal/domain"
)

type workspaceRepository struct {
	systemDB *sql.DB
	dbConfig *config.DatabaseConfig

	// Connection pool for workspace databases
	connections sync.Map
}

// NewWorkspaceRepository creates a new PostgreSQL workspace repository
func NewWorkspaceRepository(systemDB *sql.DB, dbConfig *config.DatabaseConfig) domain.WorkspaceRepository {
	return &workspaceRepository{
		systemDB: systemDB,
		dbConfig: dbConfig,
	}
}

// checkWorkspaceIDExists checks if a workspace with the given ID already exists
func (r *workspaceRepository) checkWorkspaceIDExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1)`
	err := r.systemDB.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		log.Printf("failed to check workspace ID existence: %v", err)
		return false, fmt.Errorf("failed to check workspace ID existence: %w", err)
	}
	return exists, nil
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	if workspace.ID == "" {
		return fmt.Errorf("workspace ID is required")
	}

	// Validate workspace before creating
	if err := workspace.Validate(); err != nil {
		return err
	}

	// Check if workspace ID already exists
	exists, err := r.checkWorkspaceIDExists(ctx, workspace.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("workspace with ID %s already exists", workspace.ID)
	}

	now := time.Now()
	workspace.CreatedAt = now
	workspace.UpdatedAt = now

	// Marshal settings to JSON
	settings, err := json.Marshal(workspace.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO workspaces (id, name, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = r.systemDB.ExecContext(ctx, query,
		workspace.ID,
		workspace.Name,
		settings,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	// Create the workspace database
	return r.CreateDatabase(ctx, workspace.ID)
}

func (r *workspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	query := `
		SELECT id, name, settings, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`
	workspace, err := domain.ScanWorkspace(r.systemDB.QueryRowContext(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workspace not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	return workspace, nil
}

func (r *workspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	query := `
		SELECT id, name, settings, created_at, updated_at
		FROM workspaces
		ORDER BY created_at DESC
	`
	rows, err := r.systemDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		workspace, err := domain.ScanWorkspace(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}
	return workspaces, rows.Err()
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	workspace.UpdatedAt = time.Now()

	// Validate workspace before updating
	if err := workspace.Validate(); err != nil {
		return err
	}

	// Marshal settings to JSON
	settings, err := json.Marshal(workspace.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE workspaces
		SET name = $1, settings = $2, updated_at = $3
		WHERE id = $4
	`
	result, err := r.systemDB.ExecContext(ctx, query,
		workspace.Name,
		settings,
		workspace.UpdatedAt,
		workspace.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("workspace not found")
	}
	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id string) error {
	// Delete the workspace database first
	if err := r.DeleteDatabase(ctx, id); err != nil {
		return err
	}

	// Then delete the workspace record
	query := `DELETE FROM workspaces WHERE id = $1`
	result, err := r.systemDB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("workspace not found")
	}
	return nil
}

func (r *workspaceRepository) GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error) {
	// Check if we already have a connection
	if conn, ok := r.connections.Load(workspaceID); ok {
		db := conn.(*sql.DB)
		// Test the connection
		if err := db.PingContext(ctx); err == nil {
			return db, nil
		}
		// If ping fails, remove the connection and create a new one
		r.connections.Delete(workspaceID)
	}

	// Create a new connection
	db, err := database.ConnectToWorkspace(r.dbConfig, workspaceID)
	if err != nil {
		return nil, err
	}

	// Store the connection
	r.connections.Store(workspaceID, db)
	return db, nil
}

func (r *workspaceRepository) CreateDatabase(ctx context.Context, workspaceID string) error {
	// Replace hyphens with underscores for PostgreSQL compatibility
	safeID := strings.ReplaceAll(workspaceID, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", r.dbConfig.Prefix, safeID)

	// Create the database
	_, err := r.systemDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to create workspace database: %w", err)
	}

	// Connect to the new database to create schema
	db, err := database.ConnectToWorkspace(r.dbConfig, workspaceID)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create workspace schema
	_, err = db.ExecContext(ctx, `
		CREATE TABLE contacts (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create workspace schema: %w", err)
	}

	return nil
}

func (r *workspaceRepository) DeleteDatabase(ctx context.Context, workspaceID string) error {
	// Remove the connection from the pool if it exists
	r.connections.Delete(workspaceID)

	// Replace hyphens with underscores for PostgreSQL compatibility
	safeID := strings.ReplaceAll(workspaceID, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", r.dbConfig.Prefix, safeID)

	// Drop the database
	_, err := r.systemDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to delete workspace database: %w", err)
	}

	return nil
}

func (r *workspaceRepository) AddUserToWorkspace(ctx context.Context, userWorkspace *domain.UserWorkspace) error {
	query := `
		INSERT INTO user_workspaces (user_id, workspace_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, workspace_id) DO UPDATE
		SET role = $3, updated_at = $5
	`
	_, err := r.systemDB.ExecContext(ctx, query,
		userWorkspace.UserID,
		userWorkspace.WorkspaceID,
		userWorkspace.Role,
		userWorkspace.CreatedAt,
		userWorkspace.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add user to workspace: %w", err)
	}
	return nil
}

func (r *workspaceRepository) RemoveUserFromWorkspace(ctx context.Context, userID string, workspaceID string) error {
	query := `DELETE FROM user_workspaces WHERE user_id = $1 AND workspace_id = $2`
	result, err := r.systemDB.ExecContext(ctx, query, userID, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to remove user from workspace: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user is not a member of the workspace")
	}
	return nil
}

func (r *workspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.UserWorkspace, error) {
	query := `
		SELECT user_id, workspace_id, role, created_at, updated_at
		FROM user_workspaces
		WHERE user_id = $1
	`
	rows, err := r.systemDB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}
	defer rows.Close()

	var userWorkspaces []*domain.UserWorkspace
	for rows.Next() {
		var uw domain.UserWorkspace
		err := rows.Scan(&uw.UserID, &uw.WorkspaceID, &uw.Role, &uw.CreatedAt, &uw.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user workspace: %w", err)
		}
		userWorkspaces = append(userWorkspaces, &uw)
	}
	return userWorkspaces, rows.Err()
}

func (r *workspaceRepository) GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*domain.UserWorkspace, error) {
	query := `
		SELECT user_id, workspace_id, role, created_at, updated_at
		FROM user_workspaces
		WHERE workspace_id = $1
	`
	rows, err := r.systemDB.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace users: %w", err)
	}
	defer rows.Close()

	var userWorkspaces []*domain.UserWorkspace
	for rows.Next() {
		var uw domain.UserWorkspace
		err := rows.Scan(&uw.UserID, &uw.WorkspaceID, &uw.Role, &uw.CreatedAt, &uw.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user workspace: %w", err)
		}
		userWorkspaces = append(userWorkspaces, &uw)
	}
	return userWorkspaces, rows.Err()
}

func (r *workspaceRepository) GetUserWorkspace(ctx context.Context, userID string, workspaceID string) (*domain.UserWorkspace, error) {
	query := `
		SELECT user_id, workspace_id, role, created_at, updated_at
		FROM user_workspaces
		WHERE user_id = $1 AND workspace_id = $2
	`
	var uw domain.UserWorkspace
	err := r.systemDB.QueryRowContext(ctx, query, userID, workspaceID).Scan(
		&uw.UserID, &uw.WorkspaceID, &uw.Role, &uw.CreatedAt, &uw.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user is not a member of the workspace")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspace: %w", err)
	}
	return &uw, nil
}
