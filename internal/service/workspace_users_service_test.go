package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	domainmocks "github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
)

func TestWorkspaceService_AddUserToWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := domainmocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	cfg := &config.Config{}

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

	ctx := context.Background()
	workspaceID := "workspace1"
	userID := "user1"
	requesterID := "requester1"

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	t.Run("successful_add_user_to_workspace", func(t *testing.T) {
		// Set up mock expectations
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			AddUserToWorkspace(gomock.Any(), gomock.Any()).
			Return(nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, "member")
		require.NoError(t, err)
	})

	t.Run("authentication_error", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(nil, fmt.Errorf("authentication failed"))

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, "member")
		require.Error(t, err)
		assert.Equal(t, "failed to authenticate user: authentication failed", err.Error())
	})

	t.Run("requester_not_found_in_workspace", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(nil, fmt.Errorf("user workspace not found"))

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, "member")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user workspace not found")
	})

	t.Run("requester_not_an_owner", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "member",
			}, nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, "member")
		require.Error(t, err)
		assert.Equal(t, "user is not an owner of the workspace", err.Error())
	})

	t.Run("invalid_role", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, "invalid_role")
		require.Error(t, err)
		assert.Equal(t, "invalid user workspace: role must be either 'owner' or 'member'", err.Error())
	})
}

func TestWorkspaceService_RemoveUserFromWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	cfg := &config.Config{}

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "workspace1"
	userID := "user1"
	requesterID := "requester1"

	t.Run("successful_remove_user_from_workspace", func(t *testing.T) {
		// Set up mock expectations
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			RemoveUserFromWorkspace(ctx, userID, workspaceID).
			Return(nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID)
		require.NoError(t, err)
	})

	t.Run("authentication_error", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(nil, fmt.Errorf("authentication failed"))

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.Equal(t, "failed to authenticate user: authentication failed", err.Error())
	})

	t.Run("requester_not_found_in_workspace", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(nil, fmt.Errorf("user is not a member of the workspace"))

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not a member of the workspace")
	})

	t.Run("requester_not_an_owner", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "member",
			}, nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.Equal(t, "user is not an owner of the workspace", err.Error())
	})

	t.Run("target_user_not_found", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			RemoveUserFromWorkspace(ctx, userID, workspaceID).
			Return(fmt.Errorf("user is not a member of the workspace"))

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is not a member of the workspace")
	})

	t.Run("cannot_remove_owner", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, requesterID)
		require.Error(t, err)
		assert.Equal(t, "cannot remove yourself from the workspace", err.Error())
	})

	t.Run("cannot remove self", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, requesterID)
		require.Error(t, err)
		assert.Equal(t, "cannot remove yourself from the workspace", err.Error())
	})
}

func TestWorkspaceService_TransferOwnership(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	cfg := &config.Config{}

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

	ctx := context.Background()
	workspaceID := "workspace1"
	userID := "user1"
	requesterID := "requester1"

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	t.Run("successful transfer ownership", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, userID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      userID,
				WorkspaceID: workspaceID,
				Role:        "member",
			}, nil)

		mockRepo.EXPECT().
			AddUserToWorkspace(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, uw *domain.UserWorkspace) error {
				assert.Equal(t, userID, uw.UserID)
				assert.Equal(t, workspaceID, uw.WorkspaceID)
				assert.Equal(t, "owner", uw.Role)
				return nil
			})

		mockRepo.EXPECT().
			AddUserToWorkspace(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, uw *domain.UserWorkspace) error {
				assert.Equal(t, requesterID, uw.UserID)
				assert.Equal(t, workspaceID, uw.WorkspaceID)
				assert.Equal(t, "member", uw.Role)
				return nil
			})

		err := service.TransferOwnership(ctx, workspaceID, userID, requesterID)
		require.NoError(t, err)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(nil, fmt.Errorf("authentication failed"))

		err := service.TransferOwnership(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.Equal(t, "failed to authenticate user: authentication failed", err.Error())
	})

	t.Run("requester not found in workspace", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(nil, fmt.Errorf("user workspace not found"))

		err := service.TransferOwnership(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.Equal(t, "user workspace not found", err.Error())
	})

	t.Run("requester not an owner", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "member",
			}, nil)

		err := service.TransferOwnership(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		assert.Equal(t, "user is not an owner of the workspace", err.Error())
	})

	t.Run("target user not found in workspace", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, userID, workspaceID).
			Return(nil, fmt.Errorf("user workspace not found"))

		err := service.TransferOwnership(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.Equal(t, "user workspace not found", err.Error())
	})

	t.Run("target user is already an owner", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: requesterID}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, requesterID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      requesterID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, userID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      userID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		err := service.TransferOwnership(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.Equal(t, "new owner must be a current member of the workspace", err.Error())
	})
}

