package keygenfunction

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

	// Verify keys are not empty
	if keyPair.PrivateKey == "" {
		t.Error("Private key should not be empty")
	}
	if keyPair.PublicKey == "" {
		t.Error("Public key should not be empty")
	}

	// Verify keys are base64 encoded (basic check)
	if !strings.Contains(keyPair.PrivateKey, "=") && len(keyPair.PrivateKey) < 10 {
		t.Error("Private key doesn't look like base64")
	}
	if !strings.Contains(keyPair.PublicKey, "=") && len(keyPair.PublicKey) < 10 {
		t.Error("Public key doesn't look like base64")
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
