package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type NotificationCenterHandler struct {
	service domain.NotificationCenterService
	logger  logger.Logger
}

func NewNotificationCenterHandler(service domain.NotificationCenterService, logger logger.Logger) *NotificationCenterHandler {
	return &NotificationCenterHandler{
		service: service,
		logger:  logger,
	}
}

func (h *NotificationCenterHandler) RegisterRoutes(mux *http.ServeMux) {
	// Register public routes
	mux.HandleFunc("/notification-center", h.handleNotificationCenter)
	mux.HandleFunc("/notification-center/unsubscribe", h.handleUnsubscribe)
	mux.HandleFunc("/notification-center/subscribe", h.handleSubscribe)
}

func (h *NotificationCenterHandler) handleNotificationCenter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.NotificationCenterRequest
	if err := req.FromURLValues(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get notification center data for the contact
	response, err := h.service.GetNotificationCenter(r.Context(), req.WorkspaceID, req.Email, req.EmailHMAC)
	if err != nil {
		if strings.Contains(err.Error(), "invalid email verification") {
			WriteJSONError(w, "Unauthorized: invalid verification", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "contact not found") {
			WriteJSONError(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get notification center data")
		WriteJSONError(w, "Failed to get notification center data", http.StatusInternalServerError)
		return
	}

	// Write the response
	writeJSON(w, http.StatusOK, response)
}

// handleUnsubscribe handles unsubscribing a contact from a list
func (h *NotificationCenterHandler) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.UnsubscribeFromListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unsubscribe the contact from the list
	err := h.service.UnsubscribeFromList(r.Context(), req.WorkspaceID, req.Email, req.EmailHMAC, req.ListID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid email verification") {
			WriteJSONError(w, "Unauthorized: invalid verification", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "contact not found") {
			WriteJSONError(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to unsubscribe from list")
		WriteJSONError(w, "Failed to unsubscribe from list", http.StatusInternalServerError)
		return
	}

	// Write response
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// handleSubscribe handles subscribing a contact to a list
func (h *NotificationCenterHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.SubscribeToListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Subscribe the contact to the list
	err := h.service.SubscribeToList(r.Context(), req.WorkspaceID, req.Email, req.ListID, req.EmailHMAC)
	if err != nil {
		if strings.Contains(err.Error(), "invalid email verification") {
			WriteJSONError(w, "Unauthorized: invalid verification", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "list is not public") {
			WriteJSONError(w, "List is not public", http.StatusForbidden)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to subscribe to list")
		WriteJSONError(w, "Failed to subscribe to list", http.StatusInternalServerError)
		return
	}

	// Write response
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
