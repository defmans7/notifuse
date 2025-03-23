package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// Request types for testing
type addContactToListRequest struct {
	Email  string `json:"email"`
	ListID string `json:"list_id"`
	Status string `json:"status"`
}

type updateContactListStatusRequest struct {
	Email  string `json:"email"`
	ListID string `json:"list_id"`
	Status string `json:"status"`
}

type removeContactFromListRequest struct {
	Email  string `json:"email"`
	ListID string `json:"list_id"`
}

// MockContactListService is a mock implementation of domain.ContactListService
type MockContactListService struct {
	contactLists map[string]map[string]*domain.ContactList // map[listID]map[Email]*ContactList

	// Function call trackers
	AddContactToListCalled        bool
	GetContactListByIDsCalled     bool
	GetContactsByListIDCalled     bool
	GetListsByEmailCalled         bool
	UpdateContactListStatusCalled bool
	RemoveContactFromListCalled   bool

	LastEmail       string
	LastListID      string
	LastStatus      domain.ContactListStatus
	LastContactList *domain.ContactList

	ErrToReturn                    error
	ErrContactListNotFoundToReturn bool
}

func NewMockContactListService() *MockContactListService {
	return &MockContactListService{
		contactLists: make(map[string]map[string]*domain.ContactList),
	}
}

func (m *MockContactListService) AddContactToList(ctx context.Context, contactList *domain.ContactList) error {
	m.AddContactToListCalled = true
	m.LastEmail = contactList.Email
	m.LastListID = contactList.ListID
	m.LastStatus = contactList.Status
	m.LastContactList = contactList

	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Create map for list if it doesn't exist
	if _, exists := m.contactLists[contactList.ListID]; !exists {
		m.contactLists[contactList.ListID] = make(map[string]*domain.ContactList)
	}

	// Set timestamps
	contactList.CreatedAt = time.Now()
	contactList.UpdatedAt = contactList.CreatedAt

	// Store contact list
	m.contactLists[contactList.ListID][contactList.Email] = contactList
	return nil
}

func (m *MockContactListService) GetContactListByIDs(ctx context.Context, email, listID string) (*domain.ContactList, error) {
	m.GetContactListByIDsCalled = true
	m.LastEmail = email
	m.LastListID = listID

	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	if m.ErrContactListNotFoundToReturn {
		return nil, &domain.ErrContactListNotFound{}
	}

	// Check if list exists
	contacts, exists := m.contactLists[listID]
	if !exists {
		return nil, &domain.ErrContactListNotFound{}
	}

	// Check if contact exists in list
	contactList, exists := contacts[email]
	if !exists {
		return nil, &domain.ErrContactListNotFound{}
	}

	return contactList, nil
}

func (m *MockContactListService) GetContactsByListID(ctx context.Context, listID string) ([]*domain.ContactList, error) {
	m.GetContactsByListIDCalled = true
	m.LastListID = listID

	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	// Check if list exists
	contacts, exists := m.contactLists[listID]
	if !exists {
		return []*domain.ContactList{}, nil
	}

	// Convert map to slice
	results := make([]*domain.ContactList, 0, len(contacts))
	for _, cl := range contacts {
		results = append(results, cl)
	}

	return results, nil
}

func (m *MockContactListService) GetListsByEmail(ctx context.Context, email string) ([]*domain.ContactList, error) {
	m.GetListsByEmailCalled = true
	m.LastEmail = email

	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	// Always return an empty slice instead of nil if no results
	results := []*domain.ContactList{}

	// Find all lists that contain this contact
	for _, contacts := range m.contactLists {
		if cl, exists := contacts[email]; exists {
			results = append(results, cl)
		}
	}

	return results, nil
}

func (m *MockContactListService) UpdateContactListStatus(ctx context.Context, email, listID string, status domain.ContactListStatus) error {
	m.UpdateContactListStatusCalled = true
	m.LastEmail = email
	m.LastListID = listID
	m.LastStatus = status

	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	if m.ErrContactListNotFoundToReturn {
		return &domain.ErrContactListNotFound{}
	}

	// Check if list exists
	contacts, exists := m.contactLists[listID]
	if !exists {
		return &domain.ErrContactListNotFound{}
	}

	// Check if contact exists in list
	contactList, exists := contacts[email]
	if !exists {
		return &domain.ErrContactListNotFound{}
	}

	// Update status
	contactList.Status = status
	contactList.UpdatedAt = time.Now()
	return nil
}

