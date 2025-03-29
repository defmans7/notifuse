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
	"github.com/stretchr/testify/assert"
)

func TestContactListHandler_RegisterRoutes(t *testing.T) {
	mockService := &service.MockContactListService{}
	mockLogger := &MockLoggerForContact{}
	handler := NewContactListHandler(mockService, mockLogger)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Check if routes were registered
	endpoints := []string{
		"/api/contactLists.addContact",
		"/api/contactLists.getByIDs",
		"/api/contactLists.getContactsByList",
		"/api/contactLists.getListsByContact",
		"/api/contactLists.updateStatus",
		"/api/contactLists.removeContact",
	}

	for _, endpoint := range endpoints {
		h, _ := mux.Handler(&http.Request{URL: &url.URL{Path: endpoint}})
		if h == nil {
			t.Errorf("Expected handler to be registered for %s, but got nil", endpoint)
		}
	}
}

func TestContactListHandler_HandleAddContact(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*service.MockContactListService)
		expectedStatus int
		checkResult    func(*testing.T, *service.MockContactListService)
	}{
		{
			name:   "Success",
			method: http.MethodPost,
			reqBody: domain.AddContactToListRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
				Status:      "active",
			},
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = nil
			},
			expectedStatus: http.StatusCreated,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.AddContactToListCalled)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *service.MockContactListService) {
				m.AddContactToListCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.AddContactToListCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.AddContactToListRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
				Status:      "active",
			},
			setupMock: func(m *service.MockContactListService) {
				m.AddContactToListCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.AddContactToListCalled)
			},
		},
		{
			name:   "Service Error",
			method: http.MethodPost,
			reqBody: domain.AddContactToListRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
				Status:      "active",
			},
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.AddContactToListCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockContactListService{}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactListHandler(mockService, mockLogger)

			tt.setupMock(mockService)

			var reqBody bytes.Buffer
			if tt.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tt.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/contactLists.addContact", &reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleAddContact(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotNil(t, response["contact_list"])
			}

			tt.checkResult(t, mockService)
		})
	}
}

func TestContactListHandler_HandleGetByIDs(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryParams    string
		setupMock      func(*service.MockContactListService)
		expectedStatus int
		checkResult    func(*testing.T, *service.MockContactListService)
	}{
		{
			name:        "Success",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&email=test@example.com&list_id=list123",
			setupMock: func(m *service.MockContactListService) {
				m.ContactList = &domain.ContactList{
					Email:  "test@example.com",
					ListID: "list123",
					Status: "active",
				}
				m.ErrToReturn = nil
			},
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.GetContactListByIDsCalled)
			},
		},
		{
			name:        "Missing Required Parameters",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123",
			setupMock: func(m *service.MockContactListService) {
				m.GetContactListByIDsCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.GetContactListByIDsCalled)
			},
		},
		{
			name:        "Method Not Allowed",
			method:      http.MethodPost,
			queryParams: "workspace_id=workspace123&email=test@example.com&list_id=list123",
			setupMock: func(m *service.MockContactListService) {
				m.GetContactListByIDsCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.GetContactListByIDsCalled)
			},
		},
		{
			name:        "Service Error",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&email=test@example.com&list_id=list123",
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.GetContactListByIDsCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockContactListService{}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactListHandler(mockService, mockLogger)

			tt.setupMock(mockService)

			req := httptest.NewRequest(tt.method, "/api/contactLists.getByIDs?"+tt.queryParams, nil)
			rr := httptest.NewRecorder()
			handler.handleGetByIDs(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotNil(t, response["contact_list"])
			}

			tt.checkResult(t, mockService)
		})
	}
}

func TestContactListHandler_HandleGetContactsByList(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryParams    string
		setupMock      func(*service.MockContactListService)
		expectedStatus int
		checkResult    func(*testing.T, *service.MockContactListService)
	}{
		{
			name:        "Success",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&list_id=list123",
			setupMock: func(m *service.MockContactListService) {
				m.ContactLists = []*domain.ContactList{
					{
						Email:  "test1@example.com",
						ListID: "list123",
						Status: "active",
					},
					{
						Email:  "test2@example.com",
						ListID: "list123",
						Status: "active",
					},
				}
				m.ErrToReturn = nil
			},
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.GetContactsByListCalled)
			},
		},
		{
			name:        "Missing Required Parameters",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123",
			setupMock: func(m *service.MockContactListService) {
				m.GetContactsByListCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.GetContactsByListCalled)
			},
		},
		{
			name:        "Method Not Allowed",
			method:      http.MethodPost,
			queryParams: "workspace_id=workspace123&list_id=list123",
			setupMock: func(m *service.MockContactListService) {
				m.GetContactsByListCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.GetContactsByListCalled)
			},
		},
		{
			name:        "Service Error",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&list_id=list123",
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.GetContactsByListCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockContactListService{}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactListHandler(mockService, mockLogger)

			tt.setupMock(mockService)

			req := httptest.NewRequest(tt.method, "/api/contactLists.getContactsByList?"+tt.queryParams, nil)
			rr := httptest.NewRecorder()
			handler.handleGetContactsByList(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotNil(t, response["contact_lists"])
			}

			tt.checkResult(t, mockService)
		})
	}
}

