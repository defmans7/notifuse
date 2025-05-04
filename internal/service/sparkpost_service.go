package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for listing SparkPost webhooks: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("SparkPost API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
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
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
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
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	// For this API functionality, we're not tied to a specific workspace
	// We'll use a placeholder workspace ID for authentication
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to authenticate user: %w", err)
	}

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
