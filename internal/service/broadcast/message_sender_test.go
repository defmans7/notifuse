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
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Test creating message sender with default config
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		nil, // Passing nil config to test the default config behavior
		"",
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
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		customConfig,
		"",
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
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Create a logger that returns itself for chaining
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	trackingEnabled := true
	broadcast := &domain.Broadcast{
		ID: "broadcast-456",
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
	}
	recipientEmail := "test@example.com"

	// Use a simplified template without setting the VisualEditorTree field
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := domain.MapOfAny{
		"name": "Test User",
	}

	// Setup compiled template result
	compiledHTML := "<html><body>Hello Test User</body></html>"
	compiledTemplate := &domain.CompileTemplateResponse{
		Success: true,
		HTML:    &compiledHTML,
	}

	// Setup expectations
	// Mock CompileTemplate with the correct signature
	mockTemplateService.EXPECT().
		CompileTemplate(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, req domain.CompileTemplateRequest) (*domain.CompileTemplateResponse, error) {
			// Verify request fields
			assert.Equal(t, workspaceID, req.WorkspaceID)
			assert.Equal(t, templateData, req.TemplateData)
			assert.Equal(t, trackingEnabled, req.TrackingEnabled)
			return compiledTemplate, nil
		})

	// 2. Expect email sending
	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			true, // isMarketing
			template.Email.FromAddress,
			template.Email.FromName,
			recipientEmail,
			template.Email.Subject,
			compiledHTML,
			nil,
			"",  // replyTo
			nil, // cc
			nil, // bcc
		).
		Return(nil)

	// Create message sender with circuit breaker disabled
	config := DefaultConfig()
	config.EnableCircuitBreaker = false
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
		"",
	)

	// Call the method being tested
	messageID := "test-message-id"
	err := sender.SendToRecipient(ctx, workspaceID, trackingEnabled, broadcast, messageID, recipientEmail, template, templateData, nil)

	// Verify results
	assert.NoError(t, err)
}

// TestSendToRecipientCompileFailure tests failure in template compilation
func TestSendToRecipientCompileFailure(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
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
	trackingEnabled := true
	broadcast := &domain.Broadcast{
		ID: "broadcast-456",
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
	}
	messageID := "test-message-id"
	recipientEmail := "test@example.com"

	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := domain.MapOfAny{
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
		CompileTemplate(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, req domain.CompileTemplateRequest) (*domain.CompileTemplateResponse, error) {
			// Verify request fields
			assert.Equal(t, workspaceID, req.WorkspaceID)
			assert.Equal(t, templateData, req.TemplateData)
			assert.Equal(t, trackingEnabled, req.TrackingEnabled)
			return compiledTemplate, nil
		})

	// We should NOT call SendEmail since compilation failed
	// mockEmailService.EXPECT().SendEmail(...).Times(0)  // No need to explicitly set this with gomock

	// Create message sender with circuit breaker disabled
	config := DefaultConfig()
	config.EnableCircuitBreaker = false
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
		"",
	)

	// Call the method being tested
	err := sender.SendToRecipient(ctx, workspaceID, trackingEnabled, broadcast, messageID, recipientEmail, template, templateData, nil)

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
	workspaceSecretKey := "secret-key"
	trackingEnabled := true
	broadcast := &domain.Broadcast{
		ID: "broadcast-123",
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
	}
	recipientEmail := "test@example.com"
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := domain.MapOfAny{
		"name": "John",
	}

	// Set expectations on the mock
	messageID := "test-message-id"
	mockSender.EXPECT().
		SendToRecipient(ctx, workspaceID, trackingEnabled, broadcast, messageID, recipientEmail, template, templateData, nil).
		Return(nil)

	// Use the mock (normally this would be in the system under test)
	err := mockSender.SendToRecipient(ctx, workspaceID, trackingEnabled, broadcast, messageID, recipientEmail, template, templateData, nil)

	// Verify the result
	assert.NoError(t, err)

	// We can also set up expectations for SendBatch
	mockContacts := []*domain.ContactWithList{
		{Contact: &domain.Contact{Email: recipientEmail}},
	}
	mockTemplates := map[string]*domain.Template{
		"template-123": template,
	}

	// Set up expectations with specific return values
	mockSender.EXPECT().
		SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcast.ID, mockContacts, mockTemplates, nil).
		Return(1, 0, nil)

	// Use the mock
	sent, failed, err := mockSender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcast.ID, mockContacts, mockTemplates, nil)

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
	workspaceSecretKey := "secret-key"
	trackingEnabled := true
	broadcast := &domain.Broadcast{
		ID: "broadcast-123",
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
	}
	recipientEmail := "test@example.com"
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templateData := domain.MapOfAny{
		"name": "John",
	}

	// Set up mock to return an error
	mockError := errors.New("send failed: service unavailable")
	messageID := "test-message-id"
	mockSender.EXPECT().
		SendToRecipient(ctx, workspaceID, trackingEnabled, broadcast, messageID, recipientEmail, template, templateData, nil).
		Return(mockError)

	// Call the method
	err := mockSender.SendToRecipient(ctx, workspaceID, trackingEnabled, broadcast, messageID, recipientEmail, template, templateData, nil)

	// Verify error handling
	assert.Error(t, err)
	assert.Equal(t, mockError, err)
	assert.Contains(t, err.Error(), "service unavailable")

	// Test batch processing with error
	mockContacts := []*domain.ContactWithList{
		{Contact: &domain.Contact{Email: recipientEmail}},
	}
	mockTemplates := map[string]*domain.Template{
		"template-123": template,
	}
	batchError := errors.New("batch processing failed")

	mockSender.EXPECT().
		SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcast.ID, mockContacts, mockTemplates, nil).
		Return(0, 0, batchError)

	sent, failed, err := mockSender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcast.ID, mockContacts, mockTemplates, nil)
	assert.Error(t, err)
	assert.Equal(t, batchError, err)
	assert.Equal(t, 0, sent)
	assert.Equal(t, 0, failed)
}

