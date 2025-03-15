package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"notifuse/server/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, email, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrUserNotFound{Message: "user not found"}
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, email, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrUserNotFound{Message: "user not found"}
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now().UTC()
	session.ExpiresAt = session.ExpiresAt.UTC()

	query := `
		INSERT INTO user_sessions (id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.ExpiresAt,
		session.CreatedAt,
	)
	return err
}

func (r *userRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM user_sessions
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, &domain.ErrSessionNotFound{Message: "session not found"}
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userRepository) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM user_sessions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return &domain.ErrSessionNotFound{Message: "session not found"}
	}
	return nil
}
