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
	"github.com/Notifuse/notifuse/pkg/logger"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) SignIn(ctx context.Context, input service.SignInInput) (string, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *mockUserService) VerifyCode(ctx context.Context, input service.VerifyCodeInput) (*service.AuthResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResponse), args.Error(1)
}

func (m *mockUserService) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type mockUserWorkspaceService struct {
	mock.Mock
}

func (m *mockUserWorkspaceService) CreateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockUserWorkspaceService) GetWorkspaceMembersWithEmail(ctx context.Context, id string) ([]*domain.UserWorkspaceWithEmail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserWorkspaceWithEmail), args.Error(1)
}

func (m *mockUserWorkspaceService) InviteMember(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, string, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.WorkspaceInvitation), args.String(1), args.Error(2)
}

func TestUserHandler_SignIn(t *testing.T) {
	mockService := new(mockUserService)
	mockWorkspaceSvc := new(mockUserWorkspaceService)

	// Test with different configs
	devConfig := &config.Config{Environment: "development"}
	prodConfig := &config.Config{Environment: "production"}

	// Create a test key
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	mockLogger := &MockLogger{}

	// Use type assertion to treat mockService as a service.UserServiceInterface
	devHandler := NewUserHandler(mockService, mockWorkspaceSvc, devConfig, publicKey, mockLogger)
	prodHandler := NewUserHandler(mockService, mockWorkspaceSvc, prodConfig, publicKey, mockLogger)

	tests := []struct {
		name         string
		handler      *UserHandler
		input        service.SignInInput
		setupMock    func()
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name:    "successful sign in production",
			handler: prodHandler,
			input: service.SignInInput{
				Email: "test@example.com",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, service.SignInInput{
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
			input: service.SignInInput{
				Email: "test@example.com",
			},
			setupMock: func() {
				// Mock the SignIn method to return the 6-digit code for development mode
				mockService.Mock = mock.Mock{} // Reset mock to avoid conflicts
				mockService.On("SignIn", mock.Anything, service.SignInInput{
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
			input: service.SignInInput{
				Email: "",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, service.SignInInput{
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
			input: service.SignInInput{
				Email: "",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, service.SignInInput{
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
	mockService := new(mockUserService)
	mockWorkspaceSvc := new(mockUserWorkspaceService)
	config := &config.Config{}

	// Create a test key
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	mockLogger := &MockLogger{}

	// Use type assertion for mockService
	handler := NewUserHandler(mockService, mockWorkspaceSvc, config, publicKey, mockLogger)

	user := &domain.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name          string
		input         service.VerifyCodeInput
		setupMock     func()
		expectedCode  int
		checkResponse func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "successful verification",
			input: service.VerifyCodeInput{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMock: func() {
				mockService.On("VerifyCode", mock.Anything, service.VerifyCodeInput{
					Email: "test@example.com",
					Code:  "123456",
				}).Return(&service.AuthResponse{
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
			input: service.VerifyCodeInput{
				Email: "test@example.com",
				Code:  "000000",
			},
			setupMock: func() {
				mockService.On("VerifyCode", mock.Anything, service.VerifyCodeInput{
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
	mockUserSvc := new(mockUserService)
	mockWorkspaceSvc := new(mockUserWorkspaceService)
	cfg := &config.Config{}
	publicKey := paseto.V4AsymmetricPublicKey{}
	mockLogger := &MockLogger{}

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
	mockUserSvc := new(mockUserService)
	mockWorkspaceSvc := new(mockUserWorkspaceService)
	cfg := &config.Config{}

	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	mockLogger := &MockLogger{}

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
		setupMocks func(userSvc *mockUserService, workspaceSvc *mockUserWorkspaceService)
	}{
		{
			name:  "public routes",
			route: "/api/user.signin",
			setupMocks: func(userSvc *mockUserService, workspaceSvc *mockUserWorkspaceService) {
				// No mock setup needed for testing route registration
			},
		},
		{
			name:  "protected routes with auth service",
			route: "/api/user.me",
			setupMocks: func(userSvc *mockUserService, workspaceSvc *mockUserWorkspaceService) {
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
			mockUserSvc := &mockUserService{}
			mockWorkspaceSvc := &mockUserWorkspaceService{}

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

// Add a mock logger struct
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string) {
}

func (m *MockLogger) Error(msg string) {
}

func (m *MockLogger) Debug(msg string) {
}

func (m *MockLogger) Warn(msg string) {
}

func (m *MockLogger) Fatal(msg string) {
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	return m
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	return m
}

func (m *MockLogger) WithError(err error) logger.Logger {
	return m
}
