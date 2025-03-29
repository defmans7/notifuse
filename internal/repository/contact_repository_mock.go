package repository

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockContactRepository is a mock implementation of domain.ContactRepository
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) GetContactByEmail(ctx context.Context, email string, workspaceID string) (*domain.Contact, error) {
	args := m.Called(ctx, email, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) BatchImportContacts(ctx context.Context, workspaceID string, contacts []*domain.Contact) error {
	args := m.Called(ctx, workspaceID, contacts)
	return args.Error(0)
}

func (m *MockContactRepository) DeleteContact(ctx context.Context, email string, workspaceID string) error {
	args := m.Called(ctx, email, workspaceID)
	return args.Error(0)
}

func (m *MockContactRepository) GetContactByExternalID(ctx context.Context, externalID string, workspaceID string) (*domain.Contact, error) {
	args := m.Called(ctx, externalID, workspaceID)
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

func (m *MockContactRepository) UpsertContact(ctx context.Context, workspaceID string, contact *domain.Contact) (bool, error) {
	args := m.Called(ctx, workspaceID, contact)
	return args.Bool(0), args.Error(1)
}
