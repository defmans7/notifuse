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

// PostmarkService implements the domain.PostmarkServiceInterface
type PostmarkService struct {
	httpClient  domain.HTTPClient
	authService domain.AuthService
	logger      logger.Logger
}

// NewPostmarkService creates a new instance of PostmarkService
func NewPostmarkService(httpClient domain.HTTPClient, authService domain.AuthService, logger logger.Logger) *PostmarkService {
	return &PostmarkService{
		httpClient:  httpClient,
		authService: authService,
		logger:      logger,
	}
}

// ListWebhooks retrieves all registered webhooks
func (s *PostmarkService) ListWebhooks(ctx context.Context, config domain.PostmarkConfig) (*domain.PostmarkListWebhooksResponse, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	url := fmt.Sprintf("%s/webhooks", config.APIEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for listing Postmark webhooks: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", config.ServerToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for listing Postmark webhooks: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Postmark API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	var result domain.PostmarkListWebhooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Postmark webhook list response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// RegisterWebhook registers a new webhook
func (s *PostmarkService) RegisterWebhook(ctx context.Context, config domain.PostmarkConfig, webhook domain.PostmarkWebhookConfig) (*domain.PostmarkWebhookResponse, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	url := fmt.Sprintf("%s/webhooks", config.APIEndpoint)

	jsonData, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for registering Postmark webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", config.ServerToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for registering Postmark webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Postmark API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	var result domain.PostmarkWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Postmark webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UnregisterWebhook removes a webhook by ID
func (s *PostmarkService) UnregisterWebhook(ctx context.Context, config domain.PostmarkConfig, webhookID int) error {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	webhookIDStr := strconv.Itoa(webhookID)
	s.logger = s.logger.WithField("webhook_id", webhookIDStr)

	url := fmt.Sprintf("%s/webhooks/%d", config.APIEndpoint, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for deleting Postmark webhook: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Server-Token", config.ServerToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for deleting Postmark webhook: %v", err))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Postmark API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	return nil
}

// GetWebhook retrieves a specific webhook by ID
func (s *PostmarkService) GetWebhook(ctx context.Context, config domain.PostmarkConfig, webhookID int) (*domain.PostmarkWebhookResponse, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	webhookIDStr := strconv.Itoa(webhookID)
	s.logger = s.logger.WithField("webhook_id", webhookIDStr)

	url := fmt.Sprintf("%s/webhooks/%d", config.APIEndpoint, webhookID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for getting Postmark webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Server-Token", config.ServerToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for getting Postmark webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Postmark API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	var result domain.PostmarkWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Postmark webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateWebhook updates an existing webhook
func (s *PostmarkService) UpdateWebhook(ctx context.Context, config domain.PostmarkConfig, webhookID int, webhook domain.PostmarkWebhookConfig) (*domain.PostmarkWebhookResponse, error) {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	webhookIDStr := strconv.Itoa(webhookID)
	s.logger = s.logger.WithField("webhook_id", webhookIDStr)

	url := fmt.Sprintf("%s/webhooks/%d", config.APIEndpoint, webhookID)

	jsonData, err := json.Marshal(webhook)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal webhook configuration: %v", err))
		return nil, fmt.Errorf("failed to marshal webhook configuration: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for updating Postmark webhook: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", config.ServerToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for updating Postmark webhook: %v", err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Postmark API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	var result domain.PostmarkWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to decode Postmark webhook response: %v", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// TestWebhook sends a test event to the webhook
func (s *PostmarkService) TestWebhook(ctx context.Context, config domain.PostmarkConfig, webhookID int, eventType domain.EmailEventType) error {
	// For this API functionality, we're not tied to a specific workspace
	// But we still need to authenticate the user to ensure proper access control
	// We'll use a placeholder workspace ID
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	webhookIDStr := strconv.Itoa(webhookID)
	s.logger = s.logger.WithField("webhook_id", webhookIDStr)

	// Map our standard event types to Postmark test trigger types
	var triggerName string
	switch eventType {
	case domain.EmailEventDelivered:
		triggerName = "Delivery"
	case domain.EmailEventBounce:
		triggerName = "Bounce"
	case domain.EmailEventComplaint:
		triggerName = "SpamComplaint"
	default:
		return fmt.Errorf("unsupported event type: %s", eventType)
	}

	url := fmt.Sprintf("%s/webhooks/%d/trigger", config.APIEndpoint, webhookID)

	payload := map[string]string{"Trigger": triggerName}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to marshal test trigger payload: %v", err))
		return fmt.Errorf("failed to marshal test trigger payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create request for testing Postmark webhook: %v", err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", config.ServerToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to execute request for testing Postmark webhook: %v", err))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error(fmt.Sprintf("Postmark API returned non-OK status code %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned non-OK status code %d", resp.StatusCode)
	}

	return nil
}
