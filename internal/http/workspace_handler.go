package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/pkg/logger"

	"aidanwoods.dev/go-paseto"
)

// WorkspaceHandler handles HTTP requests for workspace operations
type WorkspaceHandler struct {
	workspaceService domain.WorkspaceServiceInterface
	authService      middleware.AuthServiceInterface
	publicKey        paseto.V4AsymmetricPublicKey
	logger           logger.Logger
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(
	workspaceService domain.WorkspaceServiceInterface,
	authService middleware.AuthServiceInterface,
	publicKey paseto.V4AsymmetricPublicKey,
	logger logger.Logger,
) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
		authService:      authService,
		publicKey:        publicKey,
		logger:           logger,
	}
}

// Request/Response types
type createWorkspaceRequest struct {
	ID       string                `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name     string                `json:"name" valid:"required,stringlength(1|32)"`
	Settings workspaceSettingsData `json:"settings"`
}

type workspaceSettingsData struct {
	Name       string `json:"name"`
	WebsiteURL string `json:"website_url"`
	LogoURL    string `json:"logo_url"`
	CoverURL   string `json:"cover_url"`
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
	CoverURL   string `json:"cover_url"`
	Timezone   string `json:"timezone"`
}

type deleteWorkspaceRequest struct {
	ID string `json:"id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

// RegisterRoutes registers all workspace RPC-style routes with authentication middleware
func (h *WorkspaceHandler) RegisterRoutes(mux *http.ServeMux) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
	requireAuth := authMiddleware.RequireAuth(h.authService)

	// Register RPC-style endpoints with dot notation
	mux.Handle("/api/workspaces.list", requireAuth(http.HandlerFunc(h.handleList)))
	mux.Handle("/api/workspaces.get", requireAuth(http.HandlerFunc(h.handleGet)))
	mux.Handle("/api/workspaces.create", requireAuth(http.HandlerFunc(h.handleCreate)))
	mux.Handle("/api/workspaces.update", requireAuth(http.HandlerFunc(h.handleUpdate)))
	mux.Handle("/api/workspaces.delete", requireAuth(http.HandlerFunc(h.handleDelete)))
	mux.Handle("/api/workspaces.members", requireAuth(http.HandlerFunc(h.handleMembers)))
	mux.Handle("/api/workspaces.inviteMember", requireAuth(http.HandlerFunc(h.handleInviteMember)))
}

func (h *WorkspaceHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	workspaces, err := h.workspaceService.ListWorkspaces(r.Context(), authUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list workspaces")
		return
	}

	writeJSON(w, http.StatusOK, workspaces)
}

func (h *WorkspaceHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

	// Wrap the workspace in a response object with a workspace field to match frontend expectations
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"workspace": workspace,
	})
}

func (h *WorkspaceHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate workspace ID first
	if req.ID == "" {
		writeError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	// Support name from either root or settings
	name := req.Name
	if name == "" {
		name = req.Settings.Name
	}

	// Validate name
	if name == "" {
		writeError(w, http.StatusBadRequest, "Workspace name is required")
		return
	}

	// Validate timezone
	if req.Settings.Timezone == "" {
		writeError(w, http.StatusBadRequest, "Timezone is required")
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(
		r.Context(),
		req.ID,
		name,
		req.Settings.WebsiteURL,
		req.Settings.LogoURL,
		req.Settings.CoverURL,
		req.Settings.Timezone,
		authUser.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create workspace")
		return
	}

	writeJSON(w, http.StatusCreated, workspace)
}

// Helper function to get bytes from request body
func getBytesFromBody(body io.ReadCloser) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	return buf.Bytes()
}

func (h *WorkspaceHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	var req updateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspace(
		r.Context(),
		req.ID,
		req.Name,
		req.WebsiteURL,
		req.LogoURL,
		req.CoverURL,
		req.Timezone,
		authUser.ID,
	)
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
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

// handleMembers handles the request to get members of a workspace
func (h *WorkspaceHandler) handleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	// Get workspace ID from query params
	workspaceID := r.URL.Query().Get("id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "Missing workspace ID")
		return
	}

	// Use the new method that includes emails
	members, err := h.workspaceService.GetWorkspaceMembersWithEmail(r.Context(), workspaceID, authUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get workspace members")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"members": members,
	})
}

// handleInviteMember handles the request to invite a member to a workspace
func (h *WorkspaceHandler) handleInviteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the authenticated user from the context
	authUser := r.Context().Value(middleware.AuthUserKey).(*middleware.AuthenticatedUser)

	var req struct {
		WorkspaceID string `json:"workspace_id"`
		Email       string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Create the invitation or add the user directly if they already exist
	invitation, token, err := h.workspaceService.InviteMember(r.Context(), req.WorkspaceID, authUser.ID, req.Email)
	if err != nil {
		h.logger.WithField("workspace_id", req.WorkspaceID).WithField("email", req.Email).WithField("error", err.Error()).Error("Failed to invite member")
		writeError(w, http.StatusInternalServerError, "Failed to invite member")
		return
	}

	// If invitation is nil, it means the user was directly added to the workspace
	if invitation == nil {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": "User added to workspace",
		})
		return
	}

	// Return the invitation details and token
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "success",
		"message":    "Invitation sent",
		"invitation": invitation,
		"token":      token,
	})
}
