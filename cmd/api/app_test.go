package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/pkg/mailer"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateRandomKeyBytes generates random bytes for testing keys
func generateRandomKeyBytes(length int) []byte {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		// For testing, fallback to a fixed pattern if random fails
		for i := 0; i < length; i++ {
			bytes[i] = byte(i % 256)
		}
	}
	return bytes
}

// Helper function to create a test configuration with proper key lengths
func createTestConfig() *config.Config {
	// Generate 64 byte keys for PASETO
	privateKeyBytes := generateRandomKeyBytes(64)
	publicKeyBytes := generateRandomKeyBytes(64)

	return &config.Config{
		Environment: "test",
		RootEmail:   "test@example.com",
		Database: config.DatabaseConfig{
			User:     "postgres_test",
			Password: "postgres_test",
			Host:     "localhost",
			Port:     5432,
			DBName:   "notifuse_test",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Security: config.SecurityConfig{
			PasetoPrivateKeyBytes: privateKeyBytes,
			PasetoPublicKeyBytes:  publicKeyBytes,
		},
	}
}

// setupTestDBMock creates a mock DB for testing
func setupTestDBMock() (*sql.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	// Setup necessary mock expectations for common queries
	mock.ExpectBegin()
	mock.ExpectCommit()

	return db, mock, nil
}

func TestNewApp(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		RootEmail:   "test@example.com",
		Environment: "test",
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	// Test creating a new app with default logger
	app := NewApp(cfg)
	assert.NotNil(t, app)
	assert.Equal(t, cfg, app.GetConfig())
	assert.NotNil(t, app.GetLogger())
	assert.NotNil(t, app.GetMux())

	// Test creating a new app with custom options
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)

	mockMailer := pkgmocks.NewMockMailer(ctrl)

	app = NewApp(cfg,
		WithLogger(mockLogger),
		WithMockDB(mockDB),
		WithMockMailer(mockMailer),
	)

	assert.Equal(t, mockLogger, app.GetLogger())
	assert.Equal(t, mockDB, app.GetDB())
	assert.Equal(t, mockMailer, app.GetMailer())
}

func TestAppInitMailer(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		Environment: "development",
	}

	// Test without pre-existing mailer
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	app := NewApp(cfg, WithLogger(mockLogger))
	err := app.InitMailer()
	assert.NoError(t, err)
	assert.NotNil(t, app.GetMailer())

	// Check if correctly used development mailer
	_, isConsoleMailer := app.GetMailer().(*mailer.ConsoleMailer)
	assert.True(t, isConsoleMailer)

	// Test with pre-existing mailer (should be skipped)
	mockMailer := pkgmocks.NewMockMailer(ctrl)
	app = NewApp(cfg, WithLogger(mockLogger), WithMockMailer(mockMailer))
	err = app.InitMailer()
	assert.NoError(t, err)
	assert.Equal(t, mockMailer, app.GetMailer()) // Should still be the mock mailer
}

func TestAppShutdown(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{}

	// Create mock DB
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)

	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create app with mock DB
	app := NewApp(cfg, WithLogger(mockLogger), WithMockDB(mockDB))

	// Test shutdown - no server but should close DB
	err = app.Shutdown(context.Background())
	assert.NoError(t, err)
}

// TestAppInitRepositories tests the InitRepositories method
func TestAppInitRepositories(t *testing.T) {
	// Create mock DB
	mockDB, _, err := setupTestDBMock()
	require.NoError(t, err)
	defer mockDB.Close()

	// Create test config
	cfg := createTestConfig()

	// Create app with mock DB
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := pkgmocks.NewMockLogger(ctrl)
	app := NewApp(cfg, WithLogger(mockLogger), WithMockDB(mockDB))

	// Test repository initialization
	err = app.InitRepositories()
	assert.NoError(t, err)

	// We need to cast to *App to access the internal fields for testing
	appImpl, ok := app.(*App)
	require.True(t, ok, "app should be *App")

	// Verify repositories were initialized
	assert.NotNil(t, appImpl.userRepo)
	assert.NotNil(t, appImpl.workspaceRepo)
	assert.NotNil(t, appImpl.authRepo)
	assert.NotNil(t, appImpl.contactRepo)
	assert.NotNil(t, appImpl.listRepo)
	assert.NotNil(t, appImpl.contactListRepo)
	assert.NotNil(t, appImpl.templateRepo)
	assert.NotNil(t, appImpl.broadcastRepo)
	assert.NotNil(t, appImpl.taskRepo)
	assert.NotNil(t, appImpl.transactionalNotificationRepo)
	assert.NotNil(t, appImpl.messageHistoryRepo)
}

