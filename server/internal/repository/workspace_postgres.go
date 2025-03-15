package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

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

func (r *workspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	if workspace.ID == "" {
		workspace.ID = uuid.New().String()
	}
	now := time.Now()
	workspace.CreatedAt = now
	workspace.UpdatedAt = now

	query := `
		INSERT INTO workspaces (id, name, website_url, logo_url, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.systemDB.ExecContext(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.WebsiteURL,
		workspace.LogoURL,
		workspace.Timezone,
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
	workspace := &domain.Workspace{}
	query := `
		SELECT id, name, website_url, logo_url, timezone, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`
	err := r.systemDB.QueryRowContext(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.Name,
		&workspace.WebsiteURL,
		&workspace.LogoURL,
		&workspace.Timezone,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)
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
		SELECT id, name, website_url, logo_url, timezone, created_at, updated_at
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
		workspace := &domain.Workspace{}
		err := rows.Scan(
			&workspace.ID,
			&workspace.Name,
			&workspace.WebsiteURL,
			&workspace.LogoURL,
			&workspace.Timezone,
			&workspace.CreatedAt,
			&workspace.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}
	return workspaces, nil
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	workspace.UpdatedAt = time.Now()
	query := `
		UPDATE workspaces
		SET name = $1, website_url = $2, logo_url = $3, timezone = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.systemDB.ExecContext(ctx, query,
		workspace.Name,
		workspace.WebsiteURL,
		workspace.LogoURL,
		workspace.Timezone,
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

	// Create workspace schema (implement your schema creation here)
	// For example:
	_, err = db.ExecContext(ctx, `
		CREATE TABLE contacts (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE TABLE subscription_lists (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
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
	if conn, ok := r.connections.Load(workspaceID); ok {
		db := conn.(*sql.DB)
		db.Close()
		r.connections.Delete(workspaceID)
	}

	// Drop the database
	safeID := strings.ReplaceAll(workspaceID, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", r.dbConfig.Prefix, safeID)
	_, err := r.systemDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to delete workspace database: %w", err)
	}

	return nil
}
