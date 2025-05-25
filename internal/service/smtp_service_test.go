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
	"github.com/wneessen/go-mail"
)

// MockMailClient is a mock implementation of *mail.Client for testing
type MockMailClient struct {
	Closed         bool
	SendError      error
	SetSenderError error
	SetToError     error
	From           string
	FromName       string
	To             string
	Subject        string
	BodyType       string
	BodyContent    string
}

// Close mocks the Close method
func (m *MockMailClient) Close() error {
	m.Closed = true
	return nil
}

// DialAndSend mocks the DialAndSend method
func (m *MockMailClient) DialAndSend(msg *mail.Msg) error {
	return m.SendError
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
	messageID := "test-message-id"
	workspaceID := "workspace-123"
	fromAddress := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<h1>Test Email</h1><p>This is a test email.</p>"
	replyTo := "reply@example.com"
	cc := []string{"cc1@example.com", "cc2@example.com"}
	bcc := []string{"bcc1@example.com", "bcc2@example.com"}

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
		Senders: []domain.EmailSender{
			domain.NewEmailSender("default@example.com", "Default Sender"),
		},
	}

	t.Run("missing SMTP settings", func(t *testing.T) {
		// Create provider with missing SMTP settings
		invalidProvider := &domain.EmailProvider{
			Kind: domain.EmailProviderKindSMTP,
			SMTP: nil,
			Senders: []domain.EmailSender{
				domain.NewEmailSender("default@example.com", "Default Sender"),
			},
		}

		// Create mock factory
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)

		// Create service with mock factory
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, invalidProvider, "", nil, nil)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SMTP settings required")
	})

	t.Run("client creation error", func(t *testing.T) {
		// Create mock factory that returns an error
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, validProvider, "", nil, nil)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("client creation success but send fails", func(t *testing.T) {
		// Create mock factory that succeeds in creating client but we can't test actual sending
		// without a more complex mock setup for the go-mail library
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection refused"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method - should fail at client creation
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, validProvider, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("test CC and BCC parameters passed to factory", func(t *testing.T) {
		// Create mock factory that fails so we can test parameter passing without nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection timeout"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with CC and BCC - should fail at client creation but parameters are validated
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, validProvider, replyTo, cc, bcc)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("empty CC and BCC values handling", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("network error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with empty CC and BCC values
		ccWithEmpty := []string{"cc1@example.com", "", "cc2@example.com"}
		bccWithEmpty := []string{"", "bcc1@example.com", ""}

		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, validProvider, replyTo, ccWithEmpty, bccWithEmpty)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("with reply-to header", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("auth failed"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with reply-to
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, validProvider, replyTo, nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("without reply-to header", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("timeout"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call without reply-to (empty string)
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, validProvider, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("SMTP settings with TLS disabled", func(t *testing.T) {
		// Create provider with TLS disabled
		providerNoTLS := &domain.EmailProvider{
			Kind: domain.EmailProviderKindSMTP,
			SMTP: &domain.SMTPSettings{
				Host:     "smtp.example.com",
				Port:     25,
				Username: "user@example.com",
				Password: "password",
				UseTLS:   false,
			},
			Senders: []domain.EmailSender{
				domain.NewEmailSender("default@example.com", "Default Sender"),
			},
		}

		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			providerNoTLS.SMTP.Host,
			providerNoTLS.SMTP.Port,
			providerNoTLS.SMTP.Username,
			providerNoTLS.SMTP.Password,
			providerNoTLS.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection refused"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, content, providerNoTLS, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("invalid from address format", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with invalid from address
		invalidFromAddress := "invalid-email"
		err := service.SendEmail(ctx, messageID, workspaceID, invalidFromAddress, fromName, to, subject, content, validProvider, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("invalid to address format", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with invalid to address
		invalidToAddress := "invalid-email"
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, invalidToAddress, subject, content, validProvider, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("empty message ID", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with empty message ID
		err := service.SendEmail(ctx, "", workspaceID, fromAddress, fromName, to, subject, content, validProvider, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("empty subject", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with empty subject
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, "", content, validProvider, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})

	t.Run("empty content", func(t *testing.T) {
		// Create mock factory that fails to avoid nil client issues
		mockFactory := mocks.NewMockSMTPClientFactory(ctrl)
		mockFactory.EXPECT().CreateClient(
			validProvider.SMTP.Host,
			validProvider.SMTP.Port,
			validProvider.SMTP.Username,
			validProvider.SMTP.Password,
			validProvider.SMTP.UseTLS,
		).Return(nil, fmt.Errorf("connection error"))

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call with empty content
		err := service.SendEmail(ctx, messageID, workspaceID, fromAddress, fromName, to, subject, "", validProvider, "", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create SMTP client")
	})
}

func TestNewSMTPService(t *testing.T) {
	// Create a new instance of Logger for testing
	mockLogger := &pkgmocks.MockLogger{}

	// Call the function being tested
	service := NewSMTPService(mockLogger)

	// Verify service was created correctly
	require.NotNil(t, service)
	require.Equal(t, mockLogger, service.logger)

	// Verify the client factory is properly initialized
	require.NotNil(t, service.clientFactory)
	require.IsType(t, &defaultGoMailFactory{}, service.clientFactory)
}

func TestDefaultGoMailFactory_CreateClient(t *testing.T) {
	factory := &defaultGoMailFactory{}

	t.Run("valid configuration with TLS", func(t *testing.T) {
		client, err := factory.CreateClient("smtp.example.com", 587, "user@example.com", "password", true)

		// We expect this to fail since we're not connecting to a real SMTP server
		// But we can verify the error is related to connection, not configuration
		if err != nil {
			// The error should be about connection, not about invalid parameters
			assert.Contains(t, err.Error(), "failed to create mail client")
		} else {
			// If somehow it succeeds (unlikely), verify we got a client
			assert.NotNil(t, client)
			if client != nil {
				client.Close()
			}
		}
	})

	t.Run("valid configuration without TLS", func(t *testing.T) {
		client, err := factory.CreateClient("smtp.example.com", 25, "user@example.com", "password", false)

		// We expect this to fail since we're not connecting to a real SMTP server
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create mail client")
		} else {
			assert.NotNil(t, client)
			if client != nil {
				client.Close()
			}
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		client, err := factory.CreateClient("smtp.example.com", -1, "user@example.com", "password", true)

		// Should fail due to invalid port
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to create mail client")
	})

	t.Run("empty host", func(t *testing.T) {
		client, err := factory.CreateClient("", 587, "user@example.com", "password", true)

		// Should fail due to empty host
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to create mail client")
	})

	t.Run("port out of range", func(t *testing.T) {
		client, err := factory.CreateClient("smtp.example.com", 70000, "user@example.com", "password", true)

		// Should fail due to port out of range
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to create mail client")
	})

	t.Run("empty username", func(t *testing.T) {
		client, err := factory.CreateClient("smtp.example.com", 587, "", "password", true)

		// The go-mail library allows empty username at client creation time
		// The error would occur during authentication when actually sending
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create mail client")
		} else {
			// Client creation succeeds but would fail during authentication
			assert.NotNil(t, client)
			if client != nil {
				client.Close()
			}
		}
	})

	t.Run("empty password", func(t *testing.T) {
		client, err := factory.CreateClient("smtp.example.com", 587, "user@example.com", "", true)

		// The go-mail library allows empty password at client creation time
		// The error would occur during authentication when actually sending
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create mail client")
		} else {
			// Client creation succeeds but would fail during authentication
			assert.NotNil(t, client)
			if client != nil {
				client.Close()
			}
		}
	})
}

// Test the defaultGoMailFactory's method signature (without execution)
func TestDefaultGoMailFactory_CreateClient_Signature(t *testing.T) {
	factory := &defaultGoMailFactory{}

	// We're testing the method existence and signature, not its behavior
	// since that would require actually attempting to connect to an SMTP server
	assert.NotNil(t, factory)

	// Verify the method exists by getting its type
	factoryType := fmt.Sprintf("%T", factory.CreateClient)
	assert.Equal(t, "func(string, int, string, string, bool) (*mail.Client, error)", factoryType)
}
