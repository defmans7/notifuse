package service

import (
	"context"

	"github.com/Notifuse/notifuse/internal/domain"
)

// MockContactListService is a mock implementation of domain.ContactListService for testing
type MockContactListService struct {
	AddContactToListCalled      bool
	GetContactListByIDsCalled   bool
	GetContactsByListCalled     bool
	GetListsByEmailCalled       bool
	UpdateContactListCalled     bool
	RemoveContactFromListCalled bool

	ErrToReturn  error
	ContactList  *domain.ContactList
	ContactLists []*domain.ContactList
}

func (m *MockContactListService) AddContactToList(ctx context.Context, workspaceID string, contactList *domain.ContactList) error {
	m.AddContactToListCalled = true
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	return nil
}

func (m *MockContactListService) GetContactListByIDs(ctx context.Context, workspaceID string, email string, listID string) (*domain.ContactList, error) {
	m.GetContactListByIDsCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	return m.ContactList, nil
}

func (m *MockContactListService) GetContactsByListID(ctx context.Context, workspaceID string, listID string) ([]*domain.ContactList, error) {
	m.GetContactsByListCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	return m.ContactLists, nil
}

func (m *MockContactListService) GetListsByEmail(ctx context.Context, workspaceID string, email string) ([]*domain.ContactList, error) {
	m.GetListsByEmailCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	return m.ContactLists, nil
}

func (m *MockContactListService) UpdateContactListStatus(ctx context.Context, workspaceID string, email string, listID string, status domain.ContactListStatus) error {
	m.UpdateContactListCalled = true
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	return nil
}

func (m *MockContactListService) RemoveContactFromList(ctx context.Context, workspaceID string, email string, listID string) error {
	m.RemoveContactFromListCalled = true
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	return nil
}
