package http

import (
	"encoding/json"
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

// Request/Response types
type createContactRequest struct {
	UUID       string `json:"uuid,omitempty"`
	ExternalID string `json:"external_id" valid:"required"`
	Email      string `json:"email" valid:"required,email"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	Timezone   string `json:"timezone" valid:"required"`
}

type getContactRequest struct {
	UUID string `json:"uuid"`
}

type getContactByEmailRequest struct {
	Email string `json:"email" valid:"required,email"`
}

type getContactByExternalIDRequest struct {
	ExternalID string `json:"external_id" valid:"required"`
}

type updateContactRequest struct {
	UUID       string `json:"uuid" valid:"required,uuid"`
	ExternalID string `json:"external_id" valid:"required"`
	Email      string `json:"email" valid:"required,email"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	Timezone   string `json:"timezone" valid:"required"`
}

type deleteContactRequest struct {
	UUID string `json:"uuid" valid:"required,uuid"`
}

// Add the request type for batch importing contacts
type batchImportContactsRequest struct {
	Contacts []createContactRequest `json:"contacts" valid:"required"`
}

func (h *ContactHandler) RegisterRoutes(mux *http.ServeMux) {
	// Register RPC-style endpoints with dot notation
	mux.HandleFunc("/api/contacts.list", h.handleList)
	mux.HandleFunc("/api/contacts.get", h.handleGet)
	mux.HandleFunc("/api/contacts.getByEmail", h.handleGetByEmail)
	mux.HandleFunc("/api/contacts.getByExternalID", h.handleGetByExternalID)
	mux.HandleFunc("/api/contacts.create", h.handleCreate)
	mux.HandleFunc("/api/contacts.update", h.handleUpdate)
	mux.HandleFunc("/api/contacts.delete", h.handleDelete)
	mux.HandleFunc("/api/contacts.import", h.handleImport)
}

func (h *ContactHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contacts, err := h.service.GetContacts(r.Context())
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to get contacts")
		WriteJSONError(w, "Failed to get contacts", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

func (h *ContactHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get contact UUID from query params
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		WriteJSONError(w, "Missing UUID", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContactByUUID(r.Context(), uuid)
	if err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			WriteJSONError(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get contact")
		WriteJSONError(w, "Failed to get contact", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"contact": contact,
	})
}

func (h *ContactHandler) handleGetByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get email from query params
	email := r.URL.Query().Get("email")
	if email == "" {
		WriteJSONError(w, "Missing email", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContactByEmail(r.Context(), email)
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
	externalID := r.URL.Query().Get("external_id")
	if externalID == "" {
		WriteJSONError(w, "Missing external ID", http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContactByExternalID(r.Context(), externalID)
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

func (h *ContactHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contact := &domain.Contact{
		UUID:       req.UUID,
		ExternalID: req.ExternalID,
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Timezone:   req.Timezone,
	}

	if err := h.service.CreateContact(r.Context(), contact); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to create contact")
		WriteJSONError(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"contact": contact,
	})
}

func (h *ContactHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req updateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UUID == "" {
		WriteJSONError(w, "Missing UUID", http.StatusBadRequest)
		return
	}

	contact := &domain.Contact{
		UUID:       req.UUID,
		ExternalID: req.ExternalID,
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Timezone:   req.Timezone,
	}

	if err := h.service.UpdateContact(r.Context(), contact); err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			WriteJSONError(w, "Contact not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to update contact")
		WriteJSONError(w, "Failed to update contact", http.StatusInternalServerError)
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

	var req deleteContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UUID == "" {
		WriteJSONError(w, "Missing UUID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContact(r.Context(), req.UUID); err != nil {
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

	var req batchImportContactsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if batch size is within limit
	if len(req.Contacts) > 50 {
		WriteJSONError(w, "Batch size exceeds maximum limit of 50 contacts", http.StatusBadRequest)
		return
	}

	// Convert request to domain contacts
	contacts := make([]*domain.Contact, len(req.Contacts))
	for i, contactReq := range req.Contacts {
		contacts[i] = &domain.Contact{
			UUID:       contactReq.UUID,
			ExternalID: contactReq.ExternalID,
			Email:      contactReq.Email,
			FirstName:  contactReq.FirstName,
			LastName:   contactReq.LastName,
			Timezone:   contactReq.Timezone,
		}
	}

	// Process the batch
	if err := h.service.BatchImportContacts(r.Context(), contacts); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to import contacts")
		WriteJSONError(w, "Failed to import contacts", http.StatusInternalServerError)
		return
	}

	// Return success response with imported contacts
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Successfully imported contacts",
		"count":   len(contacts),
	})
}
