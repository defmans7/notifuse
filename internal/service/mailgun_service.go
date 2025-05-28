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
	httpClient      domain.HTTPClient
	authService     domain.AuthService
	logger          logger.Logger
	webhookEndpoint string
}

// NewMailgunService creates a new instance of MailgunService
func NewMailgunService(httpClient domain.HTTPClient, authService domain.AuthService, logger logger.Logger, webhookEndpoint string) *MailgunService {
	return &MailgunService{
		httpClient:      httpClient,
		authService:     authService,
		logger:          logger,
		webhookEndpoint: webhookEndpoint,
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

	// Format according to Mailgun API documentation: https://api.mailgun.net/v3/domains/{domain}/webhooks
	apiURL := fmt.Sprintf("%s/domains/%s/webhooks", endpoint, config.Domain)
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
	var response domain.MailgunWebhookListResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Mailgun webhook list response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// only keep URLs that contain the webhookEndpoint
	if response.Webhooks.Delivered.URLs != nil {
		filteredURLs := []string{}
		for _, url := range response.Webhooks.Delivered.URLs {
			if strings.Contains(url, s.webhookEndpoint) {
				filteredURLs = append(filteredURLs, url)
			}
		}
		response.Webhooks.Delivered.URLs = filteredURLs
	}

	// Filter other event types too
	if response.Webhooks.PermanentFail.URLs != nil {
		filteredURLs := []string{}
		for _, url := range response.Webhooks.PermanentFail.URLs {
			if strings.Contains(url, s.webhookEndpoint) {
				filteredURLs = append(filteredURLs, url)
			}
		}
		response.Webhooks.PermanentFail.URLs = filteredURLs
	}

	if response.Webhooks.TemporaryFail.URLs != nil {
		filteredURLs := []string{}
		for _, url := range response.Webhooks.TemporaryFail.URLs {
			if strings.Contains(url, s.webhookEndpoint) {
				filteredURLs = append(filteredURLs, url)
			}
		}
		response.Webhooks.TemporaryFail.URLs = filteredURLs
	}

	if response.Webhooks.Complained.URLs != nil {
		filteredURLs := []string{}
		for _, url := range response.Webhooks.Complained.URLs {
			if strings.Contains(url, s.webhookEndpoint) {
				filteredURLs = append(filteredURLs, url)
			}
		}
		response.Webhooks.Complained.URLs = filteredURLs
	}
	return &response, nil
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

	apiURL := fmt.Sprintf("%s/domains/%s/webhooks", endpoint, config.Domain)

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

	apiURL := fmt.Sprintf("%s/domains/%s/webhooks/%s", endpoint, config.Domain, webhookID)
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

	apiURL := fmt.Sprintf("%s/domains/%s/webhooks/%s", endpoint, config.Domain, webhookID)

	// Create the form data
	form := url.Values{}
	// The Mailgun API has inconsistent documentation/implementation
	// Use 'urls' parameter instead of 'url' to avoid 405 Method Not Allowed error
	form.Add("urls", webhook.URL)

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

	apiURL := fmt.Sprintf("%s/domains/%s/webhooks/%s", endpoint, config.Domain, webhookID)
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
	// Process delivered webhooks
	for eventType, urls := range map[string]domain.MailgunUrls{
		"delivered":      existingWebhooks.Webhooks.Delivered,
		"permanent_fail": existingWebhooks.Webhooks.PermanentFail,
		"temporary_fail": existingWebhooks.Webhooks.TemporaryFail,
		"complained":     existingWebhooks.Webhooks.Complained,
	} {
		for _, url := range urls.URLs {
			if strings.Contains(url, baseURL) &&
				strings.Contains(url, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
				strings.Contains(url, fmt.Sprintf("integration_id=%s", integrationID)) {

				err := s.DeleteWebhook(ctx, *providerConfig.Mailgun, eventType)
				if err != nil {
					s.logger.WithField("webhook_id", eventType).
						Error(fmt.Sprintf("Failed to delete Mailgun webhook: %v", err))
					// Continue with other webhooks even if one fails
				}
			}
		}
	}

	// Create a new webhook for each event type
	endpoints := []domain.WebhookEndpointStatus{}
	providerDetails := map[string]interface{}{
		"integration_id": integrationID,
		"workspace_id":   workspaceID,
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
			WebhookID: webhook.ID,
			URL:       webhookURL,
			EventType: mapMailgunEventType(eventType),
			Active:    webhook.Active,
		})
	}

	// Create webhook registration status
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindMailgun,
		IsRegistered:      len(endpoints) > 0,
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
		},
	}

	// Get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, *providerConfig.Mailgun)
	if err != nil {
		return nil, fmt.Errorf("failed to list Mailgun webhooks: %w", err)
	}

	// Check for webhooks that match our integration
	registeredEventMap := make(map[domain.EmailEventType]bool)

	// Check for webhooks that match our integration
	for eventType, urls := range map[string]domain.MailgunUrls{
		"delivered":      existingWebhooks.Webhooks.Delivered,
		"permanent_fail": existingWebhooks.Webhooks.PermanentFail,
		"temporary_fail": existingWebhooks.Webhooks.TemporaryFail,
		"complained":     existingWebhooks.Webhooks.Complained,
	} {
		for _, url := range urls.URLs {
			if strings.Contains(url, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
				strings.Contains(url, fmt.Sprintf("integration_id=%s", integrationID)) {

				status.IsRegistered = true

				// Add endpoint
				status.Endpoints = append(status.Endpoints, domain.WebhookEndpointStatus{
					WebhookID: eventType,
					URL:       url,
					EventType: mapMailgunEventType(eventType), // In Mailgun, the ID is the event type
					Active:    true,                           // Assume active if listed
				})

				// Track registered event types
				eventType := mapMailgunEventType(eventType)
				if eventType != "" {
					registeredEventMap[eventType] = true
				}
			}
		}
	}

	// Convert registered event map to slice
	var registeredEvents []domain.EmailEventType
	for eventType := range registeredEventMap {
		registeredEvents = append(registeredEvents, eventType)
	}

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

	// Delete webhooks that match our integration
	for eventType, urls := range map[string]domain.MailgunUrls{
		"delivered":      existingWebhooks.Webhooks.Delivered,
		"permanent_fail": existingWebhooks.Webhooks.PermanentFail,
		"temporary_fail": existingWebhooks.Webhooks.TemporaryFail,
		"complained":     existingWebhooks.Webhooks.Complained,
	} {
		for _, url := range urls.URLs {
			if strings.Contains(url, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
				strings.Contains(url, fmt.Sprintf("integration_id=%s", integrationID)) {

				err := s.DeleteWebhook(ctx, *providerConfig.Mailgun, eventType)
				if err != nil {
					s.logger.WithField("webhook_id", eventType).
						Error(fmt.Sprintf("Failed to delete Mailgun webhook: %v", err))
					lastError = err
					// Continue deleting other webhooks even if one fails
				} else {
					s.logger.WithField("webhook_id", eventType).
						Info("Successfully deleted Mailgun webhook")
				}
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

// SendEmail sends an email using Mailgun
func (s *MailgunService) SendEmail(ctx context.Context, workspaceID string, messageID string, fromAddress, fromName, to, subject, content string, provider *domain.EmailProvider, emailOptions domain.EmailOptions) error {
	if provider.Mailgun == nil {
		return fmt.Errorf("Mailgun provider is not configured")
	}

	// Determine endpoint based on region
	endpoint := ""
	if strings.ToLower(provider.Mailgun.Region) == "eu" {
		endpoint = "https://api.eu.mailgun.net/v3"
	} else {
		endpoint = "https://api.mailgun.net/v3"
	}

	// Format the API URL
	apiURL := fmt.Sprintf("%s/%s/messages", endpoint, provider.Mailgun.Domain)

	// Create the form data for the email
	form := url.Values{}
	form.Add("from", fmt.Sprintf("%s <%s>", fromName, fromAddress))
	form.Add("to", to)
	form.Add("subject", subject)
	form.Add("html", content)

	// Add cc recipients if provided
	for _, ccAddress := range emailOptions.CC {
		if ccAddress != "" {
			form.Add("cc", ccAddress)
		}
	}

	// Add bcc recipients if provided
	for _, bccAddress := range emailOptions.BCC {
		if bccAddress != "" {
			form.Add("bcc", bccAddress)
		}
	}

	// Add reply-to if provided
	if emailOptions.ReplyTo != "" {
		form.Add("h:Reply-To", emailOptions.ReplyTo)
	}

	// Add messageID as a custom variable for tracking
	form.Add("v:notifuse_message_id", messageID)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for sending Mailgun email: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth header
	req.SetBasicAuth("api", provider.Mailgun.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for sending Mailgun email: %v", err))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Mailgun API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	return nil
}
