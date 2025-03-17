package service

import (
	"context"
	"database/sql"
	"testing"

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

func TestWorkspaceService_ListWorkspaces(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
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

	mockRepo.On("List", ctx).Return(expectedWorkspaces, nil)

	workspaces, err := service.ListWorkspaces(ctx, "test-owner")
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspaces, workspaces)
	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_GetWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	expectedWorkspace := &domain.Workspace{
		ID:   "1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	mockRepo.On("GetByID", ctx, "1").Return(expectedWorkspace, nil)

	workspace, err := service.GetWorkspace(ctx, "1", "test-owner")
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace, workspace)
	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_CreateWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()
	expectedWorkspace := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(w *domain.Workspace) bool {
		return w.ID == expectedWorkspace.ID &&
			w.Name == expectedWorkspace.Name &&
			w.Settings.WebsiteURL == expectedWorkspace.Settings.WebsiteURL &&
			w.Settings.LogoURL == expectedWorkspace.Settings.LogoURL &&
			w.Settings.Timezone == expectedWorkspace.Settings.Timezone
	})).Return(nil)

	workspace, err := service.CreateWorkspace(ctx, "testworkspace1", "Test Workspace", "https://example.com", "https://example.com/logo.png", "UTC", "test-owner")
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
	expectedWorkspace := &domain.Workspace{
		ID:   "1",
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://updated.com",
			LogoURL:    "https://updated.com/logo.png",
			Timezone:   "UTC",
		},
	}

	mockRepo.On("Update", ctx, expectedWorkspace).Return(nil)

	workspace, err := service.UpdateWorkspace(ctx, "1", "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "UTC", "test-owner")
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace, workspace)
	mockRepo.AssertExpectations(t)
}

func TestWorkspaceService_DeleteWorkspace(t *testing.T) {
	mockRepo := new(MockWorkspaceRepository)
	service := NewWorkspaceService(mockRepo)

	ctx := context.Background()

	mockRepo.On("Delete", ctx, "1").Return(nil)

	err := service.DeleteWorkspace(ctx, "1", "test-owner")
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
