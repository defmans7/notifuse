package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// UUID constants for testing that pass validation
const (
	testContactUUID1 = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	testContactUUID2 = "123e4567-e89b-12d3-a456-426614174000"
)

type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) CreateContact(ctx context.Context, contact *domain.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepository) GetContactByUUID(ctx context.Context, uuid string) (*domain.Contact, error) {
	args := m.Called(ctx, uuid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Contact), args.Error(1)
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

func (m *MockContactRepository) GetContacts(ctx context.Context) ([]*domain.Contact, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Contact), args.Error(1)
}

func (m *MockContactRepository) UpdateContact(ctx context.Context, contact *domain.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepository) DeleteContact(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *MockContactRepository) BatchImportContacts(ctx context.Context, contacts []*domain.Contact) error {
	args := m.Called(ctx, contacts)
	return args.Error(0)
}

// Tests begin here

func TestContactService_CreateContact(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should create contact successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			UUID:       testContactUUID1,
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
			FirstName:  "Test",
			LastName:   "Contact",
		}

		mockRepo.On("CreateContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			// Verify timestamps are set
			return c.UUID == contact.UUID && c.Email == contact.Email &&
				!c.CreatedAt.IsZero() && !c.UpdatedAt.IsZero()
		})).Return(nil).Once()

		// Act
		err := service.CreateContact(ctx, contact)

		// Assert
		assert.NoError(t, err)
		assert.False(t, contact.CreatedAt.IsZero())
		assert.False(t, contact.UpdatedAt.IsZero())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should generate UUID if not provided", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
			FirstName:  "Test",
			LastName:   "Contact",
		}

		mockRepo.On("CreateContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			return c.UUID != "" && c.Email == contact.Email
		})).Return(nil).Once()

		// Act
		err := service.CreateContact(ctx, contact)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, contact.UUID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if contact is invalid", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			UUID: "invalid-uuid", // Invalid UUID format
			// Email missing, which is required
		}

		// Act
		err := service.CreateContact(ctx, contact)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertNotCalled(t, "CreateContact")
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			UUID:       testContactUUID1,
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
		}

		repoErr := errors.New("repository error")
		mockRepo.On("CreateContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			return c.UUID == contact.UUID && c.Email == contact.Email
		})).Return(repoErr).Once()

		// Act
		err := service.CreateContact(ctx, contact)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create contact")
		mockRepo.AssertExpectations(t)
	})
}

func TestContactService_GetContactByUUID(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should get contact by UUID successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		uuid := testContactUUID1
		expectedContact := &domain.Contact{
			UUID:       uuid,
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
			FirstName:  "Test",
			LastName:   "Contact",
		}

		mockRepo.On("GetContactByUUID", ctx, uuid).Return(expectedContact, nil).Once()

		// Act
		contact, err := service.GetContactByUUID(ctx, uuid)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedContact, contact)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return not found error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		uuid := testContactUUID2
		notFoundErr := &domain.ErrContactNotFound{Message: "contact not found"}

		mockRepo.On("GetContactByUUID", ctx, uuid).Return(nil, notFoundErr).Once()

		// Act
		contact, err := service.GetContactByUUID(ctx, uuid)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contact)
		assert.IsType(t, &domain.ErrContactNotFound{}, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		uuid := testContactUUID1
		repoErr := errors.New("repository error")

		mockRepo.On("GetContactByUUID", ctx, uuid).Return(nil, repoErr).Once()

		// Act
		contact, err := service.GetContactByUUID(ctx, uuid)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, contact)
		assert.Contains(t, err.Error(), "failed to get contact")
		mockRepo.AssertExpectations(t)
	})
}

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
			UUID:       testContactUUID1,
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
			UUID:       testContactUUID1,
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
				UUID:       testContactUUID1,
				Email:      "contact1@example.com",
				ExternalID: "ext-1",
				Timezone:   "UTC",
				FirstName:  "Contact",
				LastName:   "One",
			},
			{
				UUID:       testContactUUID2,
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

func TestContactService_UpdateContact(t *testing.T) {
	mockRepo := new(MockContactRepository)
	mockLogger := new(MockLogger)
	mockLogger.On("WithField", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything).Maybe()

	service := NewContactService(mockRepo, mockLogger)

	t.Run("should update contact successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			UUID:       testContactUUID1,
			Email:      "updated@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
			FirstName:  "Updated",
			LastName:   "Name",
		}

		mockRepo.On("UpdateContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			return c.UUID == contact.UUID &&
				c.Email == contact.Email &&
				c.FirstName == contact.FirstName &&
				!c.UpdatedAt.IsZero()
		})).Return(nil).Once()

		// Act
		err := service.UpdateContact(ctx, contact)

		// Assert
		assert.NoError(t, err)
		assert.False(t, contact.UpdatedAt.IsZero())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if contact is invalid", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			UUID: "invalid-uuid", // Invalid UUID format
			// Email missing, which is required
		}

		// Act
		err := service.UpdateContact(ctx, contact)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertNotCalled(t, "UpdateContact")
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		contact := &domain.Contact{
			UUID:       testContactUUID1,
			Email:      "test@example.com",
			ExternalID: "test-external-id",
			Timezone:   "UTC",
		}

		repoErr := errors.New("repository error")
		mockRepo.On("UpdateContact", ctx, mock.MatchedBy(func(c *domain.Contact) bool {
			return c.UUID == contact.UUID && c.Email == contact.Email
		})).Return(repoErr).Once()

		// Act
		err := service.UpdateContact(ctx, contact)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update contact")
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
		uuid := testContactUUID1

		mockRepo.On("DeleteContact", ctx, uuid).Return(nil).Once()

		// Act
		err := service.DeleteContact(ctx, uuid)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if repository fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		uuid := testContactUUID1
		repoErr := errors.New("repository error")

		mockRepo.On("DeleteContact", ctx, uuid).Return(repoErr).Once()

		// Act
		err := service.DeleteContact(ctx, uuid)

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

				// Verify UUIDs and timestamps were set
				for _, contact := range importedContacts {
					assert.NotEmpty(t, contact.UUID)
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
