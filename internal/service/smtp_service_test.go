package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClientFactory implements ClientFactory for testing
type MockClientFactory struct {
	NewClientFunc func(host string, port int, username, password string, useTLS bool) (MailClient, error)
}

func (m *MockClientFactory) NewClient(host string, port int, username, password string, useTLS bool) (MailClient, error) {
	if m.NewClientFunc != nil {
		return m.NewClientFunc(host, port, username, password, useTLS)
	}
	return nil, nil
}

// MockMailClient implements MailClient for testing
type MockMailClient struct {
	SendFunc  func(from, fromName, to, subject, content string) error
	CloseFunc func() error
}

func (m *MockMailClient) Send(from, fromName, to, subject, content string) error {
	if m.SendFunc != nil {
		return m.SendFunc(from, fromName, to, subject, content)
	}
	return nil
}

func (m *MockMailClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
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
		// Create mock client and factory
		mockClient := &MockMailClient{
			SendFunc: func(from, fromName, to, subject, content string) error {
				return nil
			},
		}

		mockFactory := &MockClientFactory{
			NewClientFunc: func(host string, port int, username, password string, useTLS bool) (MailClient, error) {
				return mockClient, nil
			},
		}

		// Create service with mocks
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
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
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: &MockClientFactory{},
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, invalidProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SMTP settings required")
	})

	t.Run("client creation error", func(t *testing.T) {
		// Create factory that returns error
		mockFactory := &MockClientFactory{
			NewClientFunc: func(host string, port int, username, password string, useTLS bool) (MailClient, error) {
				return nil, fmt.Errorf("connection error")
			},
		}

		// Create service with mock factory
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, validProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("send error", func(t *testing.T) {
		// Create mock client that returns error on send
		mockClient := &MockMailClient{
			SendFunc: func(from, fromName, to, subject, content string) error {
				return fmt.Errorf("send error")
			},
		}

		// Create factory that returns the mock client
		mockFactory := &MockClientFactory{
			NewClientFunc: func(host string, port int, username, password string, useTLS bool) (MailClient, error) {
				return mockClient, nil
			},
		}

		// Create service with mocks
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, validProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "send error")
	})
}

func TestNewSMTPService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create service
	service := NewSMTPService(mockLogger)

	// Verify service was created correctly
	assert.NotNil(t, service)
	assert.Equal(t, mockLogger, service.logger)
	assert.IsType(t, &defaultGoMailFactory{}, service.clientFactory)
}

func TestSMTPService_Validations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Create the service
	service := &SMTPService{
		logger:        mockLogger,
		clientFactory: &MockClientFactory{},
	}

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