// TestSendBatch tests the SendBatch method
func TestSendBatch(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	workspaceSecretKey := "secret-key"
	trackingEnabled := true
	broadcastID := "broadcast-456"
	apiEndpoint := "https://api.example.com"

	// Create contacts with lists
	recipients := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "recipient1@example.com",
			},
			ListID:   "list-1",
			ListName: "Test List 1",
		},
		{
			Contact: &domain.Contact{
				Email: "recipient2@example.com",
			},
			ListID:   "list-2",
			ListName: "Test List 2",
		},
	}

	// Create templates
	template := &domain.Template{
		ID:      "template-123",
		Version: 2,
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templates := map[string]*domain.Template{
		"template-123": template,
	}

	// Create A/B test variations
	variations := []domain.BroadcastVariation{
		{
			ID:         "var-1",
			TemplateID: "template-123",
		},
	}

	// Create broadcast
	broadcast := &domain.Broadcast{
		ID:               broadcastID,
		WorkspaceID:      workspaceID,
		WinningVariation: "",
		TestSettings: domain.BroadcastTestSettings{
			Enabled:    false,
			Variations: variations,
		},
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
	}

	// Set up expectations for GetBroadcast and GetAPIEndpoint
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	// Setup compiled template result
	compiledHTML := "<html><body>Hello User</body></html>"
	compiledTemplate := &domain.CompileTemplateResponse{
		Success: true,
		HTML:    &compiledHTML,
	}

	// Expect template compilation and email sending for each recipient
	for _, recipient := range recipients {
		mockTemplateService.EXPECT().
			CompileTemplate(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, req domain.CompileTemplateRequest) (*domain.CompileTemplateResponse, error) {
				// Verify request fields
				assert.Equal(t, workspaceID, req.WorkspaceID)
				assert.Equal(t, trackingEnabled, req.TrackingEnabled)
				return compiledTemplate, nil
			})

		mockEmailService.EXPECT().
			SendEmail(
				ctx,
				workspaceID,
				true, // isMarketing
				template.Email.FromAddress,
				template.Email.FromName,
				recipient.Contact.Email,
				template.Email.Subject,
				compiledHTML,
				nil,
				"",  // replyTo
				nil, // cc
				nil, // bcc
			).
			Return(nil)

		// Expect message history recording
		mockMessageHistoryRepo.EXPECT().
			Create(ctx, workspaceID, gomock.Any()).
			Do(func(_ context.Context, _ string, message *domain.MessageHistory) {
				// Verify the message history is correct
				assert.Equal(t, recipient.Contact.Email, message.ContactEmail)
				assert.Equal(t, broadcastID, *message.BroadcastID)
				assert.Equal(t, "template-123", message.TemplateID)
				assert.Equal(t, int(template.Version), message.TemplateVersion)
				assert.Equal(t, "email", message.Channel)
				assert.Equal(t, domain.MessageStatusSent, message.Status)

				// Verify message data
				assert.Contains(t, message.MessageData.Data, "broadcast_id")
				assert.Contains(t, message.MessageData.Data, "email")
				assert.Contains(t, message.MessageData.Data, "template_id")

				// Verify timestamps
				assert.NotZero(t, message.SentAt)
				assert.NotZero(t, message.CreatedAt)
				assert.NotZero(t, message.UpdatedAt)

				// Verify ID has correct format (should contain workspace_ID followed by UUID)
				assert.True(t, len(message.ID) > len(workspaceID))
				assert.Contains(t, message.ID, workspaceID)
			}).
			Return(nil)
	}

	// Create message sender
	config := DefaultConfig()
	config.EnableCircuitBreaker = false
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
		apiEndpoint,
	)

	// Call the method being tested
	sent, failed, err := sender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcastID, recipients, templates, nil)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 2, sent)
	assert.Equal(t, 0, failed)
}

