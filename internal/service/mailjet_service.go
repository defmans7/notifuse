package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// MailjetService implements the domain.MailjetServiceInterface
type MailjetService struct {
	httpClient  domain.HTTPClient
	authService domain.AuthService
	logger      logger.Logger
}

// NewMailjetService creates a new instance of MailjetService
func NewMailjetService(httpClient domain.HTTPClient, authService domain.AuthService, logger logger.Logger) *MailjetService {
	return &MailjetService{
		httpClient:  httpClient,
		authService: authService,
		logger:      logger,
	}
}

// ListWebhooks retrieves all registered webhooks
func (s *MailjetService) ListWebhooks(ctx context.Context, config domain.MailjetSettings) (*domain.MailjetWebhookResponse, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.mailjet.com/v3/eventcallback", nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for listing Mailjet webhooks: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Mailjet uses Basic Auth with API Key and Secret Key
	req.SetBasicAuth(config.APIKey, config.SecretKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for listing Mailjet webhooks: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailjet API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response
	var response domain.MailjetWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailjet webhook list response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CreateWebhook creates a new webhook
func (s *MailjetService) CreateWebhook(ctx context.Context, config domain.MailjetSettings, webhook domain.MailjetWebhook) (*domain.MailjetWebhook, error) {

	// Prepare the request body
	requestBody, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.mailjet.com/v3/eventcallback", bytes.NewBuffer(requestBody))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for creating Mailjet webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Mailjet uses Basic Auth with API Key and Secret Key
	req.SetBasicAuth(config.APIKey, config.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for creating Mailjet webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailjet API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response to get the created webhook details
	var createdWebhook domain.MailjetWebhook
	if err := json.NewDecoder(resp.Body).Decode(&createdWebhook); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailjet webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdWebhook, nil
}

// GetWebhook retrieves a webhook by ID
func (s *MailjetService) GetWebhook(ctx context.Context, config domain.MailjetSettings, webhookID int64) (*domain.MailjetWebhook, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.mailjet.com/v3/eventcallback/%d", webhookID), nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for getting Mailjet webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Mailjet uses Basic Auth with API Key and Secret Key
	req.SetBasicAuth(config.APIKey, config.SecretKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for getting Mailjet webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailjet API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response to get the webhook details
	var webhookResponse struct {
		Data []domain.MailjetWebhook `json:"Data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&webhookResponse); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailjet webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(webhookResponse.Data) == 0 {
		return nil, fmt.Errorf("webhook with ID %d not found", webhookID)
	}

	return &webhookResponse.Data[0], nil
}

// UpdateWebhook updates an existing webhook
func (s *MailjetService) UpdateWebhook(ctx context.Context, config domain.MailjetSettings, webhookID int64, webhook domain.MailjetWebhook) (*domain.MailjetWebhook, error) {

	// Ensure the webhook ID in the URL matches the one in the body
	webhook.ID = webhookID

	// Prepare the request body
	requestBody, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("https://api.mailjet.com/v3/eventcallback/%d", webhookID), bytes.NewBuffer(requestBody))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for updating Mailjet webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Mailjet uses Basic Auth with API Key and Secret Key
	req.SetBasicAuth(config.APIKey, config.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for updating Mailjet webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailjet API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response to get the updated webhook details
	var updatedWebhook domain.MailjetWebhook
	if err := json.NewDecoder(resp.Body).Decode(&updatedWebhook); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailjet webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedWebhook, nil
}

// DeleteWebhook deletes a webhook by ID
func (s *MailjetService) DeleteWebhook(ctx context.Context, config domain.MailjetSettings, webhookID int64) error {

	// Log webhook ID for debugging
	webhookIDStr := strconv.FormatInt(webhookID, 10)
	s.logger = s.logger.WithField("webhook_id", webhookIDStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("https://api.mailjet.com/v3/eventcallback/%d", webhookID), nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for deleting Mailjet webhook: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Mailjet uses Basic Auth with API Key and Secret Key
	req.SetBasicAuth(config.APIKey, config.SecretKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for deleting Mailjet webhook: %v", err))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailjet API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	return nil
}

// RegisterWebhooks implements the domain.WebhookProvider interface for Mailjet
func (s *MailjetService) RegisterWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.EmailProvider,
) (*domain.WebhookRegistrationStatus, error) {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.Mailjet == nil ||
		providerConfig.Mailjet.APIKey == "" || providerConfig.Mailjet.SecretKey == "" {
		return nil, fmt.Errorf("Mailjet configuration is missing or invalid")
	}

	// Create webhook URL that includes workspace_id and integration_id
	webhookURL := domain.GenerateWebhookCallbackURL(baseURL, domain.EmailProviderKindMailjet, workspaceID, integrationID)

	// Map our event types to Mailjet event types
	var registeredEvents []domain.EmailEventType
	var mailjetEvents []domain.MailjetWebhookEventType

	for _, eventType := range eventTypes {
		switch eventType {
		case domain.EmailEventDelivered:
			mailjetEvents = append(mailjetEvents, domain.MailjetEventSent)
			registeredEvents = append(registeredEvents, domain.EmailEventDelivered)
		case domain.EmailEventBounce:
			mailjetEvents = append(mailjetEvents, domain.MailjetEventBounce)
			mailjetEvents = append(mailjetEvents, domain.MailjetEventBlocked)
			registeredEvents = append(registeredEvents, domain.EmailEventBounce)
		case domain.EmailEventComplaint:
			mailjetEvents = append(mailjetEvents, domain.MailjetEventSpam)
			registeredEvents = append(registeredEvents, domain.EmailEventComplaint)
		}
	}

	// First, get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, *providerConfig.Mailjet)
	if err != nil {
		return nil, fmt.Errorf("failed to list Mailjet webhooks: %w", err)
	}

	// Check for existing webhooks that match our criteria
	var notifuseWebhooks []domain.MailjetWebhook
	for _, webhook := range existingWebhooks.Data {
		if strings.Contains(webhook.Endpoint, baseURL) &&
			strings.Contains(webhook.Endpoint, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.Endpoint, fmt.Sprintf("integration_id=%s", integrationID)) {
			notifuseWebhooks = append(notifuseWebhooks, webhook)
		}
	}

	// Delete existing webhooks
	for _, webhook := range notifuseWebhooks {
		err := s.DeleteWebhook(ctx, *providerConfig.Mailjet, webhook.ID)
		if err != nil {
			s.logger.WithField("webhook_id", webhook.ID).
				Error(fmt.Sprintf("Failed to delete Mailjet webhook: %v", err))
			// Continue with other webhooks even if one fails
		}
	}

	// Create a new webhook
	// Mailjet allows us to create a single webhook for multiple event types
	webhookConfig := domain.MailjetWebhook{
		Endpoint:  webhookURL,
		EventType: string(mailjetEvents[0]), // The primary event type
		Status:    "active",
	}

	webhook, err := s.CreateWebhook(ctx, *providerConfig.Mailjet, webhookConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mailjet webhook: %w", err)
	}

	// Create webhook registration status
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailjet,
		IsRegistered:      true,
		RegisteredEvents:  registeredEvents,
		Endpoints: []domain.WebhookEndpointStatus{
			{
				URL:    webhookURL,
				Active: webhook.Status == "active",
			},
		},
		ProviderDetails: map[string]interface{}{
			"webhook_id":     webhook.ID,
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}

	return status, nil
}

// GetWebhookStatus implements the domain.WebhookProvider interface for Mailjet
func (s *MailjetService) GetWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.EmailProvider,
) (*domain.WebhookRegistrationStatus, error) {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.Mailjet == nil ||
		providerConfig.Mailjet.APIKey == "" || providerConfig.Mailjet.SecretKey == "" {
		return nil, fmt.Errorf("Mailjet configuration is missing or invalid")
	}

	// Create webhook status response
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailjet,
		IsRegistered:      false,
		Endpoints:         []domain.WebhookEndpointStatus{},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}

	// Get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, *providerConfig.Mailjet)
	if err != nil {
		return nil, fmt.Errorf("failed to list Mailjet webhooks: %w", err)
	}

	// Look for webhooks that match our integration
	for _, webhook := range existingWebhooks.Data {
		if strings.Contains(webhook.Endpoint, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.Endpoint, fmt.Sprintf("integration_id=%s", integrationID)) {

			status.IsRegistered = true
			status.Endpoints = append(status.Endpoints, domain.WebhookEndpointStatus{
				URL:    webhook.Endpoint,
				Active: webhook.Status == "active",
			})

			// Map event types based on webhook.EventType
			var registeredEvents []domain.EmailEventType
			switch domain.MailjetWebhookEventType(webhook.EventType) {
			case domain.MailjetEventSent:
				registeredEvents = append(registeredEvents, domain.EmailEventDelivered)
			case domain.MailjetEventBounce, domain.MailjetEventBlocked:
				registeredEvents = append(registeredEvents, domain.EmailEventBounce)
			case domain.MailjetEventSpam:
				registeredEvents = append(registeredEvents, domain.EmailEventComplaint)
			}

			status.RegisteredEvents = registeredEvents
			status.ProviderDetails["webhook_id"] = webhook.ID
			break
		}
	}

	return status, nil
}

// UnregisterWebhooks implements the domain.WebhookProvider interface for Mailjet
func (s *MailjetService) UnregisterWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.EmailProvider,
) error {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.Mailjet == nil ||
		providerConfig.Mailjet.APIKey == "" || providerConfig.Mailjet.SecretKey == "" {
		return fmt.Errorf("Mailjet configuration is missing or invalid")
	}

	// Get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, *providerConfig.Mailjet)
	if err != nil {
		return fmt.Errorf("failed to list Mailjet webhooks: %w", err)
	}

	// Delete webhooks that match our criteria
	var lastError error
	for _, webhook := range existingWebhooks.Data {
		if strings.Contains(webhook.Endpoint, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.Endpoint, fmt.Sprintf("integration_id=%s", integrationID)) {

			err := s.DeleteWebhook(ctx, *providerConfig.Mailjet, webhook.ID)
			if err != nil {
				s.logger.WithField("webhook_id", webhook.ID).
					Error(fmt.Sprintf("Failed to delete Mailjet webhook: %v", err))
				lastError = err
				// Continue deleting other webhooks even if one fails
			} else {
				s.logger.WithField("webhook_id", webhook.ID).
					Info("Successfully deleted Mailjet webhook")
			}
		}
	}

	if lastError != nil {
		return fmt.Errorf("failed to delete one or more Mailjet webhooks: %w", lastError)
	}

	return nil
}
