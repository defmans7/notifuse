package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/Notifuse/notifuse/pkg/tracing"
	"github.com/google/uuid"
)

// WebhookEventService implements the domain.WebhookEventServiceInterface
type WebhookEventService struct {
	repo               domain.WebhookEventRepository
	authService        domain.AuthService
	logger             logger.Logger
	workspaceRepo      domain.WorkspaceRepository
	messageHistoryRepo domain.MessageHistoryRepository
}

// NewWebhookEventService creates a new WebhookEventService
func NewWebhookEventService(
	repo domain.WebhookEventRepository,
	authService domain.AuthService,
	logger logger.Logger,
	workspaceRepo domain.WorkspaceRepository,
	messageHistoryRepo domain.MessageHistoryRepository,
) *WebhookEventService {
	return &WebhookEventService{
		repo:               repo,
		authService:        authService,
		logger:             logger,
		workspaceRepo:      workspaceRepo,
		messageHistoryRepo: messageHistoryRepo,
	}
}

// ProcessWebhook processes a webhook event from an email provider
func (s *WebhookEventService) ProcessWebhook(ctx context.Context, workspaceID string, integrationID string, rawPayload []byte) error {
	// codecov:ignore:start
	ctx, span := tracing.StartServiceSpan(ctx, "WebhookEventService", "ProcessWebhook")
	defer tracing.EndSpan(span, nil)
	tracing.AddAttribute(ctx, "workspaceID", workspaceID)
	tracing.AddAttribute(ctx, "integrationID", integrationID)
	// codecov:ignore:end

	// get workspace and integration
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return fmt.Errorf("failed to get workspace: %w", err)
	}
	var integration domain.Integration
	for _, i := range workspace.Integrations {
		if i.ID == integrationID {
			integration = i
			break
		}
	}
	var event *domain.WebhookEvent
	switch integration.EmailProvider.Kind {
	case domain.EmailProviderKindSES:
		event, err = s.processSESWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindPostmark:
		event, err = s.processPostmarkWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindMailgun:
		event, err = s.processMailgunWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindSparkPost:
		event, err = s.processSparkPostWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindMailjet:
		event, err = s.processMailjetWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindSMTP:
		event, err = s.processSMTPWebhook(integration.ID, rawPayload)
	default:
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, fmt.Errorf("unsupported email provider kind: %s", integration.EmailProvider.Kind))
		// codecov:ignore:end
		return fmt.Errorf("unsupported email provider kind: %s", integration.EmailProvider.Kind)
	}

	if err != nil {
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return fmt.Errorf("failed to process webhook: %w", err)
	}

	// Store the event
	// No authentication needed for webhook events as they come from external providers
	if err := s.repo.StoreEvent(ctx, event); err != nil {
		s.logger.WithField("event_id", event.ID).
			WithField("event_type", event.Type).
			WithField("provider", event.EmailProviderKind).
			Error(fmt.Sprintf("Failed to store webhook event: %v", err))
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return fmt.Errorf("failed to store webhook event: %w", err)
	}

	// Update message history status if we have a message ID
	if event.MessageID != "" {
		var status domain.MessageStatus
		switch event.Type {
		case domain.EmailEventDelivered:
			status = domain.MessageStatusDelivered
		case domain.EmailEventBounce:
			status = domain.MessageStatusBounced
		case domain.EmailEventComplaint:
			status = domain.MessageStatusComplained
		default:
			// Skip other event types
			return nil
		}

		// Update the message status with the timestamp from the event
		err = s.messageHistoryRepo.SetStatusIfNotSet(ctx, workspaceID, event.MessageID, status, event.Timestamp)
		if err != nil {
			s.logger.WithField("event_id", event.ID).
				WithField("message_id", event.MessageID).
				WithField("status", status).
				Error(fmt.Sprintf("Failed to update message status: %v", err))
			// We don't fail the webhook processing if status update fails
			// Just log the error and continue
		}
	}

	return nil
}

// GetEventByID retrieves a webhook event by its ID
func (s *WebhookEventService) GetEventByID(ctx context.Context, id string) (*domain.WebhookEvent, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		s.logger.WithField("event_id", id).
			Error(fmt.Sprintf("Failed to get webhook event: %v", err))
		return nil, err
	}
	return event, nil
}

