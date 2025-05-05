package domain

import (
	"context"
	"fmt"

	"github.com/Notifuse/notifuse/pkg/crypto"
)

//go:generate mockgen -destination mocks/mock_ses_service.go -package mocks github.com/Notifuse/notifuse/internal/domain SESServiceInterface

// SESWebhookPayload represents an Amazon SES webhook payload
type SESWebhookPayload struct {
	Type              string                         `json:"Type"`
	MessageID         string                         `json:"MessageId"`
	TopicARN          string                         `json:"TopicArn"`
	Message           string                         `json:"Message"`
	Timestamp         string                         `json:"Timestamp"`
	SignatureVersion  string                         `json:"SignatureVersion"`
	Signature         string                         `json:"Signature"`
	SigningCertURL    string                         `json:"SigningCertURL"`
	UnsubscribeURL    string                         `json:"UnsubscribeURL"`
	MessageAttributes map[string]SESMessageAttribute `json:"MessageAttributes"`
}

// SESMessageAttribute represents a message attribute in SES webhook
type SESMessageAttribute struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

// SESBounceNotification represents an SES bounce notification
type SESBounceNotification struct {
	NotificationType string    `json:"notificationType"`
	Bounce           SESBounce `json:"bounce"`
	Mail             SESMail   `json:"mail"`
}

// SESComplaintNotification represents an SES complaint notification
type SESComplaintNotification struct {
	NotificationType string       `json:"notificationType"`
	Complaint        SESComplaint `json:"complaint"`
	Mail             SESMail      `json:"mail"`
}

// SESDeliveryNotification represents an SES delivery notification
type SESDeliveryNotification struct {
	NotificationType string      `json:"notificationType"`
	Delivery         SESDelivery `json:"delivery"`
	Mail             SESMail     `json:"mail"`
}

// SESMail represents the mail part of an SES notification
type SESMail struct {
	Timestamp        string            `json:"timestamp"`
	MessageID        string            `json:"messageId"`
	Source           string            `json:"source"`
	Destination      []string          `json:"destination"`
	HeadersTruncated bool              `json:"headersTruncated"`
	Headers          []SESHeader       `json:"headers"`
	CommonHeaders    SESCommonHeaders  `json:"commonHeaders"`
	Tags             map[string]string `json:"tags"`
}

// SESHeader represents a header in an SES notification
type SESHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SESCommonHeaders represents common headers in an SES notification
type SESCommonHeaders struct {
	From      []string `json:"from"`
	To        []string `json:"to"`
	MessageID string   `json:"messageId"`
	Subject   string   `json:"subject"`
}

// SESBounce represents a bounce in an SES notification
type SESBounce struct {
	BounceType        string                `json:"bounceType"`
	BounceSubType     string                `json:"bounceSubType"`
	BouncedRecipients []SESBouncedRecipient `json:"bouncedRecipients"`
	Timestamp         string                `json:"timestamp"`
	FeedbackID        string                `json:"feedbackId"`
	ReportingMTA      string                `json:"reportingMTA"`
}

// SESBouncedRecipient represents a bounced recipient in an SES notification
type SESBouncedRecipient struct {
	EmailAddress   string `json:"emailAddress"`
	Action         string `json:"action"`
	Status         string `json:"status"`
	DiagnosticCode string `json:"diagnosticCode"`
}

// SESComplaint represents a complaint in an SES notification
type SESComplaint struct {
	ComplainedRecipients  []SESComplainedRecipient `json:"complainedRecipients"`
	Timestamp             string                   `json:"timestamp"`
	FeedbackID            string                   `json:"feedbackId"`
	ComplaintFeedbackType string                   `json:"complaintFeedbackType"`
}

// SESComplainedRecipient represents a complained recipient in an SES notification
type SESComplainedRecipient struct {
	EmailAddress string `json:"emailAddress"`
}

// SESDelivery represents a delivery in an SES notification
type SESDelivery struct {
	Timestamp            string   `json:"timestamp"`
	ProcessingTimeMillis int      `json:"processingTimeMillis"`
	Recipients           []string `json:"recipients"`
	SMTPResponse         string   `json:"smtpResponse"`
	ReportingMTA         string   `json:"reportingMTA"`
}

// SESTopicConfig represents AWS SNS topic configuration
type SESTopicConfig struct {
	TopicARN             string `json:"topic_arn"`
	TopicName            string `json:"topic_name,omitempty"`
	NotificationEndpoint string `json:"notification_endpoint"`
	Protocol             string `json:"protocol"` // Usually "https"
}

// SESConfigurationSetEventDestination represents SES event destination configuration
type SESConfigurationSetEventDestination struct {
	Name                 string          `json:"name"`
	ConfigurationSetName string          `json:"configuration_set_name"`
	Enabled              bool            `json:"enabled"`
	MatchingEventTypes   []string        `json:"matching_event_types"`
	SNSDestination       *SESTopicConfig `json:"sns_destination,omitempty"`
}

// AmazonSESSettings contains SES email provider settings
type AmazonSESSettings struct {
	Region             string `json:"region"`
	AccessKey          string `json:"access_key"`
	EncryptedSecretKey string `json:"encrypted_secret_key,omitempty"`
	SandboxMode        bool   `json:"sandbox_mode"`

	// decoded secret key, not stored in the database
	SecretKey string `json:"secret_key,omitempty"`
}

func (a *AmazonSESSettings) DecryptSecretKey(passphrase string) error {
	secretKey, err := crypto.DecryptFromHexString(a.EncryptedSecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt SES secret key: %w", err)
	}
	a.SecretKey = secretKey
	return nil
}

func (a *AmazonSESSettings) EncryptSecretKey(passphrase string) error {
	encryptedSecretKey, err := crypto.EncryptString(a.SecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt SES secret key: %w", err)
	}
	a.EncryptedSecretKey = encryptedSecretKey
	return nil
}

func (a *AmazonSESSettings) Validate(passphrase string) error {
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

//go:generate mockgen -destination mocks/mock_ses_service.go -package mocks github.com/Notifuse/notifuse/internal/domain SESServiceInterface

// SESServiceInterface defines operations for managing Amazon SES webhooks via SNS
type SESServiceInterface interface {
	// ListConfigurationSets lists all configuration sets
	ListConfigurationSets(ctx context.Context, config AmazonSESSettings) ([]string, error)

	// CreateConfigurationSet creates a new configuration set
	CreateConfigurationSet(ctx context.Context, config AmazonSESSettings, name string) error

	// DeleteConfigurationSet deletes a configuration set
	DeleteConfigurationSet(ctx context.Context, config AmazonSESSettings, name string) error

	// CreateSNSTopic creates a new SNS topic for notifications
	CreateSNSTopic(ctx context.Context, config AmazonSESSettings, topicConfig SESTopicConfig) (string, error)

	// DeleteSNSTopic deletes an SNS topic
	DeleteSNSTopic(ctx context.Context, config AmazonSESSettings, topicARN string) error

	// CreateEventDestination creates an event destination in a configuration set
	CreateEventDestination(ctx context.Context, config AmazonSESSettings, destination SESConfigurationSetEventDestination) error

	// UpdateEventDestination updates an event destination
	UpdateEventDestination(ctx context.Context, config AmazonSESSettings, destination SESConfigurationSetEventDestination) error

	// DeleteEventDestination deletes an event destination
	DeleteEventDestination(ctx context.Context, config AmazonSESSettings, configSetName, destinationName string) error

	// ListEventDestinations lists all event destinations for a configuration set
	ListEventDestinations(ctx context.Context, config AmazonSESSettings, configSetName string) ([]SESConfigurationSetEventDestination, error)
}
