package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"notifuse/server/config"
	"notifuse/server/internal/domain"
	"notifuse/server/internal/service"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) SignIn(ctx context.Context, input service.SignInInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *mockUserService) SignInDev(ctx context.Context, input service.SignInInput) (string, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *mockUserService) VerifyCode(ctx context.Context, input service.VerifyCodeInput) (*service.AuthResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResponse), args.Error(1)
}

func (m *mockUserService) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*service.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.User), args.Error(1)
}

func TestUserHandler_SignIn(t *testing.T) {
	mockService := new(mockUserService)
	devConfig := &config.Config{Environment: "development"}
	prodConfig := &config.Config{Environment: "production"}
	devHandler := NewUserHandler(mockService, devConfig)
	prodHandler := NewUserHandler(mockService, prodConfig)

	tests := []struct {
		name         string
		handler      *UserHandler
		input        service.SignInInput
		setupMock    func()
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name:    "successful sign in production",
			handler: prodHandler,
			input: service.SignInInput{
				Email: "test@example.com",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, service.SignInInput{
					Email: "test@example.com",
				}).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{
				"message": "Magic code sent to your email",
			},
		},
		{
			name:    "successful sign in development",
			handler: devHandler,
			input: service.SignInInput{
				Email: "test@example.com",
			},
			setupMock: func() {
				mockService.On("SignInDev", mock.Anything, service.SignInInput{
					Email: "test@example.com",
				}).Return("123456", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{
				"message": "Magic code sent to your email",
				"code":    "123456",
			},
		},
		{
			name:    "invalid email production",
			handler: prodHandler,
			input: service.SignInInput{
				Email: "",
			},
			setupMock: func() {
				mockService.On("SignIn", mock.Anything, service.SignInInput{
					Email: "",
				}).Return(fmt.Errorf("invalid email"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"error": "invalid email",
			},
		},
		{
			name:    "invalid email development",
			handler: devHandler,
			input: service.SignInInput{
				Email: "",
			},
			setupMock: func() {
				mockService.On("SignInDev", mock.Anything, service.SignInInput{
					Email: "",
				}).Return("", fmt.Errorf("invalid email"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"error": "invalid email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/user.signin", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			tt.handler.SignIn(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]string
			err = json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_VerifyCode(t *testing.T) {
	mockService := new(mockUserService)
	config := &config.Config{Environment: "production"}
	handler := NewUserHandler(mockService, config)

	user := &domain.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name          string
		input         service.VerifyCodeInput
		setupMock     func()
		expectedCode  int
		checkResponse func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "successful verification",
			input: service.VerifyCodeInput{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMock: func() {
				mockService.On("VerifyCode", mock.Anything, service.VerifyCodeInput{
					Email: "test@example.com",
					Code:  "123456",
				}).Return(&service.AuthResponse{
					Token:     "auth-token",
					User:      *user,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "auth-token", response["token"])
				assert.Equal(t, user.ID, response["user"].(map[string]interface{})["id"])
				assert.Equal(t, user.Email, response["user"].(map[string]interface{})["email"])
				assert.Equal(t, user.Name, response["user"].(map[string]interface{})["name"])
				assert.NotEmpty(t, response["expires_at"])
			},
		},
		{
			name: "invalid code",
			input: service.VerifyCodeInput{
				Email: "test@example.com",
				Code:  "000000",
			},
			setupMock: func() {
				mockService.On("VerifyCode", mock.Anything, service.VerifyCodeInput{
					Email: "test@example.com",
					Code:  "000000",
				}).Return(nil, fmt.Errorf("invalid or expired code"))
			},
			expectedCode: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "invalid or expired code", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/user.verify", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.VerifyCode(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]interface{}
			err = json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			tt.checkResponse(t, response)

			mockService.AssertExpectations(t)
		})
	}
}
