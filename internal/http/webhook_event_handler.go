package http

import (
	"io"
	"net/http"
	"strconv"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// WebhookEventHandler handles HTTP requests for webhook events
type WebhookEventHandler struct {
	service   domain.WebhookEventServiceInterface
	logger    logger.Logger
	publicKey paseto.V4AsymmetricPublicKey
}

// NewWebhookEventHandler creates a new webhook event handler
func NewWebhookEventHandler(service domain.WebhookEventServiceInterface, publicKey paseto.V4AsymmetricPublicKey, logger logger.Logger) *WebhookEventHandler {
	return &WebhookEventHandler{
		service:   service,
		logger:    logger,
		publicKey: publicKey,
	}
}

// RegisterRoutes registers the webhook event HTTP endpoints
func (h *WebhookEventHandler) RegisterRoutes(mux *http.ServeMux) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
	requireAuth := authMiddleware.RequireAuth()

	// Public webhooks endpoint for receiving events from email providers
	mux.Handle("/webhooks/email", http.HandlerFunc(h.handleIncomingWebhook))

	// Authenticated endpoints for accessing webhook event data
	mux.Handle("/api/webhookEvents.list", requireAuth(http.HandlerFunc(h.handleList)))
	mux.Handle("/api/webhookEvents.get", requireAuth(http.HandlerFunc(h.handleGet)))
	mux.Handle("/api/webhookEvents.getByMessageID", requireAuth(http.HandlerFunc(h.handleGetByMessageID)))
	mux.Handle("/api/webhookEvents.getByTransactionalID", requireAuth(http.HandlerFunc(h.handleGetByTransactionalID)))
	mux.Handle("/api/webhookEvents.getByBroadcastID", requireAuth(http.HandlerFunc(h.handleGetByBroadcastID)))
}

// handleIncomingWebhook handles incoming webhook events from email providers
func (h *WebhookEventHandler) handleIncomingWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract provider, workspace_id and integration_id from query parameters
	// Format: /webhooks/email?provider={provider}&workspace_id={id}&integration_id={id}
	provider := r.URL.Query().Get("provider")
	workspaceID := r.URL.Query().Get("workspace_id")
	integrationID := r.URL.Query().Get("integration_id")

	if provider == "" {
		WriteJSONError(w, "Provider is required", http.StatusBadRequest)
		return
	}

	if workspaceID == "" || integrationID == "" {
		WriteJSONError(w, "Workspace ID and integration ID are required", http.StatusBadRequest)
		return
	}

	// Log the incoming webhook
	h.logger.WithField("provider", provider).
		WithField("workspace_id", workspaceID).
		WithField("integration_id", integrationID).
		Info("Received webhook event")

	// Read and parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to read webhook request body")
		WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Process the webhook event
	err = h.service.ProcessWebhook(r.Context(), workspaceID, integrationID, body)
	if err != nil {
		h.logger.WithField("error", err.Error()).
			WithField("workspace_id", workspaceID).
			WithField("integration_id", integrationID).
			WithField("provider", provider).
			Error("Failed to process webhook")
		WriteJSONError(w, "Failed to process webhook", http.StatusBadRequest)
		return
	}

	// Return success
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// handleList handles requests to list webhook events by type
func (h *WebhookEventHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request parameters
	workspaceID := r.URL.Query().Get("workspace_id")
	eventTypeStr := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Create and validate request
	req := domain.GetEventsRequest{
		WorkspaceID: workspaceID,
		Type:        domain.EmailEventType(eventTypeStr),
	}

	// Parse limit and offset
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			WriteJSONError(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		req.Limit = limit
	}
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			WriteJSONError(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		req.Offset = offset
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get events
	events, err := h.service.GetEventsByType(r.Context(), req.WorkspaceID, req.Type, req.Limit, req.Offset)
	if err != nil {
		h.logger.WithField("error", err.Error()).
			WithField("workspace_id", req.WorkspaceID).
			Error("Failed to get webhook events")
		WriteJSONError(w, "Failed to get webhook events", http.StatusInternalServerError)
		return
	}

	// Get total count
	count, err := h.service.GetEventCount(r.Context(), req.WorkspaceID, req.Type)
	if err != nil {
		h.logger.WithField("error", err.Error()).
			WithField("workspace_id", req.WorkspaceID).
			Error("Failed to get webhook event count")
		WriteJSONError(w, "Failed to get webhook event count", http.StatusInternalServerError)
		return
	}

	// Return results
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  count,
	})
}

