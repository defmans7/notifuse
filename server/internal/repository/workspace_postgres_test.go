package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"notifuse/server/config"
	"notifuse/server/internal/domain"
)

func generateWorkspaceID() string {
	return "testworkspace1"
}

func cleanupTestWorkspace(t *testing.T, repo domain.WorkspaceRepository, id string) {
	ctx := context.Background()

	// Try to get the connection first
	conn, err := repo.GetConnection(ctx, id)
	if err == nil && conn != nil {
		// Close the connection if it exists
		conn.Close()
		// Give it a moment to fully close
		time.Sleep(100 * time.Millisecond)
	}

	// Now try to delete the database
	if err := repo.DeleteDatabase(ctx, id); err != nil {
		// Only log if it's not because the database doesn't exist
		if !isNotExistsError(err) {
			t.Logf("Warning: Failed to cleanup test workspace database: %v", err)
		}
	}

	// Finally try to delete the workspace record
	if err := repo.Delete(ctx, id); err != nil {
		// Only log if it's not because the workspace doesn't exist
		if !isNotFoundError(err) {
			t.Logf("Warning: Failed to cleanup test workspace record: %v", err)
		}
	}
}

// isNotExistsError checks if the error is because the database doesn't exist
func isNotExistsError(err error) bool {
	return err != nil && (err.Error() == "database does not exist" ||
		err.Error() == "workspace not found")
}

// isNotFoundError checks if the error is because the workspace doesn't exist
func isNotFoundError(err error) bool {
	return err != nil && err.Error() == "workspace not found"
}

func setupWorkspaceTest(t *testing.T) (*sql.DB, domain.WorkspaceRepository, *config.Config) {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	db := setupTestDB(t)
	repo := NewWorkspaceRepository(db, &cfg.Database)

	// Clean up any existing test workspace before starting
	cleanupTestWorkspace(t, repo, generateWorkspaceID())

	return db, repo, cfg
}

func TestWorkspaceRepository_Create(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)

	workspace := &domain.Workspace{
		ID:   generateWorkspaceID(),
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	t.Cleanup(func() {
		cleanupTestWorkspace(t, repo, workspace.ID)
		db.Close()
	})

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	assert.NotEmpty(t, workspace.ID)
	assert.NotZero(t, workspace.CreatedAt)
	assert.NotZero(t, workspace.UpdatedAt)

	// Verify workspace was created
	savedWorkspace, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, savedWorkspace.ID)
	assert.Equal(t, workspace.Name, savedWorkspace.Name)
	assert.Equal(t, workspace.Settings.WebsiteURL, savedWorkspace.Settings.WebsiteURL)
	assert.Equal(t, workspace.Settings.LogoURL, savedWorkspace.Settings.LogoURL)
	assert.Equal(t, workspace.Settings.Timezone, savedWorkspace.Settings.Timezone)
}

func TestWorkspaceRepository_GetByID(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)

	workspace := &domain.Workspace{
		ID:   generateWorkspaceID(),
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	t.Cleanup(func() {
		cleanupTestWorkspace(t, repo, workspace.ID)
		db.Close()
	})

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Test not found
	_, err = repo.GetByID(context.Background(), "nonexistent1")
	assert.Error(t, err)

	// Test found
	foundWorkspace, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, foundWorkspace.ID)
	assert.Equal(t, workspace.Name, foundWorkspace.Name)
	assert.Equal(t, workspace.Settings.WebsiteURL, foundWorkspace.Settings.WebsiteURL)
	assert.Equal(t, workspace.Settings.LogoURL, foundWorkspace.Settings.LogoURL)
	assert.Equal(t, workspace.Settings.Timezone, foundWorkspace.Settings.Timezone)
}

func TestWorkspaceRepository_List(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)

	// Create some workspaces
	workspace1 := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Workspace 1",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example1.com",
			LogoURL:    "https://example1.com/logo.png",
			Timezone:   "UTC",
		},
	}
	workspace2 := &domain.Workspace{
		ID:   "testworkspace2",
		Name: "Workspace 2",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example2.com",
			LogoURL:    "https://example2.com/logo.png",
			Timezone:   "UTC",
		},
	}

	t.Cleanup(func() {
		cleanupTestWorkspace(t, repo, workspace1.ID)
		cleanupTestWorkspace(t, repo, workspace2.ID)
		db.Close()
	})

	err := repo.Create(context.Background(), workspace1)
	require.NoError(t, err)
	err = repo.Create(context.Background(), workspace2)
	require.NoError(t, err)

	// List workspaces
	workspaces, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, workspaces, 2)
	assert.Equal(t, workspace2.ID, workspaces[0].ID) // Most recent first
	assert.Equal(t, workspace1.ID, workspaces[1].ID)
}

func TestWorkspaceRepository_Update(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)

	workspace := &domain.Workspace{
		ID:   generateWorkspaceID(),
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	t.Cleanup(func() {
		cleanupTestWorkspace(t, repo, workspace.ID)
		db.Close()
	})

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Update workspace
	workspace.Name = "Updated Workspace"
	workspace.Settings.WebsiteURL = "https://updated.com"
	workspace.Settings.LogoURL = "https://updated.com/logo.png"
	err = repo.Update(context.Background(), workspace)
	require.NoError(t, err)

	// Verify update
	updatedWorkspace, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Workspace", updatedWorkspace.Name)
	assert.Equal(t, "https://updated.com", updatedWorkspace.Settings.WebsiteURL)
	assert.Equal(t, "https://updated.com/logo.png", updatedWorkspace.Settings.LogoURL)
	assert.True(t, updatedWorkspace.UpdatedAt.After(updatedWorkspace.CreatedAt))
}

func TestWorkspaceRepository_Delete(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)

	workspace := &domain.Workspace{
		ID:   generateWorkspaceID(),
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	t.Cleanup(func() {
		cleanupTestWorkspace(t, repo, workspace.ID)
		db.Close()
	})

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Test delete non-existent workspace
	err = repo.Delete(context.Background(), "nonexistent1")
	assert.Error(t, err)

	// Delete workspace
	err = repo.Delete(context.Background(), workspace.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(context.Background(), workspace.ID)
	assert.Error(t, err)
}

func TestWorkspaceRepository_GetConnection(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)

	workspace := &domain.Workspace{
		ID:   generateWorkspaceID(),
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}

	t.Cleanup(func() {
		cleanupTestWorkspace(t, repo, workspace.ID)
		db.Close()
	})

	err := repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	// Get connection
	conn, err := repo.GetConnection(context.Background(), workspace.ID)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Test connection
	err = conn.Ping()
	require.NoError(t, err)

	// Get same connection again (should be cached)
	conn2, err := repo.GetConnection(context.Background(), workspace.ID)
	require.NoError(t, err)
	require.NotNil(t, conn2)
	assert.Equal(t, conn, conn2)
}
