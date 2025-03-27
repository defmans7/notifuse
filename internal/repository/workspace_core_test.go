package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/repository/testutil"
)

func TestWorkspaceRepository_CheckWorkspaceIDExists(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
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
	db, mock, cleanup := testutil.SetupMockDB(t)
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
}

func TestWorkspaceRepository_GetByID(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
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
	db, mock, cleanup := testutil.SetupMockDB(t)
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
	db, mock, cleanup := testutil.SetupMockDB(t)
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
	db, mock, cleanup := testutil.SetupMockDB(t)
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

	// Mock for deleting user_workspaces
	mock.ExpectExec(`DELETE FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 2)) // Assume 2 users were removed

	// Mock for deleting workspace_invitations
	mock.ExpectExec(`DELETE FROM workspace_invitations WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // Assume 1 invitation was removed

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

	// Test case: Error when deleting user_workspaces
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnError(errors.New("user_workspaces deletion failed"))

	err = repo.Delete(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user workspaces")

	// Test case: Error when deleting workspace_invitations
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	mock.ExpectExec(`DELETE FROM workspace_invitations WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnError(errors.New("invitations deletion failed"))

	err = repo.Delete(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete workspace invitations")

	// Test case: Workspace record deletion error
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	mock.ExpectExec(`DELETE FROM workspace_invitations WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnError(errors.New("database error"))

	err = repo.Delete(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete workspace")

	// Test case: No rows affected (workspace not found)
	mock.ExpectExec("DROP DATABASE IF EXISTS.*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM user_workspaces WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM workspace_invitations WHERE workspace_id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`DELETE FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace not found")
}
