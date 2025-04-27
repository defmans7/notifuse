package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/google/uuid"
)

// BroadcastService implements the domain.BroadcastService interface
type BroadcastService struct {
	repo        domain.BroadcastRepository
	contactRepo domain.ContactRepository
	emailSvc    domain.EmailServiceInterface
	templateSvc domain.TemplateService
	logger      logger.Logger
}

// NewBroadcastService creates a new broadcast service
func NewBroadcastService(
	repo domain.BroadcastRepository,
	emailSvc domain.EmailServiceInterface,
	logger logger.Logger,
	contactRepo domain.ContactRepository,
	templateSvc domain.TemplateService,
) *BroadcastService {
	return &BroadcastService{
		repo:        repo,
		emailSvc:    emailSvc,
		logger:      logger,
		contactRepo: contactRepo,
		templateSvc: templateSvc,
	}
}

// CreateBroadcast creates a new broadcast
func (s *BroadcastService) CreateBroadcast(ctx context.Context, request *domain.CreateBroadcastRequest) (*domain.Broadcast, error) {
	// Validate the request
	broadcast, err := request.Validate()
	if err != nil {
		s.logger.Error("Failed to validate broadcast creation request")
		return nil, err
	}

	// Generate a unique ID for the broadcast if not provided
	if broadcast.ID == "" {
		// Generate a UUID and trim it to 32 characters to fit VARCHAR(32)
		fullUUID := uuid.New().String()
		broadcast.ID = fullUUID[:8] + fullUUID[9:13] + fullUUID[14:18] + fullUUID[19:23] + fullUUID[24:32]
	}

	// Set default values
	broadcast.Status = domain.BroadcastStatusDraft
	now := time.Now().UTC()
	broadcast.CreatedAt = now
	broadcast.UpdatedAt = now

	// Set scheduled time if needed
	if broadcast.Schedule.IsScheduled && (broadcast.Schedule.ScheduledDate != "" && broadcast.Schedule.ScheduledTime != "") {
		// Set status to scheduled if the broadcast is scheduled
		broadcast.Status = domain.BroadcastStatusScheduled
	}

	// Persist the broadcast
	err = s.repo.CreateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to create broadcast in repository")
		return nil, err
	}

	s.logger.Info("Broadcast created successfully")

	return broadcast, nil
}

// GetBroadcast retrieves a broadcast by ID
func (s *BroadcastService) GetBroadcast(ctx context.Context, workspaceID, id string) (*domain.Broadcast, error) {
	broadcast, err := s.repo.GetBroadcast(ctx, workspaceID, id)
	if err != nil {
		// Just propagate the error, including ErrBroadcastNotFound
		return nil, err
	}
	return broadcast, nil
}

// UpdateBroadcast updates an existing broadcast
func (s *BroadcastService) UpdateBroadcast(ctx context.Context, request *domain.UpdateBroadcastRequest) (*domain.Broadcast, error) {
	// First, get the existing broadcast
	existingBroadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for update")
		return nil, err
	}

	// Validate and update broadcast fields
	updatedBroadcast, err := request.Validate(existingBroadcast)
	if err != nil {
		s.logger.Error("Failed to validate broadcast update request")
		return nil, err
	}

	// Set the updated time
	updatedBroadcast.UpdatedAt = time.Now().UTC()

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, updatedBroadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return nil, err
	}

	s.logger.Info("Broadcast updated successfully")

	return updatedBroadcast, nil
}

// ListBroadcasts retrieves a list of broadcasts
func (s *BroadcastService) ListBroadcasts(ctx context.Context, params domain.ListBroadcastsParams) (*domain.BroadcastListResponse, error) {
	// Apply default values for pagination if not provided
	if params.Limit <= 0 {
		params.Limit = 50 // Default limit
	}
	if params.Limit > 100 {
		params.Limit = 100 // Maximum limit
	}
	if params.Offset < 0 {
		params.Offset = 0 // Ensure offset is not negative
	}

	response, err := s.repo.ListBroadcasts(ctx, params)
	if err != nil {
		s.logger.Error("Failed to list broadcasts from repository")
		return nil, err
	}

	// If WithTemplates is true, fetch template details for each variation
	if params.WithTemplates {
		for _, broadcast := range response.Broadcasts {
			for i, variation := range broadcast.TestSettings.Variations {
				if variation.TemplateID != "" {
					// Fetch the template for this variation
					template, err := s.templateSvc.GetTemplateByID(ctx, params.WorkspaceID, variation.TemplateID, 1)
					if err != nil {
						s.logger.Error("Failed to fetch template for broadcast variation")
						// Continue with the next variation rather than failing the whole request
						continue
					}

					// Assign the template to the variation
					broadcast.SetTemplateForVariation(i, template)
				}
			}
		}
	}

	s.logger.Info("Broadcasts listed successfully")

	return response, nil
}