func TestWorkspaceService_GetWorkspaceMembersWithEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthSvc := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
	cfg := &config.Config{}

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

	ctx := context.Background()
	workspaceID := "workspace1"
	userID := "user1"
	now := time.Now()

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	t.Run("successful get members with email", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedMembers := []*domain.UserWorkspaceWithEmail{
			{
				UserWorkspace: domain.UserWorkspace{
					UserID:      "user1",
					WorkspaceID: workspaceID,
					Role:        "owner",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				Email: "user1@example.com",
			},
			{
				UserWorkspace: domain.UserWorkspace{
					UserID:      "user2",
					WorkspaceID: workspaceID,
					Role:        "member",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				Email: "user2@example.com",
			},
		}

		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(expectedUser, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, userID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      userID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			GetWorkspaceUsersWithEmail(ctx, workspaceID).
			Return(expectedMembers, nil)

		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID)
		require.NoError(t, err)
		assert.Equal(t, expectedMembers, members)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(nil, fmt.Errorf("authentication failed"))

		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Equal(t, "failed to authenticate user: authentication failed", err.Error())
	})

	t.Run("repository error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(expectedUser, nil)

		mockRepo.EXPECT().
			GetUserWorkspace(ctx, userID, workspaceID).
			Return(&domain.UserWorkspace{
				UserID:      userID,
				WorkspaceID: workspaceID,
				Role:        "owner",
			}, nil)

		mockRepo.EXPECT().
			GetWorkspaceUsersWithEmail(ctx, workspaceID).
			Return(nil, fmt.Errorf("database error"))

		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Equal(t, "database error", err.Error())
	})
}

