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
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
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
func setupContactHandlerTest(t *testing.T) (*mocks.MockContactService, *MockLoggerForContact, *ContactHandler) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockContactService(ctrl)
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
	_, _, handler := setupContactHandlerTest(t)
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
		queryParams      string
		setupMock        func(*mocks.MockContactService)
		expectedStatus   int
		expectedContacts bool
	}{
		{
			name:        "Get Contacts Success",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&limit=2",
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().GetContacts(gomock.Any(), &domain.GetContactsRequest{
					WorkspaceID: "workspace123",
					Limit:       2,
				}).Return(&domain.GetContactsResponse{
					Contacts: []*domain.Contact{
						{
							Email:      "test1@example.com",
							ExternalID: &domain.NullableString{String: "ext1", IsNull: false},
							Timezone:   &domain.NullableString{String: "UTC", IsNull: false},
						},
					},
				}, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedContacts: true,
		},
		{
			name:        "Get Contacts Service Error",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&limit=2",
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().GetContacts(gomock.Any(), &domain.GetContactsRequest{
					WorkspaceID: "workspace123",
					Limit:       2,
				}).Return(nil, errors.New("service error"))
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedContacts: false,
		},
		{
			name:        "Method Not Allowed",
			method:      http.MethodPost,
			queryParams: "workspace_id=workspace123&limit=2",
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedContacts: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockContactService(ctrl)
			mockLogger := &MockLoggerForContact{LoggedMessages: []string{}}

			handler := NewContactHandler(mockService, mockLogger)

			// Setup mock expectations
			if tc.setupMock != nil {
				tc.setupMock(mockService)
			}

			// Create request
			req := httptest.NewRequest(tc.method, "/api/contacts.list?"+tc.queryParams, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.handleList(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If we expect contacts, check the response body
			if tc.expectedContacts {
				var response domain.GetContactsResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Contacts)
			}
		})
	}
}

func TestContactHandler_HandleGet(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		contactEmail    string
		contact         *domain.Contact
		err             error
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:         "Get_Contact_Success",
			method:       "GET",
			contactEmail: "test1@example.com",
			contact: &domain.Contact{
				Email:     "test1@example.com",
				FirstName: &domain.NullableString{String: "Test", IsNull: false},
				LastName:  &domain.NullableString{String: "User", IsNull: false},
			},
			err:             nil,
			expectedStatus:  http.StatusOK,
			expectedContact: true,
		},
		{
			name:            "Get_Contact_Not_Found",
			method:          "GET",
			contactEmail:    "nonexistent@example.com",
			contact:         nil,
			err:             &domain.ErrContactNotFound{Message: "contact not found"},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:            "Get_Contact_Service_Error",
			method:          "GET",
			contactEmail:    "test1@example.com",
			contact:         nil,
			err:             errors.New("service error"),
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:            "Missing_Contact_Email",
			method:          "GET",
			contactEmail:    "",
			contact:         nil,
			err:             nil,
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:            "Method_Not_Allowed",
			method:          "POST",
			contactEmail:    "test1@example.com",
			contact:         nil,
			err:             nil,
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedContact: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockContactService(ctrl)
			mockLogger := &MockLoggerForContact{LoggedMessages: []string{}}

			handler := NewContactHandler(mockService, mockLogger)

			// Set up mock expectations only for test cases that should call the service
			if tc.method == http.MethodGet && tc.contactEmail != "" {
				mockService.EXPECT().
					GetContactByEmail(gomock.Any(), "workspace123", tc.contactEmail).
					Return(tc.contact, tc.err)
			}

			// Create request
			req := httptest.NewRequest(tc.method, "/api/contacts.get?workspace_id=workspace123&email="+tc.contactEmail, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.handleGetByEmail(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If we expect a contact, check the response body
			if tc.expectedContact {
				var response struct {
					Contact *domain.Contact `json:"contact"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Contact)
				assert.Equal(t, tc.contactEmail, response.Contact.Email)
			}
		})
	}
}

func TestContactHandler_HandleGetByExternalID(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		externalID      string
		setupMock       func(*mocks.MockContactService)
		expectedStatus  int
		expectedContact bool
	}{
		{
			name:       "Get Contact By External ID Success",
			method:     http.MethodGet,
			externalID: "ext1",
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().
					GetContactByExternalID(gomock.Any(), "workspace123", "ext1").
					Return(&domain.Contact{
						Email: "test@example.com",
						ExternalID: &domain.NullableString{
							String: "ext1",
							IsNull: false,
						},
						Timezone: &domain.NullableString{
							String: "UTC",
							IsNull: false,
						},
					}, nil)
			},
			expectedStatus:  http.StatusOK,
			expectedContact: true,
		},
		{
			name:       "Get Contact By External ID Not Found",
			method:     http.MethodGet,
			externalID: "nonexistent",
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().
					GetContactByExternalID(gomock.Any(), "workspace123", "nonexistent").
					Return(nil, &domain.ErrContactNotFound{Message: "contact not found"})
			},
			expectedStatus:  http.StatusNotFound,
			expectedContact: false,
		},
		{
			name:       "Get Contact By External ID Service Error",
			method:     http.MethodGet,
			externalID: "error",
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().
					GetContactByExternalID(gomock.Any(), "workspace123", "error").
					Return(nil, errors.New("service error"))
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedContact: false,
		},
		{
			name:       "Missing External ID",
			method:     http.MethodGet,
			externalID: "",
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedContact: false,
		},
		{
			name:       "Method Not Allowed",
			method:     http.MethodPost,
			externalID: "ext1",
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedContact: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockContactService(ctrl)
			mockLogger := &MockLoggerForContact{LoggedMessages: []string{}}

			handler := NewContactHandler(mockService, mockLogger)

			// Setup mock expectations
			if tc.setupMock != nil {
				tc.setupMock(mockService)
			}

			// Create request
			req := httptest.NewRequest(tc.method, "/api/contacts.getByExternalID?workspace_id=workspace123&external_id="+tc.externalID, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.handleGetByExternalID(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If we expect a contact, check the response body
			if tc.expectedContact {
				var response struct {
					Contact *domain.Contact `json:"contact"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotNil(t, response.Contact)
				assert.Equal(t, tc.externalID, response.Contact.ExternalID.String)
			}
		})
	}
}

