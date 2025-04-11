package http

import (
	"encoding/json"
	"net/http"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/http/middleware"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ListHandler struct {
	service   domain.ListService
	logger    logger.Logger
	publicKey paseto.V4AsymmetricPublicKey
}

func NewListHandler(service domain.ListService, publicKey paseto.V4AsymmetricPublicKey, logger logger.Logger) *ListHandler {
	return &ListHandler{
		service:   service,
		logger:    logger,
		publicKey: publicKey,
	}
}

func (h *ListHandler) RegisterRoutes(mux *http.ServeMux) {

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.publicKey)
	requireAuth := authMiddleware.RequireAuth()

	// Register RPC-style endpoints with dot notation
	mux.Handle("/api/lists.list", requireAuth(http.HandlerFunc(h.handleList)))
	mux.Handle("/api/lists.get", requireAuth(http.HandlerFunc(h.handleGet)))
	mux.Handle("/api/lists.create", requireAuth(http.HandlerFunc(h.handleCreate)))
	mux.Handle("/api/lists.update", requireAuth(http.HandlerFunc(h.handleUpdate)))
	mux.Handle("/api/lists.delete", requireAuth(http.HandlerFunc(h.handleDelete)))
}

func (h *ListHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.GetListsRequest
	if err := req.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	lists, err := h.service.GetLists(r.Context(), req.WorkspaceID)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to get lists")
		WriteJSONError(w, "Failed to get lists", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"lists": lists,
	})
}

func (h *ListHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.GetListRequest
	if err := req.FromURLParams(r.URL.Query()); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := h.service.GetListByID(r.Context(), req.WorkspaceID, req.ID)
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

	var req domain.CreateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	list, workspaceID, err := req.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateList(r.Context(), workspaceID, list); err != nil {
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

	var req domain.UpdateListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	list, workspaceID, err := req.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateList(r.Context(), workspaceID, list); err != nil {
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

	var req domain.DeleteListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	workspaceID, err := req.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteList(r.Context(), workspaceID, req.ID); err != nil {
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
