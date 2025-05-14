package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type NotificationCenterHandler struct {
	service     domain.NotificationCenterService
	listService domain.ListService
	logger      logger.Logger
}

func NewNotificationCenterHandler(service domain.NotificationCenterService, listService domain.ListService, logger logger.Logger) *NotificationCenterHandler {
	return &NotificationCenterHandler{
		service:     service,
		listService: listService,
		logger:      logger,
	}
}

func (h *NotificationCenterHandler) RegisterRoutes(mux *http.ServeMux) {
	// Register public routes
	mux.HandleFunc("/notification-center", h.handleNotificationCenter)
	mux.HandleFunc("/subscribe", h.handleSubscribe)
	// one-click unsubscribe for GMAIL header link
	mux.HandleFunc("/unsubscribe-oneclick", h.handleUnsubscribeOneClick)
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

func (h *NotificationCenterHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.SubscribeToListsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to validate request")
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fromAPI := false

	if err := h.listService.SubscribeToLists(r.Context(), &req, fromAPI); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to subscribe to lists")
		WriteJSONError(w, "Failed to subscribe to lists", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *NotificationCenterHandler) handleUnsubscribeOneClick(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// this is one-click unsubscribe from GMAIL header link

	var req domain.UnsubscribeFromListsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fromBearerToken := false

	if err := h.listService.UnsubscribeFromLists(r.Context(), &req, fromBearerToken); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to unsubscribe from lists")
		WriteJSONError(w, "Failed to unsubscribe from lists", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
