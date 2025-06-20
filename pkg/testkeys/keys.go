// Package testkeys provides test keys for PASETO authentication in testing environments.
// These keys should NEVER be used in production.
package testkeys

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Hardcoded keys for testing - DO NOT USE IN PRODUCTION
const (
	HardcodedPrivateKeyB64 = "UayDa4OMDpm3CvIT+iSC39iDyPlsui0pNQYDEZ1pbo1LsIrO4p/aVuCBWz6LiYvzj9pc+gn0gLwRd0CoHV+nxw=="
	HardcodedPublicKeyB64  = "S7CKzuKf2lbggVs+i4mL84/aXPoJ9IC8EXdAqB1fp8c="
)

// TestKeys holds the generated keys for testing
type TestKeys struct {
	PrivateKeyBytes []byte
	PublicKeyBytes  []byte
	PrivateKeyB64   string
	PublicKeyB64    string
}

// GetTestKeys returns hardcoded PASETO keys for testing purposes (base64 encoded).
// These keys should NEVER be used in production environments.
func GetTestKeys() (string, string) {
	return HardcodedPrivateKeyB64, HardcodedPublicKeyB64
}

// GetTestKeysBytes returns hardcoded PASETO keys as byte slices for testing.
// These keys should NEVER be used in production environments.
func GetTestKeysBytes() ([]byte, []byte, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(HardcodedPrivateKeyB64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(HardcodedPublicKeyB64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	return privateKeyBytes, publicKeyBytes, nil
}

// GetHardcodedTestKeys returns the hardcoded test keys as a TestKeys struct.
// These keys should NEVER be used in production.
func GetHardcodedTestKeys() (*TestKeys, error) {
	privateKeyBytes, publicKeyBytes, err := GetTestKeysBytes()
	if err != nil {
		return nil, err
	}

	return &TestKeys{
		PrivateKeyBytes: privateKeyBytes,
		PublicKeyBytes:  publicKeyBytes,
		PrivateKeyB64:   HardcodedPrivateKeyB64,
		PublicKeyB64:    HardcodedPublicKeyB64,
	}, nil
}

// GenerateValidPasetoKeys generates a new set of ed25519 keys for testing.
// This should be used when you need fresh keys for each test.
func GenerateValidPasetoKeys() (*TestKeys, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key pair: %w", err)
	}

	privKeyB64 := base64.StdEncoding.EncodeToString(privKey)
	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKey)

	return &TestKeys{
		PrivateKeyBytes: privKey,
		PublicKeyBytes:  pubKey,
		PrivateKeyB64:   privKeyB64,
		PublicKeyB64:    pubKeyB64,
	}, nil
}

// GenerateRandomKeyBytes generates random bytes for testing keys
func GenerateRandomKeyBytes(length int) []byte {
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
