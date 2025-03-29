package repository

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockListRepository struct {
	mock.Mock
}

func (m *MockListRepository) CreateList(ctx context.Context, workspaceID string, list *domain.List) error {
	args := m.Called(ctx, workspaceID, list)
	return args.Error(0)
}

func (m *MockListRepository) GetListByID(ctx context.Context, workspaceID string, id string) (*domain.List, error) {
	args := m.Called(ctx, workspaceID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.List), args.Error(1)
}

func (m *MockListRepository) GetLists(ctx context.Context, workspaceID string) ([]*domain.List, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.List), args.Error(1)
}

func (m *MockListRepository) UpdateList(ctx context.Context, workspaceID string, list *domain.List) error {
	args := m.Called(ctx, workspaceID, list)
	return args.Error(0)
}

func (m *MockListRepository) DeleteList(ctx context.Context, workspaceID string, id string) error {
	args := m.Called(ctx, workspaceID, id)
	return args.Error(0)
}
