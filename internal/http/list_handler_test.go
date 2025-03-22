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

// MockListService is a mock implementation of domain.ListService
type MockListService struct {
	lists map[string]*domain.List

	// Function call trackers
	GetListsCalled          bool
	GetListByIDCalled       bool
	CreateListCalled        bool
	UpdateListCalled        bool
	DeleteListCalled        bool
	LastListID              string
	LastListCreated         *domain.List
	LastListUpdated         *domain.List
	ErrToReturn             error
	ErrListNotFoundToReturn bool
}

func NewMockListService() *MockListService {
	return &MockListService{
		lists: make(map[string]*domain.List),
	}
}

func (m *MockListService) GetLists(ctx context.Context) ([]*domain.List, error) {
	m.GetListsCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	lists := make([]*domain.List, 0, len(m.lists))
	for _, list := range m.lists {
		lists = append(lists, list)
	}
	return lists, nil
}

func (m *MockListService) GetListByID(ctx context.Context, id string) (*domain.List, error) {
	m.GetListByIDCalled = true
	m.LastListID = id
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.ErrListNotFoundToReturn {
		return nil, &domain.ErrListNotFound{}
	}

	list, exists := m.lists[id]
	if !exists {
		return nil, &domain.ErrListNotFound{}
	}
	return list, nil
}

func (m *MockListService) CreateList(ctx context.Context, list *domain.List) error {
	m.CreateListCalled = true
	m.LastListCreated = list
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Set timestamps
	list.CreatedAt = time.Now()
	list.UpdatedAt = list.CreatedAt

	m.lists[list.ID] = list
	return nil
}

func (m *MockListService) UpdateList(ctx context.Context, list *domain.List) error {
	m.UpdateListCalled = true
	m.LastListUpdated = list
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrListNotFoundToReturn {
		return &domain.ErrListNotFound{}
	}

	_, exists := m.lists[list.ID]
	if !exists {
		return &domain.ErrListNotFound{}
	}

	// Update timestamp
	list.UpdatedAt = time.Now()

	m.lists[list.ID] = list
	return nil
}

func (m *MockListService) DeleteList(ctx context.Context, id string) error {
	m.DeleteListCalled = true
	m.LastListID = id
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrListNotFoundToReturn {
		return &domain.ErrListNotFound{}
	}

	_, exists := m.lists[id]
	if !exists {
		return &domain.ErrListNotFound{}
	}

	delete(m.lists, id)
	return nil
}

// MockLoggerForList is a mock implementation of logger.Logger for list tests
type MockLoggerForList struct {
	LoggedMessages []string
}

func (l *MockLoggerForList) Info(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "INFO: "+message)
}

func (l *MockLoggerForList) Debug(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "DEBUG: "+message)
}

func (l *MockLoggerForList) Warn(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "WARN: "+message)
}

func (l *MockLoggerForList) Error(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "ERROR: "+message)
}

func (l *MockLoggerForList) WithField(key string, value interface{}) logger.Logger {
	return l
}

func (l *MockLoggerForList) WithFields(fields map[string]interface{}) logger.Logger {
	return l
}

func (l *MockLoggerForList) Fatal(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "FATAL: "+message)
}

// Test setup helper
func setupListHandlerTest() (*MockListService, *MockLoggerForList, *ListHandler) {
	mockService := NewMockListService()
	mockLogger := &MockLoggerForList{LoggedMessages: []string{}}
	handler := NewListHandler(mockService, mockLogger)
	return mockService, mockLogger, handler
}

// Helper function to unmarshal JSON response
func decodeJSONResponse(body *bytes.Buffer, v interface{}) error {
	return json.Unmarshal(body.Bytes(), v)
}

func TestListHandler_RegisterRoutes(t *testing.T) {
	_, _, handler := setupListHandlerTest()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Check if routes were registered - indirect test by ensuring no panic
	endpoints := []string{
		"/api/lists.list",
		"/api/lists.get",
		"/api/lists.create",
		"/api/lists.update",
		"/api/lists.delete",
	}

	for _, endpoint := range endpoints {
		// This is a very basic check - just ensure the handler exists
		h, _ := mux.Handler(&http.Request{URL: &url.URL{Path: endpoint}})
		if h == nil {
			t.Errorf("Expected handler to be registered for %s, but got nil", endpoint)
		}
	}
}

