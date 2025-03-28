package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWorkspaceRepo is a mock implementation of domain.WorkspaceRepository
type MockWorkspaceRepo struct {
	CreateFn                  func(ctx context.Context, workspace *domain.Workspace) error
	GetByIDFn                 func(ctx context.Context, id string) (*domain.Workspace, error)
	ListFn                    func(ctx context.Context) ([]*domain.Workspace, error)
	UpdateFn                  func(ctx context.Context, workspace *domain.Workspace) error
	DeleteFn                  func(ctx context.Context, id string) error
	AddUserToWorkspaceFn      func(ctx context.Context, userWorkspace *domain.UserWorkspace) error
	RemoveUserFromWorkspaceFn func(ctx context.Context, userID, workspaceID string) error
	GetUserWorkspacesFn       func(ctx context.Context, userID string) ([]*domain.UserWorkspace, error)

	GetUserWorkspaceFn      func(ctx context.Context, userID, workspaceID string) (*domain.UserWorkspace, error)
	GetConnectionFn         func(ctx context.Context, workspaceID string) (*sql.DB, error)
	CreateDatabaseFn        func(ctx context.Context, workspaceID string) error
	DeleteDatabaseFn        func(ctx context.Context, workspaceID string) error
	CreateInvitationFn      func(ctx context.Context, invitation *domain.WorkspaceInvitation) error
	GetInvitationByIDFn     func(ctx context.Context, id string) (*domain.WorkspaceInvitation, error)
	GetInvitationByEmailFn  func(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, error)
	IsUserWorkspaceMemberFn func(ctx context.Context, userID, workspaceID string) (bool, error)
}

// Implement methods for the MockWorkspaceRepo
func (m *MockWorkspaceRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	return m.CreateFn(ctx, workspace)
}

func (m *MockWorkspaceRepo) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	return m.GetByIDFn(ctx, id)
}

func (m *MockWorkspaceRepo) List(ctx context.Context) ([]*domain.Workspace, error) {
	return m.ListFn(ctx)
}

func (m *MockWorkspaceRepo) Update(ctx context.Context, workspace *domain.Workspace) error {
	return m.UpdateFn(ctx, workspace)
}

func (m *MockWorkspaceRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockWorkspaceRepo) AddUserToWorkspace(ctx context.Context, userWorkspace *domain.UserWorkspace) error {
	return m.AddUserToWorkspaceFn(ctx, userWorkspace)
}

func (m *MockWorkspaceRepo) RemoveUserFromWorkspace(ctx context.Context, userID string, workspaceID string) error {
	return m.RemoveUserFromWorkspaceFn(ctx, userID, workspaceID)
}

func (m *MockWorkspaceRepo) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.UserWorkspace, error) {
	return m.GetUserWorkspacesFn(ctx, userID)
}

func (m *MockWorkspaceRepo) GetUserWorkspace(ctx context.Context, userID string, workspaceID string) (*domain.UserWorkspace, error) {
	return m.GetUserWorkspaceFn(ctx, userID, workspaceID)
}

func (m *MockWorkspaceRepo) GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error) {
	return m.GetConnectionFn(ctx, workspaceID)
}

func (m *MockWorkspaceRepo) CreateDatabase(ctx context.Context, workspaceID string) error {
	return m.CreateDatabaseFn(ctx, workspaceID)
}

func (m *MockWorkspaceRepo) DeleteDatabase(ctx context.Context, workspaceID string) error {
	return m.DeleteDatabaseFn(ctx, workspaceID)
}

func (m *MockWorkspaceRepo) CreateInvitation(ctx context.Context, invitation *domain.WorkspaceInvitation) error {
	return m.CreateInvitationFn(ctx, invitation)
}

func (m *MockWorkspaceRepo) GetInvitationByID(ctx context.Context, id string) (*domain.WorkspaceInvitation, error) {
	return m.GetInvitationByIDFn(ctx, id)
}

func (m *MockWorkspaceRepo) GetInvitationByEmail(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, error) {
	return m.GetInvitationByEmailFn(ctx, workspaceID, email)
}

func (m *MockWorkspaceRepo) IsUserWorkspaceMember(ctx context.Context, userID, workspaceID string) (bool, error) {
	return m.IsUserWorkspaceMemberFn(ctx, userID, workspaceID)
}

func (m *MockWorkspaceRepo) GetWorkspaceUsersWithEmail(ctx context.Context, workspaceID string) ([]*domain.UserWorkspaceWithEmail, error) {
	return nil, nil // Not used in this test file
}

// MockMailer mocks the mailer.Mailer interface
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) SendInvitationEmail(invitation *domain.WorkspaceInvitation) error {
	args := m.Called(invitation)
	return args.Error(0)
}

func (m *MockMailer) SendMagicCode(email, code string) error {
	args := m.Called(email, code)
	return args.Error(0)
}

func (m *MockMailer) SendWorkspaceInvitation(email, workspaceName, inviterName, token string) error {
	args := m.Called(email, workspaceName, inviterName, token)
	return args.Error(0)
}

// Create a mock UserService for testing
func createMockUserService() *UserService {
	mockRepo := new(mockUserRepository)
	mockLogger := new(MockLogger)

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	// Create a key for AuthService
	key := paseto.NewV4AsymmetricSecretKey()

	authService := &AuthService{
		privateKey: key,
	}

	return &UserService{
		repo:        mockRepo,
		authService: authService,
		logger:      mockLogger,
	}
}

// Create a mock AuthService for testing
func createMockAuthService() *MockAuthService {
	mockAuthService := new(MockAuthService)
	return mockAuthService
}

