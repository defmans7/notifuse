package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	var events []*domain.WebhookEvent

	switch integration.EmailProvider.Kind {
	case domain.EmailProviderKindSES:
		events, err = s.processSESWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindPostmark:
		events, err = s.processPostmarkWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindMailgun:
		events, err = s.processMailgunWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindSparkPost:
		events, err = s.processSparkPostWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindMailjet:
		events, err = s.processMailjetWebhook(integration.ID, rawPayload)
	case domain.EmailProviderKindSMTP:
		events, err = s.processSMTPWebhook(integration.ID, rawPayload)
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
	if err := s.repo.StoreEvents(ctx, workspaceID, events); err != nil {
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return fmt.Errorf("failed to store webhook events: %w", err)
	}

	updates := []domain.MessageStatusUpdate{}

	for _, event := range events {
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

			updates = append(updates, domain.MessageStatusUpdate{
				ID:        event.MessageID,
				Status:    status,
				Timestamp: event.Timestamp,
			})
		}
	}

	if err := s.messageHistoryRepo.SetStatusesIfNotSet(ctx, workspaceID, updates); err != nil {
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return fmt.Errorf("failed to update message status: %w", err)
	}

	return nil
}

// processSESWebhook processes a webhook event from Amazon SES
func (s *WebhookEventService) processSESWebhook(integrationID string, rawPayload []byte) (events []*domain.WebhookEvent, err error) {

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
		event.BounceType = bounceType
		event.BounceCategory = bounceCategory
		event.BounceDiagnostic = bounceDiagnostic
	} else if eventType == domain.EmailEventComplaint {
		event.ComplaintFeedbackType = complaintFeedbackType
	}

	// Check for transactional_id and broadcast_id in tags and validate as UUID
	// These must be valid UUIDs or empty strings to avoid database errors
	var tags map[string]string

	// Get the tags from the appropriate notification based on event type
	if eventType == domain.EmailEventBounce && len(bounceNotification.Mail.Tags) > 0 {
		tags = bounceNotification.Mail.Tags
	} else if eventType == domain.EmailEventComplaint && err == nil { // err is nil when complaintNotification was parsed successfully
		var complaintNotification domain.SESComplaintNotification
		if err := json.Unmarshal(messageBytes, &complaintNotification); err == nil && complaintNotification.NotificationType == "Complaint" {
			tags = complaintNotification.Mail.Tags
		}
	} else if eventType == domain.EmailEventDelivered && err == nil { // err is nil when deliveryNotification was parsed successfully
		var deliveryNotification domain.SESDeliveryNotification
		if err := json.Unmarshal(messageBytes, &deliveryNotification); err == nil && deliveryNotification.NotificationType == "Delivery" {
			tags = deliveryNotification.Mail.Tags
		}
	}

	// Process the tags if we have any
	if len(tags) > 0 {
		if id, ok := tags["transactional_id"]; ok && id != "" {
			event.TransactionalID = id
		}

		if id, ok := tags["broadcast_id"]; ok && id != "" {
			event.BroadcastID = id
		}
	}

	return []*domain.WebhookEvent{event}, nil
}

// processPostmarkWebhook processes a webhook event from Postmark
func (s *WebhookEventService) processPostmarkWebhook(integrationID string, rawPayload []byte) (events []*domain.WebhookEvent, err error) {

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
		event.BounceType = bounceType
		event.BounceCategory = bounceCategory
		event.BounceDiagnostic = bounceDiagnostic
	} else if eventType == domain.EmailEventComplaint {
		event.ComplaintFeedbackType = complaintFeedbackType
	}

	// Validate TransactionalID and BroadcastID fields from metadata
	// These must be valid UUIDs or empty strings to avoid database errors
	if payload.Metadata != nil {
		if id, ok := payload.Metadata["transactional_id"]; ok && id != "" {
			event.TransactionalID = id
		}

		if id, ok := payload.Metadata["broadcast_id"]; ok && id != "" {
			event.BroadcastID = id
		}
	}

	return []*domain.WebhookEvent{event}, nil
}

// processMailgunWebhook processes a webhook event from Mailgun
func (s *WebhookEventService) processMailgunWebhook(integrationID string, rawPayload []byte) (events []*domain.WebhookEvent, err error) {

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
		event.BounceType = bounceType
		event.BounceCategory = bounceCategory
		event.BounceDiagnostic = bounceDiagnostic
	} else if eventType == domain.EmailEventComplaint {
		event.ComplaintFeedbackType = complaintFeedbackType
	}

	// Validate TransactionalID and BroadcastID from user variables
	// These must be valid UUIDs or empty strings to avoid database errors
	if eventData, ok := jsonData["event-data"].(map[string]interface{}); ok {
		if userVariables, ok := eventData["user-variables"].(map[string]interface{}); ok {
			if id, ok := userVariables["transactional_id"]; ok && id != nil {
				event.TransactionalID = fmt.Sprintf("%v", id)
			}

			if id, ok := userVariables["broadcast_id"]; ok && id != nil {
				event.BroadcastID = fmt.Sprintf("%v", id)
			}
		}
	}

	return []*domain.WebhookEvent{event}, nil
}

