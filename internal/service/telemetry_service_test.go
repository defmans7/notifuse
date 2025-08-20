package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelemetryService_SendMetricsForAllWorkspaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repositories
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)

	// Create a test HTTP server
	var receivedRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedRequests++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Temporarily override the TelemetryEndpoint constant for testing
	originalEndpoint := TelemetryEndpoint
	defer func() {
		// We can't actually change a const, but we can work around it
		// by creating a custom HTTP client that redirects to our test server
	}()

	// Create custom HTTP client that redirects to test server
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &testTransport{
			testServerURL: server.URL,
			originalURL:   originalEndpoint,
		},
	}

	// Create telemetry service
	config := TelemetryServiceConfig{
		Enabled:       true,
		APIEndpoint:   "https://api.example.com",
		WorkspaceRepo: mockWorkspaceRepo,
		Logger:        logger.NewLoggerWithLevel("debug"),
		HTTPClient:    httpClient,
	}

	service := NewTelemetryService(config)

	// Mock workspace list
	workspaces := []*domain.Workspace{
		{ID: "workspace1", Name: "Test Workspace 1"},
		{ID: "workspace2", Name: "Test Workspace 2"},
	}

	mockWorkspaceRepo.EXPECT().List(gomock.Any()).Return(workspaces, nil)

	// Mock database connections returning errors to test graceful handling
	mockWorkspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace1").Return(nil, assert.AnError)
	mockWorkspaceRepo.EXPECT().GetConnection(gomock.Any(), "workspace2").Return(nil, assert.AnError)

	// Execute
	ctx := context.Background()
	err := service.SendMetricsForAllWorkspaces(ctx)

	// Verify - should succeed even with database errors
	require.NoError(t, err)
	assert.Equal(t, 2, receivedRequests, "Should have sent metrics for 2 workspaces")
}

// testTransport is a custom HTTP transport for testing that redirects requests
type testTransport struct {
	testServerURL string
	originalURL   string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() == t.originalURL {
		// Redirect to test server
		req.URL, _ = req.URL.Parse(t.testServerURL)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func TestTelemetryService_DisabledService(t *testing.T) {
	// Create telemetry service with disabled configuration
	config := TelemetryServiceConfig{
		Enabled:     false,
		APIEndpoint: "https://api.example.com",
		Logger:      logger.NewLoggerWithLevel("debug"),
	}

	service := NewTelemetryService(config)

	// Execute
	ctx := context.Background()
	err := service.SendMetricsForAllWorkspaces(ctx)

	// Verify - should return without error and without making any calls
	require.NoError(t, err)
}

func TestTelemetryService_StartDailyScheduler(t *testing.T) {
	config := TelemetryServiceConfig{
		Enabled:     true,
		APIEndpoint: "https://api.example.com",
		Logger:      logger.NewLoggerWithLevel("debug"),
	}

	service := NewTelemetryService(config)

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the scheduler
	service.StartDailyScheduler(ctx)

	// The scheduler should start without error
	// We can't easily test the daily tick without waiting 24 hours,
	// but we can verify it doesn't panic or error on startup
	time.Sleep(100 * time.Millisecond) // Give it time to start

	// Cancel the context to stop the scheduler
	cancel()
	time.Sleep(100 * time.Millisecond) // Give it time to stop

	// Test passes if we reach here without panic
}

func TestTelemetryService_HardcodedEndpoint(t *testing.T) {
	// Verify that the hardcoded endpoint is used
	assert.Equal(t, "https://telemetry.notifuse.com", TelemetryEndpoint)
}
