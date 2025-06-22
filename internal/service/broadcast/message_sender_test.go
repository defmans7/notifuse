package broadcast

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	bmocks "github.com/Notifuse/notifuse/internal/service/broadcast/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/Notifuse/notifuse/pkg/notifuse_mjml"
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
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Test creating message sender with default config
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
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
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		customConfig,
		"",
	)

	assert.NotNil(t, customSender, "Message sender with custom config should not be nil")
	_, ok = customSender.(MessageSender)
	assert.True(t, ok, "Custom sender should implement MessageSender interface")
}

// Helper function to create a simple text block
func createTestTextBlock(id, textContent string) notifuse_mjml.EmailBlock {
	content := textContent
	return &notifuse_mjml.MJTextBlock{
		BaseBlock: notifuse_mjml.BaseBlock{
			ID:   id,
			Type: notifuse_mjml.MJMLComponentMjText,
		},
		Content: &content,
	}
}

// Helper function to create a valid MJML tree structure
func createValidTestTree(textBlock notifuse_mjml.EmailBlock) notifuse_mjml.EmailBlock {
	columnBlock := &notifuse_mjml.MJColumnBlock{
		BaseBlock: notifuse_mjml.BaseBlock{
			ID:       "col1",
			Type:     notifuse_mjml.MJMLComponentMjColumn,
			Children: []interface{}{textBlock},
		},
	}
	sectionBlock := &notifuse_mjml.MJSectionBlock{
		BaseBlock: notifuse_mjml.BaseBlock{
			ID:       "sec1",
			Type:     notifuse_mjml.MJMLComponentMjSection,
			Children: []interface{}{columnBlock},
		},
	}
	bodyBlock := &notifuse_mjml.MJBodyBlock{
		BaseBlock: notifuse_mjml.BaseBlock{
			ID:       "body1",
			Type:     notifuse_mjml.MJMLComponentMjBody,
			Children: []interface{}{sectionBlock},
		},
	}
	return &notifuse_mjml.MJMLBlock{
		BaseBlock: notifuse_mjml.BaseBlock{
			ID:         "root",
			Type:       notifuse_mjml.MJMLComponentMjml,
			Attributes: map[string]interface{}{"version": "4.0.0"},
			Children:   []interface{}{bodyBlock},
		},
	}
}

// TestSendToRecipientSuccess tests successful sending to a recipient
func TestSendToRecipientSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).Return().AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	tracking := true
	broadcast := &domain.Broadcast{
		ID:          "broadcast-123",
		WorkspaceID: workspaceID,
		Name:        "Test Broadcast",
		ChannelType: "email",
		Audience:    domain.AudienceSettings{Lists: []string{"list-1"}},
		Status:      domain.BroadcastStatusDraft,
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	emailProvider := &domain.EmailProvider{
		Kind:    domain.EmailProviderKindSMTP,
		Senders: []domain.EmailSender{emailSender},
		SMTP:    &domain.SMTPSettings{Host: "smtp.example.com", Port: 587, Username: "user", Password: "pass", UseTLS: true},
	}
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			SenderID:         emailSender.ID,
			Subject:          "Test Subject",
			VisualEditorTree: createValidTestTree(createTestTextBlock("txt1", "Test content")),
		},
	}

	// Setup mock expectations - SendToRecipient only calls emailService.SendEmail
	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			gomock.Any(), // messageID
			true,         // isMarketing
			gomock.Any(), // fromAddress
			gomock.Any(), // fromName
			gomock.Any(), // to
			gomock.Any(), // subject
			gomock.Any(), // content
			gomock.Any(), // emailProvider
			gomock.Any(), // emailOptions
		).Return(nil)

	// Create message sender
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		DefaultConfig(),
		"",
	)

	// Test
	err := sender.SendToRecipient(ctx, workspaceID, tracking, broadcast, "message-123", "test@example.com", template, map[string]interface{}{}, emailProvider)
	assert.NoError(t, err)
}

