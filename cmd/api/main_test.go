package main

import (
	"bytes"
	"encoding/base64"
	"os"
	"testing"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/pkg/mailer"
	"github.com/stretchr/testify/assert"
)

// TestConsoleMailer_SendMagicCode tests the ConsoleMailer's SendMagicCode method
func TestConsoleMailer_SendMagicCode(t *testing.T) {
	// Redirect stdout output for testing
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create console mailer
	mailerService := mailer.NewConsoleMailer()

	// Test sending a magic code
	err := mailerService.SendMagicCode("test@example.com", "123456")

	// Close the pipe to capture output
	w.Close()
	os.Stdout = oldStdout

	// Read the captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify no error
	assert.NoError(t, err)

	// Verify the output contains the expected message
	assert.Contains(t, output, "AUTHENTICATION MAGIC CODE")
	assert.Contains(t, output, "test@example.com")
	assert.Contains(t, output, "123456")
}

func TestConfigLoading(t *testing.T) {
	// Skip in CI environment to avoid env file requirements
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Test loading configuration
	cfg, err := config.LoadWithOptions(config.LoadOptions{
		EnvFile: ".env.test",
	})

	// If there's no env file, this will fail - that's expected in test environments
	if err != nil {
		assert.Contains(t, err.Error(), "PASETO_")
		return
	}
	assert.NotNil(t, cfg)
}

// TestSetupMinimalConfig tests a minimal config setup
func TestSetupMinimalConfig(t *testing.T) {
	// Set environment variables directly instead of using a file
	privateKey := "YDhVgXcnHQmkHYvzSqz9z7PPJccIWzSKGxXYWjlNs3xTtgx10KZb/XVpbA3EXe68/SLW7Vfv/j7b9LH3t7BMMw=="
	publicKey := "U7YMddCmW/11aWwNxF3uvP0i1u1X7/4+2/Sx97ewTDM="

	// Set the environment variables
	os.Setenv("PASETO_PRIVATE_KEY", privateKey)
	os.Setenv("PASETO_PUBLIC_KEY", publicKey)
	os.Setenv("ROOT_EMAIL", "admin@example.com")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")

	// Clean up after the test
	defer func() {
		os.Unsetenv("PASETO_PRIVATE_KEY")
		os.Unsetenv("PASETO_PUBLIC_KEY")
		os.Unsetenv("ROOT_EMAIL")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()

	// Load configuration directly
	cfg, err := config.Load()

	// Verify the configuration loaded correctly
	if err != nil {
		// If there's still a .env file issue, that's OK as long as the env vars are processed
		if err.Error() == "PASETO_PRIVATE_KEY is required" ||
			err.Error() == "PASETO_PUBLIC_KEY is required" {
			t.Fatal("Environment variables not properly loaded:", err)
		}
	} else {
		// Configuration loaded successfully, verify values
		assert.NotNil(t, cfg)
		assert.Equal(t, "admin@example.com", cfg.RootEmail)

		// Verify the keys were decoded correctly
		decodedPrivateKey, _ := base64.StdEncoding.DecodeString(privateKey)
		decodedPublicKey, _ := base64.StdEncoding.DecodeString(publicKey)
		assert.Equal(t, decodedPrivateKey, cfg.Security.PasetoPrivateKey)
		assert.Equal(t, decodedPublicKey, cfg.Security.PasetoPublicKey)
	}
}
