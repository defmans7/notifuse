package service

import (
	"context"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
)

// MockListService is a mock implementation of domain.ListService for testing
type MockListService struct {
	Lists                   map[string]*domain.List
	ErrToReturn             error
	ErrListNotFoundToReturn bool
	GetListsCalled          bool
	GetListByIDCalled       bool
	LastListID              string
	CreateListCalled        bool
	LastListCreated         *domain.List
	UpdateListCalled        bool
	LastListUpdated         *domain.List
	DeleteListCalled        bool
	LastListDeleted         string
}

// NewMockListService creates a new mock list service for testing
func NewMockListService() *MockListService {
	return &MockListService{
		Lists: make(map[string]*domain.List),
	}
}

func (m *MockListService) GetLists(ctx context.Context, workspaceID string) ([]*domain.List, error) {
	m.GetListsCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	// Convert map to slice
	lists := make([]*domain.List, 0, len(m.Lists))
	for _, list := range m.Lists {
		lists = append(lists, list)
	}

	return lists, nil
}

func (m *MockListService) GetListByID(ctx context.Context, workspaceID string, id string) (*domain.List, error) {
	m.GetListByIDCalled = true
	m.LastListID = id
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.ErrListNotFoundToReturn {
		return nil, &domain.ErrListNotFound{}
	}

	list, exists := m.Lists[id]
	if !exists {
		return nil, &domain.ErrListNotFound{}
	}
	return list, nil
}

func (m *MockListService) CreateList(ctx context.Context, workspaceID string, list *domain.List) error {
	m.CreateListCalled = true
	m.LastListCreated = list
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Set timestamps
	now := time.Now()
	if list.CreatedAt.IsZero() {
		list.CreatedAt = now
	}
	list.UpdatedAt = now

	// Store the list
	m.Lists[list.ID] = list
	return nil
}

func (m *MockListService) UpdateList(ctx context.Context, workspaceID string, list *domain.List) error {
	m.UpdateListCalled = true
	m.LastListUpdated = list
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrListNotFoundToReturn {
		return &domain.ErrListNotFound{}
	}

	// Check if list exists
	existingList, exists := m.Lists[list.ID]
	if !exists {
		return &domain.ErrListNotFound{}
	}

	// Update fields
	existingList.Name = list.Name
	existingList.Description = list.Description
	existingList.UpdatedAt = time.Now()

	// Store the updated list
	m.Lists[list.ID] = existingList
	return nil
}

func (m *MockListService) DeleteList(ctx context.Context, workspaceID string, id string) error {
	m.DeleteListCalled = true
	m.LastListDeleted = id
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrListNotFoundToReturn {
		return &domain.ErrListNotFound{}
	}

	_, exists := m.Lists[id]
	if !exists {
		return &domain.ErrListNotFound{}
	}

	delete(m.Lists, id)
	return nil
}
