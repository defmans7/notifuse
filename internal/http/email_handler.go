package http

import (
	"encoding/json"
	"net/http"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/pkg/logger"

	"aidanwoods.dev/go-paseto"
)

// EmailHandler handles HTTP requests for email operations
type EmailHandler struct {
	emailService domain.EmailServiceInterface
	publicKey    paseto.V4AsymmetricPublicKey
	logger       logger.Logger
	secretKey    string
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(
	emailService domain.EmailServiceInterface,
	publicKey paseto.V4AsymmetricPublicKey,
	logger logger.Logger,
	secretKey string,
) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
		publicKey:    publicKey,
		logger:       logger,
		secretKey:    secretKey,
	}
}

// RegisterRoutes registers all workspace RPC-style routes with authentication middleware
func (h *EmailHandler) RegisterRoutes(mux *http.ServeMux) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
	requireAuth := authMiddleware.RequireAuth()

	// Register RPC-style endpoints with dot notation
	mux.Handle("/api/email.testProvider", requireAuth(http.HandlerFunc(h.handleTestEmailProvider)))
	mux.Handle("/api/email.testTemplate", requireAuth(http.HandlerFunc(h.handleTestTemplate)))
}

// Add the handler for testEmailProvider
func (h *EmailHandler) handleTestEmailProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req domain.TestEmailProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.To == "" {
		writeError(w, http.StatusBadRequest, "Missing recipient email (to)")
		return
	}

	if req.WorkspaceID == "" {
		writeError(w, http.StatusBadRequest, "Missing workspace ID")
		return
	}

	err := h.emailService.TestEmailProvider(r.Context(), req.WorkspaceID, req.Provider, req.To)
	resp := domain.TestEmailProviderResponse{Success: err == nil}
	if err != nil {
		resp.Error = err.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleTestTemplate handles requests to test a template
func (h *EmailHandler) handleTestTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req domain.TestTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	workspaceID, templateID, integrationID, recipientEmail, err := req.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.emailService.TestTemplate(r.Context(), workspaceID, templateID, integrationID, recipientEmail)

	// Create response
	response := domain.TestTemplateResponse{
		Success: err == nil,
	}

	// If there's an error, include it in the response
	if err != nil {
		if _, ok := err.(*domain.ErrTemplateNotFound); ok {
			writeError(w, http.StatusNotFound, "Template not found")
			return
		}

		h.logger.WithFields(map[string]interface{}{
			"error":        err.Error(),
			"workspace_id": workspaceID,
			"template_id":  templateID,
		}).Error("Failed to test template")

		response.Error = err.Error()
	}

	writeJSON(w, http.StatusOK, response)
}
