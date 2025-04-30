package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// Helper function to safely get string value from NullableString
func getStringValue(nullable *domain.NullableString, defaultValue string) string {
	if nullable == nil || nullable.IsNull {
		return defaultValue
	}
	return nullable.String
}

// SendBroadcastProcessor implements domain.TaskProcessor for sending broadcasts
type SendBroadcastProcessor struct {
	broadcastService domain.BroadcastSender
	logger           logger.Logger
	maxParallelism   int64         // Maximum number of concurrent email sends
	maxProcessTime   time.Duration // Maximum time to process before saving state
}

// NewSendBroadcastProcessor creates a new SendBroadcastProcessor
func NewSendBroadcastProcessor(broadcastService domain.BroadcastSender, logger logger.Logger) *SendBroadcastProcessor {
	return &SendBroadcastProcessor{
		broadcastService: broadcastService,
		logger:           logger,
		maxParallelism:   10,               // Process 10 emails in parallel
		maxProcessTime:   50 * time.Second, // Run for 50 seconds before saving state
	}
}

// CanProcess returns true if this processor can handle the given task type
func (p *SendBroadcastProcessor) CanProcess(taskType string) bool {
	return taskType == "send_broadcast"
}

// Process executes or continues a broadcast sending task
func (p *SendBroadcastProcessor) Process(ctx context.Context, task *domain.Task) (bool, error) {
	p.logger.WithField("task_id", task.ID).Info("Processing send_broadcast task")

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
			RecipientOffset: 0, // Track how many recipients we've processed
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
		return false, &domain.ErrTaskExecution{
			TaskID: task.ID,
			Reason: "broadcast ID is missing in task state",
		}
	}

	// If we're just starting (first run), get the total recipients
	if broadcastState.TotalRecipients == 0 {
		// Get actual recipient count from the broadcast service
		recipientCount, err := p.broadcastService.GetRecipientCount(ctx, task.WorkspaceID, broadcastState.BroadcastID)
		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"error":        err.Error(),
			}).Error("Failed to get recipient count for broadcast")

			// Check if this is a not found error
			if _, ok := err.(*domain.ErrNotFound); ok {
				return false, &domain.ErrTaskExecution{
					TaskID: task.ID,
					Reason: "broadcast not found",
					Err:    err,
				}
			}

			return false, &domain.ErrTaskExecution{
				TaskID: task.ID,
				Reason: "failed to get recipient count",
				Err:    err,
			}
		}

		broadcastState.TotalRecipients = recipientCount
		broadcastState.ChannelType = "email" // Or could be "sms", "push", etc.

		task.State.Message = fmt.Sprintf("Sending %s broadcast to %d recipients",
			broadcastState.ChannelType, broadcastState.TotalRecipients)

		p.logger.WithFields(map[string]interface{}{
			"task_id":          task.ID,
			"broadcast_id":     broadcastState.BroadcastID,
			"total_recipients": broadcastState.TotalRecipients,
			"channel_type":     broadcastState.ChannelType,
		}).Info("Broadcast sending initialized")

		// If there are no recipients, we can mark as completed immediately
		if broadcastState.TotalRecipients == 0 {
			task.State.Message = "Broadcast completed: No recipients found"
			task.Progress = 100.0
			task.State.Progress = 100.0

			p.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
			}).Info("Broadcast completed with no recipients")

			return true, nil
		}

		// Early return after initialization to save state
		return false, nil
	}

	// Check if we've already sent to all recipients
	if broadcastState.RecipientOffset >= int64(broadcastState.TotalRecipients) {
		task.State.Message = fmt.Sprintf("Broadcast completed: %d sent, %d failed",
			broadcastState.SentCount, broadcastState.FailedCount)
		task.Progress = 100.0
		task.State.Progress = 100.0

		p.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"sent_count":   broadcastState.SentCount,
			"failed_count": broadcastState.FailedCount,
		}).Info("Broadcast sending completed")

		return true, nil
	}

	// Create a semaphore to limit concurrent operations
	sem := semaphore.NewWeighted(p.maxParallelism)

	// Create a context with timeout to limit processing time
	processCtx, cancel := context.WithTimeout(ctx, p.maxProcessTime)
	defer cancel()

	// Track newly processed recipients in this run
	var mu sync.Mutex // Mutex to protect counters
	successCount := 0
	failureCount := 0
	processedCount := 0
	allDone := false

	// Use a wait group to track all goroutines
	var wg sync.WaitGroup

	// Create a ticker to periodically log progress
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Start a goroutine to log progress
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				p.logger.WithFields(map[string]interface{}{
					"task_id":      task.ID,
					"broadcast_id": broadcastState.BroadcastID,
					"processed":    processedCount,
					"success":      successCount,
					"failed":       failureCount,
					"total_sent":   broadcastState.SentCount + successCount,
					"total_failed": broadcastState.FailedCount + failureCount,
				}).Info("Broadcast processing progress")
				mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	// Get the broadcast to access its template variations
	broadcast, err := p.broadcastService.GetBroadcast(ctx, task.WorkspaceID, broadcastState.BroadcastID)
	if err != nil {
		p.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
			"error":        err.Error(),
		}).Error("Failed to get broadcast for templates")

		// Check if this is a not found error
		if _, ok := err.(*domain.ErrNotFound); ok {
			return false, &domain.ErrTaskExecution{
				TaskID: task.ID,
				Reason: "broadcast not found",
				Err:    err,
			}
		}

		return false, &domain.ErrTaskExecution{
			TaskID: task.ID,
			Reason: "failed to get broadcast",
			Err:    err,
		}
	}

	// Pre-load all templates used by variations
	variationTemplates := make(map[string]*domain.Template)
	templateData := make(map[string]interface{})

	// Populate base template data that doesn't change per recipient
	templateData["broadcast"] = broadcast

	// Add UTM parameters if tracking is enabled
	if broadcast.TrackingEnabled && broadcast.UTMParameters != nil {
		templateData["utm"] = broadcast.UTMParameters
	}

	// Pre-load all templates
	for _, variation := range broadcast.TestSettings.Variations {
		template, err := p.broadcastService.GetTemplateByID(ctx, task.WorkspaceID, variation.TemplateID)
		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"template_id":  variation.TemplateID,
				"error":        err.Error(),
			}).Error("Failed to load template for variation")
			// Log the error but continue - we'll skip this variation when sending
			// We don't want to fail the entire broadcast if one template is missing
			continue
		}
		variationTemplates[variation.TemplateID] = template
	}

	// Check if we need to bail out before processing recipients
	if len(variationTemplates) == 0 {
		p.logger.WithFields(map[string]interface{}{
			"task_id":      task.ID,
			"broadcast_id": broadcastState.BroadcastID,
		}).Error("No valid templates found for broadcast")
		return false, &domain.ErrTaskExecution{
			TaskID: task.ID,
			Reason: "no valid templates found for broadcast",
		}
	}

	// Batch size for fetching contacts
	fetchBatchSize := 100

	for {
		select {
		case <-processCtx.Done():
			// We've hit the time limit, break out of the loop
			p.logger.WithField("task_id", task.ID).Info("Processing time limit reached")
			allDone = false
			goto CLEANUP // Use goto to break out of nested loop
		default:
			// Continue processing
		}

		// Get next batch of recipients
		recipients, err := p.broadcastService.GetBroadcastRecipients(
			ctx,
			task.WorkspaceID,
			broadcastState.BroadcastID,
			fetchBatchSize,
			int(broadcastState.RecipientOffset)+processedCount,
		)

		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"offset":       int(broadcastState.RecipientOffset) + processedCount,
				"error":        err.Error(),
			}).Error("Failed to get broadcast recipients")

			// Don't fail the task for temporary errors - retry on next cycle
			if strings.Contains(err.Error(), "context deadline exceeded") ||
				strings.Contains(err.Error(), "connection refused") {
				p.logger.Info("Temporary error retrieving recipients, will retry on next run")
				break
			}

			// Continue with next batch for retriable errors
			if strings.Contains(err.Error(), "no such host") ||
				strings.Contains(err.Error(), "i/o timeout") {
				continue
			}

			// Fatal errors should stop the task
			return false, &domain.ErrTaskExecution{
				TaskID: task.ID,
				Reason: "failed to get broadcast recipients",
				Err:    err,
			}
		}

		// If we got no recipients, we're done
		if len(recipients) == 0 {
			p.logger.WithField("task_id", task.ID).Info("No more recipients to process")
			allDone = true
			break
		}

		// Process each recipient in parallel with semaphore to limit concurrency
		for _, contact := range recipients {
			// Check context cancellation
			select {
			case <-processCtx.Done():
				p.logger.WithField("task_id", task.ID).Info("Processing time limit reached during contact iteration")
				allDone = false
				goto CLEANUP // Use goto to break out of nested loop
			default:
				// Continue processing
			}

			// Acquire a slot from the semaphore (will block if we're at max parallelism)
			if err := sem.Acquire(ctx, 1); err != nil {
				p.logger.WithFields(map[string]interface{}{
					"task_id": task.ID,
					"error":   err.Error(),
				}).Error("Failed to acquire semaphore")
				// Continue with next recipient
				continue
			}

			// Increment counter for this batch
			mu.Lock()
			processedCount++
			mu.Unlock()

			// Process this recipient in a goroutine
			wg.Add(1)
			go func(contact *domain.Contact) {
				defer wg.Done()
				defer sem.Release(1)

				// Create a copy of template data for this recipient
				recipientData := make(map[string]interface{})
				for k, v := range templateData {
					recipientData[k] = v
				}

				// Add contact-specific data
				contactData, err := contact.ToMapOfAny()
				if err != nil {
					mu.Lock()
					failureCount++
					mu.Unlock()
					p.logger.WithFields(map[string]interface{}{
						"task_id": task.ID,
						"email":   contact.Email,
						"error":   err.Error(),
					}).Error("Failed to convert contact to template data")
					return
				}
				recipientData["contact"] = contactData

				// Add unsubscribe URL
				apiEndpoint := p.broadcastService.GetAPIEndpoint()
				recipientData["unsubscribe_url"] = fmt.Sprintf("%s/api/contacts/unsubscribe?workspace_id=%s&email=%s",
					apiEndpoint, task.WorkspaceID, url.QueryEscape(contact.Email))

				// Send the email using the more efficient method with pre-loaded templates
				err = p.broadcastService.SendToContactWithTemplates(
					ctx,
					task.WorkspaceID,
					broadcastState.BroadcastID,
					contact,
					variationTemplates,
					recipientData,
				)

				// Update counters based on result
				mu.Lock()
				if err != nil {
					failureCount++
					p.logger.WithFields(map[string]interface{}{
						"task_id": task.ID,
						"email":   contact.Email,
						"error":   err.Error(),
					}).Error("Failed to send email to contact")
				} else {
					successCount++
				}
				mu.Unlock()
			}(contact)
		}

		// Wait for the current batch to finish before fetching the next batch
		wg.Wait()

		// Calculate progress after each batch
		currentProgress := float64(broadcastState.RecipientOffset+int64(processedCount)) / float64(broadcastState.TotalRecipients) * 100.0
		if currentProgress > 100.0 {
			currentProgress = 100.0
		}
		task.Progress = currentProgress
		task.State.Progress = currentProgress
		task.State.Message = fmt.Sprintf("Processed %d/%d recipients (%.1f%%)",
			broadcastState.RecipientOffset+int64(processedCount),
			broadcastState.TotalRecipients,
			currentProgress)
	}

