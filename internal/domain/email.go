package domain

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Notifuse/notifuse/pkg/crypto"
	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/service/ses"
)

//go:generate mockgen -destination mocks/mock_email_service.go -package mocks github.com/Notifuse/notifuse/internal/domain EmailServiceInterface
//go:generate mockgen -destination mocks/mock_http_client.go -package mocks github.com/Notifuse/notifuse/internal/domain HTTPClient
//go:generate mockgen -destination mocks/mock_ses_client.go -package mocks github.com/Notifuse/notifuse/internal/domain SESClient

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SESClient defines the interface for AWS SES operations
type SESClient interface {
	SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error)
}

// AmazonSES contains SES email provider settings
type AmazonSES struct {
	Region             string `json:"region"`
	AccessKey          string `json:"access_key"`
	EncryptedSecretKey string `json:"encrypted_secret_key,omitempty"`
	SandboxMode        bool   `json:"sandbox_mode"`

	// decoded secret key, not stored in the database
	SecretKey string `json:"secret_key,omitempty"`
}

func (a *AmazonSES) DecryptSecretKey(passphrase string) error {
	secretKey, err := crypto.DecryptFromHexString(a.EncryptedSecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt SES secret key: %w", err)
	}
	a.SecretKey = secretKey
	return nil
}

func (a *AmazonSES) EncryptSecretKey(passphrase string) error {
	encryptedSecretKey, err := crypto.EncryptString(a.SecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt SES secret key: %w", err)
	}
	a.EncryptedSecretKey = encryptedSecretKey
	return nil
}

func (a *AmazonSES) Validate(passphrase string) error {
	// Check if any field is set to determine if we should validate
	isConfigured := a.Region != "" || a.AccessKey != "" ||
		a.EncryptedSecretKey != "" || a.SecretKey != ""

	// If no fields are set, consider it valid (optional config)
	if !isConfigured {
		return nil
	}

	// If any field is set, validate required fields are present
	if a.Region == "" {
		return fmt.Errorf("region is required when Amazon SES is configured")
	}

	if a.AccessKey == "" {
		return fmt.Errorf("access key is required when Amazon SES is configured")
	}

	// only encrypt secret key if it's not empty
	if a.SecretKey != "" {
		if err := a.EncryptSecretKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt SES secret key: %w", err)
		}
	}

	return nil
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

// EmailProvider contains configuration for an email service provider
type EmailProvider struct {
	Kind               EmailProviderKind  `json:"kind"`
	SES                *AmazonSES         `json:"ses,omitempty"`
	SMTP               *SMTPSettings      `json:"smtp,omitempty"`
	SparkPost          *SparkPostSettings `json:"sparkpost,omitempty"`
	Postmark           *PostmarkSettings  `json:"postmark,omitempty"`
	Mailgun            *MailgunSettings   `json:"mailgun,omitempty"`
	Mailjet            *MailjetSettings   `json:"mailjet,omitempty"`
	DefaultSenderEmail string             `json:"default_sender_email"`
	DefaultSenderName  string             `json:"default_sender_name"`
}

// Validate validates the email provider settings
func (e *EmailProvider) Validate(passphrase string) error {
	// If Kind is empty, consider it as not configured
	if e.Kind == "" {
		return nil
	}

	// Validate default sender fields
	if e.DefaultSenderEmail == "" {
		return fmt.Errorf("default sender email is required")
	}
	if !govalidator.IsEmail(e.DefaultSenderEmail) {
		return fmt.Errorf("invalid default sender email: %s", e.DefaultSenderEmail)
	}
	if e.DefaultSenderName == "" {
		return fmt.Errorf("default sender name is required")
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

// SMTPSettings contains configuration for SMTP email server
type SMTPSettings struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Username          string `json:"username"`
	EncryptedPassword string `json:"encrypted_password,omitempty"`
	UseTLS            bool   `json:"use_tls"`

	// decoded password, not stored in the database
	Password string `json:"password,omitempty"`
}

func (s *SMTPSettings) DecryptPassword(passphrase string) error {
	password, err := crypto.DecryptFromHexString(s.EncryptedPassword, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt SMTP password: %w", err)
	}
	s.Password = password
	return nil
}

func (s *SMTPSettings) EncryptPassword(passphrase string) error {
	encryptedPassword, err := crypto.EncryptString(s.Password, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt SMTP password: %w", err)
	}
	s.EncryptedPassword = encryptedPassword
	return nil
}

func (s *SMTPSettings) Validate(passphrase string) error {
	if s.Host == "" {
		return fmt.Errorf("host is required for SMTP configuration")
	}

	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("invalid port number for SMTP configuration: %d", s.Port)
	}

	if s.Username == "" {
		return fmt.Errorf("username is required for SMTP configuration")
	}

	// Only encrypt password if it's not empty
	if s.Password != "" {
		if err := s.EncryptPassword(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt SMTP password: %w", err)
		}
	}

	return nil
}

// SparkPostSettings contains configuration for SparkPost email provider
type SparkPostSettings struct {
	EncryptedAPIKey string `json:"encrypted_api_key,omitempty"`
	SandboxMode     bool   `json:"sandbox_mode"`
	Endpoint        string `json:"endpoint"`

	// decoded API key, not stored in the database
	APIKey string `json:"api_key,omitempty"`
}

func (s *SparkPostSettings) DecryptAPIKey(passphrase string) error {
	apiKey, err := crypto.DecryptFromHexString(s.EncryptedAPIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt SparkPost API key: %w", err)
	}
	s.APIKey = apiKey
	return nil
}

