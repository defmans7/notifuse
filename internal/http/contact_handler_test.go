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
	existingContact, exists := m.contacts[contact.Email]
	if exists {
		isNew = false
		// If updating an existing contact, merge fields
		if contact.ExternalID != nil {
			existingContact.ExternalID = contact.ExternalID
		}
		if contact.Timezone != nil {
			existingContact.Timezone = contact.Timezone
		}
		if contact.Language != nil {
			existingContact.Language = contact.Language
		}
		if contact.FirstName != nil {
			existingContact.FirstName = contact.FirstName
		}
		if contact.LastName != nil {
			existingContact.LastName = contact.LastName
		}
		if contact.Phone != nil {
			existingContact.Phone = contact.Phone
		}
		if contact.AddressLine1 != nil {
			existingContact.AddressLine1 = contact.AddressLine1
		}
		if contact.AddressLine2 != nil {
			existingContact.AddressLine2 = contact.AddressLine2
		}
		if contact.Country != nil {
			existingContact.Country = contact.Country
		}
		if contact.Postcode != nil {
			existingContact.Postcode = contact.Postcode
		}
		if contact.State != nil {
			existingContact.State = contact.State
		}
		if contact.JobTitle != nil {
			existingContact.JobTitle = contact.JobTitle
		}
		if contact.LifetimeValue != nil {
			existingContact.LifetimeValue = contact.LifetimeValue
		}
		if contact.OrdersCount != nil {
			existingContact.OrdersCount = contact.OrdersCount
		}
		if contact.LastOrderAt != nil {
			existingContact.LastOrderAt = contact.LastOrderAt
		}
		if contact.CustomString1 != nil {
			existingContact.CustomString1 = contact.CustomString1
		}
		if contact.CustomString2 != nil {
			existingContact.CustomString2 = contact.CustomString2
		}
		if contact.CustomString3 != nil {
			existingContact.CustomString3 = contact.CustomString3
		}
		if contact.CustomString4 != nil {
			existingContact.CustomString4 = contact.CustomString4
		}
		if contact.CustomString5 != nil {
			existingContact.CustomString5 = contact.CustomString5
		}
		if contact.CustomNumber1 != nil {
			existingContact.CustomNumber1 = contact.CustomNumber1
		}
		if contact.CustomNumber2 != nil {
			existingContact.CustomNumber2 = contact.CustomNumber2
		}
		if contact.CustomNumber3 != nil {
			existingContact.CustomNumber3 = contact.CustomNumber3
		}
		if contact.CustomNumber4 != nil {
			existingContact.CustomNumber4 = contact.CustomNumber4
		}
		if contact.CustomNumber5 != nil {
			existingContact.CustomNumber5 = contact.CustomNumber5
		}
		if contact.CustomDatetime1 != nil {
			existingContact.CustomDatetime1 = contact.CustomDatetime1
		}
		if contact.CustomDatetime2 != nil {
			existingContact.CustomDatetime2 = contact.CustomDatetime2
		}
		if contact.CustomDatetime3 != nil {
			existingContact.CustomDatetime3 = contact.CustomDatetime3
		}
		if contact.CustomDatetime4 != nil {
			existingContact.CustomDatetime4 = contact.CustomDatetime4
		}
		if contact.CustomDatetime5 != nil {
			existingContact.CustomDatetime5 = contact.CustomDatetime5
		}
		if contact.CustomJSON1 != nil {
			existingContact.CustomJSON1 = contact.CustomJSON1
		}
		if contact.CustomJSON2 != nil {
			existingContact.CustomJSON2 = contact.CustomJSON2
		}
		if contact.CustomJSON3 != nil {
			existingContact.CustomJSON3 = contact.CustomJSON3
		}
		if contact.CustomJSON4 != nil {
			existingContact.CustomJSON4 = contact.CustomJSON4
		}
		if contact.CustomJSON5 != nil {
			existingContact.CustomJSON5 = contact.CustomJSON5
		}

		// Set timestamps
		now := time.Now()
		existingContact.UpdatedAt = now
		contact = existingContact
	} else {
		// Set timestamps for new contact
		now := time.Now()
		if contact.CreatedAt.IsZero() {
			contact.CreatedAt = now
		}
		contact.UpdatedAt = now
	}

	// Store the contact
	m.contacts[contact.Email] = contact
	m.LastContactUpserted = contact
	return isNew, nil
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
						ExternalID: &domain.NullableString{String: "ext1", IsNull: false},
						Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
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
						ExternalID: &domain.NullableString{
							String: "ext1",
							IsNull: false,
						},
						Timezone: &domain.NullableString{
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
						ExternalID: &domain.NullableString{String: "ext1", IsNull: false},
						Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
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
