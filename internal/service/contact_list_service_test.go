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
var (
	validListID1  = "list1"
	validListID2  = "list2"
	invalidListID = ""
	validEmail1   = "contact1@example.com" // Valid email format
	validEmail2   = "contact2@example.com" // Valid email format
	invalidEmail  = "not-an-email"
)

// Use MockContactListRepository and MockContactRepository from test_mocks.go

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
		email := "test@example.com"
		listID := validListID1

		contact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "ext-123", IsNull: false},
			Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
		}

		list := &domain.List{
			ID:   listID,
			Name: "Test List",
			Type: "public",
		}

		contactList := &domain.ContactList{
			Email:  email,
			ListID: listID,
			Status: domain.ContactListStatusActive,
		}

		mockContactRepo.On("GetContactByEmail", ctx, email).Return(contact, nil).Once()
		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()
		mockContactListRepo.On("AddContactToList", ctx, mock.MatchedBy(func(cl *domain.ContactList) bool {
			return cl.Email == email &&
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
		email := "test@example.com"
		listID := validListID1

		contact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "ext-123", IsNull: false},
			Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
		}

		list := &domain.List{
			ID:            listID,
			Name:          "Test List",
			Type:          "public",
			IsDoubleOptin: true,
		}

		contactList := &domain.ContactList{
			Email:  email,
			ListID: listID,
			Status: domain.ContactListStatusActive, // This should be overridden to pending
		}

		mockContactRepo.On("GetContactByEmail", ctx, email).Return(contact, nil).Once()
		mockListRepo.On("GetListByID", ctx, listID).Return(list, nil).Once()
		mockContactListRepo.On("AddContactToList", ctx, mock.MatchedBy(func(cl *domain.ContactList) bool {
			return cl.Email == email &&
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
		email := "test@example.com"
		listID := validListID1

		contactList := &domain.ContactList{
			Email:  email,
			ListID: listID,
			Status: domain.ContactListStatusActive,
		}

		notFoundErr := &domain.ErrContactNotFound{Message: "contact not found"}
		mockContactRepo.On("GetContactByEmail", ctx, email).Return(nil, notFoundErr).Once()

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
		email := "test@example.com"
		listID := validListID2

		contact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "ext-123", IsNull: false},
			Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
		}

		contactList := &domain.ContactList{
			Email:  email,
			ListID: listID,
			Status: domain.ContactListStatusActive,
		}

		mockContactRepo.On("GetContactByEmail", ctx, email).Return(contact, nil).Once()

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
		email := "test@example.com"
		listID := validListID1

		contact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "ext-123", IsNull: false},
			Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
		}

		list := &domain.List{
			ID:   listID,
			Name: "Test List",
			Type: "public",
		}

		contactList := &domain.ContactList{
			Email:  email,
			ListID: listID,
			Status: domain.ContactListStatusActive,
		}

		mockContactRepo.On("GetContactByEmail", ctx, email).Return(contact, nil).Once()
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
		email := "test@example.com"
		listID := validListID1

		expectedContactList := &domain.ContactList{
			Email:     email,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockContactListRepo.On("GetContactListByIDs", ctx, email, listID).Return(expectedContactList, nil).Once()

		// Act
		contactList, err := service.GetContactListByIDs(ctx, email, listID)

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
		email := "test@example.com"
		listID := validListID1

		notFoundErr := &domain.ErrContactListNotFound{Message: "contact list not found"}
		mockContactListRepo.On("GetContactListByIDs", ctx, email, listID).Return(nil, notFoundErr).Once()

		// Act
		contactList, err := service.GetContactListByIDs(ctx, email, listID)

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
		email := "test@example.com"
		listID := validListID1

		repoErr := errors.New("repository error")
		mockContactListRepo.On("GetContactListByIDs", ctx, email, listID).Return(nil, repoErr).Once()

		// Act
		contactList, err := service.GetContactListByIDs(ctx, email, listID)

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
				Email:     "test@example.com",
				ListID:    listID,
				Status:    domain.ContactListStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				Email:     "test2@example.com",
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

func TestContactListService_GetListsByEmail(t *testing.T) {
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
		email := "test@example.com"

		contact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "ext-123", IsNull: false},
			Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
		}

		expectedContactLists := []*domain.ContactList{
			{
				Email:     email,
				ListID:    validListID1,
				Status:    domain.ContactListStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				Email:     email,
				ListID:    validListID2,
				Status:    domain.ContactListStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockContactRepo.On("GetContactByEmail", ctx, email).Return(contact, nil).Once()
		mockContactListRepo.On("GetListsByEmail", ctx, email).Return(expectedContactLists, nil).Once()

		// Act
		contactLists, err := service.GetListsByEmail(ctx, email)

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
		email := "test2@example.com"

		notFoundErr := &domain.ErrContactNotFound{Message: "contact not found"}
		mockContactRepo.On("GetContactByEmail", ctx, email).Return(nil, notFoundErr).Once()

		// Act
		contactLists, err := service.GetListsByEmail(ctx, email)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contactLists)
		assert.Contains(t, err.Error(), "contact not found")
		mockContactRepo.AssertExpectations(t)
		mockContactListRepo.AssertNotCalled(t, "GetListsByEmail")
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
		email := "test@example.com"

		contact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "ext-123", IsNull: false},
			Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
		}

		expectedContactLists := []*domain.ContactList{}

		mockContactRepo.On("GetContactByEmail", ctx, email).Return(contact, nil).Once()
		mockContactListRepo.On("GetListsByEmail", ctx, email).Return(expectedContactLists, nil).Once()

		// Act
		contactLists, err := service.GetListsByEmail(ctx, email)

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
		email := "test@example.com"

		contact := &domain.Contact{
			Email:      email,
			ExternalID: &domain.NullableString{String: "ext-123", IsNull: false},
			Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
		}

		repoErr := errors.New("repository error")

		mockContactRepo.On("GetContactByEmail", ctx, email).Return(contact, nil).Once()
		mockContactListRepo.On("GetListsByEmail", ctx, email).Return(nil, repoErr).Once()

		// Act
		contactLists, err := service.GetListsByEmail(ctx, email)

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
		email := "test@example.com"
		listID := validListID1
		newStatus := domain.ContactListStatusUnsubscribed

		existingContactList := &domain.ContactList{
			Email:     email,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockContactListRepo.On("GetContactListByIDs", ctx, email, listID).Return(existingContactList, nil).Once()
		mockContactListRepo.On("UpdateContactListStatus", ctx, email, listID, newStatus).Return(nil).Once()

		// Act
		err := service.UpdateContactListStatus(ctx, email, listID, newStatus)

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
		email := "test@example.com"
		listID := validListID1
		newStatus := domain.ContactListStatusUnsubscribed

		notFoundErr := &domain.ErrContactListNotFound{Message: "contact list not found"}
		mockContactListRepo.On("GetContactListByIDs", ctx, email, listID).Return(nil, notFoundErr).Once()

		// Act
		err := service.UpdateContactListStatus(ctx, email, listID, newStatus)

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
		email := "test@example.com"
		listID := validListID1
		newStatus := domain.ContactListStatusUnsubscribed

		existingContactList := &domain.ContactList{
			Email:     email,
			ListID:    listID,
			Status:    domain.ContactListStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repoErr := errors.New("repository error")

		mockContactListRepo.On("GetContactListByIDs", ctx, email, listID).Return(existingContactList, nil).Once()
		mockContactListRepo.On("UpdateContactListStatus", ctx, email, listID, newStatus).Return(repoErr).Once()

		// Act
		err := service.UpdateContactListStatus(ctx, email, listID, newStatus)

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
		email := "test@example.com"
		listID := validListID1

		mockContactListRepo.On("RemoveContactFromList", ctx, email, listID).Return(nil).Once()

		// Act
		err := service.RemoveContactFromList(ctx, email, listID)

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
		email := "test@example.com"
		listID := validListID1

		repoErr := errors.New("repository error")
		mockContactListRepo.On("RemoveContactFromList", ctx, email, listID).Return(repoErr).Once()

		// Act
		err := service.RemoveContactFromList(ctx, email, listID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove contact from list")
		mockContactListRepo.AssertExpectations(t)
	})
}
