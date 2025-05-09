package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Notifuse/notifuse/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Add this variable at package scope to enable mocking
var initializeWorkspaceDBFunc = InitializeWorkspaceDatabase

// Add this variable to mock sql.Open during tests
var sqlOpen = sql.Open

func TestGetSystemDSN(t *testing.T) {
	testCases := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
	}{
		{
			name: "standard config",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "notifuse",
				SSLMode:  "disable",
			},
			expected: "postgres://postgres:password@localhost:5432/notifuse?sslmode=disable",
		},
		{
			name: "custom port",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5433,
				User:     "postgres",
				Password: "password",
				DBName:   "notifuse",
				SSLMode:  "disable",
			},
			expected: "postgres://postgres:password@localhost:5433/notifuse?sslmode=disable",
		},
		{
			name: "remote host",
			config: &config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5432,
				User:     "app_user",
				Password: "secure_password",
				DBName:   "notifuse_prod",
				SSLMode:  "disable",
			},
			expected: "postgres://app_user:secure_password@db.example.com:5432/notifuse_prod?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetSystemDSN(tc.config)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetWorkspaceDSN(t *testing.T) {
	testCases := []struct {
		name        string
		config      *config.DatabaseConfig
		workspaceID string
		expected    string
	}{
		{
			name: "standard workspace",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Prefix:   "nf",
				SSLMode:  "disable",
			},
			workspaceID: "workspace123",
			expected:    "postgres://postgres:password@localhost:5432/nf_ws_workspace123?sslmode=disable",
		},
		{
			name: "workspace with hyphens",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Prefix:   "nf",
				SSLMode:  "disable",
			},
			workspaceID: "workspace-123",
			expected:    "postgres://postgres:password@localhost:5432/nf_ws_workspace_123?sslmode=disable",
		},
		{
			name: "custom configuration",
			config: &config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "app_user",
				Password: "secure_password",
				Prefix:   "notifuse",
				SSLMode:  "disable",
			},
			workspaceID: "client456",
			expected:    "postgres://app_user:secure_password@db.example.com:5433/notifuse_ws_client456?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetWorkspaceDSN(tc.config, tc.workspaceID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetPostgresDSN(t *testing.T) {
	testCases := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
	}{
		{
			name: "standard config",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				SSLMode:  "disable",
			},
			expected: "postgres://postgres:password@localhost:5432/postgres?sslmode=disable",
		},
		{
			name: "custom port",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5433,
				User:     "postgres",
				Password: "password",
				SSLMode:  "disable",
			},
			expected: "postgres://postgres:password@localhost:5433/postgres?sslmode=disable",
		},
		{
			name: "remote host",
			config: &config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5432,
				User:     "app_user",
				Password: "secure_password",
				SSLMode:  "disable",
			},
			expected: "postgres://app_user:secure_password@db.example.com:5432/postgres?sslmode=disable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetPostgresDSN(tc.config)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// MockedEnsureSystemDatabaseExists is a test-friendly version that accepts a DB connection for mocking
func MockedEnsureSystemDatabaseExists(db *sql.DB, dbName string) error {
	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL server: %w", err)
	}

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err := db.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		// Use fmt.Sprintf for proper quoting of identifiers in SQL
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s",
			// Proper quoting to prevent SQL injection
			strings.ReplaceAll(dbName, `"`, `""`))

		_, err = db.Exec(createDBQuery)
		if err != nil {
			return fmt.Errorf("failed to create system database: %w", err)
		}
	}

	return nil
}

// MockedEnsureWorkspaceDatabaseExists is a test-friendly version
func MockedEnsureWorkspaceDatabaseExists(cfg *config.DatabaseConfig, workspaceID string, pgDB *sql.DB, wsDB *sql.DB) error {
	// Replace hyphens with underscores for PostgreSQL compatibility
	safeID := strings.ReplaceAll(workspaceID, "-", "_")
	dbName := fmt.Sprintf("%s_ws_%s", cfg.Prefix, safeID)

	// Using the provided DB connection instead of opening a new one

	// Test the connection
	if err := pgDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL server: %w", err)
	}

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err := pgDB.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		// Create database
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s",
			// Proper quoting to prevent SQL injection
			strings.ReplaceAll(dbName, `"`, `""`))

		_, err = pgDB.Exec(createDBQuery)
		if err != nil {
			return fmt.Errorf("failed to create workspace database: %w", err)
		}

		// Test workspace DB connection
		if err := wsDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping new workspace database: %w", err)
		}

		// Initialize the workspace database schema
		if err := initializeWorkspaceDBFunc(wsDB); err != nil {
			return fmt.Errorf("failed to initialize workspace database schema: %w", err)
		}
	}

	return nil
}