// ScheduleBroadcast schedules a broadcast for sending
func (s *BroadcastService) ScheduleBroadcast(ctx context.Context, request *domain.ScheduleBroadcastRequest) error {
	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate schedule broadcast request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for scheduling")
		return err
	}

	// Only draft broadcasts can be scheduled
	if broadcast.Status != domain.BroadcastStatusDraft {
		err := fmt.Errorf("only broadcasts with draft status can be scheduled, current status: %s", broadcast.Status)
		s.logger.Error("Cannot schedule broadcast with non-draft status")
		return err
	}

	// Update broadcast status and scheduling info
	broadcast.Status = domain.BroadcastStatusScheduled
	broadcast.UpdatedAt = time.Now().UTC()

	if request.SendNow {
		// If sending immediately, set status to sending
		broadcast.Status = domain.BroadcastStatusSending
		now := time.Now().UTC()
		broadcast.StartedAt = &now
	} else {
		// Update the schedule settings with the requested settings
		broadcast.Schedule.IsScheduled = request.IsScheduled
		broadcast.Schedule.ScheduledDate = request.ScheduledDate
		broadcast.Schedule.ScheduledTime = request.ScheduledTime
		broadcast.Schedule.Timezone = request.Timezone
		broadcast.Schedule.UseRecipientTimezone = request.UseRecipientTimezone

		// Validate that we can parse the scheduled date/time
		scheduledDateTime, err := broadcast.Schedule.ParseScheduledDateTime()
		if err != nil {
			s.logger.Error("Failed to parse scheduled date/time")
			return err
		}

		// Use the scheduledDateTime in a log message to avoid unused variable error
		s.logger.Info(fmt.Sprintf("Broadcast scheduled successfully for %v", scheduledDateTime))
	}

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return err
	}

	if request.SendNow {
	}

	// TODO: If SendNow is true, trigger the actual sending process
	// This would typically involve adding the broadcast to a queue for processing
	// This implementation depends on your message processing architecture

	return nil
}

// PauseBroadcast pauses a sending broadcast
func (s *BroadcastService) PauseBroadcast(ctx context.Context, request *domain.PauseBroadcastRequest) error {
	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate pause broadcast request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for pausing")
		return err
	}

	// Only sending broadcasts can be paused
	if broadcast.Status != domain.BroadcastStatusSending {
		err := fmt.Errorf("only broadcasts with sending status can be paused, current status: %s", broadcast.Status)
		s.logger.Error("Cannot pause broadcast with non-sending status")
		return err
	}

	// Update broadcast status and pause info
	broadcast.Status = domain.BroadcastStatusPaused
	now := time.Now().UTC()
	broadcast.PausedAt = &now
	broadcast.UpdatedAt = now

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return err
	}

	s.logger.Info("Broadcast paused successfully")

	return nil
}