func TestWorkspaceService_InviteMember(t *testing.T) {
	ctx := context.Background()
	workspaceID := "workspace1"
	inviterID := "inviter1"
	email := "test@example.com"

	t.Run("successful invitation for new user in production", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockMailer := pkgmocks.NewMockMailer(ctrl)
		mockContactService := mocks.NewMockContactService(ctrl)
		mockListService := mocks.NewMockListService(ctrl)
		mockContactListService := domainmocks.NewMockContactListService(ctrl)
		mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
		cfg := &config.Config{Environment: "production"}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

		// Setup common logger expectations
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Mock inviter authentication
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: inviterID}, nil)

		// Mock workspace existence check
		mockRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&domain.Workspace{
				ID:   workspaceID,
				Name: "Test Workspace",
			}, nil)

		// Mock inviter membership check
		mockRepo.EXPECT().
			IsUserWorkspaceMember(ctx, inviterID, workspaceID).
			Return(true, nil)

		// Mock inviter details
		mockUserSvc.EXPECT().
			GetUserByID(ctx, inviterID).
			Return(&domain.User{
				ID:    inviterID,
				Name:  "Test Inviter",
				Email: "inviter@example.com",
			}, nil)

		// Mock existing user check
		mockUserSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, fmt.Errorf("user not found"))

		// Mock invitation creation
		mockRepo.EXPECT().
			CreateInvitation(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, inv *domain.WorkspaceInvitation) error {
				assert.Equal(t, workspaceID, inv.WorkspaceID)
				assert.Equal(t, inviterID, inv.InviterID)
				assert.Equal(t, email, inv.Email)
				return nil
			})

		// Mock token generation
		mockAuthSvc.EXPECT().
			GenerateInvitationToken(gomock.Any()).
			Return("test-token")

		// Mock sending invitation email
		mockMailer.EXPECT().
			SendWorkspaceInvitation(email, "Test Workspace", "Test Inviter", "test-token").
			Return(nil)

		invitation, _, err := service.InviteMember(ctx, workspaceID, email)
		require.NoError(t, err)
		assert.NotNil(t, invitation)
	})

	t.Run("successful invitation for existing user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockMailer := pkgmocks.NewMockMailer(ctrl)
		mockContactService := mocks.NewMockContactService(ctrl)
		mockListService := mocks.NewMockListService(ctrl)
		mockContactListService := domainmocks.NewMockContactListService(ctrl)
		mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
		cfg := &config.Config{Environment: "development"}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

		// Setup common logger expectations
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		existingUser := &domain.User{
			ID:    "existing-user",
			Email: email,
		}

		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: inviterID}, nil)

		mockRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&domain.Workspace{
				ID:   workspaceID,
				Name: "Test Workspace",
			}, nil)

		mockRepo.EXPECT().
			IsUserWorkspaceMember(ctx, inviterID, workspaceID).
			Return(true, nil)

		mockUserSvc.EXPECT().
			GetUserByID(ctx, inviterID).
			Return(&domain.User{
				ID:    inviterID,
				Name:  "Test Inviter",
				Email: "inviter@example.com",
			}, nil)

		mockUserSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(existingUser, nil)

		mockRepo.EXPECT().
			IsUserWorkspaceMember(ctx, existingUser.ID, workspaceID).
			Return(false, nil)

		mockRepo.EXPECT().
			AddUserToWorkspace(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, uw *domain.UserWorkspace) error {
				assert.Equal(t, existingUser.ID, uw.UserID)
				assert.Equal(t, workspaceID, uw.WorkspaceID)
				assert.Equal(t, "member", uw.Role)
				return nil
			})

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.NoError(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
	})

	t.Run("invalid_email_format", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockMailer := pkgmocks.NewMockMailer(ctrl)
		mockContactService := mocks.NewMockContactService(ctrl)
		mockListService := mocks.NewMockListService(ctrl)
		mockContactListService := domainmocks.NewMockContactListService(ctrl)
		mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
		cfg := &config.Config{Environment: "development"}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

		// Setup common logger expectations
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		// Mock authentication - this should be called before email validation
		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: inviterID}, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, "invalid-email")
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, "invalid email format", err.Error())
	})

	t.Run("authentication error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockMailer := pkgmocks.NewMockMailer(ctrl)
		mockContactService := mocks.NewMockContactService(ctrl)
		mockListService := mocks.NewMockListService(ctrl)
		mockContactListService := domainmocks.NewMockContactListService(ctrl)
		mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
		cfg := &config.Config{Environment: "development"}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

		// Setup common logger expectations
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(nil, fmt.Errorf("authentication failed"))

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, "failed to authenticate user: authentication failed", err.Error())
	})

	t.Run("workspace not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockMailer := pkgmocks.NewMockMailer(ctrl)
		mockContactService := mocks.NewMockContactService(ctrl)
		mockListService := mocks.NewMockListService(ctrl)
		mockContactListService := domainmocks.NewMockContactListService(ctrl)
		mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
		cfg := &config.Config{Environment: "development"}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

		// Setup common logger expectations
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: inviterID}, nil)

		mockRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(nil, fmt.Errorf("workspace not found"))

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, "workspace not found", err.Error())
	})

	t.Run("inviter not a member", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockMailer := pkgmocks.NewMockMailer(ctrl)
		mockContactService := mocks.NewMockContactService(ctrl)
		mockListService := mocks.NewMockListService(ctrl)
		mockContactListService := domainmocks.NewMockContactListService(ctrl)
		mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
		cfg := &config.Config{Environment: "development"}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

		// Setup common logger expectations
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: inviterID}, nil)

		mockRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&domain.Workspace{
				ID:   workspaceID,
				Name: "Test Workspace",
			}, nil)

		mockRepo.EXPECT().
			IsUserWorkspaceMember(ctx, inviterID, workspaceID).
			Return(false, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, "inviter is not a member of the workspace", err.Error())
	})

	t.Run("user already a member", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
		mockLogger := pkgmocks.NewMockLogger(ctrl)
		mockUserSvc := mocks.NewMockUserServiceInterface(ctrl)
		mockAuthSvc := mocks.NewMockAuthService(ctrl)
		mockMailer := pkgmocks.NewMockMailer(ctrl)
		mockContactService := mocks.NewMockContactService(ctrl)
		mockListService := mocks.NewMockListService(ctrl)
		mockContactListService := domainmocks.NewMockContactListService(ctrl)
		mockTemplateService := domainmocks.NewMockTemplateService(ctrl)
		cfg := &config.Config{Environment: "development"}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserSvc, mockAuthSvc, mockMailer, cfg, mockContactService, mockListService, mockContactListService, mockTemplateService, "secret_key")

		// Setup common logger expectations
		mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

		existingUser := &domain.User{
			ID:    "existing-user",
			Email: email,
		}

		mockAuthSvc.EXPECT().
			AuthenticateUserForWorkspace(ctx, workspaceID).
			Return(&domain.User{ID: inviterID}, nil)

		mockRepo.EXPECT().
			GetByID(ctx, workspaceID).
			Return(&domain.Workspace{
				ID:   workspaceID,
				Name: "Test Workspace",
			}, nil)

		mockRepo.EXPECT().
			IsUserWorkspaceMember(ctx, inviterID, workspaceID).
			Return(true, nil)

		mockUserSvc.EXPECT().
			GetUserByID(ctx, inviterID).
			Return(&domain.User{
				ID:    inviterID,
				Name:  "Test Inviter",
				Email: "inviter@example.com",
			}, nil)

		mockUserSvc.EXPECT().
			GetUserByEmail(ctx, email).
			Return(existingUser, nil)

		mockRepo.EXPECT().
			IsUserWorkspaceMember(ctx, existingUser.ID, workspaceID).
			Return(true, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, email)
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, "user is already a member of the workspace", err.Error())
	})
}
