package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"aidanwoods.dev/go-paseto"
	"github.com/go-chi/chi/v5"
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

func (m *mockAuthService) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	args := m.Called(ctx, sessionID)
	return args.String(0), args.Error(1)
}

// Test setup helper
func setupTest(t *testing.T) (*WorkspaceHandler, *mockWorkspaceService, *mockAuthService, *chi.Mux) {
	workspaceSvc := &mockWorkspaceService{}
	authSvc := &mockAuthService{}
	handler := NewWorkspaceHandler(workspaceSvc, authSvc)

	router := chi.NewRouter()
	// Create a dummy public key for testing
	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromHex("1eb9dbbbbc047c03fd70604e0071f0987e16b28b757225c11f00415d0e20b1a2")
	require.NoError(t, err)

	handler.RegisterRoutes(router, publicKey)

	return handler, workspaceSvc, authSvc, router
}

func TestWorkspaceHandler_Create(t *testing.T) {
	_, workspaceSvc, _, router := setupTest(t)

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
			req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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
	_, workspaceSvc, _, router := setupTest(t)

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
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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
	_, workspaceSvc, _, router := setupTest(t)

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
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []service.Workspace
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, workspaces[0].ID, response[0].ID)
	assert.Equal(t, workspaces[1].ID, response[1].ID)
}

func TestWorkspaceHandler_Update(t *testing.T) {
	_, workspaceSvc, _, router := setupTest(t)

	request := updateWorkspaceRequest{
		ID:         "test-id",
		Name:       "Updated Workspace",
		WebsiteURL: "https://updated.com",
		LogoURL:    "https://updated.com/logo.png",
		Timezone:   "UTC",
	}

	updatedWorkspace := &service.Workspace{
		ID:         request.ID,
		Name:       request.Name,
		WebsiteURL: request.WebsiteURL,
		LogoURL:    request.LogoURL,
		Timezone:   request.Timezone,
	}

	workspaceSvc.On("UpdateWorkspace",
		mock.Anything,
		request.ID,
		request.Name,
		request.WebsiteURL,
		request.LogoURL,
		request.Timezone,
		"test-user-id",
	).Return(updatedWorkspace, nil)

	body, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.update", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.Workspace
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, updatedWorkspace.ID, response.ID)
	assert.Equal(t, updatedWorkspace.Name, response.Name)
	assert.Equal(t, updatedWorkspace.WebsiteURL, response.WebsiteURL)
	assert.Equal(t, updatedWorkspace.LogoURL, response.LogoURL)
	assert.Equal(t, updatedWorkspace.Timezone, response.Timezone)
}

func TestWorkspaceHandler_Delete(t *testing.T) {
	_, workspaceSvc, _, router := setupTest(t)

	workspaceSvc.On("DeleteWorkspace", mock.Anything, "test-id", "test-user-id").
		Return(nil)

	request := deleteWorkspaceRequest{ID: "test-id"}
	body, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/workspaces.delete", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.AuthUserKey, &middleware.AuthenticatedUser{ID: "test-user-id"}))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}
