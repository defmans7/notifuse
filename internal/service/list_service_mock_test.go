package service

import (
	"context"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMockListService(t *testing.T) {
	mockService := NewMockListService()
	ctx := context.Background()
	workspaceID := "test-workspace"
	listID := "test-list"

	// Create a test list
	testList := &domain.List{
		ID:          listID,
		Name:        "Test List",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("GetLists", func(t *testing.T) {
		// Reset error flags
		mockService.ErrToReturn = nil
		mockService.ErrListNotFoundToReturn = false

		// Add a list to the mock service
		mockService.Lists[listID] = testList

		// Test success case
		lists, err := mockService.GetLists(ctx, workspaceID)
		assert.NoError(t, err)
		assert.Len(t, lists, 1)
		assert.Equal(t, testList, lists[0])
		assert.True(t, mockService.GetListsCalled)

		// Test error case
		mockService.ErrToReturn = &domain.ErrListNotFound{}
		_, err = mockService.GetLists(ctx, workspaceID)
		assert.Error(t, err)
	})

	t.Run("GetListByID", func(t *testing.T) {
		// Reset error flags
		mockService.ErrToReturn = nil
		mockService.ErrListNotFoundToReturn = false

		// Add a list to the mock service
		mockService.Lists[listID] = testList

		// Test success case
		list, err := mockService.GetListByID(ctx, workspaceID, listID)
		assert.NoError(t, err)
		assert.Equal(t, testList, list)
		assert.True(t, mockService.GetListByIDCalled)
		assert.Equal(t, listID, mockService.LastListID)

		// Test not found case
		_, err = mockService.GetListByID(ctx, workspaceID, "nonexistent-id")
		assert.Error(t, err)
		assert.IsType(t, &domain.ErrListNotFound{}, err)
	})

	t.Run("CreateList", func(t *testing.T) {
		// Reset error flags
		mockService.ErrToReturn = nil
		mockService.ErrListNotFoundToReturn = false

		// Test success case
		err := mockService.CreateList(ctx, workspaceID, testList)
		assert.NoError(t, err)
		assert.True(t, mockService.CreateListCalled)
		assert.Equal(t, testList, mockService.LastListCreated)
		assert.Equal(t, testList, mockService.Lists[listID])

		// Test error case
		mockService.ErrToReturn = &domain.ErrListNotFound{}
		err = mockService.CreateList(ctx, workspaceID, testList)
		assert.Error(t, err)
	})

	t.Run("UpdateList", func(t *testing.T) {
		// Reset error flags
		mockService.ErrToReturn = nil
		mockService.ErrListNotFoundToReturn = false

		// Add a list to the mock service
		mockService.Lists[listID] = testList

		// Test success case
		updatedList := &domain.List{
			ID:          listID,
			Name:        "Updated List",
			Description: "Updated Description",
		}
		err := mockService.UpdateList(ctx, workspaceID, updatedList)
		assert.NoError(t, err)
		assert.True(t, mockService.UpdateListCalled)
		assert.Equal(t, updatedList, mockService.LastListUpdated)
		assert.Equal(t, "Updated List", mockService.Lists[listID].Name)
		assert.Equal(t, "Updated Description", mockService.Lists[listID].Description)

		// Test not found case
		err = mockService.UpdateList(ctx, workspaceID, &domain.List{ID: "nonexistent-id"})
		assert.Error(t, err)
		assert.IsType(t, &domain.ErrListNotFound{}, err)
	})

	t.Run("DeleteList", func(t *testing.T) {
		// Reset error flags
		mockService.ErrToReturn = nil
		mockService.ErrListNotFoundToReturn = false

		// Add a list to the mock service
		mockService.Lists[listID] = testList

		// Test success case
		err := mockService.DeleteList(ctx, workspaceID, listID)
		assert.NoError(t, err)
		assert.True(t, mockService.DeleteListCalled)
		assert.Equal(t, listID, mockService.LastListDeleted)
		_, exists := mockService.Lists[listID]
		assert.False(t, exists)

		// Test not found case
		err = mockService.DeleteList(ctx, workspaceID, "nonexistent-id")
		assert.Error(t, err)
		assert.IsType(t, &domain.ErrListNotFound{}, err)
	})
}
