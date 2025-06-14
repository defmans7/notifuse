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
	clientFactory domain.SMTPClientFactory
}

// defaultGoMailFactory implements the domain.SMTPClientFactory interface directly using go-mail
type defaultGoMailFactory struct{}

func (f *defaultGoMailFactory) CreateClient(host string, port int, username, password string, useTLS bool) (*mail.Client, error) {
	var tlsPolicy mail.TLSPolicy
	var clientOptions []mail.Option

	// Configure TLS policy
	if useTLS {
		tlsPolicy = mail.TLSMandatory
	} else {
		// For local development servers like Mailpit, disable TLS completely
		tlsPolicy = mail.NoTLS
	}

	// Basic client options
	clientOptions = append(clientOptions,
		mail.WithPort(port),
		mail.WithTLSPolicy(tlsPolicy),
		mail.WithTimeout(10*time.Second),
	)

	// Only add authentication if username and password are provided
	// This allows for servers like Mailpit that don't require authentication
	if username != "" && password != "" {
		clientOptions = append(clientOptions,
			mail.WithUsername(username),
			mail.WithPassword(password),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
		)
	}

	client, err := mail.NewClient(host, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return client, nil
}

// NewSMTPService creates a new instance of SMTPService
func NewSMTPService(logger logger.Logger) *SMTPService {
	return &SMTPService{
		logger:        logger,
		clientFactory: &defaultGoMailFactory{},
	}
}

// SendEmail sends an email using SMTP
func (s *SMTPService) SendEmail(ctx context.Context, messageID string, workspaceID string, fromAddress, fromName, to, subject, content string, provider *domain.EmailProvider, emailOptions domain.EmailOptions) error {
	if provider.SMTP == nil {
		return fmt.Errorf("SMTP settings required")
	}

	// Create a client directly
	client, err := s.clientFactory.CreateClient(
		provider.SMTP.Host,
		provider.SMTP.Port,
		provider.SMTP.Username,
		provider.SMTP.Password,
		provider.SMTP.UseTLS,
	)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	if client == nil {
		return fmt.Errorf("SMTP client factory returned nil client")
	}
	defer client.Close()

	// Create and configure the message
	msg := mail.NewMsg(mail.WithNoDefaultUserAgent())

	if err := msg.FromFormat(fromName, fromAddress); err != nil {
		return fmt.Errorf("invalid sender: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	// Add CC recipients if specified
	for _, ccAddr := range emailOptions.CC {
		if ccAddr != "" {
			if err := msg.Cc(ccAddr); err != nil {
				return fmt.Errorf("invalid CC recipient: %w", err)
			}
		}
	}

	// Add BCC recipients if specified
	for _, bccAddr := range emailOptions.BCC {
		if bccAddr != "" {
			if err := msg.Bcc(bccAddr); err != nil {
				return fmt.Errorf("invalid BCC recipient: %w", err)
			}
		}
	}

	// Add Reply-To if specified
	if emailOptions.ReplyTo != "" {
		if err := msg.ReplyTo(emailOptions.ReplyTo); err != nil {
			return fmt.Errorf("invalid reply-to address: %w", err)
		}
	}

	// Add message ID tracking header
	msg.SetGenHeader("X-Message-ID", messageID)

	// Remove User-Agent and X-Mailer headers
	// msg.SetUserAgent("")

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, content)

	// Send the email directly
	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
