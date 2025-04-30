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
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupEmailHandlerTest prepares a test environment for email handler tests
func setupEmailHandlerTest(t *testing.T) (*mocks.MockEmailServiceInterface, *pkgmocks.MockLogger, *EmailHandler, paseto.V4AsymmetricSecretKey) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockService := mocks.NewMockEmailServiceInterface(ctrl)
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

	handler := NewEmailHandler(mockService, publicKey, mockLogger, "test-secret-key")

	return mockService, mockLogger, handler, secretKey
}

func TestNewEmailHandler(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Act
	handler := NewEmailHandler(mockService, publicKey, mockLogger, "test-secret-key")

	// Assert
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.emailService)
	assert.Equal(t, publicKey, handler.publicKey)
	assert.Equal(t, mockLogger, handler.logger)
	assert.Equal(t, "test-secret-key", handler.secretKey)
}

func TestEmailHandler_RegisterRoutes(t *testing.T) {
	// Arrange
	_, _, handler, _ := setupEmailHandlerTest(t)

	// Create a multiplexer to register routes with
	mux := http.NewServeMux()

	// Register routes with the mux
	handler.RegisterRoutes(mux)

	// Create a test server with the mux
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create an authenticated request to verify the route exists
	reqBody := bytes.NewReader([]byte("{}"))
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/email.testProvider", reqBody)
	require.NoError(t, err)

	// Set content type
	req.Header.Set("Content-Type", "application/json")

	// Act
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	// We expect a 401 Unauthorized since we didn't provide authentication
	// The important part is that the route exists and returns a response
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestEmailHandler_HandleTestEmailProvider(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*mocks.MockEmailServiceInterface)
		expectedStatus int
		expectedResp   *domain.TestEmailProviderResponse
	}{
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			reqBody:        nil,
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedResp:   nil,
		},
		{
			name:           "Invalid request body",
			method:         http.MethodPost,
			reqBody:        "invalid json",
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   nil,
		},
		{
			name:   "Missing recipient email",
			method: http.MethodPost,
			reqBody: domain.TestEmailProviderRequest{
				WorkspaceID: "workspace123",
				Provider: domain.EmailProvider{
					Kind: domain.EmailProviderKindSMTP,
				},
				// Missing To field
			},
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   nil,
		},
		{
			name:   "Missing workspace ID",
			method: http.MethodPost,
			reqBody: domain.TestEmailProviderRequest{
				To: "test@example.com",
				Provider: domain.EmailProvider{
					Kind: domain.EmailProviderKindSMTP,
				},
				// Missing WorkspaceID field
			},
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   nil,
		},
		{
			name:   "Service error",
			method: http.MethodPost,
			reqBody: domain.TestEmailProviderRequest{
				WorkspaceID: "workspace123",
				To:          "test@example.com",
				Provider: domain.EmailProvider{
					Kind:               domain.EmailProviderKindSMTP,
					DefaultSenderEmail: "sender@example.com",
					DefaultSenderName:  "Test Sender",
					SMTP: &domain.SMTPSettings{
						Host:     "smtp.example.com",
						Port:     587,
						Username: "user@example.com",
					},
				},
			},
			setupMock: func(m *mocks.MockEmailServiceInterface) {
				m.EXPECT().
					TestEmailProvider(
						gomock.Any(),
						"workspace123",
						gomock.Any(),
						"test@example.com",
					).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusOK,
			expectedResp: &domain.TestEmailProviderResponse{
				Success: false,
				Error:   "service error",
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			reqBody: domain.TestEmailProviderRequest{
				WorkspaceID: "workspace123",
				To:          "test@example.com",
				Provider: domain.EmailProvider{
					Kind:               domain.EmailProviderKindSMTP,
					DefaultSenderEmail: "sender@example.com",
					DefaultSenderName:  "Test Sender",
					SMTP: &domain.SMTPSettings{
						Host:     "smtp.example.com",
						Port:     587,
						Username: "user@example.com",
					},
				},
			},
			setupMock: func(m *mocks.MockEmailServiceInterface) {
				m.EXPECT().
					TestEmailProvider(
						gomock.Any(),
						"workspace123",
						gomock.Any(),
						"test@example.com",
					).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: &domain.TestEmailProviderResponse{
				Success: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockService, _, handler, _ := setupEmailHandlerTest(t)
			tc.setupMock(mockService)

			// Create request
			var reqBody []byte
			var err error

			if tc.reqBody != nil {
				if strBody, ok := tc.reqBody.(string); ok {
					reqBody = []byte(strBody)
				} else {
					reqBody, err = json.Marshal(tc.reqBody)
					require.NoError(t, err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/email.testProvider", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a response recorder
			w := httptest.NewRecorder()

			// Act - call the handler directly instead of through the mux
			handler.handleTestEmailProvider(w, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedResp != nil {
				var response domain.TestEmailProviderResponse
				err = json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResp.Success, response.Success)
				if tc.expectedResp.Error != "" {
					assert.Equal(t, tc.expectedResp.Error, response.Error)
				}
			}
		})
	}
}

func TestEmailHandler_HandleTestTemplate(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		reqBody        interface{}
		setupMock      func(*mocks.MockEmailServiceInterface)
		expectedStatus int
		expectedResp   *domain.TestTemplateResponse
	}{
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			reqBody:        nil,
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedResp:   nil,
		},
		{
			name:           "Invalid request body",
			method:         http.MethodPost,
			reqBody:        "invalid json",
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   nil,
		},
		{
			name:   "Missing recipient email",
			method: http.MethodPost,
			reqBody: domain.TestTemplateRequest{
				WorkspaceID:  "workspace123",
				TemplateID:   "template123",
				ProviderType: "marketing",
				// Missing RecipientEmail field
			},
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   nil,
		},
		{
			name:   "Missing workspace ID",
			method: http.MethodPost,
			reqBody: domain.TestTemplateRequest{
				TemplateID:     "template123",
				ProviderType:   "marketing",
				RecipientEmail: "test@example.com",
				// Missing WorkspaceID field
			},
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   nil,
		},
		{
			name:   "Invalid provider type",
			method: http.MethodPost,
			reqBody: domain.TestTemplateRequest{
				WorkspaceID:    "workspace123",
				TemplateID:     "template123",
				ProviderType:   "invalid",
				RecipientEmail: "test@example.com",
			},
			setupMock:      func(m *mocks.MockEmailServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   nil,
		},
		{
			name:   "Service error",
			method: http.MethodPost,
			reqBody: domain.TestTemplateRequest{
				WorkspaceID:    "workspace123",
				TemplateID:     "template123",
				ProviderType:   "marketing",
				RecipientEmail: "test@example.com",
			},
			setupMock: func(m *mocks.MockEmailServiceInterface) {
				m.EXPECT().
					TestTemplate(
						gomock.Any(),
						"workspace123",
						"template123",
						"marketing",
						"test@example.com",
					).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusOK,
			expectedResp: &domain.TestTemplateResponse{
				Success: false,
				Error:   "service error",
			},
		},
		{
			name:   "Template not found",
			method: http.MethodPost,
			reqBody: domain.TestTemplateRequest{
				WorkspaceID:    "workspace123",
				TemplateID:     "template123",
				ProviderType:   "marketing",
				RecipientEmail: "test@example.com",
			},
			setupMock: func(m *mocks.MockEmailServiceInterface) {
				m.EXPECT().
					TestTemplate(
						gomock.Any(),
						"workspace123",
						"template123",
						"marketing",
						"test@example.com",
					).
					Return(&domain.ErrTemplateNotFound{Message: "not found"})
			},
			expectedStatus: http.StatusNotFound,
			expectedResp:   nil,
		},
		{
			name:   "Success",
			method: http.MethodPost,
			reqBody: domain.TestTemplateRequest{
				WorkspaceID:    "workspace123",
				TemplateID:     "template123",
				ProviderType:   "marketing",
				RecipientEmail: "test@example.com",
			},
			setupMock: func(m *mocks.MockEmailServiceInterface) {
				m.EXPECT().
					TestTemplate(
						gomock.Any(),
						"workspace123",
						"template123",
						"marketing",
						"test@example.com",
					).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: &domain.TestTemplateResponse{
				Success: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockService, _, handler, _ := setupEmailHandlerTest(t)
			tc.setupMock(mockService)

			// Create request
			var reqBody []byte
			var err error

			if tc.reqBody != nil {
				if strBody, ok := tc.reqBody.(string); ok {
					reqBody = []byte(strBody)
				} else {
					reqBody, err = json.Marshal(tc.reqBody)
					require.NoError(t, err)
				}
			}

			req := httptest.NewRequest(tc.method, "/api/email.testTemplate", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a response recorder
			w := httptest.NewRecorder()

			// Act - call the handler directly instead of through the mux
			handler.handleTestTemplate(w, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedResp != nil {
				var response domain.TestTemplateResponse
				err = json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResp.Success, response.Success)
				if tc.expectedResp.Error != "" {
					assert.Equal(t, tc.expectedResp.Error, response.Error)
				}
			}
		})
	}
}
