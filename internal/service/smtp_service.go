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
	logger        logger.Logger
	clientFactory ClientFactory
}

// ClientFactory defines an interface for creating mail clients
type ClientFactory interface {
	NewClient(host string, port int, username, password string, useTLS bool) (MailClient, error)
}

// MailClient defines an interface for mail client operations
type MailClient interface {
	Send(from, fromName, to, subject, content string) error
	Close() error
}

// defaultGoMailFactory implements the ClientFactory interface for go-mail
type defaultGoMailFactory struct{}

func (f *defaultGoMailFactory) NewClient(host string, port int, username, password string, useTLS bool) (MailClient, error) {
	tlsPolicy := mail.TLSOpportunistic
	if useTLS {
		tlsPolicy = mail.TLSMandatory
	}

	client, err := mail.NewClient(
		host,
		mail.WithPort(port),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTLSPolicy(tlsPolicy),
		mail.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return &goMailClientAdapter{client: client}, nil
}

// goMailClientAdapter adapts go-mail.Client to our MailClient interface
type goMailClientAdapter struct {
	client *mail.Client
}

func (a *goMailClientAdapter) Send(from, fromName, to, subject, content string) error {
	msg := mail.NewMsg()
	if err := msg.FromFormat(from, fromName); err != nil {
		return fmt.Errorf("invalid sender: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, content)

	if err := a.client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (a *goMailClientAdapter) Close() error {
	return a.client.Close()
}

// NewSMTPService creates a new instance of SMTPService
func NewSMTPService(logger logger.Logger) *SMTPService {
	return &SMTPService{
		logger:        logger,
		clientFactory: &defaultGoMailFactory{},
	}
}

// SendEmail sends an email using SMTP
func (s *SMTPService) SendEmail(ctx context.Context, workspaceID string, fromAddress, fromName, to, subject, content string, provider *domain.EmailProvider) error {
	if provider.SMTP == nil {
		return fmt.Errorf("SMTP settings required")
	}

	// Create a client
	client, err := s.clientFactory.NewClient(
		provider.SMTP.Host,
		provider.SMTP.Port,
		provider.SMTP.Username,
		provider.SMTP.Password,
		provider.SMTP.UseTLS,
	)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Send the email
	if err := client.Send(fromAddress, fromName, to, subject, content); err != nil {
		return err
	}

	return nil
}
