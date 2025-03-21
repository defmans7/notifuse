package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
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
	GetWorkspaceUsersFn       func(ctx context.Context, workspaceID string) ([]*domain.UserWorkspace, error)
	GetUserWorkspaceFn        func(ctx context.Context, userID, workspaceID string) (*domain.UserWorkspace, error)
	GetConnectionFn           func(ctx context.Context, workspaceID string) (*sql.DB, error)
	CreateDatabaseFn          func(ctx context.Context, workspaceID string) error
	DeleteDatabaseFn          func(ctx context.Context, workspaceID string) error
	CreateInvitationFn        func(ctx context.Context, invitation *domain.WorkspaceInvitation) error
	GetInvitationByIDFn       func(ctx context.Context, id string) (*domain.WorkspaceInvitation, error)
	GetInvitationByEmailFn    func(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, error)
	IsUserWorkspaceMemberFn   func(ctx context.Context, userID, workspaceID string) (bool, error)
}

// MockWorkspaceRepository is a mock implementation of the WorkspaceRepository interface
type MockWorkspaceRepository struct {
	mock.Mock
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

func (m *MockWorkspaceRepository) GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*domain.UserWorkspace, error) {
	args := m.Called(ctx, workspaceID)
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

func (m *MockWorkspaceRepo) GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*domain.UserWorkspace, error) {
	return m.GetWorkspaceUsersFn(ctx, workspaceID)
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

func TestWorkspaceService_ListWorkspaces(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := createMockUserService()
	mockAuthService := createMockAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

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

func TestWorkspaceService_GetWorkspaceMembers(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := createMockUserService()
	mockAuthService := createMockAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

	ctx := context.Background()
	workspaceID := "workspace1"
	userID := "user1"

	t.Run("success", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Mock user workspace check
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

		// Mock get workspace users
		expectedMembers := []*domain.UserWorkspace{
			{
				UserID:      "user1",
				WorkspaceID: workspaceID,
				Role:        "owner",
			},
			{
				UserID:      "user2",
				WorkspaceID: workspaceID,
				Role:        "member",
			},
		}
		mockRepo.On("GetWorkspaceUsers", ctx, workspaceID).Return(expectedMembers, nil)

		// Call the service
		members, err := service.GetWorkspaceMembers(ctx, workspaceID, userID)

		// Check result
		require.NoError(t, err)
		assert.Equal(t, expectedMembers, members)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		// Mock user workspace check to return error
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(nil, fmt.Errorf("user is not a member of the workspace"))

		// Call the service
		members, err := service.GetWorkspaceMembers(ctx, workspaceID, userID)

		// Check result
		require.Error(t, err)
		assert.Nil(t, members)
		assert.IsType(t, &domain.ErrUnauthorized{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error getting members", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		userWorkspace := &domain.UserWorkspace{
			UserID:      userID,
			WorkspaceID: workspaceID,
			Role:        "owner",
		}

		// Mock user workspace check
		mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

		// Mock get workspace users to return error
		mockRepo.On("GetWorkspaceUsers", ctx, workspaceID).Return(nil, fmt.Errorf("database error"))

		// Call the service
		members, err := service.GetWorkspaceMembers(ctx, workspaceID, userID)

		// Check result
		require.Error(t, err)
		assert.Nil(t, members)
		assert.Contains(t, err.Error(), "database error")
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
	mockRepo := new(MockWorkspaceRepository)
	mockLogger := new(MockLogger)
	mockUserService := createMockUserService()
	mockAuthService := createMockAuthService()

	// Setup logger mock to return itself for WithField calls
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Return()

	service := NewWorkspaceService(mockRepo, mockLogger, mockUserService, mockAuthService)

	ctx := context.Background()
	workspaceID := "workspace-1"
	inviterID := "user-1"
	email := "test@example.com"

	t.Run("invalid email format", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		// Test with an invalid email
		invalidEmail := "invalid-email"
		invitation, token, err := service.InviteMember(ctx, workspaceID, inviterID, invalidEmail)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		mockRepo.On("GetByID", ctx, workspaceID).Return(nil, fmt.Errorf("workspace not found"))

		invitation, token, err := service.InviteMember(ctx, workspaceID, inviterID, email)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "workspace not found")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("inviter not a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, inviterID, workspaceID).Return(false, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, inviterID, email)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "inviter is not a member")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("create new invitation for non-existent user", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		// Reset userService repo mock
		mockUserService.repo = new(mockUserRepository)

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}

		// Mock the necessary repository calls
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, inviterID, workspaceID).Return(true, nil)

		// Setup UserService mock to return nil user (user doesn't exist)
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", ctx, email).Return(nil, &domain.ErrUserNotFound{Message: "user not found"})

		// Expect a new invitation to be created
		mockRepo.On("CreateInvitation", ctx, mock.MatchedBy(func(invitation *domain.WorkspaceInvitation) bool {
			return invitation.WorkspaceID == workspaceID &&
				invitation.InviterID == inviterID &&
				invitation.Email == email
		})).Return(nil)

		// Mock the AuthService to return a token
		mockAuthService.On("GenerateInvitationToken", mock.MatchedBy(func(invitation *domain.WorkspaceInvitation) bool {
			return invitation.WorkspaceID == workspaceID &&
				invitation.InviterID == inviterID &&
				invitation.Email == email
		})).Return("mock-invitation-token")

		invitation, token, err := service.InviteMember(ctx, workspaceID, inviterID, email)

		require.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, "mock-invitation-token", token)
		assert.Equal(t, workspaceID, invitation.WorkspaceID)
		assert.Equal(t, inviterID, invitation.InviterID)
		assert.Equal(t, email, invitation.Email)
		mockRepo.AssertExpectations(t)
		mockUserService.repo.(*mockUserRepository).AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("user already exists and is not a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		// Reset userService repo mock
		mockUserService.repo = new(mockUserRepository)

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		existingUser := &domain.User{
			ID:    "existing-user",
			Email: email,
		}

		// Mock the necessary repository calls
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, inviterID, workspaceID).Return(true, nil)

		// Setup UserService mock to return an existing user
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", ctx, email).Return(existingUser, nil)

		// User is not a member
		mockRepo.On("IsUserWorkspaceMember", ctx, existingUser.ID, workspaceID).Return(false, nil)

		// Expect the user to be added to the workspace
		mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
			return uw.UserID == existingUser.ID &&
				uw.WorkspaceID == workspaceID &&
				uw.Role == "member"
		})).Return(nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, inviterID, email)

		require.NoError(t, err)
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockUserService.repo.(*mockUserRepository).AssertExpectations(t)
	})

	t.Run("user already exists and is already a member", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		// Reset userService repo mock
		mockUserService.repo = new(mockUserRepository)

		workspace := &domain.Workspace{
			ID:   workspaceID,
			Name: "Test Workspace",
		}
		existingUser := &domain.User{
			ID:    "existing-user",
			Email: email,
		}

		// Mock the necessary repository calls
		mockRepo.On("GetByID", ctx, workspaceID).Return(workspace, nil)
		mockRepo.On("IsUserWorkspaceMember", ctx, inviterID, workspaceID).Return(true, nil)

		// Setup UserService mock to return an existing user
		mockUserService.repo.(*mockUserRepository).On("GetUserByEmail", ctx, email).Return(existingUser, nil)

		// User is already a member
		mockRepo.On("IsUserWorkspaceMember", ctx, existingUser.ID, workspaceID).Return(true, nil)

		invitation, token, err := service.InviteMember(ctx, workspaceID, inviterID, email)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "user is already a member")
		assert.Nil(t, invitation)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
		mockUserService.repo.(*mockUserRepository).AssertExpectations(t)
	})
}

// Mock functions for testing
