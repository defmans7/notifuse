package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/tidwall/gjson"
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
type getContactByEmailRequest struct {
	Email string `json:"email" valid:"required,email"`
}

type getContactByExternalIDRequest struct {
	ExternalID string `json:"external_id" valid:"required"`
}

type deleteContactRequest struct {
	Email string `json:"email" valid:"required,email"`
}

// Add the request type for batch importing contacts
type batchImportContactsRequest struct {
	Contacts []upsertContactRequest `json:"contacts" valid:"required"`
}

// Add upsert request type that combines create and update
type upsertContactRequest struct {
	ExternalID string `json:"external_id" valid:"required"`
	Email      string `json:"email" valid:"required,email"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	Timezone   string `json:"timezone" valid:"required"`
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

	if req.Email == "" {
		WriteJSONError(w, "Missing email", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContact(r.Context(), req.Email); err != nil {
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to read request body")
		WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Extract contacts array
	contactsArray := gjson.GetBytes(body, "contacts").Array()
	if len(contactsArray) == 0 {
		WriteJSONError(w, "No contacts provided in request", http.StatusBadRequest)
		return
	}

	// Parse each contact
	contacts := make([]*domain.Contact, 0, len(contactsArray))
	for i, contactJson := range contactsArray {
		contact, err := domain.FromJSON(contactJson)
		if err != nil {
			WriteJSONError(w, fmt.Sprintf("Contact at index %d: %s", i, err.Error()), http.StatusBadRequest)
			return
		}
		contacts = append(contacts, contact)
	}

	if err := h.service.BatchImportContacts(r.Context(), contacts); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to import contacts")
		WriteJSONError(w, "Failed to import contacts", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to read request body")
		WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the contact using domain method
	contact, err := domain.FromJSON(body)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	isNew, err := h.service.UpsertContact(r.Context(), contact)
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
