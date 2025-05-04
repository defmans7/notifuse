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
func (s *MailgunService) ListWebhooks(ctx context.Context, config domain.MailgunConfig) (*domain.MailgunWebhookListResponse, error) {
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		// Default to US region if not specified
		if strings.ToLower(config.Region) == "eu" {
			baseURL = "https://api.eu.mailgun.net/v3"
		} else {
			baseURL = "https://api.mailgun.net/v3"
		}
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks", baseURL, config.Domain)
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
func (s *MailgunService) CreateWebhook(ctx context.Context, config domain.MailgunConfig, webhook domain.MailgunWebhook) (*domain.MailgunWebhook, error) {
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	if len(webhook.Events) == 0 {
		return nil, fmt.Errorf("at least one event type is required")
	}

	// Mailgun API requires a separate call for each event type
	// We'll use the first event type in the list
	eventType := webhook.Events[0]

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		// Default to US region if not specified
		if strings.ToLower(config.Region) == "eu" {
			baseURL = "https://api.eu.mailgun.net/v3"
		} else {
			baseURL = "https://api.mailgun.net/v3"
		}
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks", baseURL, config.Domain)

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
func (s *MailgunService) GetWebhook(ctx context.Context, config domain.MailgunConfig, webhookID string) (*domain.MailgunWebhook, error) {
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		// Default to US region if not specified
		if strings.ToLower(config.Region) == "eu" {
			baseURL = "https://api.eu.mailgun.net/v3"
		} else {
			baseURL = "https://api.mailgun.net/v3"
		}
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks/%s", baseURL, config.Domain, webhookID)
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
func (s *MailgunService) UpdateWebhook(ctx context.Context, config domain.MailgunConfig, webhookID string, webhook domain.MailgunWebhook) (*domain.MailgunWebhook, error) {
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		// Default to US region if not specified
		if strings.ToLower(config.Region) == "eu" {
			baseURL = "https://api.eu.mailgun.net/v3"
		} else {
			baseURL = "https://api.mailgun.net/v3"
		}
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks/%s", baseURL, config.Domain, webhookID)

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
func (s *MailgunService) DeleteWebhook(ctx context.Context, config domain.MailgunConfig, webhookID string) error {
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		// Default to US region if not specified
		if strings.ToLower(config.Region) == "eu" {
			baseURL = "https://api.eu.mailgun.net/v3"
		} else {
			baseURL = "https://api.mailgun.net/v3"
		}
	}

	apiURL := fmt.Sprintf("%s/%s/webhooks/%s", baseURL, config.Domain, webhookID)
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
func (s *MailgunService) TestWebhook(ctx context.Context, config domain.MailgunConfig, webhookID string, eventType string) error {
	// Mailgun doesn't support testing webhooks directly through their API
	// We could potentially simulate a webhook event, but that's beyond the scope
	// of this implementation
	return fmt.Errorf("testing webhooks is not supported by the Mailgun API")
}
