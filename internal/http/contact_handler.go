package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ContactHandler struct {
	service domain.ContactService
	logger  logger.Logger
}

func NewContactHandler(service domain.ContactService, logger logger.Logger) *ContactHandler {
	return &ContactHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ContactHandler) RegisterRoutes(mux *http.ServeMux) {
	// Register RPC-style endpoints with dot notation
	mux.HandleFunc("/api/contacts.list", h.handleList)
	mux.HandleFunc("/api/contacts.getByEmail", h.handleGetByEmail)
	mux.HandleFunc("/api/contacts.getByExternalID", h.handleGetByExternalID)
	mux.HandleFunc("/api/contacts.delete", h.handleDelete)
	mux.HandleFunc("/api/contacts.import", h.handleImport)
	mux.HandleFunc("/api/contacts.upsert", h.handleUpsert)
}

func (h *ContactHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Convert to domain request
	domainReq := &domain.GetContactsRequest{}
	if err := domainReq.FromQueryParams(r.URL.Query()); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get contacts from service
	response, err := h.service.GetContacts(r.Context(), domainReq)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get contacts: %v", err))
		http.Error(w, "Failed to get contacts", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to encode response: %v", err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ContactHandler) handleGetByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get email from query params
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		WriteJSONError(w, "Missing workspace ID", http.StatusBadRequest)
		return
	}
	email := r.URL.Query().Get("email")
	if email == "" {
		WriteJSONError(w, "Missing email", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContactByEmail(r.Context(), workspaceID, email)
	if err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			WriteJSONError(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get contact by email")
		WriteJSONError(w, "Failed to get contact by email", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"contact": contact,
	})
}

func (h *ContactHandler) handleGetByExternalID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get external ID from query params
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		WriteJSONError(w, "Missing workspace ID", http.StatusBadRequest)
		return
	}
	externalID := r.URL.Query().Get("external_id")
	if externalID == "" {
		WriteJSONError(w, "Missing external ID", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContactByExternalID(r.Context(), workspaceID, externalID)
	if err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			WriteJSONError(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get contact by external ID")
		WriteJSONError(w, "Failed to get contact by external ID", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"contact": contact,
	})
}

func (h *ContactHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.DeleteContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContact(r.Context(), req.WorkspaceID, req.Email); err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			WriteJSONError(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to delete contact")
		WriteJSONError(w, "Failed to delete contact", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *ContactHandler) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to read request body")
		WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req domain.BatchImportContactsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contacts, workspaceID, err := req.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.BatchImportContacts(r.Context(), workspaceID, contacts); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to import contacts")
		WriteJSONError(w, "Failed to import contacts", http.StatusInternalServerError)
		return
	}

	// Write success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully imported %d contacts", len(contacts)),
		"count":   len(contacts),
	})
}

func (h *ContactHandler) handleUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req domain.UpsertContactRequest
	if err := json.Unmarshal(body, &req); err != nil {
		WriteJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	contact, workspaceID, err := req.Validate()
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	isNew, err := h.service.UpsertContact(r.Context(), workspaceID, contact)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to upsert contact")
		WriteJSONError(w, "Failed to upsert contact", http.StatusInternalServerError)
		return
	}

	statusCode := http.StatusOK
	action := "updated"
	if isNew {
		statusCode = http.StatusCreated
		action = "created"
	}

	writeJSON(w, statusCode, map[string]interface{}{
		"contact": contact,
		"action":  action,
	})
}