// handleGet handles requests to get a specific webhook event by ID
func (h *WebhookEventHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request parameters
	eventID := r.URL.Query().Get("id")
	if eventID == "" {
		WriteJSONError(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Get event
	event, err := h.service.GetEventByID(r.Context(), eventID)
	if err != nil {
		if _, ok := err.(*domain.ErrWebhookEventNotFound); ok {
			WriteJSONError(w, "Webhook event not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).
			WithField("event_id", eventID).
			Error("Failed to get webhook event")
		WriteJSONError(w, "Failed to get webhook event", http.StatusInternalServerError)
		return
	}

	// Return results
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"event": event,
	})
}

// handleGetByMessageID handles requests to get webhook events by message ID
func (h *WebhookEventHandler) handleGetByMessageID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request parameters
	messageID := r.URL.Query().Get("message_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Create request
	req := domain.GetEventsByMessageIDRequest{
		MessageID: messageID,
	}

	// Parse limit and offset
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			WriteJSONError(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		req.Limit = limit
	}
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			WriteJSONError(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		req.Offset = offset
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get events
	events, err := h.service.GetEventsByMessageID(r.Context(), req.MessageID, req.Limit, req.Offset)
	if err != nil {
		h.logger.WithField("error", err.Error()).
			WithField("message_id", req.MessageID).
			Error("Failed to get webhook events by message ID")
		WriteJSONError(w, "Failed to get webhook events", http.StatusInternalServerError)
		return
	}

	// Return results
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// handleGetByTransactionalID handles requests to get webhook events by transactional ID
func (h *WebhookEventHandler) handleGetByTransactionalID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request parameters
	workspaceID := r.URL.Query().Get("workspace_id")
	transactionalID := r.URL.Query().Get("transactional_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Create request
	req := domain.GetEventsByTransactionalIDRequest{
		WorkspaceID:     workspaceID,
		TransactionalID: transactionalID,
	}

	// Parse limit and offset
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			WriteJSONError(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		req.Limit = limit
	}
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			WriteJSONError(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		req.Offset = offset
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get events
	events, err := h.service.GetEventsByTransactionalID(r.Context(), req.TransactionalID, req.Limit, req.Offset)
	if err != nil {
		h.logger.WithField("error", err.Error()).
			WithField("transactional_id", req.TransactionalID).
			Error("Failed to get webhook events by transactional ID")
		WriteJSONError(w, "Failed to get webhook events", http.StatusInternalServerError)
		return
	}

	// Return results
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// handleGetByBroadcastID handles requests to get webhook events by broadcast ID
func (h *WebhookEventHandler) handleGetByBroadcastID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request parameters
	workspaceID := r.URL.Query().Get("workspace_id")
	broadcastID := r.URL.Query().Get("broadcast_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Create request
	req := domain.GetEventsByBroadcastIDRequest{
		WorkspaceID: workspaceID,
		BroadcastID: broadcastID,
	}

	// Parse limit and offset
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			WriteJSONError(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		req.Limit = limit
	}
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			WriteJSONError(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		req.Offset = offset
	}

	// Validate request
	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get events
	events, err := h.service.GetEventsByBroadcastID(r.Context(), req.BroadcastID, req.Limit, req.Offset)
	if err != nil {
		h.logger.WithField("error", err.Error()).
			WithField("broadcast_id", req.BroadcastID).
			Error("Failed to get webhook events by broadcast ID")
		WriteJSONError(w, "Failed to get webhook events", http.StatusInternalServerError)
		return
	}

	// Return results
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// splitPath splits a URL path into its components
func splitPath(path string) []string {
	var parts []string
	var current string
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
