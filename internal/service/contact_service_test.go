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

// Tests begin here

func TestContactService_GetContactByEmail(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	// Set up expected error and contact
	expectedError := &domain.ErrContactNotFound{Message: "contact not found"}
	expectedContact := &domain.Contact{
		Email:      "test@example.com",
		ExternalID: "ext1",
		Timezone:   "UTC",
		FirstName: domain.NullableString{
			String: "Test",
			IsNull: false,
		},
		LastName: domain.NullableString{
			String: "Contact",
			IsNull: false,
		},
	}

	// Test error case
	mockRepo.On("GetContactByEmail", mock.Anything, "nonexistent@example.com").Return(nil, expectedError).Once()
	_, err := service.GetContactByEmail(context.Background(), "nonexistent@example.com")
	assert.Error(t, err)
	assert.Equal(t, expectedError.Error(), err.Error())

	// Test success case
	mockRepo.On("GetContactByEmail", mock.Anything, "test@example.com").Return(expectedContact, nil).Once()
	contact, err := service.GetContactByEmail(context.Background(), "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedContact, contact)
}

func TestContactService_GetContactByExternalID(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	// Set up expected error and contact
	expectedError := &domain.ErrContactNotFound{Message: "contact not found"}
	expectedContact := &domain.Contact{
		Email:      "test@example.com",
		ExternalID: "ext1",
		Timezone:   "UTC",
		FirstName: domain.NullableString{
			String: "Test",
			IsNull: false,
		},
		LastName: domain.NullableString{
			String: "Contact",
			IsNull: false,
		},
	}

	// Test error case
	mockRepo.On("GetContactByExternalID", mock.Anything, "nonexistent").Return(nil, expectedError).Once()
	_, err := service.GetContactByExternalID(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, expectedError.Error(), err.Error())

	// Test success case
	mockRepo.On("GetContactByExternalID", mock.Anything, "ext1").Return(expectedContact, nil).Once()
	contact, err := service.GetContactByExternalID(context.Background(), "ext1")
	assert.NoError(t, err)
	assert.Equal(t, expectedContact, contact)
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
				FirstName:  domain.NullableString{String: "Contact", IsNull: false},
				LastName:   domain.NullableString{String: "One", IsNull: false},
			},
			{
				Email:      "contact2@example.com",
				ExternalID: "ext-2",
				Timezone:   "UTC",
				FirstName:  domain.NullableString{String: "Contact", IsNull: false},
				LastName:   domain.NullableString{String: "Two", IsNull: false},
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
				FirstName:  domain.NullableString{String: "John", IsNull: false},
				LastName:   domain.NullableString{String: "Doe", IsNull: false},
			},
			{
				ExternalID: "ext2",
				Email:      "contact2@example.com",
				Timezone:   "Europe/Paris",
				FirstName:  domain.NullableString{String: "Jane", IsNull: false},
				LastName:   domain.NullableString{String: "Smith", IsNull: false},
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
				FirstName: domain.NullableString{String: "Invalid", IsNull: false},
				LastName:  domain.NullableString{String: "Contact", IsNull: false},
			},
			{
				ExternalID: "ext4",
				Email:      "contact4@example.com",
				// Missing required Timezone field
				FirstName: domain.NullableString{String: "Another", IsNull: false},
				LastName:  domain.NullableString{String: "Contact", IsNull: false},
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
				FirstName:  domain.NullableString{String: "John", IsNull: false},
				LastName:   domain.NullableString{String: "Doe", IsNull: false},
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
	t.Run("successful upsert - create", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		contact := &domain.Contact{
			Email:      "contact1@example.com",
			ExternalID: "ext1",
			Timezone:   "UTC",
			FirstName:  domain.NullableString{String: "John", IsNull: false},
			LastName:   domain.NullableString{String: "Doe", IsNull: false},
		}

		mockRepo.On("UpsertContact", mock.Anything, mock.AnythingOfType("*domain.Contact")).
			Run(func(args mock.Arguments) {
				updatedContact := args.Get(1).(*domain.Contact)
				assert.Equal(t, contact.Email, updatedContact.Email)
				assert.Equal(t, contact.ExternalID, updatedContact.ExternalID)
				assert.False(t, updatedContact.CreatedAt.IsZero())
				assert.False(t, updatedContact.UpdatedAt.IsZero())
			}).
			Return(true, nil).Once()

		isNew, err := service.UpsertContact(context.Background(), contact)
		assert.NoError(t, err)
		assert.True(t, isNew)
	})

	t.Run("successful upsert - update", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		now := time.Now().UTC()
		contact := &domain.Contact{
			Email:      "contact1@example.com",
			ExternalID: "ext1",
			Timezone:   "UTC",
			FirstName:  domain.NullableString{String: "John", IsNull: false},
			LastName:   domain.NullableString{String: "Doe", IsNull: false},
			CreatedAt:  now,
		}

		mockRepo.On("UpsertContact", mock.Anything, mock.AnythingOfType("*domain.Contact")).
			Run(func(args mock.Arguments) {
				updatedContact := args.Get(1).(*domain.Contact)
				assert.Equal(t, contact.Email, updatedContact.Email)
				assert.Equal(t, contact.ExternalID, updatedContact.ExternalID)
				assert.Equal(t, now, updatedContact.CreatedAt)
				assert.False(t, updatedContact.UpdatedAt.IsZero())
				assert.True(t, updatedContact.UpdatedAt.After(now) || updatedContact.UpdatedAt.Equal(now))
			}).
			Return(false, nil).Once()

		isNew, err := service.UpsertContact(context.Background(), contact)
		assert.NoError(t, err)
		assert.False(t, isNew)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		contact := &domain.Contact{
			Email:      "invalid-email",
			ExternalID: "ext1",
			Timezone:   "UTC",
			FirstName:  domain.NullableString{String: "Invalid", IsNull: false},
			LastName:   domain.NullableString{String: "Contact", IsNull: false},
		}

		isNew, err := service.UpsertContact(context.Background(), contact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid contact")
		assert.False(t, isNew)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockContactRepository)
		mockLogger := new(MockLogger)
		mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything).Maybe()
		service := NewContactService(mockRepo, mockLogger)

		contact := &domain.Contact{
			Email:      "contact1@example.com",
			ExternalID: "ext1",
			Timezone:   "UTC",
			FirstName:  domain.NullableString{String: "John", IsNull: false},
			LastName:   domain.NullableString{String: "Doe", IsNull: false},
		}

		repoErr := errors.New("repository error")
		mockRepo.On("UpsertContact", mock.Anything, mock.AnythingOfType("*domain.Contact")).
			Return(false, repoErr).Once()

		isNew, err := service.UpsertContact(context.Background(), contact)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upsert contact")
		assert.False(t, isNew)
	})
}
