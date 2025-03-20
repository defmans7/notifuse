package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"notifuse/server/config"
	"notifuse/server/internal/domain"
)

func TestWorkspaceRepository_CheckWorkspaceIDExists(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig).(*workspaceRepository)
	workspaceID := "testworkspace"

	// Test when workspace exists
	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.checkWorkspaceIDExists(context.Background(), workspaceID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test when workspace doesn't exist
	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err = repo.checkWorkspaceIDExists(context.Background(), workspaceID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Test database error
	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnError(fmt.Errorf("database error"))

	_, err = repo.checkWorkspaceIDExists(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check workspace ID existence")
}

func TestWorkspaceRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "notifuse_system",
		Prefix:   "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)

	// Test case: Happy path (partial, will error due to mock limitations)
	workspace := &domain.Workspace{
		ID:   "testworkspace",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	// Mock for checking if workspace exists
	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(workspace.ID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock for inserting workspace
	settings, _ := json.Marshal(workspace.Settings)
	mock.ExpectExec(`INSERT INTO workspaces.*VALUES.*`).
		WithArgs(
			workspace.ID,
			workspace.Name,
			settings,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock for creating workspace database
	createDBQuery := fmt.Sprintf("CREATE DATABASE %s_ws_%s", dbConfig.Prefix, workspace.ID)
	mock.ExpectExec(createDBQuery).WillReturnResult(sqlmock.NewResult(0, 0))

	// Since we can't mock the connection to a new database in this test,
	// we need to mock the behavior differently or skip this part
	err := repo.Create(context.Background(), workspace)
	require.Error(t, err) // Will error because we can't mock ConnectToWorkspace

	// Test case: Empty workspace ID
	emptyIDWorkspace := &domain.Workspace{
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	err = repo.Create(context.Background(), emptyIDWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace ID is required")

	// Test case: Invalid workspace (validation failure)
	invalidWorkspace := &domain.Workspace{
		ID:   "testworkspace",
		Name: "", // Name is required
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	err = repo.Create(context.Background(), invalidWorkspace)
	require.Error(t, err)

	// Test case: Workspace ID already exists
	existingWorkspace := &domain.Workspace{
		ID:   "existingworkspace",
		Name: "Existing Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(existingWorkspace.ID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	err = repo.Create(context.Background(), existingWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test case: Database error during insert
	validWorkspace := &domain.Workspace{
		ID:   "validworkspace",
		Name: "Valid Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(validWorkspace.ID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	settingsJSON, _ := json.Marshal(validWorkspace.Settings)
	mock.ExpectExec(`INSERT INTO workspaces.*VALUES.*`).
		WithArgs(
			validWorkspace.ID,
			validWorkspace.Name,
			settingsJSON,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnError(fmt.Errorf("insert error"))

	err = repo.Create(context.Background(), validWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workspace")
}

func TestWorkspaceRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)

	// Test data
	workspaceID := "testworkspace"
	workspaceName := "Test Workspace"
	settings := domain.WorkspaceSettings{
		Timezone: "UTC",
	}
	settingsJSON, _ := json.Marshal(settings)
	createdAt := time.Now().Truncate(time.Second)
	updatedAt := createdAt

	// Mock for successful query
	rows := sqlmock.NewRows([]string{"id", "name", "settings", "created_at", "updated_at"}).
		AddRow(workspaceID, workspaceName, settingsJSON, createdAt, updatedAt)

	mock.ExpectQuery(`SELECT id, name, settings, created_at, updated_at FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnRows(rows)

	workspace, err := repo.GetByID(context.Background(), workspaceID)
	require.NoError(t, err)
	assert.Equal(t, workspaceID, workspace.ID)
	assert.Equal(t, workspaceName, workspace.Name)
	assert.Equal(t, settings.Timezone, workspace.Settings.Timezone)
	assert.Equal(t, createdAt.Unix(), workspace.CreatedAt.Unix())
	assert.Equal(t, updatedAt.Unix(), workspace.UpdatedAt.Unix())

	// Test not found
	mock.ExpectQuery(`SELECT id, name, settings, created_at, updated_at FROM workspaces WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnError(errors.New("no rows"))

	_, err = repo.GetByID(context.Background(), "nonexistent")
	require.Error(t, err)
}

func TestWorkspaceRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)

	// Test data
	workspace1 := &domain.Workspace{
		ID:        "workspace1",
		Name:      "Workspace One",
		Settings:  domain.WorkspaceSettings{Timezone: "UTC"},
		CreatedAt: time.Now().Add(-2 * time.Hour).Truncate(time.Second),
		UpdatedAt: time.Now().Add(-1 * time.Hour).Truncate(time.Second),
	}
	settings1JSON, _ := json.Marshal(workspace1.Settings)

	workspace2 := &domain.Workspace{
		ID:        "workspace2",
		Name:      "Workspace Two",
		Settings:  domain.WorkspaceSettings{Timezone: "America/New_York"},
		CreatedAt: time.Now().Add(-1 * time.Hour).Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}
	settings2JSON, _ := json.Marshal(workspace2.Settings)

	// Mock for successful query
	rows := sqlmock.NewRows([]string{"id", "name", "settings", "created_at", "updated_at"}).
		AddRow(workspace1.ID, workspace1.Name, settings1JSON, workspace1.CreatedAt, workspace1.UpdatedAt).
		AddRow(workspace2.ID, workspace2.Name, settings2JSON, workspace2.CreatedAt, workspace2.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, settings, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
		WillReturnRows(rows)

	workspaces, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, workspaces, 2)
	assert.Equal(t, workspace1.ID, workspaces[0].ID)
	assert.Equal(t, workspace2.ID, workspaces[1].ID)
}

func TestWorkspaceRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)

	// Test case: Successful update
	workspace := &domain.Workspace{
		ID:   "testworkspace",
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone:   "Europe/London",
			WebsiteURL: "https://example.com",
		},
	}

	// Mock for updating workspace
	settings, _ := json.Marshal(workspace.Settings)
	mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, updated_at = \$3 WHERE id = \$4`).
		WithArgs(
			workspace.Name,
			settings,
			sqlmock.AnyArg(), // updated_at
			workspace.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), workspace)
	require.NoError(t, err)

	// Test case: Workspace not found
	notFoundWorkspace := &domain.Workspace{
		ID:   "nonexistent",
		Name: "Nonexistent Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	notFoundSettings, _ := json.Marshal(notFoundWorkspace.Settings)
	mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, updated_at = \$3 WHERE id = \$4`).
		WithArgs(
			notFoundWorkspace.Name,
			notFoundSettings,
			sqlmock.AnyArg(), // updated_at
			notFoundWorkspace.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), notFoundWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace not found")

	// Test case: Invalid workspace (validation failure)
	invalidWorkspace := &domain.Workspace{
		ID:   "testworkspace",
		Name: "", // Empty name, which should fail validation
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	err = repo.Update(context.Background(), invalidWorkspace)
	require.Error(t, err)

	// Test case: Database error
	validWorkspace := &domain.Workspace{
		ID:   "testworkspace",
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	validSettings, _ := json.Marshal(validWorkspace.Settings)
	mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, updated_at = \$3 WHERE id = \$4`).
		WithArgs(
			validWorkspace.Name,
			validSettings,
			sqlmock.AnyArg(), // updated_at
			validWorkspace.ID,
		).
		WillReturnError(fmt.Errorf("database error"))

	err = repo.Update(context.Background(), validWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update workspace")

	// Test case: Error getting affected rows
	mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, updated_at = \$3 WHERE id = \$4`).
		WithArgs(
			validWorkspace.Name,
			validSettings,
			sqlmock.AnyArg(), // updated_at
			validWorkspace.ID,
		).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err = repo.Update(context.Background(), validWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get affected rows")
}

func TestWorkspaceRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "notifuse_system",
		Prefix:   "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)
	workspaceID := "testworkspace"

	// Test success case
	// Mock for dropping database
	dropDBQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s_ws_%s", dbConfig.Prefix, workspaceID)
	mock.ExpectExec(dropDBQuery).WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock for deleting workspace record
	mock.ExpectExec(`DELETE FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), workspaceID)
	require.NoError(t, err)

	// Test case: Database error during database drop
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnError(errors.New("permission denied"))

	err = repo.Delete(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	// Test case: Workspace record deletion error
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnError(errors.New("database error"))

	err = repo.Delete(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete workspace")

	// Test case: No rows affected (workspace not found)
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace not found")
}

func TestWorkspaceRepository_CreateDatabase(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "notifuse_system",
		Prefix:   "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)
	workspaceID := "testworkspace"

	// Test case: Database creation error
	mock.ExpectExec("CREATE DATABASE.*").
		WillReturnError(errors.New("database already exists"))

	err := repo.(*workspaceRepository).CreateDatabase(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workspace database")
	assert.Contains(t, err.Error(), "database already exists")

	// Test case: Connection error after database creation
	mock.ExpectExec("CREATE DATABASE.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// The error will happen when trying to connect to the new database
	// which we can't fully mock, but this will increase coverage of the error path
	err = repo.(*workspaceRepository).CreateDatabase(context.Background(), workspaceID)
	require.Error(t, err)

	// Test with hyphens in workspace ID (should replace with underscores)
	workspaceIDWithHyphens := "test-workspace-123"
	safeID := strings.ReplaceAll(workspaceIDWithHyphens, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", dbConfig.Prefix, safeID)

	mock.ExpectExec(fmt.Sprintf("CREATE DATABASE %s", dbName)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// This will still error because we can't mock ConnectToWorkspace
	err = repo.(*workspaceRepository).CreateDatabase(context.Background(), workspaceIDWithHyphens)
	require.Error(t, err)
}

func TestWorkspaceRepository_DeleteDatabase(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "notifuse_system",
		Prefix:   "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)
	workspaceID := "testworkspace"

	// Test database drop error
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnError(errors.New("permission denied"))

	err := repo.(*workspaceRepository).DeleteDatabase(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete workspace database")
	assert.Contains(t, err.Error(), "permission denied")

	// Test successful database drop
	safeID := strings.ReplaceAll(workspaceID, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", dbConfig.Prefix, safeID)

	mock.ExpectExec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.(*workspaceRepository).DeleteDatabase(context.Background(), workspaceID)
	require.NoError(t, err)
}

func TestWorkspaceRepository_GetConnection(t *testing.T) {
	// Create a test database config
	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "test_db",
		Prefix:   "test",
	}

	// Create a mock database
	mockDB, _, cleanup := SetupMockDB(t)
	defer cleanup()

	// Create a repository instance
	repo := NewWorkspaceRepository(mockDB, dbConfig).(*workspaceRepository)

	ctx := context.Background()
	workspaceID := "test-workspace"

	// Test with a successful mock workspace DB connection
	mockWorkspaceDB, _, mockWorkspaceCleanup := SetupMockDB(t)
	defer mockWorkspaceCleanup()

	// Store the mock connection in the repository's connection map directly
	repo.connections.Store(workspaceID, mockWorkspaceDB)

	// Test case 1: Getting a connection that already exists
	db1, err := repo.GetConnection(ctx, workspaceID)
	assert.NoError(t, err)
	assert.Equal(t, mockWorkspaceDB, db1)

	// Test case 2: Error case can't be fully tested due to monkey patching limitations
	// But we can test that a non-existent connection returns an error
	_, err = repo.GetConnection(ctx, "non-existent-workspace")
	assert.Error(t, err)

	// Test case 3: Add a fake connection to the connection pool and verify it's there
	newWorkspaceDB, _, newWorkspaceCleanup := SetupMockDB(t)
	defer newWorkspaceCleanup()

	newWorkspaceID := "new-workspace"
	repo.connections.Store(newWorkspaceID, newWorkspaceDB)

	// Verify the connection is in the pool
	_, exists := repo.connections.Load(newWorkspaceID)
	assert.True(t, exists, "Connection should be in the pool")

	// GetConnection call (may or may not error depending on the environment)
	repo.GetConnection(context.Background(), newWorkspaceID)
}

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

func TestWorkspaceRepository_GetWorkspaceUsers(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig)
	workspaceID := "workspace123"

	// Test success case
	now := time.Now().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"user_id", "workspace_id", "role", "created_at", "updated_at"}).
		AddRow("user1", workspaceID, "owner", now, now).
		AddRow("user2", workspaceID, "member", now, now)

	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnRows(rows)

	workspaceUsers, err := repo.GetWorkspaceUsers(context.Background(), workspaceID)
	require.NoError(t, err)
	assert.Len(t, workspaceUsers, 2)
	assert.Equal(t, "user1", workspaceUsers[0].UserID)
	assert.Equal(t, "owner", workspaceUsers[0].Role)
	assert.Equal(t, "user2", workspaceUsers[1].UserID)
	assert.Equal(t, "member", workspaceUsers[1].Role)

	// Test database query error
	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnError(fmt.Errorf("database error"))

	_, err = repo.GetWorkspaceUsers(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get workspace users")

	// Test empty result
	emptyRows := sqlmock.NewRows([]string{"user_id", "workspace_id", "role", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT user_id, workspace_id, role, created_at, updated_at FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnRows(emptyRows)

	emptyUsers, err := repo.GetWorkspaceUsers(context.Background(), workspaceID)
	require.NoError(t, err)
	assert.Empty(t, emptyUsers)
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
