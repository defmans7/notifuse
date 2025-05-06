package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupWebhookEventHandlerTest prepares test dependencies and creates a webhook event handler
func setupWebhookEventHandlerTest(t *testing.T) (*mocks.MockWebhookEventServiceInterface, *pkgmocks.MockLogger, *WebhookEventHandler) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockService := mocks.NewMockWebhookEventServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup common logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Fatal(gomock.Any()).AnyTimes()

	// Create key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	handler := NewWebhookEventHandler(mockService, publicKey, mockLogger)
	return mockService, mockLogger, handler
}

func TestWebhookEventHandler_RegisterRoutes(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockWebhookEventServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Fatal(gomock.Any()).AnyTimes()

	// Create key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	handler := NewWebhookEventHandler(mockService, publicKey, mockLogger)

	// Register routes to a new multiplexer
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Create a test server
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test only that our routes are registered and accessible
	// We will test each individual handler separately
	routes := []string{
		"/webhooks/email",
		"/api/webhookEvents.list",
		"/api/webhookEvents.get",
		"/api/webhookEvents.getByMessageID",
		"/api/webhookEvents.getByTransactionalID",
		"/api/webhookEvents.getByBroadcastID",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			// All routes except the public webhook endpoint require authentication
			var resp *http.Response
			var err error

			if route == "/webhooks/email" {
				// Test that the route exists (will return method not allowed for GET)
				resp, err = http.Get(server.URL + route)
			} else {
				// For authenticated routes, just test that they're registered
				// (will return unauthorized without a token)
				resp, err = http.Get(server.URL + route)
			}

			require.NoError(t, err)
			defer resp.Body.Close()

			// Routes exist if they don't return 404
			assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}

func TestWebhookEventHandler_HandleIncomingWebhook(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		queryParams    url.Values
		requestBody    string
		setupMock      func(*mocks.MockWebhookEventServiceInterface)
		expectedStatus int
	}{
		{
			name:   "Method not allowed",
			method: http.MethodGet,
			queryParams: url.Values{
				"provider":       []string{"sparkpost"},
				"workspace_id":   []string{"workspace123"},
				"integration_id": []string{"integration123"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Missing provider",
			method: http.MethodPost,
			queryParams: url.Values{
				"workspace_id":   []string{"workspace123"},
				"integration_id": []string{"integration123"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing workspace ID",
			method: http.MethodPost,
			queryParams: url.Values{
				"provider":       []string{"sparkpost"},
				"integration_id": []string{"integration123"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing integration ID",
			method: http.MethodPost,
			queryParams: url.Values{
				"provider":     []string{"sparkpost"},
				"workspace_id": []string{"workspace123"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Success",
			method: http.MethodPost,
			queryParams: url.Values{
				"provider":       []string{"sparkpost"},
				"workspace_id":   []string{"workspace123"},
				"integration_id": []string{"integration123"},
			},
			requestBody: `{"event": "test"}`,
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					ProcessWebhook(
						gomock.Any(),
						"workspace123",
						"integration123",
						gomock.Any(), // Raw body bytes
					).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Service error",
			method: http.MethodPost,
			queryParams: url.Values{
				"provider":       []string{"sparkpost"},
				"workspace_id":   []string{"workspace123"},
				"integration_id": []string{"integration123"},
			},
			requestBody: `{"event": "test"}`,
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					ProcessWebhook(
						gomock.Any(),
						"workspace123",
						"integration123",
						gomock.Any(), // Raw body bytes
					).
					Return(errors.New("processing error"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupWebhookEventHandlerTest(t)
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/webhooks/email?"+tc.queryParams.Encode(), bytes.NewBufferString(tc.requestBody))
			w := httptest.NewRecorder()

			handler.handleIncomingWebhook(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, true, response["success"])
			}
		})
	}
}

func TestWebhookEventHandler_HandleList(t *testing.T) {
	workspaceID := "workspace123"

	testCases := []struct {
		name           string
		method         string
		queryParams    url.Values
		setupMock      func(*mocks.MockWebhookEventServiceInterface)
		expectedStatus int
		expectedEvents bool
	}{
		{
			name:   "Method not allowed",
			method: http.MethodPost,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"type":         []string{"delivered"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedEvents: false,
		},
		{
			name:   "Invalid limit parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"type":         []string{"delivered"},
				"limit":        []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid offset parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"type":         []string{"delivered"},
				"offset":       []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid request validation",
			method: http.MethodGet,
			queryParams: url.Values{
				// Missing required workspace_id
				"type": []string{"delivered"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Success with no events",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"type":         []string{"delivered"},
				"limit":        []string{"10"},
				"offset":       []string{"0"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByType(
						gomock.Any(),
						workspaceID,
						domain.EmailEventDelivered,
						10,
						0,
					).
					Return([]*domain.WebhookEvent{}, nil)

				m.EXPECT().
					GetEventCount(
						gomock.Any(),
						workspaceID,
						domain.EmailEventDelivered,
					).
					Return(0, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
		{
			name:   "Success with events",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"type":         []string{"delivered"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				events := []*domain.WebhookEvent{
					{
						ID:             "event1",
						Type:           domain.EmailEventDelivered,
						RecipientEmail: "test@example.com",
						MessageID:      "msg123",
					},
				}

				m.EXPECT().
					GetEventsByType(
						gomock.Any(),
						workspaceID,
						domain.EmailEventDelivered,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(events, nil)

				m.EXPECT().
					GetEventCount(
						gomock.Any(),
						workspaceID,
						domain.EmailEventDelivered,
					).
					Return(1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
		{
			name:   "GetEventsByType service error",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"type":         []string{"delivered"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByType(
						gomock.Any(),
						workspaceID,
						domain.EmailEventDelivered,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedEvents: false,
		},
		{
			name:   "GetEventCount service error",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"type":         []string{"delivered"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByType(
						gomock.Any(),
						workspaceID,
						domain.EmailEventDelivered,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return([]*domain.WebhookEvent{}, nil)

				m.EXPECT().
					GetEventCount(
						gomock.Any(),
						workspaceID,
						domain.EmailEventDelivered,
					).
					Return(0, errors.New("count error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedEvents: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupWebhookEventHandlerTest(t)
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/api/webhookEvents.list?"+tc.queryParams.Encode(), nil)
			w := httptest.NewRecorder()

			handler.handleList(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response, "events")
				assert.Contains(t, response, "total")
			}
		})
	}
}

func TestWebhookEventHandler_HandleGet(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		queryParams    url.Values
		setupMock      func(*mocks.MockWebhookEventServiceInterface)
		expectedStatus int
		expectedEvent  bool
	}{
		{
			name:   "Method not allowed",
			method: http.MethodPost,
			queryParams: url.Values{
				"id": []string{"event123"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedEvent:  false,
		},
		{
			name:           "Missing event ID",
			method:         http.MethodGet,
			queryParams:    url.Values{},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvent:  false,
		},
		{
			name:   "Event not found",
			method: http.MethodGet,
			queryParams: url.Values{
				"id": []string{"event123"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventByID(
						gomock.Any(),
						"event123",
					).
					Return(nil, &domain.ErrWebhookEventNotFound{ID: "event123"})
			},
			expectedStatus: http.StatusNotFound,
			expectedEvent:  false,
		},
		{
			name:   "Service error",
			method: http.MethodGet,
			queryParams: url.Values{
				"id": []string{"event123"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventByID(
						gomock.Any(),
						"event123",
					).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedEvent:  false,
		},
		{
			name:   "Success",
			method: http.MethodGet,
			queryParams: url.Values{
				"id": []string{"event123"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				event := &domain.WebhookEvent{
					ID:             "event123",
					Type:           domain.EmailEventDelivered,
					RecipientEmail: "test@example.com",
					MessageID:      "msg123",
				}

				m.EXPECT().
					GetEventByID(
						gomock.Any(),
						"event123",
					).
					Return(event, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvent:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupWebhookEventHandlerTest(t)
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/api/webhookEvents.get?"+tc.queryParams.Encode(), nil)
			w := httptest.NewRecorder()

			handler.handleGet(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedEvent {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response, "event")
			}
		})
	}
}

func TestWebhookEventHandler_HandleGetByMessageID(t *testing.T) {
	messageID := "msg123"

	testCases := []struct {
		name           string
		method         string
		queryParams    url.Values
		setupMock      func(*mocks.MockWebhookEventServiceInterface)
		expectedStatus int
		expectedEvents bool
	}{
		{
			name:   "Method not allowed",
			method: http.MethodPost,
			queryParams: url.Values{
				"message_id": []string{messageID},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedEvents: false,
		},
		{
			name:   "Invalid limit parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"message_id": []string{messageID},
				"limit":      []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid offset parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"message_id": []string{messageID},
				"offset":     []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:        "Invalid request validation",
			method:      http.MethodGet,
			queryParams: url.Values{
				// Missing required message_id
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Service error",
			method: http.MethodGet,
			queryParams: url.Values{
				"message_id": []string{messageID},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByMessageID(
						gomock.Any(),
						messageID,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedEvents: false,
		},
		{
			name:   "Success with no events",
			method: http.MethodGet,
			queryParams: url.Values{
				"message_id": []string{messageID},
				"limit":      []string{"10"},
				"offset":     []string{"0"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByMessageID(
						gomock.Any(),
						messageID,
						10,
						0,
					).
					Return([]*domain.WebhookEvent{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
		{
			name:   "Success with events",
			method: http.MethodGet,
			queryParams: url.Values{
				"message_id": []string{messageID},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				events := []*domain.WebhookEvent{
					{
						ID:             "event1",
						Type:           domain.EmailEventDelivered,
						RecipientEmail: "test@example.com",
						MessageID:      messageID,
					},
				}

				m.EXPECT().
					GetEventsByMessageID(
						gomock.Any(),
						messageID,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(events, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupWebhookEventHandlerTest(t)
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/api/webhookEvents.getByMessageID?"+tc.queryParams.Encode(), nil)
			w := httptest.NewRecorder()

			handler.handleGetByMessageID(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedEvents {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response, "events")
				assert.Contains(t, response, "total")
			}
		})
	}
}

func TestWebhookEventHandler_HandleGetByTransactionalID(t *testing.T) {
	workspaceID := "workspace123"
	transactionalID := "trans123"

	testCases := []struct {
		name           string
		method         string
		queryParams    url.Values
		setupMock      func(*mocks.MockWebhookEventServiceInterface)
		expectedStatus int
		expectedEvents bool
	}{
		{
			name:   "Method not allowed",
			method: http.MethodPost,
			queryParams: url.Values{
				"workspace_id":     []string{workspaceID},
				"transactional_id": []string{transactionalID},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedEvents: false,
		},
		{
			name:   "Invalid limit parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id":     []string{workspaceID},
				"transactional_id": []string{transactionalID},
				"limit":            []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid offset parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id":     []string{workspaceID},
				"transactional_id": []string{transactionalID},
				"offset":           []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid request validation - missing workspace_id",
			method: http.MethodGet,
			queryParams: url.Values{
				"transactional_id": []string{transactionalID},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid request validation - missing transactional_id",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Service error",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id":     []string{workspaceID},
				"transactional_id": []string{transactionalID},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByTransactionalID(
						gomock.Any(),
						transactionalID,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedEvents: false,
		},
		{
			name:   "Success with no events",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id":     []string{workspaceID},
				"transactional_id": []string{transactionalID},
				"limit":            []string{"10"},
				"offset":           []string{"0"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByTransactionalID(
						gomock.Any(),
						transactionalID,
						10,
						0,
					).
					Return([]*domain.WebhookEvent{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
		{
			name:   "Success with events",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id":     []string{workspaceID},
				"transactional_id": []string{transactionalID},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				events := []*domain.WebhookEvent{
					{
						ID:              "event1",
						Type:            domain.EmailEventDelivered,
						RecipientEmail:  "test@example.com",
						TransactionalID: transactionalID,
					},
				}

				m.EXPECT().
					GetEventsByTransactionalID(
						gomock.Any(),
						transactionalID,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(events, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupWebhookEventHandlerTest(t)
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/api/webhookEvents.getByTransactionalID?"+tc.queryParams.Encode(), nil)
			w := httptest.NewRecorder()

			handler.handleGetByTransactionalID(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedEvents {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response, "events")
				assert.Contains(t, response, "total")
			}
		})
	}
}

func TestWebhookEventHandler_HandleGetByBroadcastID(t *testing.T) {
	workspaceID := "workspace123"
	broadcastID := "broadcast123"

	testCases := []struct {
		name           string
		method         string
		queryParams    url.Values
		setupMock      func(*mocks.MockWebhookEventServiceInterface)
		expectedStatus int
		expectedEvents bool
	}{
		{
			name:   "Method not allowed",
			method: http.MethodPost,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"broadcast_id": []string{broadcastID},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedEvents: false,
		},
		{
			name:   "Invalid limit parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"broadcast_id": []string{broadcastID},
				"limit":        []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid offset parameter",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"broadcast_id": []string{broadcastID},
				"offset":       []string{"invalid"},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid request validation - missing workspace_id",
			method: http.MethodGet,
			queryParams: url.Values{
				"broadcast_id": []string{broadcastID},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Invalid request validation - missing broadcast_id",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
			},
			setupMock:      func(m *mocks.MockWebhookEventServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedEvents: false,
		},
		{
			name:   "Service error",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"broadcast_id": []string{broadcastID},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByBroadcastID(
						gomock.Any(),
						broadcastID,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedEvents: false,
		},
		{
			name:   "Success with no events",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"broadcast_id": []string{broadcastID},
				"limit":        []string{"10"},
				"offset":       []string{"0"},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				m.EXPECT().
					GetEventsByBroadcastID(
						gomock.Any(),
						broadcastID,
						10,
						0,
					).
					Return([]*domain.WebhookEvent{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
		{
			name:   "Success with events",
			method: http.MethodGet,
			queryParams: url.Values{
				"workspace_id": []string{workspaceID},
				"broadcast_id": []string{broadcastID},
			},
			setupMock: func(m *mocks.MockWebhookEventServiceInterface) {
				events := []*domain.WebhookEvent{
					{
						ID:             "event1",
						Type:           domain.EmailEventDelivered,
						RecipientEmail: "test@example.com",
						BroadcastID:    broadcastID,
					},
				}

				m.EXPECT().
					GetEventsByBroadcastID(
						gomock.Any(),
						broadcastID,
						gomock.Any(), // Default limit
						gomock.Any(), // Default offset
					).
					Return(events, nil)
			},
			expectedStatus: http.StatusOK,
			expectedEvents: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService, _, handler := setupWebhookEventHandlerTest(t)
			tc.setupMock(mockService)

			req := httptest.NewRequest(tc.method, "/api/webhookEvents.getByBroadcastID?"+tc.queryParams.Encode(), nil)
			w := httptest.NewRecorder()

			handler.handleGetByBroadcastID(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedEvents {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response, "events")
				assert.Contains(t, response, "total")
			}
		})
	}
}
