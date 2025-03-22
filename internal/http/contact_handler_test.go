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

// MockContactService is a mock implementation of domain.ContactService
type MockContactService struct {
	contacts map[string]*domain.Contact

	// Function call trackers
	GetContactsCalled            bool
	GetContactByUUIDCalled       bool
	GetContactByEmailCalled      bool
	GetContactByExternalIDCalled bool
	CreateContactCalled          bool
	UpdateContactCalled          bool
	DeleteContactCalled          bool
	LastContactUUID              string
	LastContactEmail             string
	LastContactExternalID        string
	LastContactCreated           *domain.Contact
	LastContactUpdated           *domain.Contact
	ErrToReturn                  error
	ErrContactNotFoundToReturn   bool
}

func NewMockContactService() *MockContactService {
	return &MockContactService{
		contacts: make(map[string]*domain.Contact),
	}
}

func (m *MockContactService) GetContacts(ctx context.Context) ([]*domain.Contact, error) {
	m.GetContactsCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	contacts := make([]*domain.Contact, 0, len(m.contacts))
	for _, contact := range m.contacts {
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

func (m *MockContactService) GetContactByUUID(ctx context.Context, uuid string) (*domain.Contact, error) {
	m.GetContactByUUIDCalled = true
	m.LastContactUUID = uuid
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return nil, &domain.ErrContactNotFound{}
	}

	contact, exists := m.contacts[uuid]
	if !exists {
		return nil, &domain.ErrContactNotFound{}
	}
	return contact, nil
}

func (m *MockContactService) GetContactByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	m.GetContactByEmailCalled = true
	m.LastContactEmail = email
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return nil, &domain.ErrContactNotFound{}
	}

	for _, contact := range m.contacts {
		if contact.Email == email {
			return contact, nil
		}
	}
	return nil, &domain.ErrContactNotFound{}
}

func (m *MockContactService) GetContactByExternalID(ctx context.Context, externalID string) (*domain.Contact, error) {
	m.GetContactByExternalIDCalled = true
	m.LastContactExternalID = externalID
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return nil, &domain.ErrContactNotFound{}
	}

	for _, contact := range m.contacts {
		if contact.ExternalID == externalID {
			return contact, nil
		}
	}
	return nil, &domain.ErrContactNotFound{}
}

func (m *MockContactService) CreateContact(ctx context.Context, contact *domain.Contact) error {
	m.CreateContactCalled = true
	m.LastContactCreated = contact
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Set timestamps
	contact.CreatedAt = time.Now()
	contact.UpdatedAt = contact.CreatedAt

	m.contacts[contact.UUID] = contact
	return nil
}

func (m *MockContactService) UpdateContact(ctx context.Context, contact *domain.Contact) error {
	m.UpdateContactCalled = true
	m.LastContactUpdated = contact
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return &domain.ErrContactNotFound{}
	}

	_, exists := m.contacts[contact.UUID]
	if !exists {
		return &domain.ErrContactNotFound{}
	}

	// Update timestamp
	contact.UpdatedAt = time.Now()

	m.contacts[contact.UUID] = contact
	return nil
}

func (m *MockContactService) DeleteContact(ctx context.Context, uuid string) error {
	m.DeleteContactCalled = true
	m.LastContactUUID = uuid
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return &domain.ErrContactNotFound{}
	}

	_, exists := m.contacts[uuid]
	if !exists {
		return &domain.ErrContactNotFound{}
	}

	delete(m.contacts, uuid)
	return nil
}

// MockLoggerForContact is a mock implementation of logger.Logger for contact tests
type MockLoggerForContact struct {
	LoggedMessages []string
}

func (l *MockLoggerForContact) Info(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "INFO: "+message)
}

func (l *MockLoggerForContact) Debug(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "DEBUG: "+message)
}

func (l *MockLoggerForContact) Warn(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "WARN: "+message)
}

func (l *MockLoggerForContact) Error(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "ERROR: "+message)
}

func (l *MockLoggerForContact) WithField(key string, value interface{}) logger.Logger {
	return l
}

func (l *MockLoggerForContact) WithFields(fields map[string]interface{}) logger.Logger {
	return l
}

func (l *MockLoggerForContact) Fatal(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "FATAL: "+message)
}

