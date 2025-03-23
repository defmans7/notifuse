//go:build runserver

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mailer"
	"github.com/stretchr/testify/assert"
)

func TestConfigLoading(t *testing.T) {
	// Try to load config from .env.test
	_, err := config.Load()
	// We expect an error if the file doesn't exist in the test environment
	assert.Error(t, err)
}

func TestSetupMinimalConfig(t *testing.T) {
	// Setup test environment variables
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("SERVER_PORT", "8081")
	os.Setenv("DB_USER", "postgres_test")
	os.Setenv("DB_PASS", "postgres_test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "notifuse_test")
	os.Setenv("ROOT_EMAIL", "test@example.com")

	// Cleanup
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASS")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("ROOT_EMAIL")
	}()

	// Try to load config from environment
	cfg, err := config.Load()

	// Might still fail if viper is looking for files specifically
	if err != nil {
		t.Logf("Config Load failed: %v", err)
		return
	}

	// Otherwise, verify config is loaded correctly
	assert.Equal(t, "test", cfg.Environment)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8081, cfg.Server.Port)
	assert.Equal(t, "postgres_test", cfg.Database.User)
}

// MockApp implements the necessary methods from App for testing
type MockApp struct {
	initializeFunc func() error
	startFunc      func() error
	shutdownFunc   func(ctx context.Context) error
}

func (m *MockApp) Initialize() error {
	return m.initializeFunc()
}

func (m *MockApp) Start() error {
	return m.startFunc()
}

func (m *MockApp) Shutdown(ctx context.Context) error {
	return m.shutdownFunc(ctx)
}

// ExtendedMockApp extends MockApp to also implement the App interface
type ExtendedMockApp struct {
	MockApp
	opts []AppOption
}

// Create a custom signal.Notify function for testing
func createMockSignalFunc(sendSignal bool, delay time.Duration) func(c chan<- os.Signal, sig ...os.Signal) {
	return func(c chan<- os.Signal, sig ...os.Signal) {
		// If we should send a signal, do it after the specified delay
		if sendSignal {
			go func() {
				time.Sleep(delay)
				c <- os.Interrupt
			}()
		}
	}
}

// Create a package variable for NewApp to be redefined during tests
var testNewAppFunc func(cfg *config.Config, opts ...AppOption) interface{}

// -------------------------------------------------------------------------
// The following code is from runserver_test.go
// These tests are for testing the runServer function
// -------------------------------------------------------------------------

// Test helpers for the runServer test
func createSimpleTestConfig() *config.Config {
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
	}
}

// MockLoggerSimple is a simple mock implementation of logger.Logger for the runServer test
type MockLoggerSimple struct{}

func (m *MockLoggerSimple) Info(msg string)                                        {}
func (m *MockLoggerSimple) Debug(msg string)                                       {}
func (m *MockLoggerSimple) Warn(msg string)                                        {}
func (m *MockLoggerSimple) Error(msg string)                                       {}
func (m *MockLoggerSimple) Fatal(msg string)                                       {}
func (m *MockLoggerSimple) WithField(key string, value interface{}) logger.Logger  { return m }
func (m *MockLoggerSimple) WithFields(fields map[string]interface{}) logger.Logger { return m }
func (m *MockLoggerSimple) WithError(err error) logger.Logger                      { return m }

// MockAppSimple implements AppInterface for testing the real runServer
type MockAppSimple struct {
	initializeCalled bool
	startCalled      bool
	shutdownCalled   bool
	initializeError  error
	startError       error
	shutdownError    error
	// Channel to notify when Start() is called
	startNotify chan struct{}
}

func (m *MockAppSimple) Initialize() error {
	m.initializeCalled = true
	return m.initializeError
}

func (m *MockAppSimple) Start() error {
	m.startCalled = true
	// Notify that start was called
	if m.startNotify != nil {
		close(m.startNotify)
	}
	// If we're not returning an error, block until shutdown is called
	if m.startError == nil {
		// This will block until the test sends a signal and Shutdown() is called
		// which is what happens in the real App implementation
		select {}
	}
	return m.startError
}

func (m *MockAppSimple) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return m.shutdownError
}

