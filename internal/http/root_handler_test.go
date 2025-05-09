package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootHandler(t *testing.T) {
	handler := NewRootHandler()
	assert.NotNil(t, handler, "Root handler should not be nil")
}

func TestRootHandler_Handle(t *testing.T) {
	handler := NewRootHandler()

	// Create a test request
	req := httptest.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()

	// Call the handler
	handler.Handle(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Decode response body
	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Assert response content
	assert.Equal(t, "api running", response["status"])
}

func TestRootHandler_RegisterRoutes(t *testing.T) {
	handler := NewRootHandler()
	mux := http.NewServeMux()

	// Register routes
	handler.RegisterRoutes(mux)

	// Create a test server
	server := httptest.NewServer(mux)
	defer server.Close()

	// Send a request
	resp, err := http.Get(server.URL + "/api")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Decode response body
	var response map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	// Assert response content
	assert.Equal(t, "api running", response["status"])
}

func TestRootHandler_ServeConfigJS(t *testing.T) {
	// Create a test logger
	testLogger := logger.NewLogger()

	// Create a handler with a test API endpoint
	testAPIEndpoint := "https://api.example.com"
	handler := NewRootHandlerWithConsole("test_console_dir", testLogger, testAPIEndpoint)

	// Create a request to /config.js
	req := httptest.NewRequest("GET", "/config.js", nil)
	rr := httptest.NewRecorder()

	// Call the handler directly
	handler.serveConfigJS(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the content type
	assert.Equal(t, "application/javascript", rr.Header().Get("Content-Type"))

	// Check cache control headers
	assert.Equal(t, "no-cache, no-store, must-revalidate", rr.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", rr.Header().Get("Pragma"))
	assert.Equal(t, "0", rr.Header().Get("Expires"))

	// Check the body contains the expected JavaScript
	expectedJS := `window.API_ENDPOINT = "https://api.example.com";`
	assert.Equal(t, expectedJS, rr.Body.String())
}

func TestRootHandler_Handle_ConfigJS(t *testing.T) {
	// Create a test logger
	testLogger := logger.NewLogger()

	// Create a handler with a test API endpoint
	testAPIEndpoint := "https://api.example.com"
	handler := NewRootHandlerWithConsole("test_console_dir", testLogger, testAPIEndpoint)

	// Create a request to /config.js
	req := httptest.NewRequest("GET", "/config.js", nil)
	rr := httptest.NewRecorder()

	// Call the general handle method
	handler.Handle(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the content type
	assert.Equal(t, "application/javascript", rr.Header().Get("Content-Type"))

	// Check the body contains the expected JavaScript
	expectedJS := `window.API_ENDPOINT = "https://api.example.com";`
	assert.Equal(t, expectedJS, rr.Body.String())
}
