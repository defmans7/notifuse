package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
)

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
func (ac *AuthConfig) RequireAuth() func(http.Handler) http.Handler {
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
			userID, err := verified.GetString(string(domain.UserIDKey))
			if err != nil {
				http.Error(w, "User ID not found in token", http.StatusUnauthorized)
				return
			}

			// Get session ID from claims
			sessionID, err := verified.GetString(string(domain.SessionIDKey))
			if err != nil {
				http.Error(w, "Session ID not found in token", http.StatusUnauthorized)
				return
			}

			// put userId and sessionId in the context
			ctx := context.WithValue(r.Context(), domain.UserIDKey, userID)
			ctx = context.WithValue(ctx, domain.SessionIDKey, sessionID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
