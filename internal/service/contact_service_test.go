package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Use MockContactRepository from test_mocks.go

// Tests begin here

func TestContactService_GetContactByEmail(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should get contact by email successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "test@example.com"
		expectedContact := &domain.Contact{
			Email:      email,
			ExternalID: "test-external-id",
			Timezone:   "UTC",
			FirstName:  "Test",
			LastName:   "Contact",
		}

		mockRepo.On("GetContactByEmail", ctx, email).Return(expectedContact, nil).Once()

		// Act
		contact, err := service.GetContactByEmail(ctx, email)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedContact, contact)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return not found error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "nonexistent@example.com"
		notFoundErr := &domain.ErrContactNotFound{Message: "contact not found"}

		mockRepo.On("GetContactByEmail", ctx, email).Return(nil, notFoundErr).Once()

		// Act
		contact, err := service.GetContactByEmail(ctx, email)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contact)
		assert.IsType(t, &domain.ErrContactNotFound{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "test@example.com"
		repoErr := errors.New("repository error")

		mockRepo.On("GetContactByEmail", ctx, email).Return(nil, repoErr).Once()

		// Act
		contact, err := service.GetContactByEmail(ctx, email)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contact)
		assert.Contains(t, err.Error(), "failed to get contact by email")
		mockRepo.AssertExpectations(t)
	})
}

func TestContactService_GetContactByExternalID(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should get contact by external ID successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		externalID := "ext-123"
		expectedContact := &domain.Contact{
			Email:      "test@example.com",
			ExternalID: externalID,
			Timezone:   "UTC",
			FirstName:  "Test",
			LastName:   "Contact",
		}

		mockRepo.On("GetContactByExternalID", ctx, externalID).Return(expectedContact, nil).Once()

		// Act
		contact, err := service.GetContactByExternalID(ctx, externalID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedContact, contact)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return not found error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		externalID := "nonexistent-ext-id"
		notFoundErr := &domain.ErrContactNotFound{Message: "contact not found"}

		mockRepo.On("GetContactByExternalID", ctx, externalID).Return(nil, notFoundErr).Once()

		// Act
		contact, err := service.GetContactByExternalID(ctx, externalID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contact)
		assert.IsType(t, &domain.ErrContactNotFound{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		externalID := "ext-123"
		repoErr := errors.New("repository error")

		mockRepo.On("GetContactByExternalID", ctx, externalID).Return(nil, repoErr).Once()

		// Act
		contact, err := service.GetContactByExternalID(ctx, externalID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contact)
		assert.Contains(t, err.Error(), "failed to get contact by external ID")
		mockRepo.AssertExpectations(t)
	})
}

func TestContactService_GetContacts(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should get all contacts successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expectedContacts := []*domain.Contact{
			{
				Email:      "contact1@example.com",
				ExternalID: "ext-1",
				Timezone:   "UTC",
				FirstName:  "Contact",
				LastName:   "One",
			},
			{
				Email:      "contact2@example.com",
				ExternalID: "ext-2",
				Timezone:   "UTC",
				FirstName:  "Contact",
				LastName:   "Two",
			},
		}

		mockRepo.On("GetContacts", ctx).Return(expectedContacts, nil).Once()

		// Act
		contacts, err := service.GetContacts(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedContacts, contacts)
		assert.Len(t, contacts, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return empty slice when no contacts", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expectedContacts := []*domain.Contact{}

		mockRepo.On("GetContacts", ctx).Return(expectedContacts, nil).Once()

		// Act
		contacts, err := service.GetContacts(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, contacts)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		repoErr := errors.New("repository error")

		mockRepo.On("GetContacts", ctx).Return(nil, repoErr).Once()

		// Act
		contacts, err := service.GetContacts(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contacts)
		assert.Contains(t, err.Error(), "failed to get contacts")
		mockRepo.AssertExpectations(t)
	})
}

func TestContactService_DeleteContact(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should delete contact successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "test@example.com"

		mockRepo.On("DeleteContact", ctx, email).Return(nil).Once()

		// Act
		err := service.DeleteContact(ctx, email)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "test@example.com"
		repoErr := errors.New("repository error")

		mockRepo.On("DeleteContact", ctx, email).Return(repoErr).Once()

		// Act
		err := service.DeleteContact(ctx, email)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete contact")
		mockRepo.AssertExpectations(t)
	})
}

