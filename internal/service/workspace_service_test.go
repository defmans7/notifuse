package service

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
)

func TestWorkspaceService_ListWorkspaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

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

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{Environment: "development"}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "testworkspace"
	userID := "testuser"

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

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
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

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
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

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
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

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "testworkspace"

	t.Run("successful creation", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, workspace *domain.Workspace) error {
			// Instead of expecting an exact value, verify it's not empty and has expected format
			assert.NotEmpty(t, workspace.Settings.SecretKey, "Secret key should not be empty")
			assert.Equal(t, 64, len(workspace.Settings.SecretKey), "Secret key should be 64 hex characters (32 bytes)")
			// Verify hex encoding
			_, err := hex.DecodeString(workspace.Settings.SecretKey)
			assert.NoError(t, err, "Secret key should be valid hex")
			return nil
		})
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(nil)
		mockUserService.EXPECT().GetUserByID(ctx, expectedUser.ID).Return(expectedUser, nil)
		mockContactService.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).Return(domain.UpsertContactOperation{Action: domain.UpsertContactOperationCreate})

		// Mock template service expectations for createDefaultTemplates
		mockTemplateService.EXPECT().CreateTemplate(ctx, workspaceID, gomock.Any()).Return(nil).Times(4)

		mockListService.EXPECT().CreateList(ctx, workspaceID, gomock.Any()).Return(nil)
		mockContactListService.EXPECT().AddContactToList(ctx, workspaceID, gomock.Any()).Return(nil)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
		assert.Equal(t, "Test Workspace", workspace.Name)

		// Verify the structure of settings but don't check the exact SecretKey value
		assert.Equal(t, "https://example.com", workspace.Settings.WebsiteURL)
		assert.Equal(t, "https://example.com/logo.png", workspace.Settings.LogoURL)
		assert.Equal(t, "https://example.com/cover.png", workspace.Settings.CoverURL)
		assert.Equal(t, "UTC", workspace.Settings.Timezone)

		// Verify SecretKey format but not exact value
		assert.NotEmpty(t, workspace.Settings.SecretKey)
		assert.Equal(t, 64, len(workspace.Settings.SecretKey))
		_, err = hex.DecodeString(workspace.Settings.SecretKey)
		assert.NoError(t, err, "Secret key should be valid hex")
	})

	t.Run("validation error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		// No need to mock GetByID here as the validation fails before that check

		// Invalid timezone
		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "INVALID_TIMEZONE", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "invalid timezone: INVALID_TIMEZONE")
	})

	t.Run("repository error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("add user error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("get user error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(nil)
		mockUserService.EXPECT().GetUserByID(ctx, expectedUser.ID).Return(nil, assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("upsert contact error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
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

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "failed to upsert contact")
	})

	t.Run("template creation error still allows workspace creation", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, nil)
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRepo.EXPECT().AddUserToWorkspace(ctx, gomock.Any()).Return(nil)
		mockUserService.EXPECT().GetUserByID(ctx, expectedUser.ID).Return(expectedUser, nil)
		mockContactService.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).Return(domain.UpsertContactOperation{Action: domain.UpsertContactOperationCreate})

		// Simulate template creation error for all four templates
		mockTemplateService.EXPECT().CreateTemplate(ctx, workspaceID, gomock.Any()).Return(errors.New("template creation failed")).AnyTimes()

		mockListService.EXPECT().CreateList(ctx, workspaceID, gomock.Any()).Return(nil)
		mockContactListService.EXPECT().AddContactToList(ctx, workspaceID, gomock.Any()).Return(nil)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})

		// Should still succeed despite template error
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
	})

	t.Run("workspace already exists", func(t *testing.T) {
		expectedUser := &domain.User{
			ID:    "testowner",
			Email: "test@example.com",
			Name:  "Test User",
		}

		existingWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.EXPECT().AuthenticateUserFromContext(ctx).Return(expectedUser, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(existingWorkspace, nil)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", domain.FileManagerSettings{
			Endpoint:  "https://s3.amazonaws.com",
			Bucket:    "my-bucket",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
		})
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "workspace already exists")
	})
}

