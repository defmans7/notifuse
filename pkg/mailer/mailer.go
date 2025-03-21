package mailer

import (
	"fmt"
	"log"
)

// Mailer is the interface for sending emails
type Mailer interface {
	// SendWorkspaceInvitation sends an invitation email with the given token
	SendWorkspaceInvitation(email, workspaceName, inviterName, token string) error
}

// Config holds the configuration for the mailer
type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BaseURL      string
}

// SMTPMailer implements the Mailer interface using SMTP
type SMTPMailer struct {
	config *Config
}

// NewSMTPMailer creates a new SMTP mailer
func NewSMTPMailer(config *Config) *SMTPMailer {
	return &SMTPMailer{
		config: config,
	}
}

// SendWorkspaceInvitation sends an invitation email with the given token
func (m *SMTPMailer) SendWorkspaceInvitation(email, workspaceName, inviterName, token string) error {
	// For now, just log the email details
	// In a real implementation, use a real SMTP client to send the email
	inviteURL := fmt.Sprintf("%s/invitation?token=%s", m.config.BaseURL, token)

	log.Printf("Sending invitation email to: %s", email)
	log.Printf("From: %s <%s>", m.config.FromName, m.config.FromEmail)
	log.Printf("Subject: You've been invited to join %s on Notifuse", workspaceName)
	log.Printf("Invitation URL: %s", inviteURL)

	// In a real implementation, you would:
	// 1. Create a proper HTML/text email template
	// 2. Set up SMTP client with authentication
	// 3. Send the email through the SMTP server

	return nil
}

// ConsoleMailer is a development implementation that just logs emails
type ConsoleMailer struct{}

// NewConsoleMailer creates a new console mailer for development
func NewConsoleMailer() *ConsoleMailer {
	return &ConsoleMailer{}
}

// SendWorkspaceInvitation logs the invitation details to console
func (m *ConsoleMailer) SendWorkspaceInvitation(email, workspaceName, inviterName, token string) error {
	fmt.Println("==============================================================")
	fmt.Println("                 WORKSPACE INVITATION EMAIL                   ")
	fmt.Println("==============================================================")
	fmt.Printf("To: %s\n", email)
	fmt.Printf("Subject: You've been invited to join %s\n\n", workspaceName)
	fmt.Println("Email Content:")
	fmt.Printf("Hello,\n\n")
	fmt.Printf("%s has invited you to join the %s workspace on Notifuse.\n\n", inviterName, workspaceName)
	fmt.Printf("Use the following token to join: %s\n\n", token)
	fmt.Println("==============================================================")

	return nil
}