func TestContactListHandler_HandleGetListsByContact(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryParams    string
		setupMock      func(*service.MockContactListService)
		expectedStatus int
		checkResult    func(*testing.T, *service.MockContactListService)
	}{
		{
			name:        "Success",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&email=test@example.com",
			setupMock: func(m *service.MockContactListService) {
				m.ContactLists = []*domain.ContactList{
					{
						Email:  "test@example.com",
						ListID: "list1",
						Status: "active",
					},
					{
						Email:  "test@example.com",
						ListID: "list2",
						Status: "active",
					},
				}
				m.ErrToReturn = nil
			},
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.GetListsByEmailCalled)
			},
		},
		{
			name:        "Missing Required Parameters",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123",
			setupMock: func(m *service.MockContactListService) {
				m.GetListsByEmailCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.GetListsByEmailCalled)
			},
		},
		{
			name:        "Method Not Allowed",
			method:      http.MethodPost,
			queryParams: "workspace_id=workspace123&email=test@example.com",
			setupMock: func(m *service.MockContactListService) {
				m.GetListsByEmailCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.GetListsByEmailCalled)
			},
		},
		{
			name:        "Service Error",
			method:      http.MethodGet,
			queryParams: "workspace_id=workspace123&email=test@example.com",
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.GetListsByEmailCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockContactListService{}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactListHandler(mockService, mockLogger)

			tt.setupMock(mockService)

			req := httptest.NewRequest(tt.method, "/api/contactLists.getListsByContact?"+tt.queryParams, nil)
			rr := httptest.NewRecorder()
			handler.handleGetListsByContact(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotNil(t, response["contact_lists"])
			}

			tt.checkResult(t, mockService)
		})
	}
}

func TestContactListHandler_HandleUpdateStatus(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*service.MockContactListService)
		expectedStatus int
		checkResult    func(*testing.T, *service.MockContactListService)
	}{
		{
			name:   "Success",
			method: http.MethodPost,
			reqBody: domain.UpdateContactListStatusRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
				Status:      "unsubscribed",
			},
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = nil
			},
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.UpdateContactListCalled)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *service.MockContactListService) {
				m.UpdateContactListCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.UpdateContactListCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.UpdateContactListStatusRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
				Status:      "unsubscribed",
			},
			setupMock: func(m *service.MockContactListService) {
				m.UpdateContactListCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.UpdateContactListCalled)
			},
		},
		{
			name:   "Service Error",
			method: http.MethodPost,
			reqBody: domain.UpdateContactListStatusRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
				Status:      "unsubscribed",
			},
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.UpdateContactListCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockContactListService{}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactListHandler(mockService, mockLogger)

			tt.setupMock(mockService)

			var reqBody bytes.Buffer
			if tt.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tt.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/contactLists.updateStatus", &reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleUpdateStatus(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			}

			tt.checkResult(t, mockService)
		})
	}
}

func TestContactListHandler_HandleRemoveContact(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*service.MockContactListService)
		expectedStatus int
		checkResult    func(*testing.T, *service.MockContactListService)
	}{
		{
			name:   "Success",
			method: http.MethodPost,
			reqBody: domain.RemoveContactFromListRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
			},
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = nil
			},
			expectedStatus: http.StatusOK,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.RemoveContactFromListCalled)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *service.MockContactListService) {
				m.RemoveContactFromListCalled = false
			},
			expectedStatus: http.StatusBadRequest,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.RemoveContactFromListCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.RemoveContactFromListRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
			},
			setupMock: func(m *service.MockContactListService) {
				m.RemoveContactFromListCalled = false
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.False(t, m.RemoveContactFromListCalled)
			},
		},
		{
			name:   "Service Error",
			method: http.MethodPost,
			reqBody: domain.RemoveContactFromListRequest{
				WorkspaceID: "workspace123",
				Email:       "test@example.com",
				ListID:      "list123",
			},
			setupMock: func(m *service.MockContactListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkResult: func(t *testing.T, m *service.MockContactListService) {
				assert.True(t, m.RemoveContactFromListCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &service.MockContactListService{}
			mockLogger := &MockLoggerForContact{}
			handler := NewContactListHandler(mockService, mockLogger)

			tt.setupMock(mockService)

			var reqBody bytes.Buffer
			if tt.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tt.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/contactLists.removeContact", &reqBody)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.handleRemoveContact(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			}

			tt.checkResult(t, mockService)
		})
	}
}