// Test setup helper
func setupContactHandlerTest() (*MockContactService, *MockLoggerForContact, *ContactHandler) {
	mockService := NewMockContactService()
	mockLogger := &MockLoggerForContact{LoggedMessages: []string{}}
	handler := NewContactHandler(mockService, mockLogger)
	return mockService, mockLogger, handler
}

// Helper function to unmarshal JSON response
func decodeContactJSONResponse(body *bytes.Buffer, v interface{}) error {
	return json.Unmarshal(body.Bytes(), v)
}

func TestContactHandler_RegisterRoutes(t *testing.T) {
	_, _, handler := setupContactHandlerTest()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Check if routes were registered - indirect test by ensuring no panic
	endpoints := []string{
		"/api/contacts.list",
		"/api/contacts.get",
		"/api/contacts.getByEmail",
		"/api/contacts.getByExternalID",
		"/api/contacts.create",
		"/api/contacts.update",
		"/api/contacts.delete",
	}

	for _, endpoint := range endpoints {
		// This is a basic check - just ensure the handler exists
		h, _ := mux.Handler(&http.Request{URL: &url.URL{Path: endpoint}})
		if h == nil {
			t.Errorf("Expected handler to be registered for %s, but got nil", endpoint)
		}
	}
}

func TestContactHandler_HandleList(t *testing.T) {
	testCases := []struct {
		name             string
		method           string
		setupMock        func(*MockContactService)
		expectedStatus   int
		expectedContacts bool
	}{
		{
			name:   "Get Contacts Success",
			method: http.MethodGet,
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"uuid1": {UUID: "uuid1", Email: "test1@example.com", ExternalID: "ext1", Timezone: "UTC"},
					"uuid2": {UUID: "uuid2", Email: "test2@example.com", ExternalID: "ext2", Timezone: "UTC"},
				}
			},
			expectedStatus:   http.StatusOK,
			expectedContacts: true,
		},
		{
			name:   "Get Contacts Empty Result",
			method: http.MethodGet,
			setupMock: func(m *MockContactService) {
				// No contacts in the mock
			},
			expectedStatus:   http.StatusOK,
			expectedContacts: true,
		},
		{
			name:   "Get Contacts Service Error",
			method: http.MethodGet,
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedContacts: false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedContacts: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerTest()
			tc.setupMock(mockService)

			req, err := http.NewRequest(tc.method, "/api/contacts.list", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleList(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedContacts {
				if tc.expectedStatus == http.StatusOK {
					var contacts []*domain.Contact
					if err := decodeContactJSONResponse(rr.Body, &contacts); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					if len(contacts) != len(mockService.contacts) {
						t.Errorf("Expected %d contacts, got %d", len(mockService.contacts), len(contacts))
					}
				}
			}

			if !mockService.GetContactsCalled && tc.method == http.MethodGet {
				t.Error("Expected GetContacts to be called, but it wasn't")
			}
		})
	}
}

func TestContactHandler_HandleGet(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		contactUUID     string
		setupMock       func(*MockContactService)
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:        "Get Contact Success",
			method:      http.MethodGet,
			contactUUID: "uuid1",
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"uuid1": {UUID: "uuid1", Email: "test1@example.com", ExternalID: "ext1", Timezone: "UTC"},
				}
			},
			expectedStatus:  http.StatusOK,
			expectedContact: true,
		},
		{
			name:        "Get Contact Not Found",
			method:      http.MethodGet,
			contactUUID: "nonexistent",
			setupMock: func(m *MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:        "Get Contact Service Error",
			method:      http.MethodGet,
			contactUUID: "uuid1",
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:        "Missing Contact UUID",
			method:      http.MethodGet,
			contactUUID: "",
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:        "Method Not Allowed",
			method:      http.MethodPost,
			contactUUID: "uuid1",
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedContact: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerTest()
			tc.setupMock(mockService)

			url := "/api/contacts.get"
			if tc.contactUUID != "" {
				url += "?uuid=" + tc.contactUUID
			}

			req, err := http.NewRequest(tc.method, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleGet(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedContact {
				if tc.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					contactData, exists := response["contact"]
					if !exists {
						t.Error("Expected 'contact' field in response, but not found")
					}

					// Convert to map to access fields
					contactMap, ok := contactData.(map[string]interface{})
					if !ok {
						t.Errorf("Expected contact to be a map, got %T", contactData)
					} else if contactMap["uuid"] != tc.contactUUID {
						t.Errorf("Expected contact UUID %s, got %v", tc.contactUUID, contactMap["uuid"])
					}
				}
			}

			if tc.method == http.MethodGet && tc.contactUUID != "" && tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetContactByUUIDCalled {
					t.Error("Expected GetContactByUUID to be called, but it wasn't")
				}
				if mockService.LastContactUUID != tc.contactUUID {
					t.Errorf("Expected ContactUUID %s, got %s", tc.contactUUID, mockService.LastContactUUID)
				}
			}
		})
	}
}