func TestListHandler_HandleList(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		setupMock      func(*MockListService)
		expectedStatus int
		expectedLists  bool
	}{
		{
			name:   "Get Lists Success",
			method: http.MethodGet,
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {ID: "list1", Name: "List 1", Type: "public"},
					"list2": {ID: "list2", Name: "List 2", Type: "private"},
				}
			},
			expectedStatus: http.StatusOK,
			expectedLists:  true,
		},
		{
			name:   "Get Lists Empty Result",
			method: http.MethodGet,
			setupMock: func(m *MockListService) {
				// No lists in the mock
			},
			expectedStatus: http.StatusOK,
			expectedLists:  true,
		},
		{
			name:   "Get Lists Service Error",
			method: http.MethodGet,
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLists:  false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			setupMock: func(m *MockListService) {
				// No setup needed for this test
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedLists:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
			tc.setupMock(mockService)

			req, err := http.NewRequest(tc.method, "/api/lists.list", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleList(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v, expected %v", status, tc.expectedStatus)
			}

			if tc.expectedLists {
				if tc.expectedStatus == http.StatusOK {
					var lists []*domain.List
					if err := decodeJSONResponse(rr.Body, &lists); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					if len(lists) != len(mockService.lists) {
						t.Errorf("Expected %d lists, got %d", len(mockService.lists), len(lists))
					}
				}
			}

			if !mockService.GetListsCalled && tc.method == http.MethodGet {
				t.Error("Expected GetLists to be called, but it wasn't")
			}
		})
	}
}

func TestListHandler_HandleGet(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		listID         string
		setupMock      func(*MockListService)
		expectedStatus int
		expectedList   bool
	}{
		{
			name:   "Get List Success",
			method: http.MethodGet,
			listID: "list1",
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {ID: "list1", Name: "List 1", Type: "public"},
				}
			},
			expectedStatus: http.StatusOK,
			expectedList:   true,
		},
		{
			name:   "Get List Not Found",
			method: http.MethodGet,
			listID: "nonexistent",
			setupMock: func(m *MockListService) {
				m.ErrListNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			expectedList:   false,
		},
		{
			name:   "Get List Service Error",
			method: http.MethodGet,
			listID: "list1",
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedList:   false,
		},
		{
			name:   "Missing List ID",
			method: http.MethodGet,
			listID: "",
			setupMock: func(m *MockListService) {
				// No setup needed for this test
			},
			expectedStatus: http.StatusBadRequest,
			expectedList:   false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			listID: "list1",
			setupMock: func(m *MockListService) {
				// No setup needed for this test
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedList:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
			tc.setupMock(mockService)

			url := "/api/lists.get"
			if tc.listID != "" {
				url += "?id=" + tc.listID
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

			if tc.expectedList {
				if tc.expectedStatus == http.StatusOK {
					var response map[string]interface{}
					if err := decodeJSONResponse(rr.Body, &response); err != nil {
						t.Errorf("Failed to decode response body: %v", err)
					}

					listData, exists := response["list"]
					if !exists {
						t.Error("Expected 'list' field in response, but not found")
					}

					// Convert to map to access fields
					listMap, ok := listData.(map[string]interface{})
					if !ok {
						t.Errorf("Expected list to be a map, got %T", listData)
					} else if listMap["id"] != tc.listID {
						t.Errorf("Expected list ID %s, got %v", tc.listID, listMap["id"])
					}
				}
			}

			if tc.method == http.MethodGet && tc.listID != "" && tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				if !mockService.GetListByIDCalled {
					t.Error("Expected GetListByID to be called, but it wasn't")
				}
				if mockService.LastListID != tc.listID {
					t.Errorf("Expected ListID %s, got %s", tc.listID, mockService.LastListID)
				}
			}
		})
	}
}

