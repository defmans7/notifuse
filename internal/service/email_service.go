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

// TestEmailProvider validates and tests an email provider
func (s *EmailService) TestEmailProvider(ctx context.Context, workspaceID string, provider domain.EmailProvider, to string) error {
	// Authenticate user for the workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user for workspace: %w", err)
	}

	// Validate provider config using the service's secret key
	if err := provider.Validate(s.secretKey); err != nil {
		return err
	}

	// Prepare email content
	subject := "Notifuse: Test Email Provider"
	htmlContent := "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your provider is working!</p>"

	// Send email using SendEmail method with the direct provider
	return s.SendEmail(ctx, workspaceID, false, provider.DefaultSenderEmail, provider.DefaultSenderName, to, subject, htmlContent, &provider, "", nil, nil)
}

// TestTemplate tests a template by sending a test email
func (s *EmailService) TestTemplate(ctx context.Context, workspaceID string, templateID string, integrationID string, recipientEmail string, cc []string, bcc []string, replyTo string) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Get the workspace to retrieve email provider settings
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get the template by ID - use latest version (pass 0 for version)
	template, err := s.templateRepo.GetTemplateByID(ctx, workspaceID, templateID, 0)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Get the integrationFound by ID
	var integrationFound *domain.Integration
	for _, integration := range workspace.Integrations {
		if integration.ID == integrationID {
			integrationFound = &integration
			break
		}
	}

	if integrationFound == nil {
		return fmt.Errorf("integration not found: %s", integrationID)
	}

	// Validate that the provider is configured
	if integrationFound.EmailProvider.Kind == "" {
		return fmt.Errorf("no email provider configured for type: %s", integrationID)
	}

	// Use test data from the template if available, otherwise use a default test data object
	var testData map[string]interface{}
	if template.TestData != nil && len(template.TestData) > 0 {
		testData = template.TestData
	} else {
		// Create a simple test data object with dummy values
		testData = map[string]interface{}{
			"name":    "Test User",
			"company": "Notifuse",
			"url":     "https://example.com/test",
		}
	}

	// Compile the template with test data
	var emailContent string
	var emailSubject string

	if template.Email != nil {
		if template.Email.Subject != "" {
			emailSubject = template.Email.Subject
		} else {
			emailSubject = "Notifuse: Test Template Email"
		}

		// Use templateService to compile the template with the tree
		compileResult, err := s.templateService.CompileTemplate(ctx, domain.CompileTemplateRequest{
			WorkspaceID:      workspaceID,
			VisualEditorTree: template.Email.VisualEditorTree,
			TemplateData:     testData,
		})
		if err != nil {
			return fmt.Errorf("failed to compile template: %w", err)
		}

		if !compileResult.Success || compileResult.HTML == nil {
			errorMsg := "Unknown error"
			if compileResult.Error != nil {
				errorMsg = compileResult.Error.Message
			}
			return fmt.Errorf("template compilation failed: %s", errorMsg)
		}

		emailContent = *compileResult.HTML

	} else {
		emailSubject = "Notifuse: Test Template Email"
		emailContent = "<h1>Notifuse: Test Template Email</h1><p>This is a test email from template " + template.Name + ".</p>"
	}

	// Get reply-to from the request or the template if available
	templateReplyTo := ""
	if replyTo != "" {
		templateReplyTo = replyTo
	} else if template.Email != nil && template.Email.ReplyTo != "" {
		templateReplyTo = template.Email.ReplyTo
	}

	// Send the email using SendEmail method with the provider from the integration
	return s.SendEmail(ctx, workspaceID, false, integrationFound.EmailProvider.DefaultSenderEmail, integrationFound.EmailProvider.DefaultSenderName, recipientEmail, emailSubject, emailContent, &integrationFound.EmailProvider, templateReplyTo, cc, bcc)
}

// SendEmail sends an email using the specified provider
func (s *EmailService) SendEmail(ctx context.Context, workspaceID string, isMarketing bool, fromAddress string, fromName string, to string, subject string, content string, provider *domain.EmailProvider, replyTo string, cc []string, bcc []string) error {

	// If fromAddress is not provided, use the default sender email from the provider
	if fromAddress == "" {
		fromAddress = provider.DefaultSenderEmail
	}

	// If fromName is not provided, use the default sender name from the provider
	if fromName == "" {
		fromName = provider.DefaultSenderName
	}

	// Get the appropriate provider service
	providerService, err := s.getProviderService(provider.Kind)
	if err != nil {
		return err
	}

	// Delegate to the provider-specific implementation
	return providerService.SendEmail(ctx, workspaceID, fromAddress, fromName, to, subject, content, provider, replyTo, cc, bcc)
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
	cc []string,
	bcc []string,
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

	span.AddAttributes(
		trace.StringAttribute("template.subject", template.Email.Subject),
		trace.StringAttribute("template.from_email", template.Email.FromAddress),
	)

	// override tracking settings with template settings if not set
	if trackingSettings.UTMSource == "" && template.UTMSource != nil {
		trackingSettings.UTMSource = *template.UTMSource
	}
	if trackingSettings.UTMMedium == "" && template.UTMMedium != nil {
		trackingSettings.UTMMedium = *template.UTMMedium
	}
	if trackingSettings.UTMCampaign == "" && template.UTMCampaign != nil {
		trackingSettings.UTMCampaign = *template.UTMCampaign
	}
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
	fromEmail := template.Email.FromAddress
	fromName := template.Email.FromName
	subject := template.Email.Subject
	htmlContent := *compiledTemplate.HTML

	// Create message history record
	messageHistory := &domain.MessageHistory{
		ID:           messageID,
		ContactEmail: contact.Email,
		TemplateID:   templateConfig.TemplateID,
		Channel:      "email",
		Status:       domain.MessageStatusSent,
		MessageData:  messageData,
		SentAt:       time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
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

	err = s.SendEmail(
		ctx,
		workspaceID,
		false, // Use transactional provider type
		fromEmail,
		fromName,
		contact.Email, // To address
		subject,
		htmlContent,
		emailProvider,
		template.Email.ReplyTo,
		cc,
		bcc,
	)

	if err != nil {
		// Update message history with error status
		messageHistory.Status = domain.MessageStatusFailed
		messageHistory.UpdatedAt = time.Now().UTC()
		errorMsg := err.Error()
		messageHistory.Error = &errorMsg

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
