package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// WebhookRegistrationService implements the domain.WebhookRegistrationService interface
type WebhookRegistrationService struct {
	workspaceRepo    domain.WorkspaceRepository
	authService      domain.AuthService
	postmarkService  domain.PostmarkServiceInterface
	mailgunService   domain.MailgunServiceInterface
	mailjetService   domain.MailjetServiceInterface
	sparkPostService domain.SparkPostServiceInterface
	sesService       domain.SESServiceInterface
	logger           logger.Logger
}

// NewWebhookRegistrationService creates a new webhook registration service
func NewWebhookRegistrationService(
	workspaceRepo domain.WorkspaceRepository,
	authService domain.AuthService,
	postmarkService domain.PostmarkServiceInterface,
	mailgunService domain.MailgunServiceInterface,
	mailjetService domain.MailjetServiceInterface,
	sparkPostService domain.SparkPostServiceInterface,
	sesService domain.SESServiceInterface,
	logger logger.Logger,
) *WebhookRegistrationService {
	return &WebhookRegistrationService{
		workspaceRepo:    workspaceRepo,
		authService:      authService,
		postmarkService:  postmarkService,
		mailgunService:   mailgunService,
		mailjetService:   mailjetService,
		sparkPostService: sparkPostService,
		sesService:       sesService,
		logger:           logger,
	}
}

// RegisterWebhooks registers webhook URLs with the email provider
func (s *WebhookRegistrationService) RegisterWebhooks(
	ctx context.Context,
	workspaceID string,
	config *domain.WebhookRegistrationConfig,
) (*domain.WebhookRegistrationStatus, error) {
	// Authenticate the user for this workspace
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Get email provider configuration from workspace settings
	emailProvider, err := s.getEmailProviderConfig(ctx, workspaceID, config.IntegrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get email provider configuration: %w", err)
	}

	// Make sure the provider kind matches
	if string(emailProvider.Kind) != config.IntegrationID {
		return nil, fmt.Errorf("email provider kind mismatch: config has %s, workspace has %s",
			config.IntegrationID, emailProvider.Kind)
	}

	// Convert webhook base URL if needed (remove trailing slash)
	baseURL := strings.TrimSuffix(config.BaseURL, "/")

	// Register webhooks based on provider kind
	switch emailProvider.Kind {
	case domain.EmailProviderKindPostmark:
		return s.registerPostmarkWebhooks(ctx, workspaceID, config.IntegrationID, baseURL, config.EventTypes, emailProvider.Postmark)
	case domain.EmailProviderKindMailgun:
		return s.registerMailgunWebhooks(ctx, workspaceID, config.IntegrationID, baseURL, config.EventTypes, emailProvider.Mailgun)
	case domain.EmailProviderKindMailjet:
		return s.registerMailjetWebhooks(ctx, workspaceID, config.IntegrationID, baseURL, config.EventTypes, emailProvider.Mailjet)
	case domain.EmailProviderKindSparkPost:
		return s.registerSparkPostWebhooks(ctx, workspaceID, config.IntegrationID, baseURL, config.EventTypes, emailProvider.SparkPost)
	case domain.EmailProviderKindSES:
		return s.registerSESWebhooks(ctx, workspaceID, config.IntegrationID, baseURL, config.EventTypes, emailProvider.SES)
	case domain.EmailProviderKindSMTP:
		// For SMTP, we can't register webhooks automatically - it depends on the SMTP provider
		return &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      false,
			Error:             "SMTP does not support automatic webhook registration",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported email provider kind: %s", emailProvider.Kind)
	}
}

