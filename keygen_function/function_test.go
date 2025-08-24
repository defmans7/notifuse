package keygenfunction

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestKeygenHandler(t *testing.T) {
	// Test GET request (HTML page)
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(KeygenHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/html")
	}

	// Test POST request (key generation)
	req, err = http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	// Verify JSON response structure
	var keyPair KeyPair
	if err := json.NewDecoder(rr.Body).Decode(&keyPair); err != nil {
		t.Errorf("Failed to decode JSON response: %v", err)
	}

	// Verify metadata
	if keyPair.Algorithm != "PASETO v4" {
		t.Errorf("Expected algorithm 'PASETO v4', got %s", keyPair.Algorithm)
	}
	if keyPair.KeyType != "asymmetric" {
		t.Errorf("Expected keyType 'asymmetric', got %s", keyPair.KeyType)
	}
	if time.Since(keyPair.GeneratedAt) > time.Minute {
		t.Error("GeneratedAt timestamp should be recent")
	}

	// Verify private key data
	if keyPair.PrivateKey.Base64 == "" {
		t.Error("Private key base64 should not be empty")
	}
	if keyPair.PrivateKey.Hex == "" {
		t.Error("Private key hex should not be empty")
	}
	if keyPair.PrivateKey.ByteLength == 0 {
		t.Error("Private key byte length should not be zero")
	}

	// Verify public key data
	if keyPair.PublicKey.Base64 == "" {
		t.Error("Public key base64 should not be empty")
	}
	if keyPair.PublicKey.Hex == "" {
		t.Error("Public key hex should not be empty")
	}
	if keyPair.PublicKey.ByteLength == 0 {
		t.Error("Public key byte length should not be zero")
	}

	// Verify base64 encoding validity
	_, err = base64.StdEncoding.DecodeString(keyPair.PrivateKey.Base64)
	if err != nil {
		t.Errorf("Private key base64 is invalid: %v", err)
	}
	_, err = base64.StdEncoding.DecodeString(keyPair.PublicKey.Base64)
	if err != nil {
		t.Errorf("Public key base64 is invalid: %v", err)
	}

	// Verify hex encoding validity
	_, err = hex.DecodeString(keyPair.PrivateKey.Hex)
	if err != nil {
		t.Errorf("Private key hex is invalid: %v", err)
	}
	_, err = hex.DecodeString(keyPair.PublicKey.Hex)
	if err != nil {
		t.Errorf("Public key hex is invalid: %v", err)
	}

	// Verify that base64 and hex represent the same data
	privateBase64Bytes, _ := base64.StdEncoding.DecodeString(keyPair.PrivateKey.Base64)
	privateHexBytes, _ := hex.DecodeString(keyPair.PrivateKey.Hex)
	if string(privateBase64Bytes) != string(privateHexBytes) {
		t.Error("Private key base64 and hex should represent the same data")
	}

	publicBase64Bytes, _ := base64.StdEncoding.DecodeString(keyPair.PublicKey.Base64)
	publicHexBytes, _ := hex.DecodeString(keyPair.PublicKey.Hex)
	if string(publicBase64Bytes) != string(publicHexBytes) {
		t.Error("Public key base64 and hex should represent the same data")
	}

	// Verify byte lengths match actual decoded data
	if len(privateBase64Bytes) != keyPair.PrivateKey.ByteLength {
		t.Errorf("Private key byte length mismatch: expected %d, got %d", len(privateBase64Bytes), keyPair.PrivateKey.ByteLength)
	}
	if len(publicBase64Bytes) != keyPair.PublicKey.ByteLength {
		t.Errorf("Public key byte length mismatch: expected %d, got %d", len(publicBase64Bytes), keyPair.PublicKey.ByteLength)
	}
}

func TestKeygenHandlerMethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("PUT", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(KeygenHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
