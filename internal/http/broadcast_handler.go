package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// MissingParameterError is an error type for missing URL parameters
type MissingParameterError struct {
	Param string
}

// Error returns the error message
func (e *MissingParameterError) Error() string {
	return fmt.Sprintf("Missing parameter: %s", e.Param)
}

type BroadcastHandler struct {
	service   domain.BroadcastService
	logger    logger.Logger
	publicKey paseto.V4AsymmetricPublicKey
}

func NewBroadcastHandler(service domain.BroadcastService, publicKey paseto.V4AsymmetricPublicKey, logger logger.Logger) *BroadcastHandler {
	return &BroadcastHandler{
		service:   service,
		logger:    logger,
		publicKey: publicKey,
	}
}

func (h *BroadcastHandler) RegisterRoutes(mux *http.ServeMux) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
	requireAuth := authMiddleware.RequireAuth()

	// Register RPC-style endpoints with dot notation
	mux.Handle("/api/broadcasts.list", requireAuth(http.HandlerFunc(h.handleList)))
	mux.Handle("/api/broadcasts.get", requireAuth(http.HandlerFunc(h.handleGet)))
	mux.Handle("/api/broadcasts.create", requireAuth(http.HandlerFunc(h.handleCreate)))
	mux.Handle("/api/broadcasts.update", requireAuth(http.HandlerFunc(h.handleUpdate)))
	mux.Handle("/api/broadcasts.schedule", requireAuth(http.HandlerFunc(h.handleSchedule)))
	mux.Handle("/api/broadcasts.pause", requireAuth(http.HandlerFunc(h.handlePause)))
	mux.Handle("/api/broadcasts.resume", requireAuth(http.HandlerFunc(h.handleResume)))
	mux.Handle("/api/broadcasts.cancel", requireAuth(http.HandlerFunc(h.handleCancel)))
	mux.Handle("/api/broadcasts.send", requireAuth(http.HandlerFunc(h.handleSend)))
	mux.Handle("/api/broadcasts.sendToIndividual", requireAuth(http.HandlerFunc(h.handleSendToIndividual)))
}

// GetBroadcastsRequest is used to extract query parameters for listing broadcasts
type GetBroadcastsRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Status      string `json:"status,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// FromURLParams parses URL query parameters into the request
func (r *GetBroadcastsRequest) FromURLParams(values url.Values) error {
	r.WorkspaceID = values.Get("workspace_id")
	if r.WorkspaceID == "" {
		return &MissingParameterError{Param: "workspace_id"}
	}

	r.Status = values.Get("status")

	if limitStr := values.Get("limit"); limitStr != "" {
		var err error
		r.Limit, err = parseIntParam(limitStr)
		if err != nil {
			return fmt.Errorf("invalid limit parameter: %w", err)
		}
	}

	if offsetStr := values.Get("offset"); offsetStr != "" {
		var err error
		r.Offset, err = parseIntParam(offsetStr)
		if err != nil {
			return fmt.Errorf("invalid offset parameter: %w", err)
		}
	}

	return nil
}

// parseIntParam parses a string to an integer
func parseIntParam(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// GetBroadcastRequest is used to extract query parameters for getting a single broadcast
type GetBroadcastRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

// FromURLParams parses URL query parameters into the request
func (r *GetBroadcastRequest) FromURLParams(values url.Values) error {
	r.WorkspaceID = values.Get("workspace_id")
	if r.WorkspaceID == "" {
		return &MissingParameterError{Param: "workspace_id"}
	}

	r.ID = values.Get("id")
	if r.ID == "" {
		return &MissingParameterError{Param: "id"}
	}

	return nil
}

func (h *BroadcastHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GetBroadcastsRequest
	if err := req.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := domain.ListBroadcastsParams{
		WorkspaceID: req.WorkspaceID,
		Status:      domain.BroadcastStatus(req.Status),
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	response, err := h.service.ListBroadcasts(r.Context(), params)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to list broadcasts")
		WriteJSONError(w, "Failed to list broadcasts", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"broadcasts":  response.Broadcasts,
		"total_count": response.TotalCount,
	})
}

// HandleList is an exported version of handleList for testing purposes
func (h *BroadcastHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	h.handleList(w, r)
}

// HandleGet is an exported version of handleGet for testing purposes
func (h *BroadcastHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	h.handleGet(w, r)
}

func (h *BroadcastHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GetBroadcastRequest
	if err := req.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	broadcast, err := h.service.GetBroadcast(r.Context(), req.WorkspaceID, req.ID)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get broadcast")
		WriteJSONError(w, "Failed to get broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"broadcast": broadcast,
	})
}

// HandleCreate is an exported version of handleCreate for testing purposes
func (h *BroadcastHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	h.handleCreate(w, r)
}

func (h *BroadcastHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreateBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := req.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	broadcast, err := h.service.CreateBroadcast(r.Context(), &req)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to create broadcast")
		WriteJSONError(w, "Failed to create broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"broadcast": broadcast,
	})
}

func (h *BroadcastHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.UpdateBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the existing broadcast to pass to Validate
	existingBroadcast, err := h.service.GetBroadcast(r.Context(), req.WorkspaceID, req.ID)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get existing broadcast")
		WriteJSONError(w, "Failed to get existing broadcast", http.StatusInternalServerError)
		return
	}

	_, err = req.Validate(existingBroadcast)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedBroadcast, err := h.service.UpdateBroadcast(r.Context(), &req)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to update broadcast")
		WriteJSONError(w, "Failed to update broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"broadcast": updatedBroadcast,
	})
}

// HandleSchedule is an exported version of handleSchedule for testing purposes
func (h *BroadcastHandler) HandleSchedule(w http.ResponseWriter, r *http.Request) {
	h.handleSchedule(w, r)
}

func (h *BroadcastHandler) handleSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.ScheduleBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.ScheduleBroadcast(r.Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to schedule broadcast")
		WriteJSONError(w, "Failed to schedule broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// HandlePause is an exported version of handlePause for testing purposes
func (h *BroadcastHandler) HandlePause(w http.ResponseWriter, r *http.Request) {
	h.handlePause(w, r)
}

func (h *BroadcastHandler) handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.PauseBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.PauseBroadcast(r.Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to pause broadcast")
		WriteJSONError(w, "Failed to pause broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *BroadcastHandler) handleResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.ResumeBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.ResumeBroadcast(r.Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to resume broadcast")
		WriteJSONError(w, "Failed to resume broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// HandleCancel is an exported version of handleCancel for testing purposes
func (h *BroadcastHandler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	h.handleCancel(w, r)
}

func (h *BroadcastHandler) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CancelBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.CancelBroadcast(r.Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to cancel broadcast")
		WriteJSONError(w, "Failed to cancel broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *BroadcastHandler) handleSendToIndividual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.SendToIndividualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.SendToIndividual(r.Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to send broadcast to individual")
		WriteJSONError(w, "Failed to send broadcast to individual", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// HandleSend is an exported version of handleSend for testing purposes
func (h *BroadcastHandler) HandleSend(w http.ResponseWriter, r *http.Request) {
	h.handleSend(w, r)
}

func (h *BroadcastHandler) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.SendBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.SendBroadcast(r.Context(), &req)
	if err != nil {
		if _, ok := err.(*domain.ErrBroadcastNotFound); ok {
			WriteJSONError(w, "Broadcast not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to send broadcast")
		WriteJSONError(w, "Failed to send broadcast", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
