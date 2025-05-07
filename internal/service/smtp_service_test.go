package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMailClientAdapter adapts domain/mocks.MockSMTPClient to the service.MailClient interface
type MockMailClientAdapter struct {
	mockClient *mocks.MockSMTPClient
}

func (a *MockMailClientAdapter) Send(from, fromName, to, subject, content string) error {
	if err := a.mockClient.SetSender(from, fromName); err != nil {
		return err
	}
	if err := a.mockClient.SetRecipient(to); err != nil {
		return err
	}
	a.mockClient.SetSubject(subject)
	a.mockClient.SetBodyString("text/html", content)
	return a.mockClient.DialAndSend()
}

func (a *MockMailClientAdapter) Close() error {
	return a.mockClient.Close()
}

// MockFactoryAdapter adapts domain/mocks.MockSMTPClientFactory to the service.ClientFactory interface
type MockFactoryAdapter struct {
	mockFactory *mocks.MockSMTPClientFactory
}

func (a *MockFactoryAdapter) NewClient(host string, port int, username, password string, useTLS bool) (MailClient, error) {
	// Create options slice similar to what would be passed to the real factory
	options := []interface{}{
		username,
		password,
		useTLS,
	}

	// Call the underlying mock factory
	client, err := a.mockFactory.NewClient(host, port, options...)
	if err != nil {
		return nil, err
	}

	// Adapt the client to our interface - need explicit type assertion
	mockClient, ok := client.(*mocks.MockSMTPClient)
	if !ok {
		return nil, fmt.Errorf("unexpected client type from factory")
	}

	return &MockMailClientAdapter{mockClient: mockClient}, nil
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
		// Create mock SMTP client
		mockSMTPClient := mocks.NewMockSMTPClient(ctrl)
		mockSMTPClient.EXPECT().SetSender(fromAddress, fromName).Return(nil)
		mockSMTPClient.EXPECT().SetRecipient(to).Return(nil)
		mockSMTPClient.EXPECT().SetSubject(subject).Return(nil)
		mockSMTPClient.EXPECT().SetBodyString("text/html", content).Return(nil)
		mockSMTPClient.EXPECT().DialAndSend().Return(nil)
		mockSMTPClient.EXPECT().Close().Return(nil)

		// Create mock factory
		mockSMTPFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockSMTPFactory.EXPECT().NewClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(mockSMTPClient, nil)

		// Create service with mocks using adapter
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: &MockFactoryAdapter{mockFactory: mockSMTPFactory},
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

		// Create mock factory
		mockSMTPFactory := mocks.NewMockSMTPClientFactory(ctrl)

		// Create service with mock factory
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: &MockFactoryAdapter{mockFactory: mockSMTPFactory},
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, invalidProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SMTP settings required")
	})

	t.Run("client creation error", func(t *testing.T) {
		// Create mock factory that returns error
		mockSMTPFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockSMTPFactory.EXPECT().NewClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(nil, fmt.Errorf("connection error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: &MockFactoryAdapter{mockFactory: mockSMTPFactory},
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, validProvider)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("send error", func(t *testing.T) {
		// Create mock SMTP client that returns error on DialAndSend
		mockSMTPClient := mocks.NewMockSMTPClient(ctrl)
		mockSMTPClient.EXPECT().SetSender(fromAddress, fromName).Return(nil)
		mockSMTPClient.EXPECT().SetRecipient(to).Return(nil)
		mockSMTPClient.EXPECT().SetSubject(subject).Return(nil)
		mockSMTPClient.EXPECT().SetBodyString("text/html", content).Return(nil)
		mockSMTPClient.EXPECT().DialAndSend().Return(fmt.Errorf("send error"))
		mockSMTPClient.EXPECT().Close().Return(nil)

		// Create mock factory
		mockSMTPFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockSMTPFactory.EXPECT().NewClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(mockSMTPClient, nil)

		// Create service with mocks
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: &MockFactoryAdapter{mockFactory: mockSMTPFactory},
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

	// Create mock factory
	mockSMTPFactory := mocks.NewMockSMTPClientFactory(ctrl)

	// Create the service
	service := &SMTPService{
		logger:        mockLogger,
		clientFactory: &MockFactoryAdapter{mockFactory: mockSMTPFactory},
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
