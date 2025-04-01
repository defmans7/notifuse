package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestContactService_GetContactByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockContactRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	email := "test@example.com"
	contact := &domain.Contact{
		Email: email,
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContactByEmail(ctx, workspaceID, email).Return(contact, nil)

		result, err := service.GetContactByEmail(ctx, workspaceID, email)
		assert.NoError(t, err)
		assert.Equal(t, contact, result)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContactByEmail(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContactByEmail(ctx, workspaceID, email).Return(nil, fmt.Errorf("contact not found"))

		result, err := service.GetContactByEmail(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContactByEmail(ctx, workspaceID, email).Return(nil, errors.New("repo error"))
		mockLogger.EXPECT().WithField("email", email).Return(mockLogger)
		mockLogger.EXPECT().Error("Failed to get contact by email: repo error")

		result, err := service.GetContactByEmail(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestContactService_GetContactByExternalID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockContactRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	externalID := "ext123"
	contact := &domain.Contact{
		ExternalID: &domain.NullableString{String: externalID, IsNull: false},
	}

	t.Run("successful retrieval", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContactByExternalID(ctx, externalID, workspaceID).Return(contact, nil)

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.NoError(t, err)
		assert.Equal(t, contact, result)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContactByExternalID(ctx, externalID, workspaceID).Return(nil, fmt.Errorf("contact not found"))

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContactByExternalID(ctx, externalID, workspaceID).Return(nil, errors.New("repo error"))
		mockLogger.EXPECT().WithField("external_id", externalID).Return(mockLogger)
		mockLogger.EXPECT().Error("Failed to get contact by external ID: repo error")

		result, err := service.GetContactByExternalID(ctx, externalID, workspaceID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestContactService_GetContacts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockContactRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

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
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContacts(ctx, req).Return(response, nil)

		result, err := service.GetContacts(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, response, result)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))

		result, err := service.GetContacts(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().GetContacts(ctx, req).Return(nil, errors.New("repo error"))
		mockLogger.EXPECT().Error("Failed to get contacts: repo error")

		result, err := service.GetContacts(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestContactService_DeleteContact(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	service := NewContactService(mockContactRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "test-workspace"
	email := "test@example.com"

	t.Run("contact not found", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockLogger.EXPECT().WithField("email", email).Return(mockLogger)
		mockContactRepo.EXPECT().DeleteContact(ctx, email, workspaceID).Return(fmt.Errorf("contact not found"))
		mockLogger.EXPECT().Error(fmt.Sprintf("Failed to delete contact: %v", fmt.Errorf("contact not found")))

		err := service.DeleteContact(ctx, email, workspaceID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete contact")
	})
}

func TestContactService_UpsertContact(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockContactRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"
	contact := &domain.Contact{
		Email: "test@example.com",
	}

	t.Run("successful create", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpsertContact(ctx, workspaceID, contact).Return(true, nil)

		result := service.UpsertContact(ctx, workspaceID, contact)
		assert.Equal(t, domain.UpsertContactOperationCreate, result.Action)
		assert.Empty(t, result.Error)
	})

	t.Run("successful update", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpsertContact(ctx, workspaceID, contact).Return(false, nil)

		result := service.UpsertContact(ctx, workspaceID, contact)
		assert.Equal(t, domain.UpsertContactOperationUpdate, result.Action)
		assert.Empty(t, result.Error)
	})

	t.Run("authentication error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(nil, errors.New("auth error"))
		mockLogger.EXPECT().WithField("email", contact.Email).Return(mockLogger)
		mockLogger.EXPECT().Error("Failed to authenticate user: auth error")

		result := service.UpsertContact(ctx, workspaceID, contact)
		assert.Equal(t, domain.UpsertContactOperationError, result.Action)
		assert.Contains(t, result.Error, "auth error")
	})

	t.Run("repository error", func(t *testing.T) {
		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpsertContact(ctx, workspaceID, contact).Return(false, errors.New("repo error"))
		mockLogger.EXPECT().WithField("email", contact.Email).Return(mockLogger)
		mockLogger.EXPECT().Error("Failed to upsert contact: repo error")

		result := service.UpsertContact(ctx, workspaceID, contact)
		assert.Equal(t, domain.UpsertContactOperationError, result.Action)
		assert.Contains(t, result.Error, "repo error")
	})
}

func TestContactService_UpsertContactWithPartialUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockContactRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	service := NewContactService(mockRepo, mockWorkspaceRepo, mockAuthService, mockLogger)

	ctx := context.Background()
	workspaceID := "workspace123"

	t.Run("upsert with only email", func(t *testing.T) {
		// Create a contact with only email
		minimalContact := &domain.Contact{
			Email: "minimal@example.com",
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).DoAndReturn(
			func(ctx context.Context, workspaceID string, contact *domain.Contact) (bool, error) {
				// Verify that contact has CreatedAt and UpdatedAt set
				assert.NotZero(t, contact.CreatedAt.Unix())
				assert.NotZero(t, contact.UpdatedAt.Unix())
				assert.Equal(t, "minimal@example.com", contact.Email)
				return true, nil
			})

		result := service.UpsertContact(ctx, workspaceID, minimalContact)
		assert.Equal(t, domain.UpsertContactOperationCreate, result.Action)
		assert.Empty(t, result.Error)
	})

	t.Run("upsert with partial fields", func(t *testing.T) {
		// Create a contact with partial fields
		partialContact := &domain.Contact{
			Email:     "partial@example.com",
			FirstName: &domain.NullableString{String: "Jane", IsNull: false},
			LastName:  &domain.NullableString{String: "Smith", IsNull: false},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).DoAndReturn(
			func(ctx context.Context, workspaceID string, contact *domain.Contact) (bool, error) {
				// Verify that only the specified fields are set
				assert.NotZero(t, contact.CreatedAt.Unix())
				assert.NotZero(t, contact.UpdatedAt.Unix())
				assert.Equal(t, "partial@example.com", contact.Email)
				assert.Equal(t, "Jane", contact.FirstName.String)
				assert.Equal(t, "Smith", contact.LastName.String)
				assert.False(t, contact.FirstName.IsNull)
				assert.False(t, contact.LastName.IsNull)
				// Other fields should be nil
				assert.Nil(t, contact.ExternalID)
				assert.Nil(t, contact.Phone)
				assert.Nil(t, contact.CustomJSON1)
				return false, nil
			})

		result := service.UpsertContact(ctx, workspaceID, partialContact)
		assert.Equal(t, domain.UpsertContactOperationUpdate, result.Action)
		assert.Empty(t, result.Error)
	})

	t.Run("upsert with custom JSON", func(t *testing.T) {
		// Create a contact with custom JSON fields
		jsonData := map[string]interface{}{
			"preference": "email",
			"frequency":  "weekly",
		}
		jsonContact := &domain.Contact{
			Email:       "json@example.com",
			CustomJSON1: &domain.NullableJSON{Data: jsonData, IsNull: false},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).DoAndReturn(
			func(ctx context.Context, workspaceID string, contact *domain.Contact) (bool, error) {
				// Verify that JSON field is properly set
				assert.NotZero(t, contact.CreatedAt.Unix())
				assert.NotZero(t, contact.UpdatedAt.Unix())
				assert.Equal(t, "json@example.com", contact.Email)
				assert.NotNil(t, contact.CustomJSON1)
				assert.Equal(t, jsonData, contact.CustomJSON1.Data)
				assert.False(t, contact.CustomJSON1.IsNull)
				// Other fields should be nil
				assert.Nil(t, contact.FirstName)
				assert.Nil(t, contact.LastName)
				return true, nil
			})

		result := service.UpsertContact(ctx, workspaceID, jsonContact)
		assert.Equal(t, domain.UpsertContactOperationCreate, result.Action)
		assert.Empty(t, result.Error)
	})

	t.Run("upsert with explicit null field", func(t *testing.T) {
		// Create a contact with some fields explicitly set to null
		contactWithNulls := &domain.Contact{
			Email:       "null@example.com",
			FirstName:   &domain.NullableString{String: "", IsNull: true},
			CustomJSON1: &domain.NullableJSON{Data: nil, IsNull: true},
		}

		mockAuthService.EXPECT().AuthenticateUserForWorkspace(ctx, workspaceID).Return(&domain.User{}, nil)
		mockRepo.EXPECT().UpsertContact(ctx, workspaceID, gomock.Any()).DoAndReturn(
			func(ctx context.Context, workspaceID string, contact *domain.Contact) (bool, error) {
				// Verify that null fields are properly set
				assert.NotZero(t, contact.CreatedAt.Unix())
				assert.NotZero(t, contact.UpdatedAt.Unix())
				assert.Equal(t, "null@example.com", contact.Email)
				assert.NotNil(t, contact.FirstName)
				assert.True(t, contact.FirstName.IsNull)
				assert.NotNil(t, contact.CustomJSON1)
				assert.True(t, contact.CustomJSON1.IsNull)
				assert.Nil(t, contact.CustomJSON1.Data)
				// Other fields should be nil
				assert.Nil(t, contact.LastName)
				return false, nil
			})

		result := service.UpsertContact(ctx, workspaceID, contactWithNulls)
		assert.Equal(t, domain.UpsertContactOperationUpdate, result.Action)
		assert.Empty(t, result.Error)
	})
}
