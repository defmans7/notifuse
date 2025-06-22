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

	// LoadTemplates loads all templates for a broadcast's variations
	LoadTemplates(ctx context.Context, workspaceID string, templateIDs []string) (map[string]*domain.Template, error)

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
	messageSender   MessageSender
	broadcastRepo   domain.BroadcastRepository
	templateRepo    domain.TemplateRepository
	contactRepo     domain.ContactRepository
	taskRepo        domain.TaskRepository
	workspaceRepo   domain.WorkspaceRepository
	abTestEvaluator *ABTestEvaluator
	logger          logger.Logger
	config          *Config
	timeProvider    TimeProvider
}

// NewBroadcastOrchestrator creates a new broadcast orchestrator
func NewBroadcastOrchestrator(
	messageSender MessageSender,
	broadcastRepo domain.BroadcastRepository,
	templateRepo domain.TemplateRepository,
	contactRepo domain.ContactRepository,
	taskRepo domain.TaskRepository,
	workspaceRepo domain.WorkspaceRepository,
	abTestEvaluator *ABTestEvaluator,
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
		messageSender:   messageSender,
		broadcastRepo:   broadcastRepo,
		templateRepo:    templateRepo,
		contactRepo:     contactRepo,
		taskRepo:        taskRepo,
		workspaceRepo:   workspaceRepo,
		abTestEvaluator: abTestEvaluator,
		logger:          logger,
		config:          config,
		timeProvider:    timeProvider,
	}
}

// CanProcess returns true if this processor can handle the given task type
func (o *BroadcastOrchestrator) CanProcess(taskType string) bool {
	return taskType == "send_broadcast"
}

// LoadTemplates loads all templates for a broadcast's variations
func (o *BroadcastOrchestrator) LoadTemplates(ctx context.Context, workspaceID string, templateIDs []string) (map[string]*domain.Template, error) {

	// Load all templates
	templates := make(map[string]*domain.Template)
	for _, templateID := range templateIDs {
		template, err := o.templateRepo.GetTemplateByID(ctx, workspaceID, templateID, 0) // Always use version 0
		if err != nil {
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"workspace_id": workspaceID,
				"template_id":  templateID,
				"error":        err.Error(),
			}).Error("Failed to load template for broadcast")
			// codecov:ignore:end
			continue // Don't fail the whole broadcast for one template
		}
		templates[templateID] = template
	}

	// Validate that we found at least one template
	if len(templates) == 0 {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"workspace_id": workspaceID,
		}).Error("No valid templates found for broadcast")
		// codecov:ignore:end
		return nil, NewBroadcastError(ErrCodeTemplateMissing, "no valid templates found for broadcast", false, nil)
	}

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
			// codecov:ignore:start
			o.logger.WithField("template_id", id).Error("Template missing email configuration")
			// codecov:ignore:end
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing email configuration", false, nil)
		}

		if template.Email.Subject == "" {
			// codecov:ignore:start
			o.logger.WithField("template_id", id).Error("Template missing subject")
			// codecov:ignore:end
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing subject", false, nil)
		}

		if template.Email.VisualEditorTree.GetType() == "" {
			// codecov:ignore:start
			o.logger.WithField("template_id", id).Error("Template missing content")
			// codecov:ignore:end
			return NewBroadcastError(ErrCodeTemplateInvalid, "template missing content", false, nil)
		}
	}

	return nil
}

// GetTotalRecipientCount gets the total number of recipients for a broadcast
func (o *BroadcastOrchestrator) GetTotalRecipientCount(ctx context.Context, workspaceID, broadcastID string) (int, error) {
	startTime := time.Now()
	defer func() {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
		}).Debug("Recipient count completed")
		// codecov:ignore:end
	}()

	// Get the broadcast to access audience settings
	broadcast, err := o.broadcastRepo.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient count")
		// codecov:ignore:end
		return 0, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Use the contact repository to count recipients
	count, err := o.contactRepo.CountContactsForBroadcast(ctx, workspaceID, broadcast.Audience)
	if err != nil {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to count recipients for broadcast")
		// codecov:ignore:end
		return 0, NewBroadcastError(ErrCodeRecipientFetch, "failed to count recipients", true, err)
	}

	// codecov:ignore:start
	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":      broadcastID,
		"workspace_id":      workspaceID,
		"recipient_count":   count,
		"audience_lists":    len(broadcast.Audience.Lists),
		"audience_segments": len(broadcast.Audience.Segments),
	}).Info("Got recipient count for broadcast")
	// codecov:ignore:end

	return count, nil
}

