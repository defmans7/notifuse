package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	http_handler "github.com/Notifuse/notifuse/internal/http"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBroadcastService implements domain.BroadcastService for testing
type MockBroadcastService struct {
	mock.Mock
}

func (m *MockBroadcastService) CreateBroadcast(ctx context.Context, request *domain.CreateBroadcastRequest) (*domain.Broadcast, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Broadcast), args.Error(1)
}

func (m *MockBroadcastService) GetBroadcast(ctx context.Context, workspaceID, id string) (*domain.Broadcast, error) {
	args := m.Called(ctx, workspaceID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Broadcast), args.Error(1)
}

func (m *MockBroadcastService) UpdateBroadcast(ctx context.Context, request *domain.UpdateBroadcastRequest) (*domain.Broadcast, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Broadcast), args.Error(1)
}

func (m *MockBroadcastService) ListBroadcasts(ctx context.Context, params domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BroadcastListResponse), args.Error(1)
}

func (m *MockBroadcastService) ScheduleBroadcast(ctx context.Context, request *domain.ScheduleBroadcastRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockBroadcastService) PauseBroadcast(ctx context.Context, request *domain.PauseBroadcastRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockBroadcastService) ResumeBroadcast(ctx context.Context, request *domain.ResumeBroadcastRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockBroadcastService) CancelBroadcast(ctx context.Context, request *domain.CancelBroadcastRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockBroadcastService) DeleteBroadcast(ctx context.Context, request *domain.DeleteBroadcastRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockBroadcastService) SendToIndividual(ctx context.Context, request *domain.SendToIndividualRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockBroadcastService) SendWinningVariation(ctx context.Context, request *domain.SendWinningVariationRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

// MockLogger implements the logger.Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string)                                        {}
func (m *MockLogger) Debug(msg string)                                       {}
func (m *MockLogger) Warn(msg string)                                        {}
func (m *MockLogger) Error(msg string)                                       {}
func (m *MockLogger) Fatal(msg string)                                       {}
func (m *MockLogger) WithField(key string, value interface{}) logger.Logger  { return m }
func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger { return m }
func (m *MockLogger) WithError(err error) logger.Logger                      { return m }

// Helper function to create a test broadcast
func createTestBroadcast() *domain.Broadcast {
	now := time.Now()
	return &domain.Broadcast{
		ID:          "broadcast123",
		WorkspaceID: "workspace123",
		Name:        "Test Broadcast",
		Status:      domain.BroadcastStatusDraft,
		Audience: domain.AudienceSettings{
			Segments: []string{"segment123"},
		},
		Schedule: domain.ScheduleSettings{
			IsScheduled: false,
		},
		TotalSent:         100,
		TotalDelivered:    95,
		TotalFailed:       2,
		TotalBounced:      3,
		TotalComplained:   1,
		TotalOpens:        80,
		TotalClicks:       50,
		TotalUnsubscribed: 5,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// setupHandler creates a test handler and mock service
func setupHandler() (*http_handler.BroadcastHandler, *MockBroadcastService) {
	mockService := new(MockBroadcastService)
	mockTemplateService := new(mocks.MockTemplateService)
	mockLogger := &MockLogger{}

	// Create a mock public key for the handler
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Create the handler with a mock public key
	handler := http_handler.NewBroadcastHandler(mockService, mockTemplateService, publicKey, mockLogger)

	return handler, mockService
}

// TestHandleList tests the handleList function
func TestHandleList(t *testing.T) {
	handler, mockService := setupHandler()
	broadcasts := []*domain.Broadcast{createTestBroadcast()}

	// Create a response with broadcasts and a total count
	responseWithTotal := &domain.BroadcastListResponse{
		Broadcasts: broadcasts,
		TotalCount: 1,
	}

	// Test successful list
	t.Run("Success", func(t *testing.T) {
		mockService.On("ListBroadcasts", mock.Anything, mock.MatchedBy(func(params domain.ListBroadcastsParams) bool {
			return params.WorkspaceID == "workspace123" && params.Status == ""
		})).Return(responseWithTotal, nil).Once()

		// Create a test request with query parameters
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.list?workspace_id=workspace123", nil)
		w := httptest.NewRecorder()

		// Call the exported handler method directly
		handler.HandleList(w, req)

		// Verify the response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "broadcasts")
		assert.Contains(t, response, "total_count")
		assert.Equal(t, float64(1), response["total_count"]) // JSON unmarshals numbers as float64

		mockService.AssertExpectations(t)
	})

	// Test with pagination parameters
	t.Run("WithPagination", func(t *testing.T) {
		mockService.On("ListBroadcasts", mock.Anything, mock.MatchedBy(func(params domain.ListBroadcastsParams) bool {
			return params.WorkspaceID == "workspace123" &&
				params.Status == "draft" &&
				params.Limit == 10 &&
				params.Offset == 20
		})).Return(responseWithTotal, nil).Once()

		// Create a test request with query parameters including pagination
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.list?workspace_id=workspace123&status=draft&limit=10&offset=20", nil)
		w := httptest.NewRecorder()

		// Call the exported handler method directly
		handler.HandleList(w, req)

		// Verify the response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "broadcasts")
		assert.Contains(t, response, "total_count")

		mockService.AssertExpectations(t)
	})

	// Test invalid pagination parameters
	t.Run("InvalidPaginationParams", func(t *testing.T) {
		// Create a test request with invalid pagination parameters
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.list?workspace_id=workspace123&limit=invalid", nil)
		w := httptest.NewRecorder()

		// Call the exported handler method directly
		handler.HandleList(w, req)

		// Verify the response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test missing workspace_id
	t.Run("MissingWorkspaceID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.list", nil)
		w := httptest.NewRecorder()

		// Call the exported handler method directly
		handler.HandleList(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		mockService.On("ListBroadcasts", mock.Anything, mock.MatchedBy(func(params domain.ListBroadcastsParams) bool {
			return params.WorkspaceID == "workspace123" && params.Status == ""
		})).Return(nil, errors.New("service error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.list?workspace_id=workspace123", nil)
		w := httptest.NewRecorder()

		// Call the exported handler method directly
		handler.HandleList(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})

	// Test method not allowed
	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.list?workspace_id=workspace123", nil)
		w := httptest.NewRecorder()

		// Call the exported handler method directly
		handler.HandleList(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

// TestHandleGet tests the handleGet function
func TestHandleGet(t *testing.T) {
	handler, mockService := setupHandler()
	broadcast := createTestBroadcast()

	// Test successful get
	t.Run("Success", func(t *testing.T) {
		mockService.On("GetBroadcast", mock.Anything, "workspace123", "broadcast123").
			Return(broadcast, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.get?workspace_id=workspace123&id=broadcast123", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "broadcast")

		mockService.AssertExpectations(t)
	})

	// Test broadcast not found
	t.Run("BroadcastNotFound", func(t *testing.T) {
		mockService.On("GetBroadcast", mock.Anything, "workspace123", "nonexistent").
			Return(nil, &domain.ErrBroadcastNotFound{ID: "nonexistent"}).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.get?workspace_id=workspace123&id=nonexistent", nil)
		w := httptest.NewRecorder()

		handler.HandleGet(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

// TestHandleCreate tests the handleCreate function
func TestHandleCreate(t *testing.T) {
	handler, mockService := setupHandler()
	broadcast := createTestBroadcast()

	// Test successful create
	t.Run("Success", func(t *testing.T) {
		createRequest := &domain.CreateBroadcastRequest{
			WorkspaceID: "workspace123",
			Name:        "Test Broadcast",
			Audience: domain.AudienceSettings{
				Segments: []string{"segment123"},
			},
			Schedule: domain.ScheduleSettings{
				IsScheduled: false,
			},
		}

		mockService.On("CreateBroadcast", mock.Anything, mock.MatchedBy(func(req *domain.CreateBroadcastRequest) bool {
			return req.WorkspaceID == createRequest.WorkspaceID && req.Name == createRequest.Name
		})).Return(broadcast, nil).Once()

		requestBody, _ := json.Marshal(createRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.create", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleCreate(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "broadcast")

		mockService.AssertExpectations(t)
	})
}

// TestHandleSchedule tests the handleSchedule function
func TestHandleSchedule(t *testing.T) {
	handler, mockService := setupHandler()

	// Test successful scheduling for later
	t.Run("ScheduleForLater", func(t *testing.T) {
		scheduledTime := time.Now().Add(24 * time.Hour).UTC()
		scheduledDate := scheduledTime.Format("2006-01-02")
		scheduledTimeStr := scheduledTime.Format("15:04")

		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID:          "workspace123",
			ID:                   "broadcast123",
			SendNow:              false,
			ScheduledDate:        scheduledDate,
			ScheduledTime:        scheduledTimeStr,
			Timezone:             "UTC",
			UseRecipientTimezone: false,
		}

		mockService.On("ScheduleBroadcast", mock.Anything, mock.MatchedBy(func(req *domain.ScheduleBroadcastRequest) bool {
			return req.WorkspaceID == request.WorkspaceID &&
				req.ID == request.ID &&
				!req.SendNow &&
				req.ScheduledDate == scheduledDate &&
				req.ScheduledTime == scheduledTimeStr &&
				req.Timezone == "UTC"
		})).Return(nil).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.schedule", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockService.AssertExpectations(t)
	})

	// Test successful send now
	t.Run("SendNow", func(t *testing.T) {
		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
			SendNow:     true,
		}

		mockService.On("ScheduleBroadcast", mock.Anything, mock.MatchedBy(func(req *domain.ScheduleBroadcastRequest) bool {
			return req.WorkspaceID == request.WorkspaceID &&
				req.ID == request.ID &&
				req.SendNow
		})).Return(nil).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.schedule", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockService.AssertExpectations(t)
	})

	// Test validation error
	t.Run("ValidationError", func(t *testing.T) {
		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
			SendNow:     false,
			// Missing scheduled date and time
		}

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.schedule", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test broadcast not found
	t.Run("BroadcastNotFound", func(t *testing.T) {
		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "nonexistent",
			SendNow:     true,
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a broadcast not found error
		customMock.On("ScheduleBroadcast", mock.Anything, mock.Anything).Return(
			&domain.ErrBroadcastNotFound{ID: "nonexistent"},
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.schedule", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test invalid status (not draft)
	t.Run("InvalidStatus", func(t *testing.T) {
		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
			SendNow:     true,
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a status error
		customMock.On("ScheduleBroadcast", mock.Anything, mock.Anything).Return(
			fmt.Errorf("only broadcasts with draft status can be scheduled, current status: sending"),
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.schedule", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		request := &domain.ScheduleBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
			SendNow:     true,
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a generic error
		customMock.On("ScheduleBroadcast", mock.Anything, mock.Anything).Return(
			errors.New("service error"),
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.schedule", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test method not allowed
	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.schedule", nil)
		w := httptest.NewRecorder()

		handler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	// Test invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.schedule", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleSchedule(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestHandleCancel tests the handleCancel function
func TestHandleCancel(t *testing.T) {
	handler, mockService := setupHandler()

	// Test successful cancel
	t.Run("Success", func(t *testing.T) {
		request := &domain.CancelBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		mockService.On("CancelBroadcast", mock.Anything, mock.MatchedBy(func(req *domain.CancelBroadcastRequest) bool {
			return req.WorkspaceID == request.WorkspaceID &&
				req.ID == request.ID
		})).Return(nil).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Direct method call for testing
		handler.HandleCancel(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockService.AssertExpectations(t)
	})

	// Test validation error
	t.Run("ValidationError", func(t *testing.T) {
		request := &domain.CancelBroadcastRequest{
			// Missing required fields
		}

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleCancel(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test broadcast not found
	t.Run("BroadcastNotFound", func(t *testing.T) {
		request := &domain.CancelBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "nonexistent",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a broadcast not found error
		customMock.On("CancelBroadcast", mock.Anything, mock.Anything).Return(
			&domain.ErrBroadcastNotFound{ID: "nonexistent"},
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleCancel(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test invalid status
	t.Run("InvalidStatus", func(t *testing.T) {
		request := &domain.CancelBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a status error
		customMock.On("CancelBroadcast", mock.Anything, mock.Anything).Return(
			fmt.Errorf("only broadcasts with scheduled or paused status can be cancelled, current status: draft"),
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleCancel(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		request := &domain.CancelBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a generic error
		customMock.On("CancelBroadcast", mock.Anything, mock.Anything).Return(
			errors.New("service error"),
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.cancel", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleCancel(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test method not allowed
	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.cancel", nil)
		w := httptest.NewRecorder()

		handler.HandleCancel(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	// Test invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.cancel", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleCancel(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestHandlePause tests the handlePause function
func TestHandlePause(t *testing.T) {
	handler, mockService := setupHandler()

	// Test successful pause
	t.Run("Success", func(t *testing.T) {
		request := &domain.PauseBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		mockService.On("PauseBroadcast", mock.Anything, mock.MatchedBy(func(req *domain.PauseBroadcastRequest) bool {
			return req.WorkspaceID == request.WorkspaceID &&
				req.ID == request.ID
		})).Return(nil).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.pause", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Direct method call for testing
		handler.HandlePause(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockService.AssertExpectations(t)
	})

	// Test validation error
	t.Run("ValidationError", func(t *testing.T) {
		request := &domain.PauseBroadcastRequest{
			// Missing required fields
		}

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.pause", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePause(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test broadcast not found
	t.Run("BroadcastNotFound", func(t *testing.T) {
		request := &domain.PauseBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "nonexistent",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a broadcast not found error
		customMock.On("PauseBroadcast", mock.Anything, mock.Anything).Return(
			&domain.ErrBroadcastNotFound{ID: "nonexistent"},
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.pause", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandlePause(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test invalid status
	t.Run("InvalidStatus", func(t *testing.T) {
		request := &domain.PauseBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a status error
		customMock.On("PauseBroadcast", mock.Anything, mock.Anything).Return(
			fmt.Errorf("only broadcasts with sending status can be paused, current status: draft"),
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.pause", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandlePause(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		request := &domain.PauseBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a generic error
		customMock.On("PauseBroadcast", mock.Anything, mock.Anything).Return(
			errors.New("service error"),
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.pause", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandlePause(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test method not allowed
	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.pause", nil)
		w := httptest.NewRecorder()

		handler.HandlePause(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	// Test invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.pause", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePause(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestHandleDelete tests the handleDelete function
func TestHandleDelete(t *testing.T) {
	handler, mockService := setupHandler()

	// Test successful delete
	t.Run("Success", func(t *testing.T) {
		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		mockService.On("DeleteBroadcast", mock.Anything, mock.MatchedBy(func(req *domain.DeleteBroadcastRequest) bool {
			return req.WorkspaceID == request.WorkspaceID &&
				req.ID == request.ID
		})).Return(nil).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.delete", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Direct method call for testing
		handler.HandleDelete(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))

		mockService.AssertExpectations(t)
	})

	// Test validation error
	t.Run("ValidationError", func(t *testing.T) {
		request := &domain.DeleteBroadcastRequest{
			// Missing required fields
		}

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.delete", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleDelete(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test broadcast not found
	t.Run("BroadcastNotFound", func(t *testing.T) {
		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "nonexistent",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a broadcast not found error
		customMock.On("DeleteBroadcast", mock.Anything, mock.Anything).Return(
			&domain.ErrBroadcastNotFound{ID: "nonexistent"},
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.delete", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleDelete(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test service error
	t.Run("ServiceError", func(t *testing.T) {
		request := &domain.DeleteBroadcastRequest{
			WorkspaceID: "workspace123",
			ID:          "broadcast123",
		}

		// Create a custom handler just for this test
		customHandler, customMock := setupHandler()

		// Setup mock to return a generic error
		customMock.On("DeleteBroadcast", mock.Anything, mock.Anything).Return(
			errors.New("service error"),
		).Once()

		requestBody, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.delete", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		customHandler.HandleDelete(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		customMock.AssertExpectations(t)
	})

	// Test method not allowed
	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/broadcasts.delete", nil)
		w := httptest.NewRecorder()

		handler.HandleDelete(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	// Test invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/broadcasts.delete", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleDelete(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