func (m *MockContactListService) RemoveContactFromList(ctx context.Context, email, listID string) error {
	m.RemoveContactFromListCalled = true
	m.LastEmail = email
	m.LastListID = listID

	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Check if list exists
	contacts, exists := m.contactLists[listID]
	if !exists {
		return nil // No error if list doesn't exist
	}

	// Remove contact from list
	delete(contacts, email)
	return nil
}

// MockLoggerForContactList is a mock implementation of logger.Logger for contact list tests
type MockLoggerForContactList struct {
	LoggedMessages []string
}

func (l *MockLoggerForContactList) Info(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "INFO: "+message)
}

func (l *MockLoggerForContactList) Debug(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "DEBUG: "+message)
}

func (l *MockLoggerForContactList) Warn(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "WARN: "+message)
}

func (l *MockLoggerForContactList) Error(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "ERROR: "+message)
}

func (l *MockLoggerForContactList) WithField(key string, value interface{}) logger.Logger {
	return l
}

func (l *MockLoggerForContactList) WithFields(fields map[string]interface{}) logger.Logger {
	return l
}

func (l *MockLoggerForContactList) Fatal(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "FATAL: "+message)
}

// Test setup helper
func setupContactListHandlerTest() (*MockContactListService, *MockLoggerForContactList, *ContactListHandler) {
	mockService := NewMockContactListService()
	mockLogger := &MockLoggerForContactList{LoggedMessages: []string{}}
	handler := NewContactListHandler(mockService, mockLogger)
	return mockService, mockLogger, handler
}

// Helper function to unmarshal JSON response
func decodeContactListJSONResponse(body *bytes.Buffer, v interface{}) error {
	return json.Unmarshal(body.Bytes(), v)
}

func TestContactListHandler_RegisterRoutes(t *testing.T) {
	_, _, handler := setupContactListHandlerTest()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Check if routes were registered - indirect test by ensuring no panic
	endpoints := []string{
		"/api/contactLists.addContact",
		"/api/contactLists.getByIDs",
		"/api/contactLists.getContactsByList",
		"/api/contactLists.getListsByContact",
		"/api/contactLists.updateStatus",
		"/api/contactLists.removeContact",
	}

	for _, endpoint := range endpoints {
		// This is a basic check - just ensure the handler exists
		h, _ := mux.Handler(&http.Request{URL: &url.URL{Path: endpoint}})
		if h == nil {
			t.Errorf("Expected handler to be registered for %s, but got nil", endpoint)
		}
	}
}

