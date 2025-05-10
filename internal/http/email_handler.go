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
	mux.Handle("/visit", requireAuth(http.HandlerFunc(h.handleClickRedirection)))
	mux.Handle("/api/email.testProvider", requireAuth(http.HandlerFunc(h.handleTestEmailProvider)))
	mux.Handle("/api/email.testTemplate", requireAuth(http.HandlerFunc(h.handleTestTemplate)))
}

// Add the handler for testEmailProvider
func (h *EmailHandler) handleTestEmailProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.TestEmailProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.To == "" {
		WriteJSONError(w, "Missing recipient email (to)", http.StatusBadRequest)
		return
	}

	if req.WorkspaceID == "" {
		WriteJSONError(w, "Missing workspace ID", http.StatusBadRequest)
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
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.TestTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	workspaceID, templateID, integrationID, recipientEmail, cc, bcc, replyTo, err := req.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.emailService.TestTemplate(r.Context(), workspaceID, templateID, integrationID, recipientEmail, cc, bcc, replyTo)

	// Create response
	response := domain.TestTemplateResponse{
		Success: err == nil,
	}

	// If there's an error, include it in the response
	if err != nil {
		if _, ok := err.(*domain.ErrTemplateNotFound); ok {
			WriteJSONError(w, "Template not found", http.StatusNotFound)
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

func (h *EmailHandler) handleClickRedirection(w http.ResponseWriter, r *http.Request) {
	// Get the message id (mid) and workspace id (wid) from the query parameters
	messageID := r.URL.Query().Get("mid")
	workspaceID := r.URL.Query().Get("wid")
	redirectTo := r.URL.Query().Get("url")

	// Check if URL is provided, show error if missing
	if redirectTo == "" {
		http.Error(w, "Missing redirect URL", http.StatusBadRequest)
		return
	}

	// redirect to the url if mid and wid are present
	if messageID == "" || workspaceID == "" {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	// increment the click count
	h.emailService.VisitLink(r.Context(), messageID, workspaceID)

	// redirect to the url
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
