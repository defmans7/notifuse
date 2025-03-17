package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"

	"notifuse/server/internal/domain"
)

type UserService struct {
	repo          domain.UserRepository
	privateKey    paseto.V4AsymmetricSecretKey
	publicKey     paseto.V4AsymmetricPublicKey
	emailSender   EmailSender
	sessionExpiry time.Duration
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
}

func NewUserService(cfg UserServiceConfig) (*UserService, error) {
	privateKey, err := paseto.NewV4AsymmetricSecretKeyFromBytes(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating private key: %w", err)
	}

	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromBytes(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("error creating public key: %w", err)
	}

	return &UserService{
		repo:          cfg.Repository,
		privateKey:    privateKey,
		publicKey:     publicKey,
		emailSender:   cfg.EmailSender,
		sessionExpiry: cfg.SessionExpiry,
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
	VerifyUserSession(ctx context.Context, userID string, sessionID string) (*User, error)
}

// Ensure UserService implements UserServiceInterface
var _ UserServiceInterface = (*UserService)(nil)

func (s *UserService) SignIn(ctx context.Context, input SignInInput) error {
	// Check if user exists
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Generate 6-digit magic code
	code := s.generateMagicCode()

	// Create a temporary session with the magic code
	session := &domain.Session{
		UserID:           user.ID,
		MagicCode:        code,
		MagicCodeExpires: time.Now().Add(15 * time.Minute),
		ExpiresAt:        time.Now().Add(s.sessionExpiry),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}

	// Send magic code email
	if err := s.emailSender.SendMagicCode(user.Email, code); err != nil {
		return fmt.Errorf("error sending magic code: %w", err)
	}

	return nil
}

func (s *UserService) VerifyCode(ctx context.Context, input VerifyCodeInput) (*AuthResponse, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Find session with matching code
	sessions, err := s.repo.GetSessionsByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting sessions: %w", err)
	}

	var validSession *domain.Session
	for _, session := range sessions {
		if session.MagicCode == input.Code && time.Now().Before(session.MagicCodeExpires) {
			validSession = session
			break
		}
	}

	if validSession == nil {
		return nil, fmt.Errorf("invalid or expired code")
	}

	// Clear the magic code after successful verification
	validSession.MagicCode = ""
	validSession.MagicCodeExpires = time.Time{}

	if err := s.repo.UpdateSession(ctx, validSession); err != nil {
		return nil, fmt.Errorf("error updating session: %w", err)
	}

	// Generate auth token
	signedToken := s.generateAuthToken(user, validSession.ID, validSession.ExpiresAt)

	return &AuthResponse{
		Token:     signedToken,
		User:      *user,
		ExpiresAt: validSession.ExpiresAt,
	}, nil
}

func (s *UserService) generateMagicCode() string {
	const digits = "0123456789"
	code := make([]byte, 6)
	_, err := rand.Read(code)
	if err != nil {
		// If we can't generate random numbers, use timestamp as fallback
		now := time.Now().UnixNano()
		for i := range code {
			code[i] = digits[now%10]
			now /= 10
		}
		return string(code)
	}

	for i := range code {
		code[i] = digits[int(code[i])%len(digits)]
	}
	return string(code)
}

func (s *UserService) generateAuthToken(user *domain.User, sessionID string, expiresAt time.Time) string {
	token := paseto.NewToken()
	token.SetExpiration(expiresAt)
	token.SetString("user_id", user.ID)
	token.SetString("email", user.Email)
	token.SetString("name", user.Name)
	token.SetString("session_id", sessionID)

	return token.V4Sign(s.privateKey, []byte{})
}

// VerifyUserSession verifies a user session and returns the associated user
func (s *UserService) VerifyUserSession(ctx context.Context, userID string, sessionID string) (*User, error) {
	// Get user by ID
	domainUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Get session by ID
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, ErrSessionExpired
	}

	// Verify session belongs to user
	if session.UserID != userID {
		return nil, ErrSessionExpired
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	// Convert domain user to service user
	serviceUser := &User{
		ID:        domainUser.ID,
		Email:     domainUser.Email,
		CreatedAt: domainUser.CreatedAt,
	}

	return serviceUser, nil
}

// SignInDev is a development-only version of SignIn that returns the magic code
func (s *UserService) SignInDev(ctx context.Context, input SignInInput) (string, error) {
	// Check if user exists
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return "", fmt.Errorf("error getting user: %w", err)
	}

	// Generate 6-digit magic code
	code := s.generateMagicCode()

	// Create a temporary session with the magic code
	session := &domain.Session{
		UserID:           user.ID,
		MagicCode:        code,
		MagicCodeExpires: time.Now().Add(15 * time.Minute),
		ExpiresAt:        time.Now().Add(s.sessionExpiry),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return "", fmt.Errorf("error creating session: %w", err)
	}

	// In development mode, we don't actually send the email
	return code, nil
}
