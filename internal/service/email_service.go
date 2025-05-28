package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/mjml"
	"github.com/Notifuse/notifuse/pkg/tracing"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/google/uuid"
	"go.opencensus.io/trace"
)

type EmailService struct {
	logger           logger.Logger
	authService      domain.AuthService
	secretKey        string
	workspaceRepo    domain.WorkspaceRepository
	templateRepo     domain.TemplateRepository
	templateService  domain.TemplateService
	messageRepo      domain.MessageHistoryRepository
	httpClient       domain.HTTPClient
	webhookEndpoint  string
	apiEndpoint      string
	smtpService      domain.EmailProviderService
	sesService       domain.EmailProviderService
	sparkPostService domain.EmailProviderService
	postmarkService  domain.EmailProviderService
	mailgunService   domain.EmailProviderService
	mailjetService   domain.EmailProviderService
}

// NewEmailService creates a new EmailService instance
func NewEmailService(
	logger logger.Logger,
	authService domain.AuthService,
	secretKey string,
	workspaceRepo domain.WorkspaceRepository,
	templateRepo domain.TemplateRepository,
	templateService domain.TemplateService,
	messageRepo domain.MessageHistoryRepository,
	httpClient domain.HTTPClient,
	webhookEndpoint string,
	apiEndpoint string,
) *EmailService {
	// Initialize provider services
	smtpService := NewSMTPService(logger)
	sesService := NewSESService(authService, logger)
	sparkPostService := NewSparkPostService(httpClient, authService, logger)
	postmarkService := NewPostmarkService(httpClient, authService, logger)
	mailgunService := NewMailgunService(httpClient, authService, logger, webhookEndpoint)
	mailjetService := NewMailjetService(httpClient, authService, logger)

	return &EmailService{
		logger:           logger,
		authService:      authService,
		secretKey:        secretKey,
		workspaceRepo:    workspaceRepo,
		templateRepo:     templateRepo,
		templateService:  templateService,
		messageRepo:      messageRepo,
		httpClient:       httpClient,
		webhookEndpoint:  webhookEndpoint,
		apiEndpoint:      apiEndpoint,
		smtpService:      smtpService,
		sesService:       sesService,
		sparkPostService: sparkPostService,
		postmarkService:  postmarkService,
		mailgunService:   mailgunService,
		mailjetService:   mailjetService,
	}
}

// CreateSESClient creates a new SES client with the provided credentials
func CreateSESClient(region, accessKey, secretKey string) domain.SESClient {
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	return ses.New(sess)
}

// TestEmailProvider sends a test email to verify the provider configuration works
func (s *EmailService) TestEmailProvider(ctx context.Context, workspaceID string, provider domain.EmailProvider, to string) error {
	ctx, span := tracing.StartServiceSpan(ctx, "EmailService", "TestEmailProvider")
	defer tracing.EndSpan(span, nil)

	// Authenticate user
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		tracing.MarkSpanError(ctx, err)
		return err
	}

	// Validate the provider has the required fields
	if len(provider.Senders) == 0 {
		return fmt.Errorf("at least one sender is required for the provider")
	}

	// Use the first sender in the list
	defaultSender := provider.Senders[0]

	// Ensure sender has ID
	if defaultSender.ID == "" {
		defaultSender.ID = uuid.New().String()
		provider.Senders[0] = defaultSender
	}

	// Generate email content
	subject := "Notifuse: Test Email Provider"
	content := "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your provider is working!</p>"

	// Send email with the provider details
	messageID := uuid.New().String()

	err = s.SendEmail(
		ctx,
		workspaceID,
		messageID,
		false, // Use transactional provider type
		defaultSender.Email,
		defaultSender.Name,
		to,
		subject,
		content,
		&provider,
		domain.EmailOptions{
			ReplyTo: "",
			CC:      nil,
			BCC:     nil,
		},
	)

	if err != nil {
		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to test provider: %w", err)
	}

	return nil
}

