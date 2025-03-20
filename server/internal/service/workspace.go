package service

import (
	"context"
	"fmt"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"time"
)

type WorkspaceService struct {
	repo   domain.WorkspaceRepository
	logger logger.Logger
}

func NewWorkspaceService(repo domain.WorkspaceRepository, logger logger.Logger) *WorkspaceService {
	return &WorkspaceService{
		repo:   repo,
		logger: logger,
	}
}

// ListWorkspaces returns all workspaces for a user
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	userWorkspaces, err := s.repo.GetUserWorkspaces(ctx, userID)
	if err != nil {
		s.logger.WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get user workspaces")
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
			s.logger.WithField("workspace_id", uw.WorkspaceID).WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get workspace by ID")
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
		s.logger.WithField("workspace_id", id).WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get user workspace")
		return nil, err
	}

	workspace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to get workspace by ID")
		return nil, err
	}

	return workspace, nil
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
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to validate workspace")
		return nil, err
	}

	if err := s.repo.Create(ctx, workspace); err != nil {
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to create workspace")
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
		s.logger.WithField("workspace_id", id).WithField("user_id", ownerID).WithField("error", err.Error()).Error("Failed to validate user workspace")
		return nil, err
	}

	if err := s.repo.AddUserToWorkspace(ctx, userWorkspace); err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", ownerID).WithField("error", err.Error()).Error("Failed to add user to workspace")
		return nil, err
	}

	return workspace, nil
}

// UpdateWorkspace updates a workspace if the user is an owner
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, timezone string, userID string) (*domain.Workspace, error) {
	// Check if user is an owner
	userWorkspace, err := s.repo.GetUserWorkspace(ctx, userID, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get user workspace")
		return nil, err
	}

	if userWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", id).WithField("user_id", userID).WithField("role", userWorkspace.Role).Error("User is not an owner of the workspace")
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
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to validate workspace")
		return nil, err
	}

	if err := s.repo.Update(ctx, workspace); err != nil {
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to update workspace")
		return nil, err
	}

	return workspace, nil
}

// DeleteWorkspace deletes a workspace if the user is an owner
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string, userID string) error {
	// Check if user is an owner
	userWorkspace, err := s.repo.GetUserWorkspace(ctx, userID, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get user workspace")
		return err
	}

	if userWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", id).WithField("user_id", userID).WithField("role", userWorkspace.Role).Error("User is not an owner of the workspace")
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to delete workspace")
		return err
	}

	return nil
}

// AddUserToWorkspace adds a user to a workspace if the requester is an owner
func (s *WorkspaceService) AddUserToWorkspace(ctx context.Context, workspaceID string, userID string, role string, requesterID string) error {
	// Check if requester is an owner
	requesterWorkspace, err := s.repo.GetUserWorkspace(ctx, requesterID, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", requesterID).WithField("error", err.Error()).Error("Failed to get requester workspace")
		return err
	}

	if requesterWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", requesterID).WithField("role", requesterWorkspace.Role).Error("Requester is not an owner of the workspace")
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
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to validate user workspace")
		return err
	}

	if err := s.repo.AddUserToWorkspace(ctx, userWorkspace); err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to add user to workspace")
		return err
	}

	return nil
}

// RemoveUserFromWorkspace removes a user from a workspace if the requester is an owner
func (s *WorkspaceService) RemoveUserFromWorkspace(ctx context.Context, workspaceID string, userID string, requesterID string) error {
	// Check if requester is an owner
	requesterWorkspace, err := s.repo.GetUserWorkspace(ctx, requesterID, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", requesterID).WithField("error", err.Error()).Error("Failed to get requester workspace")
		return err
	}

	if requesterWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", requesterID).WithField("role", requesterWorkspace.Role).Error("Requester is not an owner of the workspace")
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	// Prevent users from removing themselves
	if userID == requesterID {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).Error("Cannot remove self from workspace")
		return fmt.Errorf("cannot remove yourself from the workspace")
	}

	if err := s.repo.RemoveUserFromWorkspace(ctx, userID, workspaceID); err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to remove user from workspace")
		return err
	}

	return nil
}

// TransferOwnership transfers the ownership of a workspace from the current owner to a member
func (s *WorkspaceService) TransferOwnership(ctx context.Context, workspaceID string, newOwnerID string, currentOwnerID string) error {
	// Check if current owner is actually an owner
	currentOwnerWorkspace, err := s.repo.GetUserWorkspace(ctx, currentOwnerID, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("current_owner_id", currentOwnerID).WithField("new_owner_id", newOwnerID).WithField("error", err.Error()).Error("Failed to get current owner workspace")
		return err
	}

	if currentOwnerWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", workspaceID).WithField("current_owner_id", currentOwnerID).WithField("role", currentOwnerWorkspace.Role).Error("Current owner is not an owner of the workspace")
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	// Check if new owner exists and is a member
	newOwnerWorkspace, err := s.repo.GetUserWorkspace(ctx, newOwnerID, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("new_owner_id", newOwnerID).WithField("error", err.Error()).Error("Failed to get new owner workspace")
		return err
	}

	if newOwnerWorkspace.Role != "member" {
		s.logger.WithField("workspace_id", workspaceID).WithField("new_owner_id", newOwnerID).WithField("role", newOwnerWorkspace.Role).Error("New owner must be a current member of the workspace")
		return fmt.Errorf("new owner must be a current member of the workspace")
	}

	// Update new owner's role to owner
	newOwnerWorkspace.Role = "owner"
	newOwnerWorkspace.UpdatedAt = time.Now()
	if err := s.repo.AddUserToWorkspace(ctx, newOwnerWorkspace); err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("new_owner_id", newOwnerID).WithField("error", err.Error()).Error("Failed to update new owner's role")
		return err
	}

	// Update current owner's role to member
	currentOwnerWorkspace.Role = "member"
	currentOwnerWorkspace.UpdatedAt = time.Now()
	if err := s.repo.AddUserToWorkspace(ctx, currentOwnerWorkspace); err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("current_owner_id", currentOwnerID).WithField("error", err.Error()).Error("Failed to update current owner's role")
		return err
	}

	return nil
}
