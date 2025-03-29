package repository

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockContactListRepository struct {
	mock.Mock
}

func (m *MockContactListRepository) AddContactToList(ctx context.Context, workspaceID string, contactList *domain.ContactList) error {
	args := m.Called(ctx, workspaceID, contactList)
	return args.Error(0)
}

func (m *MockContactListRepository) GetContactListByIDs(ctx context.Context, workspaceID string, email, listID string) (*domain.ContactList, error) {
	args := m.Called(ctx, workspaceID, email, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) GetContactsByListID(ctx context.Context, workspaceID string, listID string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, workspaceID, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) GetListsByEmail(ctx context.Context, workspaceID string, email string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) UpdateContactListStatus(ctx context.Context, workspaceID string, email, listID string, status domain.ContactListStatus) error {
	args := m.Called(ctx, workspaceID, email, listID, status)
	return args.Error(0)
}

func (m *MockContactListRepository) RemoveContactFromList(ctx context.Context, workspaceID string, email, listID string) error {
	args := m.Called(ctx, workspaceID, email, listID)
	return args.Error(0)
}