func TestWorkspaceService_UpdateWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "testworkspace"
	userID := "testuser"

	t.Run("successful update", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		settings := domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
			FileManager: domain.FileManagerSettings{
				Endpoint:  "https://s3.amazonaws.com",
				Bucket:    "my-bucket",
				AccessKey: "AKIAIOSFODNN7EXAMPLE",
			},
		}

		existingWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Original Workspace Name",
			Settings: domain.WorkspaceSettings{
				WebsiteURL: "https://old-example.com",
			},
			CreatedAt: time.Now().Add(-24 * time.Hour), // Created a day ago
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		}

		expectedWorkspace := &domain.Workspace{
			ID:        workspaceID,
			Name:      "Updated Workspace",
			Settings:  settings,
			CreatedAt: existingWorkspace.CreatedAt,
			UpdatedAt: time.Now(),
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(existingWorkspace, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", settings)
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspace.ID, workspace.ID)
		assert.Equal(t, expectedWorkspace.Name, workspace.Name)
		assert.Equal(t, expectedWorkspace.Settings, workspace.Settings)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, nil, assert.AnError)

		settings := domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
			FileManager: domain.FileManagerSettings{
				Endpoint:  "https://s3.amazonaws.com",
				Bucket:    "my-bucket",
				AccessKey: "AKIAIOSFODNN7EXAMPLE",
			},
		}

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", settings)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), assert.AnError.Error())
	})

	t.Run("user not workspace owner", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		settings := domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
			FileManager: domain.FileManagerSettings{
				Endpoint:  "https://s3.amazonaws.com",
				Bucket:    "my-bucket",
				AccessKey: "AKIAIOSFODNN7EXAMPLE",
			},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", settings)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
	})

	t.Run("error getting user workspace", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		settings := domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
			FileManager: domain.FileManagerSettings{
				Endpoint:  "https://s3.amazonaws.com",
				Bucket:    "my-bucket",
				AccessKey: "AKIAIOSFODNN7EXAMPLE",
			},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(nil, assert.AnError)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", settings)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestWorkspaceService_DeleteWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{Environment: "development"}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	ctx := context.Background()
	userID := "testuser"
	workspaceID := "testworkspace"

	t.Run("successful delete as owner with no integrations", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Workspace with no integrations
		workspace := &domain.Workspace{
			ID:           workspaceID,
			Name:         "Test Workspace",
			Integrations: []domain.Integration{},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(workspace, nil)
		mockRepo.EXPECT().Delete(ctx, workspaceID).Return(nil)

		err := service.DeleteWorkspace(ctx, workspaceID)
		require.NoError(t, err)
	})

	t.Run("successful delete as owner with integrations", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Workspace with two integrations
		integrations := []domain.Integration{
			{
				ID:   "integration-1",
				Name: "Integration 1",
				Type: domain.IntegrationTypeEmail,
				EmailProvider: domain.EmailProvider{
					Kind: domain.EmailProviderKindSMTP,
				},
			},
			{
				ID:   "integration-2",
				Name: "Integration 2",
				Type: domain.IntegrationTypeEmail,
				EmailProvider: domain.EmailProvider{
					Kind: domain.EmailProviderKindSMTP,
				},
			},
		}

		workspace := &domain.Workspace{
			ID:           workspaceID,
			Name:         "Test Workspace",
			Integrations: integrations,
		}

		// Initial authentication for the DeleteWorkspace itself
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(workspace, nil)

		// For each DeleteIntegration call inside DeleteWorkspace, expect these mocks
		// The DeleteIntegration method will call AuthenticateUserForWorkspace again for each integration
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil).Times(2)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil).Times(2)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(workspace, nil).Times(2)

		// For first integration
		webhookStatus1 := &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      true,
		}
		mockWebhookRegService.EXPECT().GetWebhookStatus(ctx, workspaceID, "integration-1").Return(webhookStatus1, nil)
		mockWebhookRegService.EXPECT().UnregisterWebhooks(ctx, workspaceID, "integration-1").Return(nil)

		// For second integration
		webhookStatus2 := &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      true,
		}
		mockWebhookRegService.EXPECT().GetWebhookStatus(ctx, workspaceID, "integration-2").Return(webhookStatus2, nil)
		mockWebhookRegService.EXPECT().UnregisterWebhooks(ctx, workspaceID, "integration-2").Return(nil)

		// Once for each integration deletion
		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil).Times(2)

		// Final workspace deletion
		mockRepo.EXPECT().Delete(ctx, workspaceID).Return(nil)

		err := service.DeleteWorkspace(ctx, workspaceID)
		require.NoError(t, err)
	})

	t.Run("continues deletion despite integration deletion failure", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Workspace with one integration
		integrations := []domain.Integration{
			{
				ID:   "integration-1",
				Name: "Integration 1",
				Type: domain.IntegrationTypeEmail,
				EmailProvider: domain.EmailProvider{
					Kind: domain.EmailProviderKindSMTP,
				},
			},
		}

		workspace := &domain.Workspace{
			ID:           workspaceID,
			Name:         "Test Workspace",
			Integrations: integrations,
		}

		// Initial authentication for DeleteWorkspace
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(workspace, nil)

		// Authentication for DeleteIntegration
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(workspace, nil)

		// The integration deletion fails
		webhookStatus := &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      true,
		}
		mockWebhookRegService.EXPECT().GetWebhookStatus(ctx, workspaceID, "integration-1").Return(webhookStatus, nil)
		mockWebhookRegService.EXPECT().UnregisterWebhooks(ctx, workspaceID, "integration-1").Return(errors.New("webhook error"))
		// The update fails
		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(errors.New("integration delete error"))

		// Should still proceed with workspace deletion
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

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)

		err := service.DeleteWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
	})

	t.Run("error getting workspace details", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, errors.New("error getting workspace"))

		err := service.DeleteWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error getting workspace")
	})
}

