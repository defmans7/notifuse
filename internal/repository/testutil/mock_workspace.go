package testutil

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Notifuse/notifuse/internal/domain"
)

// MockWorkspaceRepository is a mock implementation of domain.WorkspaceRepository
type MockWorkspaceRepository struct {
	DB           *sql.DB
	WorkspaceDBs map[string]*sql.DB
}

// NewMockWorkspaceRepository creates a new mock workspace repository
func NewMockWorkspaceRepository(db *sql.DB) *MockWorkspaceRepository {
	return &MockWorkspaceRepository{
		DB:           db,
		WorkspaceDBs: make(map[string]*sql.DB),
	}
}

// AddWorkspaceDB adds a workspace database to the mock
func (m *MockWorkspaceRepository) AddWorkspaceDB(workspaceID string, db *sql.DB) {
	m.WorkspaceDBs[workspaceID] = db
}

func (m *MockWorkspaceRepository) GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error) {
	if db, ok := m.WorkspaceDBs[workspaceID]; ok {
		return db, nil
	}
	return nil, fmt.Errorf("workspace %s not found", workspaceID)
}

// Implement other required methods with empty implementations
func (m *MockWorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	return nil
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	return nil, nil
}

func (m *MockWorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	return nil, nil
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	return nil
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockWorkspaceRepository) CreateDatabase(ctx context.Context, workspaceID string) error {
	return nil
}

func (m *MockWorkspaceRepository) DeleteDatabase(ctx context.Context, workspaceID string) error {
	return nil
}

func (m *MockWorkspaceRepository) AddUserToWorkspace(ctx context.Context, userWorkspace *domain.UserWorkspace) error {
	return nil
}

func (m *MockWorkspaceRepository) RemoveUserFromWorkspace(ctx context.Context, userID string, workspaceID string) error {
	return nil
}

func (m *MockWorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.UserWorkspace, error) {
	return nil, nil
}

func (m *MockWorkspaceRepository) GetUserWorkspace(ctx context.Context, userID string, workspaceID string) (*domain.UserWorkspace, error) {
	return nil, nil
}

func (m *MockWorkspaceRepository) CreateInvitation(ctx context.Context, invitation *domain.WorkspaceInvitation) error {
	return nil
}

func (m *MockWorkspaceRepository) GetInvitationByID(ctx context.Context, id string) (*domain.WorkspaceInvitation, error) {
	return nil, nil
}

func (m *MockWorkspaceRepository) GetInvitationByEmail(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, error) {
	return nil, nil
}

func (m *MockWorkspaceRepository) IsUserWorkspaceMember(ctx context.Context, userID, workspaceID string) (bool, error) {
	return false, nil
}

func (m *MockWorkspaceRepository) GetWorkspaceUsersWithEmail(ctx context.Context, workspaceID string) ([]*domain.UserWorkspaceWithEmail, error) {
	return nil, nil
}
