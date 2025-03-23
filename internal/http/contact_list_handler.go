package http

import (
	"encoding/json"
	"net/http"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ContactListHandler struct {
	service domain.ContactListService
	logger  logger.Logger
}

func NewContactListHandler(service domain.ContactListService, logger logger.Logger) *ContactListHandler {
	return &ContactListHandler{
		service: service,
		logger:  logger,
	}
}

// Request/Response types
type AddContactToListRequest struct {
	Email  string `json:"email" valid:"required,email"`
	ListID string `json:"list_id" valid:"required"`
	Status string `json:"status" valid:"required,in(active|pending|unsubscribed|blacklisted)"`
}

type GetContactListRequest struct {
	Email  string `json:"email" valid:"required,email"`
	ListID string `json:"list_id" valid:"required"`
}

type GetContactsByListRequest struct {
	ListID string `json:"list_id" valid:"required,alphanum"`
}

type GetListsByContactRequest struct {
	Email string `json:"email" valid:"required,email"`
}

type UpdateContactListStatusRequest struct {
	Email  string `json:"email" valid:"required,email"`
	ListID string `json:"list_id" valid:"required"`
	Status string `json:"status" valid:"required,in(active|pending|unsubscribed|blacklisted)"`
}

type RemoveContactFromListRequest struct {
	Email  string `json:"email" valid:"required,email"`
	ListID string `json:"list_id" valid:"required"`
}

func (h *ContactListHandler) RegisterRoutes(mux *http.ServeMux) {
	// Register RPC-style endpoints with dot notation
	mux.HandleFunc("/api/contactLists.addContact", h.handleAddContact)
	mux.HandleFunc("/api/contactLists.getByIDs", h.handleGetByIDs)
	mux.HandleFunc("/api/contactLists.getContactsByList", h.handleGetContactsByList)
	mux.HandleFunc("/api/contactLists.getListsByContact", h.handleGetListsByContact)
	mux.HandleFunc("/api/contactLists.updateStatus", h.handleUpdateStatus)
	mux.HandleFunc("/api/contactLists.removeContact", h.handleRemoveContact)
}

func (h *ContactListHandler) handleAddContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddContactToListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contactList := &domain.ContactList{
		Email:  req.Email,
		ListID: req.ListID,
	}

	if req.Status != "" {
		contactList.Status = domain.ContactListStatus(req.Status)
	} else {
		contactList.Status = domain.ContactListStatusActive
	}

	if err := h.service.AddContactToList(r.Context(), contactList); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to add contact to list")
		WriteJSONError(w, "Failed to add contact to list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"contact_list": contactList,
	})
}

func (h *ContactListHandler) handleGetByIDs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get IDs from query params
	email := r.URL.Query().Get("email")
	listID := r.URL.Query().Get("list_id")

	if email == "" || listID == "" {
		WriteJSONError(w, "Missing email or listID", http.StatusBadRequest)
		return
	}

	contactList, err := h.service.GetContactListByIDs(r.Context(), email, listID)
	if err != nil {
		if _, ok := err.(*domain.ErrContactListNotFound); ok {
			WriteJSONError(w, "Contact list relationship not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to get contact list")
		WriteJSONError(w, "Failed to get contact list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"contact_list": contactList,
	})
}

func (h *ContactListHandler) handleGetContactsByList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get list ID from query params
	listID := r.URL.Query().Get("list_id")
	if listID == "" {
		WriteJSONError(w, "Missing list ID", http.StatusBadRequest)
		return
	}

	contactLists, err := h.service.GetContactsByListID(r.Context(), listID)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to get contacts by list")
		WriteJSONError(w, "Failed to get contacts by list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"contact_lists": contactLists,
	})
}

func (h *ContactListHandler) handleGetListsByContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get contact ID from query params
	email := r.URL.Query().Get("email")
	if email == "" {
		WriteJSONError(w, "Missing contact ID", http.StatusBadRequest)
		return
	}

	contactLists, err := h.service.GetListsByEmail(r.Context(), email)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to get lists by contact")
		WriteJSONError(w, "Failed to get lists by contact", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"contact_lists": contactLists,
	})
}

func (h *ContactListHandler) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateContactListStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Status == "" || req.ListID == "" {
		WriteJSONError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.service.UpdateContactListStatus(r.Context(), req.Email, req.ListID, domain.ContactListStatus(req.Status))
	if err != nil {
		if _, ok := err.(*domain.ErrContactListNotFound); ok {
			WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to update contact list status")
		WriteJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func (h *ContactListHandler) handleRemoveContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RemoveContactFromListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.ListID == "" {
		WriteJSONError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.service.RemoveContactFromList(r.Context(), req.Email, req.ListID)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to remove contact from list")
		WriteJSONError(w, "Failed to remove contact from list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