func TestContactListHandler_HandleAddContact(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactListService)
		expectedStatus int
		checkAdded     func(*testing.T, *MockContactListService)
	}{
		{
			name:   "Add Contact to List Success",
			method: http.MethodPost,
			reqBody: addContactToListRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
				Status: string(domain.ContactListStatusActive),
			},
			setupMock: func(m *MockContactListService) {
				// No special setup
			},
			expectedStatus: http.StatusCreated,
			checkAdded: func(t *testing.T, m *MockContactListService) {
				if !m.AddContactToListCalled {
					t.Error("Expected AddContactToList to be called, but it wasn't")
				}
				if m.LastEmail != "contact-uuid-1" {
					t.Errorf("Expected email 'contact-uuid-1', got '%s'", m.LastEmail)
				}
				if m.LastListID != "list-id-1" {
					t.Errorf("Expected list ID 'list-id-1', got '%s'", m.LastListID)
				}
				if m.LastStatus != domain.ContactListStatusActive {
					t.Errorf("Expected status '%s', got '%s'", domain.ContactListStatusActive, m.LastStatus)
				}
			},
		},
		{
			name:   "Add Contact to List Default Status",
			method: http.MethodPost,
			reqBody: addContactToListRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
				// No status provided - should default to Active
			},
			setupMock: func(m *MockContactListService) {
				// No special setup
			},
			expectedStatus: http.StatusCreated,
			checkAdded: func(t *testing.T, m *MockContactListService) {
				if !m.AddContactToListCalled {
					t.Error("Expected AddContactToList to be called, but it wasn't")
				}
				if m.LastStatus != domain.ContactListStatusActive {
					t.Errorf("Expected default status '%s', got '%s'", domain.ContactListStatusActive, m.LastStatus)
				}
			},
		},
		{
			name:   "Add Contact to List Service Error",
			method: http.MethodPost,
			reqBody: addContactToListRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
			},
			setupMock: func(m *MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkAdded: func(t *testing.T, m *MockContactListService) {
				if !m.AddContactToListCalled {
					t.Error("Expected AddContactToList to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactListService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkAdded: func(t *testing.T, m *MockContactListService) {
				if m.AddContactToListCalled {
					t.Error("Expected AddContactToList not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: addContactToListRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
			},
			setupMock: func(m *MockContactListService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkAdded: func(t *testing.T, m *MockContactListService) {
				if m.AddContactToListCalled {
					t.Error("Expected AddContactToList not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactListHandlerTest()
			tc.setupMock(mockService)

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				// If it's a string, just use it directly
				if str, ok := tc.reqBody.(string); ok {
					reqBody = *bytes.NewBufferString(str)
				} else {
					// Otherwise encode as JSON
					if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
						t.Fatalf("Failed to encode request body: %v", err)
					}
				}
			}

			req, err := http.NewRequest(tc.method, "/api/contactLists.addContact", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleAddContact(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := decodeContactListJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				contactListData, exists := response["contact_list"]
				if !exists {
					t.Error("Expected 'contact_list' field in response, but not found")
				}

				// Verify the response contains the contact list with correct data
				if contactListMap, ok := contactListData.(map[string]interface{}); ok {
					if req, ok := tc.reqBody.(addContactToListRequest); ok {
						if contactListMap["email"] != req.Email {
							t.Errorf("Expected email %s, got %v", req.Email, contactListMap["email"])
						}
						if contactListMap["list_id"] != req.ListID {
							t.Errorf("Expected list_id %s, got %v", req.ListID, contactListMap["list_id"])
						}
					}
				}
			}

			// Run additional checks
			tc.checkAdded(t, mockService)
		})
	}
}

func TestContactListHandler_HandleGetByIDs(t *testing.T) {
	testCases := []struct {
		name                string
		method              string
		email               string
		listID              string
		setupMock           func(*MockContactListService)
		expectedStatus      int
		expectedContactList bool
	}{
		{
			name:   "Get Contact List Success",
			method: http.MethodGet,
			email:  "contact-uuid-1",
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				// Initialize the map structure
				m.contactLists = map[string]map[string]*domain.ContactList{
					"list-id-1": {
						"contact-uuid-1": {
							Email:     "contact-uuid-1",
							ListID:    "list-id-1",
							Status:    domain.ContactListStatusActive,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
				}
			},
			expectedStatus:      http.StatusOK,
			expectedContactList: true,
		},
		{
			name:   "Get Contact List Not Found",
			method: http.MethodGet,
			email:  "nonexistent",
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				m.ErrContactListNotFoundToReturn = true
			},
			expectedStatus:      http.StatusNotFound,
			expectedContactList: false,
		},
		{
			name:   "Get Contact List Service Error",
			method: http.MethodGet,
			email:  "contact-uuid-1",
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:      http.StatusInternalServerError,
			expectedContactList: false,
		},
		{
			name:   "Missing Contact ID",
			method: http.MethodGet,
			email:  "",
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus:      http.StatusBadRequest,
			expectedContactList: false,
		},
		{
			name:   "Missing List ID",
			method: http.MethodGet,
			email:  "contact-uuid-1",
			listID: "",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus:      http.StatusBadRequest,
			expectedContactList: false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			email:  "contact-uuid-1",
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus:      http.StatusMethodNotAllowed,
			expectedContactList: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactListHandlerTest()
			tc.setupMock(mockService)

			url := "/api/contactLists.getByIDs"
			if tc.email != "" || tc.listID != "" {
				url += "?"
				if tc.email != "" {
					url += "email=" + tc.email
				}
				if tc.email != "" && tc.listID != "" {
					url += "&"
				}
				if tc.listID != "" {
					url += "list_id=" + tc.listID
				}
			}

			req, err := http.NewRequest(tc.method, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleGetByIDs(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedContactList {
				if tc.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					if err := decodeContactListJSONResponse(rr.Body, &response); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					contactListData, exists := response["contact_list"]
					if !exists {
						t.Error("Expected 'contact_list' field in response, but not found")
					}

					// Verify the response contains the contact list with correct data
					if contactListMap, ok := contactListData.(map[string]interface{}); ok {
						if contactListMap["email"] != tc.email {
							t.Errorf("Expected email %s, got %v", tc.email, contactListMap["email"])
						}
						if contactListMap["list_id"] != tc.listID {
							t.Errorf("Expected list_id %s, got %v", tc.listID, contactListMap["list_id"])
						}
					}
				}
			}

			if tc.method == http.MethodGet && tc.email != "" && tc.listID != "" &&
				tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetContactListByIDsCalled {
					t.Error("Expected GetContactListByIDs to be called, but it wasn't")
				}
				if mockService.LastEmail != tc.email {
					t.Errorf("Expected Email %s, got %s", tc.email, mockService.LastEmail)
				}
				if mockService.LastListID != tc.listID {
					t.Errorf("Expected ListID %s, got %s", tc.listID, mockService.LastListID)
				}
			}
		})
	}
}

func TestContactListHandler_HandleGetContactsByList(t *testing.T) {
	testCases := []struct {
		name                 string
		method               string
		listID               string
		setupMock            func(*MockContactListService)
		expectedStatus       int
		expectedContactLists bool
	}{
		{
			name:   "Get Contacts By List Success",
			method: http.MethodGet,
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				// Initialize the map structure with multiple contacts in the list
				m.contactLists = map[string]map[string]*domain.ContactList{
					"list-id-1": {
						"contact-uuid-1": {
							Email:  "contact-uuid-1",
							ListID: "list-id-1",
							Status: domain.ContactListStatusActive,
						},
						"contact-uuid-2": {
							Email:  "contact-uuid-2",
							ListID: "list-id-1",
							Status: domain.ContactListStatusPending,
						},
					},
				}
			},
			expectedStatus:       http.StatusOK,
			expectedContactLists: true,
		},
		{
			name:   "Get Contacts By List Empty Result",
			method: http.MethodGet,
			listID: "empty-list",
			setupMock: func(m *MockContactListService) {
				// No contacts in this list
				m.contactLists = map[string]map[string]*domain.ContactList{
					"empty-list": {},
				}
			},
			expectedStatus:       http.StatusOK,
			expectedContactLists: true,
		},
		{
			name:   "Get Contacts By List Service Error",
			method: http.MethodGet,
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedContactLists: false,
		},
		{
			name:   "Missing List ID",
			method: http.MethodGet,
			listID: "",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus:       http.StatusBadRequest,
			expectedContactLists: false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			listID: "list-id-1",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus:       http.StatusMethodNotAllowed,
			expectedContactLists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactListHandlerTest()
			tc.setupMock(mockService)

			url := "/api/contactLists.getContactsByList"
			if tc.listID != "" {
				url += "?list_id=" + tc.listID
			}

			req, err := http.NewRequest(tc.method, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleGetContactsByList(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedContactLists {
				if tc.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					if err := decodeContactListJSONResponse(rr.Body, &response); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					contactListsData, exists := response["contact_lists"]
					if !exists {
						t.Error("Expected 'contact_lists' field in response, but not found")
					}

					// Check we got an array response (even if empty)
					if _, ok := contactListsData.([]interface{}); !ok {
						t.Errorf("Expected contact_lists to be an array, got %T", contactListsData)
					}
				}
			}

			if tc.method == http.MethodGet && tc.listID != "" &&
				tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetContactsByListIDCalled {
					t.Error("Expected GetContactsByListID to be called, but it wasn't")
				}
				if mockService.LastListID != tc.listID {
					t.Errorf("Expected ListID %s, got %s", tc.listID, mockService.LastListID)
				}
			}
		})
	}
}

func TestContactListHandler_HandleGetListsByContact(t *testing.T) {
	testCases := []struct {
		name                 string
		method               string
		email                string
		setupMock            func(*MockContactListService)
		expectedStatus       int
		expectedContactLists bool
	}{
		{
			name:   "Get Lists By Contact Success",
			method: http.MethodGet,
			email:  "contact-uuid-1",
			setupMock: func(m *MockContactListService) {
				// Initialize the map structure with one contact in multiple lists
				m.contactLists = map[string]map[string]*domain.ContactList{
					"list-id-1": {
						"contact-uuid-1": {
							Email:  "contact-uuid-1",
							ListID: "list-id-1",
							Status: domain.ContactListStatusActive,
						},
					},
					"list-id-2": {
						"contact-uuid-1": {
							Email:  "contact-uuid-1",
							ListID: "list-id-2",
							Status: domain.ContactListStatusPending,
						},
					},
				}
			},
			expectedStatus:       http.StatusOK,
			expectedContactLists: true,
		},
		{
			name:   "Get Lists By Contact Empty Result",
			method: http.MethodGet,
			email:  "contact-with-no-lists",
			setupMock: func(m *MockContactListService) {
				// Return empty array but not nil
				// This simulates what the actual service would do
			},
			expectedStatus:       http.StatusOK,
			expectedContactLists: true,
		},
		{
			name:   "Get Lists By Contact Service Error",
			method: http.MethodGet,
			email:  "contact-uuid-1",
			setupMock: func(m *MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedContactLists: false,
		},
		{
			name:   "Missing Contact ID",
			method: http.MethodGet,
			email:  "",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus:       http.StatusBadRequest,
			expectedContactLists: false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			email:  "contact-uuid-1",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus:       http.StatusMethodNotAllowed,
			expectedContactLists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactListHandlerTest()
			tc.setupMock(mockService)

			url := "/api/contactLists.getListsByContact"
			if tc.email != "" {
				url += "?email=" + tc.email
			}

			req, err := http.NewRequest(tc.method, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleGetListsByContact(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedContactLists {
				if tc.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					if err := decodeContactListJSONResponse(rr.Body, &response); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
						return
					}

					contactListsData, exists := response["contact_lists"]
					if !exists {
						t.Error("Expected 'contact_lists' field in response, but not found")
						return
					}

					// Check if we got an array (even if empty)
					contactLists, ok := contactListsData.([]interface{})
					if !ok {
						t.Errorf("Expected contact_lists to be an array, got %T", contactListsData)
					} else {
						// For empty result test, verify array is empty
						if tc.name == "Get Lists By Contact Empty Result" && len(contactLists) > 0 {
							t.Errorf("Expected empty array, got %d items", len(contactLists))
						}
					}
				}
			}

			if tc.method == http.MethodGet && tc.email != "" &&
				tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetListsByEmailCalled {
					t.Error("Expected GetListsByEmail to be called, but it wasn't")
				}
				if mockService.LastEmail != tc.email {
					t.Errorf("Expected Email %s, got %s", tc.email, mockService.LastEmail)
				}
			}
		})
	}
}

