package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/google/uuid"
)

type UserService struct {
	repo          domain.UserRepository
	authService   domain.AuthService
	emailSender   EmailSender
	sessionExpiry time.Duration
	logger        logger.Logger
	isDevelopment bool
}

type EmailSender interface {
	SendMagicCode(email, code string) error
}

type UserServiceConfig struct {
	Repository    domain.UserRepository
	AuthService   domain.AuthService
	EmailSender   EmailSender
	SessionExpiry time.Duration
	Logger        logger.Logger
	IsDevelopment bool
}

func NewUserService(cfg UserServiceConfig) (*UserService, error) {
	return &UserService{
		repo:          cfg.Repository,
		authService:   cfg.AuthService,
		emailSender:   cfg.EmailSender,
		sessionExpiry: cfg.SessionExpiry,
		logger:        cfg.Logger,
		isDevelopment: cfg.IsDevelopment,
	}, nil
}

// Ensure UserService implements UserServiceInterface
var _ domain.UserServiceInterface = (*UserService)(nil)

func (s *UserService) SignIn(ctx context.Context, input domain.SignInInput) (string, error) {
	// Check if user exists, if not create a new one
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if _, ok := err.(*domain.ErrUserNotFound); !ok {
			s.logger.WithField("email", input.Email).WithField("error", err.Error()).Error("Failed to get user by email")
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
			s.logger.WithField("email", input.Email).WithField("error", err.Error()).Error("Failed to create user")
			return "", err
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
		return "", err
	}

	// In development mode, return the code directly
	// In production, send the code via email
	if s.isDevelopment {
		return code, nil
	}

	// Send magic code via email in production
	if err := s.emailSender.SendMagicCode(user.Email, code); err != nil {
		s.logger.WithField("user_id", user.ID).WithField("email", user.Email).WithField("error", err.Error()).Error("Failed to send magic code")
		return "", err
	}

	return "", nil
}

func (s *UserService) VerifyCode(ctx context.Context, input domain.VerifyCodeInput) (*domain.AuthResponse, error) {
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
	token := s.authService.GenerateAuthToken(user, matchingSession.ID, matchingSession.ExpiresAt)

	return &domain.AuthResponse{
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

// generateID generates a proper UUID
func generateID() string {
	// Use the github.com/google/uuid package to generate a standard UUID
	return uuid.New().String()
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

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.WithField("user_id", userID).WithField("error", err.Error()).Error("Failed to get user by ID")
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email address
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.WithField("email", email).WithField("error", err.Error()).Error("Failed to get user by email")
		return nil, err
	}
	return user, nil
}
