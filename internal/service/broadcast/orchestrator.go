package broadcast

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

//go:generate mockgen -destination=./mocks/mock_broadcast_orchestrator.go -package=mocks github.com/Notifuse/notifuse/internal/service/broadcast BroadcastOrchestratorInterface

// BroadcastOrchestratorInterface defines the interface for broadcast orchestration
type BroadcastOrchestratorInterface interface {
	// CanProcess returns true if this processor can handle the given task type
	CanProcess(taskType string) bool

	// Process executes or continues a broadcast sending task
	Process(ctx context.Context, task *domain.Task) (bool, error)

	// LoadTemplatesForBroadcast loads all templates for a broadcast's variations
	LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error)

	// ValidateTemplates validates that the required templates are loaded and valid
	ValidateTemplates(templates map[string]*domain.Template) error

	// GetTotalRecipientCount gets the total number of recipients for a broadcast
	GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error)

	// FetchBatch retrieves a batch of recipients for a broadcast
	FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.ContactWithList, error)

	// SaveProgressState saves the current task progress to the repository
	SaveProgressState(ctx context.Context, workspaceID, taskID, broadcastID string, totalRecipients, sentCount, failedCount, processedCount int, lastSaveTime time.Time, startTime time.Time) (time.Time, error)
}

// BroadcastOrchestrator is the main processor for sending broadcasts
type BroadcastOrchestrator struct {
	messageSender    MessageSender
	broadcastService domain.BroadcastSender
	templateService  domain.TemplateService
	contactRepo      domain.ContactRepository
	taskRepo         domain.TaskRepository
	workspaceRepo    domain.WorkspaceRepository
	logger           logger.Logger
	config           *Config
	timeProvider     TimeProvider
}

// NewBroadcastOrchestrator creates a new broadcast orchestrator
func NewBroadcastOrchestrator(
	messageSender MessageSender,
	broadcastService domain.BroadcastSender,
	templateService domain.TemplateService,
	contactRepo domain.ContactRepository,
	taskRepo domain.TaskRepository,
	workspaceRepo domain.WorkspaceRepository,
	logger logger.Logger,
	config *Config,
	timeProvider TimeProvider,
) BroadcastOrchestratorInterface {
	if config == nil {
		config = DefaultConfig()
	}

	if timeProvider == nil {
		timeProvider = NewRealTimeProvider()
	}

	return &BroadcastOrchestrator{
		messageSender:    messageSender,
		broadcastService: broadcastService,
		templateService:  templateService,
		contactRepo:      contactRepo,
		taskRepo:         taskRepo,
		logger:           logger,
		config:           config,
		timeProvider:     timeProvider,
	}
}

// CanProcess returns true if this processor can handle the given task type
func (o *BroadcastOrchestrator) CanProcess(taskType string) bool {
	return taskType == "send_broadcast"
}

// LoadTemplatesForBroadcast loads all templates for a broadcast's variations
func (o *BroadcastOrchestrator) LoadTemplatesForBroadcast(ctx context.Context, workspaceID, broadcastID string) (map[string]*domain.Template, error) {
	startTime := time.Now()
	defer func() {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Debug("Template loading completed")
		// codecov:ignore end
	}()

	// Get the broadcast to access its template variations
	broadcast, err := o.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for templates")
		// codecov:ignore end
		return nil, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Process the broadcast's variations to get template IDs
	templateIDs := make(map[string]bool)
	for _, variation := range broadcast.TestSettings.Variations {
		templateIDs[variation.TemplateID] = true
	}

	if len(templateIDs) == 0 {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Error("No template variations found in broadcast")
		// codecov:ignore end
		return nil, NewBroadcastError(ErrCodeTemplateMissing, "no template variations found in broadcast", false, nil)
	}

	// Load all templates
	templates := make(map[string]*domain.Template)
	for templateID := range templateIDs {
		template, err := o.templateService.GetTemplateByID(ctx, workspaceID, templateID, 0) // Always use version 0
		if err != nil {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"broadcast_id": broadcastID,
				"workspace_id": workspaceID,
				"template_id":  templateID,
				"error":        err.Error(),
			}).Error("Failed to load template for broadcast")
			// codecov:ignore end
			continue // Don't fail the whole broadcast for one template
		}
		templates[templateID] = template
	}

	// Validate that we found at least one template
	if len(templates) == 0 {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Error("No valid templates found for broadcast")
		// codecov:ignore end
		return nil, NewBroadcastError(ErrCodeTemplateMissing, "no valid templates found for broadcast", false, nil)
	}

	// codecov:ignore start
	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":    broadcastID,
		"workspace_id":    workspaceID,
		"template_count":  len(templates),
		"variation_count": len(broadcast.TestSettings.Variations),
	}).Info("Templates loaded for broadcast")
	// codecov:ignore end

	return templates, nil
}

