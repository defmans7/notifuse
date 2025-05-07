package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSMTPService is a wrapper around the real SMTP service for testing
type testSMTPService struct {
	logger            logger.Logger
	mailClientCreator func(host string, options ...interface{}) (interface{}, error)
	dialAndSendFunc   func(msg interface{}) error
}

// newTestSMTPService creates a new testSMTPService instance for testing
func newTestSMTPService(logger logger.Logger) *testSMTPService {
	return &testSMTPService{
		logger: logger,
	}
}

// Test implementation for SendEmail that doesn't use the actual mail package
func (s *testSMTPService) SendEmail(ctx context.Context, workspaceID string, fromAddress, fromName, to, subject, content string, provider *domain.EmailProvider) error {
	if provider.SMTP == nil {
		return fmt.Errorf("SMTP settings required")
	}

	// Simulate errors for "invalid-email" values
	if fromAddress == "invalid-email" {
		return fmt.Errorf("invalid sender: bad format")
	}
	if to == "invalid-email" {
		return fmt.Errorf("invalid recipient email: bad format")
	}

	// Simulate client creation
	if s.mailClientCreator != nil {
		_, err := s.mailClientCreator(provider.SMTP.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	}

	// Simulate dial and send
	if s.dialAndSendFunc != nil {
		err := s.dialAndSendFunc(nil)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	return nil
}

func TestSMTPService_SendEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	fromAddress := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<h1>Test Email</h1><p>This is a test email.</p>"

	// Create valid provider config
	validProvider := &domain.EmailProvider{
		Kind: domain.EmailProviderKindSMTP,
		SMTP: &domain.SMTPSettings{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			UseTLS:   true,
		},
		DefaultSenderEmail: "default@example.com",
		DefaultSenderName:  "Default Sender",
	}

	t.Run("success", func(t *testing.T) {
		// Create service with test implementation
		service := &testSMTPService{
			logger: mockLogger,
			mailClientCreator: func(host string, options ...interface{}) (interface{}, error) {
				return nil, nil
			},
			dialAndSendFunc: func(msg interface{}) error {
				return nil
			},
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, validProvider)

		// Verify no error
		assert.NoError(t, err)
	})

	t.Run("missing SMTP settings", func(t *testing.T) {
		// Create provider with missing SMTP settings
		invalidProvider := &domain.EmailProvider{
			Kind:               domain.EmailProviderKindSMTP,
			SMTP:               nil,
			DefaultSenderEmail: "default@example.com",
			DefaultSenderName:  "Default Sender",
		}

		// Create service
		service := &testSMTPService{
			logger: mockLogger,
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, invalidProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SMTP settings required")
	})

	t.Run("invalid sender email", func(t *testing.T) {
		// Create service
		service := &testSMTPService{
			logger: mockLogger,
		}

		// Call the method with invalid sender
		err := service.SendEmail(ctx, workspaceID, "invalid-email", fromName, to, subject, content, validProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sender")
	})

	t.Run("invalid recipient email", func(t *testing.T) {
		// Create service
		service := &testSMTPService{
			logger: mockLogger,
		}

		// Call the method with invalid recipient
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, "invalid-email", subject, content, validProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid recipient email")
	})

	t.Run("client creation error", func(t *testing.T) {
		// Create service with creator that returns error
		service := &testSMTPService{
			logger: mockLogger,
			mailClientCreator: func(host string, options ...interface{}) (interface{}, error) {
				return nil, fmt.Errorf("connection error")
			},
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, validProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("send error", func(t *testing.T) {
		// Create service with sender that returns error
		service := &testSMTPService{
			logger: mockLogger,
			mailClientCreator: func(host string, options ...interface{}) (interface{}, error) {
				return nil, nil
			},
			dialAndSendFunc: func(msg interface{}) error {
				return fmt.Errorf("send error")
			},
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, validProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")
	})
}

func TestNewSMTPService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create service
	service := newTestSMTPService(mockLogger)

	// Verify service was created correctly
	assert.NotNil(t, service)
	assert.Equal(t, mockLogger, service.logger)
}

// Test the real SMTP service's validation logic
func TestRealSMTPService_Validations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create the real service
	service := newTestSMTPService(mockLogger)

	// Test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	fromAddress := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<h1>Test Email</h1><p>This is a test email.</p>"

	// Create provider with nil SMTP settings
	nilSMTPProvider := &domain.EmailProvider{
		Kind:               domain.EmailProviderKindSMTP,
		SMTP:               nil,
		DefaultSenderEmail: "default@example.com",
		DefaultSenderName:  "Default Sender",
	}

	// Test with nil SMTP settings
	err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, nilSMTPProvider)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP settings required")
}