// GetEventsByType retrieves webhook events by type for a workspace
func (s *WebhookEventService) GetEventsByType(ctx context.Context, workspaceID string, eventType domain.EmailEventType, limit, offset int) ([]*domain.WebhookEvent, error) {
	// Authenticate user for workspace
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	events, err := s.repo.GetEventsByType(ctx, workspaceID, eventType, limit, offset)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).
			WithField("event_type", eventType).
			Error(fmt.Sprintf("Failed to get webhook events by type: %v", err))
		return nil, err
	}
	return events, nil
}

// GetEventsByMessageID retrieves all webhook events associated with a message ID
func (s *WebhookEventService) GetEventsByMessageID(ctx context.Context, messageID string, limit, offset int) ([]*domain.WebhookEvent, error) {
	events, err := s.repo.GetEventsByMessageID(ctx, messageID, limit, offset)
	if err != nil {
		s.logger.WithField("message_id", messageID).
			Error(fmt.Sprintf("Failed to get webhook events by message ID: %v", err))
		return nil, err
	}
	return events, nil
}

// GetEventsByTransactionalID retrieves all webhook events associated with a transactional ID
func (s *WebhookEventService) GetEventsByTransactionalID(ctx context.Context, transactionalID string, limit, offset int) ([]*domain.WebhookEvent, error) {
	events, err := s.repo.GetEventsByTransactionalID(ctx, transactionalID, limit, offset)
	if err != nil {
		s.logger.WithField("transactional_id", transactionalID).
			Error(fmt.Sprintf("Failed to get webhook events by transactional ID: %v", err))
		return nil, err
	}
	return events, nil
}

// GetEventsByBroadcastID retrieves all webhook events associated with a broadcast ID
func (s *WebhookEventService) GetEventsByBroadcastID(ctx context.Context, broadcastID string, limit, offset int) ([]*domain.WebhookEvent, error) {
	events, err := s.repo.GetEventsByBroadcastID(ctx, broadcastID, limit, offset)
	if err != nil {
		s.logger.WithField("broadcast_id", broadcastID).
			Error(fmt.Sprintf("Failed to get webhook events by broadcast ID: %v", err))
		return nil, err
	}
	return events, nil
}

// GetEventCount retrieves the count of events by type for a workspace
func (s *WebhookEventService) GetEventCount(ctx context.Context, workspaceID string, eventType domain.EmailEventType) (int, error) {
	// Authenticate user for workspace
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return 0, fmt.Errorf("failed to authenticate user: %w", err)
	}

	count, err := s.repo.GetEventCount(ctx, workspaceID, eventType)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).
			WithField("event_type", eventType).
			Error(fmt.Sprintf("Failed to get webhook event count: %v", err))
		return 0, err
	}
	return count, nil
}

