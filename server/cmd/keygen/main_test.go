package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
)

func TestKeyGeneration(t *testing.T) {
	// Capture stdout to test the output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the main function
	main()

	// Restore stdout and get the output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Test that the output contains our expected text portions
	if !strings.Contains(output, "Generated PASETO v4 key pair") {
		t.Error("Output doesn't contain expected header")
	}

	if !strings.Contains(output, "Private Key (keep this secret!)") {
		t.Error("Output doesn't mention private key")
	}

	if !strings.Contains(output, "Public Key") {
		t.Error("Output doesn't mention public key")
	}

	// Extract keys from output
	lines := strings.Split(output, "\n")
	var privateKeyBase64, publicKeyBase64 string

	for i, line := range lines {
		if strings.Contains(line, "Private Key") && i+1 < len(lines) {
			privateKeyBase64 = lines[i+1]
		}
		if strings.Contains(line, "Public Key") && i+1 < len(lines) {
			publicKeyBase64 = lines[i+1]
		}
	}

	// Test that keys are valid base64 and can be decoded
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		t.Errorf("Failed to decode private key: %v", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		t.Errorf("Failed to decode public key: %v", err)
	}

	// Test that we can recreate PASETO keys from the bytes
	_, err = paseto.NewV4AsymmetricSecretKeyFromBytes(privateKeyBytes)
	if err != nil {
		t.Errorf("Failed to create secret key from bytes: %v", err)
	}

	_, err = paseto.NewV4AsymmetricPublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		t.Errorf("Failed to create public key from bytes: %v", err)
	}
}

func TestSignAndVerify(t *testing.T) {
	// Generate a new key pair
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Create a token and sign it
	token := paseto.NewToken()

	// Use time.Now() for setting token times
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(time.Hour))
	token.Set("data", "test value")

	signed := token.V4Sign(secretKey, nil)

	// Verify the token
	parser := paseto.NewParser()
	parsedToken, err := parser.ParseV4Public(publicKey, signed, nil)
	if err != nil {
		t.Errorf("Failed to verify token: %v", err)
	}

	// Check that the data is preserved
	value, err := parsedToken.GetString("data")
	if err != nil {
		t.Errorf("Failed to get data from token: %v", err)
	}

	if value != "test value" {
		t.Errorf("Expected value 'test value', got '%s'", value)
	}
}

// TestKeyGenerationMock tests the key generation without calling main(),
// which allows us to test the core functionality without capturing stdout.
func TestKeyGenerationMock(t *testing.T) {
	// Generate a new key pair
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Convert keys to base64 for storage
	privateKeyBase64 := base64.StdEncoding.EncodeToString(secretKey.ExportBytes())
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey.ExportBytes())

	// Validate the encoded keys can be decoded back
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		t.Errorf("Failed to decode private key: %v", err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		t.Errorf("Failed to decode public key: %v", err)
	}

	// Test that the decoded keys can be used to recreate paseto keys
	recoveredSecretKey, err := paseto.NewV4AsymmetricSecretKeyFromBytes(privateKeyBytes)
	if err != nil {
		t.Errorf("Failed to create secret key from bytes: %v", err)
	}

	recoveredPublicKey, err := paseto.NewV4AsymmetricPublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		t.Errorf("Failed to create public key from bytes: %v", err)
	}

	// Verify that the original and recovered keys work together
	// Create a test token
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(time.Hour))
	token.Set("test", "value")

	// Sign with original key
	signed := token.V4Sign(secretKey, nil)

	// Verify with recovered public key
	parser := paseto.NewParser()
	_, err = parser.ParseV4Public(recoveredPublicKey, signed, nil)
	if err != nil {
		t.Errorf("Failed to verify token with recovered public key: %v", err)
	}

	// Now sign with recovered key
	token2 := paseto.NewToken()
	token2.SetIssuedAt(time.Now())
	token2.SetNotBefore(time.Now())
	token2.SetExpiration(time.Now().Add(time.Hour))
	token2.Set("test", "value2")

	signed2 := token2.V4Sign(recoveredSecretKey, nil)

	// Verify with original public key
	_, err = parser.ParseV4Public(publicKey, signed2, nil)
	if err != nil {
		t.Errorf("Failed to verify token signed with recovered key: %v", err)
	}
}
