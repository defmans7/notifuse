package domain

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Notifuse/notifuse/pkg/mjml"
	"github.com/asaskevich/govalidator"
)

//go:generate mockgen -destination mocks/mock_email_service.go -package mocks github.com/Notifuse/notifuse/internal/domain EmailServiceInterface
//go:generate mockgen -destination mocks/mock_http_client.go -package mocks github.com/Notifuse/notifuse/internal/domain HTTPClient
//go:generate mockgen -destination mocks/mock_ses_client.go -package mocks github.com/Notifuse/notifuse/internal/domain SESClient

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// EmailProviderKind defines the type of email provider
type EmailProviderKind string

const (
	EmailProviderKindSMTP      EmailProviderKind = "smtp"
	EmailProviderKindSES       EmailProviderKind = "ses"
	EmailProviderKindSparkPost EmailProviderKind = "sparkpost"
	EmailProviderKindPostmark  EmailProviderKind = "postmark"
	EmailProviderKindMailgun   EmailProviderKind = "mailgun"
	EmailProviderKindMailjet   EmailProviderKind = "mailjet"
)

// Sender represents an email sender with name and email address
type Sender struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// EmailProvider contains configuration for an email service provider
type EmailProvider struct {
	Kind      EmailProviderKind  `json:"kind"`
	SES       *AmazonSESSettings `json:"ses,omitempty"`
	SMTP      *SMTPSettings      `json:"smtp,omitempty"`
	SparkPost *SparkPostSettings `json:"sparkpost,omitempty"`
	Postmark  *PostmarkSettings  `json:"postmark,omitempty"`
	Mailgun   *MailgunSettings   `json:"mailgun,omitempty"`
	Mailjet   *MailjetSettings   `json:"mailjet,omitempty"`
	Senders   []Sender           `json:"senders"`
}

// Validate validates the email provider settings
func (e *EmailProvider) Validate(passphrase string) error {
	// If Kind is empty, consider it as not configured
	if e.Kind == "" {
		return nil
	}

	// Validate senders
	if len(e.Senders) == 0 {
		return fmt.Errorf("at least one sender is required")
	}

	for i, sender := range e.Senders {
		if sender.Email == "" {
			return fmt.Errorf("sender email is required for sender at index %d", i)
		}
		if !govalidator.IsEmail(sender.Email) {
			return fmt.Errorf("invalid sender email: %s at index %d", sender.Email, i)
		}
		if sender.Name == "" {
			return fmt.Errorf("sender name is required for sender at index %d", i)
		}
	}

	// Validate Kind value
	switch e.Kind {
	case EmailProviderKindSMTP:
		if e.SMTP == nil {
			return fmt.Errorf("SMTP settings required when email provider kind is smtp")
		}
		return e.SMTP.Validate(passphrase)
	case EmailProviderKindSES:
		if e.SES == nil {
			return fmt.Errorf("SES settings required when email provider kind is ses")
		}
		return e.SES.Validate(passphrase)
	case EmailProviderKindSparkPost:
		if e.SparkPost == nil {
			return fmt.Errorf("SparkPost settings required when email provider kind is sparkpost")
		}
		return e.SparkPost.Validate(passphrase)
	case EmailProviderKindPostmark:
		if e.Postmark == nil {
			return fmt.Errorf("Postmark settings required when email provider kind is postmark")
		}
		return e.Postmark.Validate(passphrase)
	case EmailProviderKindMailgun:
		if e.Mailgun == nil {
			return fmt.Errorf("Mailgun settings required when email provider kind is mailgun")
		}
		return e.Mailgun.Validate(passphrase)
	case EmailProviderKindMailjet:
		if e.Mailjet == nil {
			return fmt.Errorf("Mailjet settings required when email provider kind is mailjet")
		}
		return e.Mailjet.Validate(passphrase)
	default:
		return fmt.Errorf("invalid email provider kind: %s", e.Kind)
	}
}