func TestListHandler_HandleCreate(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockListService)
		expectedStatus int
		checkCreated   func(*testing.T, *MockListService)
	}{
		{
			name:   "Create List Success",
			method: http.MethodPost,
			reqBody: createListRequest{
				ID:            "newlist",
				Name:          "New List",
				Type:          "public",
				IsDoubleOptin: false,
				Description:   "Test list",
			},
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusCreated,
			checkCreated: func(t *testing.T, m *MockListService) {
				if !m.CreateListCalled {
					t.Error("Expected CreateList to be called, but it wasn't")
				}
				if m.LastListCreated == nil {
					t.Fatal("Expected list to be created, but it wasn't")
				}
				if m.LastListCreated.ID != "newlist" {
					t.Errorf("Expected list ID 'newlist', got '%s'", m.LastListCreated.ID)
				}
				if m.LastListCreated.Name != "New List" {
					t.Errorf("Expected list name 'New List', got '%s'", m.LastListCreated.Name)
				}
			},
		},
		{
			name:   "Create List Service Error",
			method: http.MethodPost,
			reqBody: createListRequest{
				ID:            "errorlist",
				Name:          "Error List",
				Type:          "public",
				IsDoubleOptin: false,
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkCreated: func(t *testing.T, m *MockListService) {
				if !m.CreateListCalled {
					t.Error("Expected CreateList to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkCreated: func(t *testing.T, m *MockListService) {
				if m.CreateListCalled {
					t.Error("Expected CreateList not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: createListRequest{
				ID:            "newlist",
				Name:          "New List",
				Type:          "public",
				IsDoubleOptin: false,
			},
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkCreated: func(t *testing.T, m *MockListService) {
				if m.CreateListCalled {
					t.Error("Expected CreateList not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
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

			req, err := http.NewRequest(tc.method, "/api/lists.create", &reqBody)
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
				if err := decodeJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				listData, exists := response["list"]
				if !exists {
					t.Error("Expected 'list' field in response, but not found")
				}

				// Verify the response contains the created list
				if listMap, ok := listData.(map[string]interface{}); ok {
					if req, ok := tc.reqBody.(createListRequest); ok {
						if listMap["id"] != req.ID {
							t.Errorf("Expected list ID %s, got %v", req.ID, listMap["id"])
						}
						if listMap["name"] != req.Name {
							t.Errorf("Expected list name %s, got %v", req.Name, listMap["name"])
						}
					}
				}
			}

			// Run additional checks
			tc.checkCreated(t, mockService)
		})
	}
}

func TestListHandler_HandleUpdate(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockListService)
		expectedStatus int
		checkUpdated   func(*testing.T, *MockListService)
	}{
		{
			name:   "Update List Success",
			method: http.MethodPost,
			reqBody: updateListRequest{
				ID:            "list1",
				Name:          "Updated List",
				Type:          "private",
				IsDoubleOptin: true,
				Description:   "Updated description",
			},
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {ID: "list1", Name: "List 1", Type: "public"},
				}
			},
			expectedStatus: http.StatusOK,
			checkUpdated: func(t *testing.T, m *MockListService) {
				if !m.UpdateListCalled {
					t.Error("Expected UpdateList to be called, but it wasn't")
				}
				if m.LastListUpdated == nil {
					t.Fatal("Expected list to be updated, but it wasn't")
				}
				if m.LastListUpdated.ID != "list1" {
					t.Errorf("Expected list ID 'list1', got '%s'", m.LastListUpdated.ID)
				}
				if m.LastListUpdated.Name != "Updated List" {
					t.Errorf("Expected list name 'Updated List', got '%s'", m.LastListUpdated.Name)
				}
				if m.LastListUpdated.Type != "private" {
					t.Errorf("Expected list type 'private', got '%s'", m.LastListUpdated.Type)
				}
				if !m.LastListUpdated.IsDoubleOptin {
					t.Error("Expected list IsDoubleOptin to be true, but it wasn't")
				}
				if m.LastListUpdated.Description != "Updated description" {
					t.Errorf("Expected list description 'Updated description', got '%s'", m.LastListUpdated.Description)
				}
			},
		},
		{
			name:   "Update List Not Found",
			method: http.MethodPost,
			reqBody: updateListRequest{
				ID:            "nonexistent",
				Name:          "Updated List",
				Type:          "private",
				IsDoubleOptin: true,
			},
			setupMock: func(m *MockListService) {
				m.ErrListNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkUpdated: func(t *testing.T, m *MockListService) {
				if !m.UpdateListCalled {
					t.Error("Expected UpdateList to be called, but it wasn't")
				}
			},
		},
		{
			name:   "Update List Service Error",
			method: http.MethodPost,
			reqBody: updateListRequest{
				ID:            "list1",
				Name:          "Updated List",
				Type:          "private",
				IsDoubleOptin: true,
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkUpdated: func(t *testing.T, m *MockListService) {
				if !m.UpdateListCalled {
					t.Error("Expected UpdateList to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkUpdated: func(t *testing.T, m *MockListService) {
				if m.UpdateListCalled {
					t.Error("Expected UpdateList not to be called, but it was")
				}
			},
		},
		{
			name:   "Missing ID in Request",
			method: http.MethodPost,
			reqBody: updateListRequest{
				ID:            "", // Empty ID
				Name:          "Updated List",
				Type:          "private",
				IsDoubleOptin: true,
			},
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkUpdated: func(t *testing.T, m *MockListService) {
				if m.UpdateListCalled {
					t.Error("Expected UpdateList not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: updateListRequest{
				ID:            "list1",
				Name:          "Updated List",
				Type:          "private",
				IsDoubleOptin: true,
			},
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkUpdated: func(t *testing.T, m *MockListService) {
				if m.UpdateListCalled {
					t.Error("Expected UpdateList not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
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

			req, err := http.NewRequest(tc.method, "/api/lists.update", &reqBody)
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
				if err := decodeJSONResponse(rr.Body, &response); err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}

				listData, exists := response["list"]
				if !exists {
					t.Error("Expected 'list' field in response, but not found")
				}

				// Verify the response contains the updated list
				if listMap, ok := listData.(map[string]interface{}); ok {
					if req, ok := tc.reqBody.(updateListRequest); ok {
						if listMap["id"] != req.ID {
							t.Errorf("Expected list ID %s, got %v", req.ID, listMap["id"])
						}
						if listMap["name"] != req.Name {
							t.Errorf("Expected list name %s, got %v", req.Name, listMap["name"])
						}
					}
				}
			}

			// Run additional checks
			tc.checkUpdated(t, mockService)
		})
	}
}

func TestListHandler_HandleDelete(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*MockListService)
		expectedStatus int
		checkDeleted   func(*testing.T, *MockListService)
	}{
		{
			name:   "Delete List Success",
			method: http.MethodPost,
			reqBody: deleteListRequest{
				ID: "list1",
			},
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {ID: "list1", Name: "List 1", Type: "public"},
				}
			},
			expectedStatus: http.StatusOK,
			checkDeleted: func(t *testing.T, m *MockListService) {
				if !m.DeleteListCalled {
					t.Error("Expected DeleteList to be called, but it wasn't")
				}
				if m.LastListID != "list1" {
					t.Errorf("Expected list ID 'list1', got '%s'", m.LastListID)
				}
			},
		},
		{
			name:   "Delete List Not Found",
			method: http.MethodPost,
			reqBody: deleteListRequest{
				ID: "nonexistent",
			},
			setupMock: func(m *MockListService) {
				m.ErrListNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkDeleted: func(t *testing.T, m *MockListService) {
				if !m.DeleteListCalled {
					t.Error("Expected DeleteList to be called, but it wasn't")
				}
			},
		},
		{
			name:   "Delete List Service Error",
			method: http.MethodPost,
			reqBody: deleteListRequest{
				ID: "list1",
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkDeleted: func(t *testing.T, m *MockListService) {
				if !m.DeleteListCalled {
					t.Error("Expected DeleteList to be called, but it wasn't")
				}
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkDeleted: func(t *testing.T, m *MockListService) {
				if m.DeleteListCalled {
					t.Error("Expected DeleteList not to be called, but it was")
				}
			},
		},
		{
			name:   "Missing ID in Request",
			method: http.MethodPost,
			reqBody: deleteListRequest{
				ID: "", // Empty ID
			},
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusBadRequest,
			checkDeleted: func(t *testing.T, m *MockListService) {
				if m.DeleteListCalled {
					t.Error("Expected DeleteList not to be called, but it was")
				}
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: deleteListRequest{
				ID: "list1",
			},
			setupMock: func(m *MockListService) {
				// No special setup
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkDeleted: func(t *testing.T, m *MockListService) {
				if m.DeleteListCalled {
					t.Error("Expected DeleteList not to be called, but it was")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
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

			req, err := http.NewRequest(tc.method, "/api/lists.delete", &reqBody)
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
				if err := decodeJSONResponse(rr.Body, &response); err != nil {
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