// ResumeBroadcast resumes a paused broadcast
func (s *BroadcastService) ResumeBroadcast(ctx context.Context, request *domain.ResumeBroadcastRequest) error {
	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate resume broadcast request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for resuming")
		return err
	}

	// Only paused broadcasts can be resumed
	if broadcast.Status != domain.BroadcastStatusPaused {
		err := fmt.Errorf("only broadcasts with paused status can be resumed, current status: %s", broadcast.Status)
		s.logger.Error("Cannot resume broadcast with invalid status")
		return err
	}

	// Update broadcast status
	now := time.Now().UTC()
	broadcast.UpdatedAt = now

	// If broadcast was originally scheduled and scheduled time is in the future
	if broadcast.Schedule.IsScheduled {
		scheduledTime, err := broadcast.Schedule.ParseScheduledDateTime()
		isScheduledInFuture := err == nil && scheduledTime.After(now) && broadcast.StartedAt == nil

		if isScheduledInFuture {
			broadcast.Status = domain.BroadcastStatusScheduled
			s.logger.Info("Broadcast resumed to scheduled status")
		} else {
			// If scheduled time has passed or there was an error parsing it
			broadcast.Status = domain.BroadcastStatusSending
			if broadcast.StartedAt == nil {
				broadcast.StartedAt = &now
			}
			s.logger.Info("Broadcast resumed to sending status")
		}
	} else {
		// If broadcast wasn't scheduled, resume sending
		broadcast.Status = domain.BroadcastStatusSending
		if broadcast.StartedAt == nil {
			broadcast.StartedAt = &now
		}
		s.logger.Info("Broadcast resumed to sending status")
	}

	// Clear the paused timestamp
	broadcast.PausedAt = nil

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return err
	}

	s.logger.Info("Broadcast resumed successfully")

	// TODO: Trigger message processing if status is sending

	return nil
}

// CancelBroadcast cancels a scheduled broadcast
func (s *BroadcastService) CancelBroadcast(ctx context.Context, request *domain.CancelBroadcastRequest) error {
	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate cancel broadcast request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for cancellation")
		return err
	}

	// Only scheduled or paused broadcasts can be cancelled
	if broadcast.Status != domain.BroadcastStatusScheduled && broadcast.Status != domain.BroadcastStatusPaused {
		err := fmt.Errorf("only broadcasts with scheduled or paused status can be cancelled, current status: %s", broadcast.Status)
		s.logger.Error("Cannot cancel broadcast with invalid status")
		return err
	}

	// Update broadcast status and cancellation info
	broadcast.Status = domain.BroadcastStatusCancelled
	now := time.Now().UTC()
	broadcast.CancelledAt = &now
	broadcast.UpdatedAt = now

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return err
	}

	s.logger.Info("Broadcast cancelled successfully")

	return nil
}

// DeleteBroadcast deletes a broadcast
func (s *BroadcastService) DeleteBroadcast(ctx context.Context, request *domain.DeleteBroadcastRequest) error {
	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate delete broadcast request")
		return err
	}

	// Retrieve the broadcast to check its status
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for deletion")
		return err
	}

	// Broadcasts in 'sending' status cannot be deleted
	if broadcast.Status == domain.BroadcastStatusSending {
		err := fmt.Errorf("broadcasts in 'sending' status cannot be deleted")
		s.logger.Error("Cannot delete broadcast with sending status")
		return err
	}

	// Delete the broadcast
	err = s.repo.DeleteBroadcast(ctx, request.WorkspaceID, request.ID)
	if err != nil {
		s.logger.Error("Failed to delete broadcast from repository")
		return err
	}

	s.logger.Info("Broadcast deleted successfully")

	return nil
}