// ValidateTemplates validates that the required templates are loaded and valid
func (o *BroadcastOrchestrator) ValidateTemplates(templates map[string]*domain.Template) error {
	if len(templates) == 0 {
		return NewBroadcastError(ErrCodeTemplateMissing, "no templates provided for validation", false, nil)
	}

	// Validate each template
	for id, template := range templates {
		if template == nil {
			return NewBroadcastError(ErrCodeTemplateInvalid, "template is nil", false, nil)
		}

		// Ensure the template has the required fields for sending emails
		if template.Email == nil {
			// codecov:ignore start
			o.logger.WithField("template_id", id).Error("Template missing email configuration")
			// codecov:ignore end
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing email configuration", false, nil)
		}

		if template.Email.FromAddress == "" {
			// codecov:ignore start
			o.logger.WithField("template_id", id).Error("Template missing from address")
			// codecov:ignore end
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing from address", false, nil)
		}

		if template.Email.Subject == "" {
			// codecov:ignore start
			o.logger.WithField("template_id", id).Error("Template missing subject")
			// codecov:ignore end
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing subject", false, nil)
		}

		if template.Email.VisualEditorTree.Kind == "" {
			// codecov:ignore start
			o.logger.WithField("template_id", id).Error("Template missing content")
			// codecov:ignore end
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing content", false, nil)
		}
	}

	return nil
}

// GetTotalRecipientCount gets the total number of recipients for a broadcast
func (o *BroadcastOrchestrator) GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error) {
	startTime := time.Now()
	defer func() {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Debug("Recipient count completed")
		// codecov:ignore end
	}()

	// Get the broadcast to access audience settings
	broadcast, err := o.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient count")
		// codecov:ignore end
		return 0, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Use the contact repository to count recipients
	count, err := o.contactRepo.CountContactsForBroadcast(ctx, workspaceID, broadcast.Audience)
	if err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to count recipients for broadcast")
		// codecov:ignore end
		return 0, NewBroadcastError(ErrCodeRecipientFetch, "failed to count recipients", true, err)
	}

	// codecov:ignore start
	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":      broadcastID,
		"workspace_id":      workspaceID,
		"recipient_count":   count,
		"audience_lists":    len(broadcast.Audience.Lists),
		"audience_segments": len(broadcast.Audience.Segments),
	}).Info("Got recipient count for broadcast")
	// codecov:ignore end

	return count, nil
}

