package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mailer"
	"github.com/google/uuid"
)

type WorkspaceService struct {
	repo        domain.WorkspaceRepository
	logger      logger.Logger
	userService domain.UserServiceInterface
	authService domain.AuthService
	mailer      mailer.Mailer
	config      *config.Config
}

func NewWorkspaceService(
	repo domain.WorkspaceRepository,
	logger logger.Logger,
	userService domain.UserServiceInterface,
	authService domain.AuthService,
	mailerInstance mailer.Mailer,
	config *config.Config,
) *WorkspaceService {
	return &WorkspaceService{
		repo:        repo,
		logger:      logger,
		userService: userService,
		authService: authService,
		mailer:      mailerInstance,
		config:      config,
	}
}

// ListWorkspaces returns all workspaces for a user
func (s *WorkspaceService) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {

	user, err := s.authService.AuthenticateUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	userWorkspaces, err := s.repo.GetUserWorkspaces(ctx, user.ID)
	if err != nil {
		s.logger.WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to get user workspaces")
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
			s.logger.WithField("workspace_id", uw.WorkspaceID).WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to get workspace by ID")
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

// GetWorkspace returns a workspace by ID if the user has access
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	// Check if user has access to the workspace
	user, err := s.authService.AuthenticateUserForWorkspace(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	_, err = s.repo.GetUserWorkspace(ctx, user.ID, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to get user workspace")
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
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, coverURL string, timezone string) (*domain.Workspace, error) {

	user, err := s.authService.AuthenticateUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	workspace := &domain.Workspace{
		ID:   id,
		Name: name,
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			CoverURL:   coverURL,
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
		UserID:      user.ID,
		WorkspaceID: id,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := userWorkspace.Validate(); err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to validate user workspace")
		return nil, err
	}

	if err := s.repo.AddUserToWorkspace(ctx, userWorkspace); err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to add user to workspace")
		return nil, err
	}

	return workspace, nil
}

// UpdateWorkspace updates a workspace if the user is an owner
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, name string, websiteURL string, logoURL string, coverURL string, timezone string) (*domain.Workspace, error) {

	user, err := s.authService.AuthenticateUserForWorkspace(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Check if user is an owner
	userWorkspace, err := s.repo.GetUserWorkspace(ctx, user.ID, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to get user workspace")
		return nil, err
	}

	if userWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("role", userWorkspace.Role).Error("User is not an owner of the workspace")
		return nil, &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	workspace := &domain.Workspace{
		ID:   id,
		Name: name,
		Settings: domain.WorkspaceSettings{
			WebsiteURL: websiteURL,
			LogoURL:    logoURL,
			CoverURL:   coverURL,
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
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	// Check if user is an owner
	user, err := s.authService.AuthenticateUserForWorkspace(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	userWorkspace, err := s.repo.GetUserWorkspace(ctx, user.ID, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to get user workspace")
		return err
	}

	if userWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("role", userWorkspace.Role).Error("User is not an owner of the workspace")
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to delete workspace")
		return err
	}

	return nil
}

// AddUserToWorkspace adds a user to a workspace if the requester is an owner
func (s *WorkspaceService) AddUserToWorkspace(ctx context.Context, workspaceID string, userID string, role string) error {

	user, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Check if requester is an owner
	requesterWorkspace, err := s.repo.GetUserWorkspace(ctx, user.ID, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", user.ID).WithField("error", err.Error()).Error("Failed to get requester workspace")
		return err
	}

	if requesterWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", user.ID).WithField("role", requesterWorkspace.Role).Error("Requester is not an owner of the workspace")
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
func (s *WorkspaceService) RemoveUserFromWorkspace(ctx context.Context, workspaceID string, userID string) error {
	// Check if requester is an owner
	owner, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	requesterWorkspace, err := s.repo.GetUserWorkspace(ctx, owner.ID, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", owner.ID).WithField("error", err.Error()).Error("Failed to get requester workspace")
		return err
	}

	if requesterWorkspace.Role != "owner" {
		s.logger.WithField("workspace_id", workspaceID).WithField("user_id", userID).WithField("requester_id", owner.ID).WithField("role", requesterWorkspace.Role).Error("Requester is not an owner of the workspace")
		return &domain.ErrUnauthorized{Message: "user is not an owner of the workspace"}
	}

	// Prevent users from removing themselves
	if userID == owner.ID {
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
	// Authenticate the user
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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

// InviteMember creates an invitation for a user to join a workspace
func (s *WorkspaceService) InviteMember(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, string, error) {

	inviter, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate email format
	if !isValidEmail(email) {
		return nil, "", fmt.Errorf("invalid email format")
	}

	// Check if workspace exists
	workspace, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("error", err.Error()).Error("Failed to get workspace for invitation")
		return nil, "", err
	}
	if workspace == nil {
		return nil, "", fmt.Errorf("workspace not found")
	}

	// Check if the inviter has permission to invite members (is a member of the workspace)
	isMember, err := s.repo.IsUserWorkspaceMember(ctx, inviter.ID, workspaceID)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("inviter_id", inviter.ID).WithField("error", err.Error()).Error("Failed to check if inviter is a member")
		return nil, "", err
	}
	if !isMember {
		return nil, "", fmt.Errorf("inviter is not a member of the workspace")
	}

	// Get inviter user details for the email
	inviterDetails, err := s.userService.GetUserByID(ctx, inviter.ID)
	if err != nil {
		s.logger.WithField("inviter_id", inviter.ID).WithField("error", err.Error()).Error("Failed to get inviter details")
		return nil, "", err
	}
	inviterName := inviterDetails.Name
	if inviterName == "" {
		inviterName = inviterDetails.Email
	}

	// Check if user already exists with this email
	existingUser, err := s.userService.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		// User exists, check if they're already a member
		isMember, err := s.repo.IsUserWorkspaceMember(ctx, existingUser.ID, workspaceID)
		if err != nil {
			s.logger.WithField("workspace_id", workspaceID).WithField("user_id", existingUser.ID).WithField("error", err.Error()).Error("Failed to check if user is already a member")
			return nil, "", err
		}
		if isMember {
			return nil, "", fmt.Errorf("user is already a member of the workspace")
		}

		// User exists but is not a member, add them as a member
		userWorkspace := &domain.UserWorkspace{
			UserID:      existingUser.ID,
			WorkspaceID: workspaceID,
			Role:        "member", // Always set invited users as members
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err = s.repo.AddUserToWorkspace(ctx, userWorkspace)
		if err != nil {
			s.logger.WithField("workspace_id", workspaceID).WithField("user_id", existingUser.ID).WithField("error", err.Error()).Error("Failed to add user to workspace")
			return nil, "", err
		}

		// Return nil invitation since user was directly added
		return nil, "", nil
	}

	// User doesn't exist or there was an error (treat as user doesn't exist for security)
	// Create an invitation
	invitationID := uuid.New().String()
	expiresAt := time.Now().Add(15 * 24 * time.Hour) // 15 days

	invitation := &domain.WorkspaceInvitation{
		ID:          invitationID,
		WorkspaceID: workspaceID,
		InviterID:   inviter.ID,
		Email:       email,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.repo.CreateInvitation(ctx, invitation)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).WithField("email", email).WithField("error", err.Error()).Error("Failed to create workspace invitation")
		return nil, "", err
	}

	// Generate a PASETO token with the invitation details
	token := s.authService.GenerateInvitationToken(invitation)

	// Send invitation email in production mode
	if !s.config.IsDevelopment() {
		err = s.mailer.SendWorkspaceInvitation(email, workspace.Name, inviterName, token)
		if err != nil {
			s.logger.WithField("workspace_id", workspaceID).WithField("email", email).WithField("error", err.Error()).Error("Failed to send invitation email")
			// Continue even if email sending fails
		}

		// Only return the token in development mode
		return invitation, "", nil
	}

	// In development mode, return the token
	return invitation, token, nil
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Basic email validation - could be more sophisticated in production
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// GetWorkspaceMembersWithEmail returns all users with emails for a workspace, verifying the requester has access
func (s *WorkspaceService) GetWorkspaceMembersWithEmail(ctx context.Context, id string) ([]*domain.UserWorkspaceWithEmail, error) {
	// Check if requester has access to the workspace
	user, err := s.authService.AuthenticateUserForWorkspace(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	_, err = s.repo.GetUserWorkspace(ctx, user.ID, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to get user workspace")
		return nil, &domain.ErrUnauthorized{Message: "You do not have access to this workspace"}
	}

	// Get all workspace users with emails
	members, err := s.repo.GetWorkspaceUsersWithEmail(ctx, id)
	if err != nil {
		s.logger.WithField("workspace_id", id).WithField("error", err.Error()).Error("Failed to get workspace users with email")
		return nil, err
	}

	return members, nil
}
