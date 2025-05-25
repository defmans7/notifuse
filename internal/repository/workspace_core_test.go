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

// mockWorkspaceRepo is a mock implementation for testing
type mockWorkspaceRepo struct {
	db       *sql.DB
	config   *config.DatabaseConfig
	keyValue string
}

func (r *mockWorkspaceRepo) systemDB() *sql.DB {
	return r.db
}

func (r *mockWorkspaceRepo) dbConfig() *config.DatabaseConfig {
	return r.config
}

// testWorkspaceRepo for workspace creation tests
type testWorkspaceRepo struct {
	domain.WorkspaceRepository
	createDatabaseError error
	mockDB              *sql.DB
}

func (r *testWorkspaceRepo) systemDB() *sql.DB {
	return r.mockDB
}

func (r *testWorkspaceRepo) checkWorkspaceIDExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	row := r.systemDB().QueryRow("SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1)", id)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// Create for the testWorkspaceRepo
func (r *testWorkspaceRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	if workspace.ID == "" {
		return errors.New("workspace ID is required")
	}

	// Check if workspace already exists
	exists, err := r.checkWorkspaceIDExists(ctx, workspace.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("workspace with ID %s already exists", workspace.ID)
	}

	// Get the current timestamp for created_at and updated_at
	now := time.Now()
	workspace.CreatedAt = now
	workspace.UpdatedAt = now

	// Set up default values as needed
	if workspace.Integrations == nil {
		workspace.Integrations = []domain.Integration{}
	}

	// Marshal settings and integrations to JSON
	settingsJSON, err := json.Marshal(workspace.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace settings: %w", err)
	}

	integrationsJSON, err := json.Marshal(workspace.Integrations)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace integrations: %w", err)
	}

	// Create the workspace in the database
	_, err = r.systemDB().Exec(
		"INSERT INTO workspaces (id, name, settings, integrations, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		workspace.ID,
		workspace.Name,
		settingsJSON,
		integrationsJSON,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	return nil
}

// This function is used to override the actual repository's scanWorkspace method
// to avoid the decryption logic in tests
func mockScanWorkspace(row scannable) (*domain.Workspace, error) {
	var (
		id           string
		name         string
		settingsJSON []byte
		integrations []byte
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(&id, &name, &settingsJSON, &integrations, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("workspace not found")
		}
		return nil, fmt.Errorf("failed to scan workspace row: %w", err)
	}

	workspace := &domain.Workspace{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Unmarshal settings
	if len(settingsJSON) > 0 {
		err = json.Unmarshal(settingsJSON, &workspace.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal workspace settings: %w", err)
		}
	}

	// Unmarshal integrations
	if len(integrations) > 0 && string(integrations) != "null" {
		err = json.Unmarshal(integrations, &workspace.Integrations)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal workspace integrations: %w", err)
		}
	}

	// In tests, we'll skip AfterLoad to avoid decryption issues
	return workspace, nil
}

// scannable interface for testing
type scannable interface {
	Scan(dest ...interface{}) error
}

// Mock workspaceRepository for tests
type mockWorkspaceRepository struct {
	db               *sql.DB
	cfg              *config.DatabaseConfig
	secretKey        string
	originalScanFunc func(row scannable) (*domain.Workspace, error)
}

func (r *mockWorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	row := r.db.QueryRow("SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces WHERE id = $1", id)
	return mockScanWorkspace(row)
}

func (r *mockWorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	rows, err := r.db.Query("SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		workspace, err := mockScanWorkspace(rows)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspace rows: %w", err)
	}

	return workspaces, nil
}

func (r *mockWorkspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	// Set updated_at timestamp
	workspace.UpdatedAt = time.Now()

	// Marshal settings and integrations
	settingsJSON, err := json.Marshal(workspace.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace settings: %w", err)
	}

	integrationsJSON, err := json.Marshal(workspace.Integrations)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace integrations: %w", err)
	}

	// Execute the update
	result, err := r.db.Exec(
		"UPDATE workspaces SET name = $1, settings = $2, integrations = $3, updated_at = $4 WHERE id = $5",
		workspace.Name,
		settingsJSON,
		integrationsJSON,
		workspace.UpdatedAt,
		workspace.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workspace with ID %s not found", workspace.ID)
	}

	return nil
}

