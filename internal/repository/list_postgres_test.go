package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
)

func TestCreateList(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewListRepository(db)

	// Test case 1: Successful list creation
	list := &domain.List{
		ID:            "list-123",
		Name:          "Test List",
		Type:          "default",
		IsDoubleOptin: true,
		Description:   "A test list",
	}

	mock.ExpectExec(`INSERT INTO lists`).
		WithArgs(
			list.ID, list.Name, list.Type, list.IsDoubleOptin, list.Description,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateList(context.Background(), list)
	require.NoError(t, err)

	// Test case 2: Error during list creation
	listWithError := &domain.List{
		ID:          "list-error",
		Name:        "Error List",
		Type:        "default",
		Description: "A list that will cause an error",
	}

	mock.ExpectExec(`INSERT INTO lists`).
		WithArgs(
			listWithError.ID, listWithError.Name, listWithError.Type, listWithError.IsDoubleOptin,
			listWithError.Description, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(errors.New("database error"))

	err = repo.CreateList(context.Background(), listWithError)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create list")
}

func TestGetListByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewListRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	listID := "list-123"

	// Test case 1: List found
	rows := sqlmock.NewRows([]string{
		"id", "name", "type", "is_double_optin", "description", "created_at", "updated_at",
	}).
		AddRow(
			listID, "Test List", "default", true, "A test list", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM lists WHERE id = \$1`).
		WithArgs(listID).
		WillReturnRows(rows)

	list, err := repo.GetListByID(context.Background(), listID)
	require.NoError(t, err)
	assert.Equal(t, listID, list.ID)
	assert.Equal(t, "Test List", list.Name)
	assert.Equal(t, "default", list.Type)
	assert.True(t, list.IsDoubleOptin)

	// Test case 2: List not found
	mock.ExpectQuery(`SELECT (.+) FROM lists WHERE id = \$1`).
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	list, err = repo.GetListByID(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrListNotFound{}, err)
	assert.Nil(t, list)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM lists WHERE id = \$1`).
		WithArgs("error-id").
		WillReturnError(errors.New("database error"))

	list, err = repo.GetListByID(context.Background(), "error-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get list")
	assert.Nil(t, list)
}

func TestGetLists(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewListRepository(db)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Test case 1: Multiple lists found
	rows := sqlmock.NewRows([]string{
		"id", "name", "type", "is_double_optin", "description", "created_at", "updated_at",
	}).
		AddRow(
			"list-1", "First List", "default", true, "First test list", now, now,
		).
		AddRow(
			"list-2", "Second List", "custom", false, "Second test list", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM lists ORDER BY created_at DESC`).
		WillReturnRows(rows)

	lists, err := repo.GetLists(context.Background())
	require.NoError(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, "list-1", lists[0].ID)
	assert.Equal(t, "list-2", lists[1].ID)

	// Test case 2: No lists found (empty result)
	emptyRows := sqlmock.NewRows([]string{
		"id", "name", "type", "is_double_optin", "description", "created_at", "updated_at",
	})

	mock.ExpectQuery(`SELECT (.+) FROM lists ORDER BY created_at DESC`).
		WillReturnRows(emptyRows)

	lists, err = repo.GetLists(context.Background())
	require.NoError(t, err)
	assert.Empty(t, lists)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM lists ORDER BY created_at DESC`).
		WillReturnError(errors.New("database error"))

	lists, err = repo.GetLists(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get lists")
	assert.Nil(t, lists)
}

func TestUpdateList(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewListRepository(db)
	listID := "list-123"

	// Test case 1: Successful update
	list := &domain.List{
		ID:            listID,
		Name:          "Updated List",
		Type:          "custom",
		IsDoubleOptin: false,
		Description:   "An updated list",
	}

	mock.ExpectExec(`UPDATE lists SET (.+) WHERE id = \$6`).
		WithArgs(
			list.Name, list.Type, list.IsDoubleOptin, list.Description, sqlmock.AnyArg(), list.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateList(context.Background(), list)
	require.NoError(t, err)

	// Test case 2: List not found
	nonExistentList := &domain.List{
		ID:            "non-existent-id",
		Name:          "Non-existent List",
		Type:          "default",
		IsDoubleOptin: true,
		Description:   "A non-existent list",
	}

	mock.ExpectExec(`UPDATE lists SET (.+) WHERE id = \$6`).
		WithArgs(
			nonExistentList.Name, nonExistentList.Type, nonExistentList.IsDoubleOptin,
			nonExistentList.Description, sqlmock.AnyArg(), nonExistentList.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateList(context.Background(), nonExistentList)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrListNotFound{}, err)
}

func TestDeleteList(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewListRepository(db)
	listID := "list-123"

	// Test case 1: Successful deletion
	mock.ExpectExec(`DELETE FROM lists WHERE id = \$1`).
		WithArgs(listID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteList(context.Background(), listID)
	require.NoError(t, err)

	// Test case 2: List not found
	mock.ExpectExec(`DELETE FROM lists WHERE id = \$1`).
		WithArgs("non-existent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteList(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrListNotFound{}, err)

	// Test case 3: Database error
	mock.ExpectExec(`DELETE FROM lists WHERE id = \$1`).
		WithArgs("error-id").
		WillReturnError(errors.New("database error"))

	err = repo.DeleteList(context.Background(), "error-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete list")
}