func TestContactHandler_HandleGetByEmail(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		email           string
		setupMock       func(*MockContactService)
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:   "Get Contact By Email Success",
			method: http.MethodGet,
			email:  "test@example.com",
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"uuid1": {UUID: "uuid1", Email: "test@example.com", ExternalID: "ext1", Timezone: "UTC"},
				}
			},
			expectedStatus:  http.StatusOK,
			expectedContact: true,
		},
		{
			name:   "Get Contact By Email Not Found",
			method: http.MethodGet,
			email:  "nonexistent@example.com",
			setupMock: func(m *MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:   "Get Contact By Email Service Error",
			method: http.MethodGet,
			email:  "test@example.com",
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:   "Missing Email",
			method: http.MethodGet,
			email:  "",
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			email:  "test@example.com",
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedContact: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerTest()
			tc.setupMock(mockService)

			url := "/api/contacts.getByEmail"
			if tc.email != "" {
				url += "?email=" + tc.email
			}

			req, err := http.NewRequest(tc.method, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleGetByEmail(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedContact {
				if tc.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					contactData, exists := response["contact"]
					if !exists {
						t.Error("Expected 'contact' field in response, but not found")
					}

					// Convert to map to access fields
					contactMap, ok := contactData.(map[string]interface{})
					if !ok {
						t.Errorf("Expected contact to be a map, got %T", contactData)
					} else if contactMap["email"] != tc.email {
						t.Errorf("Expected contact email %s, got %v", tc.email, contactMap["email"])
					}
				}
			}

			if tc.method == http.MethodGet && tc.email != "" && tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetContactByEmailCalled {
					t.Error("Expected GetContactByEmail to be called, but it wasn't")
				}
				if mockService.LastContactEmail != tc.email {
					t.Errorf("Expected Email %s, got %s", tc.email, mockService.LastContactEmail)
				}
			}
		})
	}
}

func TestContactHandler_HandleGetByExternalID(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		externalID      string
		setupMock       func(*MockContactService)
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:       "Get Contact By External ID Success",
			method:     http.MethodGet,
			externalID: "ext1",
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"uuid1": {UUID: "uuid1", Email: "test@example.com", ExternalID: "ext1", Timezone: "UTC"},
				}
			},
			expectedStatus:  http.StatusOK,
			expectedContact: true,
		},
		{
			name:       "Get Contact By External ID Not Found",
			method:     http.MethodGet,
			externalID: "nonexistent",
			setupMock: func(m *MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:       "Get Contact By External ID Service Error",
			method:     http.MethodGet,
			externalID: "ext1",
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:       "Missing External ID",
			method:     http.MethodGet,
			externalID: "",
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:       "Method Not Allowed",
			method:     http.MethodPost,
			externalID: "ext1",
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedContact: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerTest()
			tc.setupMock(mockService)

			url := "/api/contacts.getByExternalID"
			if tc.externalID != "" {
				url += "?external_id=" + tc.externalID
			}

			req, err := http.NewRequest(tc.method, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleGetByExternalID(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedContact {
				if tc.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					contactData, exists := response["contact"]
					if !exists {
						t.Error("Expected 'contact' field in response, but not found")
					}

					// Convert to map to access fields
					contactMap, ok := contactData.(map[string]interface{})
					if !ok {
						t.Errorf("Expected contact to be a map, got %T", contactData)
					} else if contactMap["external_id"] != tc.externalID {
						t.Errorf("Expected contact external_id %s, got %v", tc.externalID, contactMap["external_id"])
					}
				}
			}

			if tc.method == http.MethodGet && tc.externalID != "" && tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetContactByExternalIDCalled {
					t.Error("Expected GetContactByExternalID to be called, but it wasn't")
				}
				if mockService.LastContactExternalID != tc.externalID {
					t.Errorf("Expected ExternalID %s, got %s", tc.externalID, mockService.LastContactExternalID)
				}
			}
		})
	}
}

