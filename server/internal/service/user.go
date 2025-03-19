package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"

	"notifuse/server/internal/domain"
	"notifuse/server/pkg/logger"
)

type UserService struct {
	repo          domain.UserRepository
	privateKey    paseto.V4AsymmetricSecretKey
	publicKey     paseto.V4AsymmetricPublicKey
	emailSender   EmailSender
	sessionExpiry time.Duration
	logger        logger.Logger
}

type EmailSender interface {
	SendMagicCode(email, code string) error
}

type UserServiceConfig struct {
	Repository    domain.UserRepository
	PrivateKey    []byte
	PublicKey     []byte
	EmailSender   EmailSender
	SessionExpiry time.Duration
	Logger        logger.Logger
}

func NewUserService(cfg UserServiceConfig) (*UserService, error) {
	privateKey, err := paseto.NewV4AsymmetricSecretKeyFromBytes(cfg.PrivateKey)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithField("error", err.Error()).Error("Error creating PASETO private key")
		}
		return nil, fmt.Errorf("error creating private key: %w", err)
	}

	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromBytes(cfg.PublicKey)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.WithField("error", err.Error()).Error("Error creating PASETO public key")
		}
		return nil, fmt.Errorf("error creating public key: %w", err)
	}

	return &UserService{
		repo:          cfg.Repository,
		privateKey:    privateKey,
		publicKey:     publicKey,
		emailSender:   cfg.EmailSender,
		sessionExpiry: cfg.SessionExpiry,
		logger:        cfg.Logger,
	}, nil
}

type SignInInput struct {
	Email string `json:"email"`
}

type VerifyCodeInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type AuthResponse struct {
	Token     string      `json:"token"`
	User      domain.User `json:"user"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// UserServiceInterface defines the interface for user operations
type UserServiceInterface interface {
	SignIn(ctx context.Context, input SignInInput) error
	SignInDev(ctx context.Context, input SignInInput) (string, error)
	VerifyCode(ctx context.Context, input VerifyCodeInput) (*AuthResponse, error)
	VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
}

// Ensure UserService implements UserServiceInterface
var _ UserServiceInterface = (*UserService)(nil)

func (s *UserService) SignIn(ctx context.Context, input SignInInput) error {
	// Check if user exists, if not create a new one
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if _, ok := err.(*domain.ErrUserNotFound); !ok {
			s.logger.WithField("email", input.Email).WithField("error", err.Error()).Error("Failed to get user by email")
			return err
		}

		// User not found, create a new one
		user = &domain.User{
			ID:        generateID(),
			Email:     input.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			s.logger.WithField("email", input.Email).WithField("error", err.Error()).Error("Failed to create user")
			return err
		}
	}

	// Generate magic code
	code := s.generateMagicCode()
	expiresAt := time.Now().Add(s.sessionExpiry)
	codeExpiresAt := time.Now().Add(15 * time.Minute)

	// Create new session
	session := &domain.Session{
		ID:               generateID(),
		UserID:           user.ID,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
		MagicCode:        code,
		MagicCodeExpires: codeExpiresAt,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		s.logger.WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to create session")
		return err
	}

	// Send magic code via email
	if err := s.emailSender.SendMagicCode(user.Email, code); err != nil {
		s.logger.WithField("user_id", user.ID).WithField("email", user.Email).WithField("error", err.Error()).Error("Failed to send magic code")
		return err
	}

	return nil
}

func (s *UserService) VerifyCode(ctx context.Context, input VerifyCodeInput) (*AuthResponse, error) {
	// Find user by email
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		s.logger.WithField("email", input.Email).WithField("error", err.Error()).Error("Failed to get user by email for code verification")
		return nil, err
	}

	// Find all sessions for this user
	sessions, err := s.repo.GetSessionsByUserID(ctx, user.ID)
	if err != nil {
		s.logger.WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to get sessions for user")
		return nil, err
	}

	// Find the session with the matching code
	var matchingSession *domain.Session
	for _, session := range sessions {
		if session.MagicCode == input.Code {
			matchingSession = session
			break
		}
	}

	if matchingSession == nil {
		s.logger.WithField("user_id", user.ID).WithField("email", input.Email).Error("Invalid magic code")
		return nil, fmt.Errorf("invalid magic code")
	}

	// Check if magic code is expired
	if time.Now().After(matchingSession.MagicCodeExpires) {
		s.logger.WithField("user_id", user.ID).WithField("email", input.Email).WithField("session_id", matchingSession.ID).Error("Magic code expired")
		return nil, fmt.Errorf("magic code expired")
	}

	// Clear the magic code from the session
	matchingSession.MagicCode = ""
	matchingSession.MagicCodeExpires = time.Time{}

	if err := s.repo.UpdateSession(ctx, matchingSession); err != nil {
		s.logger.WithField("user_id", user.ID).WithField("session_id", matchingSession.ID).WithField("error", err.Error()).Error("Failed to update session")
		return nil, err
	}

	// Generate authentication token
	token := s.generateAuthToken(user, matchingSession.ID, matchingSession.ExpiresAt)

	return &AuthResponse{
		Token:     token,
		User:      *user,
		ExpiresAt: matchingSession.ExpiresAt,
	}, nil
}

func (s *UserService) generateMagicCode() string {
	// Generate a 6-digit code
	code := make([]byte, 3)
	_, err := rand.Read(code)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to generate random bytes for magic code")
		return "123456" // Fallback code in case of error
	}

	// Convert to 6 digits
	codeNum := int(code[0])<<16 | int(code[1])<<8 | int(code[2])
	codeNum = codeNum % 1000000 // Ensure it's 6 digits
	return fmt.Sprintf("%06d", codeNum)
}

// generateID generates a random ID
func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

func (s *UserService) generateAuthToken(user *domain.User, sessionID string, expiresAt time.Time) string {
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

// VerifyUserSession verifies a user session and returns the associated user
func (s *UserService) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*domain.User, error) {
	// First check if the session is valid and not expired
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		s.logger.WithField("user_id", userID).WithField("session_id", sessionID).WithField("error", err.Error()).Error("Failed to get session by ID")
		return nil, err
	}

	// Verify that the session belongs to the user
	if session.UserID != userID {
		s.logger.WithField("user_id", userID).WithField("session_id", sessionID).WithField("session_user_id", session.UserID).Error("Session does not belong to user")
		return nil, fmt.Errorf("session does not belong to user")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.logger.WithField("user_id", userID).WithField("session_id", sessionID).WithField("expires_at", session.ExpiresAt).Error("Session expired")
		return nil, ErrSessionExpired
	}

	// Get user details
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get user by ID")
		return nil, err
	}

	return user, nil
}

// SignInDev is a development-only version of SignIn that returns the magic code
func (s *UserService) SignInDev(ctx context.Context, input SignInInput) (string, error) {
	// This method is only for development environment
	// Check if user exists, if not create a new one
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if _, ok := err.(*domain.ErrUserNotFound); !ok {
			s.logger.WithField("email", input.Email).WithField("error", err.Error()).Error("Failed to get user by email in dev mode")
			return "", err
		}

		// User not found, create a new one
		user = &domain.User{
			ID:        generateID(),
			Email:     input.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			s.logger.WithField("email", input.Email).WithField("error", err.Error()).Error("Failed to create user in dev mode")
			return "", err
		}
	}

	// Create new session
	expiresAt := time.Now().Add(s.sessionExpiry)
	session := &domain.Session{
		ID:        generateID(),
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		s.logger.WithField("user_id", user.ID).WithField("error", err.Error()).Error("Failed to create session in dev mode")
		return "", err
	}

	// Generate authentication token
	token := s.generateAuthToken(user, session.ID, expiresAt)
	return token, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get user by ID")
		return nil, err
	}
	return user, nil
}
