package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	domainmocks "github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
)

func TestWorkspaceService_ListWorkspaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := domainmocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{}
	mockContactService := domainmocks.NewMockContactService(ctrl)
	mockListService := domainmocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig, mockContactService, mockListService, mockContactListService)

	ctx := context.Background()
	user := &domain.User{ID: "test-user"}

	t.Run("successful list with workspaces", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(user, nil)
		mockRepo.EXPECT().GetUserWorkspaces(ctx, user.ID).Return([]*domain.UserWorkspace{
			{WorkspaceID: "1"},
			{WorkspaceID: "2"},
		}, nil)
		mockRepo.EXPECT().GetByID(ctx, "1").Return(&domain.Workspace{ID: "1"}, nil)
		mockRepo.EXPECT().GetByID(ctx, "2").Return(&domain.Workspace{ID: "2"}, nil)

		workspaces, err := service.ListWorkspaces(ctx)
		assert.NoError(t, err)
		assert.Len(t, workspaces, 2)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(nil, errors.New("auth error"))

		workspaces, err := service.ListWorkspaces(ctx)
		assert.Error(t, err)
		assert.Nil(t, workspaces)
	})

	t.Run("get user workspaces error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(user, nil)
		mockRepo.EXPECT().GetUserWorkspaces(ctx, user.ID).Return(nil, errors.New("repo error"))
		mockLogger.EXPECT().WithField("user_id", user.ID).Return(mockLogger)
		mockLogger.EXPECT().WithField("error", "repo error").Return(mockLogger)
		mockLogger.EXPECT().Error("Failed to get user workspaces")

		workspaces, err := service.ListWorkspaces(ctx)
		assert.Error(t, err)
		assert.Nil(t, workspaces)
	})

	t.Run("get workspace by ID error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(user, nil)
		mockRepo.EXPECT().GetUserWorkspaces(ctx, user.ID).Return([]*domain.UserWorkspace{
			{WorkspaceID: "1"},
		}, nil)
		mockRepo.EXPECT().GetByID(ctx, "1").Return(nil, errors.New("repo error"))
		mockLogger.EXPECT().WithField("workspace_id", "1").Return(mockLogger)
		mockLogger.EXPECT().WithField("user_id", user.ID).Return(mockLogger)
		mockLogger.EXPECT().WithField("error", "repo error").Return(mockLogger)
		mockLogger.EXPECT().Error("Failed to get workspace by ID")

		workspaces, err := service.ListWorkspaces(ctx)
		assert.Error(t, err)
		assert.Nil(t, workspaces)
	})

	t.Run("no workspaces", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(user, nil)
		mockRepo.EXPECT().GetUserWorkspaces(ctx, user.ID).Return([]*domain.UserWorkspace{}, nil)

		workspaces, err := service.ListWorkspaces(ctx)
		assert.NoError(t, err)
		assert.Empty(t, workspaces)
	})
}

