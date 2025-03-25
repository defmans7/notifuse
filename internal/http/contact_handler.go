package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/asaskevich/govalidator"
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

// Add the request type for listing contacts
type listContactsRequest struct {
	WorkspaceID string `json:"workspace_id" valid:"required,alphanum,stringlength(1|20)"`
	Email       string `json:"email,omitempty" valid:"optional,email"`
	ExternalID  string `json:"external_id,omitempty" valid:"optional"`
	FirstName   string `json:"first_name,omitempty" valid:"optional"`
	LastName    string `json:"last_name,omitempty" valid:"optional"`
	Phone       string `json:"phone,omitempty" valid:"optional"`
	Country     string `json:"country,omitempty" valid:"optional"`
	Limit       int    `json:"limit,omitempty" valid:"optional,range(1|100)"`
	Cursor      string `json:"cursor,omitempty" valid:"optional"`
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

	// Parse query parameters
	query := r.URL.Query()
	req := listContactsRequest{
		WorkspaceID: query.Get("workspaceId"),
		Email:       query.Get("email"),
		ExternalID:  query.Get("externalId"),
		FirstName:   query.Get("firstName"),
		LastName:    query.Get("lastName"),
		Phone:       query.Get("phone"),
		Country:     query.Get("country"),
	}

	// Parse limit if provided
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		req.Limit = limit
	}

	// Get cursor if provided
	req.Cursor = query.Get("cursor")

	// Validate the request
	if _, err := govalidator.ValidateStruct(req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to domain request
	domainReq := &domain.GetContactsRequest{
		WorkspaceID: req.WorkspaceID,
		Email:       req.Email,
		ExternalID:  req.ExternalID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Phone:       req.Phone,
		Country:     req.Country,
		Limit:       req.Limit,
		Cursor:      req.Cursor,
	}

	// Validate domain request
	if err := domainReq.Validate(); err != nil {
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
	body, err := io.ReadAll(r.Body)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Validate that the body is valid JSON
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(body, &rawJSON); err != nil {
		WriteJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Parse the contact using domain method
	contact, err := domain.FromJSON(body)
	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse custom JSON fields if they exist
	jsonData := gjson.ParseBytes(body)
	for i := 1; i <= 5; i++ {
		field := fmt.Sprintf("custom_json_%d", i)
		if value := jsonData.Get(field); value.Exists() {
			// Check if the value is explicitly null
			if value.Type == gjson.Null {
				continue // Leave the field as invalid
			}

			// Set the custom JSON field
			switch i {
			case 1:
				contact.CustomJSON1 = domain.NullableJSON{Data: value.Value(), Valid: true}
			case 2:
				contact.CustomJSON2 = domain.NullableJSON{Data: value.Value(), Valid: true}
			case 3:
				contact.CustomJSON3 = domain.NullableJSON{Data: value.Value(), Valid: true}
			case 4:
				contact.CustomJSON4 = domain.NullableJSON{Data: value.Value(), Valid: true}
			case 5:
				contact.CustomJSON5 = domain.NullableJSON{Data: value.Value(), Valid: true}
			}
		}
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
