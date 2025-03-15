package service

import (
	"context"
	"database/sql"
	"time"
)

type Workspace struct {
	ID         string
	Name       string
	WebsiteURL string
	LogoURL    string
	Timezone   string
	OwnerID    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type WorkspaceService struct {
	db *sql.DB
}

func NewWorkspaceService(db *sql.DB) *WorkspaceService {
	return &WorkspaceService{
		db: db,
	}
}

// ListWorkspaces returns all workspaces owned by the given user
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, ownerID string) ([]Workspace, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, website_url, logo_url, timezone, owner_id, created_at, updated_at FROM workspaces WHERE owner_id = $1",
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.WebsiteURL, &w.LogoURL, &w.Timezone, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, rows.Err()
}

// GetWorkspace returns a workspace by ID if it belongs to the given owner
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string, ownerID string) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, website_url, logo_url, timezone, owner_id, created_at, updated_at FROM workspaces WHERE id = $1 AND owner_id = $2",
		id, ownerID,
	).Scan(&w.ID, &w.Name, &w.WebsiteURL, &w.LogoURL, &w.Timezone, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// CreateWorkspace creates a new workspace for the given owner
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, name string, websiteURL string, logoURL string, timezone string, ownerID string) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO workspaces (name, website_url, logo_url, timezone, owner_id) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, name, website_url, logo_url, timezone, owner_id, created_at, updated_at`,
		name, websiteURL, logoURL, timezone, ownerID,
	).Scan(&w.ID, &w.Name, &w.WebsiteURL, &w.LogoURL, &w.Timezone, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &w, nil
}

// UpdateWorkspace updates a workspace if it belongs to the given owner
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, timezone string, ownerID string) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRowContext(ctx,
		`UPDATE workspaces 
		SET name = $1, website_url = $2, logo_url = $3, timezone = $4, updated_at = NOW() 
		WHERE id = $5 AND owner_id = $6 
		RETURNING id, name, website_url, logo_url, timezone, owner_id, created_at, updated_at`,
		name, websiteURL, logoURL, timezone, id, ownerID,
	).Scan(&w.ID, &w.Name, &w.WebsiteURL, &w.LogoURL, &w.Timezone, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// DeleteWorkspace deletes a workspace if it belongs to the given owner
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string, ownerID string) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM workspaces WHERE id = $1 AND owner_id = $2",
		id, ownerID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return nil
	}
	return nil
}
