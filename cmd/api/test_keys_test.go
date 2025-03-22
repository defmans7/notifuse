package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHardcodedKeys(t *testing.T) {
	// Test hardcoded keys
	keys, err := GetHardcodedTestKeys()
	require.NoError(t, err)

	// Verify keys have expected values
	assert.Equal(t, HardcodedPrivateKeyB64, keys.PrivateKeyB64)
	assert.Equal(t, HardcodedPublicKeyB64, keys.PublicKeyB64)
	assert.NotEmpty(t, keys.PrivateKeyBytes)
	assert.NotEmpty(t, keys.PublicKeyBytes)
}

func TestGenerateValidPasetoKeys(t *testing.T) {
	// Test generating new keys
	keys, err := GenerateValidPasetoKeys()
	require.NoError(t, err)

	// Verify keys are properly generated
	assert.NotEmpty(t, keys.PrivateKeyB64)
	assert.NotEmpty(t, keys.PublicKeyB64)
	assert.NotEmpty(t, keys.PrivateKeyBytes)
	assert.NotEmpty(t, keys.PublicKeyBytes)

	// Verify private and public key lengths
	assert.Equal(t, 64, len(keys.PrivateKeyBytes), "Private key should be 64 bytes (ed25519)")
	assert.Equal(t, 32, len(keys.PublicKeyBytes), "Public key should be 32 bytes (ed25519)")
}

func TestSaveAndLoadTestKeys(t *testing.T) {
	// Generate test keys
	keys, err := GenerateValidPasetoKeys()
	require.NoError(t, err)

	// Create a temporary file for testing
	tempDir, err := os.MkdirTemp("", "test-keys")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test_keys.env")

	// Save keys to file
	err = SaveTestKeysToFile(keys, tempFile)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tempFile)
	assert.NoError(t, err)

	// Load keys from file
	loadedKeys, err := LoadTestKeysFromFile(tempFile)
	require.NoError(t, err)

	// Verify loaded keys match original keys
	assert.Equal(t, keys.PrivateKeyB64, loadedKeys.PrivateKeyB64)
	assert.Equal(t, keys.PublicKeyB64, loadedKeys.PublicKeyB64)
}

func TestGetTestKeysFilePath(t *testing.T) {
	// Test getting test keys file path
	path := GetTestKeysFilePath()

	// Verify path is not empty and includes temp directory
	assert.NotEmpty(t, path)
	assert.Contains(t, path, os.TempDir())
	assert.Contains(t, path, "notifuse_test_keys.env")
}

func TestSetupTestKeys(t *testing.T) {
	// Test key setup (should use hardcoded keys first)
	keys, err := SetupTestKeys()
	require.NoError(t, err)

	// Verify keys match hardcoded keys
	hardcodedKeys, _ := GetHardcodedTestKeys()
	assert.Equal(t, hardcodedKeys.PrivateKeyB64, keys.PrivateKeyB64)
	assert.Equal(t, hardcodedKeys.PublicKeyB64, keys.PublicKeyB64)
}

func TestSetupTestConfig(t *testing.T) {
	// Test config setup
	cfg, err := SetupTestConfig()
	require.NoError(t, err)

	// Verify config has expected values
	assert.Equal(t, "test", cfg.Environment)
	assert.NotEmpty(t, cfg.Security.PasetoPrivateKeyBytes)
	assert.NotEmpty(t, cfg.Security.PasetoPublicKeyBytes)

	// Verify keys match hardcoded keys
	hardcodedKeys, _ := GetHardcodedTestKeys()
	assert.Equal(t, hardcodedKeys.PrivateKeyBytes, cfg.Security.PasetoPrivateKeyBytes)
	assert.Equal(t, hardcodedKeys.PublicKeyBytes, cfg.Security.PasetoPublicKeyBytes)
}
