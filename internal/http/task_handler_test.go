package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTaskHandler_ExecuteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskService := mocks.NewMockTaskService(ctrl)
	// For tests we don't need the actual key, we can use a mock or nil since we're not validating auth
	var publicKey paseto.V4AsymmetricPublicKey
	mockLogger := &mockLogger{}
	secretKey := "test-secret-key"

	handler := NewTaskHandler(mockTaskService, publicKey, mockLogger, secretKey)

	t.Run("Successful execution", func(t *testing.T) {
		// Setup
		reqBody := domain.ExecuteTaskRequest{
			WorkspaceID: "workspace1",
			ID:          "task123",
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Configure service mock to return success
		mockTaskService.EXPECT().
			ExecuteTask(gomock.Any(), reqBody.WorkspaceID, reqBody.ID).
			Return(nil)

		// Call handler
		req := httptest.NewRequest(http.MethodPost, "/api/tasks.execute", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("Method not allowed", func(t *testing.T) {
		// Call handler with wrong method
		req := httptest.NewRequest(http.MethodGet, "/api/tasks.execute", nil)
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		// Call handler with invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/api/tasks.execute", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		// Setup request with missing fields
		reqBody := map[string]interface{}{
			"WorkspaceID": "workspace1",
			// missing ID field
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Call handler
		req := httptest.NewRequest(http.MethodPost, "/api/tasks.execute", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Task not found error", func(t *testing.T) {
		// Setup
		reqBody := domain.ExecuteTaskRequest{
			WorkspaceID: "workspace1",
			ID:          "task123",
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Configure service mock to return a NotFound error
		notFoundErr := &domain.ErrNotFound{
			Entity: "task",
			ID:     reqBody.ID,
		}
		mockTaskService.EXPECT().
			ExecuteTask(gomock.Any(), reqBody.WorkspaceID, reqBody.ID).
			Return(notFoundErr)

		// Call handler
		req := httptest.NewRequest(http.MethodPost, "/api/tasks.execute", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response has correct status code for not found
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Task execution error - unsupported task type", func(t *testing.T) {
		// Setup
		reqBody := domain.ExecuteTaskRequest{
			WorkspaceID: "workspace1",
			ID:          "task123",
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Configure service mock to return a TaskExecution error for unsupported type
		execErr := &domain.ErrTaskExecution{
			TaskID: reqBody.ID,
			Reason: "no processor registered for task type",
			Err:    errors.New("unsupported_task_type"),
		}
		mockTaskService.EXPECT().
			ExecuteTask(gomock.Any(), reqBody.WorkspaceID, reqBody.ID).
			Return(execErr)

		// Call handler
		req := httptest.NewRequest(http.MethodPost, "/api/tasks.execute", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response has correct status code for bad request (client error)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Task execution error - general error", func(t *testing.T) {
		// Setup
		reqBody := domain.ExecuteTaskRequest{
			WorkspaceID: "workspace1",
			ID:          "task123",
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Configure service mock to return a general task execution error
		execErr := &domain.ErrTaskExecution{
			TaskID: reqBody.ID,
			Reason: "processing failed",
			Err:    errors.New("internal error"),
		}
		mockTaskService.EXPECT().
			ExecuteTask(gomock.Any(), reqBody.WorkspaceID, reqBody.ID).
			Return(execErr)

		// Call handler
		req := httptest.NewRequest(http.MethodPost, "/api/tasks.execute", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response has correct status code for internal server error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("Task timeout error", func(t *testing.T) {
		// Setup
		reqBody := domain.ExecuteTaskRequest{
			WorkspaceID: "workspace1",
			ID:          "task123",
		}

		reqJSON, _ := json.Marshal(reqBody)

		// Configure service mock to return a timeout error
		timeoutErr := &domain.ErrTaskTimeout{
			TaskID:     reqBody.ID,
			MaxRuntime: 60,
		}
		mockTaskService.EXPECT().
			ExecuteTask(gomock.Any(), reqBody.WorkspaceID, reqBody.ID).
			Return(timeoutErr)

		// Call handler
		req := httptest.NewRequest(http.MethodPost, "/api/tasks.execute", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ExecuteTask(rec, req)

		// Verify response has correct status code for timeout
		assert.Equal(t, http.StatusGatewayTimeout, rec.Code)
	})
}

// Mock logger to use in tests
type mockLogger struct{}

func (l *mockLogger) Debug(msg string)                                       {}
func (l *mockLogger) Info(msg string)                                        {}
func (l *mockLogger) Warn(msg string)                                        {}
func (l *mockLogger) Error(msg string)                                       {}
func (l *mockLogger) Fatal(msg string)                                       {}
func (l *mockLogger) Debugf(format string, args ...interface{})              {}
func (l *mockLogger) Infof(format string, args ...interface{})               {}
func (l *mockLogger) Warnf(format string, args ...interface{})               {}
func (l *mockLogger) Errorf(format string, args ...interface{})              {}
func (l *mockLogger) Fatalf(format string, args ...interface{})              {}
func (l *mockLogger) WithField(key string, value interface{}) logger.Logger  { return l }
func (l *mockLogger) WithFields(fields map[string]interface{}) logger.Logger { return l }
func (l *mockLogger) WithError(err error) logger.Logger                      { return l }