// processSESWebhook processes a webhook event from Amazon SES
func (s *WebhookEventService) processSESWebhook(integrationID string, rawPayload []byte) (*domain.WebhookEvent, error) {
	// First, parse the SNS message wrapper
	var snsPayload domain.SESWebhookPayload
	if err := json.Unmarshal(rawPayload, &snsPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SES webhook payload: %w", err)
	}

	// Then, parse the actual notification based on the message type
	messageBytes := []byte(snsPayload.Message)

	// Determine the type of notification
	var eventType domain.EmailEventType
	var recipientEmail, messageID string
	var bounceType, bounceCategory, bounceDiagnostic, complaintFeedbackType string
	var timestamp time.Time
	var notifuseMessageID string

	// Try to unmarshal as bounce notification
	var bounceNotification domain.SESBounceNotification
	if err := json.Unmarshal(messageBytes, &bounceNotification); err == nil && bounceNotification.NotificationType == "Bounce" {
		eventType = domain.EmailEventBounce
		if len(bounceNotification.Bounce.BouncedRecipients) > 0 {
			recipientEmail = bounceNotification.Bounce.BouncedRecipients[0].EmailAddress
			bounceDiagnostic = bounceNotification.Bounce.BouncedRecipients[0].DiagnosticCode
		}
		messageID = bounceNotification.Mail.MessageID
		bounceType = bounceNotification.Bounce.BounceType
		bounceCategory = bounceNotification.Bounce.BounceSubType

		// Check for notifuse_message_id in tags
		if len(bounceNotification.Mail.Tags) > 0 {
			if id, ok := bounceNotification.Mail.Tags["notifuse_message_id"]; ok {
				notifuseMessageID = id
			}
		}

		// Parse timestamp
		if t, err := time.Parse(time.RFC3339, bounceNotification.Bounce.Timestamp); err == nil {
			timestamp = t
		} else {
			timestamp = time.Now()
		}
	} else {
		// Try to unmarshal as complaint notification
		var complaintNotification domain.SESComplaintNotification
		if err := json.Unmarshal(messageBytes, &complaintNotification); err == nil && complaintNotification.NotificationType == "Complaint" {
			eventType = domain.EmailEventComplaint
			if len(complaintNotification.Complaint.ComplainedRecipients) > 0 {
				recipientEmail = complaintNotification.Complaint.ComplainedRecipients[0].EmailAddress
			}
			messageID = complaintNotification.Mail.MessageID
			complaintFeedbackType = complaintNotification.Complaint.ComplaintFeedbackType

			// Check for notifuse_message_id in tags
			if len(complaintNotification.Mail.Tags) > 0 {
				if id, ok := complaintNotification.Mail.Tags["notifuse_message_id"]; ok {
					notifuseMessageID = id
				}
			}

			// Parse timestamp
			if t, err := time.Parse(time.RFC3339, complaintNotification.Complaint.Timestamp); err == nil {
				timestamp = t
			} else {
				timestamp = time.Now()
			}
		} else {
			// Try to unmarshal as delivery notification
			var deliveryNotification domain.SESDeliveryNotification
			if err := json.Unmarshal(messageBytes, &deliveryNotification); err == nil && deliveryNotification.NotificationType == "Delivery" {
				eventType = domain.EmailEventDelivered
				if len(deliveryNotification.Delivery.Recipients) > 0 {
					recipientEmail = deliveryNotification.Delivery.Recipients[0]
				}
				messageID = deliveryNotification.Mail.MessageID

				// Check for notifuse_message_id in tags
				if len(deliveryNotification.Mail.Tags) > 0 {
					if id, ok := deliveryNotification.Mail.Tags["notifuse_message_id"]; ok {
						notifuseMessageID = id
					}
				}

				// Parse timestamp
				if t, err := time.Parse(time.RFC3339, deliveryNotification.Delivery.Timestamp); err == nil {
					timestamp = t
				} else {
					timestamp = time.Now()
				}
			} else {
				return nil, fmt.Errorf("unrecognized SES notification type")
			}
		}
	}

	// Use notifuseMessageID if available, otherwise fallback to provider's messageID
	if notifuseMessageID != "" {
		messageID = notifuseMessageID
	}

	// Create the webhook event
	event := domain.NewWebhookEvent(
		uuid.New().String(),
		eventType,
		domain.EmailProviderKindSES,
		integrationID,
		recipientEmail,
		messageID,
		timestamp,
		string(rawPayload),
	)

	// Set event-specific information
	if eventType == domain.EmailEventBounce {
		event.SetBounceInfo(bounceType, bounceCategory, bounceDiagnostic)
	} else if eventType == domain.EmailEventComplaint {
		event.SetComplaintInfo(complaintFeedbackType)
	}

	return event, nil
}

