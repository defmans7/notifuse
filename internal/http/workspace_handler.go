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
	authService      domain.AuthService
	publicKey        paseto.V4AsymmetricPublicKey
	logger           logger.Logger
	secretKey        string
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(
	workspaceService domain.WorkspaceServiceInterface,
	authService domain.AuthService,
	publicKey paseto.V4AsymmetricPublicKey,
	logger logger.Logger,
	secretKey string,
) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
		authService:      authService,
		publicKey:        publicKey,
		logger:           logger,
		secretKey:        secretKey,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

// RegisterRoutes registers all workspace RPC-style routes with authentication middleware
func (h *WorkspaceHandler) RegisterRoutes(mux *http.ServeMux) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
	requireAuth := authMiddleware.RequireAuth()

	// Register RPC-style endpoints with dot notation
	mux.Handle("/api/workspaces.list", requireAuth(http.HandlerFunc(h.handleList)))
	mux.Handle("/api/workspaces.get", requireAuth(http.HandlerFunc(h.handleGet)))
	mux.Handle("/api/workspaces.create", requireAuth(http.HandlerFunc(h.handleCreate)))
	mux.Handle("/api/workspaces.update", requireAuth(http.HandlerFunc(h.handleUpdate)))
	mux.Handle("/api/workspaces.delete", requireAuth(http.HandlerFunc(h.handleDelete)))
	mux.Handle("/api/workspaces.members", requireAuth(http.HandlerFunc(h.handleMembers)))
	mux.Handle("/api/workspaces.inviteMember", requireAuth(http.HandlerFunc(h.handleInviteMember)))
	mux.Handle("/api/workspaces.createAPIKey", requireAuth(http.HandlerFunc(h.handleCreateAPIKey)))
}

func (h *WorkspaceHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	workspaces, err := h.workspaceService.ListWorkspaces(r.Context())
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

	// Get workspace ID from query params
	workspaceID := r.URL.Query().Get("id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "Missing workspace ID")
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(r.Context(), workspaceID)
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

	var req domain.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(h.secretKey); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(
		r.Context(),
		req.ID,
		req.Name,
		req.Settings.WebsiteURL,
		req.Settings.LogoURL,
		req.Settings.CoverURL,
		req.Settings.Timezone,
		req.Settings.FileManager,
	)
	if err != nil {
		if err.Error() == "workspace already exists" {
			writeError(w, http.StatusConflict, "Workspace already exists")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to create workspace")
		}
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

	var req domain.UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(h.secretKey); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspace(
		r.Context(),
		req.ID,
		req.Name,
		req.Settings,
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

	var req domain.DeleteWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.workspaceService.DeleteWorkspace(r.Context(), req.ID)
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

	// Get workspace ID from query params
	workspaceID := r.URL.Query().Get("id")
	if workspaceID == "" {
		writeError(w, http.StatusBadRequest, "Missing workspace ID")
		return
	}

	// Use the new method that includes emails
	members, err := h.workspaceService.GetWorkspaceMembersWithEmail(r.Context(), workspaceID)
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

	var req domain.InviteMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create the invitation or add the user directly if they already exist
	invitation, token, err := h.workspaceService.InviteMember(r.Context(), req.WorkspaceID, req.Email)
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

// handleCreateAPIKey handles the request to create an API key for a workspace
func (h *WorkspaceHandler) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req domain.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Use the workspace service to create the API key
	token, apiEmail, err := h.workspaceService.CreateAPIKey(r.Context(), req.WorkspaceID, req.EmailPrefix)
	if err != nil {
		h.logger.WithField("workspace_id", req.WorkspaceID).WithField("error", err.Error()).Error("Failed to create API key")

		// Check if it's an authorization error
		if _, ok := err.(*domain.ErrUnauthorized); ok {
			writeError(w, http.StatusForbidden, "Only workspace owners can create API keys")
			return
		}

		writeError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	// Return the token and API details
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"token":  token,
		"email":  apiEmail,
	})
}
