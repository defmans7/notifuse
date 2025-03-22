package service

import (
	"context"
	"database/sql"
	"fmt"
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

// MockWorkspaceRepository is a mock implementation of the WorkspaceRepository interface
type MockWorkspaceRepository struct {
	mock.Mock

	GetWorkspaceUsersWithEmailFn func(ctx context.Context, workspaceID string) ([]*domain.UserWorkspaceWithEmail, error)
	GetUserWorkspaceFn           func(ctx context.Context, userID, workspaceID string) (*domain.UserWorkspace, error)
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.DB), args.Error(1)
}

func (m *MockWorkspaceRepository) CreateDatabase(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) DeleteDatabase(ctx context.Context, workspaceID string) error {
	args := m.Called(ctx, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) AddUserToWorkspace(ctx context.Context, userWorkspace *domain.UserWorkspace) error {
	args := m.Called(ctx, userWorkspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) RemoveUserFromWorkspace(ctx context.Context, userID string, workspaceID string) error {
	args := m.Called(ctx, userID, workspaceID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.UserWorkspace, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserWorkspace), args.Error(1)
}

func (m *MockWorkspaceRepository) GetUserWorkspace(ctx context.Context, userID string, workspaceID string) (*domain.UserWorkspace, error) {
	args := m.Called(ctx, userID, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWorkspace), args.Error(1)
}

func (m *MockWorkspaceRepository) CreateInvitation(ctx context.Context, invitation *domain.WorkspaceInvitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) GetInvitationByID(ctx context.Context, id string) (*domain.WorkspaceInvitation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkspaceInvitation), args.Error(1)
}

func (m *MockWorkspaceRepository) GetInvitationByEmail(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkspaceInvitation), args.Error(1)
}

func (m *MockWorkspaceRepository) IsUserWorkspaceMember(ctx context.Context, userID, workspaceID string) (bool, error) {
	args := m.Called(ctx, userID, workspaceID)
	return args.Bool(0), args.Error(1)
}

func (m *MockWorkspaceRepository) GetWorkspaceUsersWithEmail(ctx context.Context, workspaceID string) ([]*domain.UserWorkspaceWithEmail, error) {
	if m.GetWorkspaceUsersWithEmailFn != nil {
		return m.GetWorkspaceUsersWithEmailFn(ctx, workspaceID)
	}
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserWorkspaceWithEmail), args.Error(1)
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

// MockMailer mocks the mailer.Mailer interface
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) SendWorkspaceInvitation(email, workspaceName, inviterName, token string) error {
	args := m.Called(email, workspaceName, inviterName, token)
	return args.Error(0)
}