// FetchBatch retrieves a batch of recipients for a broadcast
func (o *BroadcastOrchestrator) FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.ContactWithList, error) {
	startTime := time.Now()
	defer func() {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"offset":       offset,
			"limit":        limit,
		}).Debug("Recipient batch fetch completed")
		// codecov:ignore end
	}()

	// Get the broadcast to access audience settings
	broadcast, err := o.broadcastService.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient fetch")
		// codecov:ignore end
		return nil, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Apply the actual batch limit from config if not specified
	if limit <= 0 {
		limit = o.config.FetchBatchSize
	}

	// Fetch contacts based on broadcast audience
	contactsWithList, err := o.contactRepo.GetContactsForBroadcast(ctx, workspaceID, broadcast.Audience, limit, offset)
	if err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"offset":       offset,
			"limit":        limit,
			"error":        err.Error(),
		}).Error("Failed to fetch recipients for broadcast")
		// codecov:ignore end
		return nil, NewBroadcastError(ErrCodeRecipientFetch, "failed to fetch recipients", true, err)
	}

	// codecov:ignore start
	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":     broadcastID,
		"workspace_id":     workspaceID,
		"offset":           offset,
		"limit":            limit,
		"contacts_fetched": len(contactsWithList),
		"with_list_info":   true,
	}).Info("Fetched recipient batch with list info")

	return contactsWithList, nil
}

// FormatDuration formats a duration in a human-readable form
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	} else {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", h, m)
	}
}

// CalculateProgress calculates the progress percentage (0-100)
func CalculateProgress(processed, total int) float64 {
	if total <= 0 {
		return 100.0 // Avoid division by zero
	}

	progress := float64(processed) / float64(total) * 100.0
	if progress > 100.0 {
		progress = 100.0
	}
	return progress
}

// FormatProgressMessage creates a human-readable progress message
func FormatProgressMessage(processed, total int, elapsed time.Duration) string {
	progress := CalculateProgress(processed, total)

	// Calculate remaining time if we have processed more than 5%
	var eta string
	if progress > 5.0 && processed > 0 {
		estimatedTotal := elapsed.Seconds() * float64(total) / float64(processed)
		remaining := estimatedTotal - elapsed.Seconds()

		if remaining > 0 {
			eta = fmt.Sprintf(", ETA: %s", FormatDuration(time.Duration(remaining)*time.Second))
		}
	}

	return fmt.Sprintf("Processed %d/%d recipients (%.1f%%)%s",
		processed, total, progress, eta)
}

// SaveProgressState saves the current task progress to the repository
func (o *BroadcastOrchestrator) SaveProgressState(
	ctx context.Context,
	workspaceID, taskID, broadcastID string,
	totalRecipients, sentCount, failedCount, processedCount int,
	lastSaveTime time.Time,
	startTime time.Time,
) (time.Time, error) {
	currentTime := o.timeProvider.Now()

	// Skip saving if not enough time has passed
	elapsed := currentTime.Sub(lastSaveTime)
	if elapsed < 5*time.Second {
		return lastSaveTime, nil
	}

	// Calculate progress
	elapsedSinceStart := currentTime.Sub(startTime)
	progress := CalculateProgress(processedCount, totalRecipients)
	message := FormatProgressMessage(processedCount, totalRecipients, elapsedSinceStart)

	// Create state
	state := &domain.TaskState{
		Progress: progress,
		Message:  message,
		SendBroadcast: &domain.SendBroadcastState{
			BroadcastID:     broadcastID,
			TotalRecipients: totalRecipients,
			SentCount:       sentCount,
			FailedCount:     failedCount,
			ChannelType:     "email", // Hardcoded for now
			RecipientOffset: int64(processedCount),
		},
	}

	// Save state
	err := o.taskRepo.SaveState(ctx, workspaceID, taskID, progress, state)
	if err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"task_id":      taskID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to save progress state")
		// codecov:ignore end
		return lastSaveTime, NewBroadcastError(ErrCodeTaskStateInvalid, "failed to save task state", true, err)
	}

	// Log progress
	// codecov:ignore start
	o.logger.WithFields(map[string]interface{}{
		"task_id":      taskID,
		"workspace_id": workspaceID,
		"progress":     progress,
		"sent":         sentCount,
		"failed":       failedCount,
	}).Debug("Saved progress state")

	return currentTime, nil
}

