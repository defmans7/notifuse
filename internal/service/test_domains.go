package service

import (
	"net/http"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/golang/mock/gomock"
)

// CreateMockHTTPClient creates a mock HTTP client for testing
func CreateMockHTTPClient(t *testing.T) *mocks.MockHTTPClient {
	ctrl := gomock.NewController(t)
	return mocks.NewMockHTTPClient(ctrl)
}

// CreateMockAuthService creates a mock auth service for testing
func CreateMockAuthService(t *testing.T) *mocks.MockAuthService {
	ctrl := gomock.NewController(t)
	return mocks.NewMockAuthService(ctrl)
}

// MockHTTPResponse creates a mock HTTP response with JSON data
func MockHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       http.NoBody,
	}
}
