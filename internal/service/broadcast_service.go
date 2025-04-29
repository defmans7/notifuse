package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/google/uuid"
)

// BroadcastService handles sending broadcast messages
type BroadcastService struct {
	logger      logger.Logger
	repo        domain.BroadcastRepository
	contactRepo domain.ContactRepository
	emailSvc    domain.EmailServiceInterface
	templateSvc domain.TemplateService
	taskService domain.TaskService
	authService domain.AuthService
}

// NewBroadcastService creates a new BroadcastService
func NewBroadcastService(
	logger logger.Logger,
	repository domain.BroadcastRepository,
	emailService domain.EmailServiceInterface,
	contactRepository domain.ContactRepository,
	templateService domain.TemplateService,
	taskService domain.TaskService,
	authService domain.AuthService,
) *BroadcastService {

	return &BroadcastService{
		logger:      logger,
		repo:        repository,
		emailSvc:    emailService,
		contactRepo: contactRepository,
		templateSvc: templateService,
		taskService: taskService,
		authService: authService,
	}
}

// GetBroadcast retrieves a broadcast by ID
func (s *BroadcastService) GetBroadcast(ctx context.Context, workspaceID, broadcastID string) (*domain.Broadcast, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", broadcastID).Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Fetch the broadcast from the repository
	return s.repo.GetBroadcast(ctx, workspaceID, broadcastID)
}

// GetRecipientCount gets the count of recipients for a broadcast
func (s *BroadcastService) GetRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error) {
	// In a real implementation, this would count recipients from a database
	// For this example, we'll return a random count
	return 1000 + rand.Intn(5000), nil
}

// SendBatch sends a batch of messages for a broadcast
func (s *BroadcastService) SendBatch(ctx context.Context, workspaceID, broadcastID string, batchNumber, batchSize int) (int, int, error) {
	// In a real implementation, this would send actual messages through email/SMS/etc providers
	// For this example, we'll simulate sending with some random successes/failures

	// Simulate some processing time
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	case <-time.After(500 * time.Millisecond):
		// Simulate a 5% failure rate
		failureCount := batchSize / 20
		successCount := batchSize - failureCount

		s.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
			"broadcast_id": broadcastID,
			"batch":        batchNumber,
			"successes":    successCount,
			"failures":     failureCount,
		}).Info("Sent broadcast batch")

		return successCount, failureCount, nil
	}
}

// Broadcast represents a broadcast message campaign
type Broadcast struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Name        string     `json:"name"`
	ChannelType string     `json:"channel_type"` // email, sms, push, etc.
	Status      string     `json:"status"`       // pending, sending, completed, failed
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
}

// CreateBroadcast creates a new broadcast
func (s *BroadcastService) CreateBroadcast(ctx context.Context, request *domain.CreateBroadcastRequest) (*domain.Broadcast, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

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

// UpdateBroadcast updates an existing broadcast
func (s *BroadcastService) UpdateBroadcast(ctx context.Context, request *domain.UpdateBroadcastRequest) (*domain.Broadcast, error) {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, params.WorkspaceID)
	if err != nil {
		s.logger.Error("Failed to authenticate user for workspace")
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

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
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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
		broadcast.Schedule.IsScheduled = true
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

	// Create task for sending the broadcast
	if s.taskService != nil {
		// Create a task for this broadcast
		task := &domain.Task{
			WorkspaceID: request.WorkspaceID,
			Type:        "send_broadcast",
			Status:      domain.TaskStatusPending,
			MaxRuntime:  600, // 10 minutes
			MaxRetries:  3,
			State: &domain.TaskState{
				Message: fmt.Sprintf("Send broadcast: %s", broadcast.Name),
				SendBroadcast: &domain.SendBroadcastState{
					BroadcastID: broadcast.ID,
					ChannelType: broadcast.ChannelType,
					BatchSize:   100,
				},
			},
		}

		// Set the next run time based on whether it's immediate or scheduled
		if request.SendNow {
			// Schedule immediately
			task.NextRunAfter = nil
		} else {
			// Schedule for the future
			scheduledTime, _ := broadcast.Schedule.ParseScheduledDateTime()
			task.NextRunAfter = &scheduledTime
		}

		// Create the task
		if err := s.taskService.CreateTask(ctx, request.WorkspaceID, task); err != nil {
			s.logger.WithField("error", err.Error()).Error("Failed to create task for broadcast")
			// Continue despite task creation failure - don't block broadcast scheduling
		} else {
			// Store the task ID in the broadcast
			taskID := task.ID
			broadcast.TaskID = &taskID
			s.logger.WithFields(map[string]interface{}{
				"broadcast_id": broadcast.ID,
				"task_id":      task.ID,
			}).Info("Created task for broadcast")
		}
	}

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return err
	}

	return nil
}

