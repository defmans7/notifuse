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
	LastContactsBatchImported    []*domain.Contact
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

func (m *MockContactService) GetContacts(ctx context.Context, req *domain.GetContactsRequest) (*domain.GetContactsResponse, error) {
	m.GetContactsCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	// Convert map to slice
	contacts := make([]*domain.Contact, 0, len(m.contacts))
	for _, contact := range m.contacts {
		contacts = append(contacts, contact)
	}

	// For testing purposes, we'll just return all contacts
	// In a real implementation, we would handle pagination and filtering
	return &domain.GetContactsResponse{
		Contacts:   contacts,
		NextCursor: "", // For testing, we don't implement cursor pagination
	}, nil
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
		if contact.ExternalID.String == externalID && !contact.ExternalID.IsNull {
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
	m.LastContactsBatchImported = contacts
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
	decoder := json.NewDecoder(body)
	// Use UseNumber for more precise number handling
	decoder.UseNumber()
	return decoder.Decode(v)
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
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockService := NewMockContactService()
		mockLogger := &MockLoggerForContact{}
		handler := NewContactHandler(mockService, mockLogger)

		// Create request with query parameters
		request := httptest.NewRequest(http.MethodGet, "/api/contacts.list?workspaceId=workspace123&limit=2", nil)
		response := httptest.NewRecorder()

		// Act
		handler.handleList(response, request)

		// Assert
		assert.Equal(t, http.StatusOK, response.Code)
		assert.True(t, mockService.GetContactsCalled)
	})

	t.Run("Service_Error", func(t *testing.T) {
		// Arrange
		mockService := NewMockContactService()
		mockLogger := &MockLoggerForContact{}
		handler := NewContactHandler(mockService, mockLogger)

		mockService.ErrToReturn = errors.New("service error")

		// Create request with query parameters
		request := httptest.NewRequest(http.MethodGet, "/api/contacts.list?workspaceId=workspace123&limit=2", nil)
		response := httptest.NewRecorder()

		// Act
		handler.handleList(response, request)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.True(t, mockService.GetContactsCalled)
	})

	t.Run("Wrong_Method", func(t *testing.T) {
		// Arrange
		mockService := NewMockContactService()
		mockLogger := &MockLoggerForContact{}
		handler := NewContactHandler(mockService, mockLogger)

		request := httptest.NewRequest(http.MethodPost, "/api/contacts.list", nil)
		response := httptest.NewRecorder()

		// Act
		handler.handleList(response, request)

		// Assert
		assert.Equal(t, http.StatusMethodNotAllowed, response.Code)
		assert.False(t, mockService.GetContactsCalled)
	})

	t.Run("Invalid_Request", func(t *testing.T) {
		// Arrange
		mockService := NewMockContactService()
		mockLogger := &MockLoggerForContact{}
		handler := NewContactHandler(mockService, mockLogger)

		// Create request without required workspaceId
		request := httptest.NewRequest(http.MethodGet, "/api/contacts.list", nil)
		response := httptest.NewRecorder()

		// Act
		handler.handleList(response, request)

		// Assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.False(t, mockService.GetContactsCalled)
	})
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
					"test1@example.com": {
						Email:      "test1@example.com",
						ExternalID: domain.NullableString{String: "ext1", IsNull: false},
						Timezone:   domain.NullableString{String: "UTC", IsNull: false},
					},
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
						Email: "test@example.com",
						ExternalID: domain.NullableString{
							String: "ext1",
							IsNull: false,
						},
						Timezone: domain.NullableString{
							String: "UTC",
							IsNull: false,
						},
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
			externalID: "error",
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
					if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
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
					} else {
						// Check external_id field - could be a string in the response
						externalID, ok := contactMap["external_id"]
						if !ok {
							t.Error("Expected external_id field in contact, but not found")
						} else {
							// Get the value regardless of format
							var externalIDValue string
							switch v := externalID.(type) {
							case string:
								externalIDValue = v
							case map[string]interface{}:
								externalIDValue, _ = v["String"].(string)
							}

							if externalIDValue != tc.externalID {
								t.Errorf("Expected external_id %s, got %s", tc.externalID, externalIDValue)
							}
						}
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
			reqBody: map[string]interface{}{
				"external_id": "new-ext",
				"email":       "new@example.com",
				"first_name":  "John",
				"last_name":   "Doe",
				"timezone":    "UTC",
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
				if m.LastContactUpserted.FirstName.String != "John" || m.LastContactUpserted.FirstName.IsNull {
					t.Errorf("Expected contact first name 'John', got '%+v'", m.LastContactUpserted.FirstName)
				}
			},
		},
		{
			name:   "Create Contact Service Error",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"external_id": "error-ext",
				"email":       "error@example.com",
				"timezone":    "UTC",
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
			reqBody: map[string]interface{}{
				"external_id": "new-ext",
				"email":       "new@example.com",
				"timezone":    "UTC",
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
					if req, ok := tc.reqBody.(map[string]interface{}); ok {
						// Check the external_id which could be a string or a map
						externalID := contactMap["external_id"]
						reqExternalID := req["external_id"]

						switch val := externalID.(type) {
						case string:
							if val != reqExternalID {
								t.Errorf("Expected contact external_id %s, got %s", reqExternalID, val)
							}
						case map[string]interface{}:
							if val["String"] != reqExternalID {
								t.Errorf("Expected contact external_id %s, got %v", reqExternalID, val["String"])
							}
						}

						if contactMap["email"] != req["email"] {
							t.Errorf("Expected contact email %s, got %v", req["email"], contactMap["email"])
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
			reqBody: map[string]interface{}{
				"external_id": "updated-ext",
				"email":       "updated@example.com",
				"timezone":    "UTC",
			},
			setupMock: func(m *MockContactService) {
				m.contacts = map[string]*domain.Contact{
					"test@example.com": {
						Email:      "test@example.com",
						ExternalID: domain.NullableString{String: "ext1", IsNull: false},
						Timezone:   domain.NullableString{String: "UTC", IsNull: false},
						FirstName: domain.NullableString{
							String: "Old",
							IsNull: false,
						},
						LastName: domain.NullableString{
							String: "Name",
							IsNull: false,
						},
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
				if m.LastContactUpserted.Timezone.String != "UTC" || m.LastContactUpserted.Timezone.IsNull {
					t.Errorf("Expected contact timezone 'UTC', got '%s'", m.LastContactUpserted.Timezone.String)
				}
			},
		},
		{
			name:   "Update Contact Service Error",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"external_id": "updated-ext",
				"email":       "updated@example.com",
				"timezone":    "UTC",
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
			reqBody: map[string]interface{}{
				"external_id": "updated-ext",
				"email":       "updated@example.com",
				"timezone":    "UTC",
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
					if req, ok := tc.reqBody.(map[string]interface{}); ok {
						// Check external_id which could be a string or a map
						externalID := contactMap["external_id"]
						reqExternalID := req["external_id"]

						switch val := externalID.(type) {
						case string:
							if val != reqExternalID {
								t.Errorf("Expected contact external_id %s, got %s", reqExternalID, val)
							}
						case map[string]interface{}:
							if val["String"] != reqExternalID {
								t.Errorf("Expected contact external_id %s, got %v", reqExternalID, val["String"])
							}
						}

						if contactMap["email"] != req["email"] {
							t.Errorf("Expected contact email %s, got %v", req["email"], contactMap["email"])
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
						ExternalID: domain.NullableString{String: "ext1", IsNull: false},
						Timezone:   domain.NullableString{String: "UTC", IsNull: false},
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
	testCases := []struct {
		name            string
		method          string
		reqBody         interface{}
		setupMock       func(*MockContactService)
		expectedStatus  int
		expectedMessage string
		checkImported   func(*testing.T, *MockContactService)
	}{
		{
			name:   "successful batch import",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"contacts": []map[string]interface{}{
					{
						"external_id": map[string]interface{}{
							"String": "ext1",
							"IsNull": false,
						},
						"email": "contact1@example.com",
						"timezone": map[string]interface{}{
							"String": "UTC",
							"IsNull": false,
						},
					},
					{
						"external_id": map[string]interface{}{
							"String": "ext2",
							"IsNull": false,
						},
						"email": "contact2@example.com",
						"timezone": map[string]interface{}{
							"String": "UTC",
							"IsNull": false,
						},
					},
				},
			},
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = nil
			},
			expectedStatus:  http.StatusOK,
			expectedMessage: "Successfully imported 2 contacts",
			checkImported: func(t *testing.T, m *MockContactService) {
				assert.Equal(t, true, m.BatchImportContactsCalled)
				assert.Equal(t, 2, len(m.LastContactsBatchImported))
			},
		},
		{
			name:   "service error",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"contacts": []map[string]interface{}{
					{
						"external_id": map[string]interface{}{
							"String": "ext1",
							"IsNull": false,
						},
						"email": "contact1@example.com",
						"timezone": map[string]interface{}{
							"String": "UTC",
							"IsNull": false,
						},
					},
				},
			},
			setupMock: func(m *MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Failed to import contacts",
			checkImported: func(t *testing.T, m *MockContactService) {
				assert.Equal(t, true, m.BatchImportContactsCalled)
			},
		},
		{
			name:   "invalid request - empty contacts",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"contacts": []map[string]interface{}{},
			},
			setupMock: func(m *MockContactService) {
				// Should not be called
				m.BatchImportContactsCalled = false
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "No contacts provided in request",
			checkImported: func(t *testing.T, m *MockContactService) {
				assert.Equal(t, false, m.BatchImportContactsCalled)
			},
		},
		{
			name:   "method not allowed",
			method: http.MethodGet,
			reqBody: map[string]interface{}{
				"contacts": []map[string]interface{}{
					{
						"external_id": map[string]interface{}{
							"String": "ext1",
							"IsNull": false,
						},
						"email": "contact1@example.com",
						"timezone": map[string]interface{}{
							"String": "UTC",
							"IsNull": false,
						},
					},
				},
			},
			setupMock: func(m *MockContactService) {
				// Should not be called
				m.BatchImportContactsCalled = false
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedMessage: "Method not allowed",
			checkImported: func(t *testing.T, m *MockContactService) {
				assert.Equal(t, false, m.BatchImportContactsCalled)
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

			req, err := http.NewRequest(tc.method, "/api/contacts.import", &reqBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleImport(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				message, exists := response["message"]
				assert.True(t, exists)
				assert.Equal(t, tc.expectedMessage, message)
			}

			// Run specific checks
			tc.checkImported(t, mockService)
		})
	}
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
				"timezone":    "UTC",
			},
			setupMock: func(m *MockContactService) {
				// Reset mock state
				m.UpsertContactCalled = false
				m.LastContactUpserted = nil
				m.ErrToReturn = nil
				m.UpsertIsNewToReturn = false // Indicate this is an existing contact

				// Add existing contact
				m.contacts["old@example.com"] = &domain.Contact{
					ExternalID: domain.NullableString{String: "old-ext", IsNull: false},
					Email:      "old@example.com",
					Timezone:   domain.NullableString{String: "UTC", IsNull: false},
					FirstName: domain.NullableString{
						String: "Old",
						IsNull: false,
					},
					LastName: domain.NullableString{
						String: "Name",
						IsNull: false,
					},
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

func TestContactHandler_HandleUpsertWithCustomJSON(t *testing.T) {
	mockService, _, handler := setupContactHandlerTest()

	// Test case 1: Successful upsert with custom JSON fields
	reqBody := `{
		"email": "test@example.com",
		"external_id": "ext123",
		"timezone": "Europe/Paris",
		"language": "en-US",
		"custom_json_1": {"key": "value1"},
		"custom_json_2": null,
		"custom_json_3": {"key": "value3"}
	}`

	req, err := http.NewRequest(http.MethodPost, "/api/contacts.upsert", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler.handleUpsert(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, http.StatusOK)
	}

	// Verify that the service was called with the correct contact
	if !mockService.UpsertContactCalled {
		t.Error("Expected UpsertContact to be called, but it wasn't")
	}

	if mockService.LastContactUpserted == nil {
		t.Error("Expected LastContactUpserted to be set, but it wasn't")
	} else {
		if mockService.LastContactUpserted.Email != "test@example.com" {
			t.Errorf("Expected contact email %s, got %s", "test@example.com", mockService.LastContactUpserted.Email)
		}

		// Check external_id regardless of how it's stored internally
		expectedExternalId := "ext123"
		actualExternalId := mockService.LastContactUpserted.ExternalID.String
		if actualExternalId != expectedExternalId || mockService.LastContactUpserted.ExternalID.IsNull {
			t.Errorf("Expected contact external_id %s, got %v", expectedExternalId, mockService.LastContactUpserted.ExternalID)
		}

		// Check timezone regardless of how it's stored internally
		expectedTimezone := "Europe/Paris"
		actualTimezone := mockService.LastContactUpserted.Timezone.String
		if actualTimezone != expectedTimezone || mockService.LastContactUpserted.Timezone.IsNull {
			t.Errorf("Expected contact timezone %s, got %v", expectedTimezone, mockService.LastContactUpserted.Timezone)
		}

		if mockService.LastContactUpserted.Language.String != "en-US" || mockService.LastContactUpserted.Language.IsNull {
			t.Errorf("Expected contact language %s, got %v", "en-US", mockService.LastContactUpserted.Language)
		}

		// Verify custom JSON fields
		if !mockService.LastContactUpserted.CustomJSON1.Valid {
			t.Error("Expected CustomJSON1 to be valid")
		}
		if mockService.LastContactUpserted.CustomJSON2.Valid {
			t.Error("Expected CustomJSON2 to be invalid")
		}
		if !mockService.LastContactUpserted.CustomJSON3.Valid {
			t.Error("Expected CustomJSON3 to be valid")
		}
	}

	// Verify response
	var response map[string]interface{}
	if err := decodeContactJSONResponse(rr.Body, &response); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	contact, ok := response["contact"].(map[string]interface{})
	if !ok {
		t.Error("Expected 'contact' field in response, but not found")
	}

	action, ok := response["action"].(string)
	if !ok {
		t.Error("Expected 'action' field in response, but not found")
	}

	if mockService.UpsertIsNewToReturn {
		if action != "created" {
			t.Errorf("Expected action 'created', got %s", action)
		}
	} else {
		if action != "updated" {
			t.Errorf("Expected action 'updated', got %s", action)
		}
	}

	// Verify contact fields in response
	if contact != nil {
		if email, ok := contact["email"].(string); !ok || email != "test@example.com" {
			t.Errorf("Expected contact email %s, got %v", "test@example.com", email)
		}

		// Check external_id field - could be a string or a map
		externalID, ok := contact["external_id"]
		if !ok {
			t.Errorf("Expected external_id in response, but not found")
		} else {
			// Get the value regardless of format
			var externalIDValue string
			switch v := externalID.(type) {
			case string:
				externalIDValue = v
			case map[string]interface{}:
				externalIDValue, _ = v["String"].(string)
			}

			expectedExternalId := "ext123"
			if externalIDValue != expectedExternalId {
				t.Errorf("Expected contact external_id %s, got %v", expectedExternalId, externalIDValue)
			}
		}

		// Check timezone field - could be a string or a map
		timezone, ok := contact["timezone"]
		if !ok {
			t.Errorf("Expected timezone in response, but not found")
		} else {
			// Get the value regardless of format
			var timezoneValue string
			switch v := timezone.(type) {
			case string:
				timezoneValue = v
			case map[string]interface{}:
				timezoneValue, _ = v["String"].(string)
			}

			expectedTimezone := "Europe/Paris"
			if timezoneValue != expectedTimezone {
				t.Errorf("Expected contact timezone %s, got %v", expectedTimezone, timezoneValue)
			}
		}
	}
}

func TestContactHandler_HandleUpsertWithInvalidJSON(t *testing.T) {
	mockService := NewMockContactService()
	mockLogger := &MockLoggerForContact{}
	handler := NewContactHandler(mockService, mockLogger)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "invalid JSON syntax",
			requestBody:    `{"email": "test@example.com", "language": "en-US", invalid_json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing required email field",
			requestBody:    `{"external_id": "ext123", "language": "en-US"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "invalid-email", "language": "en-US"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty request body",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/contacts.upsert", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.handleUpsert(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
