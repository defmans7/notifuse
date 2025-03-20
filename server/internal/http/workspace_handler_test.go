package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
)

// mockWorkspaceService implements WorkspaceServiceInterface
type mockWorkspaceService struct {
	mock.Mock
}

func (m *mockWorkspaceService) CreateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, timezone, ownerID string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, timezone, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) GetWorkspace(ctx context.Context, id, ownerID string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) ListWorkspaces(ctx context.Context, ownerID string) ([]*domain.Workspace, error) {
	args := m.Called(ctx, ownerID)
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, timezone, ownerID string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, timezone, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) DeleteWorkspace(ctx context.Context, id, ownerID string) error {
	args := m.Called(ctx, id, ownerID)
	return args.Error(0)
}

// mockAuthService implements middleware.AuthServiceInterface
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// Test setup helper
func setupTest(t *testing.T) (*WorkspaceHandler, *mockWorkspaceService, *mockAuthService, *http.ServeMux, paseto.V4AsymmetricSecretKey) {
	workspaceSvc := &mockWorkspaceService{}
	authSvc := &mockAuthService{}

	// Create key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	handler := NewWorkspaceHandler(workspaceSvc, authSvc, publicKey)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return handler, workspaceSvc, authSvc, mux, secretKey
}

func createTestToken(t *testing.T, secretKey paseto.V4AsymmetricSecretKey, userID string) string {
	token := paseto.NewToken()
	token.SetAudience("test")
	token.SetIssuer("test")
	token.SetNotBefore(time.Now())
	token.SetIssuedAt(time.Now())
	token.SetExpiration(time.Now().Add(24 * time.Hour))
	token.SetString("user_id", userID)
	token.SetString("session_id", "test-session")

	signed := token.V4Sign(secretKey, nil)
	return signed
}

