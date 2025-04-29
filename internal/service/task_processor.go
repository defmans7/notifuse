package service

import (
	"context"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// ImportContactsProcessor implements domain.TaskProcessor for importing contacts
type ImportContactsProcessor struct {
	contactService *ContactService
	logger         logger.Logger
}

// NewImportContactsProcessor creates a new ImportContactsProcessor
func NewImportContactsProcessor(contactService *ContactService, logger logger.Logger) *ImportContactsProcessor {
	return &ImportContactsProcessor{
		contactService: contactService,
		logger:         logger,
	}
}

// CanProcess returns true if this processor can handle the given task type
func (p *ImportContactsProcessor) CanProcess(taskType string) bool {
	return taskType == "import_contacts"
}

// Process executes or continues a task, returns whether the task has been completed
func (p *ImportContactsProcessor) Process(ctx context.Context, task *domain.Task) (bool, error) {
	p.logger.WithField("task_id", task.ID).Info("Processing import_contacts task")

	// Extract task state
	state := task.StateData
	if state == nil {
		// Initialize state for a new task
		state = map[string]interface{}{
			"total_contacts": 0,
			"processed":      0,
			"failed":         0,
			"current_page":   1,
			"total_pages":    0,
			"started_at":     time.Now().Format(time.RFC3339),
		}
	}

	// Get current progress info
	currentPage := int(state["current_page"].(float64))
	totalPages := int(state["total_pages"].(float64))
	processed := int(state["processed"].(float64))
	failed := int(state["failed"].(float64))

	// If we're just starting, get the total count
	if currentPage == 1 && totalPages == 0 {
		// In a real implementation, we would fetch the file details or source information
		// and get the total number of contacts to import

		// For this example, simulate a task with 5 pages
		totalContacts := 1000
		pageSize := 200
		totalPages = (totalContacts + pageSize - 1) / pageSize // Ceiling division

		state["total_contacts"] = totalContacts
		state["total_pages"] = totalPages
		state["page_size"] = pageSize

		p.logger.WithFields(map[string]interface{}{
			"task_id":        task.ID,
			"total_contacts": totalContacts,
			"total_pages":    totalPages,
		}).Info("Import task initialized")

		// Update task state before processing the first batch
		task.StateData = state
		return false, nil
	}

	// Process the current page
	pageSize := int(state["page_size"].(float64))

	// In a real implementation, this would fetch and process a batch of contacts
	// Simulate processing by sleeping and incrementing counters
	select {
	case <-ctx.Done():
		// Context was canceled (e.g., timeout)
		p.logger.WithField("task_id", task.ID).Warn("Import task interrupted")
		return false, ctx.Err()
	case <-time.After(2 * time.Second): // Simulate work
		// Update processing stats
		pageContacts := pageSize
		if currentPage == totalPages {
			// Last page might have fewer items
			pageContacts = int(state["total_contacts"].(float64)) - (currentPage-1)*pageSize
		}

		// Simulate a few failures
		successCount := pageContacts - (currentPage % 3) // Some arbitrary failures
		if successCount < 0 {
			successCount = 0
		}

		processed += successCount
		failed += (pageContacts - successCount)

		p.logger.WithFields(map[string]interface{}{
			"task_id":   task.ID,
			"page":      currentPage,
			"processed": processed,
			"page_size": pageSize,
			"success":   successCount,
			"failed":    failed,
		}).Info("Processed page")

		// Update state
		state["processed"] = processed
		state["failed"] = failed

		// Check if we've processed all pages
		if currentPage >= totalPages {
			// Task is complete
			state["completed_at"] = time.Now().Format(time.RFC3339)
			task.StateData = state

			// Calculate final progress percentage
			totalContacts := int(state["total_contacts"].(float64))
			if totalContacts > 0 {
				task.Progress = float64(processed) / float64(totalContacts) * 100
			} else {
				task.Progress = 100
			}

			p.logger.WithFields(map[string]interface{}{
				"task_id":        task.ID,
				"total_contacts": totalContacts,
				"processed":      processed,
				"failed":         failed,
				"progress":       task.Progress,
			}).Info("Import task completed")

			return true, nil
		}

		// Move to next page
		currentPage++
		state["current_page"] = currentPage

		// Calculate progress percentage
		totalContacts := int(state["total_contacts"].(float64))
		if totalContacts > 0 {
			task.Progress = float64(processed) / float64(totalContacts) * 100
		}

		task.StateData = state

		p.logger.WithFields(map[string]interface{}{
			"task_id":  task.ID,
			"progress": task.Progress,
			"page":     currentPage,
			"of_pages": totalPages,
		}).Info("Moving to next page")

		return false, nil
	}
}