// processPostmarkWebhook processes a webhook event from Postmark
func (s *WebhookEventService) processPostmarkWebhook(integrationID string, rawPayload []byte) (*domain.WebhookEvent, error) {
	// First, unmarshal into a map to extract the fields directly
	var jsonData map[string]interface{}
	if err := json.Unmarshal(rawPayload, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Postmark webhook payload: %w", err)
	}

	// Then unmarshal into our struct
	var payload domain.PostmarkWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Postmark webhook payload: %w", err)
	}

	var eventType domain.EmailEventType
	var recipientEmail, messageID string
	var bounceType, bounceCategory, bounceDiagnostic, complaintFeedbackType string
	var timestamp time.Time
	var notifuseMessageID string

	// Check for custom Message-ID header in the Headers field
	if headersData, ok := jsonData["Headers"].([]interface{}); ok {
		for _, header := range headersData {
			if headerMap, ok := header.(map[string]interface{}); ok {
				if name, ok := headerMap["Name"].(string); ok && name == "Message-ID" {
					if value, ok := headerMap["Value"].(string); ok {
						notifuseMessageID = value
					}
				}
			}
		}
	}

	// Determine the event type based on RecordType
	switch payload.RecordType {
	case "Delivery":
		eventType = domain.EmailEventDelivered

		// Extract Delivered fields from the raw JSON
		if deliveryData, ok := jsonData["Recipient"].(string); ok {
			recipientEmail = deliveryData
		}

		if t, ok := jsonData["DeliveredAt"].(string); ok && t != "" {
			if parsedTime, err := time.Parse(time.RFC3339, t); err == nil {
				timestamp = parsedTime
			} else {
				timestamp = time.Now()
			}
		} else {
			timestamp = time.Now()
		}

	case "Bounce":
		eventType = domain.EmailEventBounce

		// Extract Bounce fields from the raw JSON
		if email, ok := jsonData["Email"].(string); ok {
			recipientEmail = email
		}

		if typeStr, ok := jsonData["Type"].(string); ok {
			bounceType = typeStr
			bounceCategory = typeStr // Use the same value for both in Postmark
		}

		if details, ok := jsonData["Details"].(string); ok {
			bounceDiagnostic = details
		}

		if t, ok := jsonData["BouncedAt"].(string); ok && t != "" {
			if parsedTime, err := time.Parse(time.RFC3339, t); err == nil {
				timestamp = parsedTime
			} else {
				timestamp = time.Now()
			}
		} else {
			timestamp = time.Now()
		}

	case "SpamComplaint":
		eventType = domain.EmailEventComplaint

		// Extract Complaint fields from the raw JSON
		if email, ok := jsonData["Email"].(string); ok {
			recipientEmail = email
		}

		if typeStr, ok := jsonData["Type"].(string); ok {
			complaintFeedbackType = typeStr
		}

		if t, ok := jsonData["ComplainedAt"].(string); ok && t != "" {
			if parsedTime, err := time.Parse(time.RFC3339, t); err == nil {
				timestamp = parsedTime
			} else {
				timestamp = time.Now()
			}
		} else {
			timestamp = time.Now()
		}

	default:
		return nil, fmt.Errorf("unsupported Postmark record type: %s", payload.RecordType)
	}

	messageID = payload.MessageID

	// Use notifuseMessageID if available, otherwise fallback to provider's messageID
	if notifuseMessageID != "" {
		messageID = notifuseMessageID
	}

	// Create the webhook event
	event := domain.NewWebhookEvent(
		uuid.New().String(),
		eventType,
		domain.EmailProviderKindPostmark,
		integrationID,
		recipientEmail,
		messageID,
		timestamp,
		string(rawPayload),
	)

	// Set event-specific information
	if eventType == domain.EmailEventBounce {
		event.SetBounceInfo(bounceType, bounceCategory, bounceDiagnostic)
	} else if eventType == domain.EmailEventComplaint {
		event.SetComplaintInfo(complaintFeedbackType)
	}

	return event, nil
}

