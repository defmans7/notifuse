package http

import (
	"encoding/json"
	"net/http"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ListHandler struct {
	service domain.ListService
	logger  logger.Logger
}

func NewListHandler(service domain.ListService, logger logger.Logger) *ListHandler {
	return &ListHandler{
		service: service,
		logger:  logger,
	}
}

// Request/Response types
type createListRequest struct {
	ID            string `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name          string `json:"name" valid:"required,stringlength(1|255)"`
	Type          string `json:"type" valid:"required,in(public|private)"`
	IsDoubleOptin bool   `json:"is_double_optin"`
	Description   string `json:"description,omitempty"`
}

type getListRequest struct {
	ID string `json:"id"`
}

type updateListRequest struct {
	ID            string `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name          string `json:"name" valid:"required,stringlength(1|255)"`
	Type          string `json:"type" valid:"required,in(public|private)"`
	IsDoubleOptin bool   `json:"is_double_optin"`
	Description   string `json:"description,omitempty"`
}

type deleteListRequest struct {
	ID string `json:"id"`
}

func (h *ListHandler) RegisterRoutes(mux *http.ServeMux) {
	// Register RPC-style endpoints with dot notation
	mux.HandleFunc("/api/lists.list", h.handleList)
	mux.HandleFunc("/api/lists.get", h.handleGet)
	mux.HandleFunc("/api/lists.create", h.handleCreate)
	mux.HandleFunc("/api/lists.update", h.handleUpdate)
	mux.HandleFunc("/api/lists.delete", h.handleDelete)
}

func (h *ListHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lists, err := h.service.GetLists(r.Context())
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to get lists")
		WriteJSONError(w, "Failed to get lists", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, lists)
}

func (h *ListHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get list ID from query params
	listID := r.URL.Query().Get("id")
	if listID == "" {
		WriteJSONError(w, "Missing list ID", http.StatusBadRequest)
		return
	}

	list, err := h.service.GetListByID(r.Context(), listID)
	if err != nil {
		if _, ok := err.(*domain.ErrListNotFound); ok {
			WriteJSONError(w, "List not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get list")
		WriteJSONError(w, "Failed to get list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"list": list,
	})
}

func (h *ListHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	list := &domain.List{
		ID:            req.ID,
		Name:          req.Name,
		Type:          req.Type,
		IsDoubleOptin: req.IsDoubleOptin,
		Description:   req.Description,
	}

	if err := h.service.CreateList(r.Context(), list); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to create list")
		WriteJSONError(w, "Failed to create list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"list": list,
	})
}

func (h *ListHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req updateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		WriteJSONError(w, "Missing ID", http.StatusBadRequest)
		return
	}

	list := &domain.List{
		ID:            req.ID,
		Name:          req.Name,
		Type:          req.Type,
		IsDoubleOptin: req.IsDoubleOptin,
		Description:   req.Description,
	}

	if err := h.service.UpdateList(r.Context(), list); err != nil {
		if _, ok := err.(*domain.ErrListNotFound); ok {
			WriteJSONError(w, "List not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to update list")
		WriteJSONError(w, "Failed to update list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"list": list,
	})
}

func (h *ListHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deleteListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		WriteJSONError(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteList(r.Context(), req.ID); err != nil {
		if _, ok := err.(*domain.ErrListNotFound); ok {
			WriteJSONError(w, "List not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to delete list")
		WriteJSONError(w, "Failed to delete list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
