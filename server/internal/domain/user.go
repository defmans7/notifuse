package domain

import (
	"context"
	"time"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name,omitempty" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Session struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	ExpiresAt        time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	MagicCode        string    `json:"magic_code,omitempty" db:"magic_code"`
	MagicCodeExpires time.Time `json:"magic_code_expires,omitempty" db:"magic_code_expires_at"`
}

type UserRepository interface {
	// CreateUser creates a new user in the database
	CreateUser(ctx context.Context, user *User) error

	// GetUserByEmail retrieves a user by their email address
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, id string) (*User, error)

	// CreateSession creates a new session for a user
	CreateSession(ctx context.Context, session *Session) error

	// GetSessionByID retrieves a session by its ID
	GetSessionByID(ctx context.Context, id string) (*Session, error)

	// GetSessionsByUserID retrieves all sessions for a user
	GetSessionsByUserID(ctx context.Context, userID string) ([]*Session, error)

	// UpdateSession updates an existing session
	UpdateSession(ctx context.Context, session *Session) error

	// DeleteSession deletes a session by its ID
	DeleteSession(ctx context.Context, id string) error
}

// ErrUserNotFound is returned when a user is not found
type ErrUserNotFound struct {
	Message string
}

func (e *ErrUserNotFound) Error() string {
	return e.Message
}

// ErrSessionNotFound is returned when a session is not found
type ErrSessionNotFound struct {
	Message string
}

func (e *ErrSessionNotFound) Error() string {
	return e.Message
}
