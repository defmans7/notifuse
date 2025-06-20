package testkeys

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTestKeys(t *testing.T) {
	privateKeyB64, publicKeyB64 := GetTestKeys()

	assert.NotEmpty(t, privateKeyB64, "Private key should not be empty")
	assert.NotEmpty(t, publicKeyB64, "Public key should not be empty")

	// Verify keys are valid base64
	_, err := base64.StdEncoding.DecodeString(privateKeyB64)
	assert.NoError(t, err, "Private key should be valid base64")

	_, err = base64.StdEncoding.DecodeString(publicKeyB64)
	assert.NoError(t, err, "Public key should be valid base64")
}

func TestGetTestKeysBytes(t *testing.T) {
	privateKeyBytes, publicKeyBytes, err := GetTestKeysBytes()
	require.NoError(t, err, "GetTestKeysBytes should not return error")

	assert.NotEmpty(t, privateKeyBytes, "Private key bytes should not be empty")
	assert.NotEmpty(t, publicKeyBytes, "Public key bytes should not be empty")

	// Verify expected key lengths for ed25519
	assert.Equal(t, 64, len(privateKeyBytes), "Private key should be 64 bytes (ed25519)")
	assert.Equal(t, 32, len(publicKeyBytes), "Public key should be 32 bytes (ed25519)")
}

func TestGetHardcodedTestKeys(t *testing.T) {
	keys, err := GetHardcodedTestKeys()
	require.NoError(t, err, "GetHardcodedTestKeys should not return error")

	assert.NotNil(t, keys, "Keys should not be nil")
	assert.NotEmpty(t, keys.PrivateKeyB64, "Private key B64 should not be empty")
	assert.NotEmpty(t, keys.PublicKeyB64, "Public key B64 should not be empty")
	assert.NotEmpty(t, keys.PrivateKeyBytes, "Private key bytes should not be empty")
	assert.NotEmpty(t, keys.PublicKeyBytes, "Public key bytes should not be empty")

	// Verify consistency between base64 and bytes
	expectedPrivateBytes, _ := base64.StdEncoding.DecodeString(keys.PrivateKeyB64)
	expectedPublicBytes, _ := base64.StdEncoding.DecodeString(keys.PublicKeyB64)

	assert.Equal(t, expectedPrivateBytes, keys.PrivateKeyBytes, "Private key bytes should match decoded base64")
	assert.Equal(t, expectedPublicBytes, keys.PublicKeyBytes, "Public key bytes should match decoded base64")
}

func TestGenerateValidPasetoKeys(t *testing.T) {
	keys, err := GenerateValidPasetoKeys()
	require.NoError(t, err, "GenerateValidPasetoKeys should not return error")

	assert.NotNil(t, keys, "Generated keys should not be nil")
	assert.NotEmpty(t, keys.PrivateKeyB64, "Generated private key B64 should not be empty")
	assert.NotEmpty(t, keys.PublicKeyB64, "Generated public key B64 should not be empty")
	assert.NotEmpty(t, keys.PrivateKeyBytes, "Generated private key bytes should not be empty")
	assert.NotEmpty(t, keys.PublicKeyBytes, "Generated public key bytes should not be empty")

	// Verify key lengths
	assert.Equal(t, 64, len(keys.PrivateKeyBytes), "Generated private key should be 64 bytes (ed25519)")
	assert.Equal(t, 32, len(keys.PublicKeyBytes), "Generated public key should be 32 bytes (ed25519)")

	// Verify that generated keys are different from hardcoded ones
	hardcodedKeys, _ := GetHardcodedTestKeys()
	assert.NotEqual(t, hardcodedKeys.PrivateKeyB64, keys.PrivateKeyB64, "Generated keys should be different from hardcoded")
	assert.NotEqual(t, hardcodedKeys.PublicKeyB64, keys.PublicKeyB64, "Generated keys should be different from hardcoded")
}