// EncryptSecretKeys encrypts all secret keys in the email provider
func (e *EmailProvider) EncryptSecretKeys(passphrase string) error {
	if e.Kind == EmailProviderKindSES && e.SES != nil && e.SES.SecretKey != "" {
		if err := e.SES.EncryptSecretKey(passphrase); err != nil {
			return err
		}
		e.SES.SecretKey = ""
	}

	if e.Kind == EmailProviderKindSMTP && e.SMTP != nil && e.SMTP.Password != "" {
		if err := e.SMTP.EncryptPassword(passphrase); err != nil {
			return err
		}
		e.SMTP.Password = ""
	}

	if e.Kind == EmailProviderKindSparkPost && e.SparkPost != nil && e.SparkPost.APIKey != "" {
		if err := e.SparkPost.EncryptAPIKey(passphrase); err != nil {
			return err
		}
		e.SparkPost.APIKey = ""
	}

	if e.Kind == EmailProviderKindPostmark && e.Postmark != nil && e.Postmark.ServerToken != "" {
		if err := e.Postmark.EncryptServerToken(passphrase); err != nil {
			return err
		}
		e.Postmark.ServerToken = ""
	}

	if e.Kind == EmailProviderKindMailgun && e.Mailgun != nil && e.Mailgun.APIKey != "" {
		if err := e.Mailgun.EncryptAPIKey(passphrase); err != nil {
			return err
		}
		e.Mailgun.APIKey = ""
	}

	if e.Kind == EmailProviderKindMailjet && e.Mailjet != nil {
		if e.Mailjet.APIKey != "" {
			if err := e.Mailjet.EncryptAPIKey(passphrase); err != nil {
				return err
			}
			e.Mailjet.APIKey = ""
		}

		if e.Mailjet.SecretKey != "" {
			if err := e.Mailjet.EncryptSecretKey(passphrase); err != nil {
				return err
			}
			e.Mailjet.SecretKey = ""
		}
	}

	return nil
}

// DecryptSecretKeys decrypts all encrypted secret keys in the email provider
func (e *EmailProvider) DecryptSecretKeys(passphrase string) error {
	if e.Kind == EmailProviderKindSES && e.SES != nil && e.SES.EncryptedSecretKey != "" {
		if err := e.SES.DecryptSecretKey(passphrase); err != nil {
			return err
		}
	}

	if e.Kind == EmailProviderKindSMTP && e.SMTP != nil && e.SMTP.EncryptedPassword != "" {
		if err := e.SMTP.DecryptPassword(passphrase); err != nil {
			return err
		}
	}

	if e.Kind == EmailProviderKindSparkPost && e.SparkPost != nil && e.SparkPost.EncryptedAPIKey != "" {
		if err := e.SparkPost.DecryptAPIKey(passphrase); err != nil {
			return err
		}
	}

	if e.Kind == EmailProviderKindPostmark && e.Postmark != nil && e.Postmark.EncryptedServerToken != "" {
		if err := e.Postmark.DecryptServerToken(passphrase); err != nil {
			return err
		}
	}

	if e.Kind == EmailProviderKindMailgun && e.Mailgun != nil && e.Mailgun.EncryptedAPIKey != "" {
		if err := e.Mailgun.DecryptAPIKey(passphrase); err != nil {
			return err
		}
	}

	if e.Kind == EmailProviderKindMailjet && e.Mailjet != nil {
		if e.Mailjet.EncryptedAPIKey != "" {
			if err := e.Mailjet.DecryptAPIKey(passphrase); err != nil {
				return err
			}
		}

		if e.Mailjet.EncryptedSecretKey != "" {
			if err := e.Mailjet.DecryptSecretKey(passphrase); err != nil {
				return err
			}
		}
	}

	return nil
}

// EmailServiceInterface defines the interface for the email service
type EmailServiceInterface interface {
	TestEmailProvider(ctx context.Context, workspaceID string, provider EmailProvider, to string) error
	TestTemplate(ctx context.Context, workspaceID string, templateID string, integrationID string, recipientEmail string, cc []string, bcc []string, replyTo string) error
	SendEmail(ctx context.Context, workspaceID string, messageID string, isMarketing bool, fromAddress string, fromName string, to string, subject string, content string, provider *EmailProvider, replyTo string, cc []string, bcc []string) error
	SendEmailForTemplate(
		ctx context.Context,
		workspaceID string,
		messageID string,
		contact *Contact,
		templateConfig ChannelTemplate,
		messageData MessageData,
		trackingSettings mjml.TrackingSettings,
		emailProvider *EmailProvider,
		cc []string,
		bcc []string,
	) error
	VisitLink(ctx context.Context, messageID string, workspaceID string) error
	OpenEmail(ctx context.Context, messageID string, workspaceID string) error
}

type EmailProviderService interface {
	SendEmail(ctx context.Context, workspaceID string, messageID string, fromAddress string, fromName string, to string, subject string, content string, provider *EmailProvider, replyTo string, cc []string, bcc []string) error
}