func TestContactHandler_HandleDelete(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		reqBody         interface{}
		setupMock       func(*mocks.MockContactService)
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:   "Delete Contact Success",
			method: http.MethodPost,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().DeleteContact(gomock.Any(), "workspace123", "test@example.com").Return(nil)
			},
			expectedStatus:  http.StatusOK,
			expectedMessage: "Contact deleted successfully",
		},
		{
			name:   "Delete Contact Not Found",
			method: http.MethodPost,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "nonexistent@example.com",
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().DeleteContact(gomock.Any(), "workspace123", "nonexistent@example.com").Return(&domain.ErrContactNotFound{Message: "contact not found"})
			},
			expectedStatus:  http.StatusNotFound,
			expectedMessage: "Contact not found",
		},
		{
			name:   "Delete Contact Service Error",
			method: http.MethodPost,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "error@example.com",
			},
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().DeleteContact(gomock.Any(), "workspace123", "error@example.com").Return(errors.New("service error"))
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Failed to delete contact",
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid request body",
		},
		{
			name:   "Missing Email in Request",
			method: http.MethodPost,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "",
			},
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "email is required",
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.DeleteContactRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
			},
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed for this test
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedMessage: "Method not allowed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockContactService(ctrl)
			mockLogger := &MockLoggerForContact{LoggedMessages: []string{}}

			handler := NewContactHandler(mockService, mockLogger)

			// Setup mock expectations
			if tc.setupMock != nil {
				tc.setupMock(mockService)
			}

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/contacts.delete", &reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleDelete(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check response body
			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			} else {
				var response map[string]string
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMessage, response["error"])
			}
		})
	}
}

func TestContactHandler_HandleImport(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		reqBody         interface{}
		setupMock       func(*mocks.MockContactService)
		expectedStatus  int
		expectedMessage string
		expectedCount   int
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
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().BatchImportContacts(gomock.Any(), "workspace123", gomock.Any()).Return(nil)
			},
			expectedStatus:  http.StatusOK,
			expectedMessage: "Successfully imported 2 contacts",
			expectedCount:   2,
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
			setupMock: func(m *mocks.MockContactService) {
				m.EXPECT().BatchImportContacts(gomock.Any(), "workspace123", gomock.Any()).Return(errors.New("service error"))
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Failed to import contacts",
			expectedCount:   0,
		},
		{
			name:   "invalid request - empty contacts",
			method: http.MethodPost,
			reqBody: map[string]interface{}{
				"workspace_id": "workspace123",
				"contacts":     []map[string]interface{}{},
			},
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "contacts array is empty",
			expectedCount:   0,
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
			setupMock: func(m *mocks.MockContactService) {
				// No setup needed
			},
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedMessage: "Method not allowed",
			expectedCount:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockContactService(ctrl)
			mockLogger := &MockLoggerForContact{LoggedMessages: []string{}}

			handler := NewContactHandler(mockService, mockLogger)

			// Setup mock expectations
			if tc.setupMock != nil {
				tc.setupMock(mockService)
			}

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/contacts.import", &reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleImport(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Equal(t, tc.expectedMessage, response["message"])
				assert.Equal(t, float64(tc.expectedCount), response["count"])
			}
		})
	}
}
