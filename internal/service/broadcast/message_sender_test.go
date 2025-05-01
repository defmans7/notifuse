package broadcast

import (
	"context"
	"errors"
	"testing"
	"time"

	mjmlgo "github.com/Boostport/mjml-go"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	bmocks "github.com/Notifuse/notifuse/internal/service/broadcast/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestMessageSenderCreation tests creation of the message sender
func TestMessageSenderCreation(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Test creating message sender with default config
	sender := NewMessageSender(
		mockBroadcastService,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		nil, // Passing nil config to test the default config behavior
	)

	// Assert the sender was created and implements the interface
	assert.NotNil(t, sender, "Message sender should not be nil")
	_, ok := sender.(MessageSender)
	assert.True(t, ok, "Sender should implement MessageSender interface")

	// Test creating with custom config
	customConfig := &Config{
		MaxParallelism:          5,
		MaxProcessTime:          30 * time.Second,
		DefaultRateLimit:        300, // 5 per second
		EnableCircuitBreaker:    true,
		CircuitBreakerThreshold: 3,
		CircuitBreakerCooldown:  30 * time.Second,
	}

	customSender := NewMessageSender(
		mockBroadcastService,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		customConfig,
	)

	assert.NotNil(t, customSender, "Message sender with custom config should not be nil")
	_, ok = customSender.(MessageSender)
	assert.True(t, ok, "Custom sender should implement MessageSender interface")
}

// TestSendToRecipientSuccess tests successful sending to a recipient
func TestSendToRecipientSuccess(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a logger that returns itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	recipient := &domain.Contact{
		Email: "test@example.com",
	}

	// Use a simplified template without setting the VisualEditorTree field
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := map[string]interface{}{
		"name": "Test User",
	}

	// Setup compiled template result
	compiledHTML := "<html><body>Hello Test User</body></html>"
	compiledTemplate := &domain.CompileTemplateResponse{
		Success: true,
		HTML:    &compiledHTML,
	}

	// Setup expectations
	// Mock CompileTemplate to accept any value for the third parameter and return our test response
	mockTemplateService.EXPECT().
		CompileTemplate(ctx, workspaceID, gomock.Any(), templateData).
		Return(compiledTemplate, nil)

	// 2. Expect email sending
	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			"marketing",
			template.Email.FromAddress,
			template.Email.FromName,
			recipient.Email,
			template.Email.Subject,
			compiledHTML,
		).
		Return(nil)

	// Create message sender with circuit breaker disabled
	config := DefaultConfig()
	config.EnableCircuitBreaker = false
	sender := NewMessageSender(
		mockBroadcastService,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
	)

	// Call the method being tested
	err := sender.SendToRecipient(ctx, workspaceID, broadcastID, recipient, template, templateData)

	// Verify results
	assert.NoError(t, err)
}

// TestSendToRecipientCompileFailure tests failure in template compilation
func TestSendToRecipientCompileFailure(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastService := mocks.NewMockBroadcastSender(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a logger that returns itself for chaining and allow all logger methods
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-456"
	recipient := &domain.Contact{
		Email: "test@example.com",
	}

	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := map[string]interface{}{
		"name": "Test User",
	}

	// Create a failed compile result
	errorMsg := "Template compilation failed: invalid liquid syntax"
	compiledTemplate := &domain.CompileTemplateResponse{
		Success: false,
		HTML:    nil,
		Error: &mjmlgo.Error{
			Message: errorMsg,
		},
	}

	// Setup expectations
	mockTemplateService.EXPECT().
		CompileTemplate(ctx, workspaceID, gomock.Any(), templateData).
		Return(compiledTemplate, nil)

	// We should NOT call SendEmail since compilation failed
	// mockEmailService.EXPECT().SendEmail(...).Times(0)  // No need to explicitly set this with gomock

	// Create message sender with circuit breaker disabled
	config := DefaultConfig()
	config.EnableCircuitBreaker = false
	sender := NewMessageSender(
		mockBroadcastService,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
	)

	// Call the method being tested
	err := sender.SendToRecipient(ctx, workspaceID, broadcastID, recipient, template, templateData)

	// Verify error is returned
	assert.Error(t, err)
	broadcastErr, ok := err.(*BroadcastError)
	assert.True(t, ok, "Error should be of type BroadcastError")
	assert.Equal(t, ErrCodeTemplateCompile, broadcastErr.Code)
	assert.Equal(t, errorMsg, broadcastErr.Message)
}

// TestWithMockMessageSender shows how to use the MockMessageSender
func TestWithMockMessageSender(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock message sender
	mockSender := bmocks.NewMockMessageSender(ctrl)

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	contact := &domain.Contact{
		Email: "test@example.com",
	}
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := map[string]interface{}{
		"name": "John",
	}

	// Set expectations on the mock
	mockSender.EXPECT().
		SendToRecipient(ctx, workspaceID, broadcastID, contact, template, templateData).
		Return(nil)

	// Use the mock (normally this would be in the system under test)
	err := mockSender.SendToRecipient(ctx, workspaceID, broadcastID, contact, template, templateData)

	// Verify the result
	assert.NoError(t, err)

	// We can also set up expectations for SendBatch
	mockContacts := []*domain.ContactWithList{
		{Contact: contact},
	}
	mockTemplates := map[string]*domain.Template{
		"template-123": template,
	}

	// Set up expectations with specific return values
	mockSender.EXPECT().
		SendBatch(ctx, workspaceID, broadcastID, mockContacts, mockTemplates, templateData).
		Return(1, 0, nil)

	// Use the mock
	sent, failed, err := mockSender.SendBatch(ctx, workspaceID, broadcastID, mockContacts, mockTemplates, templateData)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 1, sent)
	assert.Equal(t, 0, failed)
}

// TestErrorHandlingWithMock demonstrates error handling with mocks
func TestErrorHandlingWithMock(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock message sender
	mockSender := bmocks.NewMockMessageSender(ctrl)

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	contact := &domain.Contact{
		Email: "test@example.com",
	}
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := map[string]interface{}{
		"name": "John",
	}

	// Set up mock to return an error
	mockError := errors.New("send failed: service unavailable")
	mockSender.EXPECT().
		SendToRecipient(ctx, workspaceID, broadcastID, contact, template, templateData).
		Return(mockError)

	// Call the method
	err := mockSender.SendToRecipient(ctx, workspaceID, broadcastID, contact, template, templateData)

	// Verify error handling
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
	assert.Contains(t, err.Error(), "service unavailable")

	// Test batch processing with error
	mockContacts := []*domain.ContactWithList{
		{Contact: contact},
	}
	mockTemplates := map[string]*domain.Template{
		"template-123": template,
	}
	batchError := errors.New("batch processing failed")

	mockSender.EXPECT().
		SendBatch(ctx, workspaceID, broadcastID, mockContacts, mockTemplates, templateData).
		Return(0, 0, batchError)

	sent, failed, err := mockSender.SendBatch(ctx, workspaceID, broadcastID, mockContacts, mockTemplates, templateData)
	assert.Error(t, err)
	assert.Equal(t, batchError, err)
	assert.Equal(t, 0, sent)
	assert.Equal(t, 0, failed)
}
