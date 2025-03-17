package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notifuse/server/internal/http/middleware"
	"notifuse/server/internal/service"
)

// mockWorkspaceService implements WorkspaceServiceInterface
type mockWorkspaceService struct {
	mock.Mock
}

func (m *mockWorkspaceService) CreateWorkspace(ctx context.Context, name, websiteURL, logoURL, timezone, ownerID string) (*service.Workspace, error) {
	args := m.Called(ctx, name, websiteURL, logoURL, timezone, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) GetWorkspace(ctx context.Context, id, ownerID string) (*service.Workspace, error) {
	args := m.Called(ctx, id, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) ListWorkspaces(ctx context.Context, ownerID string) ([]service.Workspace, error) {
	args := m.Called(ctx, ownerID)
	return args.Get(0).([]service.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, timezone, ownerID string) (*service.Workspace, error) {
	args := m.Called(ctx, id, name, websiteURL, logoURL, timezone, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Workspace), args.Error(1)
}

func (m *mockWorkspaceService) DeleteWorkspace(ctx context.Context, id, ownerID string) error {
	args := m.Called(ctx, id, ownerID)
	return args.Error(0)
}

// mockAuthService implements middleware.AuthServiceInterface
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*service.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.User), args.Error(1)
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

	// Setup auth mock expectation
	authSvc.On("VerifyUserSession", mock.Anything, "test-user-id", "test-session").
		Return(&service.User{ID: "test-user-id", Email: "test@example.com"}, nil)

	tests := []struct {
		name           string
		request        createWorkspaceRequest
		mockWorkspace  *service.Workspace
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful creation",
			request: createWorkspaceRequest{
				Name:       "Test Workspace",
				WebsiteURL: "https://test.com",
				LogoURL:    "https://test.com/logo.png",
				Timezone:   "UTC",
			},
			mockWorkspace: &service.Workspace{
				ID:         "test-id",
				Name:       "Test Workspace",
				WebsiteURL: "https://test.com",
				LogoURL:    "https://test.com/logo.png",
				Timezone:   "UTC",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			workspaceSvc.On("CreateWorkspace",
				mock.Anything,
				tt.request.Name,
				tt.request.WebsiteURL,
				tt.request.LogoURL,
				tt.request.Timezone,
				"test-user-id",
			).Return(tt.mockWorkspace, tt.mockError)

			// Create request
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/workspaces.create", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user-id"))
			req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

			// Execute request
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response service.Workspace
				err = json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockWorkspace.ID, response.ID)
				assert.Equal(t, tt.mockWorkspace.Name, response.Name)
				assert.Equal(t, tt.mockWorkspace.WebsiteURL, response.WebsiteURL)
				assert.Equal(t, tt.mockWorkspace.LogoURL, response.LogoURL)
				assert.Equal(t, tt.mockWorkspace.Timezone, response.Timezone)
			}
		})
	}
}

func TestWorkspaceHandler_Get(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Setup auth mock expectation
	authSvc.On("VerifyUserSession", mock.Anything, "test-user-id", "test-session").
		Return(&service.User{ID: "test-user-id", Email: "test@example.com"}, nil)

	workspace := &service.Workspace{
		ID:         "test-id",
		Name:       "Test Workspace",
		WebsiteURL: "https://test.com",
		LogoURL:    "https://test.com/logo.png",
		Timezone:   "UTC",
	}

	workspaceSvc.On("GetWorkspace", mock.Anything, "test-id", "test-user-id").
		Return(workspace, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.get?id=test-id", nil)
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user-id"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.Workspace
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, response.ID)
	assert.Equal(t, workspace.Name, response.Name)
	assert.Equal(t, workspace.WebsiteURL, response.WebsiteURL)
	assert.Equal(t, workspace.LogoURL, response.LogoURL)
	assert.Equal(t, workspace.Timezone, response.Timezone)
}

func TestWorkspaceHandler_List(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Setup auth mock expectation
	authSvc.On("VerifyUserSession", mock.Anything, "test-user-id", "test-session").
		Return(&service.User{ID: "test-user-id", Email: "test@example.com"}, nil)

	workspaces := []service.Workspace{
		{
			ID:         "test-id-1",
			Name:       "Test Workspace 1",
			WebsiteURL: "https://test1.com",
			LogoURL:    "https://test1.com/logo.png",
			Timezone:   "UTC",
		},
		{
			ID:         "test-id-2",
			Name:       "Test Workspace 2",
			WebsiteURL: "https://test2.com",
			LogoURL:    "https://test2.com/logo.png",
			Timezone:   "UTC",
		},
	}

	workspaceSvc.On("ListWorkspaces", mock.Anything, "test-user-id").
		Return(workspaces, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/workspaces.list", nil)
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user-id"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []service.Workspace
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, workspaces[0].ID, response[0].ID)
	assert.Equal(t, workspaces[1].ID, response[1].ID)
}

func TestWorkspaceHandler_Update(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Setup auth mock expectation
	authSvc.On("VerifyUserSession", mock.Anything, "test-user-id", "test-session").
		Return(&service.User{ID: "test-user-id", Email: "test@example.com"}, nil)

	workspace := &service.Workspace{
		ID:         "test-id",
		Name:       "Updated Workspace",
		WebsiteURL: "https://updated.com",
		LogoURL:    "https://updated.com/logo.png",
		Timezone:   "UTC",
	}

	workspaceSvc.On("UpdateWorkspace", mock.Anything, "test-id", "Updated Workspace", "https://updated.com", "https://updated.com/logo.png", "UTC", "test-user-id").
		Return(workspace, nil)

	request := updateWorkspaceRequest{
		ID:         "test-id",
		Name:       "Updated Workspace",
		WebsiteURL: "https://updated.com",
		LogoURL:    "https://updated.com/logo.png",
		Timezone:   "UTC",
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user-id"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.Workspace
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, response.ID)
	assert.Equal(t, workspace.Name, response.Name)
	assert.Equal(t, workspace.WebsiteURL, response.WebsiteURL)
	assert.Equal(t, workspace.LogoURL, response.LogoURL)
	assert.Equal(t, workspace.Timezone, response.Timezone)
}

func TestWorkspaceHandler_Delete(t *testing.T) {
	_, workspaceSvc, authSvc, mux, secretKey := setupTest(t)

	// Setup auth mock expectation
	authSvc.On("VerifyUserSession", mock.Anything, "test-user-id", "test-session").
		Return(&service.User{ID: "test-user-id", Email: "test@example.com"}, nil)

	workspaceSvc.On("DeleteWorkspace", mock.Anything, "test-id", "test-user-id").
		Return(nil)

	request := deleteWorkspaceRequest{
		ID: "test-id",
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+createTestToken(t, secretKey, "test-user-id"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}