CLEANUP:
	// Signal the progress logger to stop
	close(done)

	// Wait for any remaining goroutines
	wg.Wait()

	// Update the task state with the new counters
	broadcastState.SentCount += successCount
	broadcastState.FailedCount += failureCount
	broadcastState.RecipientOffset += int64(processedCount)

	// Update the progress based on the number of recipients processed
	currentProgress := float64(broadcastState.RecipientOffset) / float64(broadcastState.TotalRecipients) * 100.0
	if currentProgress > 100.0 {
		currentProgress = 100.0
	}
	task.Progress = currentProgress
	task.State.Progress = currentProgress

	// Update the task message
	task.State.Message = fmt.Sprintf("Processed %d/%d recipients (%.1f%%)",
		broadcastState.RecipientOffset,
		broadcastState.TotalRecipients,
		currentProgress)

	p.logger.WithFields(map[string]interface{}{
		"task_id":          task.ID,
		"broadcast_id":     broadcastState.BroadcastID,
		"sent_total":       broadcastState.SentCount,
		"failed_total":     broadcastState.FailedCount,
		"processed":        processedCount,
		"offset":           broadcastState.RecipientOffset,
		"total_recipients": broadcastState.TotalRecipients,
		"progress":         task.Progress,
		"all_done":         allDone,
	}).Info("Broadcast processing cycle completed")

	return allDone, nil
}