// processMailgunWebhook processes a webhook event from Mailgun
func (s *WebhookEventService) processMailgunWebhook(integrationID string, rawPayload []byte) (*domain.WebhookEvent, error) {
	// First unmarshal into a map to access all fields
	var jsonData map[string]interface{}
	if err := json.Unmarshal(rawPayload, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Mailgun webhook payload: %w", err)
	}

	var payload domain.MailgunWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Mailgun webhook payload: %w", err)
	}

	var eventType domain.EmailEventType
	var recipientEmail, messageID string
	var bounceType, bounceCategory, bounceDiagnostic, complaintFeedbackType string
	var timestamp time.Time
	var notifuseMessageID string

	// Set timestamp from event data
	timestamp = time.Unix(int64(payload.EventData.Timestamp), 0)

	// Check for notifuse_message_id in the custom variables
	if eventData, ok := jsonData["event-data"].(map[string]interface{}); ok {
		if userVariables, ok := eventData["user-variables"].(map[string]interface{}); ok {
			if id, ok := userVariables["notifuse_message_id"]; ok {
				notifuseMessageID = fmt.Sprintf("%v", id)
			}
		}
	}

	// Map Mailgun event types to our event types
	switch payload.EventData.Event {
	case "delivered":
		eventType = domain.EmailEventDelivered
		recipientEmail = payload.EventData.Recipient
		messageID = payload.EventData.Message.Headers.MessageID
	case "failed":
		eventType = domain.EmailEventBounce
		recipientEmail = payload.EventData.Recipient
		messageID = payload.EventData.Message.Headers.MessageID

		// Set bounce details
		bounceType = "Failed"
		if payload.EventData.Severity == "permanent" {
			bounceCategory = "HardBounce"
		} else {
			bounceCategory = "SoftBounce"
		}
		bounceDiagnostic = payload.EventData.Reason
	case "complained":
		eventType = domain.EmailEventComplaint
		recipientEmail = payload.EventData.Recipient
		messageID = payload.EventData.Message.Headers.MessageID
		complaintFeedbackType = "abuse"
	default:
		return nil, fmt.Errorf("unsupported Mailgun event type: %s", payload.EventData.Event)
	}

	// Use notifuseMessageID if available, otherwise fallback to provider's messageID
	if notifuseMessageID != "" {
		messageID = notifuseMessageID
	}

	// Create the webhook event
	event := domain.NewWebhookEvent(
		uuid.New().String(),
		eventType,
		domain.EmailProviderKindMailgun,
		integrationID,
		recipientEmail,
		messageID,
		timestamp,
		string(rawPayload),
	)

	// Set event-specific information
	if eventType == domain.EmailEventBounce {
		event.SetBounceInfo(bounceType, bounceCategory, bounceDiagnostic)
	} else if eventType == domain.EmailEventComplaint {
		event.SetComplaintInfo(complaintFeedbackType)
	}

	return event, nil
}

// processSparkPostWebhook processes a webhook event from SparkPost
func (s *WebhookEventService) processSparkPostWebhook(integrationID string, rawPayload []byte) (*domain.WebhookEvent, error) {
	var payload domain.SparkPostWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SparkPost webhook payload: %w", err)
	}

	var eventType domain.EmailEventType
	var recipientEmail, messageID string
	var bounceType, bounceCategory, bounceDiagnostic, complaintFeedbackType string
	var timestamp time.Time
	var notifuseMessageID string

	if payload.MSys.MessageEvent == nil {
		return nil, fmt.Errorf("no message_event found in SparkPost webhook payload")
	}

	if id, ok := payload.MSys.MessageEvent.RecipientMeta["notifuse_message_id"]; ok {
		notifuseMessageID = fmt.Sprintf("%v", id)
	}

	// Set common fields
	recipientEmail = payload.MSys.MessageEvent.RecipientTo
	messageID = payload.MSys.MessageEvent.MessageID

	// Parse timestamp
	if t, err := time.Parse(time.RFC3339, payload.MSys.MessageEvent.Timestamp); err == nil {
		timestamp = t
	} else {
		timestamp = time.Now()
	}

	// Determine event type based on the type field
	switch payload.MSys.MessageEvent.Type {
	case "delivery":
		eventType = domain.EmailEventDelivered

	case "bounce":
		eventType = domain.EmailEventBounce
		bounceType = "Bounce"
		bounceCategory = payload.MSys.MessageEvent.BounceClass
		bounceDiagnostic = payload.MSys.MessageEvent.Reason

	case "spam_complaint":
		eventType = domain.EmailEventComplaint
		complaintFeedbackType = payload.MSys.MessageEvent.FeedbackType

	default:
		return nil, fmt.Errorf("unsupported SparkPost event type: %s", payload.MSys.MessageEvent.Type)
	}

	// Use notifuseMessageID if available, otherwise fallback to provider's messageID
	if notifuseMessageID != "" {
		messageID = notifuseMessageID
	}

	// Create the webhook event
	event := domain.NewWebhookEvent(
		uuid.New().String(),
		eventType,
		domain.EmailProviderKindSparkPost,
		integrationID,
		recipientEmail,
		messageID,
		timestamp,
		string(rawPayload),
	)

	// Set event-specific information
	if eventType == domain.EmailEventBounce {
		event.SetBounceInfo(bounceType, bounceCategory, bounceDiagnostic)
	} else if eventType == domain.EmailEventComplaint {
		event.SetComplaintInfo(complaintFeedbackType)
	}

	return event, nil
}