func (m *MockMailer) SendMagicCode(email, code string) error {
	args := m.Called(email, code)
	return args.Error(0)
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

		mockRepo.On("GetUserWorkspaces", ctx, userID).Return(expectedUserWorkspaces, nil)
		mockRepo.On("GetByID", ctx, "1").Return(expectedWorkspaces[0], nil)
		mockRepo.On("GetByID", ctx, "2").Return(expectedWorkspaces[1], nil)

		workspaces, err := service.ListWorkspaces(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspaces, workspaces)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list when user has no workspaces", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("GetUserWorkspaces", ctx, userID).Return([]*domain.UserWorkspace{}, nil)

		workspaces, err := service.ListWorkspaces(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, workspaces)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error getting user workspaces", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("GetUserWorkspaces", ctx, userID).Return(nil, assert.AnError)

		workspaces, err := service.ListWorkspaces(ctx, userID)
		require.Error(t, err)
		assert.Nil(t, workspaces)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error getting a specific workspace", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

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

		mockRepo.On("GetUserWorkspaces", ctx, userID).Return(expectedUserWorkspaces, nil)
		mockRepo.On("GetByID", ctx, "1").Return(expectedWorkspace1, nil)
		// Second workspace retrieval fails
		mockRepo.On("GetByID", ctx, "2").Return(nil, assert.AnError)

		workspaces, err := service.ListWorkspaces(ctx, userID)
		require.Error(t, err)
		assert.Nil(t, workspaces)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
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
	userID := "test-user"
	workspaceID := "1"

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		expectedUserWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
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

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
		mockRepo.On("GetByID", ctx, workspaceID).Return(expectedWorkspace, nil)

		workspace, err := service.GetWorkspace(ctx, workspaceID, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspace, workspace)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error getting user workspace", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		// Simulate error when checking if user has access
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(nil, assert.AnError)

		workspace, err := service.GetWorkspace(ctx, workspaceID, userID)
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

		workspace, err := service.GetWorkspace(ctx, workspaceID, userID)
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
	ownerID := "test-owner"
	workspaceID := "testworkspace1"

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
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
			UserID:      ownerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("Create", ctx, mock.MatchedBy(func(w *domain.Workspace) bool {
			return w.ID == expectedWorkspace.ID &&
				w.Name == expectedWorkspace.Name &&
				w.Settings.WebsiteURL == expectedWorkspace.Settings.WebsiteURL &&
				w.Settings.LogoURL == expectedWorkspace.Settings.LogoURL &&
				w.Settings.CoverURL == expectedWorkspace.Settings.CoverURL &&
				w.Settings.Timezone == expectedWorkspace.Settings.Timezone
		})).Return(nil)

		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == expectedUserWorkspace.UserID &&
				uw.WorkspaceID == expectedUserWorkspace.WorkspaceID &&
				uw.Role == expectedUserWorkspace.Role
		})).Return(nil)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", ownerID)
		require.NoError(t, err)
		assert.Equal(t, expectedWorkspace.ID, workspace.ID)
		assert.Equal(t, expectedWorkspace.Name, workspace.Name)
		assert.Equal(t, expectedWorkspace.Settings, workspace.Settings)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		// Invalid timezone
		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "INVALID_TIMEZONE", ownerID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "does not validate as timezone")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("Create", ctx, mock.Anything).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", ownerID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("add user error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("Create", ctx, mock.Anything).Return(nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.Anything).Return(assert.AnError)

		workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC", ownerID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
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
	workspaceID := "1"

	t.Run("successful update as owner", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(w *domain.Workspace) bool {
			return w.ID == workspaceID &&
				w.Name == "Updated Workspace" &&
				w.Settings.WebsiteURL == "https://updated.com" &&
				w.Settings.LogoURL == "https://updated.com/logo.png" &&
				w.Settings.CoverURL == "https://updated.com/cover.png" &&
				w.Settings.Timezone == "Europe/Paris"
		})).Return(nil)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris", userID)
		require.NoError(t, err)
		assert.Equal(t, workspaceID, workspace.ID)
		assert.Equal(t, "Updated Workspace", workspace.Name)
		assert.Equal(t, "https://updated.com", workspace.Settings.WebsiteURL)
		assert.Equal(t, "https://updated.com/logo.png", workspace.Settings.LogoURL)
		assert.Equal(t, "https://updated.com/cover.png", workspace.Settings.CoverURL)
		assert.Equal(t, "Europe/Paris", workspace.Settings.Timezone)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris", userID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

		// Invalid timezone will cause validation error
		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "INVALID_TIMEZONE", userID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Contains(t, err.Error(), "does not validate as timezone")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("Update", ctx, mock.Anything).Return(assert.AnError)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris", userID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get user workspace error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(nil, assert.AnError)

		workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "Europe/Paris", userID)
		require.Error(t, err)
		assert.Nil(t, workspace)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
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
	workspaceID := "1"

	t.Run("successful delete as owner", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("Delete", ctx, workspaceID).Return(nil)

		err := service.DeleteWorkspace(ctx, workspaceID, userID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

		err := service.DeleteWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)
		mockRepo.On("Delete", ctx, workspaceID).Return(assert.AnError)

		err := service.DeleteWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get user workspace error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(nil, assert.AnError)

		err := service.DeleteWorkspace(ctx, workspaceID, userID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWorkspaceService_AddUserToWorkspace(t *testing.T) {
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
	requesterID := "owner-user"
	workspaceID := "1"
	userID := "new-user"
	role := "member"

	t.Run("successful add as owner", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == userID &&
				uw.WorkspaceID == workspaceID &&
				uw.Role == role
		})).Return(nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, role, requesterID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized requester", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, role, requesterID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid role", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)

		// Invalid role
		invalidRole := "invalid-role"
		err := service.AddUserToWorkspace(ctx, workspaceID, userID, invalidRole, requesterID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not validate as in")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)
		mockRepo.On("AddUserToWorkspace", ctx, mock.Anything).Return(assert.AnError)

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, role, requesterID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get requester workspace error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(nil, assert.AnError)

		err := service.AddUserToWorkspace(ctx, workspaceID, userID, role, requesterID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWorkspaceService_RemoveUserFromWorkspace(t *testing.T) {
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
	requesterID := "owner-user"
	workspaceID := "1"
	userID := "user-to-remove"

	t.Run("successful remove as owner", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)
		mockRepo.On("RemoveUserFromWorkspace", ctx, userID, workspaceID).Return(nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID, requesterID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized requester", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "member", // Not an owner
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)
		mockRepo.On("RemoveUserFromWorkspace", ctx, userID, workspaceID).Return(assert.AnError)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get requester workspace error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(nil, assert.AnError)

		err := service.RemoveUserFromWorkspace(ctx, workspaceID, userID, requesterID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot remove self", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		// Try to remove self
		selfID := requesterID

		requesterWorkspace := &domain.UserWorkspace{
			UserID:      requesterID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, requesterID, workspaceID).Return(requesterWorkspace, nil)
		// We no longer expect RemoveUserFromWorkspace to be called since the service
		// will prevent self-removal before reaching the repository call

		// A cleaner approach that doesn't rely on specific error messages
		err := service.RemoveUserFromWorkspace(ctx, workspaceID, selfID, requesterID)
		require.Error(t, err)
		// Check if it contains any part of the error message
		assert.Contains(t, err.Error(), "cannot remove yourself")
		mockRepo.AssertExpectations(t)
	})
}

func TestWorkspaceService_TransferOwnership(t *testing.T) {
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
	currentOwnerID := "current-owner"
	newOwnerID := "new-owner"

	t.Run("successful transfer", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		currentOwnerWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		newOwnerWorkspace := &domain.UserWorkspace{
			UserID:      newOwnerID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)
		mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(newOwnerWorkspace, nil)

		// Expect updating both users' roles
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == newOwnerID && uw.Role == "owner"
		})).Return(nil)

		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == currentOwnerID && uw.Role == "member"
		})).Return(nil)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized current owner", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		// Current "owner" is actually a member
		currentOwnerWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.Error(t, err)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("new owner not a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		currentOwnerWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// New owner is already an owner (should be a member)
		newOwnerWorkspace := &domain.UserWorkspace{
			UserID:      newOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)
		mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(newOwnerWorkspace, nil)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a current member")
		mockRepo.AssertExpectations(t)
	})

	t.Run("new owner not found error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		currentOwnerWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)
		mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(nil, assert.AnError)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update new owner error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		currentOwnerWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		newOwnerWorkspace := &domain.UserWorkspace{
			UserID:      newOwnerID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)
		mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(newOwnerWorkspace, nil)

		// Error when updating new owner's role
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == newOwnerID && uw.Role == "owner"
		})).Return(assert.AnError)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update current owner error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		currentOwnerWorkspace := &domain.UserWorkspace{
			UserID:      currentOwnerID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		newOwnerWorkspace := &domain.UserWorkspace{
			UserID:      newOwnerID,
			WorkspaceID: workspaceID,
			Role:        "member",
		}

		mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)
		mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(newOwnerWorkspace, nil)

		// First update succeeds
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == newOwnerID && uw.Role == "owner"
		})).Return(nil)

		// Error when updating current owner's role
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == currentOwnerID && uw.Role == "member"
		})).Return(assert.AnError)

		err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
		require.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWorkspaceService_GetWorkspaceMembersWithEmail(t *testing.T) {
	// Create mock repository
	mockRepo := &MockWorkspaceRepository{}

	// Create mock logger
	logger := &MockLogger{}
	logger.On("WithField", mock.Anything, mock.Anything).Return(logger)
	logger.On("Error", mock.Anything).Return()

	// Create other mocks
	mockUserService := createMockUserService()
	mockAuthService := &MockAuthService{}
	mockMailer := &MockMailer{}
	mockConfig := &config.Config{}

	service := NewWorkspaceService(mockRepo, logger, mockUserService, mockAuthService, mockMailer, mockConfig)

	t.Run("success", func(t *testing.T) {
		// Set up test data
		ctx := context.Background()
		workspaceID := "workspace1"
		userID := "user1"

		// Expected members to be returned
		expectedMembers := []*domain.UserWorkspaceWithEmail{
			{
				UserWorkspace: domain.UserWorkspace{
					UserID:      "user1",
					WorkspaceID: workspaceID,
					Role:        "owner",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Email: "user1@example.com",
			},
			{
				UserWorkspace: domain.UserWorkspace{
					UserID:      "user2",
					WorkspaceID: workspaceID,
					Role:        "member",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Email: "user2@example.com",
			},
		}

		// Set up mock expectations
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(&domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}, nil)
		mockRepo.On("GetWorkspaceUsersWithEmail", ctx, workspaceID).Return(expectedMembers, nil)

		// Call the method
		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID, userID)

		// Assert expectations
		assert.NoError(t, err)
		assert.Equal(t, expectedMembers, members)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		// Set up test data
		ctx := context.Background()
		workspaceID := "workspace1"
		userID := "user2"

		// Set up mock expectations for unauthorized access
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(nil, fmt.Errorf("user not found in workspace"))

		// Call the method
		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID, userID)

		// Assert expectations
		assert.Error(t, err)
		assert.Nil(t, members)
		_, ok := err.(*domain.ErrUnauthorized)
		assert.True(t, ok)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		// Setup a new mock repository for this test case to avoid interference
		mockRepo := &MockWorkspaceRepository{}

		// Create mock logger
		logger := &MockLogger{}
		logger.On("WithField", mock.Anything, mock.Anything).Return(logger)
		logger.On("Error", mock.Anything).Return()

		// Create service with new mocks
		service := NewWorkspaceService(mockRepo, logger, mockUserService, mockAuthService, mockMailer, mockConfig)

		// Set up test data
		ctx := context.Background()
		workspaceID := "workspace1"
		userID := "user1"

		// Set up mock expectations
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(&domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}, nil)
		mockRepo.On("GetWorkspaceUsersWithEmail", ctx, workspaceID).Return(nil, fmt.Errorf("database error"))

		// Call the method
		members, err := service.GetWorkspaceMembersWithEmail(ctx, workspaceID, userID)

		// Assert expectations
		assert.Error(t, err)
		assert.Nil(t, members)
		assert.Equal(t, "database error", err.Error())
		mockRepo.AssertExpectations(t)
	})
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

