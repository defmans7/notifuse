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

// MockSMTPClientFactory is a mock implementation of domain.SMTPClientFactory for testing
type MockSMTPClientFactory struct {
	createErr error
}

// CreateClient returns a mock for testing
func (f *MockSMTPClientFactory) CreateClient(host string, port int, username, password string, useTLS bool) (*mail.Client, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}

	// For testing, we don't actually need to return a real client
	// In a real test, we might use a test double or a mocking library with monkey patching
	// But for now, just return nil and handle in the tests
	return nil, nil
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
		mockFactory := &MockSMTPClientFactory{}

		// Create service with mock factory
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, "test-message-id", fromAddress, fromName, to, subject, content, invalidProvider, "", nil, nil)

		// Verify error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SMTP settings required")
	})

	t.Run("client creation error", func(t *testing.T) {
		// Create a mock factory that returns an error
		mockFactory := &MockSMTPClientFactory{
			createErr: fmt.Errorf("connection error"),
		}

		// Create service
		service := &SMTPService{
			logger:        mockLogger,
			clientFactory: mockFactory,
		}

		// Call the method
		err := service.SendEmail(ctx, workspaceID, "test-message-id", fromAddress, fromName, to, subject, content, validProvider, "", nil, nil)

		// Verify error
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