// FetchBatch retrieves a batch of recipients for a broadcast
func (o *BroadcastOrchestrator) FetchBatch(ctx context.Context, workspaceID, broadcastID string, offset, limit int) ([]*domain.ContactWithList, error) {
	startTime := time.Now()
	defer func() {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"duration_ms":  time.Since(startTime).Milliseconds(),
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"offset":       offset,
			"limit":        limit,
		}).Debug("Recipient batch fetch completed")
		// codecov:ignore:end
	}()

	// Get the broadcast to access audience settings
	broadcast, err := o.broadcastRepo.GetBroadcast(ctx, workspaceID, broadcastID)
	if err != nil {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for recipient fetch")
		// codecov:ignore:end
		return nil, NewBroadcastError(ErrCodeBroadcastNotFound, "broadcast not found", false, err)
	}

	// Apply the actual batch limit from config if not specified
	if limit <= 0 {
		limit = o.config.FetchBatchSize
	}

	// Fetch contacts based on broadcast audience
	contactsWithList, err := o.contactRepo.GetContactsForBroadcast(ctx, workspaceID, broadcast.Audience, limit, offset)
	if err != nil {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcastID,
			"workspace_id": workspaceID,
			"audience":     broadcast.Audience,
			"offset":       offset,
			"limit":        limit,
			"error":        err.Error(),
		}).Error("Failed to fetch recipients for broadcast")
		// codecov:ignore:end
		return nil, NewBroadcastError(ErrCodeRecipientFetch, "failed to fetch recipients", true, err)
	}

	// codecov:ignore:start
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
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"task_id":      taskID,
			"workspace_id": workspaceID,
			"error":        err.Error(),
		}).Error("Failed to save progress state")
		// codecov:ignore:end
		return lastSaveTime, NewBroadcastError(ErrCodeTaskStateInvalid, "failed to save task state", true, err)
	}

	// Log progress
	// codecov:ignore:start
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

	// Store initial state for use in the defer function
	var broadcastID string
	isLastRetry := task.RetryCount >= task.MaxRetries-1 // Current attempt is the last one before max

	// Create a deferred function to update broadcast status to failed if we're returning an error on the last retry
	var err error
	var allDone bool

	// Defer function to mark broadcast as failed if we're returning an error on the last retry
	defer func() {
		if err != nil && isLastRetry && broadcastID != "" {
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastID,
				"retry_count":  task.RetryCount,
				"max_retries":  task.MaxRetries,
				"error":        err.Error(),
			}).Info("Task failed on last retry attempt, marking broadcast as failed")

			// Get the broadcast
			broadcast, getBroadcastErr := o.broadcastRepo.GetBroadcast(ctx, task.WorkspaceID, broadcastID)
			if getBroadcastErr != nil {
				o.logger.WithFields(map[string]interface{}{
					"task_id":      task.ID,
					"broadcast_id": broadcastID,
					"error":        getBroadcastErr.Error(),
				}).Error("Failed to get broadcast for status update on last retry")
				return
			}

			// Update broadcast status to failed
			broadcast.Status = domain.BroadcastStatusFailed
			broadcast.UpdatedAt = time.Now().UTC()

			// Save the updated broadcast
			updateErr := o.broadcastRepo.UpdateBroadcast(ctx, broadcast)
			if updateErr != nil {
				o.logger.WithFields(map[string]interface{}{
					"task_id":      task.ID,
					"broadcast_id": broadcastID,
					"error":        updateErr.Error(),
				}).Error("Failed to update broadcast status to failed")
			} else {
				o.logger.WithFields(map[string]interface{}{
					"task_id":      task.ID,
					"broadcast_id": broadcastID,
				}).Info("Broadcast marked as failed due to max retries reached")
			}

		}
	}()

	// Initialize structured state if needed
	if task.State == nil {
		// Initialize a new state for the broadcast task
		task.State = &domain.TaskState{
			Progress: 0,
			Message:  "Starting broadcast",
			SendBroadcast: &domain.SendBroadcastState{
				SentCount:                 0,
				FailedCount:               0,
				RecipientOffset:           0,        // Track how many recipients we've processed
				Phase:                     "single", // Default phase
				TestPhaseCompleted:        false,
				TestRecipientOffset:       0,
				WinnerPhaseRecipientCount: 0,
			},
		}
	}

	// Initialize the SendBroadcast state if it doesn't exist yet
	if task.State.SendBroadcast == nil {
		task.State.SendBroadcast = &domain.SendBroadcastState{
			SentCount:                 0,
			FailedCount:               0,
			RecipientOffset:           0,
			Phase:                     "single", // Default phase
			TestPhaseCompleted:        false,
			TestRecipientOffset:       0,
			WinnerPhaseRecipientCount: 0,
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
		err = NewBroadcastErrorWithTask(
			ErrCodeTaskStateInvalid,
			"broadcast ID is missing in task state",
			task.ID,
			false,
			nil,
		)
		return false, err
	}

	// Store the broadcast ID for the defer function
	broadcastID = broadcastState.BroadcastID

	// Track progress
	sentCount := broadcastState.SentCount
	failedCount := broadcastState.FailedCount
	processedCount := int(broadcastState.RecipientOffset)
	startTime := o.timeProvider.Now()
	lastSaveTime := o.timeProvider.Now()
	lastLogTime := o.timeProvider.Now()

	// Phase 1: Get recipient count if not already set
	if broadcastState.TotalRecipients == 0 {
		count, countErr := o.GetTotalRecipientCount(ctx, task.WorkspaceID, broadcastState.BroadcastID)
		if countErr != nil {
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        countErr.Error(),
			}).Error("Failed to get recipient count for broadcast")
			// codecov:ignore:end
			err = countErr
			return false, err
		}

		broadcastState.TotalRecipients = count
		broadcastState.ChannelType = "email" // Hardcoded for now

		task.State.Message = "Preparing to send broadcast"
		task.Progress = 0

		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"task_id":          task.ID,
			"broadcast_id":     broadcastState.BroadcastID,
			"total_recipients": broadcastState.TotalRecipients,
			"channel_type":     broadcastState.ChannelType,
		}).Info("Broadcast sending initialized")
		// codecov:ignore:end

		// If there are no recipients, we can mark as completed immediately
		if broadcastState.TotalRecipients == 0 {
			task.State.Message = "Broadcast completed: No recipients found"
			task.Progress = 100.0
			task.State.Progress = 100.0

			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
			}).Info("Broadcast completed with no recipients")
			// codecov:ignore:end

			allDone = true
			return allDone, err
		}

		// Update the task state with the broadcast state
		task.State.SendBroadcast = broadcastState

		// Early return to save state before processing
		allDone = false
		return allDone, err
	}

	// Get the workspace to retrieve email provider settings
	workspace, workspaceErr := o.workspaceRepo.GetByID(ctx, task.WorkspaceID)
	if workspaceErr != nil {
		err = fmt.Errorf("failed to get workspace: %w", workspaceErr)
		return false, err
	}

	// Get the email provider using the workspace's GetEmailProvider method
	emailProvider, providerErr := workspace.GetEmailProvider(true)
	if providerErr != nil {
		err = providerErr
		return false, err
	}

	// Validate that the provider is configured
	if emailProvider == nil || emailProvider.Kind == "" {
		err = fmt.Errorf("no email provider configured for marketing emails")
		return false, err
	}

	// Get the broadcast to access its template variations
	broadcast, err := o.broadcastRepo.GetBroadcast(ctx, task.WorkspaceID, broadcastState.BroadcastID)
	if err != nil {
		return false, err
	}

	// Check if we should perform auto winner evaluation
	if broadcastState.Phase == "test" && broadcast.Status == domain.BroadcastStatusTestCompleted {
		if o.shouldEvaluateWinner(broadcast) {
			if err := o.evaluateWinner(ctx, broadcast, broadcastState); err != nil {
				// Log error but continue - will fall back to manual selection
				o.logger.WithFields(map[string]interface{}{
					"broadcast_id": broadcast.ID,
					"error":        err.Error(),
				}).Error("Auto winner evaluation failed, continuing with manual selection")
			}

			// Refresh broadcast after potential evaluation
			broadcast, err = o.broadcastRepo.GetBroadcast(ctx, task.WorkspaceID, broadcastState.BroadcastID)
			if err != nil {
				return false, err
			}
		}
	}

	// Determine the current phase based on broadcast status and task state
	if broadcastState.Phase == "" || broadcastState.Phase == "single" {
		if broadcast.TestSettings.Enabled && len(broadcast.TestSettings.Variations) > 1 {
			// This is an A/B test
			if broadcast.Status == domain.BroadcastStatusSending {
				if broadcast.WinningTemplate != "" {
					// Winner already selected, proceed to winner phase
					broadcastState.Phase = "winner"
					// Initialize winner offset to start after test recipients if not already set
					if broadcastState.WinnerRecipientOffset == 0 && broadcastState.TestRecipientOffset > 0 {
						broadcastState.WinnerRecipientOffset = broadcastState.TestRecipientOffset
					}
				} else {
					// Start test phase
					broadcastState.Phase = "test"
				}
			} else if broadcast.Status == domain.BroadcastStatusWinnerSelected {
				// Winner has been selected, proceed to winner phase
				broadcastState.Phase = "winner"
				// Initialize winner offset to start after test recipients if not already set
				if broadcastState.WinnerRecipientOffset == 0 && broadcastState.TestRecipientOffset > 0 {
					broadcastState.WinnerRecipientOffset = broadcastState.TestRecipientOffset
				}
			}
		} else {
			// Single template broadcast
			broadcastState.Phase = "single"
		}
	}

	// Calculate recipient counts based on phase
	testRecipientCount := 0
	winnerRecipientCount := 0

	if broadcastState.Phase == "test" || broadcastState.Phase == "winner" {
		// Calculate test sample size
		testRecipientCount = (broadcastState.TotalRecipients * broadcast.TestSettings.SamplePercentage) / 100
		if testRecipientCount < 1 {
			testRecipientCount = 1 // Minimum test size
		}
		winnerRecipientCount = broadcastState.TotalRecipients - testRecipientCount

		// Store calculated counts in state
		broadcastState.TestPhaseRecipientCount = testRecipientCount
		broadcastState.WinnerPhaseRecipientCount = winnerRecipientCount
	}

	// Determine which templates to load based on phase
	var templateIDs []string

	if broadcastState.Phase == "test" {
		// Load all variations for testing
		templateIDs = make([]string, len(broadcast.TestSettings.Variations))
		for i, variation := range broadcast.TestSettings.Variations {
			templateIDs[i] = variation.TemplateID
		}
	} else if broadcastState.Phase == "winner" {
		// Load only the winning template
		if broadcast.WinningTemplate != "" {
			templateIDs = []string{broadcast.WinningTemplate}
		} else {
			return false, fmt.Errorf("winner phase but no winning template selected")
		}
	} else {
		// Single template broadcast
		templateIDs = make([]string, len(broadcast.TestSettings.Variations))
		for i, variation := range broadcast.TestSettings.Variations {
			templateIDs[i] = variation.TemplateID
		}
	}

	// Phase 2: Load templates

	templates, templatesErr := o.LoadTemplates(ctx, task.WorkspaceID, templateIDs)
	if templatesErr != nil {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        templatesErr.Error(),
		}).Error("Failed to load templates for broadcast")
		// codecov:ignore:end
		err = templatesErr
		return false, err
	}

	// Validate templates
	if validateErr := o.ValidateTemplates(templates); validateErr != nil {
		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        validateErr.Error(),
		}).Error("Templates validation failed")
		// codecov:ignore:end
		err = validateErr
		return false, err
	}

	// Phase 3: Process recipients in batches with a timeout
	processCtx, cancel := context.WithTimeout(ctx, o.config.MaxProcessTime)
	defer cancel()

	// Whether we've processed all recipients
	allDone = false

	// Determine phase-specific limits
	var recipientLimit int
	var currentOffset int

	if broadcastState.Phase == "test" {
		recipientLimit = broadcastState.TestPhaseRecipientCount
		currentOffset = int(broadcastState.TestRecipientOffset)
	} else if broadcastState.Phase == "winner" {
		recipientLimit = broadcastState.WinnerPhaseRecipientCount
		currentOffset = int(broadcastState.WinnerRecipientOffset)
	} else {
		// Single template - process all recipients
		recipientLimit = broadcastState.TotalRecipients
		currentOffset = int(broadcastState.RecipientOffset)
	}

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

		// Check if phase is complete
		if broadcastState.Phase == "test" && int(broadcastState.TestRecipientOffset) >= recipientLimit {
			// IMPORTANT: Check if winner has already been selected to avoid race condition
			if broadcast.WinningTemplate != "" || broadcast.Status == domain.BroadcastStatusWinnerSelected {
				// Winner already selected - don't overwrite status, transition to winner phase
				broadcastState.Phase = "winner"
				// Initialize winner offset to start after test recipients if not already set
				if broadcastState.WinnerRecipientOffset == 0 && broadcastState.TestRecipientOffset > 0 {
					broadcastState.WinnerRecipientOffset = broadcastState.TestRecipientOffset
				}
				// Update phase-specific variables after phase transition
				recipientLimit = broadcastState.WinnerPhaseRecipientCount
				currentOffset = int(broadcastState.WinnerRecipientOffset)
				o.logger.WithFields(map[string]interface{}{
					"broadcast_id":     broadcast.ID,
					"task_id":          task.ID,
					"winning_template": broadcast.WinningTemplate,
					"broadcast_status": string(broadcast.Status),
				}).Info("Test phase complete but winner already selected - transitioning to winner phase")
				// Continue processing in winner phase instead of marking test as complete
				continue
			} else {
				// No winner selected yet - mark test as complete and await winner selection
				allDone = o.handleTestPhaseCompletion(ctx, broadcast, broadcastState)
				break
			}
		} else if broadcastState.Phase == "winner" && int(broadcastState.WinnerRecipientOffset) >= recipientLimit {
			// Winner phase complete
			allDone = true
			break
		} else if broadcastState.Phase == "single" && int(broadcastState.RecipientOffset) >= recipientLimit {
			// Single template broadcast complete
			allDone = true
			break
		}

		// Calculate remaining recipients for this phase
		remainingInPhase := recipientLimit - currentOffset
		batchSize := o.config.FetchBatchSize
		if remainingInPhase < batchSize {
			batchSize = remainingInPhase
		}

		if batchSize <= 0 {
			// Phase complete
			if broadcastState.Phase == "test" {
				// Check if winner has already been selected to avoid race condition
				if broadcast.WinningTemplate != "" || broadcast.Status == domain.BroadcastStatusWinnerSelected {
					// Winner already selected - transition to winner phase
					broadcastState.Phase = "winner"
					// Initialize winner offset to start after test recipients if not already set
					if broadcastState.WinnerRecipientOffset == 0 && broadcastState.TestRecipientOffset > 0 {
						broadcastState.WinnerRecipientOffset = broadcastState.TestRecipientOffset
					}
					// Update phase-specific variables after phase transition
					recipientLimit = broadcastState.WinnerPhaseRecipientCount
					currentOffset = int(broadcastState.WinnerRecipientOffset)
					o.logger.WithFields(map[string]interface{}{
						"broadcast_id":     broadcast.ID,
						"task_id":          task.ID,
						"winning_template": broadcast.WinningTemplate,
						"broadcast_status": string(broadcast.Status),
					}).Info("Test phase complete (no more recipients) but winner already selected - transitioning to winner phase")
					// Continue processing in winner phase
					continue
				} else {
					// No winner selected yet - mark test as complete
					allDone = o.handleTestPhaseCompletion(ctx, broadcast, broadcastState)
				}
			} else {
				allDone = true
			}
			break
		}

		// Fetch the next batch of recipients
		recipients, batchErr := o.FetchBatch(
			ctx,
			task.WorkspaceID,
			broadcastState.BroadcastID,
			currentOffset,
			batchSize,
		)
		if batchErr != nil {
			err = batchErr
			return false, err
		}

		// If no more recipients, we're done
		if len(recipients) == 0 {
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"phase":        broadcastState.Phase,
			}).Info("No more recipients to process")
			// codecov:ignore:end
			if broadcastState.Phase == "test" {
				// Check if winner has already been selected to avoid race condition
				if broadcast.WinningTemplate != "" || broadcast.Status == domain.BroadcastStatusWinnerSelected {
					// Winner already selected - transition to winner phase
					broadcastState.Phase = "winner"
					// Initialize winner offset to start after test recipients if not already set
					if broadcastState.WinnerRecipientOffset == 0 && broadcastState.TestRecipientOffset > 0 {
						broadcastState.WinnerRecipientOffset = broadcastState.TestRecipientOffset
					}
					// Update phase-specific variables after phase transition
					recipientLimit = broadcastState.WinnerPhaseRecipientCount
					currentOffset = int(broadcastState.WinnerRecipientOffset)
					o.logger.WithFields(map[string]interface{}{
						"broadcast_id":     broadcast.ID,
						"task_id":          task.ID,
						"winning_template": broadcast.WinningTemplate,
						"broadcast_status": string(broadcast.Status),
					}).Info("No more recipients but winner already selected - transitioning to winner phase")
					// Continue processing in winner phase
					continue
				} else {
					// No winner selected yet - mark test as complete
					allDone = o.handleTestPhaseCompletion(ctx, broadcast, broadcastState)
				}
			} else {
				allDone = true
			}
			break
		}

		// Process this batch of recipients
		sent, failed, sendErr := o.messageSender.SendBatch(
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
		if sendErr != nil {
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"error":        sendErr.Error(),
			}).Error("Error sending batch")
			// codecov:ignore:end
			// Continue despite errors as we want to make progress
		}

		// Update progress counters
		sentCount += sent
		failedCount += failed
		processedCount += sent + failed

		// Update phase-specific counters
		if broadcastState.Phase == "test" {
			broadcastState.TestRecipientOffset += int64(sent + failed)
			currentOffset = int(broadcastState.TestRecipientOffset)
		} else if broadcastState.Phase == "winner" {
			broadcastState.WinnerRecipientOffset += int64(sent + failed)
			currentOffset = int(broadcastState.WinnerRecipientOffset)
		} else {
			broadcastState.RecipientOffset += int64(sent + failed)
			currentOffset = int(broadcastState.RecipientOffset)
		}

		// Log progress at regular intervals
		if o.timeProvider.Since(lastLogTime) >= o.config.ProgressLogInterval {
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":         task.ID,
				"broadcast_id":    broadcastState.BroadcastID,
				"phase":           broadcastState.Phase,
				"sent_count":      sentCount,
				"failed_count":    failedCount,
				"processed_count": processedCount,
				"phase_offset":    currentOffset,
				"phase_limit":     recipientLimit,
				"total_count":     broadcastState.TotalRecipients,
				"progress":        CalculateProgress(processedCount, broadcastState.TotalRecipients),
				"elapsed":         o.timeProvider.Since(startTime).String(),
			}).Info("Broadcast progress update")
			// codecov:ignore:end
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
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        saveErr.Error(),
			}).Error("Failed to save progress state")
			// codecov:ignore:end
			// Continue processing despite save error
		}

		// Update local counters for state
		broadcastState.SentCount = sentCount
		broadcastState.FailedCount = failedCount

		// If we processed fewer recipients than requested, we're done
		if len(recipients) < o.config.FetchBatchSize {
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       currentOffset,
				"count":        len(recipients),
			}).Info("Reached end of recipient list")
			// codecov:ignore:end
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

	// If the task is complete, update the broadcast status appropriately
	if allDone {
		var statusMessage string

		if broadcastState.Phase == "winner" || broadcastState.Phase == "single" {
			// Winner phase or single template complete - mark as sent
			broadcast.Status = domain.BroadcastStatusSent
			broadcast.UpdatedAt = time.Now().UTC()

			// Set completion time
			completedAt := time.Now().UTC()
			broadcast.CompletedAt = &completedAt

			// For winner phase, set winner sent time
			if broadcastState.Phase == "winner" {
				broadcast.WinnerSentAt = &completedAt
			}

			statusMessage = "sent"
		} else if broadcastState.Phase == "test" {
			// Test phase complete - should have been handled by handleTestPhaseCompletion
			// This shouldn't happen, but handle it gracefully
			o.logger.WithField("broadcast_id", broadcastState.BroadcastID).Warn("Test phase marked as complete in final processing - this should have been handled earlier")
			return false, nil // Pause task for winner selection
		}

		// Save the updated broadcast
		updateErr := o.broadcastRepo.UpdateBroadcast(ctx, broadcast)
		if updateErr != nil {
			// codecov:ignore:start
			o.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        updateErr.Error(),
			}).Error("Failed to update broadcast status to " + statusMessage)
			// codecov:ignore:end
			err = fmt.Errorf("failed to update broadcast status to %s: %w", statusMessage, updateErr)
			return false, err
		}

		// codecov:ignore:start
		o.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"sent_count":   sentCount,
			"failed_count": failedCount,
			"phase":        broadcastState.Phase,
		}).Info("Broadcast marked as " + statusMessage + " successfully")
		// codecov:ignore:end
	}

	// codecov:ignore:start
	o.logger.WithFields(map[string]interface{}{
		"task_id":          task.ID,
		"broadcast_id":     broadcastState.BroadcastID,
		"sent_total":       broadcastState.SentCount,
		"failed_total":     broadcastState.FailedCount,
		"total_recipients": broadcastState.TotalRecipients,
		"progress":         task.Progress,
		"all_done":         allDone,
	}).Info("Broadcast processing cycle completed")
	// codecov:ignore:end
	return allDone, err
}

