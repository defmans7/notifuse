package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
)

func TestWorkspaceRepository_AddUserToWorkspace(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)

	userWorkspace := &domain.UserWorkspace{
		UserID:      "user123",
		WorkspaceID: "workspace123",
		Role:        "member",
	}

	// Test success case
	mock.ExpectExec(`INSERT INTO user_workspaces.*VALUES.*`).
		WithArgs(
			userWorkspace.UserID,
			userWorkspace.WorkspaceID,
			userWorkspace.Role,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddUserToWorkspace(context.Background(), userWorkspace)
	require.NoError(t, err)

	// Test database error
	mock.ExpectExec(`INSERT INTO user_workspaces.*VALUES.*`).
		WithArgs(
			userWorkspace.UserID,
			userWorkspace.WorkspaceID,
			userWorkspace.Role,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(fmt.Errorf("database error"))

	err = repo.AddUserToWorkspace(context.Background(), userWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add user to workspace")
}

func TestWorkspaceRepository_RemoveUserFromWorkspace(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)
	userID := "user123"
	workspaceID := "workspace123"

	// Test success case
	mock.ExpectExec(`DELETE FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
		WithArgs(userID, workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveUserFromWorkspace(context.Background(), userID, workspaceID)
	require.NoError(t, err)

	// Test not found case
	mock.ExpectExec(`DELETE FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
		WithArgs("nonexistent", workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.RemoveUserFromWorkspace(context.Background(), "nonexistent", workspaceID)
	require.Error(t, err)

	// Test database error
	mock.ExpectExec(`DELETE FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
		WithArgs(userID, workspaceID).
		WillReturnError(fmt.Errorf("database error"))

	err = repo.RemoveUserFromWorkspace(context.Background(), userID, workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove user from workspace")

	// Test error getting affected rows
	mock.ExpectExec(`DELETE FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
		WithArgs(userID, workspaceID).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err = repo.RemoveUserFromWorkspace(context.Background(), userID, workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get affected rows")
}

func TestWorkspaceRepository_GetUserWorkspaces(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)
	userID := "user123"

	// Test success case
	now := time.Now().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"user_id", "workspace_id", "role", "created_at", "updated_at"}).
		AddRow(userID, "workspace1", "owner", now, now).
		AddRow(userID, "workspace2", "member", now, now)

	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	userWorkspaces, err := repo.GetUserWorkspaces(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, userWorkspaces, 2)
	assert.Equal(t, "workspace1", userWorkspaces[0].WorkspaceID)
	assert.Equal(t, "owner", userWorkspaces[0].Role)
	assert.Equal(t, "workspace2", userWorkspaces[1].WorkspaceID)
	assert.Equal(t, "member", userWorkspaces[1].Role)

	// Test database query error
	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("database error"))

	_, err = repo.GetUserWorkspaces(context.Background(), userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user workspaces")

	// Test empty result
	emptyRows := sqlmock.NewRows([]string{"user_id", "workspace_id", "role", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(emptyRows)

	emptyWorkspaces, err := repo.GetUserWorkspaces(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, emptyWorkspaces)
}

func TestWorkspaceRepository_GetUserWorkspace(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)
	userID := "user123"
	workspaceID := "workspace123"

	// Test success case
	now := time.Now().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"user_id", "workspace_id", "role", "created_at", "updated_at"}).
		AddRow(userID, workspaceID, "owner", now, now)

	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
		WithArgs(userID, workspaceID).
		WillReturnRows(rows)

	userWorkspace, err := repo.GetUserWorkspace(context.Background(), userID, workspaceID)
	require.NoError(t, err)
	assert.Equal(t, userID, userWorkspace.UserID)
	assert.Equal(t, workspaceID, userWorkspace.WorkspaceID)
	assert.Equal(t, "owner", userWorkspace.Role)

	// Test not found case
	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
		WithArgs("nonexistent", workspaceID).
		WillReturnError(errors.New("no rows"))

	_, err = repo.GetUserWorkspace(context.Background(), "nonexistent", workspaceID)
	require.Error(t, err)

	// Test database query error
	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
		WithArgs(userID, workspaceID).
		WillReturnError(fmt.Errorf("database error"))

	_, err = repo.GetUserWorkspace(context.Background(), userID, workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user workspace")
}

func TestWorkspaceRepository_IsUserWorkspaceMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &workspaceRepository{
		systemDB: db,
		dbConfig: &config.DatabaseConfig{},
	}

	userID := "user-123"
	workspaceID := "ws-123"

	t.Run("user is a member", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
			WithArgs(userID, workspaceID).
			WillReturnRows(rows)

		isMember, err := repo.IsUserWorkspaceMember(context.Background(), userID, workspaceID)
		require.NoError(t, err)
		assert.True(t, isMember)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("user is not a member", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
			WithArgs(userID, workspaceID).
			WillReturnRows(rows)

		isMember, err := repo.IsUserWorkspaceMember(context.Background(), userID, workspaceID)
		require.NoError(t, err)
		assert.False(t, isMember)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_workspaces WHERE user_id = \$1 AND workspace_id = \$2`).
			WithArgs(userID, workspaceID).
			WillReturnError(fmt.Errorf("database error"))

		isMember, err := repo.IsUserWorkspaceMember(context.Background(), userID, workspaceID)
		require.Error(t, err)
		assert.False(t, isMember)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}
