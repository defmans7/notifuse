package http

import (
	"context"
	"encoding/json"
	"net/http"
	"notifuse/server/internal/http/middleware"
	"notifuse/server/internal/service"

	"aidanwoods.dev/go-paseto"
	"github.com/go-chi/chi/v5"
)

// WorkspaceServiceInterface defines the interface for workspace operations
type WorkspaceServiceInterface interface {
	CreateWorkspace(ctx context.Context, name, websiteURL, logoURL, timezone, ownerID string) (*service.Workspace, error)
	GetWorkspace(ctx context.Context, id, ownerID string) (*service.Workspace, error)
	ListWorkspaces(ctx context.Context, ownerID string) ([]service.Workspace, error)
	UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, timezone, ownerID string) (*service.Workspace, error)
	DeleteWorkspace(ctx context.Context, id, ownerID string) error
}

type WorkspaceHandler struct {
	workspaceService WorkspaceServiceInterface
	authService      middleware.AuthServiceInterface
}

func NewWorkspaceHandler(workspaceService WorkspaceServiceInterface, authService middleware.AuthServiceInterface) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
		authService:      authService,
	}
}

// Request/Response types
type createWorkspaceRequest struct {
	Name       string `json:"name"`
	WebsiteURL string `json:"website_url"`
	LogoURL    string `json:"logo_url"`
	Timezone   string `json:"timezone"`
}

type getWorkspaceRequest struct {
	ID string `json:"id"`
}

type updateWorkspaceRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	WebsiteURL string `json:"website_url"`
	LogoURL    string `json:"logo_url"`
	Timezone   string `json:"timezone"`
}

type deleteWorkspaceRequest struct {
	ID string `json:"id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// RegisterRoutes registers all workspace RPC-style routes with authentication middleware
func (h *WorkspaceHandler) RegisterRoutes(r chi.Router, publicKey paseto.V4AsymmetricPublicKey) {
	// Create a subrouter for workspace routes
	workspaceRouter := chi.NewRouter()

	// Add authentication middlewares
	workspaceRouter.Use(middleware.NewAuthMiddleware(publicKey).VerifyToken)
	workspaceRouter.Use(middleware.RequireAuth(h.authService))

	// Register RPC-style endpoints with dot notation
	workspaceRouter.Get("/workspaces.list", h.handleList) // GET for listing (no request body needed)
	workspaceRouter.Get("/workspaces.get", h.handleGet)   // GET with query params
	workspaceRouter.Post("/workspaces.create", h.handleCreate)
	workspaceRouter.Post("/workspaces.update", h.handleUpdate)
	workspaceRouter.Post("/workspaces.delete", h.handleDelete)

	// Mount the router at /api
	r.Mount("/api", workspaceRouter)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func (h *WorkspaceHandler) handleList(w http.ResponseWriter, r *http.Request) {
	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	workspaces, err := h.workspaceService.ListWorkspaces(r.Context(), authUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list workspaces")
		return
	}

	writeJSON(w, http.StatusOK, workspaces)
}

func (h *WorkspaceHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	// Get workspace ID from query params
	workspaceID := r.URL.Query().Get("id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "Missing workspace ID")
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(r.Context(), workspaceID, authUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get workspace")
		return
	}
	if workspace == nil {
		writeError(w, http.StatusNotFound, "Workspace not found")
		return
	}

	writeJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(r.Context(), req.Name, req.WebsiteURL, req.LogoURL, req.Timezone, authUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create workspace")
		return
	}

	writeJSON(w, http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	var req updateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspace(r.Context(), req.ID, req.Name, req.WebsiteURL, req.LogoURL, req.Timezone, authUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update workspace")
		return
	}
	if workspace == nil {
		writeError(w, http.StatusNotFound, "Workspace not found")
		return
	}

	writeJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	var req deleteWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.workspaceService.DeleteWorkspace(r.Context(), req.ID, authUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete workspace")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