// handleTestPhaseCompletion handles the transition from test phase to test_completed status
func (o *BroadcastOrchestrator) handleTestPhaseCompletion(ctx context.Context, broadcast *domain.Broadcast, broadcastState *domain.SendBroadcastState) bool {
	// Mark test phase as completed in state
	broadcastState.TestPhaseCompleted = true

	// Update broadcast status to test_completed
	broadcast.Status = domain.BroadcastStatusTestCompleted
	broadcast.UpdatedAt = time.Now().UTC()

	// Set test completion time
	now := time.Now().UTC()
	if broadcast.TestSentAt == nil {
		broadcast.TestSentAt = &now
	}

	// Save the updated broadcast status
	if err := o.broadcastRepo.UpdateBroadcast(ctx, broadcast); err != nil {
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id": broadcast.ID,
			"error":        err.Error(),
		}).Error("Failed to update broadcast status to test_completed")
		return false // Continue processing despite error
	}

	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":    broadcast.ID,
		"test_sent_count": broadcastState.TestRecipientOffset,
	}).Info("A/B test phase completed, awaiting winner selection")

	// Log completion - auto evaluation will happen on next task run if enabled
	if broadcast.TestSettings.AutoSendWinner {
		evaluationTime := now.Add(time.Duration(broadcast.TestSettings.TestDurationHours) * time.Hour)
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id":  broadcast.ID,
			"evaluation_at": evaluationTime,
			"test_duration": broadcast.TestSettings.TestDurationHours,
		}).Info("A/B test phase completed, auto winner evaluation will occur after test duration")
	} else {
		o.logger.WithField("broadcast_id", broadcast.ID).Info("A/B test phase completed, awaiting manual winner selection")
	}

	// Task should pause here - winner selection will resume it
	return false // Not all done, waiting for winner selection
}