// PauseBroadcast pauses a sending broadcast
func (s *BroadcastService) PauseBroadcast(ctx context.Context, request *domain.PauseBroadcastRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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

	// Pause the associated task if exists
	if s.taskService != nil && broadcast.TaskID != nil {
		// Get the task
		task, err := s.taskService.GetTask(ctx, request.WorkspaceID, *broadcast.TaskID)
		if err == nil && task != nil {
			// Only update task if it's in a running or pending state
			if task.Status == domain.TaskStatusRunning || task.Status == domain.TaskStatusPending {
				// Pause the task by setting NextRunAfter to a future time (1 day)
				pauseUntil := now.Add(24 * time.Hour)
				// Update task to be paused by setting its next run time
				if task.State == nil {
					task.State = &domain.TaskState{}
				}
				task.NextRunAfter = &pauseUntil
				err = s.taskService.SaveTaskProgress(ctx, request.WorkspaceID, *broadcast.TaskID, task.Progress, task.State)
				if err != nil {
					s.logger.WithField("error", err.Error()).
						WithField("task_id", *broadcast.TaskID).
						Warn("Failed to pause task for broadcast")
					// Continue despite task pause failure
				} else {
					s.logger.WithField("task_id", *broadcast.TaskID).
						Info("Successfully paused task for broadcast")
				}
			}
		}
	}

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
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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

	// Resume the associated task if exists
	if s.taskService != nil && broadcast.TaskID != nil {
		// Get the task
		task, err := s.taskService.GetTask(ctx, request.WorkspaceID, *broadcast.TaskID)
		if err == nil && task != nil {
			// Only update task if it's in a paused state
			if task.Status == domain.TaskStatusPaused {
				// Set NextRunAfter to current time to run immediately
				task.NextRunAfter = &now
				err = s.taskService.SaveTaskProgress(ctx, request.WorkspaceID, *broadcast.TaskID, task.Progress, task.State)
				if err != nil {
					s.logger.WithField("error", err.Error()).
						WithField("task_id", *broadcast.TaskID).
						Warn("Failed to resume task for broadcast")
					// Continue despite task resume failure
				} else {
					s.logger.WithField("task_id", *broadcast.TaskID).
						Info("Successfully resumed task for broadcast")
				}
			}
		}
	}

	// Persist the changes
	err = s.repo.UpdateBroadcast(ctx, broadcast)
	if err != nil {
		s.logger.Error("Failed to update broadcast in repository")
		return err
	}

	s.logger.Info("Broadcast resumed successfully")

	return nil
}

// CancelBroadcast cancels a scheduled broadcast
func (s *BroadcastService) CancelBroadcast(ctx context.Context, request *domain.CancelBroadcastRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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

	// Cancel the associated task if exists
	if s.taskService != nil && broadcast.TaskID != nil {
		// Delete the task
		err = s.taskService.DeleteTask(ctx, request.WorkspaceID, *broadcast.TaskID)
		if err != nil {
			s.logger.WithField("error", err.Error()).
				WithField("task_id", *broadcast.TaskID).
				Warn("Failed to delete task for cancelled broadcast")
			// Continue despite task deletion failure
		} else {
			s.logger.WithField("task_id", *broadcast.TaskID).
				Info("Successfully deleted task for cancelled broadcast")
		}
	}

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
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.ID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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

	// Delete the associated task if exists
	if s.taskService != nil && broadcast.TaskID != nil {
		err = s.taskService.DeleteTask(ctx, request.WorkspaceID, *broadcast.TaskID)
		if err != nil {
			s.logger.WithField("error", err.Error()).
				WithField("task_id", *broadcast.TaskID).
				Warn("Failed to delete task for deleted broadcast")
			// Continue despite task deletion failure
		}
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
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.BroadcastID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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

	return nil
}

// SendWinningVariation sends the winning variation of an A/B test to remaining recipients
func (s *BroadcastService) SendWinningVariation(ctx context.Context, request *domain.SendWinningVariationRequest) error {
	// Authenticate user for workspace
	var err error
	ctx, _, err = s.authService.AuthenticateUserForWorkspace(ctx, request.WorkspaceID)
	if err != nil {
		s.logger.WithField("broadcast_id", request.BroadcastID).Error("Failed to authenticate user for workspace")
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

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

	// Create task for sending the winning variation if needed
	if s.taskService != nil && broadcast.TaskID == nil {
		// Create a task for sending the winning variation
		task := &domain.Task{
			WorkspaceID: request.WorkspaceID,
			Type:        "send_broadcast_winning",
			Status:      domain.TaskStatusPending,
			MaxRuntime:  600, // 10 minutes
			MaxRetries:  3,
			State: &domain.TaskState{
				Message: fmt.Sprintf("Send winning variation for broadcast: %s", broadcast.Name),
				SendBroadcast: &domain.SendBroadcastState{
					BroadcastID: broadcast.ID,
					ChannelType: broadcast.ChannelType,
					BatchSize:   100,
				},
			},
		}

		// Create the task
		if err := s.taskService.CreateTask(ctx, request.WorkspaceID, task); err != nil {
			s.logger.WithField("error", err.Error()).Error("Failed to create task for sending winning variation")
			// Continue despite task creation failure
		} else {
			// Store the task ID in the broadcast and update again
			taskID := task.ID
			broadcast.TaskID = &taskID

			if err := s.repo.UpdateBroadcast(ctx, broadcast); err != nil {
				s.logger.WithField("error", err.Error()).Warn("Failed to update broadcast with task ID")
			}

			s.logger.WithFields(map[string]interface{}{
				"broadcast_id": broadcast.ID,
				"task_id":      task.ID,
			}).Info("Created task for sending winning variation")
		}
	}

	s.logger.Info("Broadcast updated with winning variation")

	return nil
}
