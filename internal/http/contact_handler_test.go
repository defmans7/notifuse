package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/service"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/stretchr/testify/assert"
)

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
func setupContactHandlerTest() (*service.MockContactService, *MockLoggerForContact, *ContactHandler) {
	mockService := service.NewMockContactService()
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
		mockService := &service.MockContactService{}
		mockLogger := &MockLoggerForContact{}
		handler := NewContactHandler(mockService, mockLogger)

		// Create request with query parameters
		request := httptest.NewRequest(http.MethodGet, "/api/contacts.list?workspace_id=workspace123&limit=2", nil)
		response := httptest.NewRecorder()

		// Act
		handler.handleList(response, request)

		// Assert
		assert.Equal(t, http.StatusOK, response.Code)
		assert.True(t, mockService.GetContactsCalled)
	})

	t.Run("Service_Error", func(t *testing.T) {
		// Arrange
		mockService := &service.MockContactService{}
		mockLogger := &MockLoggerForContact{}
		handler := NewContactHandler(mockService, mockLogger)

		mockService.ErrToReturn = errors.New("service error")

		// Create request with query parameters
		request := httptest.NewRequest(http.MethodGet, "/api/contacts.list?workspace_id=workspace123&limit=2", nil)
		response := httptest.NewRecorder()

		// Act
		handler.handleList(response, request)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.True(t, mockService.GetContactsCalled)
	})

	t.Run("Wrong_Method", func(t *testing.T) {
		// Arrange
		mockService := &service.MockContactService{}
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
		mockService := &service.MockContactService{}
		mockLogger := &MockLoggerForContact{}
		handler := NewContactHandler(mockService, mockLogger)

		// Create request without required workspace_id
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
		setupMock       func(*service.MockContactService)
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:         "Get Contact Success",
			method:       http.MethodGet,
			contactEmail: "test1@example.com",
			setupMock: func(m *service.MockContactService) {
				m.Contacts = map[string]*domain.Contact{
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
			setupMock: func(m *service.MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:         "Get Contact Service Error",
			method:       http.MethodGet,
			contactEmail: "test1@example.com",
			setupMock: func(m *service.MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:         "Missing Contact Email",
			method:       http.MethodGet,
			contactEmail: "",
			setupMock: func(m *service.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:         "Method Not Allowed",
			method:       http.MethodPost,
			contactEmail: "test1@example.com",
			setupMock: func(m *service.MockContactService) {
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
				url += "?workspace_id=workspace123&email=" + tc.contactEmail
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
		setupMock       func(*service.MockContactService)
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:       "Get Contact By External ID Success",
			method:     http.MethodGet,
			externalID: "ext1",
			setupMock: func(m *service.MockContactService) {
				m.Contacts = map[string]*domain.Contact{
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
			setupMock: func(m *service.MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:       "Get Contact By External ID Service Error",
			method:     http.MethodGet,
			externalID: "error",
			setupMock: func(m *service.MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:       "Missing External ID",
			method:     http.MethodGet,
			externalID: "",
			setupMock: func(m *service.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:       "Method Not Allowed",
			method:     http.MethodPost,
			externalID: "ext1",
			setupMock: func(m *service.MockContactService) {
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
				url += "?workspace_id=workspace123&external_id=" + tc.externalID
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
		name            string
		method          string
		reqBody         interface{}
		setupMock       func(*service.MockContactService)
		expectedStatus  int
		expectedMessage string
		checkDeleted    func(*testing.T, *service.MockContactService)
	}{
		{
			name:   "Delete Contact Success",
			method: http.MethodPost,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
			},
			setupMock: func(m *service.MockContactService) {
				m.Contacts = map[string]*domain.Contact{
					"test@example.com": {
						Email:      "test@example.com",
						ExternalID: &domain.NullableString{String: "ext1", IsNull: false},
						Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkDeleted: func(t *testing.T, m *service.MockContactService) {
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
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "nonexistent@example.com",
			},
			setupMock: func(m *service.MockContactService) {
				m.ErrContactNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkDeleted: func(t *testing.T, m *service.MockContactService) {
				if !m.DeleteContactCalled {
					t.Error("Expected DeleteContact to be called, but it wasn't")
				}
			},
		},
		{
			name:   "Delete Contact Service Error",
			method: http.MethodPost,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "error@example.com",
			},
			setupMock: func(m *service.MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkDeleted: func(t *testing.T, m *service.MockContactService) {
				if !m.DeleteContactCalled {
					t.Error("Expected DeleteContact to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *service.MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkDeleted: func(t *testing.T, m *service.MockContactService) {
				if m.DeleteContactCalled {
					t.Error("Expected DeleteContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Missing Email in Request",
			method: http.MethodPost,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "", // Empty Email
			},
			setupMock: func(m *service.MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkDeleted: func(t *testing.T, m *service.MockContactService) {
				if m.DeleteContactCalled {
					t.Error("Expected DeleteContact not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
			},
			setupMock: func(m *service.MockContactService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkDeleted: func(t *testing.T, m *service.MockContactService) {
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
					return
				}

				success, exists := response["success"]
				assert.True(t, exists, "Expected 'success' field in response")
				assert.True(t, success.(bool), "Expected 'success' to be true")
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
		setupMock       func(*service.MockContactService)
		expectedStatus  int
		expectedMessage string
		checkImported   func(*testing.T, *service.MockContactService)
	}{
		{
			name:   "successful batch import",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contacts": []map[string]interface{}{
					{
						"email":       "contact1@example.com",
						"external_id": "ext1",
						"timezone":    "UTC",
					},
					{
						"email":       "contact2@example.com",
						"external_id": "ext2",
						"timezone":    "UTC",
					},
				},
			},
			setupMock: func(m *service.MockContactService) {
				m.ErrToReturn = nil
			},
			expectedStatus:  http.StatusOK,
			expectedMessage: "Successfully imported 2 contacts",
			checkImported: func(t *testing.T, m *service.MockContactService) {
				assert.Equal(t, true, m.BatchImportContactsCalled)
				assert.Equal(t, 2, len(m.LastContactsBatchImported))
			},
		},
		{
			name:   "service error",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contacts": []map[string]interface{}{
					{
						"email":       "contact1@example.com",
						"external_id": "ext1",
						"timezone":    "UTC",
					},
				},
			},
			setupMock: func(m *service.MockContactService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Failed to import contacts",
			checkImported: func(t *testing.T, m *service.MockContactService) {
				assert.Equal(t, true, m.BatchImportContactsCalled)
			},
		},
		{
			name:   "invalid request - empty contacts",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contacts":     []map[string]interface{}{},
			},
			setupMock: func(m *service.MockContactService) {
				// Should not be called
				m.BatchImportContactsCalled = false
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "contacts array is empty",
			checkImported: func(t *testing.T, m *service.MockContactService) {
				assert.Equal(t, false, m.BatchImportContactsCalled)
			},
		},
		{
			name:   "method not allowed",
			method: http.MethodGet,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contacts": []map[string]interface{}{
					{
						"email":       "contact1@example.com",
						"external_id": "ext1",
						"timezone":    "UTC",
					},
				},
			},
			setupMock: func(m *service.MockContactService) {
				// Should not be called
				m.BatchImportContactsCalled = false
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedMessage: "Method not allowed",
			checkImported: func(t *testing.T, m *service.MockContactService) {
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
					return
				}

				success, exists := response["success"]
				assert.True(t, exists, "Expected 'success' field in response")
				assert.True(t, success.(bool), "Expected 'success' to be true")

				message, exists := response["message"]
				assert.True(t, exists, "Expected 'message' field in response")
				assert.Equal(t, tc.expectedMessage, message)

				count, exists := response["count"]
				assert.True(t, exists, "Expected 'count' field in response")
				expectedCount := 0
				if tc.name == "successful batch import" {
					expectedCount = 2
				} else if tc.name == "service error" {
					expectedCount = 1
				}
				countNum, ok := count.(json.Number)
				assert.True(t, ok, "Expected count to be json.Number")
				countVal, err := countNum.Int64()
				assert.NoError(t, err, "Failed to convert count to int64")
				assert.Equal(t, int64(expectedCount), countVal)
			}

			// Run specific checks
			tc.checkImported(t, mockService)
		})
	}
}
