package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/google/uuid"
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
	if broadcastState.BroadcastID == "" {
		// In a real implementation, we'd expect the broadcast ID to be set when creating the task
		return false, fmt.Errorf("broadcast ID is missing in task state")
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
			return false, fmt.Errorf("failed to get recipient count: %w", err)
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
		return false, fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Pre-load all templates used by variations
	variationTemplates := make(map[string]*domain.Template)
	for _, variation := range broadcast.TestSettings.Variations {
		// Skip if we already loaded this template
		if _, exists := variationTemplates[variation.TemplateID]; exists {
			continue
		}

		// Load the template
		template, err := p.broadcastService.GetTemplateByID(ctx, task.WorkspaceID, variation.TemplateID)
		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"task_id":      task.ID,
				"broadcast_id": broadcastState.BroadcastID,
				"template_id":  variation.TemplateID,
				"error":        err.Error(),
			}).Error("Failed to load template for variation, will load on demand")
			// Continue without this template, it will be loaded on demand
			continue
		}

		// Store the template
		variationTemplates[variation.TemplateID] = template
	}

	p.logger.WithFields(map[string]interface{}{
		"task_id":          task.ID,
		"broadcast_id":     broadcastState.BroadcastID,
		"templates_loaded": len(variationTemplates),
	}).Info("Pre-loaded templates for broadcast")

	// Use a batch size of 50 for each fetch operation
	fetchBatchSize := 50

	// Process recipients until we hit a stopping condition
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
			// Try to continue with next batch rather than failing the task
			continue
		}

		// If we got no recipients, we're done
		if len(recipients) == 0 {
			p.logger.WithField("task_id", task.ID).Info("No more recipients to process")
			allDone = true
			break
		}

		// Process each recipient in parallel with semaphore to limit concurrency
		for _, contact := range recipients {
			// Check if we should stop processing
			select {
			case <-processCtx.Done():
				// We've hit the time limit, break out of the loop
				p.logger.WithField("task_id", task.ID).Info("Processing time limit reached during recipient processing")
				allDone = false
				goto CLEANUP // Use goto to break out of nested loop
			default:
				// Continue processing
			}

			// Acquire semaphore slot (blocks if we're at max parallelism)
			if err := sem.Acquire(ctx, 1); err != nil {
				p.logger.WithFields(map[string]interface{}{
					"task_id": task.ID,
					"error":   err.Error(),
				}).Error("Failed to acquire semaphore")
				continue
			}

			// Increment processed count under lock
			mu.Lock()
			processedCount++
			mu.Unlock()

			// Launch goroutine to process this recipient
			wg.Add(1)
			go func(contact *domain.Contact) {
				defer wg.Done()
				defer sem.Release(1)

				// Generate a unique message ID for tracking
				messageID := uuid.New().String()

				// Create a message history record before sending
				now := time.Now()
				broadcastIDCopy := broadcastState.BroadcastID // Make a copy to use in the pointer

				// Determine which template will be used for this contact
				var templateID string
				var templateVersion int

				// In a real implementation, this would be determined by the variation assignment logic
				// For now, we'll just use the first variation
				if len(broadcast.TestSettings.Variations) > 0 {
					templateID = broadcast.TestSettings.Variations[0].TemplateID

					// Find the template in our pre-loaded templates
					if template, ok := variationTemplates[templateID]; ok && template != nil {
						templateVersion = int(template.Version)
					} else {
						// Default to version 1 if not found
						templateVersion = 1
					}
				}

				// Get the contact list for unsubscribe link if available
				var contactList *domain.List
				// If this broadcast is targeting specific lists
				if len(broadcast.Audience.Lists) > 0 && len(broadcast.Audience.Lists[0]) > 0 {
					// Try to find the list in the contact's lists
					primaryListID := broadcast.Audience.Lists[0]
					for _, cl := range contact.ContactLists {
						if cl != nil && cl.ListID == primaryListID {
							// Found the list, create a simplified List object
							contactList = &domain.List{
								ID:   cl.ListID,
								Name: primaryListID, // Use ID as name if we don't have the actual name
							}
							break
						}
					}
				}

				// Compute list ID for metadata
				var listID string
				if contactList != nil {
					listID = contactList.ID
				}

				messageData := domain.MessageData{
					Data: map[string]interface{}{
						"contact": contact,
						// Other data used for template rendering
					},
					Metadata: map[string]interface{}{
						"broadcast_id": broadcastState.BroadcastID,
						"variation_id": broadcast.TestSettings.Variations[0].ID,
						"list_id":      listID,
					},
				}

				// Create the message history record
				messageHistory := &domain.MessageHistory{
					ID:              messageID,
					ContactID:       contact.Email,
					BroadcastID:     &broadcastIDCopy,
					TemplateID:      templateID,
					TemplateVersion: templateVersion,
					Channel:         "email",
					Status:          domain.MessageStatusSent,
					MessageData:     messageData,
					SentAt:          now,
					CreatedAt:       now,
					UpdatedAt:       now,
				}

				// Record the message in the message history
				if err := p.broadcastService.RecordMessageSent(ctx, task.WorkspaceID, messageHistory); err != nil {
					p.logger.WithFields(map[string]interface{}{
						"task_id":      task.ID,
						"broadcast_id": broadcastState.BroadcastID,
						"email":        contact.Email,
						"error":        err.Error(),
					}).Error("Failed to record message history")
					// Continue with sending even if recording fails
				}

				// Build template data using the utility method
				templateData, err := domain.BuildTemplateDataWithOptions(domain.TemplateDataOptions{
					Contact:     contact,
					Broadcast:   broadcast,
					List:        contactList,
					MessageID:   messageID,
					IncludeNow:  true,
					APIEndpoint: p.broadcastService.GetAPIEndpoint(),
					WorkspaceID: task.WorkspaceID,
				})
				if err != nil {
					p.logger.WithFields(map[string]interface{}{
						"task_id":      task.ID,
						"broadcast_id": broadcastState.BroadcastID,
						"email":        contact.Email,
						"error":        err.Error(),
					}).Error("Failed to build template data")

					// Fall back to basic template data if building fails
					templateData = map[string]interface{}{
						"contact": map[string]interface{}{
							"email": contact.Email,
						},
						"message_id": messageID,
					}
				}

				// Send the email with template data
				err = p.broadcastService.SendToContactWithTemplates(
					ctx,
					task.WorkspaceID,
					broadcastState.BroadcastID,
					contact,
					variationTemplates,
					templateData,
				)

				// Update counters and message status based on result
				mu.Lock()
				if err != nil {
					failureCount++
					p.logger.WithFields(map[string]interface{}{
						"task_id":      task.ID,
						"broadcast_id": broadcastState.BroadcastID,
						"email":        contact.Email,
						"error":        err.Error(),
					}).Error("Failed to send broadcast to contact")

					// Update message status to failed
					updateErr := p.broadcastService.UpdateMessageStatus(
						ctx,
						task.WorkspaceID,
						messageID,
						domain.MessageStatusFailed,
						time.Now(),
					)
					if updateErr != nil {
						p.logger.WithFields(map[string]interface{}{
							"task_id":      task.ID,
							"broadcast_id": broadcastState.BroadcastID,
							"email":        contact.Email,
							"message_id":   messageID,
							"error":        updateErr.Error(),
						}).Error("Failed to update message status to failed")
					}
				} else {
					successCount++

					// Update message status to delivered (this would normally happen via webhooks)
					// For demo purposes, we're marking it as delivered right away
					updateErr := p.broadcastService.UpdateMessageStatus(
						ctx,
						task.WorkspaceID,
						messageID,
						domain.MessageStatusDelivered,
						time.Now(),
					)
					if updateErr != nil {
						p.logger.WithFields(map[string]interface{}{
							"task_id":      task.ID,
							"broadcast_id": broadcastState.BroadcastID,
							"email":        contact.Email,
							"message_id":   messageID,
							"error":        updateErr.Error(),
						}).Error("Failed to update message status to delivered")
					}
				}
				mu.Unlock()
			}(contact)
		}

		// If we got fewer recipients than batch size, we're done
		if len(recipients) < fetchBatchSize {
			// Make sure to wait for all goroutines to finish
			wg.Wait()
			allDone = true
			break
		}
	}

CLEANUP:
	// Signal the progress logger to stop
	close(done)

	// Wait for all goroutines to finish
	wg.Wait()

	// Update state with results from this processing run
	broadcastState.SentCount += successCount
	broadcastState.FailedCount += failureCount
	broadcastState.RecipientOffset += int64(processedCount)

	// Calculate progress percentage
	if broadcastState.TotalRecipients > 0 {
		task.Progress = float64(broadcastState.RecipientOffset) / float64(broadcastState.TotalRecipients) * 100
		task.State.Progress = task.Progress
	}

	// Update task message
	task.State.Message = fmt.Sprintf("Sent to %d/%d recipients (%d failed)",
		broadcastState.SentCount, broadcastState.TotalRecipients, broadcastState.FailedCount)

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