// GetWebhookStatus gets the status of webhooks for an email provider
func (s *WebhookRegistrationService) GetWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
) (*domain.WebhookRegistrationStatus, error) {
	// Authenticate the user for this workspace
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Get email provider configuration from workspace settings
	emailProvider, err := s.getEmailProviderConfig(ctx, workspaceID, integrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get email provider configuration: %w", err)
	}

	// Get webhook status based on provider kind
	switch emailProvider.Kind {
	case domain.EmailProviderKindPostmark:
		return s.getPostmarkWebhookStatus(ctx, workspaceID, integrationID, emailProvider.Postmark)
	case domain.EmailProviderKindMailgun:
		return s.getMailgunWebhookStatus(ctx, workspaceID, integrationID, emailProvider.Mailgun)
	case domain.EmailProviderKindMailjet:
		return s.getMailjetWebhookStatus(ctx, workspaceID, integrationID, emailProvider.Mailjet)
	case domain.EmailProviderKindSparkPost:
		return s.getSparkPostWebhookStatus(ctx, workspaceID, integrationID, emailProvider.SparkPost)
	case domain.EmailProviderKindSES:
		return s.getSESWebhookStatus(ctx, workspaceID, integrationID, emailProvider.SES)
	case domain.EmailProviderKindSMTP:
		// For SMTP, we can't check webhook status automatically
		return &domain.WebhookRegistrationStatus{
			EmailProviderKind: domain.EmailProviderKindSMTP,
			IsRegistered:      false,
			Error:             "SMTP does not support automatic webhook status checks",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported email provider kind: %s", emailProvider.Kind)
	}
}

// getEmailProviderConfig gets the email provider configuration from workspace settings
func (s *WebhookRegistrationService) getEmailProviderConfig(ctx context.Context, workspaceID string, integrationID string) (*domain.EmailProvider, error) {
	// Get workspace settings from the database
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Find the integration by ID
	integration := workspace.GetIntegrationByID(integrationID)
	if integration == nil {
		return nil, fmt.Errorf("integration with ID %s not found", integrationID)
	}

	return &integration.EmailProvider, nil
}

// registerPostmarkWebhooks registers webhooks with Postmark
func (s *WebhookRegistrationService) registerPostmarkWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.PostmarkSettings,
) (*domain.WebhookRegistrationStatus, error) {
	if providerConfig == nil || providerConfig.ServerToken == "" {
		return nil, fmt.Errorf("Postmark configuration is missing or invalid")
	}

	// Create Postmark API config
	config := domain.PostmarkConfig{
		APIEndpoint: "https://api.postmarkapp.com",
		ServerToken: providerConfig.ServerToken,
	}

	// First, get existing webhooks
	existingWebhooks, err := s.postmarkService.ListWebhooks(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to list Postmark webhooks: %w", err)
	}

	// Check if we already have webhooks registered
	notifuseWebhooks := filterPostmarkWebhooks(existingWebhooks.Webhooks, baseURL)

	// If we have existing webhooks, unregister them
	for _, webhook := range notifuseWebhooks {
		err := s.postmarkService.UnregisterWebhook(ctx, config, webhook.ID)
		if err != nil {
			s.logger.WithField("webhook_id", webhook.ID).
				Error(fmt.Sprintf("Failed to unregister Postmark webhook: %v", err))
			// Continue with other webhooks
		}
	}

	// Create webhook configuration
	webhookURL := fmt.Sprintf("%s/webhooks/email/postmark?workspace_id=%s&integration_id=%s", baseURL, workspaceID, integrationID)
	triggers := []domain.PostmarkTriggerRule{}

	// Add triggers for each event type
	for _, eventType := range eventTypes {
		var triggerValue string
		switch eventType {
		case domain.EmailEventDelivered:
			triggerValue = "Delivery"
		case domain.EmailEventBounce:
			triggerValue = "Bounce"
		case domain.EmailEventComplaint:
			triggerValue = "SpamComplaint"
		default:
			continue // Skip unsupported event types
		}

		triggers = append(triggers, domain.PostmarkTriggerRule{
			Key:   "MessageStream",
			Match: "Equals",
			Value: "outbound",
		})

		triggers = append(triggers, domain.PostmarkTriggerRule{
			Key:   "RecordType",
			Match: "Equals",
			Value: triggerValue,
		})
	}

	// Register new webhook
	webhookConfig := domain.PostmarkWebhookConfig{
		URL:           webhookURL,
		MessageStream: "outbound",
		TriggerRules:  triggers,
	}

	webhookResponse, err := s.postmarkService.RegisterWebhook(ctx, config, webhookConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to register Postmark webhook: %w", err)
	}

	// Create webhook status
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindPostmark,
		IsRegistered:      true,
		RegisteredEvents:  eventTypes,
		Endpoints: []domain.WebhookEndpointStatus{
			{
				URL:    webhookURL,
				Active: true,
			},
		},
		ProviderDetails: map[string]interface{}{
			"webhook_id":     webhookResponse.ID,
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}

	return status, nil
}

// getPostmarkWebhookStatus gets the webhook status for Postmark
func (s *WebhookRegistrationService) getPostmarkWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.PostmarkSettings,
) (*domain.WebhookRegistrationStatus, error) {
	if providerConfig == nil || providerConfig.ServerToken == "" {
		return nil, fmt.Errorf("Postmark configuration is missing or invalid")
	}

	// Create Postmark API config
	config := domain.PostmarkConfig{
		APIEndpoint: "https://api.postmarkapp.com",
		ServerToken: providerConfig.ServerToken,
	}

	// Get existing webhooks
	existingWebhooks, err := s.postmarkService.ListWebhooks(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to list Postmark webhooks: %w", err)
	}

	// Create webhook status
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindPostmark,
		IsRegistered:      false,
		Endpoints:         []domain.WebhookEndpointStatus{},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}

	// Check each webhook in the response
	for _, webhook := range existingWebhooks.Webhooks {
		status.Endpoints = append(status.Endpoints, domain.WebhookEndpointStatus{
			URL:    webhook.URL,
			Active: true,
		})

		// Determine registered event types
		for _, trigger := range webhook.Triggers {
			if trigger.Key == "RecordType" {
				var eventType domain.EmailEventType
				switch trigger.Value {
				case "Delivery":
					eventType = domain.EmailEventDelivered
				case "Bounce":
					eventType = domain.EmailEventBounce
				case "SpamComplaint":
					eventType = domain.EmailEventComplaint
				default:
					continue
				}

				// Add to registered events if not already there
				found := false
				for _, registeredEvent := range status.RegisteredEvents {
					if registeredEvent == eventType {
						found = true
						break
					}
				}
				if !found {
					status.RegisteredEvents = append(status.RegisteredEvents, eventType)
				}
			}
		}

		// Mark as registered if we have any endpoints
		if len(status.Endpoints) > 0 {
			status.IsRegistered = true
		}
	}

	return status, nil
}

