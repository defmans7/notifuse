package domain

import (
	"context"
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

// SESConfig represents configuration for Amazon SES API
type SESConfig struct {
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

//go:generate mockgen -destination mocks/mock_ses_service.go -package mocks github.com/Notifuse/notifuse/internal/domain SESServiceInterface

// SESServiceInterface defines operations for managing Amazon SES webhooks via SNS
type SESServiceInterface interface {
	// ListConfigurationSets lists all configuration sets
	ListConfigurationSets(ctx context.Context, config SESConfig) ([]string, error)

	// CreateConfigurationSet creates a new configuration set
	CreateConfigurationSet(ctx context.Context, config SESConfig, name string) error

	// DeleteConfigurationSet deletes a configuration set
	DeleteConfigurationSet(ctx context.Context, config SESConfig, name string) error

	// CreateSNSTopic creates a new SNS topic for notifications
	CreateSNSTopic(ctx context.Context, config SESConfig, topicConfig SESTopicConfig) (string, error)

	// DeleteSNSTopic deletes an SNS topic
	DeleteSNSTopic(ctx context.Context, config SESConfig, topicARN string) error

	// CreateEventDestination creates an event destination in a configuration set
	CreateEventDestination(ctx context.Context, config SESConfig, destination SESConfigurationSetEventDestination) error

	// UpdateEventDestination updates an event destination
	UpdateEventDestination(ctx context.Context, config SESConfig, destination SESConfigurationSetEventDestination) error

	// DeleteEventDestination deletes an event destination
	DeleteEventDestination(ctx context.Context, config SESConfig, configSetName, destinationName string) error

	// ListEventDestinations lists all event destinations for a configuration set
	ListEventDestinations(ctx context.Context, config SESConfig, configSetName string) ([]SESConfigurationSetEventDestination, error)
}
