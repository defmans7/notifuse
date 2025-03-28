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

// MockListService is a mock implementation of domain.ListService for testing
type MockListService struct {
	lists                   map[string]*domain.List
	ErrToReturn             error
	ErrListNotFoundToReturn bool
	GetListsCalled          bool
	GetListByIDCalled       bool
	LastListID              string
	CreateListCalled        bool
	LastListCreated         *domain.List
	UpdateListCalled        bool
	LastListUpdated         *domain.List
	DeleteListCalled        bool
	LastListDeleted         string
}

// NewMockListService creates a new mock list service for testing
func NewMockListService() *MockListService {
	return &MockListService{
		lists: make(map[string]*domain.List),
	}
}

func (m *MockListService) GetLists(ctx context.Context, workspaceID string) ([]*domain.List, error) {
	m.GetListsCalled = true
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	// Convert map to slice
	lists := make([]*domain.List, 0, len(m.lists))
	for _, list := range m.lists {
		lists = append(lists, list)
	}

	return lists, nil
}

func (m *MockListService) GetListByID(ctx context.Context, workspaceID string, id string) (*domain.List, error) {
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

func (m *MockListService) CreateList(ctx context.Context, workspaceID string, list *domain.List) error {
	m.CreateListCalled = true
	m.LastListCreated = list
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	// Set timestamps
	now := time.Now()
	if list.CreatedAt.IsZero() {
		list.CreatedAt = now
	}
	list.UpdatedAt = now

	// Store the list
	m.lists[list.ID] = list
	return nil
}

func (m *MockListService) UpdateList(ctx context.Context, workspaceID string, list *domain.List) error {
	m.UpdateListCalled = true
	m.LastListUpdated = list
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}
	if m.ErrListNotFoundToReturn {
		return &domain.ErrListNotFound{}
	}

	// Check if list exists
	existingList, exists := m.lists[list.ID]
	if !exists {
		return &domain.ErrListNotFound{}
	}

	// Update fields
	existingList.Name = list.Name
	existingList.Description = list.Description
	existingList.UpdatedAt = time.Now()

	// Store the updated list
	m.lists[list.ID] = existingList
	return nil
}

func (m *MockListService) DeleteList(ctx context.Context, workspaceID string, id string) error {
	m.DeleteListCalled = true
	m.LastListDeleted = id
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

func (l *MockLoggerForList) Error(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "ERROR: "+message)
}

func (l *MockLoggerForList) Debug(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "DEBUG: "+message)
}

func (l *MockLoggerForList) Warn(message string) {
	l.LoggedMessages = append(l.LoggedMessages, "WARN: "+message)
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
		// This is a basic check - just ensure the handler exists
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
		queryParams    url.Values
		setupMock      func(*MockListService)
		expectedStatus int
		expectedLists  bool
	}{
		{
			name:   "Get Lists Success",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
			},
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {
						ID:          "list1",
						Name:        "Test List 1",
						Description: "Test Description 1",
					},
					"list2": {
						ID:          "list2",
						Name:        "Test List 2",
						Description: "Test Description 2",
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedLists:  true,
		},
		{
			name:   "Get Lists Service Error",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLists:  false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
			},
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedLists:  false,
		},
		{
			name:        "Missing Workspace ID",
			method:      http.MethodGet,
			queryParams: url.Values{},
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedLists:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/api/lists.list?"+tc.queryParams.Encode(), nil)
			rr := httptest.NewRecorder()

			handler.handleList(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response []*domain.List
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response)
			}

			if tc.method == http.MethodGet && tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				assert.True(t, mockService.GetListsCalled)
			}
		})
	}
}

