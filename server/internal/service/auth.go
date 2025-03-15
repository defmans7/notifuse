package service

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrSessionExpired = errors.New("session expired")
	ErrUserNotFound   = errors.New("user not found")
)

type AuthService struct {
	db *sql.DB
}

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

// VerifyUserSession checks if the user exists and the session is valid
func (s *AuthService) VerifyUserSession(ctx context.Context, userID, sessionID string) (*User, error) {
	// First check if the session is valid and not expired
	var expiresAt time.Time
	err := s.db.QueryRowContext(ctx,
		"SELECT expires_at FROM sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&expiresAt)

	if err == sql.ErrNoRows {
		return nil, ErrSessionExpired
	}
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(expiresAt) {
		return nil, ErrSessionExpired
	}

	// Get user details
	var user User
	err = s.db.QueryRowContext(ctx,
		"SELECT id, email, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
