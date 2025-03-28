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
	CreateWorkspaceFn func(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error)
	GetWorkspaceFn    func(ctx context.Context, id string) (*domain.Workspace, error)
	ListWorkspacesFn  func(ctx context.Context) ([]*domain.Workspace, error)
	UpdateWorkspaceFn func(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error)
	DeleteWorkspaceFn func(ctx context.Context, id string) error
	InviteMemberFn    func(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, string, error)

	GetWorkspaceMembersWithEmailFn func(ctx context.Context, id string) ([]*domain.UserWorkspaceWithEmail, error)
	GetUserWorkspaceFn             func(ctx context.Context, userID, workspaceID string) (*domain.UserWorkspace, error)
}

func (m *mockWorkspaceService) CreateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) ListWorkspaces(ctx context.Context) ([]*domain.Workspace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string) (*domain.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, coverURL, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockWorkspaceService) InviteMember(ctx context.Context, workspaceID, email string) (*domain.WorkspaceInvitation, string, error) {
	args := m.Called(ctx, workspaceID, email)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.WorkspaceInvitation), args.String(1), args.Error(2)
}

func (m *mockWorkspaceService) GetWorkspaceMembersWithEmail(ctx context.Context, id string) ([]*domain.UserWorkspaceWithEmail, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserWorkspaceWithEmail), args.Error(1)
}

