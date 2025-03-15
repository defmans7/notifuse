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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

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

func (m *mockUserService) SignUp(ctx context.Context, input service.SignUpInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *mockUserService) VerifyToken(ctx context.Context, input service.VerifyTokenInput) (*service.AuthResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResponse), args.Error(1)
}

func TestUserHandler_SignIn(t *testing.T) {
	mockService := new(mockUserService)
	handler := NewUserHandler(mockService)

	tests := []struct {
		name         string
		input        service.SignInInput
		setupMock    func()
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name: "successful sign in",
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
				"message": "Magic link sent to your email",
			},
		},
		{
			name: "invalid email",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/signin", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.SignIn(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]string
			err = json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_SignUp(t *testing.T) {
	mockService := new(mockUserService)
	handler := NewUserHandler(mockService)

	tests := []struct {
		name         string
		input        service.SignUpInput
		setupMock    func()
		expectedCode int
		expectedBody map[string]string
	}{
		{
			name: "successful sign up",
			input: service.SignUpInput{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setupMock: func() {
				mockService.On("SignUp", mock.Anything, service.SignUpInput{
					Email: "test@example.com",
					Name:  "Test User",
				}).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{
				"message": "Verification link sent to your email",
			},
		},
		{
			name: "user already exists",
			input: service.SignUpInput{
				Email: "existing@example.com",
				Name:  "Existing User",
			},
			setupMock: func() {
				mockService.On("SignUp", mock.Anything, service.SignUpInput{
					Email: "existing@example.com",
					Name:  "Existing User",
				}).Return(fmt.Errorf("user already exists"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"error": "user already exists",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.SignUp(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response map[string]string
			err = json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_VerifyToken(t *testing.T) {
	mockService := new(mockUserService)
	handler := NewUserHandler(mockService)

	user := &domain.User{
		ID:    "user-id",
		Email: "test@example.com",
		Name:  "Test User",
	}

	tests := []struct {
		name         string
		input        service.VerifyTokenInput
		setupMock    func()
		expectedCode int
		expectedBody interface{}
	}{
		{
			name: "successful verification",
			input: service.VerifyTokenInput{
				Token: "valid-token",
			},
			setupMock: func() {
				mockService.On("VerifyToken", mock.Anything, service.VerifyTokenInput{
					Token: "valid-token",
				}).Return(&service.AuthResponse{
					Token:     "auth-token",
					User:      *user,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid token",
			input: service.VerifyTokenInput{
				Token: "invalid-token",
			},
			setupMock: func() {
				mockService.On("VerifyToken", mock.Anything, service.VerifyTokenInput{
					Token: "invalid-token",
				}).Return(nil, fmt.Errorf("invalid token"))
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]string{
				"error": "invalid token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body, err := json.Marshal(tt.input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/verify", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.VerifyToken(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedBody != nil {
				var response map[string]string
				err = json.NewDecoder(rec.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}