// SendToIndividual sends a broadcast to an individual recipient
func (s *BroadcastService) SendToIndividual(ctx context.Context, request *domain.SendToIndividualRequest) error {
	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate send to individual request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.BroadcastID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for individual sending")
		return err
	}

	// Determine which variation to use
	variationID := request.VariationID
	if variationID == "" && len(broadcast.TestSettings.Variations) > 0 {
		// If no variation ID specified, use the first one
		variationID = broadcast.TestSettings.Variations[0].ID
		s.logger.Debug("No variation specified, using first variation")
	} else if variationID == "" {
		err := fmt.Errorf("broadcast has no variations")
		s.logger.Error("Cannot send broadcast with no variations")
		return err
	}

	// Find the specified variation
	var variation *domain.BroadcastVariation
	for _, v := range broadcast.TestSettings.Variations {
		if v.ID == variationID {
			variation = &v
			break
		}
	}

	if variation == nil {
		err := fmt.Errorf("variation with ID %s not found in broadcast", variationID)
		s.logger.Error("Variation not found in broadcast")
		return err
	}

	// Fetch the contact if it exists to get the template data
	contact, err := s.contactRepo.GetContactByEmail(ctx, request.WorkspaceID, request.RecipientEmail)
	if err != nil && !strings.Contains(err.Error(), "contact not found") {
		// Log other errors but still return nil,false
		s.logger.Error("Error fetching contact by email")
		return err
	}

	// Fetch the template
	template, err := s.templateSvc.GetTemplateByID(ctx, request.WorkspaceID, variation.TemplateID, 1)
	if err != nil {
		s.logger.Error("Failed to fetch template for broadcast")
		return err
	}

	// Prepare contact data for template
	var templateData domain.MapOfAny
	if contact != nil {
		// Convert contact to JSON-compatible map using ToMapOfAny
		contactData, err := contact.ToMapOfAny()
		if err != nil {
			s.logger.Error("Failed to convert contact to map")
			return err
		}

		templateData = domain.MapOfAny{
			"contact": contactData,
			"broadcast": domain.MapOfAny{
				"id":   broadcast.ID,
				"name": broadcast.Name,
			},
			"variation": domain.MapOfAny{
				"id": variation.ID,
			},
		}
	} else {
		// If no contact data available, use empty data
		templateData = domain.MapOfAny{
			"contact": domain.MapOfAny{
				"email": request.RecipientEmail,
			},
		}
	}

	// Compile the template with contact data
	compiledTemplate, err := s.templateSvc.CompileTemplate(ctx, request.WorkspaceID, template.Email.VisualEditorTree, templateData)
	if err != nil {
		s.logger.Error("Failed to compile template for broadcast")
		return err
	}

	if !compiledTemplate.Success || compiledTemplate.HTML == nil {
		errMsg := "Template compilation failed"
		if compiledTemplate.Error != nil {
			errMsg = compiledTemplate.Error.Message
		}
		s.logger.Error("Failed to generate HTML from template")
		return fmt.Errorf("template compilation failed: %s", errMsg)
	}

	// Send the email with compiled HTML content
	err = s.emailSvc.SendEmail(
		ctx,
		request.WorkspaceID,
		"marketing", // Email provider type - adjust as needed
		template.Email.FromAddress,
		template.Email.FromName,
		request.RecipientEmail,
		template.Email.Subject,
		*compiledTemplate.HTML, // Use the compiled HTML content
	)
	if err != nil {
		s.logger.Error("Failed to send email to individual recipient")
		return err
	}

	s.logger.Info("Email sent to individual recipient successfully")

	// TODO: Record send event in analytics or message tracking system

	return nil
}

// SendWinningVariation sends the winning variation of an A/B test to remaining recipients
func (s *BroadcastService) SendWinningVariation(ctx context.Context, request *domain.SendWinningVariationRequest) error {
	// Validate the request
	if err := request.Validate(); err != nil {
		s.logger.Error("Failed to validate send winning variation request")
		return err
	}

	// Retrieve the broadcast
	broadcast, err := s.repo.GetBroadcast(ctx, request.WorkspaceID, request.BroadcastID)
	if err != nil {
		s.logger.Error("Failed to get broadcast for sending winning variation")
		return err
	}

	// Verify that the broadcast has A/B testing enabled
	if !broadcast.TestSettings.Enabled {
		err := fmt.Errorf("broadcast does not have A/B testing enabled")
		s.logger.Error("Cannot send winning variation for broadcast without A/B testing")
		return err
	}

	// Find the specified variation
	var winningVariation *domain.BroadcastVariation
	for _, v := range broadcast.TestSettings.Variations {
		if v.ID == request.VariationID {
			winningVariation = &v
			break
		}
	}

	if winningVariation == nil {
		err := fmt.Errorf("variation with ID %s not found in broadcast", request.VariationID)
		s.logger.Error("Winning variation not found in broadcast")
		return err
	}

	// Ensure tracking is enabled when sending the winning variation
	if !request.TrackingEnabled && broadcast.TrackingEnabled {
		// If tracking was enabled for the broadcast but not specified in the request,
		// use the broadcast's tracking setting
		request.TrackingEnabled = broadcast.TrackingEnabled
	}

	// Update the broadcast with the winning variation and winner sent time
	now := time.Now().UTC()
	broadcast.WinningVariation = request.VariationID
	broadcast.WinnerSentAt = &now
	broadcast.TrackingEnabled = request.TrackingEnabled
	broadcast.UpdatedAt = now

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast with winning variation information")
		return err
	}

	s.logger.Info("Broadcast updated with winning variation")

	// TODO: Implement the actual sending of the winning variation to remaining recipients
	// This would involve fetching the remaining recipients who didn't receive any test variation
	// and sending them the winning variation email

	return nil
}