func TestWorkspaceService_CreateIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "testworkspace"
	userID := "testuser"
	integrationName := "Test SMTP Integration"

	provider := domain.EmailProvider{
		Kind: domain.EmailProviderKindSMTP,
		SMTP: &domain.SMTPSettings{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "smtp_user",
			Password: "smtp_password",
			UseTLS:   true,
		},
		DefaultSenderEmail: "test@example.com",
		DefaultSenderName:  "Test Sender",
	}

	t.Run("successful create integration", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, workspace *domain.Workspace) error {
			// Verify the integration was added to the workspace
			require.Equal(t, 1, len(workspace.Integrations))
			require.Equal(t, integrationName, workspace.Integrations[0].Name)
			require.Equal(t, domain.IntegrationTypeEmail, workspace.Integrations[0].Type)
			require.Equal(t, domain.EmailProviderKindSMTP, workspace.Integrations[0].EmailProvider.Kind)
			return nil
		})

		// Expect webhook registration call for email integration
		mockConfig.APIEndpoint = "https://api.example.com"
		// Webhook config is provided for reference only, we use gomock.Any() since ID is random
		_ = &domain.WebhookRegistrationConfig{
			IntegrationID: "integration123", // This will be a random UUID, so use Any matcher
			EventTypes: []domain.EmailEventType{
				domain.EmailEventDelivered,
				domain.EmailEventBounce,
				domain.EmailEventComplaint,
			},
		}
		mockWebhookRegService.EXPECT().RegisterWebhooks(
			ctx,
			workspaceID,
			gomock.Any(), // Use Any for the config since integrationID is random
		).Return(&domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      true,
		}, nil)

		integrationID, err := service.CreateIntegration(ctx, workspaceID, integrationName, domain.IntegrationTypeEmail, provider)
		require.NoError(t, err)
		require.NotEmpty(t, integrationID)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		// User is a member, not an owner
		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)

		integrationID, err := service.CreateIntegration(ctx, workspaceID, integrationName, domain.IntegrationTypeEmail, provider)
		require.Error(t, err)
		require.Empty(t, integrationID)
		require.IsType(t, &domain.ErrUnauthorized{}, err)
	})

	t.Run("workspace not found", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(nil, errors.New("workspace not found"))

		integrationID, err := service.CreateIntegration(ctx, workspaceID, integrationName, domain.IntegrationTypeEmail, provider)
		require.Error(t, err)
		require.Empty(t, integrationID)
		require.Contains(t, err.Error(), "workspace not found")
	})

	t.Run("update error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(errors.New("update error"))

		integrationID, err := service.CreateIntegration(ctx, workspaceID, integrationName, domain.IntegrationTypeEmail, provider)
		require.Error(t, err)
		require.Empty(t, integrationID)
		require.Contains(t, err.Error(), "update error")
	})
}