func TestListHandler_HandleGet(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		queryParams    url.Values
		setupMock      func(*MockListService)
		expectedStatus int
		expectedList   bool
	}{
		{
			name:   "Get List Success",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"list1"},
			},
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {
						ID:          "list1",
						Name:        "Test List",
						Description: "Test Description",
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedList:   true,
		},
		{
			name:   "Get List Not Found",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"nonexistent"},
			},
			setupMock: func(m *MockListService) {
				m.ErrListNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			expectedList:   false,
		},
		{
			name:   "Get List Service Error",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"list1"},
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedList:   false,
		},
		{
			name:   "Missing List ID",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
			},
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedList:   false,
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodPost,
			queryParams: url.Values{
				"workspace_id": []string{"workspace123"},
				"id":           []string{"list1"},
			},
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedList:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/api/lists.get?"+tc.queryParams.Encode(), nil)
			rr := httptest.NewRecorder()

			handler.handleGet(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "list")
			}

			if tc.method == http.MethodGet && tc.expectedStatus != http.StatusMethodNotAllowed && tc.expectedStatus != http.StatusBadRequest {
				assert.True(t, mockService.GetListByIDCalled)
				if tc.queryParams.Get("id") != "" {
					assert.Equal(t, tc.queryParams.Get("id"), mockService.LastListID)
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
			reqBody: domain.CreateListRequest{
				WorkspaceID:   "workspace123",
				ID:            "list1",
				Name:          "New List",
				Type:          "public",
				IsDoubleOptin: true,
				IsPublic:      true,
				Description:   "New Description",
			},
			setupMock: func(m *MockListService) {
				// No special setup needed
			},
			expectedStatus: http.StatusCreated,
			checkCreated: func(t *testing.T, m *MockListService) {
				assert.True(t, m.CreateListCalled)
				assert.NotNil(t, m.LastListCreated)
				assert.Equal(t, "New List", m.LastListCreated.Name)
				assert.Equal(t, "New Description", m.LastListCreated.Description)
			},
		},
		{
			name:   "Create List Service Error",
			method: http.MethodPost,
			reqBody: domain.CreateListRequest{
				WorkspaceID:   "workspace123",
				ID:            "list1",
				Name:          "New List",
				Type:          "public",
				IsDoubleOptin: true,
				IsPublic:      true,
				Description:   "New Description",
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkCreated: func(t *testing.T, m *MockListService) {
				assert.True(t, m.CreateListCalled)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			checkCreated: func(t *testing.T, m *MockListService) {
				assert.False(t, m.CreateListCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.CreateListRequest{
				WorkspaceID:   "workspace123",
				ID:            "list1",
				Name:          "New List",
				Type:          "public",
				IsDoubleOptin: true,
				IsPublic:      true,
				Description:   "New Description",
			},
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkCreated: func(t *testing.T, m *MockListService) {
				assert.False(t, m.CreateListCalled)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
			tc.setupMock(mockService)

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/lists.create", &reqBody)
			rr := httptest.NewRecorder()

			handler.handleCreate(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "list")
			}

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
			reqBody: domain.UpdateListRequest{
				WorkspaceID:   "workspace123",
				ID:            "list1",
				Name:          "Updated List",
				Type:          "public",
				IsDoubleOptin: true,
				IsPublic:      true,
				Description:   "Updated Description",
			},
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {
						ID:            "list1",
						Name:          "Original List",
						Type:          "public",
						IsDoubleOptin: true,
						IsPublic:      true,
						Description:   "Original Description",
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkUpdated: func(t *testing.T, m *MockListService) {
				assert.True(t, m.UpdateListCalled)
				assert.NotNil(t, m.LastListUpdated)
				assert.Equal(t, "Updated List", m.LastListUpdated.Name)
				assert.Equal(t, "Updated Description", m.LastListUpdated.Description)
			},
		},
		{
			name:   "Update List Not Found",
			method: http.MethodPost,
			reqBody: domain.UpdateListRequest{
				WorkspaceID:   "workspace123",
				ID:            "nonexistent",
				Name:          "Updated List",
				Type:          "public",
				IsDoubleOptin: true,
				IsPublic:      true,
				Description:   "Updated Description",
			},
			setupMock: func(m *MockListService) {
				m.ErrListNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkUpdated: func(t *testing.T, m *MockListService) {
				assert.True(t, m.UpdateListCalled)
			},
		},
		{
			name:   "Update List Service Error",
			method: http.MethodPost,
			reqBody: domain.UpdateListRequest{
				WorkspaceID:   "workspace123",
				ID:            "list1",
				Name:          "Updated List",
				Type:          "public",
				IsDoubleOptin: true,
				IsPublic:      true,
				Description:   "Updated Description",
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkUpdated: func(t *testing.T, m *MockListService) {
				assert.True(t, m.UpdateListCalled)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			checkUpdated: func(t *testing.T, m *MockListService) {
				assert.False(t, m.UpdateListCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.UpdateListRequest{
				WorkspaceID:   "workspace123",
				ID:            "list1",
				Name:          "Updated List",
				Type:          "public",
				IsDoubleOptin: true,
				IsPublic:      true,
				Description:   "Updated Description",
			},
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkUpdated: func(t *testing.T, m *MockListService) {
				assert.False(t, m.UpdateListCalled)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
			tc.setupMock(mockService)

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/lists.update", &reqBody)
			rr := httptest.NewRecorder()

			handler.handleUpdate(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response, "list")
			}

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
			reqBody: domain.DeleteListRequest{
				WorkspaceID: "workspace123",
				ID:          "list1",
			},
			setupMock: func(m *MockListService) {
				m.lists = map[string]*domain.List{
					"list1": {
						ID:          "list1",
						Name:        "Test List",
						Description: "Test Description",
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkDeleted: func(t *testing.T, m *MockListService) {
				assert.True(t, m.DeleteListCalled)
				assert.Equal(t, "list1", m.LastListDeleted)
			},
		},
		{
			name:   "Delete List Not Found",
			method: http.MethodPost,
			reqBody: domain.DeleteListRequest{
				WorkspaceID: "workspace123",
				ID:          "nonexistent",
			},
			setupMock: func(m *MockListService) {
				m.ErrListNotFoundToReturn = true
			},
			expectedStatus: http.StatusNotFound,
			checkDeleted: func(t *testing.T, m *MockListService) {
				assert.True(t, m.DeleteListCalled)
			},
		},
		{
			name:   "Delete List Service Error",
			method: http.MethodPost,
			reqBody: domain.DeleteListRequest{
				WorkspaceID: "workspace123",
				ID:          "list1",
			},
			setupMock: func(m *MockListService) {
				m.ErrToReturn = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
			checkDeleted: func(t *testing.T, m *MockListService) {
				assert.True(t, m.DeleteListCalled)
			},
		},
		{
			name:    "Invalid Request Body",
			method:  http.MethodPost,
			reqBody: "invalid json",
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusBadRequest,
			checkDeleted: func(t *testing.T, m *MockListService) {
				assert.False(t, m.DeleteListCalled)
			},
		},
		{
			name:   "Method Not Allowed",
			method: http.MethodGet,
			reqBody: domain.DeleteListRequest{
				WorkspaceID: "workspace123",
				ID:          "list1",
			},
			setupMock: func(m *MockListService) {
				// No setup needed
			},
			expectedStatus: http.StatusMethodNotAllowed,
			checkDeleted: func(t *testing.T, m *MockListService) {
				assert.False(t, m.DeleteListCalled)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupListHandlerTest()
			tc.setupMock(mockService)

			var reqBody bytes.Buffer
			if tc.reqBody != nil {
				if err := json.NewEncoder(&reqBody).Encode(tc.reqBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/lists.delete", &reqBody)
			rr := httptest.NewRecorder()

			handler.handleDelete(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
			}

			tc.checkDeleted(t, mockService)
		})
	}
}
