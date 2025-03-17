package service

import (
	"context"
	"fmt"
	"notifuse/server/internal/domain"
	"time"
)

type WorkspaceService struct {
	repo domain.WorkspaceRepository
}

func NewWorkspaceService(repo domain.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{
		repo: repo,
	}
}

// ListWorkspaces returns all workspaces for a user
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	userWorkspaces, err := s.repo.GetUserWorkspaces(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Return empty array if user has no workspaces
	if len(userWorkspaces) == 0 {
		return []*domain.Workspace{}, nil
	}

	workspaces := make([]*domain.Workspace, 0, len(userWorkspaces))
	for _, uw := range userWorkspaces {
		workspace, err := s.repo.GetByID(ctx, uw.WorkspaceID)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

// GetWorkspace returns a workspace by ID if the user has access
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string, userID string) (*domain.Workspace, error) {
	// Check if user has access to the workspace
	_, err := s.repo.GetUserWorkspace(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, id)
}

// CreateWorkspace creates a new workspace and adds the creator as owner
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, timezone string, ownerID string) (*domain.Workspace, error) {
	workspace := &domain.Workspace{
		ID:   id,
		Name: name,
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			Timezone:   timezone,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := workspace.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	// Add the creator as owner
	userWorkspace := &domain.UserWorkspace{
		UserID:      ownerID,
		WorkspaceID: id,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := userWorkspace.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.AddUserToWorkspace(ctx, userWorkspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// UpdateWorkspace updates a workspace if the user is an owner
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, timezone string, userID string) (*domain.Workspace, error) {
	// Check if user is an owner
	userWorkspace, err := s.repo.GetUserWorkspace(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if userWorkspace.Role != "owner" {
		return nil, &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	workspace := &domain.Workspace{
		ID:   id,
		Name: name,
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			Timezone:   timezone,
		},
		UpdatedAt: time.Now(),
	}

	if err := workspace.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// DeleteWorkspace deletes a workspace if the user is an owner
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string, userID string) error {
	// Check if user is an owner
	userWorkspace, err := s.repo.GetUserWorkspace(ctx, userID, id)
	if err != nil {
		return err
	}

	if userWorkspace.Role != "owner" {
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	return s.repo.Delete(ctx, id)
}

// AddUserToWorkspace adds a user to a workspace if the requester is an owner
func (s *WorkspaceService) AddUserToWorkspace(ctx context.Context, workspaceID string, userID string, role string, requesterID string) error {
	// Check if requester is an owner
	requesterWorkspace, err := s.repo.GetUserWorkspace(ctx, requesterID, workspaceID)
	if err != nil {
		return err
	}

	if requesterWorkspace.Role != "owner" {
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	userWorkspace := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        role,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := userWorkspace.Validate(); err != nil {
		return err
	}

	return s.repo.AddUserToWorkspace(ctx, userWorkspace)
}

// RemoveUserFromWorkspace removes a user from a workspace if the requester is an owner
func (s *WorkspaceService) RemoveUserFromWorkspace(ctx context.Context, workspaceID string, userID string, requesterID string) error {
	// Check if requester is an owner
	requesterWorkspace, err := s.repo.GetUserWorkspace(ctx, requesterID, workspaceID)
	if err != nil {
		return err
	}

	if requesterWorkspace.Role != "owner" {
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	return s.repo.RemoveUserFromWorkspace(ctx, userID, workspaceID)
}

// TransferOwnership transfers the ownership of a workspace from the current owner to a member
func (s *WorkspaceService) TransferOwnership(ctx context.Context, workspaceID string, newOwnerID string, currentOwnerID string) error {
	// Check if current owner is actually an owner
	currentOwnerWorkspace, err := s.repo.GetUserWorkspace(ctx, currentOwnerID, workspaceID)
	if err != nil {
		return err
	}

	if currentOwnerWorkspace.Role != "owner" {
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	// Check if new owner exists and is a member
	newOwnerWorkspace, err := s.repo.GetUserWorkspace(ctx, newOwnerID, workspaceID)
	if err != nil {
		return err
	}

	if newOwnerWorkspace.Role != "member" {
		return fmt.Errorf("new owner must be a current member of the workspace")
	}

	// Update new owner's role to owner
	newOwnerWorkspace.Role = "owner"
	newOwnerWorkspace.UpdatedAt = time.Now()
	if err := s.repo.AddUserToWorkspace(ctx, newOwnerWorkspace); err != nil {
		return err
	}

	// Update current owner's role to member
	currentOwnerWorkspace.Role = "member"
	currentOwnerWorkspace.UpdatedAt = time.Now()
	if err := s.repo.AddUserToWorkspace(ctx, currentOwnerWorkspace); err != nil {
		return err
	}

	return nil
}