// SendEmail sends an email using the specified provider
func (s *EmailService) SendEmail(ctx context.Context, workspaceID string, messageID string, isMarketing bool, fromAddress string, fromName string, to string, subject string, content string, provider *domain.EmailProvider, options domain.EmailOptions) error {

	// If fromAddress is not provided, use the first sender's email from the provider
	if fromAddress == "" && len(provider.Senders) > 0 {
		fromAddress = provider.Senders[0].Email
	}

	// If fromName is not provided, use the first sender's name from the provider
	if fromName == "" && len(provider.Senders) > 0 {
		fromName = provider.Senders[0].Name
	}

	// Get the appropriate provider service
	providerService, err := s.getProviderService(provider.Kind)
	if err != nil {
		return err
	}

	// Delegate to the provider-specific implementation
	return providerService.SendEmail(ctx, workspaceID, messageID, fromAddress, fromName, to, subject, content, provider, options)
}

// getProviderService returns the appropriate email provider service based on provider kind
func (s *EmailService) getProviderService(providerKind domain.EmailProviderKind) (domain.EmailProviderService, error) {
	switch providerKind {
	case domain.EmailProviderKindSMTP:
		return s.smtpService, nil
	case domain.EmailProviderKindSES:
		return s.sesService, nil
	case domain.EmailProviderKindSparkPost:
		return s.sparkPostService, nil
	case domain.EmailProviderKindPostmark:
		return s.postmarkService, nil
	case domain.EmailProviderKindMailgun:
		return s.mailgunService, nil
	case domain.EmailProviderKindMailjet:
		return s.mailjetService, nil
	default:
		return nil, fmt.Errorf("unsupported provider kind: %s", providerKind)
	}
}

func (s *EmailService) VisitLink(ctx context.Context, messageID string, workspaceID string) error {
	// find the message by id
	err := s.messageRepo.SetClicked(ctx, workspaceID, messageID, time.Now())
	if err != nil {
		s.logger.Error(err.Error())
		return fmt.Errorf("failed to set clicked: %w", err)
	}

	return nil
}

func (s *EmailService) OpenEmail(ctx context.Context, messageID string, workspaceID string) error {
	// find the message by id
	err := s.messageRepo.SetOpened(ctx, workspaceID, messageID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update message opened: %w", err)
	}
	return nil
}