func (m *mockWorkspaceService) GetUserWorkspace(ctx context.Context, userID, workspaceID string) (*domain.UserWorkspace, error) {
	args := m.Called(ctx, userID, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWorkspace), args.Error(1)
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

func (m *mockAuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// Test setup helper
func setupTest(t *testing.T) (*WorkspaceHandler, *mockWorkspaceService, *http.ServeMux, paseto.V4AsymmetricSecretKey) {
	workspaceSvc := new(mockWorkspaceService)

	// Create key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	mockLogger := new(MockLogger)
	handler := NewWorkspaceHandler(workspaceSvc, publicKey, mockLogger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return handler, workspaceSvc, mux, secretKey
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
	_, workspaceSvc, mux, secretKey := setupTest(t)

	// Mock successful workspace creation
	expectedWorkspace := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
		},
	}
	workspaceSvc.On("CreateWorkspace", mock.Anything, "testworkspace1", "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC").Return(expectedWorkspace, nil)

	// Create request
	reqBody := domain.CreateWorkspaceRequest{
		ID:   "testworkspace1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettingsData{
			Name:       "Test Workspace",
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
		},
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
}

func TestWorkspaceHandler_Get(t *testing.T) {
	_, workspaceSvc, mux, secretKey := setupTest(t)

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
	workspaceSvc.On("GetWorkspace", mock.Anything, "testworkspace1").Return(expectedWorkspace, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.get?id=testworkspace1", nil)
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	// Execute request
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Updated to expect a response with a workspace field
	var response struct {
		Workspace domain.Workspace `json:"workspace"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace.ID, response.Workspace.ID)
	assert.Equal(t, expectedWorkspace.Name, response.Workspace.Name)
	assert.Equal(t, expectedWorkspace.Settings, response.Workspace.Settings)

	// Assert expectations
	workspaceSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_List(t *testing.T) {
	_, workspaceSvc, mux, secretKey := setupTest(t)

	// Mock successful user session verification
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
	workspaceSvc.On("ListWorkspaces", mock.Anything).Return(expectedWorkspaces, nil)

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
}

func TestWorkspaceHandler_Update(t *testing.T) {
	_, workspaceSvc, mux, secretKey := setupTest(t)

	// Mock successful workspace update
	expectedWorkspace := &domain.Workspace{
		ID:   "testworkspace1",
		Name: "Updated Workspace",
		Settings: domain.WorkspaceSettings{
			WebsiteURL: "https://updated.com",
			LogoURL:    "https://updated.com/logo.png",
			CoverURL:   "https://updated.com/cover.png",
			Timezone:   "UTC",
		},
	}
	workspaceSvc.On("UpdateWorkspace", mock.Anything, "testworkspace1", "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "UTC").Return(expectedWorkspace, nil)

	// Create request
	reqBody := domain.UpdateWorkspaceRequest{
		ID:         "testworkspace1",
		Name:       "Updated Workspace",
		WebsiteURL: "https://updated.com",
		LogoURL:    "https://updated.com/logo.png",
		CoverURL:   "https://updated.com/cover.png",
		Timezone:   "UTC",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Workspace
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedWorkspace.ID, response.ID)
	assert.Equal(t, expectedWorkspace.Name, response.Name)
	assert.Equal(t, expectedWorkspace.Settings, response.Settings)

	workspaceSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_Delete(t *testing.T) {
	_, workspaceSvc, mux, secretKey := setupTest(t)

	// Mock successful workspace deletion
	workspaceSvc.On("DeleteWorkspace", mock.Anything, "test-id").Return(nil)

	// Create request
	reqBody := domain.DeleteWorkspaceRequest{
		ID: "test-id",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	// Execute request
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert expectations
	workspaceSvc.AssertExpectations(t)
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
	handler, _, _, secretKey := setupTest(t)

	// Try with POST instead of GET
	reqBody := bytes.NewBuffer([]byte("{}"))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.list", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleList(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_List_ServiceError(t *testing.T) {
	handler, workspaceService, _, _ := setupTest(t)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.list", nil)
	w := httptest.NewRecorder()

	// Mock workspace service to return error
	workspaceService.On("ListWorkspaces", mock.Anything).Return(
		([]*domain.Workspace)(nil), fmt.Errorf("database error"))

	// Setup context with authenticated user - no need for token validation here
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
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
	handler, _, _, secretKey := setupTest(t)

	// Try with POST instead of GET
	reqBody := bytes.NewBuffer([]byte(`{"id": "workspace123"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.get", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleGet(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Get_MissingID(t *testing.T) {
	handler, _, _, secretKey := setupTest(t)

	// Create request without ID
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.get", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
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
	handler, workspaceService, _, secretKey := setupTest(t)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.get?id=workspace123", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Mock workspace service to return error
	workspaceService.On("GetWorkspace", mock.Anything, "workspace123").Return(
		(*domain.Workspace)(nil), fmt.Errorf("database error"))

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
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
	handler, _, _, secretKey := setupTest(t)

	// Try with GET instead of POST
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.create", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleCreate(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Create_InvalidBody(t *testing.T) {
	handler, _, _, secretKey := setupTest(t)

	// Create invalid JSON request
	reqBody := bytes.NewBuffer([]byte(`{invalid json`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
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
	_, _, mux, secretKey := setupTest(t)

	// Create request with missing ID
	reqBody := domain.CreateWorkspaceRequest{
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettingsData{
			Name:       "Test Workspace",
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
		},
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "id: non zero value required")
}

func TestWorkspaceHandler_Create_MissingName(t *testing.T) {
	_, _, mux, secretKey := setupTest(t)

	// Create request with missing name
	reqBody := domain.CreateWorkspaceRequest{
		ID: "testworkspace1",
		Settings: domain.WorkspaceSettingsData{
			Name:       "Test Workspace",
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
		},
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name: non zero value required")
}

func TestWorkspaceHandler_Create_MissingTimezone(t *testing.T) {
	_, _, mux, secretKey := setupTest(t)

	// Create request with missing timezone
	reqBody := domain.CreateWorkspaceRequest{
		ID:   "testworkspace1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettingsData{
			Name:       "Test Workspace",
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
		},
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "timezone: non zero value required")
}

func TestWorkspaceHandler_Create_ServiceError(t *testing.T) {
	_, workspaceSvc, mux, secretKey := setupTest(t)

	// Mock service error
	workspaceSvc.On("CreateWorkspace", mock.Anything, "testworkspace1", "Test Workspace", "https://example.com", "https://example.com/logo.png", "https://example.com/cover.png", "UTC").Return(nil, fmt.Errorf("database error"))

	// Create request with valid data
	reqBody := domain.CreateWorkspaceRequest{
		ID:   "testworkspace1",
		Name: "Test Workspace",
		Settings: domain.WorkspaceSettingsData{
			Name:       "Test Workspace",
			WebsiteURL: "https://example.com",
			LogoURL:    "https://example.com/logo.png",
			CoverURL:   "https://example.com/cover.png",
			Timezone:   "UTC",
		},
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to create workspace")

	workspaceSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_Update_MethodNotAllowed(t *testing.T) {
	handler, _, _, secretKey := setupTest(t)

	// Try with GET instead of POST
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.update", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleUpdate(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Update_InvalidBody(t *testing.T) {
	handler, _, _, secretKey := setupTest(t)

	// Create invalid JSON request
	reqBody := bytes.NewBuffer([]byte(`{invalid json`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
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
	_, _, mux, secretKey := setupTest(t)

	// Create request with missing ID
	reqBody := domain.UpdateWorkspaceRequest{
		Name:       "Updated Workspace",
		WebsiteURL: "https://updated.com",
		LogoURL:    "https://updated.com/logo.png",
		CoverURL:   "https://updated.com/cover.png",
		Timezone:   "UTC",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "id: non zero value required")
}

func TestWorkspaceHandler_Update_ServiceError(t *testing.T) {
	_, workspaceSvc, mux, secretKey := setupTest(t)

	// Mock service error
	workspaceSvc.On("UpdateWorkspace", mock.Anything, "testworkspace1", "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "https://updated.com/cover.png", "UTC").Return(nil, fmt.Errorf("database error"))

	// Create request with valid data
	reqBody := domain.UpdateWorkspaceRequest{
		ID:         "testworkspace1",
		Name:       "Updated Workspace",
		WebsiteURL: "https://updated.com",
		LogoURL:    "https://updated.com/logo.png",
		CoverURL:   "https://updated.com/cover.png",
		Timezone:   "UTC",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to update workspace")

	workspaceSvc.AssertExpectations(t)
}

func TestWorkspaceHandler_Delete_MethodNotAllowed(t *testing.T) {
	handler, _, _, secretKey := setupTest(t)

	// Try with GET instead of POST
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.delete", nil)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
	req = req.WithContext(ctx)

	// Call handler directly
	handler.handleDelete(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestWorkspaceHandler_Delete_InvalidBody(t *testing.T) {
	handler, _, _, secretKey := setupTest(t)

	// Create invalid JSON request
	reqBody := bytes.NewBuffer([]byte(`{invalid json`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
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
	handler, workspaceService, _, secretKey := setupTest(t)

	// Create request with missing ID
	reqBody := bytes.NewBuffer([]byte(`{}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
	req = req.WithContext(ctx)

	// Mock the service call - the handler doesn't check for empty ID
	workspaceService.On("DeleteWorkspace", mock.Anything, "").
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
	handler, workspaceService, _, secretKey := setupTest(t)

	// Create valid request
	reqBody := bytes.NewBuffer([]byte(`{"id": "workspace123"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", reqBody)
	w := httptest.NewRecorder()

	// Add auth token
	token := createTestToken(t, secretKey, "user123")
	req.Header.Set("Authorization", "Bearer "+token)

	// Mock workspace service to return error
	workspaceService.On("DeleteWorkspace", mock.Anything, "workspace123").
		Return(fmt.Errorf("database error"))

	// Setup context with authenticated user
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.UserIDKey, "user123")
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

func TestWorkspaceHandler_HandleMembers(t *testing.T) {
	_, workspaceSvc, mux, secretKey := setupTest(t)

	// Mock successful members retrieval
	expectedMembers := []*domain.UserWorkspaceWithEmail{
		{
			UserWorkspace: domain.UserWorkspace{
				UserID:      "user1",
				WorkspaceID: "workspace1",
				Role:        "owner",
			},
			Email: "user1@example.com",
		},
	}
	workspaceSvc.On("GetWorkspaceMembersWithEmail", mock.Anything, "workspace1").Return(expectedMembers, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.members?id=workspace1", nil)
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user"))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response, "members")

	workspaceSvc.AssertExpectations(t)
}
