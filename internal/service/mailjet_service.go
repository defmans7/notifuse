package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
func (s *MailjetService) ListWebhooks(ctx context.Context, config domain.MailjetConfig) (*domain.MailjetWebhookResponse, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mailjet.com/v3"
	}

	apiURL := fmt.Sprintf("%s/eventcallback", baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
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
func (s *MailjetService) CreateWebhook(ctx context.Context, config domain.MailjetConfig, webhook domain.MailjetWebhook) (*domain.MailjetWebhook, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mailjet.com/v3"
	}

	apiURL := fmt.Sprintf("%s/eventcallback", baseURL)

	// Prepare the request body
	requestBody, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(requestBody))
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
func (s *MailjetService) GetWebhook(ctx context.Context, config domain.MailjetConfig, webhookID int64) (*domain.MailjetWebhook, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mailjet.com/v3"
	}

	apiURL := fmt.Sprintf("%s/eventcallback/%d", baseURL, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
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
func (s *MailjetService) UpdateWebhook(ctx context.Context, config domain.MailjetConfig, webhookID int64, webhook domain.MailjetWebhook) (*domain.MailjetWebhook, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Ensure the webhook ID in the URL matches the one in the body
	webhook.ID = webhookID

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mailjet.com/v3"
	}

	apiURL := fmt.Sprintf("%s/eventcallback/%d", baseURL, webhookID)

	// Prepare the request body
	requestBody, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, bytes.NewBuffer(requestBody))
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
func (s *MailjetService) DeleteWebhook(ctx context.Context, config domain.MailjetConfig, webhookID int64) error {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Log webhook ID for debugging
	webhookIDStr := strconv.FormatInt(webhookID, 10)
	s.logger = s.logger.WithField("webhook_id", webhookIDStr)

	// Construct the API URL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mailjet.com/v3"
	}

	apiURL := fmt.Sprintf("%s/eventcallback/%d", baseURL, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, apiURL, nil)
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
