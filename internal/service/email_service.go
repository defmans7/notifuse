package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	httpClient domain.HTTPClient,
) *EmailService {
	return &EmailService{
		logger:          logger,
		authService:     authService,
		secretKey:       secretKey,
		workspaceRepo:   workspaceRepo,
		templateRepo:    templateRepo,
		templateService: templateService,
		httpClient:      httpClient,
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
	return s.SendEmail(ctx, workspaceID, false, provider.DefaultSenderEmail, provider.DefaultSenderName, to, subject, htmlContent, &provider)
}

// TestTemplate tests a template by sending a test email
func (s *EmailService) TestTemplate(ctx context.Context, workspaceID string, templateID string, integrationID string, recipientEmail string) error {
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
	return s.SendEmail(ctx, workspaceID, false, integrationFound.EmailProvider.DefaultSenderEmail, integrationFound.EmailProvider.DefaultSenderName, recipientEmail, emailSubject, emailContent, &integrationFound.EmailProvider)
}

// SendEmail sends an email using the specified provider type or a direct provider
func (s *EmailService) SendEmail(ctx context.Context, workspaceID string, isMarketing bool, fromAddress string, fromName string, to string, subject string, content string, optionalProvider ...*domain.EmailProvider) error {
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

		// Get the email provider using the workspace's GetEmailProvider method
		provider, err := workspace.GetEmailProvider(isMarketing)
		if err != nil {
			return err
		}

		// Validate that the provider is configured
		if provider == nil || provider.Kind == "" {
			return fmt.Errorf("no email provider configured for type: %t", isMarketing)
		}

		emailProvider = *provider
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
		_, err := svc.SendEmailWithContext(ctx, input)
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

		// Prepare the API endpoint
		endpoint := "https://api.postmarkapp.com/email"

		// Prepare the request body
		requestBody := map[string]interface{}{
			"From":     fmt.Sprintf("%s <%s>", fromName, fromAddress),
			"To":       to,
			"Subject":  subject,
			"HtmlBody": content,
		}

		// Convert to JSON
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal Postmark request: %w", err)
		}

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create Postmark request: %w", err)
		}

		// Add headers
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

	case domain.EmailProviderKindMailgun:
		if emailProvider.Mailgun == nil {
			return fmt.Errorf("Mailgun provider is not configured")
		}

		// Decrypt API key if needed
		if emailProvider.Mailgun.EncryptedAPIKey != "" && emailProvider.Mailgun.APIKey == "" {
			if err := emailProvider.Mailgun.DecryptAPIKey(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt Mailgun API key: %w", err)
			}
		}

		// Determine API region
		baseURL := "https://api.mailgun.net/v3"
		if emailProvider.Mailgun.Region == "EU" {
			baseURL = "https://api.eu.mailgun.net/v3"
		}

		// Create the API URL with the domain
		apiURL := fmt.Sprintf("%s/%s/messages", baseURL, emailProvider.Mailgun.Domain)

		// Create form data
		formData := url.Values{}
		formData.Set("from", fmt.Sprintf("%s <%s>", fromName, fromAddress))
		formData.Set("to", to)
		formData.Set("subject", subject)
		formData.Set("html", content)

		// Create the HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("api", emailProvider.Mailgun.APIKey)

		// Use the injected HTTP client
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request to Mailgun API: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read Mailgun API response: %w", err)
		}

		// Check response status
		if resp.StatusCode >= 400 {
			return fmt.Errorf("Mailgun API error (%d): %s", resp.StatusCode, string(body))
		}

	case domain.EmailProviderKindMailjet:
		if emailProvider.Mailjet == nil {
			return fmt.Errorf("Mailjet provider is not configured")
		}

		// Decrypt API key and Secret Key if needed
		if emailProvider.Mailjet.EncryptedAPIKey != "" && emailProvider.Mailjet.APIKey == "" {
			if err := emailProvider.Mailjet.DecryptAPIKey(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt Mailjet API key: %w", err)
			}
		}

		if emailProvider.Mailjet.EncryptedSecretKey != "" && emailProvider.Mailjet.SecretKey == "" {
			if err := emailProvider.Mailjet.DecryptSecretKey(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt Mailjet Secret key: %w", err)
			}
		}

		// Create the HTTP request to Mailjet API
		apiURL := "https://api.mailjet.com/v3.1/send"

		// Create the request payload
		payload := map[string]interface{}{
			"SandboxMode": emailProvider.Mailjet.SandboxMode,
			"Messages": []map[string]interface{}{
				{
					"From": map[string]string{
						"Email": fromAddress,
						"Name":  fromName,
					},
					"To": []map[string]string{
						{
							"Email": to,
						},
					},
					"Subject":  subject,
					"HTMLPart": content,
				},
			},
		}

		// Convert payload to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal Mailjet request: %w", err)
		}

		// Create the HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		// Use basic auth with API key as username and Secret key as password
		req.SetBasicAuth(emailProvider.Mailjet.APIKey, emailProvider.Mailjet.SecretKey)

		// Use the injected HTTP client
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request to Mailjet API: %w", err)
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read Mailjet API response: %w", err)
		}

		// Check response status
		if resp.StatusCode >= 400 {
			return fmt.Errorf("Mailjet API error (%d): %s", resp.StatusCode, string(body))
		}

	default:
		return fmt.Errorf("unsupported provider kind: %s", emailProvider.Kind)
	}

	return nil
}