// processSparkPostWebhook processes a webhook event from SparkPost
func (s *WebhookEventService) processSparkPostWebhook(integrationID string, rawPayload []byte) (events []*domain.WebhookEvent, err error) {
	events = []*domain.WebhookEvent{}

	// payload can contain multiple events
	var payload []*domain.SparkPostWebhookPayload

	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SparkPost webhook payload: %w", err)
	}

	for _, payload := range payload {
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

		// Parse timestamp - SparkPost may send Unix timestamp as a string
		if payload.MSys.MessageEvent.Timestamp != "" {
			// First try parsing as RFC3339
			if t, err := time.Parse(time.RFC3339, payload.MSys.MessageEvent.Timestamp); err == nil {
				timestamp = t
			} else {
				// If RFC3339 parsing fails, try parsing as Unix timestamp
				if unixTimestamp, err := strconv.ParseInt(payload.MSys.MessageEvent.Timestamp, 10, 64); err == nil {
					timestamp = time.Unix(unixTimestamp, 0)
				} else {
					// Fall back to current time if parsing fails
					timestamp = time.Now()
					s.logger.WithFields(map[string]interface{}{
						"timestamp_string": payload.MSys.MessageEvent.Timestamp,
						"parse_error":      err.Error(),
					}).Warn("Failed to parse SparkPost timestamp")
				}
			}
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
			event.BounceType = bounceType
			event.BounceCategory = bounceCategory
			event.BounceDiagnostic = bounceDiagnostic
		} else if eventType == domain.EmailEventComplaint {
			event.ComplaintFeedbackType = complaintFeedbackType
		}

		events = append(events, event)
	}

	return events, nil
}

// processMailjetWebhook processes a webhook event from Mailjet
func (s *WebhookEventService) processMailjetWebhook(integrationID string, rawPayload []byte) (events []*domain.WebhookEvent, err error) {

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
		event.BounceType = bounceType
		event.BounceCategory = bounceCategory
		event.BounceDiagnostic = bounceDiagnostic
	} else if eventType == domain.EmailEventComplaint {
		event.ComplaintFeedbackType = complaintFeedbackType
	}

	// Validate TransactionalID and BroadcastID
	// These must be valid UUIDs or empty strings to avoid database errors
	// Mailjet allows custom IDs to be passed in the Payload field, which might be JSON
	if payload.Payload != "" {
		// Try to parse Payload as JSON
		var customData map[string]interface{}
		if err := json.Unmarshal([]byte(payload.Payload), &customData); err == nil {
			if id, ok := customData["transactional_id"]; ok && id != nil {
				event.TransactionalID = fmt.Sprintf("%v", id)
			}

			if id, ok := customData["broadcast_id"]; ok && id != nil {
				event.BroadcastID = fmt.Sprintf("%v", id)
			}
		}
	}

	return []*domain.WebhookEvent{event}, nil
}

// processSMTPWebhook processes a webhook event from a generic SMTP provider
func (s *WebhookEventService) processSMTPWebhook(integrationID string, rawPayload []byte) (events []*domain.WebhookEvent, err error) {

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
		event.BounceType = "Bounce"
		event.BounceCategory = payload.BounceCategory
		event.BounceDiagnostic = payload.DiagnosticCode
	} else if eventType == domain.EmailEventComplaint {
		event.ComplaintFeedbackType = payload.ComplaintType
	}

	return []*domain.WebhookEvent{event}, nil
}

// ListEvents retrieves all webhook events for a workspace
func (s *WebhookEventService) ListEvents(ctx context.Context, workspaceID string, params domain.WebhookEventListParams) (*domain.WebhookEventListResult, error) {
	// codecov:ignore:start
	ctx, span := tracing.StartServiceSpan(ctx, "WebhookEventService", "ListEvents")
	defer tracing.EndSpan(span, nil)
	tracing.AddAttribute(ctx, "workspaceID", workspaceID)
	// codecov:ignore:end

	// Authenticate user for workspace
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate params
	if err := params.Validate(); err != nil {
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Call repository method
	result, err := s.repo.ListEvents(ctx, workspaceID, params)
	if err != nil {
		s.logger.WithField("workspace_id", workspaceID).
			WithField("params", params).
			Error(fmt.Sprintf("Failed to list webhook events: %v", err))
		// codecov:ignore:start
		tracing.MarkSpanError(ctx, err)
		// codecov:ignore:end
		return nil, fmt.Errorf("failed to list webhook events: %w", err)
	}

	return result, nil
}