func TestWorkspaceHandler_Create(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Mock successful user session verification
	user := &domain.User{ID: "test-user"}
	authSvc.On("VerifyUserSession", mock.Anything, "test-user", "test-session").Return(user, nil)

	// Mock successful workspace creation
	expectedWorkspace := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}
	workspaceSvc.On("CreateWorkspace", mock.Anything, "testworkspace1", "Test Workspace", "https://example.com", "https://example.com/logo.png", "UTC", "test-user").Return(expectedWorkspace, nil)

	// Create request
	reqBody := createWorkspaceRequest{
		ID:         "testworkspace1",
		Name:       "Test Workspace",
		WebsiteURL: "https://example.com",
		LogoURL:    "https://example.com/logo.png",
		Timezone:   "UTC",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	// Execute request
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response domain.Workspace
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace.ID, response.ID)
	assert.Equal(t, expectedWorkspace.Name, response.Name)
	assert.Equal(t, expectedWorkspace.Settings, response.Settings)

	// Assert expectations
	workspaceSvc.AssertExpectations(t)
	authSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_Get(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Mock successful user session verification
	user := &domain.User{ID: "test-user"}
	authSvc.On("VerifyUserSession", mock.Anything, "test-user", "test-session").Return(user, nil)

	// Mock successful workspace retrieval
	expectedWorkspace := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			Timezone:   "UTC",
		},
	}
	workspaceSvc.On("GetWorkspace", mock.Anything, "testworkspace1", "test-user").Return(expectedWorkspace, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.get?id=testworkspace1", nil)
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	// Execute request
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Workspace
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace.ID, response.ID)
	assert.Equal(t, expectedWorkspace.Name, response.Name)
	assert.Equal(t, expectedWorkspace.Settings, response.Settings)

	// Assert expectations
	workspaceSvc.AssertExpectations(t)
	authSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_List(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Mock successful user session verification
	user := &domain.User{ID: "test-user"}
	authSvc.On("VerifyUserSession", mock.Anything, "test-user", "test-session").Return(user, nil)

	// Mock successful workspace list retrieval
	expectedWorkspaces := []*domain.Workspace{
		{
			ID:   "testworkspace1",
			Name: "Test Workspace 1",
			Settings: domain.WorkspaceSettings{
				WebsiteURL: "https://example1.com",
				LogoURL:    "https://example1.com/logo.png",
				Timezone:   "UTC",
			},
		},
		{
			ID:   "testworkspace2",
			Name: "Test Workspace 2",
			Settings: domain.WorkspaceSettings{
				WebsiteURL: "https://example2.com",
				LogoURL:    "https://example2.com/logo.png",
				Timezone:   "UTC",
			},
		},
	}
	workspaceSvc.On("ListWorkspaces", mock.Anything, "test-user").Return(expectedWorkspaces, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.list", nil)
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	// Execute request
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response []*domain.Workspace
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspaces, response)

	// Assert expectations
	workspaceSvc.AssertExpectations(t)
	authSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_Update(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Mock successful user session verification
	user := &domain.User{ID: "test-user"}
	authSvc.On("VerifyUserSession", mock.Anything, "test-user", "test-session").Return(user, nil)

	// Mock successful workspace update
	expectedWorkspace := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://updated.com",
			LogoURL:    "https://updated.com/logo.png",
			Timezone:   "UTC",
		},
	}
	workspaceSvc.On("UpdateWorkspace", mock.Anything, "testworkspace1", "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "UTC", "test-user").Return(expectedWorkspace, nil)

	// Create request
	reqBody := updateWorkspaceRequest{
		ID:         "testworkspace1",
		Name:       "Updated Workspace",
		WebsiteURL: "https://updated.com",
		LogoURL:    "https://updated.com/logo.png",
		Timezone:   "UTC",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	// Execute request
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Workspace
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace.ID, response.ID)
	assert.Equal(t, expectedWorkspace.Name, response.Name)
	assert.Equal(t, expectedWorkspace.Settings, response.Settings)

	// Assert expectations
	workspaceSvc.AssertExpectations(t)
	authSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_Delete(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Mock successful user session verification
	user := &domain.User{ID: "test-user"}
	authSvc.On("VerifyUserSession", mock.Anything, "test-user", "test-session").Return(user, nil)

	workspaceSvc.On("DeleteWorkspace", mock.Anything, "test-id", "test-user").
		Return(nil)

	request := deleteWorkspaceRequest{
		ID: "test-id",
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user"}))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

func TestWriteError(t *testing.T) {
	testCases := []struct {
		name           string
		status         int
		errorMessage   string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "bad request error",
			status:         http.StatusBadRequest,
			errorMessage:   "invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid input"}`,
		},
		{
			name:           "unauthorized error",
			status:         http.StatusUnauthorized,
			errorMessage:   "not authorized",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"not authorized"}`,
		},
		{
			name:           "internal server error",
			status:         http.StatusInternalServerError,
			errorMessage:   "server error",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"server error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test response recorder
			w := httptest.NewRecorder()

			// Call the function being tested
			writeError(w, tc.status, tc.errorMessage)

			// Assert the response status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Assert the response content type
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			// Assert the response body, trimming newlines
			assert.Equal(t, tc.expectedBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestWorkspaceHandler_List_MethodNotAllowed(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Try with POST instead of GET
	reqBody := bytes.NewBuffer([]byte("{}"))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.list", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleList(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_List_ServiceError(t *testing.T) {
	handler, workspaceService, _, _, _ := setupTest(t)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.list", nil)
	w := httptest.NewRecorder()

	// Mock workspace service to return error
	workspaceService.On("ListWorkspaces", mock.Anything, "user123").Return(
		([]*domain.Workspace)(nil), fmt.Errorf("database error"))

	// Setup context with authenticated user - no need for token validation here
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleList(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Failed to list workspaces", response.Error)

	// Verify mocks were called
	workspaceService.AssertExpectations(t)
}

func TestWorkspaceHandler_Get_MethodNotAllowed(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Try with POST instead of GET
	reqBody := bytes.NewBuffer([]byte(`{"id": "workspace123"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.get", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleGet(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Get_MissingID(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Create request without ID
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.get", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleGet(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Missing workspace ID", response.Error)
}

func TestWorkspaceHandler_Get_ServiceError(t *testing.T) {
	handler, workspaceService, _, _, secretKey := setupTest(t)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.get?id=workspace123", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Mock workspace service to return error
	workspaceService.On("GetWorkspace", mock.Anything, "workspace123", "user123").Return(
		(*domain.Workspace)(nil), fmt.Errorf("database error"))

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleGet(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Failed to get workspace", response.Error)

	// Verify mocks were called
	workspaceService.AssertExpectations(t)
}

func TestWorkspaceHandler_Create_MethodNotAllowed(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Try with GET instead of POST
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.create", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleCreate(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Create_InvalidBody(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Create invalid JSON request
	reqBody := bytes.NewBuffer([]byte(`{invalid json`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleCreate(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestWorkspaceHandler_Create_MissingID(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Create request with missing ID
	reqBody := bytes.NewBuffer([]byte(`{"name": "Test Workspace"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleCreate(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Workspace ID is required", response.Error)
}

func TestWorkspaceHandler_Create_ServiceError(t *testing.T) {
	handler, workspaceService, _, _, secretKey := setupTest(t)

	// Create valid request
	reqBody := bytes.NewBuffer([]byte(`{
		"id": "workspace123",
		"name": "Test Workspace",
		"website_url": "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"timezone": "UTC"
	}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Mock workspace service to return error
	workspaceService.On("CreateWorkspace",
		mock.Anything,
		"workspace123",
		"Test Workspace",
		"https://example.com",
		"https://example.com/logo.png",
		"UTC",
		"user123").Return(nil, fmt.Errorf("database error"))

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleCreate(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Failed to create workspace", response.Error)

	// Verify mocks were called
	workspaceService.AssertExpectations(t)
}

func TestWorkspaceHandler_Update_MethodNotAllowed(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Try with GET instead of POST
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.update", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleUpdate(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Update_InvalidBody(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Create invalid JSON request
	reqBody := bytes.NewBuffer([]byte(`{invalid json`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleUpdate(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestWorkspaceHandler_Update_MissingID(t *testing.T) {
	handler, workspaceService, _, _, secretKey := setupTest(t)

	// Create request with missing ID
	reqBody := bytes.NewBuffer([]byte(`{"name": "Test Workspace"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Mock the service call - the handler doesn't check for empty ID
	workspaceService.On("UpdateWorkspace",
		mock.Anything,
		"", // Empty ID
		"Test Workspace",
		"",
		"",
		"",
		"user123").Return(nil, fmt.Errorf("workspace ID is required"))

	// Call handler directly
	handler.handleUpdate(w, req)

	// Verify response - since the handler doesn't validate empty ID, we'll get an internal server error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Failed to update workspace", response.Error)

	// Verify mocks were called
	workspaceService.AssertExpectations(t)
}

func TestWorkspaceHandler_Update_ServiceError(t *testing.T) {
	handler, workspaceService, _, _, secretKey := setupTest(t)

	// Create valid request
	reqBody := bytes.NewBuffer([]byte(`{
		"id": "workspace123",
		"name": "Test Workspace",
		"website_url": "https://example.com",
		"logo_url": "https://example.com/logo.png",
		"timezone": "UTC"
	}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Mock workspace service to return error
	workspaceService.On("UpdateWorkspace",
		mock.Anything,
		"workspace123",
		"Test Workspace",
		"https://example.com",
		"https://example.com/logo.png",
		"UTC",
		"user123").Return(nil, fmt.Errorf("database error"))

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleUpdate(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Failed to update workspace", response.Error)

	// Verify mocks were called
	workspaceService.AssertExpectations(t)
}

func TestWorkspaceHandler_Delete_MethodNotAllowed(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Try with GET instead of POST
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.delete", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleDelete(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Delete_InvalidBody(t *testing.T) {
	handler, _, _, _, secretKey := setupTest(t)

	// Create invalid JSON request
	reqBody := bytes.NewBuffer([]byte(`{invalid json`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleDelete(w, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestWorkspaceHandler_Delete_MissingID(t *testing.T) {
	handler, workspaceService, _, _, secretKey := setupTest(t)

	// Create request with missing ID
	reqBody := bytes.NewBuffer([]byte(`{}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Mock the service call - the handler doesn't check for empty ID
	workspaceService.On("DeleteWorkspace", mock.Anything, "", "user123").
		Return(fmt.Errorf("workspace ID is required"))

	// Call handler directly
	handler.handleDelete(w, req)

	// Verify response - since the handler doesn't validate empty ID, we'll get an internal server error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Failed to delete workspace", response.Error)

	// Verify mocks were called
	workspaceService.AssertExpectations(t)
}

func TestWorkspaceHandler_Delete_ServiceError(t *testing.T) {
	handler, workspaceService, _, _, secretKey := setupTest(t)

	// Create valid request
	reqBody := bytes.NewBuffer([]byte(`{"id": "workspace123"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Mock workspace service to return error
	workspaceService.On("DeleteWorkspace", mock.Anything, "workspace123", "user123").
		Return(fmt.Errorf("database error"))

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.AuthUserKey, &middleware.AuthenticatedUser{
		ID: "user123",
	})
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleDelete(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify error message
	var response errorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Failed to delete workspace", response.Error)

	// Verify mocks were called
	workspaceService.AssertExpectations(t)
}
