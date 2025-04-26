package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/wneessen/go-mail"
)

type EmailService struct {
	logger          logger.Logger
	authService     domain.AuthService
	secretKey       string
	workspaceRepo   domain.WorkspaceRepository
	templateRepo    domain.TemplateRepository
	templateService domain.TemplateService
	httpClient      domain.HTTPClient
}

// NewEmailService creates a new EmailService instance
func NewEmailService(
	logger logger.Logger,
	authService domain.AuthService,
	secretKey string,
	workspaceRepo domain.WorkspaceRepository,
	templateRepo domain.TemplateRepository,
	templateService domain.TemplateService,
) *EmailService {
	return &EmailService{
		logger:          logger,
		authService:     authService,
		secretKey:       secretKey,
		workspaceRepo:   workspaceRepo,
		templateRepo:    templateRepo,
		templateService: templateService,
		httpClient:      &http.Client{Timeout: time.Second * 10},
	}
}

// SetHTTPClient sets a custom HTTP client (useful for testing)
func (s *EmailService) SetHTTPClient(client domain.HTTPClient) {
	s.httpClient = client
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
	return s.SendEmail(ctx, workspaceID, "", provider.DefaultSenderEmail, provider.DefaultSenderName, to, subject, htmlContent, &provider)
}

// TestTemplate tests a template by sending a test email
func (s *EmailService) TestTemplate(ctx context.Context, workspaceID string, templateID string, providerType string, recipientEmail string) error {
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

	// Determine which email provider to use based on providerType
	var emailProvider domain.EmailProvider
	if providerType == "marketing" {
		emailProvider = workspace.Settings.EmailMarketingProvider
	} else if providerType == "transactional" {
		emailProvider = workspace.Settings.EmailTransactionalProvider
	} else {
		return fmt.Errorf("invalid provider type: %s", providerType)
	}

	// Validate that the provider is configured
	if emailProvider.Kind == "" {
		return fmt.Errorf("no email provider configured for type: %s", providerType)
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
		compileResult, err := s.templateService.CompileTemplate(ctx, workspaceID, template.Email.VisualEditorTree, testData)
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

	// Send the email using SendEmail method - we pass empty string for providerType since we're providing the provider directly
	return s.SendEmail(ctx, workspaceID, "", emailProvider.DefaultSenderEmail, emailProvider.DefaultSenderName, recipientEmail, emailSubject, emailContent, &emailProvider)
}

// SendEmail sends an email using the specified provider type or a direct provider
func (s *EmailService) SendEmail(ctx context.Context, workspaceID string, providerType string, fromAddress string, fromName string, to string, subject string, content string, optionalProvider ...*domain.EmailProvider) error {
	// Direct provider takes precedence if provided
	var emailProvider domain.EmailProvider
	if len(optionalProvider) > 0 && optionalProvider[0] != nil {
		emailProvider = *optionalProvider[0]
	} else {
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

		// Determine which email provider to use based on providerType
		if providerType == "marketing" {
			emailProvider = workspace.Settings.EmailMarketingProvider
		} else if providerType == "transactional" {
			emailProvider = workspace.Settings.EmailTransactionalProvider
		} else {
			return fmt.Errorf("invalid provider type: %s", providerType)
		}

		// Validate that the provider is configured
		if emailProvider.Kind == "" {
			return fmt.Errorf("no email provider configured for type: %s", providerType)
		}
	}

	// If fromAddress is not provided, use the default sender email from the provider
	if fromAddress == "" {
		fromAddress = emailProvider.DefaultSenderEmail
	}

	// If fromName is not provided, use the default sender name from the provider
	if fromName == "" {
		fromName = emailProvider.DefaultSenderName
	}

	// Send the email using the appropriate provider
	switch emailProvider.Kind {
	case domain.EmailProviderKindSMTP:
		if emailProvider.SMTP == nil {
			return fmt.Errorf("SMTP settings required")
		}
		// Decrypt password if needed
		if emailProvider.SMTP.EncryptedPassword != "" && emailProvider.SMTP.Password == "" {
			if err := emailProvider.SMTP.DecryptPassword(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt SMTP password: %w", err)
			}
		}

		// Use go-mail to send the email
		msg := mail.NewMsg()
		if err := msg.FromFormat(fromAddress, fromName); err != nil {
			return fmt.Errorf("invalid sender: %w", err)
		}
		if err := msg.To(to); err != nil {
			return fmt.Errorf("invalid recipient email: %w", err)
		}
		msg.Subject(subject)
		msg.SetBodyString(mail.TypeTextHTML, content)

		client, err := mail.NewClient(
			emailProvider.SMTP.Host,
			mail.WithPort(emailProvider.SMTP.Port),
			mail.WithUsername(emailProvider.SMTP.Username),
			mail.WithPassword(emailProvider.SMTP.Password),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithTLSPolicy(mail.TLSMandatory),
			mail.WithTimeout(10*time.Second),
		)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		if err := client.DialAndSend(msg); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

	case domain.EmailProviderKindSES:
		if emailProvider.SES == nil {
			return fmt.Errorf("SES provider is not configured")
		}

		// Decrypt secret key if needed
		if emailProvider.SES.EncryptedSecretKey != "" && emailProvider.SES.SecretKey == "" {
			if err := emailProvider.SES.DecryptSecretKey(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt SES secret key: %w", err)
			}
		}

		// Create a fresh SES client for this request
		svc := CreateSESClient(emailProvider.SES.Region, emailProvider.SES.AccessKey, emailProvider.SES.SecretKey)

		// Format the "From" header with name and email
		fromHeader := fmt.Sprintf("%s <%s>", fromName, fromAddress)

		// Create the email
		input := &ses.SendEmailInput{
			Destination: &ses.Destination{
				ToAddresses: []*string{aws.String(to)},
			},
			Message: &ses.Message{
				Body: &ses.Body{
					Html: &ses.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(content),
					},
				},
				Subject: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(subject),
				},
			},
			Source: aws.String(fromHeader),
		}

		// Send the email
		_, err := svc.SendEmail(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				return fmt.Errorf("SES error: %s", aerr.Error())
			}
			return fmt.Errorf("failed to send email: %w", err)
		}

	case domain.EmailProviderKindSparkPost:
		if emailProvider.SparkPost == nil {
			return fmt.Errorf("SparkPost provider is not configured")
		}

		// Decrypt API key if needed
		if emailProvider.SparkPost.EncryptedAPIKey != "" && emailProvider.SparkPost.APIKey == "" {
			if err := emailProvider.SparkPost.DecryptAPIKey(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt SparkPost API key: %w", err)
			}
		}

		// Create the HTTP request to SparkPost API
		apiURL := emailProvider.SparkPost.Endpoint + "/api/v1/transmissions"

		// Create the request payload
		payload := map[string]interface{}{
			"options": map[string]interface{}{
				"sandbox": emailProvider.SparkPost.SandboxMode,
			},
			"content": map[string]interface{}{
				"from": map[string]string{
					"email": fromAddress,
					"name":  fromName,
				},
				"subject": subject,
				"html":    content,
			},
			"recipients": []map[string]string{
				{
					"address": to,
				},
			},
		}

		// Convert payload to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal SparkPost request: %w", err)
		}

		// Create the HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", emailProvider.SparkPost.APIKey)

		// Use the injected HTTP client
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request to SparkPost API: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read SparkPost API response: %w", err)
		}

		// Check response status
		if resp.StatusCode >= 400 {
			return fmt.Errorf("SparkPost API error (%d): %s", resp.StatusCode, string(body))
		}

	case domain.EmailProviderKindPostmark:
		if emailProvider.Postmark == nil {
			return fmt.Errorf("Postmark provider is not configured")
		}

		// Decrypt server token if needed
		if emailProvider.Postmark.EncryptedServerToken != "" && emailProvider.Postmark.ServerToken == "" {
			if err := emailProvider.Postmark.DecryptServerToken(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt Postmark server token: %w", err)
			}
		}

		// Create a simple HTTP request to the Postmark API
		apiURL := "https://api.postmarkapp.com/email"

		// Format the "From" header with name and email
		fromHeader := fmt.Sprintf("%s <%s>", fromName, fromAddress)

		// Create the request payload
		payload := map[string]interface{}{
			"From":          fromHeader,
			"To":            to,
			"Subject":       subject,
			"HtmlBody":      content,
			"MessageStream": "outbound",
		}

		// Convert payload to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal Postmark request: %w", err)
		}

		// Create the HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Postmark-Server-Token", emailProvider.Postmark.ServerToken)

		// Use the injected HTTP client
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request to Postmark API: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read Postmark API response: %w", err)
		}

		// Check response status
		if resp.StatusCode >= 400 {
			return fmt.Errorf("Postmark API error (%d): %s", resp.StatusCode, string(body))
		}

	default:
		return fmt.Errorf("unsupported provider kind: %s", emailProvider.Kind)
	}

	return nil
}