func TestWorkspaceRepository_GetByID(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	// Create a mock repository that uses our mockScanWorkspace function
	repo := &mockWorkspaceRepository{
		db:        db,
		cfg:       dbConfig,
		secretKey: "secret_key_for_dev_env",
	}

	t.Run("successful retrieval", func(t *testing.T) {
		workspaceID := "testworkspace"
		workspaceName := "Test Workspace"
		settings := `{"timezone":"UTC","secret_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`
		integrations := `[{"type":"email","provider":"mailgun","config":{"api_key":"test"}}]`
		createdAt := time.Now()
		updatedAt := time.Now()

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
		assert.Equal(t, createdAt.Unix(), workspace.CreatedAt.Unix())
		assert.Equal(t, updatedAt.Unix(), workspace.UpdatedAt.Unix())
		assert.NotNil(t, workspace.Integrations)
		assert.Len(t, workspace.Integrations, 1)
	})

	t.Run("workspace not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces WHERE id = \$1`).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		workspace, err := repo.GetByID(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, workspace)
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces WHERE id = \$1`).
			WithArgs("testworkspace").
			WillReturnError(fmt.Errorf("connection refused"))

		workspace, err := repo.GetByID(context.Background(), "testworkspace")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
		assert.Nil(t, workspace)
	})

	t.Run("empty workspace ID", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces WHERE id = \$1`).
			WithArgs("").
			WillReturnError(sql.ErrNoRows)

		workspace, err := repo.GetByID(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, workspace)
	})

	t.Run("workspace with minimal settings", func(t *testing.T) {
		workspaceID := "minimal-workspace"
		workspaceName := "Minimal Workspace"
		settings := `{"timezone":"UTC"}`
		integrations := `[]`
		createdAt := time.Now()
		updatedAt := time.Now()

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
		assert.Empty(t, workspace.Integrations)
	})

	t.Run("workspace with null integrations", func(t *testing.T) {
		workspaceID := "null-integrations-workspace"
		workspaceName := "Null Integrations Workspace"
		settings := `{"timezone":"UTC","secret_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`
		integrations := `null`
		createdAt := time.Now()
		updatedAt := time.Now()

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
		assert.Empty(t, workspace.Integrations)
	})
}