// Additional provider-specific webhook registration and status check methods would go here
// For brevity, I'm showing just the Postmark implementation
// Similar methods would be needed for:
// - registerMailgunWebhooks / getMailgunWebhookStatus
// - registerMailjetWebhooks / getMailjetWebhookStatus
// - registerSparkPostWebhooks / getSparkPostWebhookStatus
// - registerSESWebhooks / getSESWebhookStatus

// Helper function to filter Postmark webhooks by base URL
func filterPostmarkWebhooks(webhooks []domain.PostmarkWebhookResponse, baseURL string) []domain.PostmarkWebhookResponse {
	var filtered []domain.PostmarkWebhookResponse
	for _, webhook := range webhooks {
		if strings.Contains(webhook.URL, baseURL) {
			filtered = append(filtered, webhook)
		}
	}
	return filtered
}

// These are stub implementations that would need to be completed
func (s *WebhookRegistrationService) registerMailgunWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.MailgunSettings,
) (*domain.WebhookRegistrationStatus, error) {
	// Webhook URL would include both workspace_id and integration_id
	webhookURL := fmt.Sprintf("%s/webhooks/email/mailgun?workspace_id=%s&integration_id=%s", baseURL, workspaceID, integrationID)

	// Implementation would use s.mailgunService to register webhooks with the updated URL
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailgun,
		IsRegistered:      false,
		Error:             "Mailgun webhook registration not implemented",
		Endpoints: []domain.WebhookEndpointStatus{
			{
				URL:    webhookURL,
				Active: false,
			},
		},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}

func (s *WebhookRegistrationService) getMailgunWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.MailgunSettings,
) (*domain.WebhookRegistrationStatus, error) {
	// Implementation would use s.mailgunService to get webhook status
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailgun,
		IsRegistered:      false,
		Error:             "Mailgun webhook status check not implemented",
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}

func (s *WebhookRegistrationService) registerMailjetWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.MailjetSettings,
) (*domain.WebhookRegistrationStatus, error) {
	// Webhook URL would include both workspace_id and integration_id
	webhookURL := fmt.Sprintf("%s/webhooks/email/mailjet?workspace_id=%s&integration_id=%s", baseURL, workspaceID, integrationID)

	// Implementation would use s.mailjetService to register webhooks with the updated URL
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailjet,
		IsRegistered:      false,
		Error:             "Mailjet webhook registration not implemented",
		Endpoints: []domain.WebhookEndpointStatus{
			{
				URL:    webhookURL,
				Active: false,
			},
		},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}

func (s *WebhookRegistrationService) getMailjetWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.MailjetSettings,
) (*domain.WebhookRegistrationStatus, error) {
	// Implementation would use s.mailjetService to get webhook status
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailjet,
		IsRegistered:      false,
		Error:             "Mailjet webhook status check not implemented",
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}

func (s *WebhookRegistrationService) registerSparkPostWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.SparkPostSettings,
) (*domain.WebhookRegistrationStatus, error) {
	// Webhook URL would include both workspace_id and integration_id
	webhookURL := fmt.Sprintf("%s/webhooks/email/sparkpost?workspace_id=%s&integration_id=%s", baseURL, workspaceID, integrationID)

	// Implementation would use s.sparkPostService to register webhooks with the updated URL
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSparkPost,
		IsRegistered:      false,
		Error:             "SparkPost webhook registration not implemented",
		Endpoints: []domain.WebhookEndpointStatus{
			{
				URL:    webhookURL,
				Active: false,
			},
		},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}

func (s *WebhookRegistrationService) getSparkPostWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.SparkPostSettings,
) (*domain.WebhookRegistrationStatus, error) {
	// Implementation would use s.sparkPostService to get webhook status
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSparkPost,
		IsRegistered:      false,
		Error:             "SparkPost webhook status check not implemented",
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}

func (s *WebhookRegistrationService) registerSESWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.AmazonSES,
) (*domain.WebhookRegistrationStatus, error) {
	// Webhook URL would include both workspace_id and integration_id
	webhookURL := fmt.Sprintf("%s/webhooks/email/ses?workspace_id=%s&integration_id=%s", baseURL, workspaceID, integrationID)

	// Implementation would use s.sesService to register webhooks with the updated URL
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSES,
		IsRegistered:      false,
		Error:             "SES webhook registration not implemented",
		Endpoints: []domain.WebhookEndpointStatus{
			{
				URL:    webhookURL,
				Active: false,
			},
		},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}

func (s *WebhookRegistrationService) getSESWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.AmazonSES,
) (*domain.WebhookRegistrationStatus, error) {
	// Implementation would use s.sesService to get webhook status
	return &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSES,
		IsRegistered:      false,
		Error:             "SES webhook status check not implemented",
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}, nil
}