func TestWorkspaceService_UpdateIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "testworkspace"
	userID := "testuser"
	integrationID := "integration123"
	integrationName := "Updated SMTP Integration"

	provider := domain.EmailProvider{
		Kind: domain.EmailProviderKindSMTP,
		SMTP: &domain.SMTPSettings{
			Host:     "smtp.updated.com",
			Port:     587,
			Username: "updated_user",
			Password: "updated_password",
			UseTLS:   true,
		},
		DefaultSenderEmail: "updated@example.com",
		DefaultSenderName:  "Updated Sender",
	}

	t.Run("successful update integration", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Create a workspace with an existing integration
		existingIntegration := domain.Integration{
			ID:   integrationID,
			Name: "Original SMTP Integration",
			Type: domain.IntegrationTypeEmail,
			EmailProvider: domain.EmailProvider{
				Kind: domain.EmailProviderKindSMTP,
				SMTP: &domain.SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "smtp_user",
					Password: "smtp_password",
					UseTLS:   true,
				},
				DefaultSenderEmail: "test@example.com",
				DefaultSenderName:  "Test Sender",
			},
			CreatedAt: time.Now().Add(-24 * time.Hour), // Created 24 hours ago
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		}

		expectedWorkspace := &domain.Workspace{
			ID:           workspaceID,
			Name:         "Test Workspace",
			Integrations: []domain.Integration{existingIntegration},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, workspace *domain.Workspace) error {
			// Verify the integration was updated in the workspace
			require.Equal(t, 1, len(workspace.Integrations))
			require.Equal(t, integrationID, workspace.Integrations[0].ID)
			require.Equal(t, integrationName, workspace.Integrations[0].Name)
			require.Equal(t, domain.IntegrationTypeEmail, workspace.Integrations[0].Type)
			require.Equal(t, domain.EmailProviderKindSMTP, workspace.Integrations[0].EmailProvider.Kind)
			require.Equal(t, "smtp.updated.com", workspace.Integrations[0].EmailProvider.SMTP.Host)
			require.Equal(t, "updated_user", workspace.Integrations[0].EmailProvider.SMTP.Username)
			require.Equal(t, existingIntegration.CreatedAt, workspace.Integrations[0].CreatedAt)      // CreatedAt should remain the same
			require.True(t, workspace.Integrations[0].UpdatedAt.After(existingIntegration.UpdatedAt)) // UpdatedAt should be updated
			return nil
		})

		err := service.UpdateIntegration(ctx, workspaceID, integrationID, integrationName, provider)
		require.NoError(t, err)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		// User is a member, not an owner
		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)

		err := service.UpdateIntegration(ctx, workspaceID, integrationID, integrationName, provider)
		require.Error(t, err)
		require.IsType(t, &domain.ErrUnauthorized{}, err)
	})

	t.Run("integration not found", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Create a workspace with no integrations
		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)

		err := service.UpdateIntegration(ctx, workspaceID, integrationID, integrationName, provider)
		require.Error(t, err)
		require.Contains(t, err.Error(), "integration not found")
	})
}

