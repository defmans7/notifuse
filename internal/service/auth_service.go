package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"

	"aidanwoods.dev/go-paseto"
)

var (
	ErrSessionExpired = errors.New("session expired")
	ErrUserNotFound   = errors.New("user not found")
)

type AuthService struct {
	repo       domain.AuthRepository
	logger     logger.Logger
	privateKey paseto.V4AsymmetricSecretKey
	publicKey  paseto.V4AsymmetricPublicKey
}

type AuthServiceConfig struct {
	Repository domain.AuthRepository
	PrivateKey []byte
	PublicKey  []byte
	Logger     logger.Logger
}

func NewAuthService(cfg AuthServiceConfig) (*AuthService, error) {
	privateKey, err := paseto.NewV4AsymmetricSecretKeyFromBytes(cfg.PrivateKey)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithField("error", err.Error()).Error("Error creating PASETO private key")
		}
		return nil, err
	}

	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromBytes(cfg.PublicKey)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithField("error", err.Error()).Error("Error creating PASETO public key")
		}
		return nil, err
	}

	return &AuthService{
		repo:       cfg.Repository,
		logger:     cfg.Logger,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
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

// GenerateAuthToken generates an authentication token for a user
func (s *AuthService) GenerateAuthToken(user *domain.User, sessionID string, expiresAt time.Time) string {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(expiresAt)
	token.SetString("user_id", user.ID)
	token.SetString("session_id", sessionID)
	token.SetString("email", user.Email)

	encrypted := token.V4Sign(s.privateKey, nil)
	if encrypted == "" {
		s.logger.WithField("user_id", user.ID).WithField("session_id", sessionID).Error("Failed to sign authentication token")
	}

	return encrypted
}

// GetPrivateKey returns the private key
func (s *AuthService) GetPrivateKey() paseto.V4AsymmetricSecretKey {
	return s.privateKey
}

// GenerateInvitationToken generates a PASETO token for a workspace invitation
func (s *AuthService) GenerateInvitationToken(invitation *domain.WorkspaceInvitation) string {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(invitation.ExpiresAt)
	token.SetString("invitation_id", invitation.ID)
	token.SetString("workspace_id", invitation.WorkspaceID)
	token.SetString("email", invitation.Email)

	encrypted := token.V4Sign(s.privateKey, nil)
	if encrypted == "" {
		s.logger.WithField("invitation_id", invitation.ID).Error("Failed to sign invitation token")
	}

	return encrypted
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	// Delegate to the repository
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.logger.WithField("error", err.Error()).WithField("user_id", userID).Error("Failed to get user by ID")
		return nil, err
	}
	return user, nil
}
