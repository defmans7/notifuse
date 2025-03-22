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
type addContactToListRequest struct {
	ContactID string `json:"contact_id" valid:"required,uuid"`
	ListID    string `json:"list_id" valid:"required,alphanum"`
	Status    string `json:"status,omitempty"`
}

type getContactListRequest struct {
	ContactID string `json:"contact_id" valid:"required,uuid"`
	ListID    string `json:"list_id" valid:"required,alphanum"`
}

type getContactsByListRequest struct {
	ListID string `json:"list_id" valid:"required,alphanum"`
}

type getListsByContactRequest struct {
	ContactID string `json:"contact_id" valid:"required,uuid"`
}

type updateContactListStatusRequest struct {
	ContactID string `json:"contact_id" valid:"required,uuid"`
	ListID    string `json:"list_id" valid:"required,alphanum"`
	Status    string `json:"status" valid:"required"`
}

type removeContactFromListRequest struct {
	ContactID string `json:"contact_id" valid:"required,uuid"`
	ListID    string `json:"list_id" valid:"required,alphanum"`
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

	var req addContactToListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contactList := &domain.ContactList{
		ContactID: req.ContactID,
		ListID:    req.ListID,
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
	contactID := r.URL.Query().Get("contact_id")
	listID := r.URL.Query().Get("list_id")

	if contactID == "" || listID == "" {
		WriteJSONError(w, "Missing contactID or listID", http.StatusBadRequest)
		return
	}

	contactList, err := h.service.GetContactListByIDs(r.Context(), contactID, listID)
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
	contactID := r.URL.Query().Get("contact_id")
	if contactID == "" {
		WriteJSONError(w, "Missing contact ID", http.StatusBadRequest)
		return
	}

	contactLists, err := h.service.GetListsByContactID(r.Context(), contactID)
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

	var req updateContactListStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ContactID == "" || req.ListID == "" || req.Status == "" {
		WriteJSONError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.service.UpdateContactListStatus(r.Context(), req.ContactID, req.ListID, domain.ContactListStatus(req.Status))
	if err != nil {
		if _, ok := err.(*domain.ErrContactListNotFound); ok {
			WriteJSONError(w, "Contact list relationship not found", http.StatusNotFound)
			return
		}
		h.logger.WithField("error", err.Error()).Error("Failed to update contact list status")
		WriteJSONError(w, "Failed to update contact list status", http.StatusInternalServerError)
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

	var req removeContactFromListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to decode request body")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ContactID == "" || req.ListID == "" {
		WriteJSONError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.service.RemoveContactFromList(r.Context(), req.ContactID, req.ListID)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to remove contact from list")
		WriteJSONError(w, "Failed to remove contact from list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
