package service

import (
	"context"
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
	SendMagicLink(email, token string) error
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

type SignUpInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type VerifyTokenInput struct {
	Token string `json:"token"`
}

type AuthResponse struct {
	Token     string      `json:"token"`
	User      domain.User `json:"user"`
	ExpiresAt time.Time   `json:"expires_at"`
}

type UserServiceInterface interface {
	SignIn(ctx context.Context, input SignInInput) error
	SignUp(ctx context.Context, input SignUpInput) error
	VerifyToken(ctx context.Context, input VerifyTokenInput) (*AuthResponse, error)
}

// Ensure UserService implements UserServiceInterface
var _ UserServiceInterface = (*UserService)(nil)

func (s *UserService) SignIn(ctx context.Context, input SignInInput) error {
	// Check if user exists
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Generate magic link token
	token := s.generateMagicLinkToken(user.Email)

	// Send magic link email
	if err := s.emailSender.SendMagicLink(user.Email, token); err != nil {
		return fmt.Errorf("error sending magic link: %w", err)
	}

	return nil
}

func (s *UserService) SignUp(ctx context.Context, input SignUpInput) error {
	// Check if user already exists
	_, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err == nil {
		return fmt.Errorf("user already exists")
	}

	// Make sure the error is ErrUserNotFound
	if _, ok := err.(*domain.ErrUserNotFound); !ok {
		return fmt.Errorf("error checking user existence: %w", err)
	}

	// Generate verification token
	token := s.generateVerificationToken(input.Email, input.Name)

	// Send verification email
	if err := s.emailSender.SendMagicLink(input.Email, token); err != nil {
		return fmt.Errorf("error sending verification email: %w", err)
	}

	return nil
}

func (s *UserService) VerifyToken(ctx context.Context, input VerifyTokenInput) (*AuthResponse, error) {
	parser := paseto.NewParser()

	token, err := parser.ParseV4Public(s.publicKey, input.Token, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	email, err := token.GetString("email")
	if err != nil {
		return nil, fmt.Errorf("invalid token claims: %w", err)
	}

	// Check if this is a signup verification
	name, err := token.GetString("name")
	if err == nil {
		// Create new user
		user := &domain.User{
			Email: email,
			Name:  name,
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
		}
	}

	// Get or create user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Create session
	expiresAt := time.Now().Add(s.sessionExpiry)
	session := &domain.Session{
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	// Generate auth token
	signedToken := s.generateAuthToken(user, session.ID, expiresAt)

	return &AuthResponse{
		Token:     signedToken,
		User:      *user,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *UserService) generateMagicLinkToken(email string) string {
	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(15 * time.Minute))
	token.SetString("email", email)

	return token.V4Sign(s.privateKey, []byte{})
}

func (s *UserService) generateVerificationToken(email, name string) string {
	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(15 * time.Minute))
	token.SetString("email", email)
	token.SetString("name", name)

	return token.V4Sign(s.privateKey, []byte{})
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
