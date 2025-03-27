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
	"github.com/Notifuse/notifuse/internal/repository/testutil"
)

func TestCreateList(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	repo := NewListRepository(db, workspaceRepo)

	// Test case 1: Successful list creation
	list := &domain.List{
		ID:            "list123",
		Name:          "Test List",
		Type:          "public",
		IsDoubleOptin: true,
		Description:   "This is a test list",
	}

	mock.ExpectExec(`INSERT INTO lists`).
		WithArgs(
			list.ID, list.Name, list.Type, list.IsDoubleOptin, list.Description,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateList(context.Background(), list)
	require.NoError(t, err)

	// Test case 2: Error during insertion
	listWithError := &domain.List{
		ID:            "errorList",
		Name:          "Error List",
		Type:          "public",
		IsDoubleOptin: false,
		Description:   "This list will cause an error",
	}

	mock.ExpectExec(`INSERT INTO lists`).
		WithArgs(
			listWithError.ID, listWithError.Name, listWithError.Type, listWithError.IsDoubleOptin, listWithError.Description,
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnError(errors.New("database error"))

	err = repo.CreateList(context.Background(), listWithError)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create list")
}

func TestGetListByID(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	repo := NewListRepository(db, workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)
	listID := "list123"

	// Test case 1: List found
	rows := sqlmock.NewRows([]string{
		"id", "name", "type", "is_double_optin", "description", "created_at", "updated_at",
	}).
		AddRow(
			listID, "Test List", "public", true, "This is a test list", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM lists WHERE id = \$1`).
		WithArgs(listID).
		WillReturnRows(rows)

	list, err := repo.GetListByID(context.Background(), listID)
	require.NoError(t, err)
	assert.Equal(t, listID, list.ID)
	assert.Equal(t, "Test List", list.Name)
	assert.Equal(t, "public", list.Type)
	assert.Equal(t, true, list.IsDoubleOptin)
	assert.Equal(t, "This is a test list", list.Description)

	// Test case 2: List not found
	mock.ExpectQuery(`SELECT (.+) FROM lists WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	list, err = repo.GetListByID(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrListNotFound{}, err)
	assert.Nil(t, list)

	// Test case 3: Database error
	mock.ExpectQuery(`SELECT (.+) FROM lists WHERE id = \$1`).
		WithArgs("error").
		WillReturnError(errors.New("database error"))

	list, err = repo.GetListByID(context.Background(), "error")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get list")
	assert.Nil(t, list)
}

func TestGetLists(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	repo := NewListRepository(db, workspaceRepo)
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Test case 1: Multiple lists found
	rows := sqlmock.NewRows([]string{
		"id", "name", "type", "is_double_optin", "description", "created_at", "updated_at",
	}).
		AddRow(
			"list1", "Test List 1", "public", true, "Description 1", now, now,
		).
		AddRow(
			"list2", "Test List 2", "private", false, "Description 2", now, now,
		)

	mock.ExpectQuery(`SELECT (.+) FROM lists ORDER BY created_at DESC`).
		WillReturnRows(rows)

	lists, err := repo.GetLists(context.Background())
	require.NoError(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, "list1", lists[0].ID)
	assert.Equal(t, "list2", lists[1].ID)
	assert.Equal(t, "Test List 1", lists[0].Name)
	assert.Equal(t, "Test List 2", lists[1].Name)

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
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	repo := NewListRepository(db, workspaceRepo)

	// Test case 1: Successful update
	list := &domain.List{
		ID:            "list123",
		Name:          "Updated List",
		Type:          "private",
		IsDoubleOptin: true,
		Description:   "This list has been updated",
	}

	mock.ExpectExec(`UPDATE lists SET name = \$1, type = \$2, is_double_optin = \$3, description = \$4, updated_at = \$5 WHERE id = \$6`).
		WithArgs(
			list.Name, list.Type, list.IsDoubleOptin, list.Description,
			sqlmock.AnyArg(), list.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateList(context.Background(), list)
	require.NoError(t, err)

	// Test case 2: List not found
	notFoundList := &domain.List{
		ID:            "nonexistent",
		Name:          "Nonexistent List",
		Type:          "public",
		IsDoubleOptin: false,
		Description:   "This list doesn't exist",
	}

	mock.ExpectExec(`UPDATE lists SET name = \$1, type = \$2, is_double_optin = \$3, description = \$4, updated_at = \$5 WHERE id = \$6`).
		WithArgs(
			notFoundList.Name, notFoundList.Type, notFoundList.IsDoubleOptin, notFoundList.Description,
			sqlmock.AnyArg(), notFoundList.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateList(context.Background(), notFoundList)
	require.Error(t, err)
	assert.IsType(t, &domain.ErrListNotFound{}, err)

	// Test case 3: Database error
	errorList := &domain.List{
		ID:            "error",
		Name:          "Error List",
		Type:          "public",
		IsDoubleOptin: false,
		Description:   "This list causes an error",
	}

	mock.ExpectExec(`UPDATE lists SET name = \$1, type = \$2, is_double_optin = \$3, description = \$4, updated_at = \$5 WHERE id = \$6`).
		WithArgs(
			errorList.Name, errorList.Type, errorList.IsDoubleOptin, errorList.Description,
			sqlmock.AnyArg(), errorList.ID,
		).
		WillReturnError(errors.New("database error"))

	err = repo.UpdateList(context.Background(), errorList)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update list")
}

func TestDeleteList(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	repo := NewListRepository(db, workspaceRepo)
	listID := "list123"

	// Test case 1: Successful deletion
	mock.ExpectExec(`DELETE FROM lists WHERE id = \$1`).
		WithArgs(listID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteList(context.Background(), listID)
	require.NoError(t, err)

	// Test case 2: List not found
	mock.ExpectExec(`DELETE FROM lists WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteList(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.IsType(t, &domain.ErrListNotFound{}, err)

	// Test case 3: Database error
	mock.ExpectExec(`DELETE FROM lists WHERE id = \$1`).
		WithArgs("error").
		WillReturnError(errors.New("database error"))

	err = repo.DeleteList(context.Background(), "error")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete list")
}