func TestWorkspaceService_GetWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := domainmocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{Environment: "development"}
	mockContactService := domainmocks.NewMockContactService(ctrl)
	mockListService := domainmocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig, mockContactService, mockListService, mockContactListService)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "test-workspace"
	userID := "test-user"

	t.Run("successful get", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)

		workspace, err := service.GetWorkspace(ctx, workspaceID)
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspace, workspace)
	})

	t.Run("workspace not found", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(nil, assert.AnError)

		workspace, err := service.GetWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("error getting workspace by ID", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, assert.AnError)

		workspace, err := service.GetWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestWorkspaceService_CreateWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := domainmocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{Environment: "development"}
	mockContactService := domainmocks.NewMockContactService(ctrl)
	mockListService := domainmocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig, mockContactService, mockListService, mockContactListService)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "testworkspace1"

	t.Run("successful creation", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "test-owner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				WebsiteURL: "https://example.com",
				LogoURL:    "https://example.com/logo.png",
				CoverURL:   "https://example.com/cover.png",
				Timezone:   "UTC",
			},
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(nil)
		mockUserService.EXPECT().GetUserByID(ctx, expectedUser.ID).Return(expectedUser, nil)
		mockContactService.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).Return(domain.UpsertContactOperation{Action: domain.UpsertContactOperationCreate})
		mockListService.EXPECT().CreateList(ctx, workspaceID, gomock.Any()).Return(nil)
		mockContactListService.EXPECT().AddContactToList(ctx, workspaceID, gomock.Any()).Return(nil)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspace.ID, workspace.ID)
		assert.Equal(t, expectedWorkspace.Name, workspace.Name)
		assert.Equal(t, expectedWorkspace.Settings, workspace.Settings)
	})

	t.Run("validation error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "test-owner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		// No need to mock GetByID here as the validation fails before that check

		// Invalid timezone
		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "INVALID_TIMEZONE")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "does not validate as timezone")
	})

	t.Run("repository error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "test-owner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("add user error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "test-owner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("get user error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "test-owner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(nil)
		mockUserService.EXPECT().GetUserByID(ctx, expectedUser.ID).Return(nil, assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("upsert contact error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "test-owner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(nil)
		mockUserService.EXPECT().GetUserByID(ctx, expectedUser.ID).Return(expectedUser, nil)
		mockContactService.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).Return(domain.UpsertContactOperation{
			Action: domain.UpsertContactOperationError,
			Error:  "failed to upsert contact",
		})

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "failed to upsert contact")
	})

	t.Run("workspace already exists", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "test-owner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		existingWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(existingWorkspace, nil)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "workspace already exists")
	})
}

func TestWorkspaceService_UpdateWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := domainmocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{Environment: "development"}
	mockContactService := domainmocks.NewMockContactService(ctrl)
	mockListService := domainmocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig, mockContactService, mockListService, mockContactListService)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	userID := "test-user"
	workspaceID := "testworkspace1"

	t.Run("successful update as owner", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Updated Workspace",
			Settings: domain.WorkspaceSettings{
				WebsiteURL: "https://updated.com",
				LogoURL:    "https://updated.com/logo.png",
				CoverURL:   "https://updated.com/cover.png",
				Timezone:   "Europe/Paris",
			},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().Update(ctx, gomock.AssignableToTypeOf(&domain.Workspace{})).DoAndReturn(func(_ context.Context, w *domain.Workspace) error {
			assert.Equal(t, expectedWorkspace.ID, w.ID)
			assert.Equal(t, expectedWorkspace.Name, w.Name)
			assert.Equal(t, expectedWorkspace.Settings, w.Settings)
			return nil
		})

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris")
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
		assert.Equal(t, "Updated Workspace", workspace.Name)
		assert.Equal(t, "https://updated.com", workspace.Settings.WebsiteURL)
		assert.Equal(t, "https://updated.com/logo.png", workspace.Settings.LogoURL)
		assert.Equal(t, "https://updated.com/cover.png", workspace.Settings.CoverURL)
		assert.Equal(t, "Europe/Paris", workspace.Settings.Timezone)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
	})

	t.Run("validation error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)

		// Invalid timezone
		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "INVALID_TIMEZONE")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "does not validate as timezone")
	})

	t.Run("repository error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(assert.AnError)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestWorkspaceService_DeleteWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := domainmocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := domainmocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := domainmocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{Environment: "development"}
	mockContactService := domainmocks.NewMockContactService(ctrl)
	mockListService := domainmocks.NewMockListService(ctrl)
	mockContactListService := domainmocks.NewMockContactListService(ctrl)
	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, mockConfig, mockContactService, mockListService, mockContactListService)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	userID := "test-user"
	workspaceID := "test-workspace"

	t.Run("successful delete as owner", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().Delete(ctx, workspaceID).Return(nil)

		err := service.DeleteWorkspace(ctx, workspaceID)
		require.NoError(t, err)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)

		err := service.DeleteWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
	})
}
