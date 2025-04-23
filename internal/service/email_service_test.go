package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock auth service
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) AuthenticateUserFromContext(ctx context.Context) (*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockAuthService) AuthenticateUserForWorkspace(ctx context.Context, workspaceID string) (*domain.User, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockAuthService) VerifyUserSession(ctx context.Context, userID, sessionID string) (*domain.User, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockAuthService) GenerateAuthToken(user *domain.User, sessionID string, expiresAt time.Time) string {
	args := m.Called(user, sessionID, expiresAt)
	return args.String(0)
}

func (m *mockAuthService) GetPrivateKey() paseto.V4AsymmetricSecretKey {
	args := m.Called()
	return args.Get(0).(paseto.V4AsymmetricSecretKey)
}

func (m *mockAuthService) GenerateInvitationToken(invitation *domain.WorkspaceInvitation) string {
	args := m.Called(invitation)
	return args.String(0)
}

// Mock logger
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Fatal(msg string) {
	m.Called(msg)
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *mockLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(logger.Logger)
}

func (m *mockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(logger.Logger)
}

func (m *mockLogger) WithError(err error) logger.Logger {
	args := m.Called(err)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(logger.Logger)
}

func TestNewEmailService(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"

	// Act
	service := NewEmailService(mockLog, mockAuth, secretKey)

	// Assert
	require.NotNil(t, service)
	assert.Equal(t, mockLog, service.logger)
	assert.Equal(t, mockAuth, service.authService)
	assert.Equal(t, secretKey, service.secretKey)
}

func TestEmailService_TestEmailProvider_AuthenticationFailure(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	provider := domain.EmailProvider{}
	to := "test@example.com"

	expectedErr := errors.New("authentication failed")
	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(nil, expectedErr)

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate user for workspace")
	mockAuth.AssertExpectations(t)
}

func TestEmailService_TestEmailProvider_ValidationFailure(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}
	to := "test@example.com"

	// Invalid provider that will fail validation
	provider := domain.EmailProvider{
		Kind: domain.EmailProviderKindSMTP,
		// SMTP is nil, which will cause validation failure
	}

	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(user, nil)

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	mockAuth.AssertExpectations(t)
}

func TestEmailService_TestEmailProvider_UnsupportedProvider(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}
	to := "test@example.com"

	// Provider with unsupported kind
	provider := domain.EmailProvider{
		Kind:               "unsupported",
		DefaultSenderEmail: "test@example.com",
		DefaultSenderName:  "Test Sender",
	}

	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(user, nil)

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email provider kind: unsupported")
	mockAuth.AssertExpectations(t)
}

func TestEmailService_TestEmailProvider_SMTPMissingSettings(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}
	to := "test@example.com"

	// SMTP provider with nil settings
	provider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSMTP,
		DefaultSenderEmail: "test@example.com",
		DefaultSenderName:  "Test Sender",
		SMTP:               nil,
	}

	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(user, nil)

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP settings required")
	mockAuth.AssertExpectations(t)
}

func TestEmailService_TestEmailProvider_SESMissingSettings(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}
	to := "test@example.com"

	// SES provider with nil settings
	provider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSES,
		DefaultSenderEmail: "test@example.com",
		DefaultSenderName:  "Test Sender",
		SES:                nil,
	}

	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(user, nil)

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SES settings required")
	mockAuth.AssertExpectations(t)
}

func TestEmailService_TestEmailProvider_SparkPostMissingSettings(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}
	to := "test@example.com"

	// SparkPost provider with nil settings
	provider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSparkPost,
		DefaultSenderEmail: "test@example.com",
		DefaultSenderName:  "Test Sender",
		SparkPost:          nil,
	}

	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(user, nil)

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SparkPost settings required")
	mockAuth.AssertExpectations(t)
}

func TestEmailService_TestEmailProvider_SparkPostAPIError(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}
	to := "test@example.com"

	// SparkPost provider with valid settings
	provider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSparkPost,
		DefaultSenderEmail: "sender@example.com",
		DefaultSenderName:  "Test Sender",
		SparkPost: &domain.SparkPostSettings{
			APIKey:   "test-api-key",
			Endpoint: "https://api.sparkpost.com",
		},
	}

	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(user, nil)

	// Create a test HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors":[{"message":"API key not valid"}]}`))
	}))
	defer server.Close()

	// Override the endpoint to use our test server
	provider.SparkPost.Endpoint = server.URL

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SparkPost API error")
	mockAuth.AssertExpectations(t)
}

// Mock for AWS SES error
type mockSESError struct {
	awsErr awserr.Error
}

func (m *mockSESError) Error() string {
	return m.awsErr.Error()
}

func (m *mockSESError) OrigErr() error {
	return m.awsErr.OrigErr()
}

func (m *mockSESError) Code() string {
	return m.awsErr.Code()
}

func (m *mockSESError) Message() string {
	return m.awsErr.Message()
}

func TestEmailService_TestEmailProvider_DecryptionFailure(t *testing.T) {
	// Arrange
	mockAuth := new(mockAuthService)
	mockLog := new(mockLogger)
	secretKey := "test-secret-key"
	service := NewEmailService(mockLog, mockAuth, secretKey)

	ctx := context.Background()
	workspaceID := "workspace123"
	user := &domain.User{ID: "user123"}
	to := "test@example.com"

	// SMTP provider with encrypted password that will fail to decrypt
	provider := domain.EmailProvider{
		Kind:               domain.EmailProviderKindSMTP,
		DefaultSenderEmail: "test@example.com",
		DefaultSenderName:  "Test Sender",
		SMTP: &domain.SMTPSettings{
			Host:              "smtp.example.com",
			Port:              587,
			Username:          "user",
			EncryptedPassword: "invalid-encrypted-password", // This will fail to decrypt
			UseTLS:            true,
		},
	}

	mockAuth.On("AuthenticateUserForWorkspace", ctx, workspaceID).Return(user, nil)

	// Act
	err := service.TestEmailProvider(ctx, workspaceID, provider, to)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt SMTP password")
	mockAuth.AssertExpectations(t)
}
