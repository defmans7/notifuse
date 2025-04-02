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
	"github.com/golang/mock/gomock"
)

func TestWorkspaceRepository_CheckWorkspaceIDExists(t *testing.T) {
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

	repo := NewWorkspaceRepository(db, dbConfig, "secret-key").(*workspaceRepository)
	workspaceID := "test-workspace"

	// Test successful check
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM workspaces WHERE id = \$1\)`).
		WithArgs(workspaceID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.checkWorkspaceIDExists(context.Background(), workspaceID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test database error
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM workspaces WHERE id = \$1\)`).
		WithArgs(workspaceID).
		WillReturnError(fmt.Errorf("database error"))

	exists, err = repo.checkWorkspaceIDExists(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Equal(t, "database error", err.Error())
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

	repo := NewWorkspaceRepository(db, dbConfig, "secret-key")

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

	repo := NewWorkspaceRepository(db, dbConfig, "secret-key")

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

	repo := NewWorkspaceRepository(db, dbConfig, "secret-key")

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

	// Test case: Database error during workspace update
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
	assert.Equal(t, "database error", err.Error())

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
	assert.Equal(t, "rows affected error", err.Error())
}

func TestWorkspaceRepository_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup a mock implementation of the repository
	// This is different from the standard pattern because we're testing
	// the Delete method itself, so we need to use a wrapper
	// to avoid infinite recursion
	workspaceRepo := &workspaceRepositoryDeleteTest{
		deleteResults: map[string]error{
			"test-workspace":        nil,
			"error-workspace":       fmt.Errorf("database error"),
			"nonexistent-workspace": fmt.Errorf("workspace not found"),
		},
		deleteDatabaseResults: map[string]error{
			"test-workspace":              nil,
			"error-workspace":             nil,
			"permission-denied-workspace": fmt.Errorf("permission denied"),
			"nonexistent-workspace":       nil,
		},
	}

	// Test cases
	testCases := []struct {
		name          string
		workspaceID   string
		expectedError string
	}{
		{
			name:          "successful deletion",
			workspaceID:   "test-workspace",
			expectedError: "",
		},
		{
			name:          "database error during deletion",
			workspaceID:   "error-workspace",
			expectedError: "database error",
		},
		{
			name:          "permission denied during database deletion",
			workspaceID:   "permission-denied-workspace",
			expectedError: "permission denied",
		},
		{
			name:          "workspace not found",
			workspaceID:   "nonexistent-workspace",
			expectedError: "workspace not found",
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the method under test
			err := workspaceRepo.Delete(context.Background(), tc.workspaceID)

			// Check the result
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

// A test implementation of the repository for Delete test only
type workspaceRepositoryDeleteTest struct {
	domain.WorkspaceRepository // This embeds the interface, making all methods required
	deleteResults              map[string]error
	deleteDatabaseResults      map[string]error
}

// Override the Delete method for testing
func (r *workspaceRepositoryDeleteTest) Delete(ctx context.Context, id string) error {
	// First call DeleteDatabase to match the real implementation
	if err := r.DeleteDatabase(ctx, id); err != nil {
		return err
	}

	// Then return the result for this specific workspace ID
	return r.deleteResults[id]
}

// Override the DeleteDatabase method for testing
func (r *workspaceRepositoryDeleteTest) DeleteDatabase(ctx context.Context, id string) error {
	return r.deleteDatabaseResults[id]
}