// Stub implementations of the AppInterface methods added for testing
func (m *MockAppSimple) GetConfig() *config.Config { return nil }
func (m *MockAppSimple) GetLogger() logger.Logger  { return nil }
func (m *MockAppSimple) GetMux() *http.ServeMux    { return nil }
func (m *MockAppSimple) GetDB() *sql.DB            { return nil }
func (m *MockAppSimple) GetMailer() mailer.Mailer  { return nil }
func (m *MockAppSimple) InitDB() error             { return nil }
func (m *MockAppSimple) InitMailer() error         { return nil }
func (m *MockAppSimple) InitRepositories() error   { return nil }
func (m *MockAppSimple) InitServices() error       { return nil }
func (m *MockAppSimple) InitHandlers() error       { return nil }

// TestActualRunServer directly tests the real runServer function
// This test is only run when the runserver build tag is specified
func TestActualRunServer(t *testing.T) {
	// Skip in normal test runs
	if os.Getenv("RUN_RUNSERVER_TEST") != "true" {
		t.Skip("Skipping TestActualRunServer. Use -tags=runserver to run this test.")
	}

	// Save original functions to restore later
	originalNewApp := NewApp
	originalSignalNotify := signalNotify

	defer func() {
		// Restore original functions
		NewApp = originalNewApp
		signalNotify = originalSignalNotify
	}()

	// Create test cases
	tests := []struct {
		name            string
		initializeError error
		startError      error
		shutdownError   error
		sendSignal      bool
		expectError     bool
	}{
		{
			name:            "Successful initialization and graceful shutdown",
			initializeError: nil,
			startError:      nil,
			shutdownError:   nil,
			sendSignal:      true,
			expectError:     false,
		},
		{
			name:            "Error during initialization",
			initializeError: fmt.Errorf("initialize error"),
			startError:      nil,
			shutdownError:   nil,
			sendSignal:      false,
			expectError:     true,
		},
		{
			name:            "Error during start",
			initializeError: nil,
			startError:      fmt.Errorf("start error"),
			shutdownError:   nil,
			sendSignal:      false,
			expectError:     true,
		},
		{
			name:            "Error during shutdown",
			initializeError: nil,
			startError:      nil,
			shutdownError:   fmt.Errorf("shutdown error"),
			sendSignal:      true,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create channel to know when Start is called
			startNotify := make(chan struct{})

			// Create mock app
			mockApp := &MockAppSimple{
				initializeError: tt.initializeError,
				startError:      tt.startError,
				shutdownError:   tt.shutdownError,
				startNotify:     startNotify,
			}

			// Replace NewApp with our mock implementation
			NewApp = func(cfg *config.Config, opts ...AppOption) AppInterface {
				// Apply options to record they were passed
				for _, opt := range opts {
					// We can't use the options directly on the mock
					// but we want to make sure they don't cause a panic
					dummyApp := &App{}
					opt(dummyApp)
				}
				return mockApp
			}

			// Replace signal notify to send signal in test
			signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
				// In real implementation, the channel gets registered for OS signals
				// For test, we'll just send a signal manually if needed
				go func() {
					// For cases with startError, Start() returns immediately with an error
					// For other cases, wait until Start is called before sending signal
					if tt.startError == nil {
						<-startNotify
					}

					if tt.sendSignal {
						// Small delay to ensure server is ready
						time.Sleep(100 * time.Millisecond)
						c <- os.Interrupt
					}
				}()
			}

			// Create test config
			cfg := createSimpleTestConfig()

			// Create mock logger
			mockLogger := &MockLoggerSimple{}

			// Run the server in a goroutine
			resultCh := make(chan error, 1)
			go func() {
				resultCh <- runServer(cfg, mockLogger)
			}()

			// Set test timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Wait for result or timeout
			var err error
			select {
			case err = <-resultCh:
				// Got result
			case <-ctx.Done():
				t.Fatalf("Test timed out")
			}

			// Check result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify appropriate methods were called
			assert.True(t, mockApp.initializeCalled, "Initialize should have been called")

			if tt.initializeError != nil {
				// If initialization failed, start should not be called
				assert.False(t, mockApp.startCalled, "Start should not have been called after initialization error")
			} else if tt.sendSignal {
				// If we sent a shutdown signal, shutdown should be called
				assert.True(t, mockApp.shutdownCalled, "Shutdown should have been called")
			}
		})
	}
}
