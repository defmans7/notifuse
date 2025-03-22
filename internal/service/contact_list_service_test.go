package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test constants
const (
	validContactID1 = "f47ac10b-58cc-4372-a567-0e02b2c3d479" // Valid UUID format
	validContactID2 = "123e4567-e89b-12d3-a456-426614174000" // Valid UUID format
	validListID1    = "test123"                              // Valid alphanum format
	validListID2    = "list456"                              // Valid alphanum format
)

type MockContactListRepository struct {
	mock.Mock
}

func (m *MockContactListRepository) AddContactToList(ctx context.Context, contactList *domain.ContactList) error {
	args := m.Called(ctx, contactList)
	return args.Error(0)
}

func (m *MockContactListRepository) GetContactListByIDs(ctx context.Context, contactID, listID string) (*domain.ContactList, error) {
	args := m.Called(ctx, contactID, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) GetContactsByListID(ctx context.Context, listID string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) GetListsByContactID(ctx context.Context, contactID string) ([]*domain.ContactList, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ContactList), args.Error(1)
}

func (m *MockContactListRepository) UpdateContactListStatus(ctx context.Context, contactID, listID string, status domain.ContactListStatus) error {
	args := m.Called(ctx, contactID, listID, status)
	return args.Error(0)
}

func (m *MockContactListRepository) RemoveContactFromList(ctx context.Context, contactID, listID string) error {
	args := m.Called(ctx, contactID, listID)
	return args.Error(0)
}

func TestContactListService_AddContactToList(t *testing.T) {
	t.Run("should add contact to list successfully", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		contact := &domain.Contact{
			UUID:       contactID,
			Email:      "test@example.com",
			ExternalID: "ext-123",
			Timezone:   "UTC",
		}

		list := &domain.List{
			ID:   listID,
			Name: "Test List",
			Type: "public",
		}

		contactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
		}

		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(contact, nil).Once()
		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()
		mockContactListRepo.On("AddContactToList", ctx, mock.MatchedBy(func(cl *domain.ContactList) bool {
			return cl.ContactID == contactID &&
				cl.ListID == listID &&
				cl.Status == domain.ContactListStatusActive &&
				!cl.CreatedAt.IsZero() &&
				!cl.UpdatedAt.IsZero()
		})).Return(nil).Once()

		// Act
		err := service.AddContactToList(ctx, contactList)

		// Assert
		assert.NoError(t, err)
		assert.False(t, contactList.CreatedAt.IsZero())
		assert.False(t, contactList.UpdatedAt.IsZero())
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should set pending status for double opt-in lists", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		contact := &domain.Contact{
			UUID:       contactID,
			Email:      "test@example.com",
			ExternalID: "ext-123",
			Timezone:   "UTC",
		}

		list := &domain.List{
			ID:            listID,
			Name:          "Test List",
			Type:          "public",
			IsDoubleOptin: true,
		}

		contactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive, // This should be overridden to pending
		}

		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(contact, nil).Once()
		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()
		mockContactListRepo.On("AddContactToList", ctx, mock.MatchedBy(func(cl *domain.ContactList) bool {
			return cl.ContactID == contactID &&
				cl.ListID == listID &&
				cl.Status == domain.ContactListStatusPending // Should be changed to pending
		})).Return(nil).Once()

		// Act
		err := service.AddContactToList(ctx, contactList)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, domain.ContactListStatusPending, contactList.Status)
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if contact not found", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID2
		listID := validListID1

		contactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
		}

		notFoundErr := &domain.ErrContactNotFound{Message: "contact not found"}
		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(nil, notFoundErr).Once()

		// Act
		err := service.AddContactToList(ctx, contactList)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contact not found")
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertNotCalled(t, "GetListByID")
		mockContactListRepo.AssertNotCalled(t, "AddContactToList")
	})

	t.Run("should return error if list not found", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID2

		contact := &domain.Contact{
			UUID:       contactID,
			Email:      "test@example.com",
			ExternalID: "ext-123",
			Timezone:   "UTC",
		}

		contactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
		}

		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(contact, nil).Once()

		notFoundErr := &domain.ErrListNotFound{Message: "list not found"}
		mockListRepo.On("GetListByID", ctx, listID).Return(nil, notFoundErr).Once()

		// Act
		err := service.AddContactToList(ctx, contactList)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list not found")
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertNotCalled(t, "AddContactToList")
	})

	t.Run("should return error if contact list repository fails", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		contact := &domain.Contact{
			UUID:       contactID,
			Email:      "test@example.com",
			ExternalID: "ext-123",
			Timezone:   "UTC",
		}

		list := &domain.List{
			ID:   listID,
			Name: "Test List",
			Type: "public",
		}

		contactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
		}

		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(contact, nil).Once()
		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()

		repoErr := errors.New("repository error")
		mockContactListRepo.On("AddContactToList", ctx, mock.Anything).Return(repoErr).Once()

		// Act
		err := service.AddContactToList(ctx, contactList)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add contact to list")
		mockContactRepo.AssertExpectations(t)
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})
}