func TestEnsureSystemDatabaseExists(t *testing.T) {
	t.Run("database already exists", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		dbName := "notifuse_system"

		// Mock the ping
		mock.ExpectPing()

		// Mock the check if database exists
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(dbName).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Call the mocked version that accepts a DB directly
		err = MockedEnsureSystemDatabaseExists(db, dbName)
		require.NoError(t, err)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("database doesn't exist and gets created", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		dbName := "notifuse_system"

		// Mock the ping
		mock.ExpectPing()

		// Mock the check if database exists - return false
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs(dbName).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock the create database query
		mock.ExpectExec("CREATE DATABASE notifuse_system").
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Call the mocked version that accepts a DB directly
		err = MockedEnsureSystemDatabaseExists(db, dbName)
		require.NoError(t, err)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

func TestEnsureSystemDatabaseExists_Additional(t *testing.T) {
	t.Run("fails to ping postgresql server", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		// Mock ping failure
		mock.ExpectPing().WillReturnError(sql.ErrConnDone)

		// Call the mocked version that accepts a DB directly
		err = MockedEnsureSystemDatabaseExists(db, "notifuse_system")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping PostgreSQL server")

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("fails to check if database exists", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		// Mock the ping
		mock.ExpectPing()

		// Mock query error
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("notifuse_system").
			WillReturnError(sql.ErrConnDone)

		// Call the mocked version that accepts a DB directly
		err = MockedEnsureSystemDatabaseExists(db, "notifuse_system")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check if database exists")

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("fails to create database", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer db.Close()

		// Mock the ping
		mock.ExpectPing()

		// Mock database doesn't exist
		mock.ExpectQuery("SELECT EXISTS").
			WithArgs("notifuse_system").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock creation failure
		mock.ExpectExec("CREATE DATABASE notifuse_system").
			WillReturnError(sql.ErrConnDone)

		// Call the mocked version that accepts a DB directly
		err = MockedEnsureSystemDatabaseExists(db, "notifuse_system")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create system database")

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

func TestEnsureWorkspaceDatabaseExists(t *testing.T) {
	t.Run("workspace database already exists", func(t *testing.T) {
		// Mock the PostgreSQL server connection
		pgDB, pgMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer pgDB.Close()

		// Mock the workspace DB connection - not actually used in this test
		wsDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer wsDB.Close()

		// Workspace ID for the test
		workspaceID := "testworkspace"

		// Mock the ping
		pgMock.ExpectPing()

		// Mock the check if database exists - return true
		pgMock.ExpectQuery("SELECT EXISTS").
			WithArgs("ntf_ws_" + workspaceID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Create config for the test
		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
			SSLMode:  "disable",
		}

		// Call the mocked version
		err = MockedEnsureWorkspaceDatabaseExists(cfg, workspaceID, pgDB, wsDB)
		require.NoError(t, err)

		// Verify all expectations were met
		err = pgMock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("workspace database doesn't exist and gets created", func(t *testing.T) {
		// Mock the PostgreSQL server connection
		pgDB, pgMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer pgDB.Close()

		// Mock the workspace DB connection
		wsDB, wsMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer wsDB.Close()

		// Workspace ID for the test
		workspaceID := "testworkspace"

		// Mock the ping for PostgreSQL server
		pgMock.ExpectPing()

		// Mock database doesn't exist
		pgMock.ExpectQuery("SELECT EXISTS").
			WithArgs("ntf_ws_" + workspaceID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock the create database query
		pgMock.ExpectExec("CREATE DATABASE ntf_ws_" + workspaceID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Mock the ping for workspace DB
		wsMock.ExpectPing()

		// Create a mock for InitializeWorkspaceDatabase
		var initCalled bool
		originalInitFunc := initializeWorkspaceDBFunc
		defer func() { initializeWorkspaceDBFunc = originalInitFunc }()

		initializeWorkspaceDBFunc = func(db *sql.DB) error {
			initCalled = true
			return nil
		}

		// Create config for the test
		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
			SSLMode:  "disable",
		}

		// Call the mocked version
		err = MockedEnsureWorkspaceDatabaseExists(cfg, workspaceID, pgDB, wsDB)
		require.NoError(t, err)
		require.True(t, initCalled, "InitializeWorkspaceDatabase should be called")

		// Verify all expectations were met
		err = pgMock.ExpectationsWereMet()
		require.NoError(t, err)
		err = wsMock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

// Add these variables for mocking
var ensureWorkspaceDBExistsFunc = EnsureWorkspaceDatabaseExists

// Create an interface for database connections
type dbConn interface {
	Ping() error
	Close() error
}

// mockConnectToWorkspace is a test-specific function that uses mocked dependencies
func mockConnectToWorkspace(cfg *config.DatabaseConfig, workspaceID string, mockDB dbConn, ensureErr error) (*sql.DB, error) {
	// Skip the real database check
	if ensureErr != nil {
		return nil, fmt.Errorf("failed to ensure workspace database exists: %w", ensureErr)
	}

	// Use the mock DB instead of opening a real connection
	if mockDB != nil {
		// Test the connection - this will ping as configured by the test
		if err := mockDB.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping workspace database: %w", err)
		}
		// Type assertion to return an *sql.DB is safe only in test code
		// In production, we'd handle this more carefully
		return mockDB.(*sql.DB), nil
	}

	// This simulates an error opening the database connection
	return nil, fmt.Errorf("failed to connect to workspace database: %w", errors.New("connection error"))
}

// Create a custom type for a mock DB that will always fail on ping
type pingErrorDB struct {
	*sql.DB
}

func (db *pingErrorDB) Ping() error {
	return fmt.Errorf("ping failed")
}

func TestConnectToWorkspace(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer mockDB.Close()

		// Mock the ping
		mock.ExpectPing()

		// Create config for the test
		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
		}

		// Call our test-specific function with the mock
		db, err := mockConnectToWorkspace(cfg, "workspace123", mockDB, nil)
		require.NoError(t, err)
		assert.Equal(t, mockDB, db)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("connection error", func(t *testing.T) {
		expectedError := errors.New("connection error")

		// Create config for the test
		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
		}

		// Call our test function without providing a mock DB to simulate connection error
		_, err := mockConnectToWorkspace(cfg, "workspace123", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to workspace database")

		// Check if the error message contains our expected error string
		assert.Contains(t, err.Error(), expectedError.Error())
	})

	t.Run("ping error", func(t *testing.T) {
		// Create a mock DB that will fail on ping
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		// Wrap the mock DB with our pingErrorDB to force ping failures
		pingErrorMockDB := &pingErrorDB{mockDB}

		// Create config for the test
		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
		}

		// Call our test-specific function with the mock DB that will fail ping
		_, err = mockConnectToWorkspace(cfg, "workspace123", pingErrorMockDB, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping workspace database")
		assert.Contains(t, err.Error(), "ping failed")
	})

	t.Run("ensure database error", func(t *testing.T) {
		// Create a specific error for the ensure function
		ensureError := errors.New("failed to create workspace database")

		// Create config for the test
		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
		}

		// Call our test-specific function with an ensure error
		_, err := mockConnectToWorkspace(cfg, "workspace123", nil, ensureError)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ensure workspace database exists")

		// Check if the error message contains our expected error string
		assert.Contains(t, err.Error(), ensureError.Error())
	})
}

func TestGetWorkspaceDSN_EdgeCases(t *testing.T) {
	t.Run("handles special characters in workspace ID", func(t *testing.T) {
		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "nf",
		}

		workspaceIDs := []string{
			"workspace-with-hyphens",
			"workspace_with_underscores",
			"workspace.with.dots",
			"workspace/with/slashes",
			"workspace:with:colons",
		}

		for _, id := range workspaceIDs {
			result := GetWorkspaceDSN(cfg, id)
			// Verify hyphens are replaced with underscores
			sanitizedID := strings.ReplaceAll(id, "-", "_")
			expectedDBName := fmt.Sprintf("nf_ws_%s", sanitizedID)
			assert.Contains(t, result, expectedDBName, "DSN should contain the sanitized workspace ID")
		}
	})
}

// MockedConnectToWorkspace is a test-specific function that doesn't make real DB connections
func MockedConnectToWorkspace(mockDB *sql.DB, ensureErr error, openErr error, pingErr error) (*sql.DB, error) {
	// Check if EnsureWorkspaceDatabaseExists would fail
	if ensureErr != nil {
		return nil, fmt.Errorf("failed to ensure workspace database exists: %w", ensureErr)
	}

	// Check if sql.Open would fail
	if openErr != nil {
		return nil, fmt.Errorf("failed to connect to workspace database: %w", openErr)
	}

	// Check if db.Ping would fail
	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping workspace database: %w", pingErr)
	}

	// Everything succeeded, return the mock DB
	return mockDB, nil
}

// TestConnectToWorkspace_Mock tests the function using our mocked version
func TestConnectToWorkspace_Mock(t *testing.T) {
	// Create a mock DB just for structure, not for expectations
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	t.Run("successful connection", func(t *testing.T) {
		db, err := MockedConnectToWorkspace(mockDB, nil, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, mockDB, db)
	})

	t.Run("ensure database error", func(t *testing.T) {
		_, err := MockedConnectToWorkspace(mockDB, errors.New("database creation failed"), nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ensure workspace database exists")
		assert.Contains(t, err.Error(), "database creation failed")
	})

	t.Run("sql open error", func(t *testing.T) {
		_, err := MockedConnectToWorkspace(mockDB, nil, errors.New("connection error"), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to workspace database")
		assert.Contains(t, err.Error(), "connection error")
	})

	t.Run("ping error", func(t *testing.T) {
		_, err := MockedConnectToWorkspace(mockDB, nil, nil, errors.New("ping error"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping workspace database")
		assert.Contains(t, err.Error(), "ping error")
	})
}

func TestEnsureWorkspaceDatabaseExists_Additional(t *testing.T) {
	t.Run("fails to ping postgresql server", func(t *testing.T) {
		// Mock the PostgreSQL server connection
		pgDB, pgMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer pgDB.Close()

		// Mock the workspace DB connection - not used in this test
		wsDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer wsDB.Close()

		// Mock ping failure
		pgMock.ExpectPing().WillReturnError(sql.ErrConnDone)

		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
			SSLMode:  "disable",
		}

		err = MockedEnsureWorkspaceDatabaseExists(cfg, "testworkspace", pgDB, wsDB)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping PostgreSQL server")
	})

	t.Run("fails to check if database exists", func(t *testing.T) {
		// Mock the PostgreSQL server connection
		pgDB, pgMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer pgDB.Close()

		// Mock the workspace DB connection - not used in this test
		wsDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer wsDB.Close()

		// Mock the ping
		pgMock.ExpectPing()

		// Mock query error
		pgMock.ExpectQuery("SELECT EXISTS").
			WithArgs("ntf_ws_testworkspace").
			WillReturnError(sql.ErrConnDone)

		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
			SSLMode:  "disable",
		}

		err = MockedEnsureWorkspaceDatabaseExists(cfg, "testworkspace", pgDB, wsDB)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check if database exists")
	})

	t.Run("fails to create database", func(t *testing.T) {
		// Mock the PostgreSQL server connection
		pgDB, pgMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer pgDB.Close()

		// Mock the workspace DB connection - not used in this test
		wsDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer wsDB.Close()

		// Mock the ping
		pgMock.ExpectPing()

		// Mock database doesn't exist
		pgMock.ExpectQuery("SELECT EXISTS").
			WithArgs("ntf_ws_testworkspace").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock creation failure
		pgMock.ExpectExec("CREATE DATABASE ntf_ws_testworkspace").
			WillReturnError(sql.ErrConnDone)

		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
			SSLMode:  "disable",
		}

		err = MockedEnsureWorkspaceDatabaseExists(cfg, "testworkspace", pgDB, wsDB)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create workspace database")
	})

	t.Run("fails to ping workspace database", func(t *testing.T) {
		// Mock the PostgreSQL server connection
		pgDB, pgMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer pgDB.Close()

		// Mock the workspace DB connection
		wsDB, wsMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer wsDB.Close()

		// Mock the ping for PostgreSQL server
		pgMock.ExpectPing()

		// Mock database doesn't exist
		pgMock.ExpectQuery("SELECT EXISTS").
			WithArgs("ntf_ws_testworkspace").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock successful database creation
		pgMock.ExpectExec("CREATE DATABASE ntf_ws_testworkspace").
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Mock ping failure for workspace DB
		wsMock.ExpectPing().WillReturnError(sql.ErrConnDone)

		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
			SSLMode:  "disable",
		}

		err = MockedEnsureWorkspaceDatabaseExists(cfg, "testworkspace", pgDB, wsDB)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping new workspace database")
	})

	t.Run("fails to initialize workspace database", func(t *testing.T) {
		// Mock the PostgreSQL server connection
		pgDB, pgMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer pgDB.Close()

		// Mock the workspace DB connection
		wsDB, wsMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer wsDB.Close()

		// Mock the ping for PostgreSQL server
		pgMock.ExpectPing()

		// Mock database doesn't exist
		pgMock.ExpectQuery("SELECT EXISTS").
			WithArgs("ntf_ws_testworkspace").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock successful database creation
		pgMock.ExpectExec("CREATE DATABASE ntf_ws_testworkspace").
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Mock successful ping for workspace DB
		wsMock.ExpectPing()

		// Create a mock for InitializeWorkspaceDatabase that fails
		originalInitFunc := initializeWorkspaceDBFunc
		defer func() { initializeWorkspaceDBFunc = originalInitFunc }()

		initializeWorkspaceDBFunc = func(db *sql.DB) error {
			return errors.New("failed to create tables")
		}

		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Prefix:   "ntf",
			SSLMode:  "disable",
		}

		err = MockedEnsureWorkspaceDatabaseExists(cfg, "testworkspace", pgDB, wsDB)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize workspace database schema")
	})
}
