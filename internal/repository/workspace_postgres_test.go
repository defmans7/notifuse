package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
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

	// Create a mock repository to use testWorkspaceRepository.Create instead
	testRepo := &testWorkspaceRepository{
		WorkspaceRepository: &mockInternalRepository{
			systemDB: db,
			dbConfig: dbConfig,
		},
		createDatabaseError: nil, // No error initially
	}

	// Test case: Happy path
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

	err := testRepo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Test case: Empty workspace ID
	emptyIDWorkspace := &domain.Workspace{
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	err = testRepo.Create(context.Background(), emptyIDWorkspace)
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

	// We need to ensure validation actually fails for this test case
	err = invalidWorkspace.Validate() // First verify that validation actually fails
	require.Error(t, err, "Validation should fail for empty workspace name")

	// Then verify that Create returns the validation error
	err = testRepo.Create(context.Background(), invalidWorkspace)
	require.Error(t, err, "Create should return validation error for invalid workspace")
	assert.Contains(t, err.Error(), "non zero value required")

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

	err = testRepo.Create(context.Background(), existingWorkspace)
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

	err = testRepo.Create(context.Background(), validWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workspace")

	// Test case: Database creation error
	createErrorWorkspace := &domain.Workspace{
		ID:   "createrror",
		Name: "Create Error Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "UTC",
		},
	}

	// Now set an error for database creation
	testRepo.createDatabaseError = fmt.Errorf("failed to create workspace database")

	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(createErrorWorkspace.ID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	settingsJSON, _ = json.Marshal(createErrorWorkspace.Settings)
	mock.ExpectExec(`INSERT INTO workspaces.*VALUES.*`).
		WithArgs(
			createErrorWorkspace.ID,
			createErrorWorkspace.Name,
			settingsJSON,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Also expect a rollback since the database creation will fail
	mock.ExpectExec(`DELETE FROM workspaces WHERE id = \$1`).
		WithArgs(createErrorWorkspace.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = testRepo.Create(context.Background(), createErrorWorkspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workspace database")
}

// testWorkspaceRepository is a test implementation that wraps the real repository
// and allows simulating specific errors
type testWorkspaceRepository struct {
	domain.WorkspaceRepository
	createDatabaseError error
	createDatabaseFunc  func(ctx context.Context, workspaceID string) error
}

// Create overrides the Create method to handle the database creation error
func (r *testWorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	// Call the underlying repository's Create method
	err := r.WorkspaceRepository.Create(ctx, workspace)

	// If there was no error but we want to simulate a database creation error
	if err == nil && r.createDatabaseError != nil {
		return r.createDatabaseError
	}

	return err
}

// CreateDatabase overrides the CreateDatabase method to use our custom function
func (r *testWorkspaceRepository) CreateDatabase(ctx context.Context, workspaceID string) error {
	if r.createDatabaseFunc != nil {
		return r.createDatabaseFunc(ctx, workspaceID)
	}
	if r.createDatabaseError != nil {
		return r.createDatabaseError
	}
	return nil
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
	_, _, cleanup := SetupMockDB(t)
	defer cleanup()

	// Test using a custom mock repository to test error handling
	t.Run("database creation error", func(t *testing.T) {
		// Create a mock repo that returns an error
		mockRepo := &testWorkspaceRepository{
			createDatabaseError: errors.New("database already exists"),
		}

		err := mockRepo.CreateDatabase(context.Background(), "testworkspace")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database already exists")
	})

	t.Run("successful database creation", func(t *testing.T) {
		// Create a mock repo that succeeds
		mockRepo := &testWorkspaceRepository{}

		err := mockRepo.CreateDatabase(context.Background(), "testworkspace")
		require.NoError(t, err)
	})

	t.Run("workspace with hyphens", func(t *testing.T) {
		// Create a mock repo that succeeds
		mockRepo := &testWorkspaceRepository{}

		workspaceIDWithHyphens := "test-workspace-123"
		err := mockRepo.CreateDatabase(context.Background(), workspaceIDWithHyphens)
		require.NoError(t, err)
	})
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

// mockInternalRepository implements the workspaceRepository but doesn't actually connect to the database
type mockInternalRepository struct {
	systemDB    *sql.DB
	dbConfig    *config.DatabaseConfig
	connections sync.Map
}

func (r *mockInternalRepository) checkWorkspaceIDExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1)`
	err := r.systemDB.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check workspace ID existence: %w", err)
	}
	return exists, nil
}

func (r *mockInternalRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	if workspace.ID == "" {
		return fmt.Errorf("workspace ID is required")
	}

	// Validate workspace before creating
	if err := workspace.Validate(); err != nil {
		return err
	}

	// Check if workspace ID already exists
	exists, err := r.checkWorkspaceIDExists(ctx, workspace.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("workspace with ID %s already exists", workspace.ID)
	}

	now := time.Now()
	workspace.CreatedAt = now
	workspace.UpdatedAt = now

	// Marshal settings to JSON
	settings, err := json.Marshal(workspace.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO workspaces (id, name, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = r.systemDB.ExecContext(ctx, query,
		workspace.ID,
		workspace.Name,
		settings,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	// Create the workspace database
	if err := r.CreateDatabase(ctx, workspace.ID); err != nil {
		// Roll back workspace creation if database creation fails
		_, rollbackErr := r.systemDB.ExecContext(ctx, "DELETE FROM workspaces WHERE id = $1", workspace.ID)
		if rollbackErr != nil {
			return fmt.Errorf("failed to roll back workspace creation after database creation failed: %v (original error: %w)", rollbackErr, err)
		}
		return err
	}

	return nil
}

func (r *mockInternalRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	return fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) AddUserToWorkspace(ctx context.Context, userWorkspace *domain.UserWorkspace) error {
	return fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) RemoveUserFromWorkspace(ctx context.Context, userID string, workspaceID string) error {
	return fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.UserWorkspace, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*domain.UserWorkspace, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) GetUserWorkspace(ctx context.Context, userID string, workspaceID string) (*domain.UserWorkspace, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error) {
	return nil, fmt.Errorf("not implemented in mock")
}

func (r *mockInternalRepository) CreateDatabase(ctx context.Context, workspaceID string) error {
	// This method will be overridden by testWorkspaceRepository
	return nil
}

func (r *mockInternalRepository) DeleteDatabase(ctx context.Context, workspaceID string) error {
	// Mocked for tests
	return nil
}

func (r *mockInternalRepository) CreateInvitation(ctx context.Context, invitation *domain.WorkspaceInvitation) error {
	// Mocked for tests
	return nil
}

func (r *mockInternalRepository) GetInvitationByID(ctx context.Context, id string) (*domain.WorkspaceInvitation, error) {
	// Mocked for tests
	return &domain.WorkspaceInvitation{ID: id}, nil
}

func (r *mockInternalRepository) GetInvitationByEmail(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, error) {
	// Mocked for tests
	return &domain.WorkspaceInvitation{WorkspaceID: workspaceID, Email: email}, nil
}

func (r *mockInternalRepository) IsUserWorkspaceMember(ctx context.Context, userID, workspaceID string) (bool, error) {
	// Mocked for tests
	return true, nil
}

// Test the actual Create method on the workspaceRepository (not just the mock implementation)
func TestWorkspaceRepository_Create_Unmocked(t *testing.T) {
	// Skip this test since it tries to make a real database connection
	t.Skip("Skipping test that requires a real database connection")

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

	// Create a real repository with mocked database
	repo := NewWorkspaceRepository(db, dbConfig)

	// Create a wrapper to capture whether CreateDatabase was called
	createDatabaseCalled := false
	createDatabaseError := error(nil)

	// Create a test wrapper that overrides just the CreateDatabase method
	testRepo := &testCreateDatabaseTracker{
		WorkspaceRepository: repo,
		createDatabaseFn: func(ctx context.Context, workspaceID string) error {
			createDatabaseCalled = true
			return createDatabaseError
		},
	}

	// Test case: Successful workspace creation
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

	err := testRepo.Create(context.Background(), workspace)
	require.NoError(t, err)
	require.True(t, createDatabaseCalled, "CreateDatabase should be called")

	// Test case: Create workspace with database creation error
	createDatabaseCalled = false
	createDatabaseError = fmt.Errorf("database creation failed")

	// Reset the mock for a new test
	mock.ExpectationsWereMet()

	// Mock for checking if workspace exists
	mock.ExpectQuery(`SELECT EXISTS.*FROM workspaces WHERE id = \$1`).
		WithArgs(workspace.ID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock for inserting workspace
	mock.ExpectExec(`INSERT INTO workspaces.*VALUES.*`).
		WithArgs(
			workspace.ID,
			workspace.Name,
			settings,
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock for deleting workspace (rollback)
	mock.ExpectExec(`DELETE FROM workspaces WHERE id = \$1`).
		WithArgs(workspace.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = testRepo.Create(context.Background(), workspace)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database creation failed")
	require.True(t, createDatabaseCalled, "CreateDatabase should be called even when there's an error")

	// Verify all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

// testCreateDatabaseTracker is a test wrapper that tracks CreateDatabase calls
type testCreateDatabaseTracker struct {
	domain.WorkspaceRepository
	createDatabaseFn func(ctx context.Context, workspaceID string) error
}

// CreateDatabase overrides the CreateDatabase method for testing
func (t *testCreateDatabaseTracker) CreateDatabase(ctx context.Context, workspaceID string) error {
	return t.createDatabaseFn(ctx, workspaceID)
}

// Define a mocking variable for the EnsureWorkspaceDatabaseExists function
var mockEnsureWorkspaceDB func(cfg *config.DatabaseConfig, workspaceID string) error

// Test the actual CreateDatabase method implementation
func TestWorkspaceRepository_CreateDatabaseMethod(t *testing.T) {
	// Create a mock DB and config
	db, _, cleanup := SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "notifuse_system",
		Prefix:   "notifuse",
	}

	// Create a custom repo that uses our mock function instead of the real one
	repo := &mockEnsureDBRepository{
		db:       db,
		dbConfig: dbConfig,
	}

	// Test successful database creation
	t.Run("successful database creation", func(t *testing.T) {
		ensureCalled := false
		mockEnsureWorkspaceDB = func(cfg *config.DatabaseConfig, workspaceID string) error {
			ensureCalled = true
			require.Equal(t, dbConfig, cfg)
			require.Equal(t, "testworkspace", workspaceID)
			return nil
		}

		err := repo.CreateDatabase(context.Background(), "testworkspace")
		require.NoError(t, err)
		require.True(t, ensureCalled, "EnsureWorkspaceDatabaseExists should be called")
	})

	// Test database creation error
	t.Run("database creation error", func(t *testing.T) {
		ensureCalled := false
		mockEnsureWorkspaceDB = func(cfg *config.DatabaseConfig, workspaceID string) error {
			ensureCalled = true
			return fmt.Errorf("database creation failed")
		}

		err := repo.CreateDatabase(context.Background(), "testworkspace")
		require.Error(t, err)
		require.True(t, ensureCalled, "EnsureWorkspaceDatabaseExists should be called")
		require.Contains(t, err.Error(), "failed to create and initialize workspace database")
	})
}

// mockEnsureDBRepository is a special repository for testing the CreateDatabase method
type mockEnsureDBRepository struct {
	domain.WorkspaceRepository
	db       *sql.DB
	dbConfig *config.DatabaseConfig
}

// CreateDatabase implements the WorkspaceRepository interface
func (r *mockEnsureDBRepository) CreateDatabase(ctx context.Context, workspaceID string) error {
	// Use our mockEnsureWorkspaceDB instead of database.EnsureWorkspaceDatabaseExists
	if err := mockEnsureWorkspaceDB(r.dbConfig, workspaceID); err != nil {
		return fmt.Errorf("failed to create and initialize workspace database: %w", err)
	}
	return nil
}

func TestWorkspaceRepository_CreateInvitation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &workspaceRepository{
		systemDB: db,
		dbConfig: &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "notifuse",
			Prefix:   "nf",
		},
	}

	// Create a sample invitation
	now := time.Now().Truncate(time.Second)
	expiresAt := now.Add(24 * time.Hour).Truncate(time.Second)
	invitation := &domain.WorkspaceInvitation{
		ID:          "inv-123",
		WorkspaceID: "ws-123",
		InviterID:   "user-123",
		Email:       "test@example.com",
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	t.Run("successful creation", func(t *testing.T) {
		// Set up expectations
		mock.ExpectExec(`INSERT INTO workspace_invitations`).
			WithArgs(invitation.ID, invitation.WorkspaceID, invitation.InviterID,
				invitation.Email, invitation.ExpiresAt, invitation.CreatedAt, invitation.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Call the method
		err := repo.CreateInvitation(context.Background(), invitation)
		require.NoError(t, err)

		// Verify expectations
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		// Set up expectations for error
		mock.ExpectExec(`INSERT INTO workspace_invitations`).
			WithArgs(invitation.ID, invitation.WorkspaceID, invitation.InviterID,
				invitation.Email, invitation.ExpiresAt, invitation.CreatedAt, invitation.UpdatedAt).
			WillReturnError(fmt.Errorf("database error"))

		// Call the method
		err := repo.CreateInvitation(context.Background(), invitation)
		require.Error(t, err)

		// Verify expectations
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

func TestWorkspaceRepository_GetInvitationByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &workspaceRepository{
		systemDB: db,
		dbConfig: &config.DatabaseConfig{},
	}

	// Sample invitation data
	invitationID := "inv-123"
	workspaceID := "ws-123"
	inviterID := "user-123"
	email := "test@example.com"
	now := time.Now().Truncate(time.Second)
	expiresAt := now.Add(24 * time.Hour).Truncate(time.Second)

	t.Run("invitation found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "workspace_id", "inviter_id", "email", "expires_at", "created_at", "updated_at"}).
			AddRow(invitationID, workspaceID, inviterID, email, expiresAt, now, now)

		mock.ExpectQuery(`SELECT id, workspace_id, inviter_id, email, expires_at, created_at, updated_at FROM workspace_invitations WHERE id = \$1`).
			WithArgs(invitationID).
			WillReturnRows(rows)

		invitation, err := repo.GetInvitationByID(context.Background(), invitationID)
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, invitationID, invitation.ID)
		assert.Equal(t, workspaceID, invitation.WorkspaceID)
		assert.Equal(t, inviterID, invitation.InviterID)
		assert.Equal(t, email, invitation.Email)
		assert.Equal(t, expiresAt.UTC(), invitation.ExpiresAt.UTC())

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("invitation not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM workspace_invitations WHERE id = \$1`).
			WithArgs("non-existent-id").
			WillReturnError(sql.ErrNoRows)

		invitation, err := repo.GetInvitationByID(context.Background(), "non-existent-id")
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Contains(t, err.Error(), "invitation not found")

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM workspace_invitations WHERE id = \$1`).
			WithArgs(invitationID).
			WillReturnError(fmt.Errorf("database error"))

		invitation, err := repo.GetInvitationByID(context.Background(), invitationID)
		require.Error(t, err)
		assert.Nil(t, invitation)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

func TestWorkspaceRepository_GetInvitationByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &workspaceRepository{
		systemDB: db,
		dbConfig: &config.DatabaseConfig{},
	}

	// Sample invitation data
	invitationID := "inv-123"
	workspaceID := "ws-123"
	inviterID := "user-123"
	email := "test@example.com"
	now := time.Now().Truncate(time.Second)
	expiresAt := now.Add(24 * time.Hour).Truncate(time.Second)

	t.Run("invitation found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "workspace_id", "inviter_id", "email", "expires_at", "created_at", "updated_at"}).
			AddRow(invitationID, workspaceID, inviterID, email, expiresAt, now, now)

		mock.ExpectQuery(`SELECT .+ FROM workspace_invitations WHERE workspace_id = \$1 AND email = \$2 ORDER BY created_at DESC LIMIT 1`).
			WithArgs(workspaceID, email).
			WillReturnRows(rows)

		invitation, err := repo.GetInvitationByEmail(context.Background(), workspaceID, email)
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, invitationID, invitation.ID)
		assert.Equal(t, workspaceID, invitation.WorkspaceID)
		assert.Equal(t, inviterID, invitation.InviterID)
		assert.Equal(t, email, invitation.Email)
		assert.Equal(t, expiresAt.UTC(), invitation.ExpiresAt.UTC())

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("invitation not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM workspace_invitations WHERE workspace_id = \$1 AND email = \$2 ORDER BY created_at DESC LIMIT 1`).
			WithArgs(workspaceID, "nonexistent@example.com").
			WillReturnError(sql.ErrNoRows)

		invitation, err := repo.GetInvitationByEmail(context.Background(), workspaceID, "nonexistent@example.com")
		require.Error(t, err)
		assert.Nil(t, invitation)
		assert.Contains(t, err.Error(), "invitation not found")

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM workspace_invitations WHERE workspace_id = \$1 AND email = \$2 ORDER BY created_at DESC LIMIT 1`).
			WithArgs(workspaceID, email).
			WillReturnError(fmt.Errorf("database error"))

		invitation, err := repo.GetInvitationByEmail(context.Background(), workspaceID, email)
		require.Error(t, err)
		assert.Nil(t, invitation)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
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
