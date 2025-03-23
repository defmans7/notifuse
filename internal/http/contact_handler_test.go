package http

/*
TODO: This file needs comprehensive rewriting to use Email as the primary
identifier for contacts instead of UUID. The current test cases use the
legacy UUID field which has been removed from the domain model.
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// MockContactService is a mock implementation of domain.ContactService for testing
type MockContactService struct {
	contacts                   map[string]*domain.Contact
	ErrToReturn                error
	ErrContactNotFoundToReturn bool

	GetContactsCalled            bool
	GetContactByEmailCalled      bool
	LastContactEmail             string
	GetContactByExternalIDCalled bool
	LastContactExternalID        string
	DeleteContactCalled          bool
	BatchImportContactsCalled    bool
	LastBatchImported            []*domain.Contact
	UpsertContactCalled          bool
	LastContactUpserted          *domain.Contact
	UpsertIsNewToReturn          bool
}

// NewMockContactService creates a new mock contact service for testing
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

func (m *MockContactService) DeleteContact(ctx context.Context, email string) error {
	m.DeleteContactCalled = true
	m.LastContactEmail = email
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrContactNotFoundToReturn {
		return &domain.ErrContactNotFound{}
	}

	for key, contact := range m.contacts {
		if contact.Email == email {
			delete(m.contacts, key)
			return nil
		}
	}
	return &domain.ErrContactNotFound{}
}

func (m *MockContactService) BatchImportContacts(ctx context.Context, contacts []*domain.Contact) error {
	m.BatchImportContactsCalled = true
	m.LastBatchImported = contacts
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Set timestamps for all contacts
	now := time.Now()
	for _, contact := range contacts {
		if contact.CreatedAt.IsZero() {
			contact.CreatedAt = now
		}
		contact.UpdatedAt = now

		// Store in the map
		m.contacts[contact.Email] = contact
	}

	return nil
}

func (m *MockContactService) UpsertContact(ctx context.Context, contact *domain.Contact) (bool, error) {
	m.UpsertContactCalled = true
	m.LastContactUpserted = contact
	if m.ErrToReturn != nil {
		return false, m.ErrToReturn
	}

	// Check if contact exists
	isNew := true
	for _, existingContact := range m.contacts {
		if existingContact.Email == contact.Email {
			isNew = false
			break
		}
	}

	// Set timestamps
	now := time.Now()
	if isNew || contact.CreatedAt.IsZero() {
		contact.CreatedAt = now
	}
	contact.UpdatedAt = now

	// Store the contact
	m.contacts[contact.Email] = contact

	return m.UpsertIsNewToReturn, nil
}

// MockLoggerForContact is a mock implementation of logger.Logger for contact tests
type MockLoggerForContact struct {
	LoggedMessages []string
}

func (l *MockLoggerForContact) Info(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "INFO: "+message)
}

func (l *MockLoggerForContact) Error(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "ERROR: "+message)
}

func (l *MockLoggerForContact) Debug(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "DEBUG: "+message)
}

func (l *MockLoggerForContact) Warn(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "WARN: "+message)
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
		"/api/contacts.delete",
		"/api/contacts.import",
		"/api/contacts.upsert",
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
			name:   "Success",
			method: http.MethodGet,
			setupMock: func(mock *MockContactService) {
				mock.contacts = map[string]*domain.Contact{
					"test@example.com": {
						Email:      "test@example.com",
						ExternalID: "ext123",
						Timezone:   "UTC",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					},
				}
			},
			expectedStatus:   http.StatusOK,
			expectedContacts: true,
		},
		{
			name:             "Service Error",
			method:           http.MethodGet,
			setupMock:        func(mock *MockContactService) { mock.ErrToReturn = errors.New("service error") },
			expectedStatus:   http.StatusInternalServerError,
			expectedContacts: false,
		},
		{
			name:             "Wrong Method",
			method:           http.MethodPost,
			setupMock:        func(mock *MockContactService) {},
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
		contactEmail    string
		setupMock       func(*MockContactService)
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:         "Get Contact Success",
			method:       http.MethodGet,
			contactEmail: "test1@example.com",
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"test1@example.com": {Email: "test1@example.com", ExternalID: "ext1", Timezone: "UTC"},
				}
			},
			expectedStatus:  http.StatusOK,
			expectedContact: true,
		},
		{
			name:         "Get Contact Not Found",
			method:       http.MethodGet,
			contactEmail: "nonexistent@example.com",
			setupMock: func(m *MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:         "Get Contact Service Error",
			method:       http.MethodGet,
			contactEmail: "test1@example.com",
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:         "Missing Contact Email",
			method:       http.MethodGet,
			contactEmail: "",
			setupMock: func(m *MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:         "Method Not Allowed",
			method:       http.MethodPost,
			contactEmail: "test1@example.com",
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
			if tc.contactEmail != "" {
				url += "?email=" + tc.contactEmail
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
					} else if contactMap["email"] != tc.contactEmail {
						t.Errorf("Expected contact email %s, got %v", tc.contactEmail, contactMap["email"])
					}
				}
			}

			if tc.method == http.MethodGet && tc.contactEmail != "" && tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetContactByEmailCalled {
					t.Error("Expected GetContactByEmail to be called, but it wasn't")
				}
				if mockService.LastContactEmail != tc.contactEmail {
					t.Errorf("Expected Email %s, got %s", tc.contactEmail, mockService.LastContactEmail)
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
					"test@example.com": {
						Email:      "test@example.com",
						ExternalID: "ext1",
						Timezone:   "UTC",
					},
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

func TestContactHandler_HandleCreate_Upsert(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactService)
		expectedStatus int
		expectedAction string
		checkCreated   func(*testing.T, *MockContactService)
	}{
		{
			name:   "Create Contact Success",
			method: http.MethodPost,
			reqBody: upsertContactRequest{
				ExternalID: "new-ext",
				Email:      "new@example.com",
				FirstName:  "John",
				LastName:   "Doe",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.UpsertIsNewToReturn = true // This is a new contact
			},
			expectedStatus: http.StatusCreated,
			expectedAction: "created",
			checkCreated: func(t *testing.T, m *MockContactService) {
				if !m.UpsertContactCalled {
					t.Error("Expected UpsertContact to be called, but it wasn't")
				}
				if m.LastContactUpserted == nil {
					t.Fatal("Expected contact to be created, but it wasn't")
				}
				if m.LastContactUpserted.Email != "new@example.com" {
					t.Errorf("Expected contact email 'new@example.com', got '%s'", m.LastContactUpserted.Email)
				}
				if m.LastContactUpserted.FirstName != "John" {
					t.Errorf("Expected contact first name 'John', got '%s'", m.LastContactUpserted.FirstName)
				}
			},
		},
		{
			name:   "Create Contact Service Error",
			method: http.MethodPost,
			reqBody: upsertContactRequest{
				ExternalID: "error-ext",
				Email:      "error@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedAction: "",
			checkCreated: func(t *testing.T, m *MockContactService) {
				if !m.UpsertContactCalled {
					t.Error("Expected UpsertContact to be called, but it wasn't")
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
			expectedAction: "",
			checkCreated: func(t *testing.T, m *MockContactService) {
				if m.UpsertContactCalled {
					t.Error("Expected UpsertContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: upsertContactRequest{
				ExternalID: "new-ext",
				Email:      "new@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAction: "",
			checkCreated: func(t *testing.T, m *MockContactService) {
				if m.UpsertContactCalled {
					t.Error("Expected UpsertContact not to be called, but it was")
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

			req, err := http.NewRequest(tc.method, "/api/contacts.upsert", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleUpsert(rr, req)

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

				action, exists := response["action"]
				if !exists {
					t.Error("Expected 'action' field in response, but not found")
				}
				if exists && tc.expectedAction != "" && action != tc.expectedAction {
					t.Errorf("Expected action '%s', got '%s'", tc.expectedAction, action)
				}

				// Verify the response contains the created contact
				if contactMap, ok := contactData.(map[string]interface{}); ok {
					if req, ok := tc.reqBody.(upsertContactRequest); ok {
						if contactMap["external_id"] != req.ExternalID {
							t.Errorf("Expected contact external_id %s, got %v", req.ExternalID, contactMap["external_id"])
						}
						if contactMap["email"] != req.Email {
							t.Errorf("Expected contact email %s, got %v", req.Email, contactMap["email"])
						}
					}
				}
			}

			// Run specific checks
			tc.checkCreated(t, mockService)
		})
	}
}

func TestContactHandler_HandleUpdate_Upsert(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactService)
		expectedStatus int
		expectedAction string
		checkUpdated   func(*testing.T, *MockContactService)
	}{
		{
			name:   "Update Contact Success",
			method: http.MethodPost,
			reqBody: upsertContactRequest{
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				FirstName:  "Updated",
				LastName:   "Name",
				Timezone:   "Europe/London",
			},
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"test@example.com": {
						Email:      "test@example.com",
						ExternalID: "ext1",
						Timezone:   "UTC",
					},
				}
				m.UpsertIsNewToReturn = false // This is an existing contact
			},
			expectedStatus: http.StatusOK,
			expectedAction: "updated",
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if !m.UpsertContactCalled {
					t.Error("Expected UpsertContact to be called, but it wasn't")
				}
				if m.LastContactUpserted == nil {
					t.Fatal("Expected contact to be updated, but it wasn't")
				}
				if m.LastContactUpserted.Email != "updated@example.com" {
					t.Errorf("Expected contact email 'updated@example.com', got '%s'", m.LastContactUpserted.Email)
				}
				if m.LastContactUpserted.FirstName != "Updated" {
					t.Errorf("Expected contact first name 'Updated', got '%s'", m.LastContactUpserted.FirstName)
				}
				if m.LastContactUpserted.Timezone != "Europe/London" {
					t.Errorf("Expected contact timezone 'Europe/London', got '%s'", m.LastContactUpserted.Timezone)
				}
			},
		},
		{
			name:   "Update Contact Service Error",
			method: http.MethodPost,
			reqBody: upsertContactRequest{
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedAction: "",
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if !m.UpsertContactCalled {
					t.Error("Expected UpsertContact to be called, but it wasn't")
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
			expectedAction: "",
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if m.UpsertContactCalled {
					t.Error("Expected UpsertContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: upsertContactRequest{
				ExternalID: "updated-ext",
				Email:      "updated@example.com",
				Timezone:   "UTC",
			},
			setupMock: func(m *MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAction: "",
			checkUpdated: func(t *testing.T, m *MockContactService) {
				if m.UpsertContactCalled {
					t.Error("Expected UpsertContact not to be called, but it was")
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

			req, err := http.NewRequest(tc.method, "/api/contacts.upsert", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleUpsert(rr, req)

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

				action, exists := response["action"]
				if !exists {
					t.Error("Expected 'action' field in response, but not found")
				}
				if exists && tc.expectedAction != "" && action != tc.expectedAction {
					t.Errorf("Expected action '%s', got '%s'", tc.expectedAction, action)
				}

				// Verify the response contains the updated contact
				if contactMap, ok := contactData.(map[string]interface{}); ok {
					if req, ok := tc.reqBody.(upsertContactRequest); ok {
						if contactMap["external_id"] != req.ExternalID {
							t.Errorf("Expected contact external_id %s, got %v", req.ExternalID, contactMap["external_id"])
						}
						if contactMap["email"] != req.Email {
							t.Errorf("Expected contact email %s, got %v", req.Email, contactMap["email"])
						}
					}
				}
			}

			// Run specific checks
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
				Email: "test@example.com",
			},
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"test@example.com": {
						Email:      "test@example.com",
						ExternalID: "ext1",
						Timezone:   "UTC",
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkDeleted: func(t *testing.T, m *MockContactService) {
				if !m.DeleteContactCalled {
					t.Error("Expected DeleteContact to be called, but it wasn't")
				}
				if m.LastContactEmail != "test@example.com" {
					t.Errorf("Expected contact Email 'test@example.com', got '%s'", m.LastContactEmail)
				}
			},
		},
		{
			name:   "Delete Contact Not Found",
			method: http.MethodPost,
			reqBody: deleteContactRequest{
				Email: "nonexistent@example.com",
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
				Email: "error@example.com",
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
			name:   "Missing Email in Request",
			method: http.MethodPost,
			reqBody: deleteContactRequest{
				Email: "", // Empty Email
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
				Email: "test@example.com",
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

			// Run specific checks
			tc.checkDeleted(t, mockService)
		})
	}
}

func TestContactHandler_HandleImport(t *testing.T) {
	mockService := &MockContactService{
		contacts: make(map[string]*domain.Contact),
	}
	mockLogger := &MockLoggerForContact{}
	handler := NewContactHandler(mockService, mockLogger)

	t.Run("successful batch import", func(t *testing.T) {
		// Setup request with valid contacts
		reqBody := `{
			"contacts": [
				{
					"external_id": "ext1",
					"email": "contact1@example.com",
					"timezone": "UTC",
					"first_name": "John",
					"last_name": "Doe"
				},
				{
					"external_id": "ext2",
					"email": "contact2@example.com",
					"timezone": "Europe/Paris",
					"first_name": "Jane",
					"last_name": "Smith"
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/contacts.import", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Reset service state
		mockService.BatchImportContactsCalled = false
		mockService.LastBatchImported = nil
		mockService.ErrToReturn = nil

		// Execute
		handler.handleImport(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Successfully imported contacts", resp["message"])
		assert.Equal(t, float64(2), resp["count"])

		// Verify service was called with contacts
		assert.True(t, mockService.BatchImportContactsCalled)
		assert.Len(t, mockService.LastBatchImported, 2)
		assert.Equal(t, "contact1@example.com", mockService.LastBatchImported[0].Email)
		assert.Equal(t, "contact2@example.com", mockService.LastBatchImported[1].Email)
	})

	t.Run("invalid request method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/contacts.import", nil)
		w := httptest.NewRecorder()

		handler.handleImport(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		reqBody := `{ "contacts": [ invalid json ] }`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts.import", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleImport(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("batch size exceeds limit", func(t *testing.T) {
		// Create a request with 51 contacts (exceeding the limit of 50)
		contacts := make([]map[string]string, 51)
		for i := 0; i < 51; i++ {
			contacts[i] = map[string]string{
				"external_id": fmt.Sprintf("ext%d", i),
				"email":       fmt.Sprintf("contact%d@example.com", i),
				"timezone":    "UTC",
				"first_name":  "John",
				"last_name":   "Doe",
			}
		}

		reqData := map[string]interface{}{
			"contacts": contacts,
		}
		reqBytes, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/api/contacts.import", bytes.NewReader(reqBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleImport(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service returns error", func(t *testing.T) {
		reqBody := `{
			"contacts": [
				{
					"external_id": "ext1",
					"email": "contact1@example.com",
					"timezone": "UTC",
					"first_name": "John",
					"last_name": "Doe"
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/contacts.import", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Set service to return error
		mockService.ErrToReturn = errors.New("service error")

		handler.handleImport(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestContactHandler_HandleUpsert(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockContactService)
		expectedStatus int
		expectedAction string
		checkResult    func(*testing.T, *MockContactService)
	}{
		{
			name:   "Create Contact Without UUID",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"external_id": "new-ext",
				"email":       "new@example.com",
				"first_name":  "John",
				"last_name":   "Doe",
				"timezone":    "UTC",
			},
			setupMock: func(m *MockContactService) {
				// Reset mock state
				m.UpsertContactCalled = false
				m.LastContactUpserted = nil
				m.ErrToReturn = nil
				m.UpsertIsNewToReturn = true // Indicate this is a new contact
			},
			expectedStatus: http.StatusCreated,
			expectedAction: "created",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
				assert.NotNil(t, m.LastContactUpserted)
				assert.Equal(t, "new@example.com", m.LastContactUpserted.Email)
			},
		},
		{
			name:   "Create Contact With Email",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"external_id": "new-ext",
				"email":       "new@example.com",
				"first_name":  "John",
				"last_name":   "Doe",
				"timezone":    "UTC",
			},
			setupMock: func(m *MockContactService) {
				// Reset mock state
				m.UpsertContactCalled = false
				m.LastContactUpserted = nil
				m.ErrToReturn = nil
				m.UpsertIsNewToReturn = true // Indicate this is a new contact
			},
			expectedStatus: http.StatusCreated,
			expectedAction: "created",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
				assert.NotNil(t, m.LastContactUpserted)
				assert.Equal(t, "new@example.com", m.LastContactUpserted.Email)
			},
		},
		{
			name:   "Update Existing Contact",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"external_id": "updated-ext",
				"email":       "updated@example.com",
				"first_name":  "Updated",
				"last_name":   "Name",
				"timezone":    "Europe/London",
			},
			setupMock: func(m *MockContactService) {
				// Reset mock state
				m.UpsertContactCalled = false
				m.LastContactUpserted = nil
				m.ErrToReturn = nil
				m.UpsertIsNewToReturn = false // Indicate this is an existing contact

				// Add existing contact
				m.contacts["old@example.com"] = &domain.Contact{
					ExternalID: "old-ext",
					Email:      "old@example.com",
					FirstName:  "Old",
					LastName:   "Name",
					Timezone:   "UTC",
				}
			},
			expectedStatus: http.StatusOK,
			expectedAction: "updated",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
				assert.NotNil(t, m.LastContactUpserted)
				assert.Equal(t, "updated@example.com", m.LastContactUpserted.Email)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockContactService) {
				m.UpsertContactCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			expectedAction: "",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.False(t, m.UpsertContactCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: map[string]interface{}{
				"external_id": "updated-ext",
				"email":       "updated@example.com",
				"timezone":    "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.UpsertContactCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedAction: "",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.False(t, m.UpsertContactCalled)
			},
		},
		{
			name:   "Service Error on Upsert",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"external_id": "ext1",
				"email":       "test@example.com",
				"timezone":    "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.UpsertContactCalled = false
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedAction: "",
			checkResult: func(t *testing.T, m *MockContactService) {
				assert.True(t, m.UpsertContactCalled)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockContactService{
				contacts: make(map[string]*domain.Contact),
			}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactHandler(mockService, mockLogger)

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

			req := httptest.NewRequest(tc.method, "/api/contacts.upsert", &reqBody)
			if err := req.ParseForm(); err != nil {
				t.Fatalf("Failed to parse form: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleUpsert(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check response body for success cases
			if tc.expectedStatus == http.StatusOK || tc.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check action field
				action, exists := response["action"]
				assert.True(t, exists)
				assert.Equal(t, tc.expectedAction, action)

				// Check contact exists
				_, exists = response["contact"]
				assert.True(t, exists)
			}

			// Run specific checks
			tc.checkResult(t, mockService)
		})
	}
}

// CreateContact and UpdateContact stubs are redirecting to UpsertContact and were previously defined
// in the HTTP test file as stubs. We need to keep these stubs to maintain compatibility with existing tests,
// but they should just call UpsertContact.