// TestSendToRecipientCompileFailure tests failure in template compilation
func TestSendToRecipientCompileFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).Return().AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	tracking := true
	broadcast := &domain.Broadcast{
		ID:          "broadcast-123",
		WorkspaceID: workspaceID,
		Name:        "Test Broadcast",
		ChannelType: "email",
		Audience:    domain.AudienceSettings{Lists: []string{"list-1"}},
		Status:      domain.BroadcastStatusDraft,
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	emailProvider := &domain.EmailProvider{
		Kind:    domain.EmailProviderKindSMTP,
		Senders: []domain.EmailSender{emailSender},
		SMTP:    &domain.SMTPSettings{Host: "smtp.example.com", Port: 587, Username: "user", Password: "pass", UseTLS: true},
	}
	// Create a template with empty VisualEditorTree that should cause compilation to fail
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			SenderID:         emailSender.ID,
			Subject:          "Test Subject",
			VisualEditorTree: &notifuse_mjml.MJMLBlock{}, // Empty block should cause compilation issues
		},
	}

	// Create message sender
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		DefaultConfig(),
		"",
	)

	// Test - this should fail due to template compilation issues
	err := sender.SendToRecipient(ctx, workspaceID, tracking, broadcast, "message-123", "test@example.com", template, map[string]interface{}{}, emailProvider)
	assert.Error(t, err)
	broadcastErr, ok := err.(*BroadcastError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeTemplateCompile, broadcastErr.Code)
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
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	recipientEmail := "test@example.com"
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			SenderID: emailSender.ID,
			Subject:  "Test Subject",
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
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			SenderID: emailSender.ID,
			Subject:  "Test Subject",
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).Return().AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	tracking := true
	broadcast := &domain.Broadcast{
		ID:          broadcastID,
		WorkspaceID: workspaceID,
		Name:        "Test Broadcast",
		ChannelType: "email",
		Audience:    domain.AudienceSettings{Lists: []string{"list-1"}},
		Status:      domain.BroadcastStatusDraft,
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
		TestSettings: domain.BroadcastTestSettings{
			Enabled: false,
			Variations: []domain.BroadcastVariation{
				{
					VariationName: "variation-1",
					TemplateID:    "template-123",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	emailProvider := &domain.EmailProvider{
		Kind:    domain.EmailProviderKindSMTP,
		Senders: []domain.EmailSender{emailSender},
		SMTP:    &domain.SMTPSettings{Host: "smtp.example.com", Port: 587, Username: "user", Password: "pass", UseTLS: true},
	}
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			SenderID:         emailSender.ID,
			Subject:          "Test Subject",
			VisualEditorTree: createValidTestTree(createTestTextBlock("txt1", "Test content")),
		},
	}

	// Setup mock expectations
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			gomock.Any(), // messageID
			true,         // isMarketing
			gomock.Any(), // fromAddress
			gomock.Any(), // fromName
			gomock.Any(), // to
			gomock.Any(), // subject
			gomock.Any(), // content
			gomock.Any(), // emailProvider
			gomock.Any(), // emailOptions
		).Return(nil).Times(2)

	mockMessageHistoryRepo.EXPECT().
		Create(
			ctx,
			workspaceID,
			gomock.Any(), // message
		).Return(nil).Times(2)

	// Create message sender
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		DefaultConfig(),
		"",
	)

	// Test
	recipients := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "recipient1@example.com",
			},
			ListID:   "list-1",
			ListName: "Test List",
		},
		{
			Contact: &domain.Contact{
				Email: "recipient2@example.com",
			},
			ListID:   "list-1",
			ListName: "Test List",
		},
	}
	templates := map[string]*domain.Template{"template-123": template}
	sent, failed, err := sender.SendBatch(ctx, workspaceID, "secret-key-123", tracking, broadcastID, recipients, templates, emailProvider)
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
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
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
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	emailProvider := &domain.EmailProvider{
		Kind:    domain.EmailProviderKindSMTP,
		Senders: []domain.EmailSender{emailSender},
	}
	// Create message sender
	config := DefaultConfig()
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		config,
		"",
	)

	// Call the method being tested with empty recipients
	sent, failed, err := sender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcastID, []*domain.ContactWithList{},
		map[string]*domain.Template{}, emailProvider)

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
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
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
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	emailProvider := &domain.EmailProvider{
		Kind:    domain.EmailProviderKindSMTP,
		Senders: []domain.EmailSender{emailSender},
		SMTP:    &domain.SMTPSettings{Host: "smtp.example.com", Port: 587, Username: "user", Password: "pass", UseTLS: true},
	}

	// Create message sender with circuit breaker enabled
	config := DefaultConfig()
	config.EnableCircuitBreaker = true
	config.CircuitBreakerThreshold = 1
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		config,
		"",
	)

	// Force circuit breaker to open
	messageSenderImpl := sender.(*messageSender)
	messageSenderImpl.circuitBreaker.RecordFailure(fmt.Errorf("test error"))

	// Call the method being tested
	sent, failed, err := sender.SendBatch(ctx, workspaceID, workspaceSecretKey, trackingEnabled, broadcastID, recipients,
		map[string]*domain.Template{}, emailProvider)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).Return().AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	tracking := true
	broadcast := &domain.Broadcast{
		ID:          broadcastID,
		WorkspaceID: workspaceID,
		Name:        "Test Broadcast",
		ChannelType: "email",
		Audience:    domain.AudienceSettings{Lists: []string{"list-1"}},
		Status:      domain.BroadcastStatusDraft,
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
		TestSettings: domain.BroadcastTestSettings{
			Enabled: false,
			Variations: []domain.BroadcastVariation{
				{
					VariationName: "variation-1",
					TemplateID:    "template-123",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	emailProvider := &domain.EmailProvider{
		Kind:    domain.EmailProviderKindSMTP,
		Senders: []domain.EmailSender{emailSender},
		SMTP:    &domain.SMTPSettings{Host: "smtp.example.com", Port: 587, Username: "user", Password: "pass", UseTLS: true},
	}
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			SenderID:         emailSender.ID,
			Subject:          "Test Subject",
			VisualEditorTree: createValidTestTree(createTestTextBlock("txt1", "Test content")),
		},
	}

	// Setup mock expectations
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			gomock.Any(), // messageID
			true,         // isMarketing
			gomock.Any(), // fromAddress
			gomock.Any(), // fromName
			gomock.Any(), // to
			gomock.Any(), // subject
			gomock.Any(), // content
			gomock.Any(), // emailProvider
			gomock.Any(), // emailOptions
		).Return(fmt.Errorf("email service unavailable")).Times(1)

	mockMessageHistoryRepo.EXPECT().
		Create(
			ctx,
			workspaceID,
			gomock.Any(), // message
		).Return(nil)

	// Create message sender
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		DefaultConfig(),
		"",
	)

	// Test
	recipients := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "recipient1@example.com",
			},
			ListID:   "list-1",
			ListName: "Test List",
		},
	}
	templates := map[string]*domain.Template{"template-123": template}
	sent, failed, err := sender.SendBatch(ctx, workspaceID, "secret-key-123", tracking, broadcastID, recipients, templates, emailProvider)
	assert.NoError(t, err)
	assert.Equal(t, 0, sent)
	assert.Equal(t, 1, failed)
}

