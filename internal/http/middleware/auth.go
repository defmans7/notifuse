package middleware

import (
	"context"
	"fmt"
	"net/http"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/service"
	"strings"

	"aidanwoods.dev/go-paseto"
)

// Key for storing user ID and session ID in context
type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	SessionIDKey contextKey = "session_id"
	AuthUserKey  contextKey = "auth_user"
)

// AuthenticatedUser represents a user that has been authenticated
type AuthenticatedUser struct {
	ID    string
	Email string
}

// AuthServiceInterface defines the interface for authentication operations
type AuthServiceInterface interface {
	VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error)
}

// AuthConfig holds the configuration for the auth middleware
type AuthConfig struct {
	PublicKey paseto.V4AsymmetricPublicKey
}

// NewAuthMiddleware creates a new auth middleware with the given public key
func NewAuthMiddleware(publicKey paseto.V4AsymmetricPublicKey) *AuthConfig {
	return &AuthConfig{
		PublicKey: publicKey,
	}
}

// RequireAuth creates a middleware that verifies the PASETO token and user session
func (ac *AuthConfig) RequireAuth(authService AuthServiceInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Parse and verify the token
			parser := paseto.NewParser()
			parser.AddRule(paseto.NotExpired())

			// Verify token and get claims
			verified, err := parser.ParseV4Public(ac.PublicKey, token, nil)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Get user ID from claims
			userID, err := verified.GetString("user_id")
			if err != nil {
				http.Error(w, "User ID not found in token", http.StatusUnauthorized)
				return
			}

			// Get session ID from claims
			sessionID, err := verified.GetString("session_id")
			if err != nil {
				http.Error(w, "Session ID not found in token", http.StatusUnauthorized)
				return
			}

			// Verify user session
			user, err := authService.VerifyUserSession(r.Context(), userID, sessionID)
			if err != nil {
				switch err {
				case service.ErrSessionExpired:
					http.Error(w, "Session expired", http.StatusUnauthorized)
				case service.ErrUserNotFound:
					http.Error(w, "User not found", http.StatusUnauthorized)
				default:
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
				return
			}

			// Add authenticated user to context
			authUser := &AuthenticatedUser{
				ID:    user.ID,
				Email: user.Email,
			}
			ctx := context.WithValue(r.Context(), AuthUserKey, authUser)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
