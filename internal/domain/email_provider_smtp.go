package domain

import (
	"context"
	"fmt"

	"github.com/Notifuse/notifuse/pkg/crypto"
)

//go:generate mockgen -destination mocks/mock_smtp_client.go -package mocks github.com/Notifuse/notifuse/internal/domain SMTPClient,SMTPClientFactory,SMTPService

// SMTPWebhookPayload represents an SMTP webhook payload
// SMTP doesn't typically have a built-in webhook system, so this is a generic structure
// that could be used with a third-party SMTP provider that offers webhooks
type SMTPWebhookPayload struct {
	Event          string            `json:"event"`
	Timestamp      string            `json:"timestamp"`
	MessageID      string            `json:"message_id"`
	Recipient      string            `json:"recipient"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Reason         string            `json:"reason,omitempty"`
	Description    string            `json:"description,omitempty"`
	BounceCategory string            `json:"bounce_category,omitempty"`
	DiagnosticCode string            `json:"diagnostic_code,omitempty"`
	ComplaintType  string            `json:"complaint_type,omitempty"`
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

// SMTPClient is an interface for github.com/wneessen/go-mail Client
type SMTPClient interface {
	SetSender(email string, name ...string) error
	SetRecipient(email string, name ...string) error
	SetSubject(subject string) error
	SetBodyString(contentType, content string) error
	DialAndSend() error
	Close() error
}

// SMTPClientFactory is an interface for creating SMTP clients
type SMTPClientFactory interface {
	NewClient(host string, port int, options ...interface{}) (SMTPClient, error)
}

// SMTPService is an interface for SMTP email sending service
type SMTPService interface {
	SendEmail(ctx context.Context, workspaceID string, fromAddress, fromName, to, subject, content string, provider *EmailProvider) error
}
