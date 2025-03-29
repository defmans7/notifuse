package service

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockWorkspaceService struct {
	mock.Mock
}

func (m *MockWorkspaceService) CreateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *MockWorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockWorkspaceService) GetWorkspaceMembersWithEmail(ctx context.Context, id string) ([]*domain.UserWorkspaceWithEmail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserWorkspaceWithEmail), args.Error(1)
}

func (m *MockWorkspaceService) InviteMember(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, string, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.WorkspaceInvitation), args.String(1), args.Error(2)
}