func TestBatchImportContacts(t *testing.T) {
	t.Run("successful batch import", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		contacts := []*domain.Contact{
			{
				ExternalID: "ext1",
				Email:      "contact1@example.com",
				Timezone:   "UTC",
				FirstName:  "John",
				LastName:   "Doe",
			},
			{
				ExternalID: "ext2",
				Email:      "contact2@example.com",
				Timezone:   "Europe/Paris",
				FirstName:  "Jane",
				LastName:   "Smith",
			},
		}

		mockRepo.On("BatchImportContacts", mock.Anything, mock.AnythingOfType("[]*domain.Contact")).
			Run(func(args mock.Arguments) {
				importedContacts := args.Get(1).([]*domain.Contact)
				assert.Len(t, importedContacts, 2)

				// Verify timestamps were set
				for _, contact := range importedContacts {
					assert.False(t, contact.CreatedAt.IsZero())
					assert.False(t, contact.UpdatedAt.IsZero())
				}
			}).
			Return(nil).Once()

		err := service.BatchImportContacts(context.Background(), contacts)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid contact in batch", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		invalidContacts := []*domain.Contact{
			{
				ExternalID: "ext3",
				// Missing required Email field
				Timezone:  "UTC",
				FirstName: "Invalid",
				LastName:  "Contact",
			},
			{
				ExternalID: "ext4",
				Email:      "contact4@example.com",
				// Missing required Timezone field
				FirstName: "Another",
				LastName:  "Contact",
			},
		}

		err := service.BatchImportContacts(context.Background(), invalidContacts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid contact at index 0")
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty batch", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		emptyBatch := []*domain.Contact{}

		// Add expectation for empty batch (the service may still call the repository)
		mockRepo.On("BatchImportContacts", mock.Anything, mock.AnythingOfType("[]*domain.Contact")).
			Return(nil).Once()

		// For an empty batch, we expect no error
		err := service.BatchImportContacts(context.Background(), emptyBatch)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		contacts := []*domain.Contact{
			{
				ExternalID: "ext1",
				Email:      "contact1@example.com",
				Timezone:   "UTC",
				FirstName:  "John",
				LastName:   "Doe",
			},
		}

		mockRepo.On("BatchImportContacts", mock.Anything, mock.AnythingOfType("[]*domain.Contact")).
			Return(errors.New("repository error")).Once()

		err := service.BatchImportContacts(context.Background(), contacts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository error")
		mockRepo.AssertExpectations(t)
	})
}

func TestContactService_UpsertContact(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should create new contact successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
			FirstName:  "Test",
			LastName:   "Contact",
		}

		mockRepo.On("UpsertContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			// Verify timestamps are set
			return c.Email == contact.Email &&
				!c.CreatedAt.IsZero() && !c.UpdatedAt.IsZero()
		})).Return(true, nil).Once()

		// Act
		isNew, err := service.UpsertContact(ctx, contact)

		// Assert
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.False(t, contact.CreatedAt.IsZero())
		assert.False(t, contact.UpdatedAt.IsZero())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should update existing contact successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
			FirstName:  "Test",
			LastName:   "Contact",
		}

		mockRepo.On("UpsertContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			// Verify timestamps are set
			return c.Email == contact.Email && !c.UpdatedAt.IsZero()
		})).Return(false, nil).Once()

		// Act
		isNew, err := service.UpsertContact(ctx, contact)

		// Assert
		assert.NoError(t, err)
		assert.False(t, isNew)
		assert.False(t, contact.UpdatedAt.IsZero())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if contact is invalid", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			// Email missing, which is required
		}

		// Act
		isNew, err := service.UpsertContact(ctx, contact)

		// Assert
		assert.Error(t, err)
		assert.False(t, isNew)
		mockRepo.AssertNotCalled(t, "UpsertContact")
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
		}

		repoErr := errors.New("repository error")
		mockRepo.On("UpsertContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			return c.Email == contact.Email
		})).Return(false, repoErr).Once()

		// Act
		isNew, err := service.UpsertContact(ctx, contact)

		// Assert
		assert.Error(t, err)
		assert.False(t, isNew)
		assert.Contains(t, err.Error(), "failed to upsert contact")
		mockRepo.AssertExpectations(t)
	})
}