func TestContactListService_GetContactListByIDs(t *testing.T) {
	t.Run("should get contact list by IDs successfully", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		expectedContactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockContactListRepo.On("GetContactListByIDs", ctx, contactID, listID).Return(expectedContactList, nil).Once()

		// Act
		contactList, err := service.GetContactListByIDs(ctx, contactID, listID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedContactList, contactList)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return not found error", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		notFoundErr := &domain.ErrContactListNotFound{Message: "contact list not found"}
		mockContactListRepo.On("GetContactListByIDs", ctx, contactID, listID).Return(nil, notFoundErr).Once()

		// Act
		contactList, err := service.GetContactListByIDs(ctx, contactID, listID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contactList)
		assert.IsType(t, &domain.ErrContactListNotFound{}, err)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		repoErr := errors.New("repository error")
		mockContactListRepo.On("GetContactListByIDs", ctx, contactID, listID).Return(nil, repoErr).Once()

		// Act
		contactList, err := service.GetContactListByIDs(ctx, contactID, listID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contactList)
		assert.Contains(t, err.Error(), "failed to get contact list")
		mockContactListRepo.AssertExpectations(t)
	})
}

func TestContactListService_GetContactsByListID(t *testing.T) {
	t.Run("should get contacts by list ID successfully", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		listID := validListID1

		list := &domain.List{
			ID:   listID,
			Name: "Test List",
			Type: "public",
		}

		expectedContactLists := []*domain.ContactList{
			{
				ContactID: validContactID1,
				ListID:    listID,
				Status:    domain.ContactListStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ContactID: validContactID2,
				ListID:    listID,
				Status:    domain.ContactListStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()
		mockContactListRepo.On("GetContactsByListID", ctx, listID).Return(expectedContactLists, nil).Once()

		// Act
		contactLists, err := service.GetContactsByListID(ctx, listID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedContactLists, contactLists)
		assert.Len(t, contactLists, 2)
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if list not found", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		listID := validListID2

		notFoundErr := &domain.ErrListNotFound{Message: "list not found"}
		mockListRepo.On("GetListByID", ctx, listID).Return(nil, notFoundErr).Once()

		// Act
		contactLists, err := service.GetContactsByListID(ctx, listID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contactLists)
		assert.Contains(t, err.Error(), "list not found")
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertNotCalled(t, "GetContactsByListID")
	})

	t.Run("should return empty slice when no contacts in list", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		listID := validListID1

		list := &domain.List{
			ID:   listID,
			Name: "Test List",
			Type: "public",
		}

		expectedContactLists := []*domain.ContactList{}

		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()
		mockContactListRepo.On("GetContactsByListID", ctx, listID).Return(expectedContactLists, nil).Once()

		// Act
		contactLists, err := service.GetContactsByListID(ctx, listID)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, contactLists)
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		listID := validListID1

		list := &domain.List{
			ID:   listID,
			Name: "Test List",
			Type: "public",
		}

		repoErr := errors.New("repository error")

		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()
		mockContactListRepo.On("GetContactsByListID", ctx, listID).Return(nil, repoErr).Once()

		// Act
		contactLists, err := service.GetContactsByListID(ctx, listID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contactLists)
		assert.Contains(t, err.Error(), "failed to get contacts")
		mockListRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})
}

