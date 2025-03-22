package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Notifuse/notifuse/config"
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

// GetHardcodedTestKeys returns the hardcoded test keys
func GetHardcodedTestKeys() (*TestKeys, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(HardcodedPrivateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(HardcodedPublicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	return &TestKeys{
		PrivateKeyBytes: privateKeyBytes,
		PublicKeyBytes:  publicKeyBytes,
		PrivateKeyB64:   HardcodedPrivateKeyB64,
		PublicKeyB64:    HardcodedPublicKeyB64,
	}, nil
}

// GenerateValidPasetoKeys generates a new set of ed25519 keys for testing
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

// SaveTestKeysToFile saves the generated keys to a file
func SaveTestKeysToFile(keys *TestKeys, filename string) error {
	content := fmt.Sprintf("PASETO_PRIVATE_KEY=%s\nPASETO_PUBLIC_KEY=%s\n",
		keys.PrivateKeyB64, keys.PublicKeyB64)

	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = ioutil.WriteFile(filename, []byte(content), 0600)
	if err != nil {
		return fmt.Errorf("failed to write keys file: %w", err)
	}

	return nil
}

// LoadTestKeysFromFile loads PASETO keys from a file
func LoadTestKeysFromFile(filename string) (*TestKeys, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read keys file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var privateKeyB64, publicKeyB64 string

	for _, line := range lines {
		if strings.HasPrefix(line, "PASETO_PRIVATE_KEY=") {
			privateKeyB64 = strings.TrimPrefix(line, "PASETO_PRIVATE_KEY=")
		} else if strings.HasPrefix(line, "PASETO_PUBLIC_KEY=") {
			publicKeyB64 = strings.TrimPrefix(line, "PASETO_PUBLIC_KEY=")
		}
	}

	if privateKeyB64 == "" || publicKeyB64 == "" {
		return nil, fmt.Errorf("keys not found in file")
	}

	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	return &TestKeys{
		PrivateKeyBytes: privateKeyBytes,
		PublicKeyBytes:  publicKeyBytes,
		PrivateKeyB64:   privateKeyB64,
		PublicKeyB64:    publicKeyB64,
	}, nil
}

// GetTestKeysFilePath returns the path to the temp file for storing test keys
func GetTestKeysFilePath() string {
	return filepath.Join(os.TempDir(), "notifuse_test_keys.env")
}

// SetupTestKeys sets up keys for testing, first checking for hardcoded keys,
// then a cached file, and finally generating new ones if needed
func SetupTestKeys() (*TestKeys, error) {
	// First try to get hardcoded keys
	keys, err := GetHardcodedTestKeys()
	if err == nil {
		return keys, nil
	}

	// Then try to load from temp file
	tempFile := GetTestKeysFilePath()
	keys, err = LoadTestKeysFromFile(tempFile)
	if err == nil {
		return keys, nil
	}

	// Finally, generate new keys
	keys, err = GenerateValidPasetoKeys()
	if err != nil {
		return nil, err
	}

	// Save for future use
	_ = SaveTestKeysToFile(keys, tempFile)

	return keys, nil
}

// SetupTestConfig sets up a test configuration with valid PASETO keys
func SetupTestConfig() (*config.Config, error) {
	keys, err := SetupTestKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to set up test keys: %w", err)
	}

	config := &config.Config{
		Security: config.SecurityConfig{
			PasetoPrivateKeyBytes: keys.PrivateKeyBytes,
			PasetoPublicKeyBytes:  keys.PublicKeyBytes,
		},
		Environment: "test",
	}

	return config, nil
}
