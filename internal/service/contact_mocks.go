package service

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockContactRepository is a mock implementation of domain.ContactRepository
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) GetContactByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) GetContactByExternalID(ctx context.Context, externalID string) (*domain.Contact, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) GetContacts(ctx context.Context, req *domain.GetContactsRequest) (*domain.GetContactsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GetContactsResponse), args.Error(1)
}

func (m *MockContactRepository) DeleteContact(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockContactRepository) BatchImportContacts(ctx context.Context, contacts []*domain.Contact) error {
	args := m.Called(ctx, contacts)
	return args.Error(0)
}

func (m *MockContactRepository) UpsertContact(ctx context.Context, contact *domain.Contact) (bool, error) {
	args := m.Called(ctx, contact)
	return args.Bool(0), args.Error(1)
}

// MockContactListRepository is a mock implementation of domain.ContactListRepository
type MockContactListRepository struct {
	mock.Mock
}

func (m *MockContactListRepository) AddContactToList(ctx context.Context, contactList *domain.ContactList) error {
	args := m.Called(ctx, contactList)
	return args.Error(0)
}

func (m *MockContactListRepository) GetContactListByIDs(ctx context.Context, email, listID string) (*domain.ContactList, error) {
	args := m.Called(ctx, email, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) GetContactsByListID(ctx context.Context, email string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) UpdateContactListStatus(ctx context.Context, email, listID string, status domain.ContactListStatus) error {
	args := m.Called(ctx, email, listID, status)
	return args.Error(0)
}

func (m *MockContactListRepository) RemoveContactFromList(ctx context.Context, email, listID string) error {
	args := m.Called(ctx, email, listID)
	return args.Error(0)
}

func (m *MockContactListRepository) GetListsByEmail(ctx context.Context, email string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}