// Process executes or continues a broadcast sending task
func (o *BroadcastOrchestrator) Process(ctx context.Context, task *domain.Task) (bool, error) {
	o.logger.WithField("task_id", task.ID).Info("Processing send_broadcast task")

	// Initialize structured state if needed
	if task.State == nil {
		// Initialize a new state for the broadcast task
		task.State = &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				SentCount:       0,
				FailedCount:     0,
				RecipientOffset: 0, // Track how many recipients we've processed
			},
		}
	}

	// Initialize the SendBroadcast state if it doesn't exist yet
	if task.State.SendBroadcast == nil {
		task.State.SendBroadcast = &domain.SendBroadcastState{
			SentCount:       0,
			FailedCount:     0,
			RecipientOffset: 0,
		}
	}

	// Extract broadcast ID from task state or context
	broadcastState := task.State.SendBroadcast
	if broadcastState.BroadcastID == "" && task.BroadcastID != nil {
		// If state doesn't have broadcast ID but task does, use it
		broadcastState.BroadcastID = *task.BroadcastID
	}

	if broadcastState.BroadcastID == "" {
		// In a real implementation, we'd expect the broadcast ID to be set when creating the task
		return false, NewBroadcastErrorWithTask(
			ErrCodeTaskStateInvalid,
			"broadcast ID is missing in task state",
			task.ID,
			false,
			nil,
		)
	}

	// Track progress
	sentCount := broadcastState.SentCount
	failedCount := broadcastState.FailedCount
	processedCount := int(broadcastState.RecipientOffset)
	startTime := o.timeProvider.Now()
	lastSaveTime := o.timeProvider.Now()
	lastLogTime := o.timeProvider.Now()

	// Phase 1: Get recipient count if not already set
	if broadcastState.TotalRecipients == 0 {
		count, err := o.GetTotalRecipientCount(ctx, task.WorkspaceID, broadcastState.BroadcastID)
		if err != nil {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        err.Error(),
			}).Error("Failed to get recipient count for broadcast")
			// codecov:ignore end
			return false, err
		}

		broadcastState.TotalRecipients = count
		broadcastState.ChannelType = "email" // Hardcoded for now

		task.State.Message = "Preparing to send broadcast"
		task.Progress = 0

		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"task_id":          task.ID,
			"broadcast_id":     broadcastState.BroadcastID,
			"total_recipients": broadcastState.TotalRecipients,
			"channel_type":     broadcastState.ChannelType,
		}).Info("Broadcast sending initialized")
		// codecov:ignore end

		// If there are no recipients, we can mark as completed immediately
		if broadcastState.TotalRecipients == 0 {
			task.State.Message = "Broadcast completed: No recipients found"
			task.Progress = 100.0
			task.State.Progress = 100.0

			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
			}).Info("Broadcast completed with no recipients")
			// codecov:ignore end

			return true, nil
		}

		// Early return to save state before processing
		return false, nil
	}

	// Get the workspace to retrieve email provider settings
	workspace, err := o.workspaceRepo.GetByID(ctx, task.WorkspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Get the email provider using the workspace's GetEmailProvider method
	emailProvider, err := workspace.GetEmailProvider(true)
	if err != nil {
		return false, err
	}

	// Validate that the provider is configured
	if emailProvider == nil || emailProvider.Kind == "" {
		return false, fmt.Errorf("no email provider configured for marketing emails")
	}

	// Phase 2: Load templates
	templates, err := o.LoadTemplatesForBroadcast(ctx, task.WorkspaceID, broadcastState.BroadcastID)
	if err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        err.Error(),
		}).Error("Failed to load templates for broadcast")
		// codecov:ignore end
		return false, err
	}

	// Validate templates
	if err := o.ValidateTemplates(templates); err != nil {
		// codecov:ignore start
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        err.Error(),
		}).Error("Templates validation failed")
		// codecov:ignore end
		return false, err
	}

	// Phase 3: Process recipients in batches with a timeout
	processCtx, cancel := context.WithTimeout(ctx, o.config.MaxProcessTime)
	defer cancel()

	// Whether we've processed all recipients
	allDone := false

	// Process until timeout or completion
	for {
		select {
		case <-processCtx.Done():
			// We've hit the time limit, break out of the loop
			o.logger.WithField("task_id", task.ID).Info("Processing time limit reached")
			allDone = false
			break
		default:
			// Continue processing
		}

		// Fetch the next batch of recipients
		currentOffset := int(broadcastState.RecipientOffset)
		recipients, err := o.FetchBatch(
			ctx,
			task.WorkspaceID,
			broadcastState.BroadcastID,
			currentOffset,
			o.config.FetchBatchSize,
		)
		if err != nil {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"error":        err.Error(),
			}).Error("Failed to fetch recipients for broadcast")
			// codecov:ignore end
			return false, err
		}

		// If no more recipients, we're done
		if len(recipients) == 0 {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
			}).Info("No more recipients to process")
			// codecov:ignore end
			allDone = true
			break
		}

		// Process this batch of recipients
		sent, failed, err := o.messageSender.SendBatch(
			ctx,
			task.WorkspaceID,
			workspace.Settings.SecretKey,
			workspace.Settings.EmailTrackingEnabled,
			broadcastState.BroadcastID,
			recipients,
			templates,
			emailProvider,
		)

		// Handle errors during sending
		if err != nil {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"error":        err.Error(),
			}).Error("Error sending batch")
			// codecov:ignore end
			// Continue despite errors as we want to make progress
		}

		// Update progress counters
		sentCount += sent
		failedCount += failed
		processedCount += sent + failed

		// Log progress at regular intervals
		if o.timeProvider.Since(lastLogTime) >= o.config.ProgressLogInterval {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":         task.ID,
				"broadcast_id":    broadcastState.BroadcastID,
				"sent_count":      sentCount,
				"failed_count":    failedCount,
				"processed_count": processedCount,
				"total_count":     broadcastState.TotalRecipients,
				"progress":        CalculateProgress(processedCount, broadcastState.TotalRecipients),
				"elapsed":         o.timeProvider.Since(startTime).String(),
			}).Info("Broadcast progress update")
			// codecov:ignore end
			lastLogTime = o.timeProvider.Now()
		}

		// Save progress to the task
		var saveErr error
		lastSaveTime, saveErr = o.SaveProgressState(
			ctx,
			task.WorkspaceID,
			task.ID,
			broadcastState.BroadcastID,
			broadcastState.TotalRecipients,
			sentCount,
			failedCount,
			processedCount,
			lastSaveTime,
			startTime,
		)
		if saveErr != nil {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        saveErr.Error(),
			}).Error("Failed to save progress state")
			// codecov:ignore end
			// Continue processing despite save error
		}

		// Update local counters for state
		broadcastState.SentCount = sentCount
		broadcastState.FailedCount = failedCount
		broadcastState.RecipientOffset = int64(processedCount)

		// If we processed fewer recipients than requested, we're done
		if len(recipients) < o.config.FetchBatchSize {
			// codecov:ignore start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"count":        len(recipients),
			}).Info("Reached end of recipient list")
			// codecov:ignore end
			allDone = true
			break
		}
	}

	// Update task state with the latest progress data
	progress := CalculateProgress(processedCount, broadcastState.TotalRecipients)
	message := FormatProgressMessage(processedCount, broadcastState.TotalRecipients, time.Since(startTime))

	task.State.Progress = progress
	task.State.Message = message
	task.Progress = progress

	// codecov:ignore start
	o.logger.WithFields(map[string]interface{}{
		"task_id":          task.ID,
		"broadcast_id":     broadcastState.BroadcastID,
		"sent_total":       broadcastState.SentCount,
		"failed_total":     broadcastState.FailedCount,
		"total_recipients": broadcastState.TotalRecipients,
		"progress":         task.Progress,
		"all_done":         allDone,
	}).Info("Broadcast processing cycle completed")
	// codecov:ignore end
	return allDone, nil
}