func TestContactListService_GetListsByContactID(t *testing.T) {
	t.Run("should get lists by contact ID successfully", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1

		contact := &domain.Contact{
			UUID:       contactID,
			Email:      "test@example.com",
			ExternalID: "ext-123",
			Timezone:   "UTC",
		}

		expectedContactLists := []*domain.ContactList{
			{
				ContactID: contactID,
				ListID:    validListID1,
				Status:    domain.ContactListStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ContactID: contactID,
				ListID:    validListID2,
				Status:    domain.ContactListStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(contact, nil).Once()
		mockContactListRepo.On("GetListsByContactID", ctx, contactID).Return(expectedContactLists, nil).Once()

		// Act
		contactLists, err := service.GetListsByContactID(ctx, contactID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedContactLists, contactLists)
		assert.Len(t, contactLists, 2)
		mockContactRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if contact not found", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID2

		notFoundErr := &domain.ErrContactNotFound{Message: "contact not found"}
		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(nil, notFoundErr).Once()

		// Act
		contactLists, err := service.GetListsByContactID(ctx, contactID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contactLists)
		assert.Contains(t, err.Error(), "contact not found")
		mockContactRepo.AssertExpectations(t)
		mockContactListRepo.AssertNotCalled(t, "GetListsByContactID")
	})

	t.Run("should return empty slice when contact has no lists", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1

		contact := &domain.Contact{
			UUID:       contactID,
			Email:      "test@example.com",
			ExternalID: "ext-123",
			Timezone:   "UTC",
		}

		expectedContactLists := []*domain.ContactList{}

		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(contact, nil).Once()
		mockContactListRepo.On("GetListsByContactID", ctx, contactID).Return(expectedContactLists, nil).Once()

		// Act
		contactLists, err := service.GetListsByContactID(ctx, contactID)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, contactLists)
		mockContactRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1

		contact := &domain.Contact{
			UUID:       contactID,
			Email:      "test@example.com",
			ExternalID: "ext-123",
			Timezone:   "UTC",
		}

		repoErr := errors.New("repository error")

		mockContactRepo.On("GetContactByUUID", ctx, contactID).Return(contact, nil).Once()
		mockContactListRepo.On("GetListsByContactID", ctx, contactID).Return(nil, repoErr).Once()

		// Act
		contactLists, err := service.GetListsByContactID(ctx, contactID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contactLists)
		assert.Contains(t, err.Error(), "failed to get lists")
		mockContactRepo.AssertExpectations(t)
		mockContactListRepo.AssertExpectations(t)
	})
}

func TestContactListService_UpdateContactListStatus(t *testing.T) {
	t.Run("should update contact list status successfully", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1
		newStatus := domain.ContactListStatusUnsubscribed

		existingContactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockContactListRepo.On("GetContactListByIDs", ctx, contactID, listID).Return(existingContactList, nil).Once()
		mockContactListRepo.On("UpdateContactListStatus", ctx, contactID, listID, newStatus).Return(nil).Once()

		// Act
		err := service.UpdateContactListStatus(ctx, contactID, listID, newStatus)

		// Assert
		assert.NoError(t, err)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if contact list not found", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1
		newStatus := domain.ContactListStatusUnsubscribed

		notFoundErr := &domain.ErrContactListNotFound{Message: "contact list not found"}
		mockContactListRepo.On("GetContactListByIDs", ctx, contactID, listID).Return(nil, notFoundErr).Once()

		// Act
		err := service.UpdateContactListStatus(ctx, contactID, listID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contact list not found")
		mockContactListRepo.AssertExpectations(t)
		mockContactListRepo.AssertNotCalled(t, "UpdateContactListStatus")
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1
		newStatus := domain.ContactListStatusUnsubscribed

		existingContactList := &domain.ContactList{
			ContactID: contactID,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repoErr := errors.New("repository error")

		mockContactListRepo.On("GetContactListByIDs", ctx, contactID, listID).Return(existingContactList, nil).Once()
		mockContactListRepo.On("UpdateContactListStatus", ctx, contactID, listID, newStatus).Return(repoErr).Once()

		// Act
		err := service.UpdateContactListStatus(ctx, contactID, listID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update contact list status")
		mockContactListRepo.AssertExpectations(t)
	})
}

func TestContactListService_RemoveContactFromList(t *testing.T) {
	t.Run("should remove contact from list successfully", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		mockContactListRepo.On("RemoveContactFromList", ctx, contactID, listID).Return(nil).Once()

		// Act
		err := service.RemoveContactFromList(ctx, contactID, listID)

		// Assert
		assert.NoError(t, err)
		mockContactListRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Setup fresh mocks for this test case
		mockContactListRepo := new(MockContactListRepository)
		mockContactRepo := new(MockContactRepository)
		mockListRepo := new(MockListRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()

		service := NewContactListService(mockContactListRepo, mockContactRepo, mockListRepo, mockLogger)

		// Arrange
		ctx := context.Background()
		contactID := validContactID1
		listID := validListID1

		repoErr := errors.New("repository error")
		mockContactListRepo.On("RemoveContactFromList", ctx, contactID, listID).Return(repoErr).Once()

		// Act
		err := service.RemoveContactFromList(ctx, contactID, listID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove contact from list")
		mockContactListRepo.AssertExpectations(t)
	})
}
