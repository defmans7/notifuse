package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestContactService_GetContactByEmail(t *testing.T) {
	mockRepo := new(repository.MockContactRepository)
	mockAuthService := new(MockAuthService)
	mockWorkspaceRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	email := "test@example.com"
	contact := &domain.Contact{
		Email: email,
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactByEmail", ctx, email, workspaceID).Return(contact, nil)

		result, err := service.GetContactByEmail(ctx, email, workspaceID)
		assert.NoError(t, err)
		assert.Equal(t, contact, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContactByEmail(ctx, email, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactByEmail", ctx, email, workspaceID).Return(nil, &domain.ErrContactNotFound{})

		result, err := service.GetContactByEmail(ctx, email, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactByEmail", ctx, email, workspaceID).Return(nil, errors.New("repo error"))
		mockLogger.On("WithField", "email", email).Return(mockLogger)
		mockLogger.On("Error", "Failed to get contact by email: repo error").Return()

		result, err := service.GetContactByEmail(ctx, email, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactService_GetContactByExternalID(t *testing.T) {
	mockRepo := new(repository.MockContactRepository)
	mockAuthService := new(MockAuthService)
	mockWorkspaceRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	externalID := "ext123"
	contact := &domain.Contact{
		ExternalID: &domain.NullableString{String: externalID, IsNull: false},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactByExternalID", ctx, externalID, workspaceID).Return(contact, nil)

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.NoError(t, err)
		assert.Equal(t, contact, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactByExternalID", ctx, externalID, workspaceID).Return(nil, &domain.ErrContactNotFound{})

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContactByExternalID", ctx, externalID, workspaceID).Return(nil, errors.New("repo error"))
		mockLogger.On("WithField", "external_id", externalID).Return(mockLogger)
		mockLogger.On("Error", "Failed to get contact by external ID: repo error").Return()

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactService_GetContacts(t *testing.T) {
	mockRepo := new(repository.MockContactRepository)
	mockAuthService := new(MockAuthService)
	mockWorkspaceRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	req := &domain.GetContactsRequest{
		WorkspaceID: workspaceID,
	}
	response := &domain.GetContactsResponse{
		Contacts: []*domain.Contact{
			{Email: "test1@example.com"},
			{Email: "test2@example.com"},
		},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContacts", ctx, req).Return(response, nil)

		result, err := service.GetContacts(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, response, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContacts(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("GetContacts", ctx, req).Return(nil, errors.New("repo error"))
		mockLogger.On("Error", "Failed to get contacts: repo error").Return()

		result, err := service.GetContacts(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactService_DeleteContact(t *testing.T) {
	mockRepo := new(repository.MockContactRepository)
	mockAuthService := new(MockAuthService)
	mockWorkspaceRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	email := "test@example.com"

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("DeleteContact", ctx, email, workspaceID).Return(nil)

		err := service.DeleteContact(ctx, email, workspaceID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.DeleteContact(ctx, email, workspaceID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("DeleteContact", ctx, email, workspaceID).Return(errors.New("repo error"))
		mockLogger.On("WithField", "email", email).Return(mockLogger)
		mockLogger.On("Error", "Failed to delete contact: repo error").Return()

		err := service.DeleteContact(ctx, email, workspaceID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactService_BatchImportContacts(t *testing.T) {
	mockRepo := new(repository.MockContactRepository)
	mockAuthService := new(MockAuthService)
	mockWorkspaceRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	contacts := []*domain.Contact{
		{
			Email: "test1@example.com",
		},
		{
			Email: "test2@example.com",
		},
	}

	t.Run("successful import", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("BatchImportContacts", ctx, workspaceID, mock.MatchedBy(func(contacts []*domain.Contact) bool {
			return len(contacts) == 2 &&
				contacts[0].Email == "test1@example.com" &&
				contacts[1].Email == "test2@example.com" &&
				!contacts[0].CreatedAt.IsZero() &&
				!contacts[1].CreatedAt.IsZero()
		})).Return(nil)

		err := service.BatchImportContacts(ctx, workspaceID, contacts)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		err := service.BatchImportContacts(ctx, workspaceID, contacts)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("invalid contact", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		invalidContacts := []*domain.Contact{
			{
				Email: "", // Invalid email
			},
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)

		err := service.BatchImportContacts(ctx, workspaceID, invalidContacts)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("BatchImportContacts", ctx, workspaceID, mock.Anything).Return(errors.New("repo error"))
		mockLogger.On("WithField", "contacts_count", len(contacts)).Return(mockLogger)
		mockLogger.On("Error", "Failed to batch import contacts: repo error").Return()

		err := service.BatchImportContacts(ctx, workspaceID, contacts)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestContactService_UpsertContact(t *testing.T) {
	mockRepo := new(repository.MockContactRepository)
	mockAuthService := new(MockAuthService)
	mockWorkspaceRepo := new(repository.MockWorkspaceRepository)
	mockLogger := new(MockLogger)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	contact := &domain.Contact{
		Email: "test@example.com",
	}

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("UpsertContact", ctx, workspaceID, mock.MatchedBy(func(contact *domain.Contact) bool {
			return contact.Email == "test@example.com" &&
				!contact.CreatedAt.IsZero() &&
				!contact.UpdatedAt.IsZero()
		})).Return(true, nil)

		created, err := service.UpsertContact(ctx, workspaceID, contact)
		assert.NoError(t, err)
		assert.True(t, created)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("successful update", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		existingContact := &domain.Contact{
			Email:     "test@example.com",
			CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("UpsertContact", ctx, workspaceID, mock.MatchedBy(func(contact *domain.Contact) bool {
			return contact.Email == "test@example.com" &&
				contact.CreatedAt.Equal(existingContact.CreatedAt) &&
				!contact.UpdatedAt.IsZero()
		})).Return(false, nil)

		created, err := service.UpsertContact(ctx, workspaceID, existingContact)
		assert.NoError(t, err)
		assert.False(t, created)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, errors.New("auth error"))

		created, err := service.UpsertContact(ctx, workspaceID, contact)
		assert.Error(t, err)
		assert.False(t, created)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("invalid contact", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		invalidContact := &domain.Contact{
			Email: "", // Invalid email
		}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)

		created, err := service.UpsertContact(ctx, workspaceID, invalidContact)
		assert.Error(t, err)
		assert.False(t, created)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		mockAuthService.Mock = mock.Mock{}
		mockWorkspaceRepo.Mock = mock.Mock{}
		mockLogger.Mock = mock.Mock{}

		mockAuthService.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.On("UpsertContact", ctx, workspaceID, mock.Anything).Return(false, errors.New("repo error"))
		mockLogger.On("WithField", "email", contact.Email).Return(mockLogger)
		mockLogger.On("Error", "Failed to upsert contact: repo error").Return()

		created, err := service.UpsertContact(ctx, workspaceID, contact)
		assert.Error(t, err)
		assert.False(t, created)
		mockRepo.AssertExpectations(t)
		mockAuthService.AssertExpectations(t)
		mockWorkspaceRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