// TestSendBatch_EmptyRecipients tests SendBatch with no recipients
func TestSendBatch_EmptyRecipients(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	workspaceSecretKey := "secret-key"
	trackingEnabled := true
	broadcastID := "broadcast-456"

	// Create message sender
	config := DefaultConfig()
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
		"",
	)

	// Call the method being tested with empty recipients
	sent, failed, err := sender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcastID, []*domain.ContactWithList{},
		map[string]*domain.Template{}, nil)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 0, sent)
	assert.Equal(t, 0, failed)
}

// TestSendBatch_CircuitBreakerOpen tests SendBatch when circuit breaker is open
func TestSendBatch_CircuitBreakerOpen(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	workspaceSecretKey := "secret-key"
	trackingEnabled := true
	broadcastID := "broadcast-456"
	recipients := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "test@example.com",
			},
		},
	}

	// Create message sender with circuit breaker enabled
	config := DefaultConfig()
	config.EnableCircuitBreaker = true
	config.CircuitBreakerThreshold = 1
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
		"",
	)

	// Force circuit breaker to open
	messageSenderImpl := sender.(*messageSender)
	messageSenderImpl.circuitBreaker.RecordFailure()

	// Call the method being tested
	sent, failed, err := sender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcastID, recipients,
		map[string]*domain.Template{}, nil)

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, 0, sent)
	assert.Equal(t, 0, failed)

	// Check that we got the right error
	broadcastErr, ok := err.(*BroadcastError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeCircuitOpen, broadcastErr.Code)
}

// TestGenerateMessageID tests the generateMessageID function
func TestGenerateMessageID(t *testing.T) {
	workspaceID := "workspace-123"

	// Generate multiple message IDs
	id1 := generateMessageID(workspaceID)
	id2 := generateMessageID(workspaceID)

	// Check that IDs have the correct format (workspace_uuid)
	assert.Contains(t, id1, workspaceID+"_")
	assert.Contains(t, id2, workspaceID+"_")

	// Check that generated IDs are different
	assert.NotEqual(t, id1, id2)

	// Verify length is reasonable (workspace ID + "_" + UUID)
	expectedMinLength := len(workspaceID) + 1 + 32 // UUID strings are at least 32 chars
	assert.Greater(t, len(id1), expectedMinLength)
}

// TestSendBatch_WithFailure tests SendBatch with a failed email send
func TestSendBatch_WithFailure(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	workspaceSecretKey := "secret-key"
	trackingEnabled := true
	broadcastID := "broadcast-456"
	apiEndpoint := "https://api.example.com"

	// Create a single contact with list
	recipients := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "recipient1@example.com",
			},
			ListID:   "list-1",
			ListName: "Test List 1",
		},
	}

	// Create template
	template := &domain.Template{
		ID:      "template-123",
		Version: 2,
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templates := map[string]*domain.Template{
		"template-123": template,
	}

	// Create A/B test variations
	variations := []domain.BroadcastVariation{
		{
			ID:         "var-1",
			TemplateID: "template-123",
		},
	}

	// Create broadcast
	broadcast := &domain.Broadcast{
		ID:               broadcastID,
		WorkspaceID:      workspaceID,
		WinningVariation: "",
		TestSettings: domain.BroadcastTestSettings{
			Enabled:    false,
			Variations: variations,
		},
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
	}

	// Set up expectations for GetBroadcast and GetAPIEndpoint
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	// Setup compiled template result
	compiledHTML := "<html><body>Hello User</body></html>"
	compiledTemplate := &domain.CompileTemplateResponse{
		Success: true,
		HTML:    &compiledHTML,
	}

	// Create error for email sending
	sendError := errors.New("email service unavailable")

	// Expect template compilation to succeed but email sending to fail
	mockTemplateService.EXPECT().
		CompileTemplate(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, req domain.CompileTemplateRequest) (*domain.CompileTemplateResponse, error) {
			// Verify request fields
			assert.Equal(t, workspaceID, req.WorkspaceID)
			assert.Equal(t, trackingEnabled, req.TrackingEnabled)
			return compiledTemplate, nil
		})

	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			true, // isMarketing
			template.Email.FromAddress,
			template.Email.FromName,
			recipients[0].Contact.Email,
			template.Email.Subject,
			compiledHTML,
			nil,
			"",  // replyTo
			nil, // cc
			nil, // bcc
		).
		Return(sendError)

	// Expect message history recording with failed status
	mockMessageHistoryRepo.EXPECT().
		Create(ctx, workspaceID, gomock.Any()).
		Do(func(_ context.Context, _ string, message *domain.MessageHistory) {
			// Verify the message history is correct
			assert.Equal(t, recipients[0].Contact.Email, message.ContactEmail)
			assert.Equal(t, broadcastID, *message.BroadcastID)
			assert.Equal(t, "template-123", message.TemplateID)
			assert.Equal(t, int(template.Version), message.TemplateVersion)
			assert.Equal(t, "email", message.Channel)
			assert.Equal(t, domain.MessageStatusFailed, message.Status)

			// Verify error is stored
			assert.NotNil(t, message.Error)
			assert.Contains(t, *message.Error, "email service unavailable")

			// Verify message data
			assert.Contains(t, message.MessageData.Data, "broadcast_id")
			assert.Contains(t, message.MessageData.Data, "email")
			assert.Contains(t, message.MessageData.Data, "template_id")

			// Verify timestamps
			assert.NotZero(t, message.SentAt)
			assert.NotZero(t, message.CreatedAt)
			assert.NotZero(t, message.UpdatedAt)
		}).
		Return(nil)

	// Create message sender
	config := DefaultConfig()
	config.EnableCircuitBreaker = false
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
		apiEndpoint,
	)

	// Call the method being tested
	sent, failed, err := sender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcastID, recipients, templates, nil)

	// Verify results
	assert.NoError(t, err) // The overall operation shouldn't fail even if individual sends fail
	assert.Equal(t, 0, sent)
	assert.Equal(t, 1, failed)
}

