package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository/testutil"
)

func TestListRepository(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	workspaceRepo := testutil.NewMockWorkspaceRepository(db)
	repo := NewListRepository(workspaceRepo)

	// Create a test list
	testList := &domain.List{
		ID:            "list123",
		Name:          "Test List",
		Type:          "contact",
		IsDoubleOptin: true,
		IsPublic:      false,
		Description:   "Test list description",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	t.Run("CreateList", func(t *testing.T) {
		t.Run("successful creation", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`
				INSERT INTO lists (id, name, type, is_double_optin, is_public, description, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`)).WithArgs(
				testList.ID,
				testList.Name,
				testList.Type,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.CreateList(context.Background(), "workspace123", testList)
			require.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("database error", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`
				INSERT INTO lists (id, name, type, is_double_optin, is_public, description, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`)).WithArgs(
				testList.ID,
				testList.Name,
				testList.Type,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).WillReturnError(errors.New("database error"))

			err := repo.CreateList(context.Background(), "workspace123", testList)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to create list")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetListByID", func(t *testing.T) {
		t.Run("list found", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{
				"id", "name", "type", "is_double_optin", "is_public", "description", "created_at", "updated_at",
			}).AddRow(
				testList.ID,
				testList.Name,
				testList.Type,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				testList.CreatedAt,
				testList.UpdatedAt,
			)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
				FROM lists
				WHERE id = $1
			`)).WithArgs(testList.ID).WillReturnRows(rows)

			list, err := repo.GetListByID(context.Background(), "workspace123", testList.ID)
			require.NoError(t, err)
			assert.Equal(t, testList.ID, list.ID)
			assert.Equal(t, testList.Name, list.Name)
			assert.Equal(t, testList.Type, list.Type)
			assert.Equal(t, testList.IsDoubleOptin, list.IsDoubleOptin)
			assert.Equal(t, testList.IsPublic, list.IsPublic)
			assert.Equal(t, testList.Description, list.Description)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("list not found", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
				FROM lists
				WHERE id = $1
			`)).WithArgs(testList.ID).WillReturnError(sql.ErrNoRows)

			list, err := repo.GetListByID(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.Nil(t, list)
			assert.IsType(t, &domain.ErrListNotFound{}, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("database error", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
				FROM lists
				WHERE id = $1
			`)).WithArgs(testList.ID).WillReturnError(errors.New("database error"))

			list, err := repo.GetListByID(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.Nil(t, list)
			assert.Contains(t, err.Error(), "failed to get list")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("GetLists", func(t *testing.T) {
		t.Run("successful retrieval", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{
				"id", "name", "type", "is_double_optin", "is_public", "description", "created_at", "updated_at",
			}).AddRow(
				testList.ID,
				testList.Name,
				testList.Type,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				testList.CreatedAt,
				testList.UpdatedAt,
			)

			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
				FROM lists
				ORDER BY created_at DESC
			`)).WillReturnRows(rows)

			lists, err := repo.GetLists(context.Background(), "workspace123")
			require.NoError(t, err)
			require.Len(t, lists, 1)
			assert.Equal(t, testList.ID, lists[0].ID)
			assert.Equal(t, testList.Name, lists[0].Name)
			assert.Equal(t, testList.Type, lists[0].Type)
			assert.Equal(t, testList.IsDoubleOptin, lists[0].IsDoubleOptin)
			assert.Equal(t, testList.IsPublic, lists[0].IsPublic)
			assert.Equal(t, testList.Description, lists[0].Description)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("database error", func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT id, name, type, is_double_optin, is_public, description, created_at, updated_at
				FROM lists
				ORDER BY created_at DESC
			`)).WillReturnError(errors.New("database error"))

			lists, err := repo.GetLists(context.Background(), "workspace123")
			require.Error(t, err)
			assert.Nil(t, lists)
			assert.Contains(t, err.Error(), "failed to get lists")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("UpdateList", func(t *testing.T) {
		t.Run("successful update", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`
				UPDATE lists
				SET name = $1, type = $2, is_double_optin = $3, is_public = $4, description = $5, updated_at = $6
				WHERE id = $7
			`)).WithArgs(
				testList.Name,
				testList.Type,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				testList.ID,
			).WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.UpdateList(context.Background(), "workspace123", testList)
			require.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("list not found", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`
				UPDATE lists
				SET name = $1, type = $2, is_double_optin = $3, is_public = $4, description = $5, updated_at = $6
				WHERE id = $7
			`)).WithArgs(
				testList.Name,
				testList.Type,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				testList.ID,
			).WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.UpdateList(context.Background(), "workspace123", testList)
			require.Error(t, err)
			assert.IsType(t, &domain.ErrListNotFound{}, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("database error", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`
				UPDATE lists
				SET name = $1, type = $2, is_double_optin = $3, is_public = $4, description = $5, updated_at = $6
				WHERE id = $7
			`)).WithArgs(
				testList.Name,
				testList.Type,
				testList.IsDoubleOptin,
				testList.IsPublic,
				testList.Description,
				sqlmock.AnyArg(),
				testList.ID,
			).WillReturnError(errors.New("database error"))

			err := repo.UpdateList(context.Background(), "workspace123", testList)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to update list")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})

	t.Run("DeleteList", func(t *testing.T) {
		t.Run("successful deletion", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM lists WHERE id = $1`)).
				WithArgs(testList.ID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.DeleteList(context.Background(), "workspace123", testList.ID)
			require.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("list not found", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM lists WHERE id = $1`)).
				WithArgs(testList.ID).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.DeleteList(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.IsType(t, &domain.ErrListNotFound{}, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})

		t.Run("database error", func(t *testing.T) {
			mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM lists WHERE id = $1`)).
				WithArgs(testList.ID).
				WillReturnError(errors.New("database error"))

			err := repo.DeleteList(context.Background(), "workspace123", testList.ID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to delete list")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	})
}
