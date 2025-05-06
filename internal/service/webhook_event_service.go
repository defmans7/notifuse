package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/google/uuid"
)

// WebhookEventService implements the domain.WebhookEventServiceInterface
type WebhookEventService struct {
	repo          domain.WebhookEventRepository
	authService   domain.AuthService
	logger        logger.Logger
	workspaceRepo domain.WorkspaceRepository
}

// NewWebhookEventService creates a new WebhookEventService
func NewWebhookEventService(repo domain.WebhookEventRepository, authService domain.AuthService, logger logger.Logger, workspaceRepo domain.WorkspaceRepository) *WebhookEventService {
	return &WebhookEventService{
		repo:          repo,
		authService:   authService,
		logger:        logger,
		workspaceRepo: workspaceRepo,
	}
}

// ProcessWebhook processes a webhook event from an email provider
func (s *WebhookEventService) ProcessWebhook(ctx context.Context, workspaceID string, integrationID string, rawPayload []byte) error {
	// TODO get workspace and integration
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
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
		return fmt.Errorf("unsupported email provider kind: %s", integration.EmailProvider.Kind)
	}

	if err != nil {
		return fmt.Errorf("failed to process webhook: %w", err)
	}

	// Store the event
	// No authentication needed for webhook events as they come from external providers
	if err := s.repo.StoreEvent(ctx, event); err != nil {
		s.logger.WithField("event_id", event.ID).
			WithField("event_type", event.Type).
			WithField("provider", event.EmailProviderKind).
			Error(fmt.Sprintf("Failed to store webhook event: %v", err))
		return fmt.Errorf("failed to store webhook event: %w", err)
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
	var payload domain.PostmarkWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Postmark webhook payload: %w", err)
	}

	var eventType domain.EmailEventType
	var recipientEmail, messageID string
	var bounceType, bounceCategory, bounceDiagnostic, complaintFeedbackType string
	var timestamp time.Time

	// Determine the event type based on RecordType
	switch payload.RecordType {
	case "Delivery":
		eventType = domain.EmailEventDelivered
		if payload.DeliveredFields != nil {
			recipientEmail = payload.DeliveredFields.RecipientEmail

			// Parse timestamp
			if t, err := time.Parse(time.RFC3339, payload.DeliveredFields.DeliveredAt); err == nil {
				timestamp = t
			} else {
				timestamp = time.Now()
			}
		}
	case "Bounce":
		eventType = domain.EmailEventBounce
		if payload.BounceFields != nil {
			recipientEmail = payload.BounceFields.RecipientEmail
			bounceType = payload.BounceFields.Type
			// Map TypeCode to a category
			switch payload.BounceFields.TypeCode {
			case 1:
				bounceCategory = "HardBounce"
			case 2:
				bounceCategory = "Transient"
			case 3:
				bounceCategory = "Unsubscribe"
			case 4:
				bounceCategory = "Subscribe"
			case 5:
				bounceCategory = "AutoResponder"
			case 6:
				bounceCategory = "AddressChange"
			case 7:
				bounceCategory = "DnsError"
			case 8:
				bounceCategory = "SpamNotification"
			case 9:
				bounceCategory = "OpenRelayTest"
			case 10:
				bounceCategory = "Unknown"
			case 11:
				bounceCategory = "SoftBounce"
			case 12:
				bounceCategory = "VirusNotification"
			case 13:
				bounceCategory = "ChallengeVerification"
			default:
				bounceCategory = "Unknown"
			}
			bounceDiagnostic = payload.BounceFields.Details

			// Parse timestamp
			if t, err := time.Parse(time.RFC3339, payload.BounceFields.BouncedAt); err == nil {
				timestamp = t
			} else {
				timestamp = time.Now()
			}
		}
	case "SpamComplaint":
		eventType = domain.EmailEventComplaint
		if payload.ComplaintFields != nil {
			recipientEmail = payload.ComplaintFields.RecipientEmail
			complaintFeedbackType = payload.ComplaintFields.Type

			// Parse timestamp
			if t, err := time.Parse(time.RFC3339, payload.ComplaintFields.ComplainedAt); err == nil {
				timestamp = t
			} else {
				timestamp = time.Now()
			}
		}
	default:
		return nil, fmt.Errorf("unsupported Postmark record type: %s", payload.RecordType)
	}

	messageID = payload.MessageID

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
	var payload domain.MailgunWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Mailgun webhook payload: %w", err)
	}

	var eventType domain.EmailEventType
	var recipientEmail, messageID string
	var bounceType, bounceCategory, bounceDiagnostic, complaintFeedbackType string
	var timestamp time.Time

	// Set timestamp from event data
	timestamp = time.Unix(int64(payload.EventData.Timestamp), 0)

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

	// Check which event type is present in the payload
	if delivery := payload.MSys.DeliveryEvent; delivery != nil {
		eventType = domain.EmailEventDelivered
		recipientEmail = delivery.RecipientTo
		messageID = delivery.MessageID

		if t, err := time.Parse(time.RFC3339, delivery.Timestamp); err == nil {
			timestamp = t
		} else {
			timestamp = time.Now()
		}
	} else if bounce := payload.MSys.BounceEvent; bounce != nil {
		eventType = domain.EmailEventBounce
		recipientEmail = bounce.RecipientTo
		messageID = bounce.MessageID

		bounceType = "Bounce"
		bounceCategory = bounce.BounceClass
		bounceDiagnostic = bounce.Reason

		if t, err := time.Parse(time.RFC3339, bounce.Timestamp); err == nil {
			timestamp = t
		} else {
			timestamp = time.Now()
		}
	} else if complaint := payload.MSys.SpamComplaint; complaint != nil {
		eventType = domain.EmailEventComplaint
		recipientEmail = complaint.RecipientTo
		messageID = complaint.MessageID

		complaintFeedbackType = complaint.FeedbackType

		if t, err := time.Parse(time.RFC3339, complaint.Timestamp); err == nil {
			timestamp = t
		} else {
			timestamp = time.Now()
		}
	} else {
		return nil, fmt.Errorf("no supported event type found in SparkPost webhook")
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

	// Set timestamp from Unix timestamp
	timestamp = time.Unix(payload.Time, 0)

	// Convert message ID to string
	messageID = fmt.Sprintf("%d", payload.MessageID)
	recipientEmail = payload.Email

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
