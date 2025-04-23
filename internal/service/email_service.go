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
	logger      logger.Logger
	authService domain.AuthService
	secretKey   string
}

func NewEmailService(
	logger logger.Logger,
	authService domain.AuthService,
	secretKey string,
) *EmailService {
	return &EmailService{
		logger:      logger,
		authService: authService,
		secretKey:   secretKey,
	}
}

// TestEmailProvider validates and tests an email provider
func (s *EmailService) TestEmailProvider(ctx context.Context, workspaceID string, provider domain.EmailProvider, to string) error {
	// Authenticate user for the workspace
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user for workspace: %w", err)
	}

	// Validate provider config using the service's secret key
	if err := provider.Validate(s.secretKey); err != nil {
		return err
	}

	switch provider.Kind {
	case domain.EmailProviderKindSMTP:
		if provider.SMTP == nil {
			return fmt.Errorf("SMTP settings required")
		}
		// Decrypt password if needed
		if provider.SMTP.EncryptedPassword != "" && provider.SMTP.Password == "" {
			if err := provider.SMTP.DecryptPassword(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt SMTP password: %w", err)
			}
		}
		// Use go-mail to send a test email
		msg := mail.NewMsg()
		if err := msg.From(provider.DefaultSenderEmail); err != nil {
			return fmt.Errorf("invalid sender email: %w", err)
		}
		if err := msg.To(to); err != nil {
			return fmt.Errorf("invalid recipient email: %w", err)
		}
		msg.Subject("Notifuse: Test Email Provider")
		msg.SetBodyString(mail.TypeTextPlain, "This is a test email from Notifuse. Your provider is working!")

		client, err := mail.NewClient(
			provider.SMTP.Host,
			mail.WithPort(provider.SMTP.Port),
			mail.WithUsername(provider.SMTP.Username),
			mail.WithPassword(provider.SMTP.Password),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithTLSPolicy(mail.TLSMandatory),
			mail.WithTimeout(10*time.Second),
		)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		if err := client.DialAndSend(msg); err != nil {
			return fmt.Errorf("failed to send test email: %w", err)
		}
		return nil
	case domain.EmailProviderKindSES:
		if provider.SES == nil {
			return fmt.Errorf("SES settings required")
		}

		// Decrypt secret key if needed
		if provider.SES.EncryptedSecretKey != "" && provider.SES.SecretKey == "" {
			if err := provider.SES.DecryptSecretKey(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt SES secret key: %w", err)
			}
		}

		// Create a new AWS session with the SES provider credentials
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(provider.SES.Region),
			Credentials: credentials.NewStaticCredentials(provider.SES.AccessKey, provider.SES.SecretKey, ""),
		})
		if err != nil {
			return fmt.Errorf("failed to create AWS session: %w", err)
		}

		// Create an SES service client
		svc := ses.New(sess)

		// Prepare the email
		input := &ses.SendEmailInput{
			Destination: &ses.Destination{
				ToAddresses: []*string{
					aws.String(to),
				},
			},
			Message: &ses.Message{
				Body: &ses.Body{
					Html: &ses.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String("<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your Amazon SES provider is working!</p>"),
					},
					Text: &ses.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String("This is a test email from Notifuse. Your Amazon SES provider is working!"),
					},
				},
				Subject: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String("Notifuse: Test Email Provider"),
				},
			},
			Source: aws.String(provider.DefaultSenderEmail),
		}

		// Send the email
		_, err = svc.SendEmail(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case ses.ErrCodeMessageRejected:
					return fmt.Errorf("message rejected: %s", aerr.Error())
				case ses.ErrCodeMailFromDomainNotVerifiedException:
					return fmt.Errorf("mail from domain not verified: %s", aerr.Error())
				case ses.ErrCodeConfigurationSetDoesNotExistException:
					return fmt.Errorf("configuration set does not exist: %s", aerr.Error())
				default:
					return fmt.Errorf("SES error: %s", aerr.Error())
				}
			}
			return fmt.Errorf("failed to send test email: %w", err)
		}

		return nil
	case domain.EmailProviderKindSparkPost:
		if provider.SparkPost == nil {
			return fmt.Errorf("SparkPost settings required")
		}

		// Decrypt API key if needed
		if provider.SparkPost.EncryptedAPIKey != "" && provider.SparkPost.APIKey == "" {
			if err := provider.SparkPost.DecryptAPIKey(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt SparkPost API key: %w", err)
			}
		}

		// Since we can't directly include the SparkPost API client as a dependency,
		// we'll use a simple HTTP request to the SparkPost API
		apiURL := provider.SparkPost.Endpoint + "/api/v1/transmissions"

		// Create the request payload
		payload := map[string]interface{}{
			"options": map[string]interface{}{
				"sandbox": provider.SparkPost.SandboxMode,
			},
			"content": map[string]interface{}{
				"from": map[string]string{
					"email": provider.DefaultSenderEmail,
					"name":  provider.DefaultSenderName,
				},
				"subject": "Notifuse: Test Email Provider",
				"text":    "This is a test email from Notifuse. Your SparkPost provider is working!",
				"html":    "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your SparkPost provider is working!</p>",
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
		req.Header.Set("Authorization", provider.SparkPost.APIKey)

		// Send the request
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := client.Do(req)
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

		return nil
	case domain.EmailProviderKindPostmark:
		if provider.Postmark == nil {
			return fmt.Errorf("Postmark settings required")
		}

		// Decrypt server token if needed
		if provider.Postmark.EncryptedServerToken != "" && provider.Postmark.ServerToken == "" {
			if err := provider.Postmark.DecryptServerToken(s.secretKey); err != nil {
				return fmt.Errorf("failed to decrypt Postmark server token: %w", err)
			}
		}

		// Create a simple HTTP request to the Postmark API
		apiURL := "https://api.postmarkapp.com/email"

		// Create the request payload
		payload := map[string]interface{}{
			"From":          provider.DefaultSenderEmail,
			"To":            to,
			"Subject":       "Notifuse: Test Email Provider",
			"TextBody":      "This is a test email from Notifuse. Your Postmark provider is working!",
			"HtmlBody":      "<h1>Notifuse: Test Email Provider</h1><p>This is a test email from Notifuse. Your Postmark provider is working!</p>",
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
		req.Header.Set("X-Postmark-Server-Token", provider.Postmark.ServerToken)

		// Send the request
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := client.Do(req)
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

		return nil
	default:
		return fmt.Errorf("unsupported provider kind: %s", provider.Kind)
	}
}