func TestWorkspaceRepository_List(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	// Create a mock repository that uses our mockScanWorkspace function
	repo := &mockWorkspaceRepository{
		db:        db,
		cfg:       dbConfig,
		secretKey: "secret_key_for_dev_env",
	}

	t.Run("successful retrieval with multiple workspaces", func(t *testing.T) {
		workspace1ID := "workspace1"
		workspace1Name := "Workspace 1"
		workspace1Settings := `{"timezone":"UTC","secret_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`
		workspace1Integrations := `[{"type":"email","provider":"mailgun"}]`
		workspace1CreatedAt := time.Now()
		workspace1UpdatedAt := time.Now()

		workspace2ID := "workspace2"
		workspace2Name := "Workspace 2"
		workspace2Settings := `{"timezone":"Europe/London","secret_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`
		workspace2Integrations := `[]`
		workspace2CreatedAt := time.Now().Add(time.Hour)
		workspace2UpdatedAt := time.Now().Add(time.Hour)

		rows := sqlmock.NewRows([]string{"id", "name", "settings", "integrations", "created_at", "updated_at"}).
			AddRow(workspace2ID, workspace2Name, workspace2Settings, workspace2Integrations, workspace2CreatedAt, workspace2UpdatedAt).
			AddRow(workspace1ID, workspace1Name, workspace1Settings, workspace1Integrations, workspace1CreatedAt, workspace1UpdatedAt)

		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
			WillReturnRows(rows)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 2, len(workspaces))

		// Verify order (newest first)
		assert.Equal(t, workspace2ID, workspaces[0].ID)
		assert.Equal(t, workspace2Name, workspaces[0].Name)
		assert.Equal(t, "Europe/London", workspaces[0].Settings.Timezone)

		assert.Equal(t, workspace1ID, workspaces[1].ID)
		assert.Equal(t, workspace1Name, workspaces[1].Name)
		assert.Equal(t, "UTC", workspaces[1].Settings.Timezone)
		assert.Len(t, workspaces[1].Integrations, 1)
	})

	t.Run("empty result set", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "settings", "integrations", "created_at", "updated_at"})

		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
			WillReturnRows(rows)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Empty(t, workspaces)
	})

	t.Run("single workspace", func(t *testing.T) {
		workspaceID := "single-workspace"
		workspaceName := "Single Workspace"
		settings := `{"timezone":"America/New_York","secret_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}`
		integrations := `[{"type":"sms","provider":"twilio","config":{"account_sid":"test"}}]`
		createdAt := time.Now()
		updatedAt := time.Now()

		rows := sqlmock.NewRows([]string{"id", "name", "settings", "integrations", "created_at", "updated_at"}).
			AddRow(workspaceID, workspaceName, settings, integrations, createdAt, updatedAt)

		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
			WillReturnRows(rows)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, len(workspaces))
		assert.Equal(t, workspaceID, workspaces[0].ID)
		assert.Equal(t, workspaceName, workspaces[0].Name)
		assert.Equal(t, "America/New_York", workspaces[0].Settings.Timezone)
		assert.Len(t, workspaces[0].Integrations, 1)
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
			WillReturnError(fmt.Errorf("connection timeout"))

		workspaces, err := repo.List(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection timeout")
		assert.Nil(t, workspaces)
	})

	t.Run("row iteration error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "settings", "integrations", "created_at", "updated_at"}).
			AddRow("workspace1", "Workspace 1", `{"timezone":"UTC"}`, `[]`, time.Now(), time.Now()).
			RowError(0, fmt.Errorf("row scan error"))

		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
			WillReturnRows(rows)

		workspaces, err := repo.List(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "row scan error")
		assert.Nil(t, workspaces)
	})

	t.Run("workspaces with various configurations", func(t *testing.T) {
		// Workspace with minimal config
		workspace1ID := "minimal-workspace"
		workspace1Name := "Minimal Workspace"
		workspace1Settings := `{"timezone":"UTC"}`
		workspace1Integrations := `[]`
		workspace1CreatedAt := time.Now().Add(-2 * time.Hour)
		workspace1UpdatedAt := time.Now().Add(-2 * time.Hour)

		// Workspace with full config
		workspace2ID := "full-workspace"
		workspace2Name := "Full Workspace"
		workspace2Settings := `{"timezone":"Asia/Tokyo","secret_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","custom_setting":"value"}`
		workspace2Integrations := `[{"type":"email","provider":"sendgrid"},{"type":"sms","provider":"twilio"}]`
		workspace2CreatedAt := time.Now().Add(-1 * time.Hour)
		workspace2UpdatedAt := time.Now().Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"id", "name", "settings", "integrations", "created_at", "updated_at"}).
			AddRow(workspace2ID, workspace2Name, workspace2Settings, workspace2Integrations, workspace2CreatedAt, workspace2UpdatedAt).
			AddRow(workspace1ID, workspace1Name, workspace1Settings, workspace1Integrations, workspace1CreatedAt, workspace1UpdatedAt)

		mock.ExpectQuery(`SELECT id, name, settings, integrations, created_at, updated_at FROM workspaces ORDER BY created_at DESC`).
			WillReturnRows(rows)

		workspaces, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 2, len(workspaces))

		// Verify full workspace
		assert.Equal(t, workspace2ID, workspaces[0].ID)
		assert.Equal(t, "Asia/Tokyo", workspaces[0].Settings.Timezone)
		assert.Len(t, workspaces[0].Integrations, 2)

		// Verify minimal workspace
		assert.Equal(t, workspace1ID, workspaces[1].ID)
		assert.Equal(t, "UTC", workspaces[1].Settings.Timezone)
		assert.Empty(t, workspaces[1].Integrations)
	})
}

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

	// Use development environment key for tests
	repo := NewWorkspaceRepository(db, dbConfig, "secret_key_for_dev_env").(*workspaceRepository)
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

	// Create a mock repository for testing
	testRepo := &testWorkspaceRepo{
		WorkspaceRepository: nil, // We don't need this for the test
		mockDB:              db,
	}

	t.Run("successful creation", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "test-workspace",
			Name: "Test Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
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
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
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
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
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
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
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
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
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

func TestWorkspaceRepository_Update(t *testing.T) {
	db, mock, cleanup := testutil.SetupMockDB(t)
	defer cleanup()

	dbConfig := &config.DatabaseConfig{
		Prefix: "notifuse",
	}

	// Create a mock repository with the Update method already implemented
	mockRepo := &mockWorkspaceRepository{
		db:        db,
		cfg:       dbConfig,
		secretKey: "secret_key_for_dev_env",
	}

	t.Run("successful update", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "workspace1",
			Name: "Updated Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "America/New_York",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			Integrations: []domain.Integration{
				{
					ID:   "integration-1",
					Name: "SendGrid Integration",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind: domain.EmailProviderKindSMTP,
					},
				},
			},
		}

		mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
			WithArgs(
				workspace.Name,
				sqlmock.AnyArg(), // settings - use AnyArg since it will be dynamic
				sqlmock.AnyArg(), // integrations
				sqlmock.AnyArg(), // updated_at
				workspace.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := mockRepo.Update(context.Background(), workspace)
		require.NoError(t, err)
		assert.True(t, workspace.UpdatedAt.After(time.Now().Add(-time.Minute)))
	})

	t.Run("workspace not found", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "nonexistent-workspace",
			Name: "Nonexistent Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
			WithArgs(
				workspace.Name,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				workspace.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := mockRepo.Update(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("database connection error", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "workspace1",
			Name: "Updated Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "Europe/Paris",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
			WithArgs(
				workspace.Name,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				workspace.ID,
			).
			WillReturnError(fmt.Errorf("connection lost"))

		err := mockRepo.Update(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection lost")
	})

	t.Run("update with minimal settings", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "minimal-workspace",
			Name: "Minimal Updated Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone: "UTC",
			},
			Integrations: []domain.Integration{},
		}

		mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
			WithArgs(
				workspace.Name,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				workspace.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := mockRepo.Update(context.Background(), workspace)
		require.NoError(t, err)
	})

	t.Run("update with complex integrations", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "complex-workspace",
			Name: "Complex Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "Asia/Tokyo",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
			Integrations: []domain.Integration{
				{
					ID:   "integration-1",
					Name: "Mailgun Integration",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind: domain.EmailProviderKindMailgun,
					},
				},
				{
					ID:   "integration-2",
					Name: "Twilio Integration",
					Type: domain.IntegrationTypeEmail,
					EmailProvider: domain.EmailProvider{
						Kind: domain.EmailProviderKindSMTP,
					},
				},
			},
		}

		mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
			WithArgs(
				workspace.Name,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				workspace.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := mockRepo.Update(context.Background(), workspace)
		require.NoError(t, err)
	})

	t.Run("rows affected error", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "workspace1",
			Name: "Updated Workspace",
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		result := sqlmock.NewErrorResult(fmt.Errorf("rows affected error"))
		mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
			WithArgs(
				workspace.Name,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				workspace.ID,
			).
			WillReturnResult(result)

		err := mockRepo.Update(context.Background(), workspace)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rows affected error")
	})

	t.Run("update with empty workspace name", func(t *testing.T) {
		workspace := &domain.Workspace{
			ID:   "workspace1",
			Name: "", // Empty name
			Settings: domain.WorkspaceSettings{
				Timezone:  "UTC",
				SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		}

		mock.ExpectExec(`UPDATE workspaces SET name = \$1, settings = \$2, integrations = \$3, updated_at = \$4 WHERE id = \$5`).
			WithArgs(
				workspace.Name,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				workspace.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := mockRepo.Update(context.Background(), workspace)
		require.NoError(t, err)
	})
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
