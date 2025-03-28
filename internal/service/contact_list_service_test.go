package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
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

type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) GetContactByEmail(ctx context.Context, workspaceID string, email string) (*domain.Contact, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) BatchImportContacts(ctx context.Context, workspaceID string, contacts []*domain.Contact) error {
	args := m.Called(ctx, workspaceID, contacts)
	return args.Error(0)
}

func (m *MockContactRepository) DeleteContact(ctx context.Context, workspaceID string, email string) error {
	args := m.Called(ctx, workspaceID, email)
	return args.Error(0)
}

func (m *MockContactRepository) GetContactByExternalID(ctx context.Context, workspaceID string, externalID string) (*domain.Contact, error) {
	args := m.Called(ctx, workspaceID, externalID)
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

func TestContactListService_AddContactToList(t *testing.T) {
	mockRepo := new(MockContactListRepository)
	mockAuthService := new(MockAuthService)
	mockContactRepo := new(MockContactRepository)
	mockListRepo := new(MockListRepository)
	mockLogger := new(MockLogger)

	service := NewContactListService(mockRepo, mockAuthService, mockContactRepo, mockListRepo, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	contactList := &domain.ContactList{
		Email:  "test@example.com",
		ListID: "testlist123",
		Status: domain.ContactListStatusActive,
	}
	contact := &domain.Contact{
		Email: "test@example.com",
	}
	list := &domain.List{
		ID:            "testlist123",
		IsDoubleOptin: true,
	}

	t.Run("successful addition", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockContactRepo.On("GetContactByEmail", ctx, workspaceID, contactList.Email).Return(contact, nil)
		mockListRepo.On("GetListByID", ctx, workspaceID, contactList.ListID).Return(list, nil)
		mockRepo.On("AddContactToList", ctx, workspaceID, mock.MatchedBy(func(cl *domain.ContactList) bool {
			return cl.Email == contactList.Email &&
				cl.ListID == contactList.ListID &&
				cl.Status == domain.ContactListStatusPending // Should be pending due to double opt-in
		})).Return(nil)

		err := service.AddContactToList(ctx, workspaceID, contactList)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.AddContactToList(ctx, workspaceID, contactList)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockContactRepo.On("GetContactByEmail", ctx, workspaceID, contactList.Email).Return(nil, errors.New("contact not found"))

		err := service.AddContactToList(ctx, workspaceID, contactList)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("list not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockContactRepo.On("GetContactByEmail", ctx, workspaceID, contactList.Email).Return(contact, nil)
		mockListRepo.On("GetListByID", ctx, workspaceID, contactList.ListID).Return(nil, errors.New("list not found"))

		err := service.AddContactToList(ctx, workspaceID, contactList)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockContactRepo.On("GetContactByEmail", ctx, workspaceID, contactList.Email).Return(contact, nil)
		mockListRepo.On("GetListByID", ctx, workspaceID, contactList.ListID).Return(list, nil)
		mockRepo.On("AddContactToList", ctx, workspaceID, mock.Anything).Return(errors.New("repo error"))
		mockLogger.On("WithField", "email", contactList.Email).Return(mockLogger)
		mockLogger.On("WithField", "list_id", contactList.ListID).Return(mockLogger)
		mockLogger.On("Error", "Failed to add contact to list: repo error").Return()

		err := service.AddContactToList(ctx, workspaceID, contactList)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactListService_GetContactListByIDs(t *testing.T) {
	mockRepo := new(MockContactListRepository)
	mockAuthService := new(MockAuthService)
	mockContactRepo := new(MockContactRepository)
	mockListRepo := new(MockListRepository)
	mockLogger := new(MockLogger)

	service := NewContactListService(mockRepo, mockAuthService, mockContactRepo, mockListRepo, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	email := "test@example.com"
	listID := "testlist123"
	contactList := &domain.ContactList{
		Email:  email,
		ListID: listID,
		Status: domain.ContactListStatusActive,
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactListByIDs", ctx, workspaceID, email, listID).Return(contactList, nil)

		result, err := service.GetContactListByIDs(ctx, workspaceID, email, listID)
		assert.NoError(t, err)
		assert.Equal(t, contactList, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContactListByIDs(ctx, workspaceID, email, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("contact list not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactListByIDs", ctx, workspaceID, email, listID).Return(nil, &domain.ErrContactListNotFound{})

		result, err := service.GetContactListByIDs(ctx, workspaceID, email, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactListByIDs", ctx, workspaceID, email, listID).Return(nil, errors.New("repo error"))
		mockLogger.On("WithField", "email", email).Return(mockLogger)
		mockLogger.On("WithField", "list_id", listID).Return(mockLogger)
		mockLogger.On("Error", "Failed to get contact list: repo error").Return()

		result, err := service.GetContactListByIDs(ctx, workspaceID, email, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactListService_GetContactsByListID(t *testing.T) {
	mockRepo := new(MockContactListRepository)
	mockAuthService := new(MockAuthService)
	mockContactRepo := new(MockContactRepository)
	mockListRepo := new(MockListRepository)
	mockLogger := new(MockLogger)

	service := NewContactListService(mockRepo, mockAuthService, mockContactRepo, mockListRepo, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	listID := "testlist123"
	contactLists := []*domain.ContactList{
		{
			Email:  "test1@example.com",
			ListID: listID,
			Status: domain.ContactListStatusActive,
		},
		{
			Email:  "test2@example.com",
			ListID: listID,
			Status: domain.ContactListStatusPending,
		},
	}
	list := &domain.List{
		ID: listID,
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockListRepo.On("GetListByID", ctx, workspaceID, listID).Return(list, nil)
		mockRepo.On("GetContactsByListID", ctx, workspaceID, listID).Return(contactLists, nil)

		result, err := service.GetContactsByListID(ctx, workspaceID, listID)
		assert.NoError(t, err)
		assert.Equal(t, contactLists, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContactsByListID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("list not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockListRepo.On("GetListByID", ctx, workspaceID, listID).Return(nil, errors.New("list not found"))

		result, err := service.GetContactsByListID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockListRepo.On("GetListByID", ctx, workspaceID, listID).Return(list, nil)
		mockRepo.On("GetContactsByListID", ctx, workspaceID, listID).Return(nil, errors.New("repo error"))
		mockLogger.On("WithField", "list_id", listID).Return(mockLogger)
		mockLogger.On("Error", "Failed to get contacts for list: repo error").Return()

		result, err := service.GetContactsByListID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactListService_GetListsByEmail(t *testing.T) {
	mockRepo := new(MockContactListRepository)
	mockAuthService := new(MockAuthService)
	mockContactRepo := new(MockContactRepository)
	mockListRepo := new(MockListRepository)
	mockLogger := new(MockLogger)

	service := NewContactListService(mockRepo, mockAuthService, mockContactRepo, mockListRepo, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	email := "test@example.com"
	contactLists := []*domain.ContactList{
		{
			Email:  email,
			ListID: "list1",
			Status: domain.ContactListStatusActive,
		},
		{
			Email:  email,
			ListID: "list2",
			Status: domain.ContactListStatusPending,
		},
	}
	contact := &domain.Contact{
		Email: email,
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockContactRepo.On("GetContactByEmail", ctx, email, workspaceID).Return(contact, nil)
		mockRepo.On("GetListsByEmail", ctx, workspaceID, email).Return(contactLists, nil)

		result, err := service.GetListsByEmail(ctx, workspaceID, email)
		assert.NoError(t, err)
		assert.Equal(t, contactLists, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetListsByEmail(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockContactRepo.On("GetContactByEmail", ctx, email, workspaceID).Return(nil, errors.New("contact not found"))

		result, err := service.GetListsByEmail(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockContactRepo.On("GetContactByEmail", ctx, email, workspaceID).Return(contact, nil)
		mockRepo.On("GetListsByEmail", ctx, workspaceID, email).Return(nil, errors.New("repo error"))
		mockLogger.On("WithField", "email", email).Return(mockLogger)
		mockLogger.On("Error", "Failed to get lists for contact: repo error").Return()

		result, err := service.GetListsByEmail(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactListService_UpdateContactListStatus(t *testing.T) {
	mockRepo := new(MockContactListRepository)
	mockAuthService := new(MockAuthService)
	mockContactRepo := new(MockContactRepository)
	mockListRepo := new(MockListRepository)
	mockLogger := new(MockLogger)

	service := NewContactListService(mockRepo, mockAuthService, mockContactRepo, mockListRepo, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	email := "test@example.com"
	listID := "testlist123"
	status := domain.ContactListStatusActive
	contactList := &domain.ContactList{
		Email:  email,
		ListID: listID,
		Status: status,
	}

	t.Run("successful update", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactListByIDs", ctx, workspaceID, email, listID).Return(contactList, nil)
		mockRepo.On("UpdateContactListStatus", ctx, workspaceID, email, listID, status).Return(nil)

		err := service.UpdateContactListStatus(ctx, workspaceID, email, listID, status)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.UpdateContactListStatus(ctx, workspaceID, email, listID, status)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("contact list not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactListByIDs", ctx, workspaceID, email, listID).Return(nil, &domain.ErrContactListNotFound{})

		err := service.UpdateContactListStatus(ctx, workspaceID, email, listID, status)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactListByIDs", ctx, workspaceID, email, listID).Return(contactList, nil)
		mockRepo.On("UpdateContactListStatus", ctx, workspaceID, email, listID, status).Return(errors.New("repo error"))
		mockLogger.On("WithField", "email", email).Return(mockLogger)
		mockLogger.On("WithField", "list_id", listID).Return(mockLogger)
		mockLogger.On("Error", "Failed to update contact list status: repo error").Return()

		err := service.UpdateContactListStatus(ctx, workspaceID, email, listID, status)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactListService_RemoveContactFromList(t *testing.T) {
	mockRepo := new(MockContactListRepository)
	mockAuthService := new(MockAuthService)
	mockContactRepo := new(MockContactRepository)
	mockListRepo := new(MockListRepository)
	mockLogger := new(MockLogger)

	service := NewContactListService(mockRepo, mockAuthService, mockContactRepo, mockListRepo, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	email := "test@example.com"
	listID := "testlist123"

	t.Run("successful removal", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("RemoveContactFromList", ctx, workspaceID, email, listID).Return(nil)

		err := service.RemoveContactFromList(ctx, workspaceID, email, listID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.RemoveContactFromList(ctx, workspaceID, email, listID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockContactRepo.Mock = mock.Mock{}
		mockListRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("RemoveContactFromList", ctx, workspaceID, email, listID).Return(errors.New("repo error"))
		mockLogger.On("WithField", "email", email).Return(mockLogger)
		mockLogger.On("WithField", "list_id", listID).Return(mockLogger)
		mockLogger.On("Error", "Failed to remove contact from list: repo error").Return()

		err := service.RemoveContactFromList(ctx, workspaceID, email, listID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