func (s *SparkPostSettings) EncryptAPIKey(passphrase string) error {
	encryptedAPIKey, err := crypto.EncryptString(s.APIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt SparkPost API key: %w", err)
	}
	s.EncryptedAPIKey = encryptedAPIKey
	return nil
}

func (s *SparkPostSettings) Validate(passphrase string) error {
	if s.Endpoint == "" {
		return fmt.Errorf("endpoint is required for SparkPost configuration")
	}

	// Encrypt API key if it's not empty
	if s.APIKey != "" {
		if err := s.EncryptAPIKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt SparkPost API key: %w", err)
		}
	}

	return nil
}

// PostmarkSettings contains configuration for Postmark email provider
type PostmarkSettings struct {
	EncryptedServerToken string `json:"encrypted_server_token,omitempty"`
	ServerToken          string `json:"server_token,omitempty"`
}

func (p *PostmarkSettings) DecryptServerToken(passphrase string) error {
	serverToken, err := crypto.DecryptFromHexString(p.EncryptedServerToken, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt Postmark server token: %w", err)
	}
	p.ServerToken = serverToken
	return nil
}

func (p *PostmarkSettings) EncryptServerToken(passphrase string) error {
	encryptedServerToken, err := crypto.EncryptString(p.ServerToken, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt Postmark server token: %w", err)
	}
	p.EncryptedServerToken = encryptedServerToken
	return nil
}

func (p *PostmarkSettings) Validate(passphrase string) error {
	// Encrypt server token if it's not empty
	if p.ServerToken != "" {
		if err := p.EncryptServerToken(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt Postmark server token: %w", err)
		}
	}

	return nil
}

// MailgunSettings contains configuration for Mailgun
type MailgunSettings struct {
	EncryptedAPIKey string `json:"encrypted_api_key,omitempty"`
	Domain          string `json:"domain"`
	Region          string `json:"region,omitempty"` // "US" or "EU"

	// decoded API key, not stored in the database
	APIKey string `json:"api_key,omitempty"`
}

func (m *MailgunSettings) DecryptAPIKey(passphrase string) error {
	apiKey, err := crypto.DecryptFromHexString(m.EncryptedAPIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt Mailgun API key: %w", err)
	}
	m.APIKey = apiKey
	return nil
}

func (m *MailgunSettings) EncryptAPIKey(passphrase string) error {
	encryptedAPIKey, err := crypto.EncryptString(m.APIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt Mailgun API key: %w", err)
	}
	m.EncryptedAPIKey = encryptedAPIKey
	return nil
}

func (m *MailgunSettings) Validate(passphrase string) error {
	if m.Domain == "" {
		return fmt.Errorf("domain is required for Mailgun configuration")
	}

	// Encrypt API key if it's not empty
	if m.APIKey != "" {
		if err := m.EncryptAPIKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt Mailgun API key: %w", err)
		}
		m.APIKey = "" // Clear the API key after encryption
	}

	return nil
}

// MailjetSettings contains configuration for Mailjet
type MailjetSettings struct {
	EncryptedAPIKey    string `json:"encrypted_api_key,omitempty"`
	EncryptedSecretKey string `json:"encrypted_secret_key,omitempty"`
	SandboxMode        bool   `json:"sandbox_mode"`

	// decoded keys, not stored in the database
	APIKey    string `json:"api_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

func (m *MailjetSettings) DecryptAPIKey(passphrase string) error {
	apiKey, err := crypto.DecryptFromHexString(m.EncryptedAPIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt Mailjet API key: %w", err)
	}
	m.APIKey = apiKey
	return nil
}

func (m *MailjetSettings) EncryptAPIKey(passphrase string) error {
	encryptedAPIKey, err := crypto.EncryptString(m.APIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt Mailjet API key: %w", err)
	}
	m.EncryptedAPIKey = encryptedAPIKey
	return nil
}

func (m *MailjetSettings) DecryptSecretKey(passphrase string) error {
	secretKey, err := crypto.DecryptFromHexString(m.EncryptedSecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt Mailjet Secret key: %w", err)
	}
	m.SecretKey = secretKey
	return nil
}

func (m *MailjetSettings) EncryptSecretKey(passphrase string) error {
	encryptedSecretKey, err := crypto.EncryptString(m.SecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt Mailjet Secret key: %w", err)
	}
	m.EncryptedSecretKey = encryptedSecretKey
	return nil
}

func (m *MailjetSettings) Validate(passphrase string) error {
	// API Key is required for Mailjet
	if m.APIKey != "" {
		if err := m.EncryptAPIKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt Mailjet API key: %w", err)
		}
		m.APIKey = "" // Clear the API key after encryption
	}

	// Secret Key is required for Mailjet
	if m.SecretKey != "" {
		if err := m.EncryptSecretKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt Mailjet Secret key: %w", err)
		}
		m.SecretKey = "" // Clear the Secret key after encryption
	}

	return nil
}

// EmailServiceInterface defines the interface for the email service
type EmailServiceInterface interface {
	TestEmailProvider(ctx context.Context, workspaceID string, provider EmailProvider, to string) error
	TestTemplate(ctx context.Context, workspaceID string, templateID string, providerType string, recipientEmail string) error
	SendEmail(ctx context.Context, workspaceID string, providerType string, fromAddress string, fromName string, to string, subject string, content string, optionalProvider ...*EmailProvider) error
}
