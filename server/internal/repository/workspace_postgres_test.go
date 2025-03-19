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

	"github.com/google/uuid"
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

func TestWorkspaceRepository_AddUserToWorkspace(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)
	defer db.Close()

	ctx := context.Background()
	workspaceID := generateWorkspaceID()
	defer cleanupTestWorkspace(t, repo, workspaceID)

	// Create a test workspace first
	workspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, workspace)
	require.NoError(t, err)

	// Now test adding a user to the workspace
	userID := uuid.New().String()
	userWorkspace := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        "member",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Test adding a user
	err = repo.AddUserToWorkspace(ctx, userWorkspace)
	assert.NoError(t, err)

	// Test updating existing user role (should use ON CONFLICT)
	userWorkspace.Role = "owner"
	userWorkspace.UpdatedAt = time.Now()
	err = repo.AddUserToWorkspace(ctx, userWorkspace)
	assert.NoError(t, err)

	// Verify the user was added with the updated role
	userWorkspaces, err := repo.GetWorkspaceUsers(ctx, workspaceID)
	assert.NoError(t, err)
	assert.Len(t, userWorkspaces, 1)
	assert.Equal(t, userID, userWorkspaces[0].UserID)
	assert.Equal(t, "owner", userWorkspaces[0].Role)
}

