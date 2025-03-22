package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/service"
)

// MockAuthService mocks the AuthServiceInterface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestNewAuthMiddleware(t *testing.T) {
	// Generate a key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Create the middleware
	middleware := NewAuthMiddleware(publicKey)

	// Assert the middleware is created with the given key
	assert.Equal(t, publicKey, middleware.PublicKey)
}

func TestRequireAuth(t *testing.T) {
	// Generate a key pair for testing
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Create the middleware
	authConfig := NewAuthMiddleware(publicKey)
	mockAuthService := new(MockAuthService)

	t.Run("missing authorization header", func(t *testing.T) {
		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is required")
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with invalid header
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("invalid token", func(t *testing.T) {
		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with invalid token
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalidtoken")
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("missing user_id in token", func(t *testing.T) {
		// Create a token with missing user_id
		token := paseto.NewToken()
		token.SetExpiration(time.Now().Add(time.Hour))
		// Intentionally omit setting user_id
		token.SetString("session_id", "test-session")

		signedToken := token.V4Sign(secretKey, nil)

		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with the token
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User ID not found in token")
	})

	t.Run("missing session_id in token", func(t *testing.T) {
		// Create a token with missing session_id
		token := paseto.NewToken()
		token.SetExpiration(time.Now().Add(time.Hour))
		token.SetString("user_id", "test-user")
		// Intentionally omit setting session_id

		signedToken := token.V4Sign(secretKey, nil)

		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with the token
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Session ID not found in token")
	})

	t.Run("session expired", func(t *testing.T) {
		// Create a valid token
		token := paseto.NewToken()
		token.SetExpiration(time.Now().Add(time.Hour))
		token.SetString("user_id", "test-user")
		token.SetString("session_id", "test-session")

		signedToken := token.V4Sign(secretKey, nil)

		// Mock the auth service to return an expired session error
		mockAuthService.On("VerifyUserSession", mock.Anything, "test-user", "test-session").
			Return(nil, service.ErrSessionExpired)

		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with the token
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Session expired")
		mockAuthService.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Reset mock
		mockAuthService.ExpectedCalls = nil

		// Create a valid token
		token := paseto.NewToken()
		token.SetExpiration(time.Now().Add(time.Hour))
		token.SetString("user_id", "test-user")
		token.SetString("session_id", "test-session")

		signedToken := token.V4Sign(secretKey, nil)

		// Mock the auth service to return a user not found error
		mockAuthService.On("VerifyUserSession", mock.Anything, "test-user", "test-session").
			Return(nil, service.ErrUserNotFound)

		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with the token
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
		mockAuthService.AssertExpectations(t)
	})

	t.Run("other error", func(t *testing.T) {
		// Reset mock
		mockAuthService.ExpectedCalls = nil

		// Create a valid token
		token := paseto.NewToken()
		token.SetExpiration(time.Now().Add(time.Hour))
		token.SetString("user_id", "test-user")
		token.SetString("session_id", "test-session")

		signedToken := token.V4Sign(secretKey, nil)

		// Mock the auth service to return some other error
		mockAuthService.On("VerifyUserSession", mock.Anything, "test-user", "test-session").
			Return(nil, assert.AnError)

		// Create a test handler
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with the token
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		mockAuthService.AssertExpectations(t)
	})

	t.Run("successful auth", func(t *testing.T) {
		// Reset mock
		mockAuthService.ExpectedCalls = nil

		// Create a valid token
		token := paseto.NewToken()
		token.SetExpiration(time.Now().Add(time.Hour))
		token.SetString("user_id", "test-user")
		token.SetString("session_id", "test-session")

		signedToken := token.V4Sign(secretKey, nil)

		// Mock the auth service to return a valid user
		user := &domain.User{
			ID:    "test-user",
			Email: "test@example.com",
		}
		mockAuthService.On("VerifyUserSession", mock.Anything, "test-user", "test-session").
			Return(user, nil)

		// Create a test handler that checks for the auth user in context
		var authUserFromContext *AuthenticatedUser
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authUserFromContext = r.Context().Value(AuthUserKey).(*AuthenticatedUser)
			w.WriteHeader(http.StatusOK)
		})

		// Apply the middleware
		handler := authConfig.RequireAuth(mockAuthService)(next)

		// Create a test request with the token
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		w := httptest.NewRecorder()

		// Call the handler
		handler.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, authUserFromContext)
		assert.Equal(t, "test-user", authUserFromContext.ID)
		assert.Equal(t, "test@example.com", authUserFromContext.Email)
		mockAuthService.AssertExpectations(t)
	})
}
