package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"notifuse/server/config"
	"notifuse/server/internal/domain"
)

func TestWorkspaceRepository_Create(t *testing.T) {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	db := setupTestDB(t)
	repo := NewWorkspaceRepository(db, &cfg.Database)

	workspace := &domain.Workspace{
		Name: "Test Workspace",
	}

	err = repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.DeleteDatabase(context.Background(), workspace.ID)
		db.Close()
	})

	assert.NotEmpty(t, workspace.ID)
	assert.NotZero(t, workspace.CreatedAt)
	assert.NotZero(t, workspace.UpdatedAt)

	// Verify workspace was created
	savedWorkspace, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, savedWorkspace.ID)
	assert.Equal(t, workspace.Name, savedWorkspace.Name)
}

func TestWorkspaceRepository_GetByID(t *testing.T) {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	db := setupTestDB(t)
	repo := NewWorkspaceRepository(db, &cfg.Database)

	workspace := &domain.Workspace{
		Name: "Test Workspace",
	}
	err = repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.DeleteDatabase(context.Background(), workspace.ID)
		db.Close()
	})

	// Test not found
	_, err = repo.GetByID(context.Background(), uuid.New().String())
	assert.Error(t, err)

	// Test found
	foundWorkspace, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, foundWorkspace.ID)
	assert.Equal(t, workspace.Name, foundWorkspace.Name)
}

func TestWorkspaceRepository_List(t *testing.T) {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	db := setupTestDB(t)
	repo := NewWorkspaceRepository(db, &cfg.Database)

	// Create some workspaces
	workspace1 := &domain.Workspace{Name: "Workspace 1"}
	workspace2 := &domain.Workspace{Name: "Workspace 2"}

	err = repo.Create(context.Background(), workspace1)
	require.NoError(t, err)
	err = repo.Create(context.Background(), workspace2)
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.DeleteDatabase(context.Background(), workspace1.ID)
		repo.DeleteDatabase(context.Background(), workspace2.ID)
		db.Close()
	})

	// List workspaces
	workspaces, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, workspaces, 2)
	assert.Equal(t, workspace2.ID, workspaces[0].ID) // Most recent first
	assert.Equal(t, workspace1.ID, workspaces[1].ID)
}

func TestWorkspaceRepository_Update(t *testing.T) {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	db := setupTestDB(t)
	repo := NewWorkspaceRepository(db, &cfg.Database)

	workspace := &domain.Workspace{
		Name: "Test Workspace",
	}
	err = repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.DeleteDatabase(context.Background(), workspace.ID)
		db.Close()
	})

	// Update workspace
	workspace.Name = "Updated Workspace"
	err = repo.Update(context.Background(), workspace)
	require.NoError(t, err)

	// Verify update
	updatedWorkspace, err := repo.GetByID(context.Background(), workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Workspace", updatedWorkspace.Name)
	assert.True(t, updatedWorkspace.UpdatedAt.After(updatedWorkspace.CreatedAt))
}

func TestWorkspaceRepository_Delete(t *testing.T) {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	db := setupTestDB(t)
	repo := NewWorkspaceRepository(db, &cfg.Database)

	workspace := &domain.Workspace{
		Name: "Test Workspace",
	}
	err = repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.DeleteDatabase(context.Background(), workspace.ID)
		db.Close()
	})

	// Test delete non-existent workspace
	err = repo.Delete(context.Background(), uuid.New().String())
	assert.Error(t, err)

	// Delete workspace
	err = repo.Delete(context.Background(), workspace.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(context.Background(), workspace.ID)
	assert.Error(t, err)
}

func TestWorkspaceRepository_GetConnection(t *testing.T) {
	// Load test configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{EnvFile: ".env.test"})
	require.NoError(t, err)

	db := setupTestDB(t)
	repo := NewWorkspaceRepository(db, &cfg.Database)

	workspace := &domain.Workspace{
		Name: "Test Workspace",
	}
	err = repo.Create(context.Background(), workspace)
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.DeleteDatabase(context.Background(), workspace.ID)
		db.Close()
	})

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