func TestContactHandler_HandleCreate(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactService)
		expectedStatus int
		checkCreated   func(*testing.T, *MockContactService)
	}{
		{
			name:   "Create Contact Success",
			method: http.MethodPost,
			reqBody: createContactRequest{
				UUID:       "new-uuid",
				ExternalID: "new-ext",
				Email:      "new@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusCreated,
			checkCreated: func(t *testing.T, m *MockContactService) {
				if !m.CreateContactCalled {
					t.Error("Expected CreateContact to be called, but it wasn't")
				}
				if m.LastContactCreated == nil {
					t.Fatal("Expected contact to be created, but it wasn't")
				}
				if m.LastContactCreated.UUID != "new-uuid" {
					t.Errorf("Expected contact UUID 'new-uuid', got '%s'", m.LastContactCreated.UUID)
				}
				if m.LastContactCreated.Email != "new@example.com" {
					t.Errorf("Expected contact email 'new@example.com', got '%s'", m.LastContactCreated.Email)
				}
				if m.LastContactCreated.FirstName != "John" {
					t.Errorf("Expected contact first name 'John', got '%s'", m.LastContactCreated.FirstName)
				}
			},
		},
		{
			name:   "Create Contact Service Error",
			method: http.MethodPost,
			reqBody: createContactRequest{
				UUID:       "error-uuid",
				ExternalID: "error-ext",
				Email:      "error@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkCreated: func(t *testing.T, m *MockContactService) {
				if !m.CreateContactCalled {
					t.Error("Expected CreateContact to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkCreated: func(t *testing.T, m *MockContactService) {
				if m.CreateContactCalled {
					t.Error("Expected CreateContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: createContactRequest{
				UUID:       "new-uuid",
				ExternalID: "new-ext",
				Email:      "new@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkCreated: func(t *testing.T, m *MockContactService) {
				if m.CreateContactCalled {
					t.Error("Expected CreateContact not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerTest()
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

			req, err := http.NewRequest(tc.method, "/api/contacts.create", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleCreate(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				contactData, exists := response["contact"]
				if !exists {
					t.Error("Expected 'contact' field in response, but not found")
				}

				// Verify the response contains the created contact
				if contactMap, ok := contactData.(map[string]interface{}); ok {
					if req, ok := tc.reqBody.(createContactRequest); ok {
						if contactMap["uuid"] != req.UUID {
							t.Errorf("Expected contact UUID %s, got %v", req.UUID, contactMap["uuid"])
						}
						if contactMap["email"] != req.Email {
							t.Errorf("Expected contact email %s, got %v", req.Email, contactMap["email"])
						}
					}
				}
			}

			// Run additional checks
			tc.checkCreated(t, mockService)
		})
	}
}

func TestContactHandler_HandleUpdate(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactService)
		expectedStatus int
		checkUpdated   func(*testing.T, *MockContactService)
	}{
		{
			name:   "Update Contact Success",
			method: http.MethodPost,
			reqBody: updateContactRequest{
				UUID:       "uuid1",
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				FirstName:  "Updated",
				LastName:   "Name",
				Timezone:   "Europe/London",
			},
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"uuid1": {UUID: "uuid1", Email: "test@example.com", ExternalID: "ext1", Timezone: "UTC"},
				}
			},
			expectedStatus: http.StatusOK,
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if !m.UpdateContactCalled {
					t.Error("Expected UpdateContact to be called, but it wasn't")
				}
				if m.LastContactUpdated == nil {
					t.Fatal("Expected contact to be updated, but it wasn't")
				}
				if m.LastContactUpdated.UUID != "uuid1" {
					t.Errorf("Expected contact UUID 'uuid1', got '%s'", m.LastContactUpdated.UUID)
				}
				if m.LastContactUpdated.Email != "updated@example.com" {
					t.Errorf("Expected contact email 'updated@example.com', got '%s'", m.LastContactUpdated.Email)
				}
				if m.LastContactUpdated.FirstName != "Updated" {
					t.Errorf("Expected contact first name 'Updated', got '%s'", m.LastContactUpdated.FirstName)
				}
				if m.LastContactUpdated.Timezone != "Europe/London" {
					t.Errorf("Expected contact timezone 'Europe/London', got '%s'", m.LastContactUpdated.Timezone)
				}
			},
		},
		{
			name:   "Update Contact Not Found",
			method: http.MethodPost,
			reqBody: updateContactRequest{
				UUID:       "nonexistent",
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if !m.UpdateContactCalled {
					t.Error("Expected UpdateContact to be called, but it wasn't")
				}
			},
		},
		{
			name:   "Update Contact Service Error",
			method: http.MethodPost,
			reqBody: updateContactRequest{
				UUID:       "uuid1",
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if !m.UpdateContactCalled {
					t.Error("Expected UpdateContact to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if m.UpdateContactCalled {
					t.Error("Expected UpdateContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Missing UUID in Request",
			method: http.MethodPost,
			reqBody: updateContactRequest{
				UUID:       "", // Empty UUID
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if m.UpdateContactCalled {
					t.Error("Expected UpdateContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: updateContactRequest{
				UUID:       "uuid1",
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if m.UpdateContactCalled {
					t.Error("Expected UpdateContact not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerTest()
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

			req, err := http.NewRequest(tc.method, "/api/contacts.update", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleUpdate(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				contactData, exists := response["contact"]
				if !exists {
					t.Error("Expected 'contact' field in response, but not found")
				}

				// Verify the response contains the updated contact
				if contactMap, ok := contactData.(map[string]interface{}); ok {
					if req, ok := tc.reqBody.(updateContactRequest); ok {
						if contactMap["uuid"] != req.UUID {
							t.Errorf("Expected contact UUID %s, got %v", req.UUID, contactMap["uuid"])
						}
						if contactMap["email"] != req.Email {
							t.Errorf("Expected contact email %s, got %v", req.Email, contactMap["email"])
						}
					}
				}
			}

			// Run additional checks
			tc.checkUpdated(t, mockService)
		})
	}
}

func TestContactHandler_HandleDelete(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactService)
		expectedStatus int
		checkDeleted   func(*testing.T, *MockContactService)
	}{
		{
			name:   "Delete Contact Success",
			method: http.MethodPost,
			reqBody: deleteContactRequest{
				UUID: "uuid1",
			},
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"uuid1": {UUID: "uuid1", Email: "test@example.com", ExternalID: "ext1", Timezone: "UTC"},
				}
			},
			expectedStatus: http.StatusOK,
			checkDeleted: func(t *testing.T, m *MockContactService) {
				if !m.DeleteContactCalled {
					t.Error("Expected DeleteContact to be called, but it wasn't")
				}
				if m.LastContactUUID != "uuid1" {
					t.Errorf("Expected contact UUID 'uuid1', got '%s'", m.LastContactUUID)
				}
			},
		},
		{
			name:   "Delete Contact Not Found",
			method: http.MethodPost,
			reqBody: deleteContactRequest{
				UUID: "nonexistent",
			},
			setupMock: func(m *MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkDeleted: func(t *testing.T, m *MockContactService) {
				if !m.DeleteContactCalled {
					t.Error("Expected DeleteContact to be called, but it wasn't")
				}
			},
		},
		{
			name:   "Delete Contact Service Error",
			method: http.MethodPost,
			reqBody: deleteContactRequest{
				UUID: "uuid1",
			},
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkDeleted: func(t *testing.T, m *MockContactService) {
				if !m.DeleteContactCalled {
					t.Error("Expected DeleteContact to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkDeleted: func(t *testing.T, m *MockContactService) {
				if m.DeleteContactCalled {
					t.Error("Expected DeleteContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Missing UUID in Request",
			method: http.MethodPost,
			reqBody: deleteContactRequest{
				UUID: "", // Empty UUID
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkDeleted: func(t *testing.T, m *MockContactService) {
				if m.DeleteContactCalled {
					t.Error("Expected DeleteContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: deleteContactRequest{
				UUID: "uuid1",
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkDeleted: func(t *testing.T, m *MockContactService) {
				if m.DeleteContactCalled {
					t.Error("Expected DeleteContact not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupContactHandlerTest()
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

			req, err := http.NewRequest(tc.method, "/api/contacts.delete", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleDelete(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
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
			tc.checkDeleted(t, mockService)
		})
	}
}
