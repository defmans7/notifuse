package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notifuse/server/internal/domain"
)

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

func TestWorkspaceService_ListWorkspaces(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	userID := "test-user"
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
}

func TestWorkspaceService_GetWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	userID := "test-user"
	workspaceID := "1"

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
			Timezone:   "UTC",
		},
	}

	mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
	mockRepo.On("GetByID", ctx, workspaceID).Return(expectedWorkspace, nil)

	workspace, err := service.GetWorkspace(ctx, workspaceID, userID)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace, workspace)
	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_CreateWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	ownerID := "test-owner"
	workspaceID := "testworkspace1"

	expectedWorkspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
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
			w.Settings.Timezone == expectedWorkspace.Settings.Timezone
	})).Return(nil)

	mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
		return uw.UserID == expectedUserWorkspace.UserID &&
			uw.WorkspaceID == expectedUserWorkspace.WorkspaceID &&
			uw.Role == expectedUserWorkspace.Role
	})).Return(nil)

	workspace, err := service.CreateWorkspace(ctx, workspaceID, "Test Workspace", "https://example.com", "https://example.com/logo.png", "UTC", ownerID)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace.ID, workspace.ID)
	assert.Equal(t, expectedWorkspace.Name, workspace.Name)
	assert.Equal(t, expectedWorkspace.Settings, workspace.Settings)
	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_UpdateWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	userID := "test-user"
	workspaceID := "1"

	expectedUserWorkspace := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	expectedWorkspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://updated.com",
			LogoURL:    "https://updated.com/logo.png",
			Timezone:   "UTC",
		},
	}

	mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(w *domain.Workspace) bool {
		return w.ID == expectedWorkspace.ID &&
			w.Name == expectedWorkspace.Name &&
			w.Settings.WebsiteURL == expectedWorkspace.Settings.WebsiteURL &&
			w.Settings.LogoURL == expectedWorkspace.Settings.LogoURL &&
			w.Settings.Timezone == expectedWorkspace.Settings.Timezone
	})).Return(nil)

	workspace, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "UTC", userID)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace.ID, workspace.ID)
	assert.Equal(t, expectedWorkspace.Name, workspace.Name)
	assert.Equal(t, expectedWorkspace.Settings, workspace.Settings)
	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_DeleteWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	userID := "test-user"
	workspaceID := "1"

	expectedUserWorkspace := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(expectedUserWorkspace, nil)
	mockRepo.On("Delete", ctx, workspaceID).Return(nil)

	err := service.DeleteWorkspace(ctx, workspaceID, userID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_Validation(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()

	// Test invalid timezone
	_, err := service.CreateWorkspace(ctx, "testworkspace1", "Test Workspace", "https://example.com", "https://example.com/logo.png", "Invalid/Timezone", "test-owner")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Settings.timezone: Invalid/Timezone does not validate as timezone")

	// Test invalid website URL
	_, err = service.CreateWorkspace(ctx, "testworkspace1", "Test Workspace", "not-a-url", "https://example.com/logo.png", "UTC", "test-owner")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Settings.website_url: not-a-url does not validate as url")

	// Test invalid logo URL
	_, err = service.CreateWorkspace(ctx, "testworkspace1", "Test Workspace", "https://example.com", "not-a-url", "UTC", "test-owner")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Settings.logo_url: not-a-url does not validate as url")

	// Test empty name
	_, err = service.CreateWorkspace(ctx, "testworkspace1", "", "https://example.com", "https://example.com/logo.png", "UTC", "test-owner")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name: non zero value required")

	// Test invalid workspace ID
	_, err = service.CreateWorkspace(ctx, "test-workspace-1", "Test Workspace", "https://example.com", "https://example.com/logo.png", "UTC", "test-owner")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id: test-workspace-1 does not validate as alphanum")
}

func TestWorkspaceService_Unauthorized(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	userID := "test-user"
	workspaceID := "1"

	// Test unauthorized update
	userWorkspace := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        "member",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

	_, err := service.UpdateWorkspace(ctx, workspaceID, "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "UTC", userID)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrUnauthorized{}, err)

	// Test unauthorized delete
	mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

	err = service.DeleteWorkspace(ctx, workspaceID, userID)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrUnauthorized{}, err)

	// Test unauthorized add user
	mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

	err = service.AddUserToWorkspace(ctx, workspaceID, "new-user", "member", userID)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrUnauthorized{}, err)

	// Test unauthorized remove user
	mockRepo.On("GetUserWorkspace", ctx, userID, workspaceID).Return(userWorkspace, nil)

	err = service.RemoveUserFromWorkspace(ctx, workspaceID, "other-user", userID)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrUnauthorized{}, err)

	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_TransferOwnership(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	workspaceID := "1"
	currentOwnerID := "owner-id"
	newOwnerID := "member-id"

	currentOwnerWorkspace := &domain.UserWorkspace{
		UserID:      currentOwnerID,
		WorkspaceID: workspaceID,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	newOwnerWorkspace := &domain.UserWorkspace{
		UserID:      newOwnerID,
		WorkspaceID: workspaceID,
		Role:        "member",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test successful transfer
	mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)
	mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(newOwnerWorkspace, nil)

	// Expect new owner role update
	mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
		return uw.UserID == newOwnerID &&
			uw.WorkspaceID == workspaceID &&
			uw.Role == "owner"
	})).Return(nil)

	// Expect current owner role update
	mockRepo.On("AddUserToWorkspace", ctx, mock.MatchedBy(func(uw *domain.UserWorkspace) bool {
		return uw.UserID == currentOwnerID &&
			uw.WorkspaceID == workspaceID &&
			uw.Role == "member"
	})).Return(nil)

	err := service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Test unauthorized transfer (current owner is not an owner)
	mockRepo = new(MockWorkspaceRepository)
	service = NewWorkspaceService(mockRepo)

	currentOwnerWorkspace.Role = "member"
	mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)

	err = service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrUnauthorized{}, err)
	mockRepo.AssertExpectations(t)

	// Test invalid transfer (new owner is not a member)
	mockRepo = new(MockWorkspaceRepository)
	service = NewWorkspaceService(mockRepo)

	currentOwnerWorkspace.Role = "owner"
	newOwnerWorkspace.Role = "owner"
	mockRepo.On("GetUserWorkspace", ctx, currentOwnerID, workspaceID).Return(currentOwnerWorkspace, nil)
	mockRepo.On("GetUserWorkspace", ctx, newOwnerID, workspaceID).Return(newOwnerWorkspace, nil)

	err = service.TransferOwnership(ctx, workspaceID, newOwnerID, currentOwnerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new owner must be a current member")
	mockRepo.AssertExpectations(t)
}
