package service

import (
	"context"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockContactService(t *testing.T) {
	mockService := NewMockContactService()
	ctx := context.Background()
	workspaceID := "test-workspace"
	email := "test@example.com"

	// Create a test contact
	testContact := &domain.Contact{
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("GetContacts", func(t *testing.T) {
		// Reset mock service state
		mockService.ErrToReturn = nil
		mockService.ErrContactNotFoundToReturn = false

		// Add a contact to the mock service
		mockService.Contacts[email] = testContact

		req := &domain.GetContactsRequest{
			WorkspaceID: workspaceID,
		}

		// Test success case
		resp, err := mockService.GetContacts(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, resp.Contacts, 1)
		assert.Equal(t, testContact, resp.Contacts[0])
		assert.True(t, mockService.GetContactsCalled)

		// Test error case
		mockService.ErrToReturn = &domain.ErrContactNotFound{}
		_, err = mockService.GetContacts(ctx, req)
		assert.Error(t, err)
	})

	t.Run("GetContactByEmail", func(t *testing.T) {
		// Reset mock service state
		mockService.ErrToReturn = nil
		mockService.ErrContactNotFoundToReturn = false

		// Add a contact to the mock service
		mockService.Contacts[email] = testContact

		// Test success case
		contact, err := mockService.GetContactByEmail(ctx, workspaceID, email)
		assert.NoError(t, err)
		assert.Equal(t, testContact, contact)
		assert.True(t, mockService.GetContactByEmailCalled)
		assert.Equal(t, email, mockService.LastContactEmail)

		// Test not found case
		_, err = mockService.GetContactByEmail(ctx, workspaceID, "nonexistent@example.com")
		assert.Error(t, err)
		assert.IsType(t, &domain.ErrContactNotFound{}, err)
	})

	t.Run("GetContactByExternalID", func(t *testing.T) {
		// Reset mock service state
		mockService.ErrToReturn = nil
		mockService.ErrContactNotFoundToReturn = false

		externalID := "ext-123"
		testContact.ExternalID = &domain.NullableString{String: externalID, IsNull: false}
		mockService.Contacts[email] = testContact

		// Test success case
		contact, err := mockService.GetContactByExternalID(ctx, workspaceID, externalID)
		assert.NoError(t, err)
		assert.Equal(t, testContact, contact)
		assert.True(t, mockService.GetContactByExternalIDCalled)
		assert.Equal(t, externalID, mockService.LastContactExternalID)

		// Test not found case
		_, err = mockService.GetContactByExternalID(ctx, workspaceID, "nonexistent-id")
		assert.Error(t, err)
		assert.IsType(t, &domain.ErrContactNotFound{}, err)
	})

	t.Run("DeleteContact", func(t *testing.T) {
		// Reset mock service state
		mockService.ErrToReturn = nil
		mockService.ErrContactNotFoundToReturn = false

		// Add a contact to the mock service
		mockService.Contacts[email] = testContact

		// Test success case
		err := mockService.DeleteContact(ctx, workspaceID, email)
		assert.NoError(t, err)
		assert.True(t, mockService.DeleteContactCalled)
		assert.Equal(t, email, mockService.LastContactEmail)
		_, exists := mockService.Contacts[email]
		assert.False(t, exists)

		// Test not found case
		err = mockService.DeleteContact(ctx, workspaceID, "nonexistent@example.com")
		assert.Error(t, err)
		assert.IsType(t, &domain.ErrContactNotFound{}, err)
	})

	t.Run("BatchImportContacts", func(t *testing.T) {
		// Reset mock service state
		mockService.ErrToReturn = nil
		mockService.ErrContactNotFoundToReturn = false

		contacts := []*domain.Contact{testContact}

		// Test success case
		err := mockService.BatchImportContacts(ctx, workspaceID, contacts)
		assert.NoError(t, err)
		assert.True(t, mockService.BatchImportContactsCalled)
		assert.Equal(t, contacts, mockService.LastContactsBatchImported)
		assert.Equal(t, testContact, mockService.Contacts[email])

		// Test error case
		mockService.ErrToReturn = &domain.ErrContactNotFound{}
		err = mockService.BatchImportContacts(ctx, workspaceID, contacts)
		assert.Error(t, err)
	})

	t.Run("UpsertContact", func(t *testing.T) {
		// Reset mock service state
		mockService.ErrToReturn = nil
		mockService.ErrContactNotFoundToReturn = false
		mockService.UpsertIsNewToReturn = true

		// Test create new contact
		isNew, err := mockService.UpsertContact(ctx, workspaceID, testContact)
		assert.NoError(t, err)
		assert.True(t, isNew)
		assert.True(t, mockService.UpsertContactCalled)
		assert.Equal(t, testContact, mockService.LastContactUpserted)
		assert.Equal(t, testContact, mockService.Contacts[email])

		// Test update existing contact
		mockService.UpsertIsNewToReturn = false
		updatedContact := &domain.Contact{
			Email: email,
			FirstName: &domain.NullableString{
				String: "John",
				IsNull: false,
			},
		}
		isNew, err = mockService.UpsertContact(ctx, workspaceID, updatedContact)
		assert.NoError(t, err)
		assert.False(t, isNew)
		assert.Equal(t, "John", mockService.Contacts[email].FirstName.String)

		// Test error case
		mockService.ErrToReturn = &domain.ErrContactNotFound{}
		_, err = mockService.UpsertContact(ctx, workspaceID, testContact)
		assert.Error(t, err)
	})
}