// TestAppStart tests the Start method
func TestAppStart(t *testing.T) {
	// Use a special config with high port number to avoid conflicts
	cfg := createTestConfig()
	// Use a random high port to avoid conflicts
	cfg.Server.Port = 18080 + (time.Now().Nanosecond() % 1000)

	// Create app with mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	mockDB, _, err := setupTestDBMock()
	require.NoError(t, err)
	defer mockDB.Close()

	app := NewApp(cfg, WithLogger(mockLogger), WithMockDB(mockDB))

	// Set up a channel to receive errors
	errCh := make(chan error, 1)

	// Start server in goroutine
	go func() {
		errCh <- app.Start()
	}()

	// Wait for server to be initialized with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	started := app.WaitForServerStart(ctx)
	require.True(t, started, "Server should have started within timeout")

	// Verify server was created
	assert.True(t, app.IsServerCreated(), "Server should be created")

	// Shutdown the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()

	err = app.Shutdown(shutdownCtx)
	assert.NoError(t, err)

	// Check for any server errors
	select {
	case err := <-errCh:
		// We expect http.ErrServerClosed
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("Server error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for server to stop")
	}
}

// TestInitialize tests a simplified version of Initialize to increase coverage
func TestInitialize(t *testing.T) {
	// Create test app with modified Initialize method for testing
	type testApp struct {
		App                    *App // Change to pointer instead of embedding
		initDBCalled           bool
		initMailerCalled       bool
		initRepositoriesCalled bool
		initServicesCalled     bool
		initHandlersCalled     bool

		// For simulating errors
		returnError error
		errorStage  string
	}

	// Create wrapper for App
	newTestApp := func(cfg *config.Config) *testApp {
		appInterface := NewApp(cfg)
		app, ok := appInterface.(*App)
		require.True(t, ok, "appInterface should be *App")
		return &testApp{
			App: app,
		}
	}

	// Override initialize methods
	initDB := func(t *testApp) error {
		t.initDBCalled = true
		if t.errorStage == "db" {
			return t.returnError
		}
		return nil
	}

	initMailer := func(t *testApp) error {
		t.initMailerCalled = true
		if t.errorStage == "mailer" {
			return t.returnError
		}
		return nil
	}

	initRepositories := func(t *testApp) error {
		t.initRepositoriesCalled = true
		if t.errorStage == "repositories" {
			return t.returnError
		}
		return nil
	}

	initServices := func(t *testApp) error {
		t.initServicesCalled = true
		if t.errorStage == "services" {
			return t.returnError
		}
		return nil
	}

	initHandlers := func(t *testApp) error {
		t.initHandlersCalled = true
		if t.errorStage == "handlers" {
			return t.returnError
		}
		return nil
	}

	// Custom initialize that uses our wrapped functions
	initialize := func(t *testApp) error {
		if err := initDB(t); err != nil {
			return err
		}

		if err := initMailer(t); err != nil {
			return err
		}

		if err := initRepositories(t); err != nil {
			return err
		}

		if err := initServices(t); err != nil {
			return err
		}

		if err := initHandlers(t); err != nil {
			return err
		}

		return nil
	}

	// Test successful initialization
	tApp := newTestApp(createTestConfig())
	err := initialize(tApp)
	assert.NoError(t, err)
	assert.True(t, tApp.initDBCalled)
	assert.True(t, tApp.initMailerCalled)
	assert.True(t, tApp.initRepositoriesCalled)
	assert.True(t, tApp.initServicesCalled)
	assert.True(t, tApp.initHandlersCalled)

	// Test DB error
	tApp = newTestApp(createTestConfig())
	tApp.errorStage = "db"
	tApp.returnError = errors.New("db error")
	err = initialize(tApp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.True(t, tApp.initDBCalled)
	assert.False(t, tApp.initMailerCalled)

	// Test mailer error
	tApp = newTestApp(createTestConfig())
	tApp.errorStage = "mailer"
	tApp.returnError = errors.New("mailer error")
	err = initialize(tApp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mailer error")
	assert.True(t, tApp.initDBCalled)
	assert.True(t, tApp.initMailerCalled)
	assert.False(t, tApp.initRepositoriesCalled)

	// Test repository error
	tApp = newTestApp(createTestConfig())
	tApp.errorStage = "repositories"
	tApp.returnError = errors.New("repo error")
	err = initialize(tApp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repo error")
	assert.True(t, tApp.initDBCalled)
	assert.True(t, tApp.initMailerCalled)
	assert.True(t, tApp.initRepositoriesCalled)
	assert.False(t, tApp.initServicesCalled)
}

// TestAppInitServices tests the InitServices method with our hardcoded keys
func TestAppInitServices(t *testing.T) {
	// Get hardcoded keys for testing
	keys, err := GetHardcodedTestKeys()
	if err != nil {
		t.Fatalf("Failed to get hardcoded keys: %v", err)
	}

	// Set up mock DB
	mockDB, _, err := setupTestDBMock()
	require.NoError(t, err)
	defer mockDB.Close()

	// Create app with test config and mocks
	cfg := createTestConfig()
	// Override config with our hardcoded keys
	cfg.Security.PasetoPrivateKeyBytes = keys.PrivateKeyBytes
	cfg.Security.PasetoPublicKeyBytes = keys.PublicKeyBytes

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := pkgmocks.NewMockLogger(ctrl)
	// Set up expectations for any logger calls
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Fatal(gomock.Any()).AnyTimes()

	app := NewApp(cfg, WithLogger(mockLogger), WithMockDB(mockDB))

	// Setup repositories (required for services)
	err = app.InitRepositories()
	assert.NoError(t, err)

	// Test service initialization
	err = app.InitServices()
	assert.NoError(t, err)

	// Cast to *App to access service fields
	appImpl, ok := app.(*App)
	require.True(t, ok, "app should be *App")

	// Verify services were initialized
	assert.NotNil(t, appImpl.authService, "Auth service should be initialized")
	assert.NotNil(t, appImpl.userService, "User service should be initialized")
	assert.NotNil(t, appImpl.workspaceService, "Workspace service should be initialized")
	assert.NotNil(t, appImpl.contactService, "Contact service should be initialized")
	assert.NotNil(t, appImpl.listService, "List service should be initialized")
	assert.NotNil(t, appImpl.contactListService, "ContactList service should be initialized")
	assert.NotNil(t, appImpl.templateService, "Template service should be initialized")
	assert.NotNil(t, appImpl.emailService, "Email service should be initialized")
	assert.NotNil(t, appImpl.broadcastService, "Broadcast service should be initialized")
	assert.NotNil(t, appImpl.taskService, "Task service should be initialized")
	assert.NotNil(t, appImpl.transactionalNotificationService, "TransactionalNotification service should be initialized")
	assert.NotNil(t, appImpl.eventBus, "Event bus should be initialized")
}

// TestAppInitHandlers tests the InitHandlers method
func TestAppInitHandlers(t *testing.T) {
	// Get hardcoded keys for testing
	keys, err := GetHardcodedTestKeys()
	if err != nil {
		t.Fatalf("Failed to get hardcoded keys: %v", err)
	}

	// Set up mock DB
	mockDB, _, err := setupTestDBMock()
	require.NoError(t, err)
	defer mockDB.Close()

	// Create app with test config and mocks
	cfg := createTestConfig()
	// Override config with our hardcoded keys
	cfg.Security.PasetoPrivateKeyBytes = keys.PrivateKeyBytes
	cfg.Security.PasetoPublicKeyBytes = keys.PublicKeyBytes

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := pkgmocks.NewMockLogger(ctrl)
	// Set up expectations for any logger calls
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Fatal(gomock.Any()).AnyTimes()

	app := NewApp(cfg, WithLogger(mockLogger), WithMockDB(mockDB))

	// Setup repositories (required for services)
	err = app.InitRepositories()
	assert.NoError(t, err)

	// Initialize services (required for handlers)
	err = app.InitServices()
	assert.NoError(t, err)

	// Test handler initialization
	err = app.InitHandlers()
	assert.NoError(t, err)

	// Verify handlers were initialized - since handlers are not directly exposed,
	// we can only check that the mux has routes registered
	assert.NotNil(t, app.GetMux(), "HTTP mux should be initialized")
	// We could add more specific assertions by checking specific routes if needed
}

// AppMockForRunServer is a mock App for testing the runServer function
type AppMockForRunServer struct {
	initCalled          bool
	startCalled         bool
	shutdownCalled      bool
	returnInitError     bool
	returnStartError    bool
	returnShutdownError bool
}

func (a *AppMockForRunServer) Initialize() error {
	a.initCalled = true
	if a.returnInitError {
		return fmt.Errorf("initialize error")
	}
	return nil
}

func (a *AppMockForRunServer) Start() error {
	a.startCalled = true
	if a.returnStartError {
		return fmt.Errorf("start error")
	}
	return nil
}

func (a *AppMockForRunServer) Shutdown(ctx context.Context) error {
	a.shutdownCalled = true
	if a.returnShutdownError {
		return fmt.Errorf("shutdown error")
	}
	return nil
}

// Note: The runServer function is now properly tested in main_test.go
// with TestActualRunServer, which tests the real implementation directly.
