package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/wneessen/go-mail"
)

// SMTPService implements the domain.EmailProviderService interface for SMTP
type SMTPService struct {
	logger logger.Logger
}

// NewSMTPService creates a new instance of SMTPService
func NewSMTPService(logger logger.Logger) *SMTPService {
	return &SMTPService{
		logger: logger,
	}
}

// SendEmail sends an email using SMTP
func (s *SMTPService) SendEmail(ctx context.Context, workspaceID string, fromAddress, fromName, to, subject, content string, provider *domain.EmailProvider) error {
	if provider.SMTP == nil {
		return fmt.Errorf("SMTP settings required")
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
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