func TestContactListHandler_HandleUpdateStatus(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactListService)
		expectedStatus int
		checkUpdated   func(*testing.T, *MockContactListService)
	}{
		{
			name:   "Update Contact List Status Success",
			method: http.MethodPost,
			reqBody: updateContactListStatusRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
				Status: string(domain.ContactListStatusActive),
			},
			setupMock: func(m *MockContactListService) {
				// Initialize the contact list in the mock
				m.contactLists = map[string]map[string]*domain.ContactList{
					"list-id-1": {
						"contact-uuid-1": {
							Email:  "contact-uuid-1",
							ListID: "list-id-1",
							Status: domain.ContactListStatusPending,
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkUpdated: func(t *testing.T, m *MockContactListService) {
				if !m.UpdateContactListStatusCalled {
					t.Error("Expected UpdateContactListStatus to be called, but it wasn't")
				}
				if m.LastEmail != "contact-uuid-1" {
					t.Errorf("Expected email 'contact-uuid-1', got '%s'", m.LastEmail)
				}
				if m.LastListID != "list-id-1" {
					t.Errorf("Expected list ID 'list-id-1', got '%s'", m.LastListID)
				}
				if m.LastStatus != domain.ContactListStatusActive {
					t.Errorf("Expected status '%s', got '%s'", domain.ContactListStatusActive, m.LastStatus)
				}
			},
		},
		{
			name:   "Update Contact List Status Not Found",
			method: http.MethodPost,
			reqBody: updateContactListStatusRequest{
				Email:  "nonexistent",
				ListID: "list-id-1",
				Status: string(domain.ContactListStatusActive),
			},
			setupMock: func(m *MockContactListService) {
				m.ErrContactListNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkUpdated: func(t *testing.T, m *MockContactListService) {
				if !m.UpdateContactListStatusCalled {
					t.Error("Expected UpdateContactListStatus to be called, but it wasn't")
				}
			},
		},
		{
			name:   "Update Contact List Status Service Error",
			method: http.MethodPost,
			reqBody: updateContactListStatusRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
				Status: string(domain.ContactListStatusActive),
			},
			setupMock: func(m *MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkUpdated: func(t *testing.T, m *MockContactListService) {
				if !m.UpdateContactListStatusCalled {
					t.Error("Expected UpdateContactListStatus to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			checkUpdated: func(t *testing.T, m *MockContactListService) {
				if m.UpdateContactListStatusCalled {
					t.Error("Expected UpdateContactListStatus not to be called, but it was")
				}
			},
		},
		{
			name:   "Missing Required Fields",
			method: http.MethodPost,
			reqBody: updateContactListStatusRequest{
				Email: "contact-uuid-1",
				// ListID and Status are missing
			},
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			checkUpdated: func(t *testing.T, m *MockContactListService) {
				if m.UpdateContactListStatusCalled {
					t.Error("Expected UpdateContactListStatus not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: updateContactListStatusRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
				Status: string(domain.ContactListStatusActive),
			},
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkUpdated: func(t *testing.T, m *MockContactListService) {
				if m.UpdateContactListStatusCalled {
					t.Error("Expected UpdateContactListStatus not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactListHandlerTest()
			tc.setupMock(mockService)

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				// If it's a string, just use it directly
				if str, ok := tc.reqBody.(string); ok {
					reqBody = *bytes.NewBufferString(str)
				} else {
					// Otherwise encode as JSON
					if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
						t.Fatalf("Failed to encode request body: %v", err)
					}
				}
			}

			req, err := http.NewRequest(tc.method, "/api/contactLists.updateStatus", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleUpdateStatus(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := decodeContactListJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				success, exists := response["success"]
				if !exists {
					t.Error("Expected 'success' field in response, but not found")
				}
				if !success.(bool) {
					t.Error("Expected 'success' to be true, but it was false")
				}
			}

			// Run additional checks
			tc.checkUpdated(t, mockService)
		})
	}
}

func TestContactListHandler_HandleRemoveContact(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactListService)
		expectedStatus int
		checkRemoved   func(*testing.T, *MockContactListService)
	}{
		{
			name:   "Remove Contact from List Success",
			method: http.MethodPost,
			reqBody: removeContactFromListRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
			},
			setupMock: func(m *MockContactListService) {
				// Initialize the contact list in the mock
				m.contactLists = map[string]map[string]*domain.ContactList{
					"list-id-1": {
						"contact-uuid-1": {
							Email:  "contact-uuid-1",
							ListID: "list-id-1",
							Status: domain.ContactListStatusActive,
						},
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkRemoved: func(t *testing.T, m *MockContactListService) {
				if !m.RemoveContactFromListCalled {
					t.Error("Expected RemoveContactFromList to be called, but it wasn't")
				}
				if m.LastEmail != "contact-uuid-1" {
					t.Errorf("Expected email 'contact-uuid-1', got '%s'", m.LastEmail)
				}
				if m.LastListID != "list-id-1" {
					t.Errorf("Expected list ID 'list-id-1', got '%s'", m.LastListID)
				}
			},
		},
		{
			name:   "Remove Contact from List Service Error",
			method: http.MethodPost,
			reqBody: removeContactFromListRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
			},
			setupMock: func(m *MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkRemoved: func(t *testing.T, m *MockContactListService) {
				if !m.RemoveContactFromListCalled {
					t.Error("Expected RemoveContactFromList to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			checkRemoved: func(t *testing.T, m *MockContactListService) {
				if m.RemoveContactFromListCalled {
					t.Error("Expected RemoveContactFromList not to be called, but it was")
				}
			},
		},
		{
			name:   "Missing Required Fields",
			method: http.MethodPost,
			reqBody: removeContactFromListRequest{
				Email: "contact-uuid-1",
				// ListID is missing
			},
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			checkRemoved: func(t *testing.T, m *MockContactListService) {
				if m.RemoveContactFromListCalled {
					t.Error("Expected RemoveContactFromList not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: removeContactFromListRequest{
				Email:  "contact-uuid-1",
				ListID: "list-id-1",
			},
			setupMock: func(m *MockContactListService) {
				// No setup needed
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkRemoved: func(t *testing.T, m *MockContactListService) {
				if m.RemoveContactFromListCalled {
					t.Error("Expected RemoveContactFromList not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactListHandlerTest()
			tc.setupMock(mockService)

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				// If it's a string, just use it directly
				if str, ok := tc.reqBody.(string); ok {
					reqBody = *bytes.NewBufferString(str)
				} else {
					// Otherwise encode as JSON
					if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
						t.Fatalf("Failed to encode request body: %v", err)
					}
				}
			}

			req, err := http.NewRequest(tc.method, "/api/contactLists.removeContact", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleRemoveContact(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := decodeContactListJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				success, exists := response["success"]
				if !exists {
					t.Error("Expected 'success' field in response, but not found")
				}
				if !success.(bool) {
					t.Error("Expected 'success' to be true, but it was false")
				}
			}

			// Run additional checks
			tc.checkRemoved(t, mockService)
		})
	}
}
