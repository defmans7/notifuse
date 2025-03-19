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

	"notifuse/server/config"
	"notifuse/server/internal/domain"
	"notifuse/server/internal/http/middleware"
	"notifuse/server/internal/service"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) SignIn(ctx context.Context, input service.SignInInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *mockUserService) SignInDev(ctx context.Context, input service.SignInInput) (string, error) {
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

func (m *mockUserWorkspaceService) CreateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, timezone, ownerID string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, timezone, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) GetWorkspace(ctx context.Context, id, ownerID string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) ListWorkspaces(ctx context.Context, ownerID string) ([]*domain.Workspace, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, timezone, ownerID string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, timezone, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockUserWorkspaceService) DeleteWorkspace(ctx context.Context, id, ownerID string) error {
	args := m.Called(ctx, id, ownerID)
	return args.Error(0)
}

func TestUserHandler_SignIn(t *testing.T) {
	mockService := new(mockUserService)
	mockWorkspaceSvc := new(mockUserWorkspaceService)
	devConfig := &config.Config{Environment: "development"}
	prodConfig := &config.Config{Environment: "production"}

	// Create a test key
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	devHandler := NewUserHandler(mockService, mockWorkspaceSvc, devConfig, publicKey)
	prodHandler := NewUserHandler(mockService, mockWorkspaceSvc, prodConfig, publicKey)

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
				}).Return(nil)
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
				mockService.On("SignInDev", mock.Anything, service.SignInInput{
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
				}).Return(fmt.Errorf("invalid email"))
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
				mockService.On("SignInDev", mock.Anything, service.SignInInput{
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
	config := &config.Config{Environment: "production"}

	// Create a test key
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	handler := NewUserHandler(mockService, mockWorkspaceSvc, config, publicKey)

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
	config := &config.Config{Environment: "production"}

	// Create a test key
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	handler := NewUserHandler(mockUserSvc, mockWorkspaceSvc, config, publicKey)

	userID := uuid.New().String()
	user := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	workspaces := []*domain.Workspace{
		{
			ID:   uuid.New().String(),
			Name: "Workspace 1",
			Settings: domain.WorkspaceSettings{
				WebsiteURL: "https://example.com",
				LogoURL:    "https://example.com/logo.png",
				Timezone:   "UTC",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:   uuid.New().String(),
			Name: "Workspace 2",
			Settings: domain.WorkspaceSettings{
				WebsiteURL: "https://example2.com",
				LogoURL:    "https://example2.com/logo.png",
				Timezone:   "UTC",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Test successful retrieval
	mockUserSvc.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	mockWorkspaceSvc.On("ListWorkspaces", mock.Anything, userID).Return(workspaces, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	// Add authenticated user to context
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID:    userID,
		Email: user.Email,
	}))
	rec := httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Check user data
	userData := response["user"].(map[string]interface{})
	assert.Equal(t, user.ID, userData["id"])
	assert.Equal(t, user.Email, userData["email"])
	assert.Equal(t, user.Name, userData["name"])

	// Check workspaces data
	workspacesData := response["workspaces"].([]interface{})
	assert.Equal(t, 2, len(workspacesData))

	// Test unauthorized access
	req = httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	rec = httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Test user not found
	mockUserSvc.On("GetUserByID", mock.Anything, "unknown-user-id").Return(nil, fmt.Errorf("user not found"))

	req = httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID:    "unknown-user-id",
		Email: "unknown@example.com",
	}))
	rec = httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// Test workspaces retrieval error
	mockUserSvc.On("GetUserByID", mock.Anything, "error-workspace-user").Return(user, nil)
	mockWorkspaceSvc.On("ListWorkspaces", mock.Anything, "error-workspace-user").Return(nil, fmt.Errorf("database error"))

	req = httptest.NewRequest(http.MethodGet, "/api/user.me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID:    "error-workspace-user",
		Email: user.Email,
	}))
	rec = httptest.NewRecorder()

	handler.GetCurrentUser(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockUserSvc.AssertExpectations(t)
	mockWorkspaceSvc.AssertExpectations(t)
}

func TestUserHandler_RegisterRoutes(t *testing.T) {
	// Generate a PASETO key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Test cases for different scenarios
	testCases := []struct {
		name            string
		setupMocks      func(mockUserSvc *mockUserService, mockWorkspaceSvc *mockWorkspaceService)
		testPath        string
		expectedHandler bool
	}{
		{
			name: "public routes",
			setupMocks: func(mockUserSvc *mockUserService, mockWorkspaceSvc *mockWorkspaceService) {
				// No need to set up AuthServiceInterface expectations
				// because we're testing that these routes don't require auth
			},
			testPath:        "/api/user.signin",
			expectedHandler: true,
		},
		{
			name: "protected routes with auth service",
			setupMocks: func(mockUserSvc *mockUserService, mockWorkspaceSvc *mockWorkspaceService) {
				// Make mockUserService implement AuthServiceInterface
				mockUserSvc.On("VerifyUserSession", mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.User{ID: "user1", Email: "test@example.com"}, nil)

				// Mock GetUserByID which is called in GetCurrentUser
				mockUserSvc.On("GetUserByID", mock.Anything, "user1").
					Return(&domain.User{ID: "user1", Email: "test@example.com"}, nil)

				// Mock ListWorkspaces which is also called in GetCurrentUser
				mockWorkspaceSvc.On("ListWorkspaces", mock.Anything, "user1").
					Return([]*domain.Workspace{}, nil)
			},
			testPath:        "/api/user.me",
			expectedHandler: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup new mock services for each test case to avoid interference
			mockUserSvc := &mockUserService{}
			mockWorkspaceSvc := &mockWorkspaceService{}

			// Create a config with test values
			cfg := &config.Config{
				Security: config.SecurityConfig{
					PasetoPublicKey: []byte("key"),
				},
			}

			// Create the handler
			handler := NewUserHandler(mockUserSvc, mockWorkspaceSvc, cfg, publicKey)

			// Create a new HTTP multiplexer for each test case
			mux := http.NewServeMux()

			// Set up mocks for this test case
			tc.setupMocks(mockUserSvc, mockWorkspaceSvc)

			// Register routes
			handler.RegisterRoutes(mux)

			// Test server for this multiplexer
			server := httptest.NewServer(mux)
			defer server.Close()

			// Make a request to the test path
			req, err := http.NewRequest("GET", server.URL+tc.testPath, nil)
			require.NoError(t, err)

			// For protected routes, we need to add a valid token
			if tc.testPath == "/api/user.me" {
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
			if tc.expectedHandler {
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}