// TestSendBatch_RecordMessageFails tests that SendBatch continues even if recording message history fails
func TestSendBatch_RecordMessageFails(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks for all dependencies
	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateService := mocks.NewMockTemplateService(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	workspaceSecretKey := "secret-key"
	trackingEnabled := true
	broadcastID := "broadcast-456"
	apiEndpoint := "https://api.example.com"

	// Create contact with list
	recipients := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "recipient1@example.com",
			},
			ListID:   "list-1",
			ListName: "Test List 1",
		},
	}

	// Create template
	template := &domain.Template{
		ID:      "template-123",
		Version: 2,
		Email: &domain.EmailTemplate{
			FromAddress: "sender@example.com",
			FromName:    "Sender",
			Subject:     "Test Subject",
		},
	}
	templates := map[string]*domain.Template{
		"template-123": template,
	}

	// Create A/B test variations
	variations := []domain.BroadcastVariation{
		{
			ID:         "var-1",
			TemplateID: "template-123",
		},
	}

	// Create broadcast
	broadcast := &domain.Broadcast{
		ID:               broadcastID,
		WorkspaceID:      workspaceID,
		WinningVariation: "",
		TestSettings: domain.BroadcastTestSettings{
			Enabled:    false,
			Variations: variations,
		},
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
	}

	// Set up expectations for GetBroadcast and GetAPIEndpoint
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	// Setup compiled template result
	compiledHTML := "<html><body>Hello User</body></html>"
	compiledTemplate := &domain.CompileTemplateResponse{
		Success: true,
		HTML:    &compiledHTML,
	}

	// Expect template compilation and email sending to succeed
	mockTemplateService.EXPECT().
		CompileTemplate(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, req domain.CompileTemplateRequest) (*domain.CompileTemplateResponse, error) {
			// Verify request fields
			assert.Equal(t, workspaceID, req.WorkspaceID)
			assert.Equal(t, trackingEnabled, req.TrackingEnabled)
			return compiledTemplate, nil
		})

	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			true, // isMarketing
			template.Email.FromAddress,
			template.Email.FromName,
			recipients[0].Contact.Email,
			template.Email.Subject,
			compiledHTML,
			nil,
			"",  // replyTo
			nil, // cc
			nil, // bcc
		).
		Return(nil)

	// Create error for message recording
	recordError := errors.New("database connection error")

	// Expect message history recording to fail
	mockMessageHistoryRepo.EXPECT().
		Create(ctx, workspaceID, gomock.Any()).
		Return(recordError)

	// Create message sender
	config := DefaultConfig()
	config.EnableCircuitBreaker = false
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateService,
		mockEmailService,
		mockLogger,
		config,
		apiEndpoint,
	)

	// Call the method being tested
	sent, failed, err := sender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcastID, recipients, templates, nil)

	// Verify results - the SendBatch should still succeed even if recording failed
	assert.NoError(t, err)
	assert.Equal(t, 1, sent)
	assert.Equal(t, 0, failed)
}
