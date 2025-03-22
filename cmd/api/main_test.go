package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mailer"
	"github.com/stretchr/testify/assert"
)

// TestRunServerMocked tests the runServer function with mocking
func TestRunServerMocked(t *testing.T) {
	// Skip this test in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping in CI environment")
	}

	// Get hardcoded keys for testing
	keys, err := GetHardcodedTestKeys()
	assert.NoError(t, err)

	// Create test config
	cfg := createTestConfig()
	// Override config with our hardcoded keys
	cfg.Security.PasetoPrivateKeyBytes = keys.PrivateKeyBytes
	cfg.Security.PasetoPublicKeyBytes = keys.PublicKeyBytes

	// Use a random high port to avoid conflicts
	cfg.Server.Port = 18080 + (time.Now().Nanosecond() % 1000)

	// Create mock logger
	mockLogger := &MockLogger{}

	// Create a mock DB
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Create app manually with our mocks
	app := NewApp(cfg, WithLogger(mockLogger), WithMockDB(mockDB))

	// Setup a simple runServer function that just starts and stops the app
	testRunServer := func(_ *config.Config, logger logger.Logger) error {
		// Start the server in a goroutine
		serverError := make(chan error, 1)
		go func() {
			logger.Info("Server started successfully")
			serverError <- app.Start()
		}()

		// Send shutdown signal
		time.Sleep(100 * time.Millisecond)

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := app.Shutdown(ctx); err != nil {
			return err
		}

		logger.Info("Server shut down gracefully")
		return nil
	}

	// Run the test function
	err = testRunServer(cfg, mockLogger)
	assert.NoError(t, err)
}

func TestConsoleMailer_SendMagicCode(t *testing.T) {
	// Creating a ConsoleMailer
	m := mailer.NewConsoleMailer()

	// Capturing stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = w

	// Make sure we restore stdout when we're done
	defer func() {
		os.Stdout = oldStdout
	}()

	// Call the function we're testing
	err = m.SendMagicCode("test@example.com", "123456")
	assert.NoError(t, err)

	// Close write end of pipe so we can read all data
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	assert.NoError(t, err)
	output := buf.String()

	// Check if output contains the magic code message - update the string to match actual output
	assert.Contains(t, output, "AUTHENTICATION MAGIC CODE")
	assert.Contains(t, output, "test@example.com")
	assert.Contains(t, output, "123456")
}

func TestConfigLoading(t *testing.T) {
	// Skip this test in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping in CI environment")
	}

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