func TestWorkspaceService_ListWorkspaces(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
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
	userID := "test-user"

	t.Run("successful list with workspaces", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspaces := []*domain.UserWorkspace{
			{
				UserID:      userID,
				WorkspaceID: "1",
				Role:        "owner",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				UserID:      userID,
				WorkspaceID: "2",
				Role:        "member",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		expectedWorkspaces := []*domain.Workspace{
			{
				ID:   "1",
				Name: "Test Workspace 1",
				Settings: domain.WorkspaceSettings{
					WebsiteURL: "https://example.com",
					LogoURL:    "https://example.com/logo.png",
					Timezone:   "UTC",
				},
			},
			{
				ID:   "2",
				Name: "Test Workspace 2",
				Settings: domain.WorkspaceSettings{
					WebsiteURL: "https://example2.com",
					LogoURL:    "https://example2.com/logo.png",
					Timezone:   "UTC",
				},
			},
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspaces", ctx, userID).Return(expectedUserWorkspaces, nil)
		mockRepo.On("GetByID", ctx, "1").Return(expectedWorkspaces[0], nil)
		mockRepo.On("GetByID", ctx, "2").Return(expectedWorkspaces[1], nil)

		workspaces, err := service.ListWorkspaces(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspaces, workspaces)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("empty list when user has no workspaces", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspaces", ctx, userID).Return([]*domain.UserWorkspace{}, nil)

		workspaces, err := service.ListWorkspaces(ctx)
		require.NoError(t, err)
		assert.Empty(t, workspaces)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("error getting user workspaces", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspaces", ctx, userID).Return(nil, assert.AnError)

		workspaces, err := service.ListWorkspaces(ctx)
		require.Error(t, err)
		assert.Nil(t, workspaces)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("error getting a specific workspace", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		expectedUserWorkspaces := []*domain.UserWorkspace{
			{
				UserID:      userID,
				WorkspaceID: "1",
				Role:        "owner",
			},
			{
				UserID:      userID,
				WorkspaceID: "2",
				Role:        "member",
			},
		}

		// First workspace retrieval succeeds
		expectedWorkspace1 := &domain.Workspace{
			ID:   "1",
			Name: "Test Workspace 1",
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspaces", ctx, userID).Return(expectedUserWorkspaces, nil)
		mockRepo.On("GetByID", ctx, "1").Return(expectedWorkspace1, nil)
		// Second workspace retrieval fails
		mockRepo.On("GetByID", ctx, "2").Return(nil, assert.AnError)

		workspaces, err := service.ListWorkspaces(ctx)
		require.Error(t, err)
		assert.Nil(t, workspaces)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}

func TestWorkspaceService_GetWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
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

	t.Run("successful get", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

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

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(expectedWorkspace, nil)

		workspace, err := service.GetWorkspace(ctx, workspaceID)
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspace, workspace)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: userID,
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(nil, assert.AnError)

		workspace, err := service.GetWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error getting workspace by ID", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// User has access but workspace retrieval fails
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(nil, assert.AnError)

		workspace, err := service.GetWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWorkspaceService_CreateWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
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
	workspaceID := "testworkspace1"

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: "test-owner",
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

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      "test-owner",
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)
		mockRepo.On("Create", ctx, mock.MatchedBy(func(w *domain.Workspace) bool {
			return w.ID == expectedWorkspace.ID &&
				w.Name == expectedWorkspace.Name &&
				w.Settings == expectedWorkspace.Settings
		})).Return(nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == expectedUserWorkspace.UserID &&
				uw.WorkspaceID == expectedUserWorkspace.WorkspaceID &&
				uw.Role == expectedUserWorkspace.Role
		})).Return(nil)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspace.ID, workspace.ID)
		assert.Equal(t, expectedWorkspace.Name, workspace.Name)
		assert.Equal(t, expectedWorkspace.Settings, workspace.Settings)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: "test-owner",
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)

		// Invalid timezone
		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "INVALID_TIMEZONE")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "does not validate as timezone")
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: "test-owner",
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("add user error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}

		expectedUser := &domain.User{
			ID: "test-owner",
		}

		mockAuthService.On("AuthenticateUserFromContext", ctx).Return(expectedUser, nil)
		mockRepo.On("Create", ctx, mock.Anything).Return(nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.Anything).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}

func TestWorkspaceService_UpdateWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
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
	userID := "test-user"
	workspaceID := "testworkspace1"

	t.Run("successful update as owner", func(t *testing.T) {
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

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(w *domain.Workspace) bool {
			return w.ID == expectedWorkspace.ID &&
				w.Name == expectedWorkspace.Name &&
				w.Settings == expectedWorkspace.Settings
		})).Return(nil)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris")
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
		assert.Equal(t, "Updated Workspace", workspace.Name)
		assert.Equal(t, "https://updated.com", workspace.Settings.WebsiteURL)
		assert.Equal(t, "https://updated.com/logo.png", workspace.Settings.LogoURL)
		assert.Equal(t, "https://updated.com/cover.png", workspace.Settings.CoverURL)
		assert.Equal(t, "Europe/Paris", workspace.Settings.Timezone)
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

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
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

		// Invalid timezone will cause validation error
		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "INVALID_TIMEZONE")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "does not validate as timezone")
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
			Role:        "owner",
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(expectedUser, nil)
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("Update", ctx, mock.Anything).Return(assert.AnError)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris")
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}

func TestWorkspaceService_DeleteWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
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
	userID := "test-user"
	workspaceID := "test-workspace"

	t.Run("successful delete as owner", func(t *testing.T) {
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
		mockRepo.On("Delete", ctx, workspaceID).Return(nil)

		err := service.DeleteWorkspace(ctx, workspaceID)
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

		err := service.DeleteWorkspace(ctx, workspaceID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})
}
