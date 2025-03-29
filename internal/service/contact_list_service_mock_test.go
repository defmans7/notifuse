package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockContactListService(t *testing.T) {
	mockService := &MockContactListService{}
	ctx := context.Background()
	workspaceID := "test-workspace"
	email := "test@example.com"
	listID := "test-list"

	// Test contact list for mock responses
	testContactList := &domain.ContactList{
		Email:     email,
		ListID:    listID,
		Status:    domain.ContactListStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("AddContactToList", func(t *testing.T) {
		err := mockService.AddContactToList(ctx, workspaceID, testContactList)
		assert.NoError(t, err)
		assert.True(t, mockService.AddContactToListCalled)
	})

	t.Run("GetContactListByIDs", func(t *testing.T) {
		mockService.ContactList = testContactList
		contactList, err := mockService.GetContactListByIDs(ctx, workspaceID, email, listID)
		assert.NoError(t, err)
		assert.Equal(t, testContactList, contactList)
		assert.True(t, mockService.GetContactListByIDsCalled)
	})

	t.Run("GetContactsByListID", func(t *testing.T) {
		mockService.ContactLists = []*domain.ContactList{testContactList}
		contactLists, err := mockService.GetContactsByListID(ctx, workspaceID, listID)
		assert.NoError(t, err)
		assert.Equal(t, mockService.ContactLists, contactLists)
		assert.True(t, mockService.GetContactsByListCalled)
	})

	t.Run("GetListsByEmail", func(t *testing.T) {
		mockService.ContactLists = []*domain.ContactList{testContactList}
		contactLists, err := mockService.GetListsByEmail(ctx, workspaceID, email)
		assert.NoError(t, err)
		assert.Equal(t, mockService.ContactLists, contactLists)
		assert.True(t, mockService.GetListsByEmailCalled)
	})

	t.Run("UpdateContactListStatus", func(t *testing.T) {
		err := mockService.UpdateContactListStatus(ctx, workspaceID, email, listID, domain.ContactListStatusActive)
		assert.NoError(t, err)
		assert.True(t, mockService.UpdateContactListCalled)
	})

	t.Run("RemoveContactFromList", func(t *testing.T) {
		err := mockService.RemoveContactFromList(ctx, workspaceID, email, listID)
		assert.NoError(t, err)
		assert.True(t, mockService.RemoveContactFromListCalled)
	})

	t.Run("Error handling", func(t *testing.T) {
		testError := errors.New("test error")
		mockService.ErrToReturn = testError

		// Test error in AddContactToList
		err := mockService.AddContactToList(ctx, workspaceID, testContactList)
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		// Test error in GetContactListByIDs
		contactList, err := mockService.GetContactListByIDs(ctx, workspaceID, email, listID)
		assert.Error(t, err)
		assert.Nil(t, contactList)
		assert.Equal(t, testError, err)

		// Test error in GetContactsByListID
		contactLists, err := mockService.GetContactsByListID(ctx, workspaceID, listID)
		assert.Error(t, err)
		assert.Nil(t, contactLists)
		assert.Equal(t, testError, err)

		// Test error in GetListsByEmail
		contactLists, err = mockService.GetListsByEmail(ctx, workspaceID, email)
		assert.Error(t, err)
		assert.Nil(t, contactLists)
		assert.Equal(t, testError, err)

		// Test error in UpdateContactListStatus
		err = mockService.UpdateContactListStatus(ctx, workspaceID, email, listID, domain.ContactListStatusActive)
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		// Test error in RemoveContactFromList
		err = mockService.RemoveContactFromList(ctx, workspaceID, email, listID)
		assert.Error(t, err)
		assert.Equal(t, testError, err)
	})
}
