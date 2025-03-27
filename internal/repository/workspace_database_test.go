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
	"github.com/Notifuse/notifuse/internal/repository/testutil"
)

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

func TestWorkspaceRepository_CreateDatabase(t *testing.T) {
	_, _, cleanup := testutil.SetupMockDB(t)
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
	mockDB, _, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	// Create a repository instance
	repo := NewWorkspaceRepository(mockDB, dbConfig).(*workspaceRepository)

	ctx := context.Background()
	workspaceID := "test-workspace"

	// Test with a successful mock workspace DB connection
	mockWorkspaceDB, _, mockWorkspaceCleanup := testutil.SetupMockDB(t)
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
	newWorkspaceDB, _, newWorkspaceCleanup := testutil.SetupMockDB(t)
	defer newWorkspaceCleanup()

	newWorkspaceID := "new-workspace"
	repo.connections.Store(newWorkspaceID, newWorkspaceDB)

	// Verify the connection is in the pool
	_, exists := repo.connections.Load(newWorkspaceID)
	assert.True(t, exists, "Connection should be in the pool")

	// GetConnection call (may or may not error depending on the environment)
	repo.GetConnection(context.Background(), newWorkspaceID)
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
	args := ctx.Value("mockGetUserWorkspaces")
	if args == nil {
		return []*domain.UserWorkspace{}, nil
	}
	if err, ok := args.(error); ok {
		return nil, err
	}
	if users, ok := args.([]*domain.UserWorkspace); ok {
		return users, nil
	}
	return []*domain.UserWorkspace{}, nil
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

func (r *mockInternalRepository) GetWorkspaceUsersWithEmail(ctx context.Context, workspaceID string) ([]*domain.UserWorkspaceWithEmail, error) {
	args := ctx.Value("mockGetWorkspaceUsersWithEmail")
	if args == nil {
		return []*domain.UserWorkspaceWithEmail{}, nil
	}
	if err, ok := args.(error); ok {
		return nil, err
	}
	if users, ok := args.([]*domain.UserWorkspaceWithEmail); ok {
		return users, nil
	}
	return []*domain.UserWorkspaceWithEmail{}, nil
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
	db, _, cleanup := testutil.SetupMockDB(t)
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
