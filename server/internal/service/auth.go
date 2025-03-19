package service

import (
	"context"
	"database/sql"
	"errors"
	"notifuse/server/internal/domain"
	"notifuse/server/pkg/logger"
	"time"
)

var (
	ErrSessionExpired = errors.New("session expired")
	ErrUserNotFound   = errors.New("user not found")
)

type AuthService struct {
	repo   domain.AuthRepository
	logger logger.Logger
}

func NewAuthService(repo domain.AuthRepository, logger logger.Logger) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: logger,
	}
}

// VerifyUserSession checks if the user exists and the session is valid
func (s *AuthService) VerifyUserSession(ctx context.Context, userID, sessionID string) (*domain.User, error) {
	// First check if the session is valid and not expired
	expiresAt, err := s.repo.GetSessionByID(ctx, sessionID, userID)

	if err == sql.ErrNoRows {
		s.logger.WithField("user_id", userID).WithField("session_id", sessionID).Error("Session not found")
		return nil, ErrSessionExpired
	}
	if err != nil {
		s.logger.WithField("user_id", userID).WithField("session_id", sessionID).WithField("error", err.Error()).Error("Failed to query session")
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(*expiresAt) {
		s.logger.WithField("user_id", userID).WithField("session_id", sessionID).WithField("expires_at", expiresAt).Error("Session expired")
		return nil, ErrSessionExpired
	}

	// Get user details
	user, err := s.repo.GetUserByID(ctx, userID)

	if err == sql.ErrNoRows {
		s.logger.WithField("user_id", userID).Error("User not found")
		return nil, ErrUserNotFound
	}
	if err != nil {
		s.logger.WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to query user")
		return nil, err
	}

	return user, nil
}
