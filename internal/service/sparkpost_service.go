package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// SparkPostService implements the domain.SparkPostServiceInterface
type SparkPostService struct {
	httpClient  domain.HTTPClient
	authService domain.AuthService
	logger      logger.Logger
}

// NewSparkPostService creates a new instance of SparkPostService
func NewSparkPostService(httpClient domain.HTTPClient, authService domain.AuthService, logger logger.Logger) *SparkPostService {
	return &SparkPostService{
		httpClient:  httpClient,
		authService: authService,
		logger:      logger,
	}
}

// ListWebhooks retrieves all registered webhooks
func (s *SparkPostService) ListWebhooks(ctx context.Context, config domain.SparkPostConfig) (*domain.SparkPostWebhookListResponse, error) {

	// Construct the API URL
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.sparkpost.com/api/v1"
	}

	apiURL := fmt.Sprintf("%s/webhooks", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for listing SparkPost webhooks: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// SparkPost uses API key in header
	authHeader := fmt.Sprintf("Bearer %s", config.APIKey)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for listing SparkPost webhooks: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		responseBody := string(body)
		s.logger.Error(fmt.Sprintf("SparkPost API returned non-OK status code %d: %s", resp.StatusCode, responseBody))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response
	var response domain.SparkPostWebhookListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode SparkPost webhook list response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CreateWebhook creates a new webhook
func (s *SparkPostService) CreateWebhook(ctx context.Context, config domain.SparkPostConfig, webhook domain.SparkPostWebhook) (*domain.SparkPostWebhookResponse, error) {

	// Construct the API URL
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.sparkpost.com/api/v1"
	}

	apiURL := fmt.Sprintf("%s/webhooks", baseURL)

	// Prepare the request body
	requestBody, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for creating SparkPost webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// SparkPost uses API key in header
	authHeader := fmt.Sprintf("Bearer %s", config.APIKey)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for creating SparkPost webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("SparkPost API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response to get the created webhook details
	var response domain.SparkPostWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode SparkPost webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// GetWebhook retrieves a webhook by ID
func (s *SparkPostService) GetWebhook(ctx context.Context, config domain.SparkPostConfig, webhookID string) (*domain.SparkPostWebhookResponse, error) {

	// Construct the API URL
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.sparkpost.com/api/v1"
	}

	apiURL := fmt.Sprintf("%s/webhooks/%s", baseURL, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for getting SparkPost webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// SparkPost uses API key in header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for getting SparkPost webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("SparkPost API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response to get the webhook details
	var response domain.SparkPostWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode SparkPost webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// UpdateWebhook updates an existing webhook
func (s *SparkPostService) UpdateWebhook(ctx context.Context, config domain.SparkPostConfig, webhookID string, webhook domain.SparkPostWebhook) (*domain.SparkPostWebhookResponse, error) {

	// Construct the API URL
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.sparkpost.com/api/v1"
	}

	apiURL := fmt.Sprintf("%s/webhooks/%s", baseURL, webhookID)

	// Prepare the request body
	requestBody, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for updating SparkPost webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// SparkPost uses API key in header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for updating SparkPost webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("SparkPost API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	// Parse the response to get the updated webhook details
	var response domain.SparkPostWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode SparkPost webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// DeleteWebhook deletes a webhook by ID
func (s *SparkPostService) DeleteWebhook(ctx context.Context, config domain.SparkPostConfig, webhookID string) error {

	// Log webhook ID for debugging
	s.logger = s.logger.WithField("webhook_id", webhookID)

	// Construct the API URL
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.sparkpost.com/api/v1"
	}

	apiURL := fmt.Sprintf("%s/webhooks/%s", baseURL, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, apiURL, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for deleting SparkPost webhook: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// SparkPost uses API key in header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for deleting SparkPost webhook: %v", err))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("SparkPost API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	return nil
}

// TestWebhook sends a test event to a webhook
func (s *SparkPostService) TestWebhook(ctx context.Context, config domain.SparkPostConfig, webhookID string) error {

	// Log webhook ID for debugging
	s.logger = s.logger.WithField("webhook_id", webhookID)

	// Construct the API URL
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.sparkpost.com/api/v1"
	}

	apiURL := fmt.Sprintf("%s/webhooks/%s/validate", baseURL, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for testing SparkPost webhook: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// SparkPost uses API key in header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for testing SparkPost webhook: %v", err))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("SparkPost API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	return nil
}

// ValidateWebhook validates a webhook's configuration
func (s *SparkPostService) ValidateWebhook(ctx context.Context, config domain.SparkPostConfig, webhook domain.SparkPostWebhook) (bool, error) {

	// Construct the API URL
	baseURL := config.APIEndpoint
	if baseURL == "" {
		baseURL = "https://api.sparkpost.com/api/v1"
	}

	apiURL := fmt.Sprintf("%s/webhooks/validate", baseURL)

	// Prepare the request body with just the target URL to validate
	requestBody := map[string]string{
		"target": webhook.Target,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal validation request: %v", err))
		return false, fmt.Errorf("failed to marshal validation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for validating SparkPost webhook: %v", err))
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// SparkPost uses API key in header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for validating SparkPost webhook: %v", err))
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response to check if the webhook is valid
	var response struct {
		Results struct {
			Valid bool `json:"valid"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode SparkPost webhook validation response: %v", err))
		return false, fmt.Errorf("failed to decode validation response: %w", err)
	}

	return response.Results.Valid, nil
}

// RegisterWebhooks implements the domain.WebhookProvider interface for SparkPost
func (s *SparkPostService) RegisterWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	baseURL string,
	eventTypes []domain.EmailEventType,
	providerConfig *domain.EmailProvider,
) (*domain.WebhookRegistrationStatus, error) {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.SparkPost == nil || providerConfig.SparkPost.APIKey == "" {
		return nil, fmt.Errorf("SparkPost configuration is missing or invalid")
	}

	// Create SparkPost API config
	apiConfig := domain.SparkPostConfig{
		APIKey:      providerConfig.SparkPost.APIKey,
		APIEndpoint: "https://api.sparkpost.com/api/v1",
	}

	// Generate webhook URL that includes workspace_id and integration_id
	webhookURL := domain.GenerateWebhookCallbackURL(baseURL, domain.EmailProviderKindSparkPost, workspaceID, integrationID)

	// Map our event types to SparkPost event types
	sparkpostEvents := []string{}
	for _, eventType := range eventTypes {
		switch eventType {
		case domain.EmailEventDelivered:
			sparkpostEvents = append(sparkpostEvents, "delivery")
		case domain.EmailEventBounce:
			sparkpostEvents = append(sparkpostEvents, "bounce")
		case domain.EmailEventComplaint:
			sparkpostEvents = append(sparkpostEvents, "spam_complaint")
		}
	}

	// First check for existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, apiConfig)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list SparkPost webhooks: %v", err))
		return nil, fmt.Errorf("failed to list SparkPost webhooks: %w", err)
	}

	// Check if we already have a webhook with our URL
	var existingWebhook *domain.SparkPostWebhook
	for _, webhook := range existingWebhooks.Results {
		if strings.Contains(webhook.Target, baseURL) &&
			strings.Contains(webhook.Target, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.Target, fmt.Sprintf("integration_id=%s", integrationID)) {
			existingWebhook = &webhook
			break
		}
	}

	var webhookResponse *domain.SparkPostWebhookResponse
	if existingWebhook != nil {
		// Update the webhook with new events
		existingWebhook.Events = sparkpostEvents
		webhookResponse, err = s.UpdateWebhook(ctx, apiConfig, existingWebhook.ID, *existingWebhook)
		if err != nil {
			return nil, fmt.Errorf("failed to update SparkPost webhook: %w", err)
		}
	} else {
		// Create a new webhook
		newWebhook := domain.SparkPostWebhook{
			Name:     fmt.Sprintf("Notifuse-%s", integrationID),
			Target:   webhookURL,
			Events:   sparkpostEvents,
			Active:   true,
			AuthType: "none",
		}

		webhookResponse, err = s.CreateWebhook(ctx, apiConfig, newWebhook)
		if err != nil {
			return nil, fmt.Errorf("failed to create SparkPost webhook: %w", err)
		}
	}

	// Create webhook registration status
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSparkPost,
		IsRegistered:      true,
		RegisteredEvents:  eventTypes,
		Endpoints: []domain.WebhookEndpointStatus{
			{
				URL:    webhookURL,
				Active: true,
			},
		},
		ProviderDetails: map[string]interface{}{
			"webhook_id":     webhookResponse.Results.ID,
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}

	return status, nil
}

// GetWebhookStatus implements the domain.WebhookProvider interface for SparkPost
func (s *SparkPostService) GetWebhookStatus(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.EmailProvider,
) (*domain.WebhookRegistrationStatus, error) {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.SparkPost == nil || providerConfig.SparkPost.APIKey == "" {
		return nil, fmt.Errorf("SparkPost configuration is missing or invalid")
	}

	// Create SparkPost API config
	apiConfig := domain.SparkPostConfig{
		APIKey:      providerConfig.SparkPost.APIKey,
		APIEndpoint: "https://api.sparkpost.com/api/v1",
	}

	// Check if we should use EU endpoint based on endpoint setting
	if providerConfig.SparkPost.Endpoint != "" &&
		strings.Contains(strings.ToLower(providerConfig.SparkPost.Endpoint), "eu.sparkpost") {
		apiConfig.APIEndpoint = "https://api.eu.sparkpost.com/api/v1"
	}

	// Create webhook status response
	status := &domain.WebhookRegistrationStatus{
		EmailProviderKind: domain.EmailProviderKindSparkPost,
		IsRegistered:      false,
		Endpoints:         []domain.WebhookEndpointStatus{},
		ProviderDetails: map[string]interface{}{
			"integration_id": integrationID,
			"workspace_id":   workspaceID,
		},
	}

	// Get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, apiConfig)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list SparkPost webhooks: %v", err))
		return nil, fmt.Errorf("failed to list SparkPost webhooks: %w", err)
	}

	// Look for webhooks that match our integration
	for _, webhook := range existingWebhooks.Results {
		if strings.Contains(webhook.Target, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.Target, fmt.Sprintf("integration_id=%s", integrationID)) {

			status.IsRegistered = true
			status.Endpoints = append(status.Endpoints, domain.WebhookEndpointStatus{
				URL:    webhook.Target,
				Active: webhook.Active,
			})

			// Map SparkPost events to our event types
			var registeredEvents []domain.EmailEventType
			for _, event := range webhook.Events {
				switch event {
				case "delivery":
					registeredEvents = append(registeredEvents, domain.EmailEventDelivered)
				case "bounce":
					registeredEvents = append(registeredEvents, domain.EmailEventBounce)
				case "spam_complaint":
					registeredEvents = append(registeredEvents, domain.EmailEventComplaint)
				}
			}
			status.RegisteredEvents = registeredEvents
			status.ProviderDetails["webhook_id"] = webhook.ID
			break
		}
	}

	return status, nil
}

// UnregisterWebhooks implements the domain.WebhookProvider interface for SparkPost
func (s *SparkPostService) UnregisterWebhooks(
	ctx context.Context,
	workspaceID string,
	integrationID string,
	providerConfig *domain.EmailProvider,
) error {
	// Validate the provider configuration
	if providerConfig == nil || providerConfig.SparkPost == nil || providerConfig.SparkPost.APIKey == "" {
		return fmt.Errorf("SparkPost configuration is missing or invalid")
	}

	// Create SparkPost API config
	apiConfig := domain.SparkPostConfig{
		APIKey:      providerConfig.SparkPost.APIKey,
		APIEndpoint: "https://api.sparkpost.com/api/v1",
	}

	// Get existing webhooks
	existingWebhooks, err := s.ListWebhooks(ctx, apiConfig)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list SparkPost webhooks: %v", err))
		return fmt.Errorf("failed to list SparkPost webhooks: %w", err)
	}

	// Delete webhooks that match our integration
	var lastError error
	for _, webhook := range existingWebhooks.Results {
		if strings.Contains(webhook.Target, fmt.Sprintf("workspace_id=%s", workspaceID)) &&
			strings.Contains(webhook.Target, fmt.Sprintf("integration_id=%s", integrationID)) {

			err := s.DeleteWebhook(ctx, apiConfig, webhook.ID)
			if err != nil {
				s.logger.WithField("webhook_id", webhook.ID).
					Error(fmt.Sprintf("Failed to delete SparkPost webhook: %v", err))
				lastError = err
				// Continue deleting other webhooks even if one fails
			} else {
				s.logger.WithField("webhook_id", webhook.ID).
					Info("Successfully deleted SparkPost webhook")
			}
		}
	}

	if lastError != nil {
		return fmt.Errorf("failed to delete one or more SparkPost webhooks: %w", lastError)
	}

	return nil
}