func TestWorkspaceService_DeleteIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	mockConfig := &config.Config{}
	mockContactService := mocks.NewMockContactService(ctrl)
	mockListService := mocks.NewMockListService(ctrl)
	mockContactListService := mocks.NewMockContactListService(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockWebhookRegService := mocks.NewMockWebhookRegistrationService(ctrl)

	service := NewWorkspaceService(
		mockRepo,
		mockUserRepo,
		mockLogger,
		mockUserService,
		mockAuthService,
		mockMailer,
		mockConfig,
		mockContactService,
		mockListService,
		mockContactListService,
		mockTemplateService,
		mockWebhookRegService,
		"secret_key",
	)

	// Set up mockLogger to allow any calls
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	ctx := context.Background()
	workspaceID := "testworkspace"
	userID := "testuser"
	integrationID := "integration123"

	t.Run("successful delete integration", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Create a workspace with an existing integration
		existingIntegration := domain.Integration{
			ID:   integrationID,
			Name: "SMTP Integration",
			Type: domain.IntegrationTypeEmail,
			EmailProvider: domain.EmailProvider{
				Kind: domain.EmailProviderKindSMTP,
				SMTP: &domain.SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "smtp_user",
					Password: "smtp_password",
					UseTLS:   true,
				},
			},
		}

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				TransactionalEmailProviderID: integrationID, // Reference the integration
			},
			Integrations: []domain.Integration{existingIntegration},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)

		// Expect webhook status check
		webhookStatus := &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      true,
			Endpoints: []domain.WebhookEndpointStatus{
				{
					URL:    "https://api.example.com/webhooks",
					Active: true,
				},
			},
		}
		mockWebhookRegService.EXPECT().GetWebhookStatus(ctx, workspaceID, integrationID).Return(webhookStatus, nil)

		// Expect webhook unregistration
		mockWebhookRegService.EXPECT().UnregisterWebhooks(ctx, workspaceID, integrationID).Return(nil)

		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, workspace *domain.Workspace) error {
			// Verify the integration was removed from the workspace
			require.Empty(t, workspace.Integrations)
			// Verify the reference was removed from settings
			require.Empty(t, workspace.Settings.TransactionalEmailProviderID)
			return nil
		})

		err := service.DeleteIntegration(ctx, workspaceID, integrationID)
		require.NoError(t, err)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		// User is a member, not an owner
		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)

		err := service.DeleteIntegration(ctx, workspaceID, integrationID)
		require.Error(t, err)
		require.IsType(t, &domain.ErrUnauthorized{}, err)
	})

	t.Run("integration not found", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Create a workspace with no integrations
		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)

		err := service.DeleteIntegration(ctx, workspaceID, integrationID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "integration not found")
	})

	t.Run("webhook unregistration error", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Create a workspace with an existing integration
		existingIntegration := domain.Integration{
			ID:   integrationID,
			Name: "SMTP Integration",
			Type: domain.IntegrationTypeEmail,
			EmailProvider: domain.EmailProvider{
				Kind: domain.EmailProviderKindSMTP,
				SMTP: &domain.SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "smtp_user",
					Password: "smtp_password",
					UseTLS:   true,
				},
			},
		}

		expectedWorkspace := &domain.Workspace{
			ID:           workspaceID,
			Name:         "Test Workspace",
			Integrations: []domain.Integration{existingIntegration},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)

		// Expect webhook status check
		mockWebhookRegService.EXPECT().GetWebhookStatus(ctx, workspaceID, integrationID).Return(&domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      true,
			Endpoints: []domain.WebhookEndpointStatus{
				{
					URL:    "https://api.example.com/webhooks",
					Active: true,
				},
			},
		}, nil)

		// Skip logger checks

		// The unregistration fails
		webhookError := errors.New("failed to unregister webhooks")
		mockWebhookRegService.EXPECT().UnregisterWebhooks(ctx, workspaceID, integrationID).Return(webhookError)

		// Skip logger checks

		mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		err := service.DeleteIntegration(ctx, workspaceID, integrationID)
		require.NoError(t, err) // Should still succeed despite webhook unregistration error
	})

	t.Run("removes marketing reference", func(t *testing.T) {
		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Create a workspace with an existing integration
		existingIntegration := domain.Integration{
			ID:   integrationID,
			Name: "SMTP Integration",
			Type: domain.IntegrationTypeEmail,
			EmailProvider: domain.EmailProvider{
				Kind: domain.EmailProviderKindSMTP,
			},
		}

		expectedWorkspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				MarketingEmailProviderID: integrationID, // Reference the integration as marketing provider
			},
			Integrations: []domain.Integration{existingIntegration},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(ctx, expectedUser, nil)
		mockRepo.EXPECT().GetUserWorkspace(ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.EXPECT().GetByID(ctx, workspaceID).Return(expectedWorkspace, nil)

		// Expect webhook status check
		mockWebhookRegService.EXPECT().GetWebhookStatus(ctx, workspaceID, integrationID).Return(&domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      false, // Not registered
		}, nil)

		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, workspace *domain.Workspace) error {
			// Verify the reference was removed from settings
			require.Empty(t, workspace.Settings.MarketingEmailProviderID)
			return nil
		})

		err := service.DeleteIntegration(ctx, workspaceID, integrationID)
		require.NoError(t, err)
	})
}

func TestGenerateSecureKey(t *testing.T) {
	t.Run("generates key of expected length", func(t *testing.T) {
		// Test with different byte lengths
		byteLengths := []int{16, 32, 64}

		for _, byteLen := range byteLengths {
			// Each byte becomes 2 hex chars
			expectedHexLen := byteLen * 2

			// Generate the key
			key, err := GenerateSecureKey(byteLen)

			// Verify results
			require.NoError(t, err)
			assert.Len(t, key, expectedHexLen)

			// Verify it's valid hex
			_, err = hex.DecodeString(key)
			require.NoError(t, err, "Generated key is not valid hex")
		}
	})

	t.Run("generates unique keys", func(t *testing.T) {
		// Generate multiple keys to ensure uniqueness
		iterations := 10
		keys := make([]string, iterations)

		for i := 0; i < iterations; i++ {
			key, err := GenerateSecureKey(32)
			require.NoError(t, err)
			keys[i] = key
		}

		// Check for duplicates
		seen := make(map[string]bool)
		for _, key := range keys {
			assert.False(t, seen[key], "Duplicate key generated")
			seen[key] = true
		}
	})
}
