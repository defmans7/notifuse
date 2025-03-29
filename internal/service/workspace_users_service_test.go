package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceService_AddUserToWorkspace(t *testing.T) {
	mockRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := createMockUserService()
	mockAuthService := createMockAuthService()
	mockMailer := &MockMailer{}
	mockConfig := &config.Config{Environment: "development"}

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig)

	ctx := context.Background()
	workspaceID := "test-workspace"
	userID := "test-user"
	newUserID := "new-user"

	t.Run("successful addition", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(&domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}, nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == newUserID && uw.WorkspaceID == workspaceID && uw.Role == "member"
		})).Return(nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, newUserID, "member")
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(&domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}, nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, newUserID, "member")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not an owner of the workspace")
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(&domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}, nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, newUserID, "invalid-role")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user workspace: role: invalid-role does not validate as in(owner|member)")
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}

func TestWorkspaceService_RemoveUserFromWorkspace(t *testing.T) {
	mockRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := createMockUserService()
	mockAuthService := createMockAuthService()
	mockMailer := &MockMailer{}
	mockConfig := &config.Config{Environment: "development"}

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig)

	ctx := context.Background()
	workspaceID := "test-workspace"
	userID := "test-user"
	targetUserID := "target-user"

	t.Run("successful removal", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("RemoveUserFromWorkspace", ctx, targetUserID, workspaceID).Return(nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, targetUserID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, targetUserID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("cannot remove self", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove yourself")
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}

func TestWorkspaceService_TransferOwnership(t *testing.T) {
	mockRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := createMockUserService()
	mockAuthService := createMockAuthService()
	mockMailer := &MockMailer{}
	mockConfig := &config.Config{Environment: "development"}

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig)

	ctx := context.Background()
	workspaceID := "test-workspace"
	currentOwnerID := "current-owner"
	newOwnerID := "new-owner"

	t.Run("successful transfer", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: currentOwnerID,
		}

		expectedCurrentUserWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		expectedNewOwnerWorkspace := &domain.UserWorkspace{
			UserID:      newOwnerID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(expectedCurrentUserWorkspace, nil)
		mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(expectedNewOwnerWorkspace, nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == newOwnerID && uw.WorkspaceID == workspaceID && uw.Role == "owner"
		})).Return(nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == currentOwnerID && uw.WorkspaceID == workspaceID && uw.Role == "member"
		})).Return(nil)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("current owner is not an owner", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: currentOwnerID,
		}

		expectedCurrentUserWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(expectedCurrentUserWorkspace, nil)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not an owner of the workspace")
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("new owner is not a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: currentOwnerID,
		}

		expectedCurrentUserWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(expectedCurrentUserWorkspace, nil)
		mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(nil, fmt.Errorf("user is not a member of the workspace"))

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not a member of the workspace")
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}

func TestWorkspaceService_InviteMember(t *testing.T) {
	mockRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := createMockUserService()
	mockAuthService := createMockAuthService()
	mockMailer := &MockMailer{}
	mockConfig := &config.Config{Environment: "development"}

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig)

	ctx := context.Background()
	workspaceID := "test-workspace"
	userID := "test-user"
	email := "test@example.com"

	t.Run("successful invitation", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockUserService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "inviter@example.com",
		}

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, userID, workspaceID).Return(true, nil)
		mockUserService.On("GetUserByID", ctx, userID).Return(expectedUser, nil)
		mockUserService.On("GetUserByEmail", ctx, email).Return(nil, fmt.Errorf("not found"))
		mockRepo.On("CreateInvitation", ctx, mock.MatchedBy(func(i *domain.WorkspaceInvitation) bool {
			return i.WorkspaceID == workspaceID && i.Email == email && i.InviterID == userID
		})).Return(nil)
		mockAuthService.On("GenerateInvitationToken", mock.AnythingOfType("*domain.WorkspaceInvitation")).Return("test-token")

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, "test-token", token)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid email format", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockUserService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "inviter@example.com",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, "invalid-email")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockUserService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "inviter@example.com",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(nil, fmt.Errorf("workspace not found"))

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "workspace not found")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})

	t.Run("inviter not a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockUserService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "inviter@example.com",
		}

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, userID, workspaceID).Return(false, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "inviter is not a member")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})

	t.Run("user already exists and is a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockUserService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "inviter@example.com",
		}

		existingUser := &domain.User{
			ID:    "existing-user",
			Name:  "Existing User",
			Email: email,
		}

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, userID, workspaceID).Return(true, nil)
		mockUserService.On("GetUserByID", ctx, userID).Return(expectedUser, nil)
		mockUserService.On("GetUserByEmail", ctx, email).Return(existingUser, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, existingUser.ID, workspaceID).Return(true, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is already a member")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})

	t.Run("user already exists but not a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockUserService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "inviter@example.com",
		}

		existingUser := &domain.User{
			ID:    "existing-user",
			Name:  "Existing User",
			Email: email,
		}

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, userID, workspaceID).Return(true, nil)
		mockUserService.On("GetUserByID", ctx, userID).Return(expectedUser, nil)
		mockUserService.On("GetUserByEmail", ctx, email).Return(existingUser, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, existingUser.ID, workspaceID).Return(false, nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == existingUser.ID && uw.WorkspaceID == workspaceID && uw.Role == "member"
		})).Return(nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.NoError(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})
}

func TestWorkspaceService_GetWorkspaceMembersWithEmail(t *testing.T) {
	mockRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := new(MockUserService)
	mockAuthService := new(MockAuthService)
	mockMailer := &MockMailer{}
	mockConfig := &config.Config{Environment: "development"}

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig)

	ctx := context.Background()
	workspaceID := "test-workspace"
	userID := "test-user"

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		expectedMembers := []*domain.UserWorkspaceWithEmail{
			{
				UserWorkspace: *userWorkspace,
				Email:         "test@example.com",
			},
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("GetWorkspaceUsersWithEmail", ctx, workspaceID).Return(expectedMembers, nil)

		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID)
		require.NoError(t, err)
		assert.Equal(t, expectedMembers, members)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(nil, fmt.Errorf("user is not a member of the workspace"))

		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Equal(t, "You do not have access to this workspace", err.Error())
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("GetWorkspaceUsersWithEmail", ctx, workspaceID).Return(nil, fmt.Errorf("repository error"))

		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Equal(t, "repository error", err.Error())
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}