// TestSendBatch_RecordMessageFails tests that SendBatch continues even if recording message history fails
func TestSendBatch_RecordMessageFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBroadcastRepository := mocks.NewMockBroadcastRepository(ctrl)
	mockMessageHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
	mockTemplateRepo := mocks.NewMockTemplateRepository(ctrl)
	mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Setup logger expectations
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).Return().AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).Return().AnyTimes()

	// Setup test data
	ctx := context.Background()
	workspaceID := "workspace-123"
	broadcastID := "broadcast-123"
	tracking := true
	broadcast := &domain.Broadcast{
		ID:          broadcastID,
		WorkspaceID: workspaceID,
		Name:        "Test Broadcast",
		ChannelType: "email",
		Audience:    domain.AudienceSettings{Lists: []string{"list-1"}},
		Status:      domain.BroadcastStatusDraft,
		UTMParameters: &domain.UTMParameters{
			Source:   "test",
			Medium:   "email",
			Campaign: "unit-test",
			Content:  "test-content",
			Term:     "test-term",
		},
		TestSettings: domain.BroadcastTestSettings{
			Enabled: false,
			Variations: []domain.BroadcastVariation{
				{
					VariationName: "variation-1",
					TemplateID:    "template-123",
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	emailSender := domain.NewEmailSender("sender@example.com", "Sender")
	emailProvider := &domain.EmailProvider{
		Kind:    domain.EmailProviderKindSMTP,
		Senders: []domain.EmailSender{emailSender},
		SMTP:    &domain.SMTPSettings{Host: "smtp.example.com", Port: 587, Username: "user", Password: "pass", UseTLS: true},
	}
	template := &domain.Template{
		ID: "template-123",
		Email: &domain.EmailTemplate{
			SenderID:         emailSender.ID,
			Subject:          "Test Subject",
			VisualEditorTree: createValidTestTree(createTestTextBlock("txt1", "Test content")),
		},
	}

	// Setup mock expectations
	mockBroadcastRepository.EXPECT().
		GetBroadcast(ctx, workspaceID, broadcastID).
		Return(broadcast, nil)

	mockEmailService.EXPECT().
		SendEmail(
			ctx,
			workspaceID,
			gomock.Any(), // messageID
			true,         // isMarketing
			gomock.Any(), // fromAddress
			gomock.Any(), // fromName
			gomock.Any(), // to
			gomock.Any(), // subject
			gomock.Any(), // content
			gomock.Any(), // emailProvider
			gomock.Any(), // emailOptions
		).Return(nil)

	mockMessageHistoryRepo.EXPECT().
		Create(
			ctx,
			workspaceID,
			gomock.Any(), // message
		).Return(fmt.Errorf("database connection error"))

	// Create message sender
	sender := NewMessageSender(
		mockBroadcastRepository,
		mockMessageHistoryRepo,
		mockTemplateRepo,
		mockEmailService,
		mockLogger,
		DefaultConfig(),
		"",
	)

	// Test
	recipients := []*domain.ContactWithList{
		{
			Contact: &domain.Contact{
				Email: "recipient1@example.com",
			},
			ListID:   "list-1",
			ListName: "Test List",
		},
	}
	templates := map[string]*domain.Template{"template-123": template}
	sent, failed, err := sender.SendBatch(ctx, workspaceID, "secret-key-123", tracking, broadcastID, recipients, templates, emailProvider)
	assert.NoError(t, err)
	assert.Equal(t, 1, sent)
	assert.Equal(t, 0, failed)
}