// SendEmailForTemplate handles sending through the email channel
func (s *EmailService) SendEmailForTemplate(
	ctx context.Context,
	workspaceID string,
	messageID string,
	contact *domain.Contact,
	templateConfig domain.ChannelTemplate,
	messageData domain.MessageData,
	trackingSettings mjml.TrackingSettings,
	emailProvider *domain.EmailProvider,
	options domain.EmailOptions,
) error {
	ctx, span := tracing.StartServiceSpan(ctx, "EmailService", "SendEmailForTemplate")
	defer span.End()

	span.AddAttributes(
		trace.StringAttribute("workspace", workspaceID),
		trace.StringAttribute("message_id", messageID),
		trace.StringAttribute("contact.email", contact.Email),
		trace.StringAttribute("template_id", templateConfig.TemplateID),
	)

	s.logger.WithFields(map[string]interface{}{
		"workspace":   workspaceID,
		"message_id":  messageID,
		"contact":     contact.Email,
		"template_id": templateConfig.TemplateID,
	}).Debug("Preparing to send email notification")

	// Get the template
	template, err := s.templateService.GetTemplateByID(ctx, workspaceID, templateConfig.TemplateID, int64(0))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":       err.Error(),
			"template_id": templateConfig.TemplateID,
		}).Error("Failed to get template")

		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Find the emailSender
	emailSender := emailProvider.GetSender(template.Email.SenderID)

	if emailSender == nil {
		return fmt.Errorf("sender not found: %s", template.Email.SenderID)
	}

	span.AddAttributes(
		trace.StringAttribute("template.subject", template.Email.Subject),
		trace.StringAttribute("template.from_email", emailSender.Email),
	)

	// set utm_content to the template id if not set
	if trackingSettings.UTMContent == "" {
		trackingSettings.UTMContent = template.ID
	}

	compileTemplateRequest := domain.CompileTemplateRequest{
		WorkspaceID:      workspaceID,
		MessageID:        messageID,
		VisualEditorTree: template.Email.VisualEditorTree,
		TemplateData:     messageData.Data,
		TrackingEnabled:  trackingSettings.EnableTracking,
		UTMSource:        &trackingSettings.UTMSource,
		UTMMedium:        &trackingSettings.UTMMedium,
		UTMCampaign:      &trackingSettings.UTMCampaign,
		UTMContent:       &trackingSettings.UTMContent,
		UTMTerm:          &trackingSettings.UTMTerm,
	}

	// Compile the template with the message data
	compiledTemplate, err := s.templateService.CompileTemplate(ctx, compileTemplateRequest)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":       err.Error(),
			"template_id": templateConfig.TemplateID,
		}).Error("Failed to compile template")

		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to compile template: %w", err)
	}

	tracing.AddAttribute(ctx, "template.compilation_success", compiledTemplate.Success)

	if !compiledTemplate.Success || compiledTemplate.HTML == nil {
		errMsg := "Unknown error"
		if compiledTemplate.Error != nil {
			errMsg = compiledTemplate.Error.Message
		}
		s.logger.WithField("error", errMsg).Error("Template compilation failed")

		err := fmt.Errorf("template compilation failed: %s", errMsg)
		tracing.MarkSpanError(ctx, err)
		return err
	}

	// Get necessary email information from the template
	fromEmail := emailSender.Email
	fromName := emailSender.Name
	subject := template.Email.Subject
	htmlContent := *compiledTemplate.HTML
	now := time.Now().UTC()

	// Create message history record
	messageHistory := &domain.MessageHistory{
		ID:           messageID,
		ContactEmail: contact.Email,
		TemplateID:   templateConfig.TemplateID,
		Channel:      "email",
		MessageData:  messageData,
		SentAt:       now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save to message history
	if err := s.messageRepo.Create(ctx, workspaceID, messageHistory); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"message_id": messageID,
		}).Error("Failed to create message history")

		tracing.MarkSpanError(ctx, err)
		return fmt.Errorf("failed to create message history: %w", err)
	}

	tracing.AddAttribute(ctx, "message_history.created", true)

	// Send the email using the email service
	s.logger.WithFields(map[string]interface{}{
		"to":         contact.Email,
		"from":       fromEmail,
		"subject":    subject,
		"message_id": messageID,
	}).Debug("Sending email")

	tracing.AddAttribute(ctx, "email.sending", true)

	// optional override for reply to
	if template.Email.ReplyTo != "" {
		options.ReplyTo = template.Email.ReplyTo
	}

	err = s.SendEmail(
		ctx,
		workspaceID,
		messageID,
		false, // Use transactional provider type
		fromEmail,
		fromName,
		contact.Email, // To address
		subject,
		htmlContent,
		emailProvider,
		options,
	)

	if err != nil {
		// Update message history with error status
		messageHistory.FailedAt = &now
		messageHistory.UpdatedAt = now
		errorMsg := err.Error()
		messageHistory.StatusInfo = &errorMsg

		// Attempt to update the message history record
		updateErr := s.messageRepo.Update(ctx, workspaceID, messageHistory)
		if updateErr != nil {
			s.logger.WithFields(map[string]interface{}{
				"error":      updateErr.Error(),
				"message_id": messageID,
			}).Error("Failed to update message history with error status")

			tracing.AddAttribute(ctx, "message_history.update_error", updateErr.Error())
		}

		s.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"message_id": messageID,
			"to":         contact.Email,
		}).Error("Failed to send email")

		tracing.MarkSpanError(ctx, err)
		tracing.AddAttribute(ctx, "email.error", err.Error())
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"message_id": messageID,
		"to":         contact.Email,
	}).Info("Email sent successfully")

	tracing.AddAttribute(ctx, "email.sent", true)
	return nil
}