func TestWorkspaceRepository_RemoveUserFromWorkspace(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)
	defer db.Close()

	ctx := context.Background()
	workspaceID := generateWorkspaceID()
	defer cleanupTestWorkspace(t, repo, workspaceID)

	// Create a test workspace first
	workspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, workspace)
	require.NoError(t, err)

	// Add a couple of users to the workspace
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()
	userWorkspace1 := &domain.UserWorkspace{
		UserID:      userID1,
		WorkspaceID: workspaceID,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	userWorkspace2 := &domain.UserWorkspace{
		UserID:      userID2,
		WorkspaceID: workspaceID,
		Role:        "member",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.AddUserToWorkspace(ctx, userWorkspace1)
	require.NoError(t, err)

	err = repo.AddUserToWorkspace(ctx, userWorkspace2)
	require.NoError(t, err)

	// Test removing a user
	err = repo.RemoveUserFromWorkspace(ctx, userID2, workspaceID)
	assert.NoError(t, err)

	// Verify the user was removed
	userWorkspaces, err := repo.GetWorkspaceUsers(ctx, workspaceID)
	assert.NoError(t, err)
	assert.Len(t, userWorkspaces, 1)
	assert.Equal(t, userID1, userWorkspaces[0].UserID)

	// Test removing a user that doesn't exist in the workspace
	nonExistentUserID := uuid.New().String()
	err = repo.RemoveUserFromWorkspace(ctx, nonExistentUserID, workspaceID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user is not a member")
}

func TestWorkspaceRepository_GetUserWorkspaces(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)
	defer db.Close()

	ctx := context.Background()

	// Create a couple of test workspaces
	workspace1 := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Test Workspace 1",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example1.com",
			LogoURL:    "https://example1.com/logo.png",
			Timezone:   "UTC",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	workspace2 := &domain.Workspace{
		ID:   "testworkspace2",
		Name: "Test Workspace 2",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example2.com",
			LogoURL:    "https://example2.com/logo.png",
			Timezone:   "UTC",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	defer cleanupTestWorkspace(t, repo, workspace1.ID)
	defer cleanupTestWorkspace(t, repo, workspace2.ID)

	err := repo.Create(ctx, workspace1)
	require.NoError(t, err)

	err = repo.Create(ctx, workspace2)
	require.NoError(t, err)

	// Add a user to both workspaces
	userID := uuid.New().String()
	userWorkspace1 := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspace1.ID,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	userWorkspace2 := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspace2.ID,
		Role:        "member",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.AddUserToWorkspace(ctx, userWorkspace1)
	require.NoError(t, err)

	err = repo.AddUserToWorkspace(ctx, userWorkspace2)
	require.NoError(t, err)

	// Test getting a user's workspaces
	userWorkspaces, err := repo.GetUserWorkspaces(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, userWorkspaces, 2)

	// Verify the workspace IDs and roles
	workspaceIDs := make(map[string]string) // map of workspace ID to role
	for _, uw := range userWorkspaces {
		workspaceIDs[uw.WorkspaceID] = uw.Role
	}
	assert.Contains(t, workspaceIDs, workspace1.ID)
	assert.Contains(t, workspaceIDs, workspace2.ID)
	assert.Equal(t, "owner", workspaceIDs[workspace1.ID])
	assert.Equal(t, "member", workspaceIDs[workspace2.ID])

	// Test getting workspaces for a user that doesn't have any
	nonExistentUserID := uuid.New().String()
	userWorkspaces, err = repo.GetUserWorkspaces(ctx, nonExistentUserID)
	assert.NoError(t, err)
	assert.Empty(t, userWorkspaces)
}

func TestWorkspaceRepository_GetWorkspaceUsers(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)
	defer db.Close()

	ctx := context.Background()
	workspaceID := generateWorkspaceID()
	defer cleanupTestWorkspace(t, repo, workspaceID)

	// Create a test workspace
	workspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, workspace)
	require.NoError(t, err)

	// Add multiple users to the workspace
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()
	userID3 := uuid.New().String()
	userWorkspace1 := &domain.UserWorkspace{
		UserID:      userID1,
		WorkspaceID: workspaceID,
		Role:        "owner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	userWorkspace2 := &domain.UserWorkspace{
		UserID:      userID2,
		WorkspaceID: workspaceID,
		Role:        "member",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	userWorkspace3 := &domain.UserWorkspace{
		UserID:      userID3,
		WorkspaceID: workspaceID,
		Role:        "member",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.AddUserToWorkspace(ctx, userWorkspace1)
	require.NoError(t, err)

	err = repo.AddUserToWorkspace(ctx, userWorkspace2)
	require.NoError(t, err)

	err = repo.AddUserToWorkspace(ctx, userWorkspace3)
	require.NoError(t, err)

	// Test getting all users for a workspace
	userWorkspaces, err := repo.GetWorkspaceUsers(ctx, workspaceID)
	assert.NoError(t, err)
	assert.Len(t, userWorkspaces, 3)

	// Verify we have the expected users and roles
	userRoles := make(map[string]string) // map of user ID to role
	for _, uw := range userWorkspaces {
		userRoles[uw.UserID] = uw.Role
	}
	assert.Contains(t, userRoles, userID1)
	assert.Contains(t, userRoles, userID2)
	assert.Contains(t, userRoles, userID3)
	assert.Equal(t, "owner", userRoles[userID1])
	assert.Equal(t, "member", userRoles[userID2])
	assert.Equal(t, "member", userRoles[userID3])

	// Test getting users for a workspace that doesn't exist
	userWorkspaces, err = repo.GetWorkspaceUsers(ctx, "nonexistentworkspace")
	assert.NoError(t, err)
	assert.Empty(t, userWorkspaces)
}

func TestWorkspaceRepository_GetUserWorkspace(t *testing.T) {
	db, repo, _ := setupWorkspaceTest(t)
	defer db.Close()

	ctx := context.Background()
	workspaceID := generateWorkspaceID()
	defer cleanupTestWorkspace(t, repo, workspaceID)

	// Create a test workspace
	workspace := &domain.Workspace{
		ID:   workspaceID,
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, workspace)
	require.NoError(t, err)

	// Add a user to the workspace
	now := time.Now()
	userID := uuid.New().String()
	userWorkspace := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        "owner",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = repo.AddUserToWorkspace(ctx, userWorkspace)
	require.NoError(t, err)

	// Test getting a specific user-workspace relationship
	uw, err := repo.GetUserWorkspace(ctx, userID, workspaceID)
	assert.NoError(t, err)
	assert.NotNil(t, uw)
	assert.Equal(t, userID, uw.UserID)
	assert.Equal(t, workspaceID, uw.WorkspaceID)
	assert.Equal(t, "owner", uw.Role)

	// Test getting a user-workspace relationship that doesn't exist
	nonExistentUserID := uuid.New().String()
	uw, err = repo.GetUserWorkspace(ctx, nonExistentUserID, workspaceID)
	assert.Error(t, err)
	assert.Nil(t, uw)
	assert.Contains(t, err.Error(), "user is not a member")

	// Test getting a user-workspace relationship for a workspace that doesn't exist
	uw, err = repo.GetUserWorkspace(ctx, userID, "nonexistentworkspace")
	assert.Error(t, err)
	assert.Nil(t, uw)
	assert.Contains(t, err.Error(), "user is not a member")
}