// shouldEvaluateWinner checks if auto winner evaluation should be performed
func (o *BroadcastOrchestrator) shouldEvaluateWinner(broadcast *domain.Broadcast) bool {
	// Check if auto winner is enabled
	if !broadcast.TestSettings.AutoSendWinner {
		return false
	}

	// Check if test was sent (required for timing calculation)
	if broadcast.TestSentAt == nil {
		return false
	}

	// Calculate evaluation time and check if it has passed
	evaluationTime := broadcast.TestSentAt.Add(time.Duration(broadcast.TestSettings.TestDurationHours) * time.Hour)
	now := o.timeProvider.Now()

	if now.Before(evaluationTime) {
		o.logger.WithFields(map[string]interface{}{
			"broadcast_id":    broadcast.ID,
			"current_time":    now,
			"evaluation_time": evaluationTime,
			"test_duration":   broadcast.TestSettings.TestDurationHours,
		}).Debug("Auto evaluation time has not yet arrived")
		return false
	}

	return true
}

// evaluateWinner performs automatic winner evaluation and updates the broadcast
func (o *BroadcastOrchestrator) evaluateWinner(ctx context.Context, broadcast *domain.Broadcast, broadcastState *domain.SendBroadcastState) error {
	o.logger.WithField("broadcast_id", broadcast.ID).Info("Performing automatic winner evaluation")

	// Perform the evaluation using the ABTestEvaluator
	winnerTemplateID, err := o.abTestEvaluator.EvaluateAndSelectWinner(ctx, broadcast.WorkspaceID, broadcast.ID)
	if err != nil {
		return fmt.Errorf("auto winner evaluation failed: %w", err)
	}

	// Update task state to proceed to winner phase
	broadcastState.Phase = "winner"

	o.logger.WithFields(map[string]interface{}{
		"broadcast_id":    broadcast.ID,
		"winner_template": winnerTemplateID,
	}).Info("Auto winner evaluation completed successfully")

	return nil
}
