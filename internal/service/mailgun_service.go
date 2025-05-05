package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// MailgunService implements the domain.MailgunServiceInterface
type MailgunService struct {
	httpClient  domain.HTTPClient
	authService domain.AuthService
	logger      logger.Logger
}

// NewMailgunService creates a new instance of MailgunService
func NewMailgunService(httpClient domain.HTTPClient, authService domain.AuthService, logger logger.Logger) *MailgunService {
	return &MailgunService{
		httpClient:  httpClient,
		authService: authService,
		logger:      logger,
	}
}

// ListWebhooks retrieves all registered webhooks for a domain
func (s *MailgunService) ListWebhooks(ctx context.Context, config domain.MailgunSettings) (*domain.MailgunWebhookListResponse, error) {

	// Construct the API URL
	endpoint := ""
	if strings.ToLower(config.Region) == "eu" {
		endpoint = "https://api.eu.mailgun.net/v3"
	} else {
		endpoint = "https://api.mailgun.net/v3"
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks", endpoint, config.Domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for listing Mailgun webhooks: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for listing Mailgun webhooks: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailgun API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		Webhooks map[string]interface{} `json:"webhooks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailgun webhook list response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert the response to our domain model
	result := &domain.MailgunWebhookListResponse{
		Items: []domain.MailgunWebhook{},
		Total: len(response.Webhooks),
	}

	for eventType, webhookData := range response.Webhooks {
		if webhookMap, ok := webhookData.(map[string]interface{}); ok {
			url, _ := webhookMap["url"].(string)
			active := true
			webhook := domain.MailgunWebhook{
				ID:     eventType,
				URL:    url,
				Events: []string{eventType},
				Active: active,
			}
			result.Items = append(result.Items, webhook)
		}
	}

	return result, nil
}

// CreateWebhook creates a new webhook
func (s *MailgunService) CreateWebhook(ctx context.Context, config domain.MailgunSettings, webhook domain.MailgunWebhook) (*domain.MailgunWebhook, error) {

	if len(webhook.Events) == 0 {
		return nil, fmt.Errorf("at least one event type is required")
	}

	// Mailgun API requires a separate call for each event type
	// We'll use the first event type in the list
	eventType := webhook.Events[0]

	// Construct the API URL
	endpoint := ""
	if strings.ToLower(config.Region) == "eu" {
		endpoint = "https://api.eu.mailgun.net/v3"
	} else {
		endpoint = "https://api.mailgun.net/v3"
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks", endpoint, config.Domain)

	// Create the form data
	form := url.Values{}
	form.Add("id", eventType)
	form.Add("url", webhook.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for creating Mailgun webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", config.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for creating Mailgun webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailgun API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		Message string                 `json:"message"`
		Webhook map[string]interface{} `json:"webhook"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailgun webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Return the created webhook
	createdWebhook := &domain.MailgunWebhook{
		ID:     eventType,
		URL:    webhook.URL,
		Events: []string{eventType},
		Active: true,
	}

	return createdWebhook, nil
}

// GetWebhook retrieves a webhook by ID
func (s *MailgunService) GetWebhook(ctx context.Context, config domain.MailgunSettings, webhookID string) (*domain.MailgunWebhook, error) {

	// Construct the API URL
	endpoint := ""
	if strings.ToLower(config.Region) == "eu" {
		endpoint = "https://api.eu.mailgun.net/v3"
	} else {
		endpoint = "https://api.mailgun.net/v3"
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks/%s", endpoint, config.Domain, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for getting Mailgun webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for getting Mailgun webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailgun API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		Webhook map[string]interface{} `json:"webhook"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailgun webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract webhook details
	url, _ := response.Webhook["url"].(string)
	active, _ := response.Webhook["active"].(bool)

	// Return the webhook
	webhook := &domain.MailgunWebhook{
		ID:     webhookID,
		URL:    url,
		Events: []string{webhookID},
		Active: active,
	}

	return webhook, nil
}

// UpdateWebhook updates an existing webhook
func (s *MailgunService) UpdateWebhook(ctx context.Context, config domain.MailgunSettings, webhookID string, webhook domain.MailgunWebhook) (*domain.MailgunWebhook, error) {

	// Construct the API URL
	endpoint := ""
	if strings.ToLower(config.Region) == "eu" {
		endpoint = "https://api.eu.mailgun.net/v3"
	} else {
		endpoint = "https://api.mailgun.net/v3"
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks/%s", endpoint, config.Domain, webhookID)

	// Create the form data
	form := url.Values{}
	form.Add("url", webhook.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for updating Mailgun webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", config.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for updating Mailgun webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailgun API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		Message string                 `json:"message"`
		Webhook map[string]interface{} `json:"webhook"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailgun webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Return the updated webhook
	updatedWebhook := &domain.MailgunWebhook{
		ID:     webhookID,
		URL:    webhook.URL,
		Events: []string{webhookID},
		Active: true,
	}

	return updatedWebhook, nil
}

// DeleteWebhook deletes a webhook by ID
func (s *MailgunService) DeleteWebhook(ctx context.Context, config domain.MailgunSettings, webhookID string) error {

	// Construct the API URL
	endpoint := ""
	if strings.ToLower(config.Region) == "eu" {
		endpoint = "https://api.eu.mailgun.net/v3"
	} else {
		endpoint = "https://api.mailgun.net/v3"
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks/%s", endpoint, config.Domain, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, apiURL, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for deleting Mailgun webhook: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for deleting Mailgun webhook: %v", err))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailgun API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	return nil
}

// TestWebhook sends a test event to a webhook
func (s *MailgunService) TestWebhook(ctx context.Context, config domain.MailgunSettings, webhookID string, eventType string) error {
	// Mailgun doesn't support testing webhooks directly through their API
	// We could potentially simulate a webhook event, but that's beyond the scope
	// of this implementation
	return fmt.Errorf("testing webhooks is not supported by the Mailgun API")
}

// RegisterWebhooks implements the domain.WebhookProvider interface for Mailgun
func (s *MailgunService) RegisterWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.EmailProvider,
) (*domain.WebhookRegistrationStatus, error) {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.Mailgun == nil ||
		providerConfig.Mailgun.APIKey == "" || providerConfig.Mailgun.Domain == "" {
		return nil, fmt.Errorf("Mailgun configuration is missing or invalid")
	}

	// Generate webhook URL that includes workspace_id and integration_id
	webhookURL := domain.GenerateWebhookCallbackURL(baseURL, domain.EmailProviderKindMailgun, workspaceID, integrationID)

	// Map our event types to Mailgun event types
	var registeredEvents []domain.EmailEventType
	mailgunEvents := make(map[string]bool)

	for _, eventType := range eventTypes {
		switch eventType {
		case domain.EmailEventDelivered:
			mailgunEvents["delivered"] = true
			registeredEvents = append(registeredEvents, domain.EmailEventDelivered)
		case domain.EmailEventBounce:
			mailgunEvents["permanent_fail"] = true
			mailgunEvents["temporary_fail"] = true
			registeredEvents = append(registeredEvents, domain.EmailEventBounce)
		case domain.EmailEventComplaint:
			mailgunEvents["complained"] = true
			registeredEvents = append(registeredEvents, domain.EmailEventComplaint)
		}
	}

	// Get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, *providerConfig.Mailgun)
	if err != nil {
		return nil, fmt.Errorf("failed to list Mailgun webhooks: %w", err)
	}

	// Delete existing webhooks that match our criteria
	for _, webhook := range existingWebhooks.Items {
		if strings.Contains(webhook.URL, baseURL) &&
			strings.Contains(webhook.URL, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.URL, fmt.Sprintf("integration_id=%s", integrationID)) {

			err := s.DeleteWebhook(ctx, *providerConfig.Mailgun, webhook.ID)
			if err != nil {
				s.logger.WithField("webhook_id", webhook.ID).
					Error(fmt.Sprintf("Failed to delete Mailgun webhook: %v", err))
				// Continue with other webhooks even if one fails
			}
		}
	}

	// Create a new webhook for each event type
	endpoints := []domain.WebhookEndpointStatus{}
	providerDetails := map[string]interface{}{
		"integration_id": integrationID,
		"workspace_id":   workspaceID,
		"webhook_ids":    []string{},
	}

	for eventType := range mailgunEvents {
		webhookConfig := domain.MailgunWebhook{
			URL:    webhookURL,
			Events: []string{eventType},
			Active: true,
		}

		webhook, err := s.CreateWebhook(ctx, *providerConfig.Mailgun, webhookConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Mailgun webhook for event %s: %w", eventType, err)
		}

		endpoints = append(endpoints, domain.WebhookEndpointStatus{
			URL:       webhookURL,
			EventType: mapMailgunEventType(eventType),
			Active:    webhook.Active,
		})

		// Add webhook ID to provider details
		webhookIDs := providerDetails["webhook_ids"].([]string)
		webhookIDs = append(webhookIDs, webhook.ID)
		providerDetails["webhook_ids"] = webhookIDs
	}

	// Create webhook registration status
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailgun,
		IsRegistered:      len(endpoints) > 0,
		RegisteredEvents:  registeredEvents,
		Endpoints:         endpoints,
		ProviderDetails:   providerDetails,
	}

	return status, nil
}

// GetWebhookStatus implements the domain.WebhookProvider interface for Mailgun
func (s *MailgunService) GetWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.EmailProvider,
) (*domain.WebhookRegistrationStatus, error) {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.Mailgun == nil ||
		providerConfig.Mailgun.APIKey == "" || providerConfig.Mailgun.Domain == "" {
		return nil, fmt.Errorf("Mailgun configuration is missing or invalid")
	}

	// Create webhook status response
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailgun,
		IsRegistered:      false,
		Endpoints:         []domain.WebhookEndpointStatus{},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
			"webhook_ids":    []string{},
		},
	}

	// Get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, *providerConfig.Mailgun)
	if err != nil {
		return nil, fmt.Errorf("failed to list Mailgun webhooks: %w", err)
	}

	// Check for webhooks that match our integration
	registeredEventMap := make(map[domain.EmailEventType]bool)
	webhookIDs := []string{}

	for _, webhook := range existingWebhooks.Items {
		if strings.Contains(webhook.URL, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.URL, fmt.Sprintf("integration_id=%s", integrationID)) {

			status.IsRegistered = true
			webhookIDs = append(webhookIDs, webhook.ID)

			// Add endpoint
			status.Endpoints = append(status.Endpoints, domain.WebhookEndpointStatus{
				URL:       webhook.URL,
				EventType: mapMailgunEventType(webhook.ID), // In Mailgun, the ID is the event type
				Active:    webhook.Active,
			})

			// Track registered event types
			eventType := mapMailgunEventType(webhook.ID)
			if eventType != "" {
				registeredEventMap[eventType] = true
			}
		}
	}

	// Convert registered event map to slice
	var registeredEvents []domain.EmailEventType
	for eventType := range registeredEventMap {
		registeredEvents = append(registeredEvents, eventType)
	}
	status.RegisteredEvents = registeredEvents
	status.ProviderDetails["webhook_ids"] = webhookIDs

	return status, nil
}

// UnregisterWebhooks implements the domain.WebhookProvider interface for Mailgun
func (s *MailgunService) UnregisterWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.EmailProvider,
) error {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.Mailgun == nil ||
		providerConfig.Mailgun.APIKey == "" || providerConfig.Mailgun.Domain == "" {
		return fmt.Errorf("Mailgun configuration is missing or invalid")
	}

	// Get existing webhooks

	existingWebhooks, err := s.ListWebhooks(ctx, *providerConfig.Mailgun)
	if err != nil {
		return fmt.Errorf("failed to list Mailgun webhooks: %w", err)
	}

	// Delete webhooks that match our integration
	var lastError error
	for _, webhook := range existingWebhooks.Items {
		if strings.Contains(webhook.URL, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.URL, fmt.Sprintf("integration_id=%s", integrationID)) {

			err := s.DeleteWebhook(ctx, *providerConfig.Mailgun, webhook.ID)
			if err != nil {
				s.logger.WithField("webhook_id", webhook.ID).
					Error(fmt.Sprintf("Failed to delete Mailgun webhook: %v", err))
				lastError = err
				// Continue deleting other webhooks even if one fails
			} else {
				s.logger.WithField("webhook_id", webhook.ID).
					Info("Successfully deleted Mailgun webhook")
			}
		}
	}

	if lastError != nil {
		return fmt.Errorf("failed to delete one or more Mailgun webhooks: %w", lastError)
	}

	return nil
}

// Helper function to map Mailgun event types to our domain event types
func mapMailgunEventType(eventType string) domain.EmailEventType {
	switch eventType {
	case "delivered":
		return domain.EmailEventDelivered
	case "permanent_fail", "temporary_fail":
		return domain.EmailEventBounce
	case "complained":
		return domain.EmailEventComplaint
	default:
		return ""
	}
}
