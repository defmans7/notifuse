package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransport is a custom http.RoundTripper that returns predefined responses
type mockTransport struct {
	responses map[string]*http.Response
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, ok := t.responses[req.URL.String()]
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("Not found")),
			Header:     make(http.Header),
		}, nil
	}
	return resp, nil
}

func TestNewFaviconHandler(t *testing.T) {
	handler := NewFaviconHandler()
	assert.NotNil(t, handler)
}

func TestFaviconHandler_DetectFavicon_MethodNotAllowed(t *testing.T) {
	handler := NewFaviconHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/detect-favicon", nil)
	w := httptest.NewRecorder()

	handler.DetectFavicon(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestFaviconHandler_DetectFavicon_InvalidJSON(t *testing.T) {
	handler := NewFaviconHandler()

	// Create a request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/detect-favicon", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	handler.DetectFavicon(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestFaviconHandler_DetectFavicon_EmptyURL(t *testing.T) {
	handler := NewFaviconHandler()

	// Create a request with empty URL
	reqBody, _ := json.Marshal(FaviconRequest{URL: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/detect-favicon", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.DetectFavicon(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "URL is required")
}

func TestFaviconHandler_DetectFavicon_InvalidURL(t *testing.T) {
	handler := NewFaviconHandler()

	// Create a request with invalid URL
	reqBody, _ := json.Marshal(FaviconRequest{URL: "://invalid-url"})
	req := httptest.NewRequest(http.MethodPost, "/api/detect-favicon", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler.DetectFavicon(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid URL")
}

func TestResolveURL(t *testing.T) {
	testCases := []struct {
		name     string
		baseURL  string
		href     string
		expected string
		hasError bool
	}{
		{
			name:     "absolute URL",
			baseURL:  "https://example.com",
			href:     "https://example.com/favicon.ico",
			expected: "https://example.com/favicon.ico",
			hasError: false,
		},
		{
			name:     "relative URL",
			baseURL:  "https://example.com",
			href:     "/favicon.ico",
			expected: "https://example.com/favicon.ico",
			hasError: false,
		},
		{
			name:     "invalid base URL",
			baseURL:  "://invalid",
			href:     "/favicon.ico",
			expected: "",
			hasError: true,
		},
		{
			name:     "invalid href",
			baseURL:  "https://example.com",
			href:     "://invalid", // This isn't actually invalid for URL resolution as it's treated as a relative path
			expected: "https://example.com/://invalid",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var baseURL *url.URL
			var err error

			if tc.hasError && tc.name == "invalid base URL" {
				baseURL, err = url.Parse(tc.baseURL)
				assert.Error(t, err)
				return
			} else {
				baseURL, err = url.Parse(tc.baseURL)
				require.NoError(t, err)
			}

			result, err := resolveURL(baseURL, tc.href)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestTryDefaultFavicon(t *testing.T) {
	// Create a client with a mock transport
	originalClient := http.DefaultClient
	defer func() { http.DefaultClient = originalClient }()

	mockClient := &http.Client{
		Transport: &mockTransport{
			responses: map[string]*http.Response{
				"https://example.com/favicon.ico": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("mock icon")),
				},
				"https://noicon.com/favicon.ico": {
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("not found")),
				},
			},
		},
	}
	http.DefaultClient = mockClient

	// Test successful icon detection
	baseURL, err := url.Parse("https://example.com")
	require.NoError(t, err)
	result := tryDefaultFavicon(baseURL)
	assert.Equal(t, "https://example.com/favicon.ico", result)

	// Test failed icon detection
	baseURL, err = url.Parse("https://noicon.com")
	require.NoError(t, err)
	result = tryDefaultFavicon(baseURL)
	assert.Equal(t, "", result)
}

// Helper function to create a mock HTML document for testing
func createMockHTMLDoc(t *testing.T, htmlContent string) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	require.NoError(t, err)
	return doc
}

func TestFindAppleTouchIcon(t *testing.T) {
	baseURL, err := url.Parse("https://example.com")
	require.NoError(t, err)

	t.Run("with apple-touch-icon", func(t *testing.T) {
		html := `<html><head><link rel="apple-touch-icon" href="/apple-touch-icon.png"></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findAppleTouchIcon(doc, baseURL)
		assert.Equal(t, "https://example.com/apple-touch-icon.png", result)
	})

	t.Run("without apple-touch-icon", func(t *testing.T) {
		html := `<html><head></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findAppleTouchIcon(doc, baseURL)
		assert.Equal(t, "", result)
	})
}

func TestFindTraditionalFavicon(t *testing.T) {
	baseURL, err := url.Parse("https://example.com")
	require.NoError(t, err)

	t.Run("with favicon link", func(t *testing.T) {
		html := `<html><head><link rel="shortcut icon" href="/favicon.ico"></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findTraditionalFavicon(doc, baseURL)
		assert.Equal(t, "https://example.com/favicon.ico", result)
	})

	t.Run("with icon link", func(t *testing.T) {
		html := `<html><head><link rel="icon" href="/icon.png"></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findTraditionalFavicon(doc, baseURL)
		assert.Equal(t, "https://example.com/icon.png", result)
	})

	t.Run("without favicon", func(t *testing.T) {
		html := `<html><head></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findTraditionalFavicon(doc, baseURL)
		assert.Equal(t, "", result)
	})
}

func TestFindManifestIcon(t *testing.T) {
	baseURL, err := url.Parse("https://example.com")
	require.NoError(t, err)

	// Save and restore the default HTTP client
	originalClient := http.DefaultClient
	defer func() { http.DefaultClient = originalClient }()

	// Create a mock HTTP client
	mockClient := &http.Client{
		Transport: &mockTransport{
			responses: map[string]*http.Response{
				"https://example.com/manifest.json": {
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"icons": [
							{
								"src": "/icon-192.png",
								"sizes": "192x192"
							},
							{
								"src": "/icon-512.png",
								"sizes": "512x512"
							}
						]
					}`)),
					Header: make(http.Header),
				},
				"https://example.com/empty-manifest.json": {
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"icons": []
					}`)),
					Header: make(http.Header),
				},
				"https://example.com/invalid-manifest.json": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`invalid json`)),
					Header:     make(http.Header),
				},
				"https://failure.com/manifest.json": {
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader(`not found`)),
					Header:     make(http.Header),
				},
			},
		},
	}
	http.DefaultClient = mockClient

	t.Run("with valid manifest", func(t *testing.T) {
		html := `<html><head><link rel="manifest" href="/manifest.json"></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findManifestIcon(doc, baseURL)
		assert.Equal(t, "https://example.com/icon-512.png", result)
	})

	t.Run("with empty manifest icons", func(t *testing.T) {
		html := `<html><head><link rel="manifest" href="/empty-manifest.json"></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findManifestIcon(doc, baseURL)
		assert.Equal(t, "", result)
	})

	t.Run("with invalid manifest JSON", func(t *testing.T) {
		html := `<html><head><link rel="manifest" href="/invalid-manifest.json"></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findManifestIcon(doc, baseURL)
		assert.Equal(t, "", result)
	})

	t.Run("with manifest fetch error", func(t *testing.T) {
		html := `<html><head><link rel="manifest" href="https://failure.com/manifest.json"></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findManifestIcon(doc, baseURL)
		assert.Equal(t, "", result)
	})

	t.Run("without manifest link", func(t *testing.T) {
		html := `<html><head></head><body></body></html>`
		doc := createMockHTMLDoc(t, html)

		result := findManifestIcon(doc, baseURL)
		assert.Equal(t, "", result)
	})
}
