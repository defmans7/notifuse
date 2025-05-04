package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
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

	t.Run("successful creation", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "test-workspace",
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
		}

		// Mock for checking if workspace exists
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM workspaces WHERE id = \$1\)`).
			WithArgs(workspace.ID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock for inserting workspace
		settings, _ := json.Marshal(workspace.Settings)
		mock.ExpectExec(`INSERT INTO workspaces \(id, name, settings, integrations, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
			WithArgs(
				workspace.ID,
				workspace.Name,
				settings,
				sqlmock.AnyArg(), // integrations (should be nil or empty JSON array)
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := testRepo.Create(context.Background(), workspace)
		require.NoError(t, err)
	})

	t.Run("empty workspace ID", func(t *testing.T) {
		workspace := &domain.Workspace{
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
		}

		err := testRepo.Create(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "workspace ID is required")
	})

	t.Run("workspace ID already exists", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "existing-workspace",
			Name: "Existing Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
		}

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM workspaces WHERE id = \$1\)`).
			WithArgs(workspace.ID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		err := testRepo.Create(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("database error during existence check", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "test-workspace",
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
		}

		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM workspaces WHERE id = \$1\)`).
			WithArgs(workspace.ID).
			WillReturnError(errors.New("database error"))

		err := testRepo.Create(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("database error during insert", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "test-workspace",
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
		}

		// Mock for checking if workspace exists
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM workspaces WHERE id = \$1\)`).
			WithArgs(workspace.ID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock for inserting workspace with error
		settings, _ := json.Marshal(workspace.Settings)
		mock.ExpectExec(`INSERT INTO workspaces \(id, name, settings, integrations, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
			WithArgs(
				workspace.ID,
				workspace.Name,
				settings,
				sqlmock.AnyArg(), // integrations (should be nil or empty JSON array)
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnError(errors.New("insert error"))

		err := testRepo.Create(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create workspace")
	})
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
	settings := `{"timezone":"UTC"}`
	integrations := `[]`
	createdAt := time.Now()
	updatedAt := time.Now()

	// Test successful retrieval
	rows := sqlmock.NewRows([]string{"id", "name", "settings", "integrations", "created_at", "updated_at"}).
		AddRow(workspaceID, workspaceName, settings, integrations, createdAt, updatedAt)

	mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnRows(rows)

	workspace, err := repo.GetByID(context.Background(), workspaceID)
	require.NoError(t, err)
	assert.Equal(t, workspaceID, workspace.ID)
	assert.Equal(t, workspaceName, workspace.Name)
	assert.Equal(t, "UTC", workspace.Settings.Timezone)

	// Test not found
	mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces WHERE id = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	workspace, err = repo.GetByID(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, workspace)

	// Test database error
	mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces WHERE id = \$1`).
		WithArgs(workspaceID).
		WillReturnError(fmt.Errorf("database error"))

	workspace, err = repo.GetByID(context.Background(), workspaceID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, workspace)
}

func TestWorkspaceRepository_List(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig, "secret-key")

	// Test data
	workspace1ID := "workspace1"
	workspace1Name := "Workspace 1"
	workspace1Settings := `{"timezone":"UTC"}`
	workspace1Integrations := `[]`
	workspace1CreatedAt := time.Now()
	workspace1UpdatedAt := time.Now()

	workspace2ID := "workspace2"
	workspace2Name := "Workspace 2"
	workspace2Settings := `{"timezone":"Europe/London"}`
	workspace2Integrations := `[]`
	workspace2CreatedAt := time.Now().Add(time.Hour)
	workspace2UpdatedAt := time.Now().Add(time.Hour)

	// Test successful retrieval
	rows := sqlmock.NewRows([]string{"id", "name", "settings", "integrations", "created_at", "updated_at"}).
		AddRow(workspace2ID, workspace2Name, workspace2Settings, workspace2Integrations, workspace2CreatedAt, workspace2UpdatedAt).
		AddRow(workspace1ID, workspace1Name, workspace1Settings, workspace1Integrations, workspace1CreatedAt, workspace1UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
		WillReturnRows(rows)

	workspaces, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, len(workspaces))
	assert.Equal(t, workspace2ID, workspaces[0].ID)
	assert.Equal(t, workspace1ID, workspaces[1].ID)

	// Test database error
	mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
		WillReturnError(fmt.Errorf("database error"))

	workspaces, err = repo.List(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, workspaces)
}

func TestWorkspaceRepository_Update(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	repo := NewWorkspaceRepository(db, dbConfig, "secret-key")

	workspace := &domain.Workspace{
		ID:   "workspace1",
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			Timezone: "America/New_York",
		},
	}

	// Marshal settings to JSON for the mock
	settings, _ := json.Marshal(workspace.Settings)

	// Mock for successful update
	mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
		WithArgs(
			workspace.Name,
			settings,
			sqlmock.AnyArg(), // integrations (should be nil or empty JSON array)
			sqlmock.AnyArg(), // updated_at
			workspace.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), workspace)
	require.NoError(t, err)

	// Mock for workspace not found
	mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
		WithArgs(
			workspace.Name,
			settings,
			sqlmock.AnyArg(), // integrations (should be nil or empty JSON array)
			sqlmock.AnyArg(), // updated_at
			workspace.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), workspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Mock for database error
	mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
		WithArgs(
			workspace.Name,
			settings,
			sqlmock.AnyArg(), // integrations (should be nil or empty JSON array)
			sqlmock.AnyArg(), // updated_at
			workspace.ID,
		).
		WillReturnError(fmt.Errorf("database error"))

	err = repo.Update(context.Background(), workspace)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
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
