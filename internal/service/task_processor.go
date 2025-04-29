package service

import (
	"context"
	"fmt"
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

	// Initialize structured state if needed
	if task.State == nil {
		// Initialize a new state for the import task
		task.State = &domain.TaskState{
			Progress: 0,
			Message:  "Starting import",
			ImportContacts: &domain.ImportContactsState{
				TotalContacts:  0,
				ProcessedCount: 0,
				FailedCount:    0,
				CurrentPage:    1,
				TotalPages:     0,
				PageSize:       200,
				StartedAt:      time.Now(),
			},
		}
	}

	// Initialize the ImportContacts state if it doesn't exist yet
	if task.State.ImportContacts == nil {
		task.State.ImportContacts = &domain.ImportContactsState{
			TotalContacts:  0,
			ProcessedCount: 0,
			FailedCount:    0,
			CurrentPage:    1,
			TotalPages:     0,
			PageSize:       200,
			StartedAt:      time.Now(),
		}
	}

	// Get current state values from our structured state
	importState := task.State.ImportContacts
	currentPage := importState.CurrentPage
	totalPages := importState.TotalPages
	processed := importState.ProcessedCount
	failed := importState.FailedCount
	pageSize := importState.PageSize

	// If we're just starting, get the total count
	if currentPage == 1 && totalPages == 0 {
		// In a real implementation, we would fetch the file details or source information
		// and get the total number of contacts to import

		// For this example, simulate a task with 5 pages
		totalContacts := 1000
		pageSize = 200
		totalPages = (totalContacts + pageSize - 1) / pageSize // Ceiling division

		// Update our structured state
		importState.TotalContacts = totalContacts
		importState.TotalPages = totalPages
		importState.PageSize = pageSize

		// Set message in task state
		task.State.Message = fmt.Sprintf("Importing %d contacts", totalContacts)

		p.logger.WithFields(map[string]interface{}{
			"task_id":        task.ID,
			"total_contacts": totalContacts,
			"total_pages":    totalPages,
		}).Info("Import task initialized")

		// Return false to indicate task is not complete yet
		return false, nil
	}

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
		totalContacts := importState.TotalContacts

		if currentPage == totalPages {
			// Last page might have fewer items
			pageContacts = totalContacts - (currentPage-1)*pageSize
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

		// Update structured state
		importState.ProcessedCount = processed
		importState.FailedCount = failed

		// Update task message
		task.State.Message = fmt.Sprintf("Processed %d/%d contacts", processed, totalContacts)

		// Check if we've processed all pages
		if currentPage >= totalPages {
			// Task is complete
			now := time.Now()
			importState.CompletedAt = &now
			task.State.Message = fmt.Sprintf("Import completed: %d processed, %d failed", processed, failed)

			// Calculate final progress percentage
			if totalContacts > 0 {
				task.Progress = float64(processed) / float64(totalContacts) * 100
				task.State.Progress = task.Progress
			} else {
				task.Progress = 100
				task.State.Progress = 100
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
		importState.CurrentPage = currentPage

		// Calculate progress percentage
		if totalContacts > 0 {
			task.Progress = float64(processed) / float64(totalContacts) * 100
			task.State.Progress = task.Progress
		}

		p.logger.WithFields(map[string]interface{}{
			"task_id":  task.ID,
			"progress": task.Progress,
			"page":     currentPage,
			"of_pages": totalPages,
		}).Info("Moving to next page")

		return false, nil
	}
}
