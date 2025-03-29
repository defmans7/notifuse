package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/internal/service"
)

func TestUserHandler_SignIn(t *testing.T) {
	mockService := new(service.MockUserService)
	mockWorkspaceSvc := new(service.MockWorkspaceService)

	// Test with different configs
	devConfig := &config.Config{Environment: "development"}
	prodConfig := &config.Config{Environment: "production"}

	// Create a test key
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	mockLogger := &service.MockLogger{}

	// Use type assertion to treat mockService as a service.UserServiceInterface
	devHandler := NewUserHandler(mockService, mockWorkspaceSvc, devConfig, publicKey, mockLogger)
	prodHandler := NewUserHandler(mockService, mockWorkspaceSvc, prodConfig, publicKey, mockLogger)

	tests := []struct {
		name         string
		handler      *UserHandler
		input        domain.SignInInput
		setupMock    func()
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name:    "successful sign in production",
			handler: prodHandler,
			input: domain.SignInInput{
				Email: "test@example.com",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, domain.SignInInput{
					Email: "test@example.com",
				}).Return("", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{
				"message": "Magic code sent to your email",
			},
		},
		{
			name:    "successful sign in development",
			handler: devHandler,
			input: domain.SignInInput{
				Email: "test@example.com",
			},
			setupMock: func() {
				// Mock the SignIn method to return the 6-digit code for development mode
				mockService.Mock = mock.Mock{} // Reset mock to avoid conflicts
				mockService.On("SignIn", mock.Anything, domain.SignInInput{
					Email: "test@example.com",
				}).Return("123456", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{
				"message": "Magic code sent to your email",
				"code":    "123456",
			},
		},
		{
			name:    "invalid email production",
			handler: prodHandler,
			input: domain.SignInInput{
				Email: "",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, domain.SignInInput{
					Email: "",
				}).Return("", fmt.Errorf("invalid email"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"error": "invalid email",
			},
		},
		{
			name:    "invalid email development",
			handler: devHandler,
			input: domain.SignInInput{
				Email: "",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, domain.SignInInput{
					Email: "",
				}).Return("", fmt.Errorf("invalid email"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"error": "invalid email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/user.signin", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			tt.handler.SignIn(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]string
			err = json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_VerifyCode(t *testing.T) {
	mockService := new(service.MockUserService)
	mockWorkspaceSvc := new(service.MockWorkspaceService)
	config := &config.Config{}

	// Create a test key
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	mockLogger := &service.MockLogger{}

	// Use type assertion for mockService
	handler := NewUserHandler(mockService, mockWorkspaceSvc, config, publicKey, mockLogger)

	user := &domain.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name          string
		input         domain.VerifyCodeInput
		setupMock     func()
		expectedCode  int
		checkResponse func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "successful verification",
			input: domain.VerifyCodeInput{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMock: func() {
				mockService.On("VerifyCode", mock.Anything, domain.VerifyCodeInput{
					Email: "test@example.com",
					Code:  "123456",
				}).Return(&domain.AuthResponse{
					Token:     "auth-token",
					User:      *user,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "auth-token", response["token"])
				assert.Equal(t, user.ID, response["user"].(map[string]interface{})["id"])
				assert.Equal(t, user.Email, response["user"].(map[string]interface{})["email"])
				assert.Equal(t, user.Name, response["user"].(map[string]interface{})["name"])
				assert.NotEmpty(t, response["expires_at"])
			},
		},
		{
			name: "invalid code",
			input: domain.VerifyCodeInput{
				Email: "test@example.com",
				Code:  "000000",
			},
			setupMock: func() {
				mockService.On("VerifyCode", mock.Anything, domain.VerifyCodeInput{
					Email: "test@example.com",
					Code:  "000000",
				}).Return(nil, fmt.Errorf("invalid or expired code"))
			},
			expectedCode: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "invalid or expired code", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/user.verify", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.VerifyCode(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]interface{}
			err = json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			tt.checkResponse(t, response)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetCurrentUser(t *testing.T) {
	mockUserSvc := new(service.MockUserService)
	mockWorkspaceSvc := new(service.MockWorkspaceService)
	cfg := &config.Config{}
	publicKey := paseto.V4AsymmetricPublicKey{}
	mockLogger := &service.MockLogger{}

	handler := NewUserHandler(mockUserSvc, mockWorkspaceSvc, cfg, publicKey, mockLogger)

	// Test successful case
	userID := "test-user"
	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
	}
	workspaces := []*domain.Workspace{
		{
			ID:   "workspace1",
			Name: "Workspace 1",
		},
		{
			ID:   "workspace2",
			Name: "Workspace 2",
		},
	}

	mockUserSvc.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	mockWorkspaceSvc.On("ListWorkspaces", mock.Anything).Return(workspaces, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID))
	rec := httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	userData := response["user"].(map[string]interface{})
	assert.Equal(t, user.Email, userData["email"])
	assert.Equal(t, user.Name, userData["name"])

	workspacesData := response["workspaces"].([]interface{})
	assert.Equal(t, 2, len(workspacesData))

	// Test unauthorized access
	req = httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	rec = httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Test user not found
	notFoundUserID := "unknown-user-id"
	mockUserSvc.On("GetUserByID", mock.Anything, notFoundUserID).Return(nil, fmt.Errorf("user not found"))

	req = httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, notFoundUserID))
	rec = httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// Test workspaces retrieval error
	errorUserID := "error-workspace-user"
	mockUserSvc.On("GetUserByID", mock.Anything, errorUserID).Return(user, nil)
	mockWorkspaceSvc.On("ListWorkspaces", mock.Anything).Return(nil, fmt.Errorf("database error"))

	req = httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, errorUserID))
	rec = httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Failed to retrieve workspaces")

	mockUserSvc.AssertExpectations(t)
	mockWorkspaceSvc.AssertExpectations(t)
}

func TestUserHandler_RegisterRoutes(t *testing.T) {
	mockUserSvc := new(service.MockUserService)
	mockWorkspaceSvc := new(service.MockWorkspaceService)
	cfg := &config.Config{}

	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	mockLogger := &service.MockLogger{}

	// Set up mock expectation for VerifyUserSession to prevent unexpected call error
	mockUserSvc.On("VerifyUserSession", mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.User{ID: "user1", Email: "user@example.com"}, nil)

	// Set up mock expectation for GetUserByID with specific user ID
	mockUserSvc.On("GetUserByID", mock.Anything, "user1").
		Return(&domain.User{ID: "user1", Email: "user@example.com"}, nil)

	// Set up mock expectation for ListWorkspaces
	mockWorkspaceSvc.On("ListWorkspaces", mock.Anything).
		Return([]*domain.Workspace{}, nil)

	// Use type assertion for mockUserSvc
	handler := NewUserHandler(mockUserSvc, mockWorkspaceSvc, cfg, publicKey, mockLogger)

	// Test cases for different scenarios
	testCases := []struct {
		name       string
		route      string
		setupMocks func(userSvc *service.MockUserService, workspaceSvc *service.MockWorkspaceService)
	}{
		{
			name:  "public routes",
			route: "/api/user.signin",
			setupMocks: func(userSvc *service.MockUserService, workspaceSvc *service.MockWorkspaceService) {
				// No mock setup needed for testing route registration
			},
		},
		{
			name:  "protected routes with auth service",
			route: "/api/user.me",
			setupMocks: func(userSvc *service.MockUserService, workspaceSvc *service.MockWorkspaceService) {
				// Setup mock for auth middleware
				userSvc.On("GetUserByID", mock.Anything, mock.Anything).Return(&domain.User{
					ID:    "user1",
					Email: "user@example.com",
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup new mock services for each test case to avoid interference
			mockUserSvc := new(service.MockUserService)
			mockWorkspaceSvc := new(service.MockWorkspaceService)

			// Set up mocks for this test case
			tc.setupMocks(mockUserSvc, mockWorkspaceSvc)

			// Create a new HTTP multiplexer for each test case
			mux := http.NewServeMux()

			// Register routes
			handler.RegisterRoutes(mux)

			// Test server for this multiplexer
			server := httptest.NewServer(mux)
			defer server.Close()

			// Make a request to the test path
			req, err := http.NewRequest("GET", server.URL+tc.route, nil)
			require.NoError(t, err)

			// For protected routes, we need to add a valid token
			if tc.route == "/api/user.me" {
				token := paseto.NewToken()
				token.SetString("user_id", "user1")
				token.SetString("session_id", "session1")
				token.SetExpiration(time.Now().Add(time.Hour))

				signedToken := token.V4Sign(secretKey, nil)
				req.Header.Set("Authorization", "Bearer "+signedToken)
			}

			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)

			// We don't care about the response content, just that a handler was registered
			// and it didn't return 404 Not Found
			if tc.route == "/api/user.me" {
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}