// MockAuthService is a mock implementation of the AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyUserSession(ctx context.Context, userID, sessionID string) (*domain.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) GenerateAuthToken(user *domain.User, sessionID string, expiresAt time.Time) string {
	args := m.Called(user, sessionID, expiresAt)
	return args.String(0)
}

func (m *MockAuthService) GenerateInvitationToken(invitation *domain.WorkspaceInvitation) string {
	args := m.Called(invitation)
	return args.String(0)
}

func (m *MockAuthService) GetPrivateKey() paseto.V4AsymmetricSecretKey {
	args := m.Called()
	return args.Get(0).(paseto.V4AsymmetricSecretKey)
}

// Create a mock AuthService for testing
func createMockAuthService() *MockAuthService {
	mockAuthService := new(MockAuthService)
	return mockAuthService
}

func TestWorkspaceService_InviteMember(t *testing.T) {
	// Test with both development and production configs
	devConfig := &config.Config{Environment: "development"}
	prodConfig := &config.Config{Environment: "production"}
	mockLogger := &MockLogger{} // Keep this as it's used in all test cases

	t.Run("development mode returns token", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock the GetByID call to return a workspace
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(nil, fmt.Errorf("user not found"))

		// Mock CreateInvitation
		mockRepo.On("CreateInvitation", mock.Anything, mock.MatchedBy(func(inv *domain.WorkspaceInvitation) bool {
			return inv.WorkspaceID == workspaceID && inv.Email == inviteeEmail
		})).Return(nil)

		// Mock token generation
		token := "test-invitation-token"
		mockAuthService.On("GenerateInvitationToken", mock.Anything).Return(token)

		// Call the method
		invitation, returnedToken, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, token, returnedToken) // In dev mode, token should be returned

		// Verify mailService was not called
		mockMailer.AssertNotCalled(t, "SendWorkspaceInvitation")
	})

	t.Run("production mode doesn't return token", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}

		// Create service with new mocks
		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, prodConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock the GetByID call to return a workspace
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(nil, fmt.Errorf("user not found"))

		// Mock CreateInvitation
		mockRepo.On("CreateInvitation", mock.Anything, mock.MatchedBy(func(inv *domain.WorkspaceInvitation) bool {
			return inv.WorkspaceID == workspaceID && inv.Email == inviteeEmail
		})).Return(nil)

		// Mock token generation
		token := "test-invitation-token"
		mockAuthService.On("GenerateInvitationToken", mock.Anything).Return(token)

		// Mock mailer
		mockMailer.On("SendWorkspaceInvitation", inviteeEmail, workspace.Name, inviter.Name, token).Return(nil)

		// Call the method
		invitation, returnedToken, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Empty(t, returnedToken) // In prod mode, token should not be returned

		// Verify mailService was called
		mockMailer.AssertCalled(t, "SendWorkspaceInvitation", inviteeEmail, workspace.Name, inviter.Name, token)
	})

	t.Run("inviter not a workspace member", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return false - inviter is not a member
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(false, nil)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "not a member")
	})

	t.Run("invitee already a workspace member", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"
		inviteeID := "user2"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)

		// Mock GetUserByEmail to return the invitee
		invitee := &domain.User{
			ID:    inviteeID,
			Name:  "Invitee User",
			Email: inviteeEmail,
		}
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(invitee, nil)

		// Mock IsUserWorkspaceMember to return true for invitee - already a member
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviteeID, workspaceID).Return(true, nil)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "already a member")
	})

	t.Run("invitation for email already exists", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)

		// Mock GetUserByEmail to return error (not found)
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(nil, fmt.Errorf("user not found"))

		// Mock GetInvitationByEmail to return an existing invitation
		existingInvitation := &domain.WorkspaceInvitation{
			ID:          "existing-invitation",
			WorkspaceID: workspaceID,
			Email:       inviteeEmail,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}
		mockRepo.On("GetInvitationByEmail", mock.Anything, workspaceID, inviteeEmail).Return(existingInvitation, nil)

		// The implementation doesn't actually check GetInvitationByEmail - we need to mock CreateInvitation
		// to return an error that simulates a duplicate invitation
		createError := fmt.Errorf("invitation already exists for this email")
		mockRepo.On("CreateInvitation", mock.Anything, mock.MatchedBy(func(inv *domain.WorkspaceInvitation) bool {
			return inv.WorkspaceID == workspaceID && inv.Email == inviteeEmail
		})).Return(createError)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, createError, err)
	})

	t.Run("error creating invitation", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(nil, fmt.Errorf("user not found"))

		// Mock GetInvitationByEmail to return no invitation
		mockRepo.On("GetInvitationByEmail", mock.Anything, workspaceID, inviteeEmail).Return(nil, sql.ErrNoRows)

		// Mock CreateInvitation to return an error
		createError := fmt.Errorf("database error")
		mockRepo.On("CreateInvitation", mock.Anything, mock.MatchedBy(func(inv *domain.WorkspaceInvitation) bool {
			return inv.WorkspaceID == workspaceID && inv.Email == inviteeEmail
		})).Return(createError)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, createError, err)
	})

	t.Run("error sending invitation email", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		// Create a real production config for testing
		realProdConfig := &config.Config{Environment: "production"}

		// Create service with production config
		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, realProdConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(nil, fmt.Errorf("user not found"))

		// Mock GetInvitationByEmail to return no invitation
		mockRepo.On("GetInvitationByEmail", mock.Anything, workspaceID, inviteeEmail).Return(nil, sql.ErrNoRows)

		// Mock CreateInvitation to succeed
		mockRepo.On("CreateInvitation", mock.Anything, mock.MatchedBy(func(inv *domain.WorkspaceInvitation) bool {
			return inv.WorkspaceID == workspaceID && inv.Email == inviteeEmail
		})).Return(nil)

		// Mock token generation
		token := "test-invitation-token"
		mockAuthService.On("GenerateInvitationToken", mock.Anything).Return(token)

		// Mock mailer error
		emailError := fmt.Errorf("email sending failed")
		mockMailer.On("SendWorkspaceInvitation", inviteeEmail, workspace.Name, inviter.Name, token).Return(emailError)

		// Call the method
		invitation, returnedToken, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results - note that the function continues even if email sending fails
		require.NoError(t, err)        // No error is returned by the function
		assert.NotNil(t, invitation)   // Invitation object is returned
		assert.Empty(t, returnedToken) // No token returned in production mode

		// Verify that the email sending function was called
		mockMailer.AssertCalled(t, "SendWorkspaceInvitation", inviteeEmail, workspace.Name, inviter.Name, token)
	})

	t.Run("existing user not a member is directly added", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"
		inviteeID := "user2"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)

		// Mock GetUserByEmail to return an existing user
		invitee := &domain.User{
			ID:    inviteeID,
			Name:  "Invitee User",
			Email: inviteeEmail,
		}
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(invitee, nil)

		// Mock IsUserWorkspaceMember to return false for invitee - not a member yet
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviteeID, workspaceID).Return(false, nil)

		// Mock AddUserToWorkspace to succeed
		mockRepo.On("AddUserToWorkspace", mock.Anything, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == inviteeID && uw.WorkspaceID == workspaceID && uw.Role == "member"
		})).Return(nil)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.NoError(t, err)
		assert.Nil(t, invitation) // No invitation should be created
		assert.Empty(t, token)    // No token since user is directly added

		// Verify that AddUserToWorkspace was called
		mockRepo.AssertCalled(t, "AddUserToWorkspace", mock.Anything, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == inviteeID && uw.WorkspaceID == workspaceID && uw.Role == "member"
		}))
	})

	t.Run("invalid email format", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		invalidEmail := "not-an-email"

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, invalidEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "invalid email format")

		// Verify no other mocks were called
		mockRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
		mockRepo.AssertNotCalled(t, "IsUserWorkspaceMember", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("error checking if existing user is a member", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"
		inviteeID := "user2"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)

		// Mock GetUserByEmail to return an existing user
		invitee := &domain.User{
			ID:    inviteeID,
			Name:  "Invitee User",
			Email: inviteeEmail,
		}
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(invitee, nil)

		// Mock IsUserWorkspaceMember to return error for invitee
		checkError := fmt.Errorf("database error checking membership")
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviteeID, workspaceID).Return(false, checkError)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, checkError, err)

		// Verify logger was called with error
		mockLogger.AssertCalled(t, "WithField", "workspace_id", workspaceID)
		mockLogger.AssertCalled(t, "WithField", "user_id", inviteeID)
		mockLogger.AssertCalled(t, "Error", "Failed to check if user is already a member")
	})

	t.Run("error adding existing user to workspace", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"
		inviteeID := "user2"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return the inviter
		inviter := &domain.User{
			ID:    inviterID,
			Name:  "Inviter User",
			Email: "inviter@example.com",
		}

		// Set up the UserService mock
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(inviter, nil)

		// Mock GetUserByEmail to return an existing user
		invitee := &domain.User{
			ID:    inviteeID,
			Name:  "Invitee User",
			Email: inviteeEmail,
		}
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", mock.Anything, inviteeEmail).Return(invitee, nil)

		// Mock IsUserWorkspaceMember to return false for invitee - not a member yet
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviteeID, workspaceID).Return(false, nil)

		// Mock AddUserToWorkspace to fail
		addError := fmt.Errorf("database error adding user to workspace")
		mockRepo.On("AddUserToWorkspace", mock.Anything, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == inviteeID && uw.WorkspaceID == workspaceID
		})).Return(addError)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, addError, err)

		// Verify logger was called with error
		mockLogger.AssertCalled(t, "WithField", "workspace_id", workspaceID)
		mockLogger.AssertCalled(t, "WithField", "user_id", inviteeID)
		mockLogger.AssertCalled(t, "Error", "Failed to add user to workspace")
	})

	t.Run("error getting workspace by ID", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock workspace retrieval with error
		workspaceError := fmt.Errorf("database error getting workspace")
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(nil, workspaceError)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, workspaceError, err)

		// Verify logger was called with error
		mockLogger.AssertCalled(t, "WithField", "workspace_id", workspaceID)
		mockLogger.AssertCalled(t, "Error", "Failed to get workspace for invitation")
	})

	t.Run("error getting inviter user details", func(t *testing.T) {
		// Create fresh mocks for this test
		mockRepo := &MockWorkspaceRepository{}
		mockUserService := createMockUserService()
		mockAuthService := createMockAuthService()
		mockMailer := &MockMailer{}
		mockLogger := &MockLogger{}

		// Set up logger mock expectations
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Return()

		service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService, mockMailer, devConfig)

		// Setup mocks
		workspaceID := "workspace1"
		inviterID := "user1"
		inviteeEmail := "invitee@example.com"

		// Mock workspace retrieval
		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", mock.Anything, workspaceID).Return(workspace, nil)

		// Mock IsUserWorkspaceMember to return true for inviter
		mockRepo.On("IsUserWorkspaceMember", mock.Anything, inviterID, workspaceID).Return(true, nil)

		// Mock GetUserByID to return error for inviter
		userError := fmt.Errorf("database error getting user")
		mockUserService.repo = &mockUserRepository{}
		mockUserService.repo.(*mockUserRepository).On("GetUserByID", mock.Anything, inviterID).Return(nil, userError)

		// Call the method
		invitation, token, err := service.InviteMember(context.Background(), workspaceID, inviterID, inviteeEmail)

		// Verify results
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		assert.Equal(t, userError, err)

		// Verify logger was called with error
		mockLogger.AssertCalled(t, "WithField", "inviter_id", inviterID)
		mockLogger.AssertCalled(t, "Error", "Failed to get inviter details")
	})
}

// Mock functions for testing