// processMailjetWebhook processes a webhook event from Mailjet
func (s *WebhookEventService) processMailjetWebhook(integrationID string, rawPayload []byte) (*domain.WebhookEvent, error) {
	var payload domain.MailjetWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Mailjet webhook payload: %w", err)
	}

	var eventType domain.EmailEventType
	var recipientEmail, messageID string
	var bounceType, bounceCategory, bounceDiagnostic, complaintFeedbackType string
	var timestamp time.Time
	var notifuseMessageID string

	// Set timestamp from Unix timestamp
	timestamp = time.Unix(payload.Time, 0)

	// Convert message ID to string
	messageID = fmt.Sprintf("%d", payload.MessageID)
	recipientEmail = payload.Email

	// Check for X-MJ-CustomID in the custom variables
	if payload.CustomID != "" {
		notifuseMessageID = payload.CustomID
	}

	// Map Mailjet event types to our event types
	switch payload.Event {
	case "sent":
		eventType = domain.EmailEventDelivered
	case "bounce", "blocked":
		eventType = domain.EmailEventBounce

		// Set bounce details
		if payload.HardBounce {
			bounceType = "HardBounce"
			bounceCategory = "Permanent"
		} else {
			bounceType = "SoftBounce"
			bounceCategory = "Temporary"
		}

		bounceDiagnostic = payload.Comment
		if payload.ErrorCode != "" {
			if bounceDiagnostic != "" {
				bounceDiagnostic += ": "
			}
			bounceDiagnostic += payload.ErrorCode
		}
	case "spam":
		eventType = domain.EmailEventComplaint
		complaintFeedbackType = "abuse"
	default:
		return nil, fmt.Errorf("unsupported Mailjet event type: %s", payload.Event)
	}

	// Use notifuseMessageID if available, otherwise fallback to provider's messageID
	if notifuseMessageID != "" {
		messageID = notifuseMessageID
	}

	// Create the webhook event
	event := domain.NewWebhookEvent(
		uuid.New().String(),
		eventType,
		domain.EmailProviderKindMailjet,
		integrationID,
		recipientEmail,
		messageID,
		timestamp,
		string(rawPayload),
	)

	// Set event-specific information
	if eventType == domain.EmailEventBounce {
		event.SetBounceInfo(bounceType, bounceCategory, bounceDiagnostic)
	} else if eventType == domain.EmailEventComplaint {
		event.SetComplaintInfo(complaintFeedbackType)
	}

	return event, nil
}

// processSMTPWebhook processes a webhook event from a generic SMTP provider
func (s *WebhookEventService) processSMTPWebhook(integrationID string, rawPayload []byte) (*domain.WebhookEvent, error) {
	var payload domain.SMTPWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SMTP webhook payload: %w", err)
	}

	var eventType domain.EmailEventType
	var timestamp time.Time

	// Parse timestamp
	if t, err := time.Parse(time.RFC3339, payload.Timestamp); err == nil {
		timestamp = t
	} else {
		timestamp = time.Now()
	}

	// Map event types
	switch payload.Event {
	case "delivered":
		eventType = domain.EmailEventDelivered
	case "bounce":
		eventType = domain.EmailEventBounce
	case "complaint":
		eventType = domain.EmailEventComplaint
	default:
		return nil, fmt.Errorf("unsupported SMTP event type: %s", payload.Event)
	}

	// Create the webhook event
	event := domain.NewWebhookEvent(
		uuid.New().String(),
		eventType,
		domain.EmailProviderKindSMTP,
		integrationID,
		payload.Recipient,
		payload.MessageID,
		timestamp,
		string(rawPayload),
	)

	// Set event-specific information
	if eventType == domain.EmailEventBounce {
		event.SetBounceInfo("Bounce", payload.BounceCategory, payload.DiagnosticCode)
	} else if eventType == domain.EmailEventComplaint {
		event.SetComplaintInfo(payload.ComplaintType)
	}

	return event, nil
}
