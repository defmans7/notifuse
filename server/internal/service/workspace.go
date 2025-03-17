package service

import (
	"context"
	"notifuse/server/internal/domain"
)

type WorkspaceService struct {
	repo domain.WorkspaceRepository
}

func NewWorkspaceService(repo domain.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{
		repo: repo,
	}
}

// ListWorkspaces returns all workspaces
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, ownerID string) ([]*domain.Workspace, error) {
	return s.repo.List(ctx)
}

// GetWorkspace returns a workspace by ID
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string, ownerID string) (*domain.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, timezone string, ownerID string) (*domain.Workspace, error) {
	workspace := &domain.Workspace{
		ID:   id,
		Name: name,
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			Timezone:   timezone,
		},
	}

	if err := workspace.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// UpdateWorkspace updates a workspace
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, timezone string, ownerID string) (*domain.Workspace, error) {
	workspace := &domain.Workspace{
		ID:   id,
		Name: name,
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			Timezone:   timezone,
		},
	}

	if err := workspace.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// DeleteWorkspace deletes a workspace
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string, ownerID string) error {
	return s.repo.Delete(ctx, id)
}
