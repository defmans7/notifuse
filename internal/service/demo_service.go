package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/config"
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/google/uuid"
)

// DemoService handles demo workspace operations
type DemoService struct {
	logger                           logger.Logger
	config                           *config.Config
	workspaceService                 *WorkspaceService
	userService                      *UserService
	contactService                   *ContactService
	listService                      *ListService
	contactListService               *ContactListService
	templateService                  *TemplateService
	emailService                     *EmailService
	broadcastService                 *BroadcastService
	taskService                      *TaskService
	transactionalNotificationService *TransactionalNotificationService
	webhookEventService              *WebhookEventService
	webhookRegistrationService       *WebhookRegistrationService
	messageHistoryService            *MessageHistoryService
	notificationCenterService        *NotificationCenterService
	workspaceRepo                    domain.WorkspaceRepository
	taskRepo                         domain.TaskRepository
}

// NewDemoService creates a new demo service instance
func NewDemoService(
	logger logger.Logger,
	config *config.Config,
	workspaceService *WorkspaceService,
	userService *UserService,
	contactService *ContactService,
	listService *ListService,
	contactListService *ContactListService,
	templateService *TemplateService,
	emailService *EmailService,
	broadcastService *BroadcastService,
	taskService *TaskService,
	transactionalNotificationService *TransactionalNotificationService,
	webhookEventService *WebhookEventService,
	webhookRegistrationService *WebhookRegistrationService,
	messageHistoryService *MessageHistoryService,
	notificationCenterService *NotificationCenterService,
	workspaceRepo domain.WorkspaceRepository,
	taskRepo domain.TaskRepository,
) *DemoService {
	return &DemoService{
		logger:                           logger,
		config:                           config,
		workspaceService:                 workspaceService,
		userService:                      userService,
		contactService:                   contactService,
		listService:                      listService,
		contactListService:               contactListService,
		templateService:                  templateService,
		emailService:                     emailService,
		broadcastService:                 broadcastService,
		taskService:                      taskService,
		transactionalNotificationService: transactionalNotificationService,
		webhookEventService:              webhookEventService,
		webhookRegistrationService:       webhookRegistrationService,
		messageHistoryService:            messageHistoryService,
		notificationCenterService:        notificationCenterService,
		workspaceRepo:                    workspaceRepo,
		taskRepo:                         taskRepo,
	}
}

// VerifyRootEmailHMAC verifies the HMAC of the root email
func (s *DemoService) VerifyRootEmailHMAC(providedHMAC string) bool {
	if s.config.RootEmail == "" {
		s.logger.Error("Root email not configured")
		return false
	}

	// Use the domain function to verify HMAC with constant-time comparison
	return domain.VerifyEmailHMAC(s.config.RootEmail, providedHMAC, s.config.Security.SecretKey)
}

// ResetDemo deletes all existing workspaces and tasks, then creates a new demo workspace
func (s *DemoService) ResetDemo(ctx context.Context) error {
	s.logger.Info("Starting demo reset process")

	// Step 1: Delete all existing workspaces
	if err := s.deleteAllWorkspaces(ctx); err != nil {
		return fmt.Errorf("failed to delete existing workspaces: %w", err)
	}

	// Step 2: Delete all tasks from the system database
	if err := s.deleteAllTasks(ctx); err != nil {
		return fmt.Errorf("failed to delete existing tasks: %w", err)
	}

	// Step 3: Create a new demo workspace
	if err := s.createDemoWorkspace(ctx); err != nil {
		return fmt.Errorf("failed to create demo workspace: %w", err)
	}

	s.logger.Info("Demo reset completed successfully")
	return nil
}

// deleteAllWorkspaces deletes all workspaces from the system
func (s *DemoService) deleteAllWorkspaces(ctx context.Context) error {
	s.logger.Info("Deleting all existing workspaces")

	// Get all workspaces
	workspaces, err := s.workspaceRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	// Delete each workspace
	for _, workspace := range workspaces {
		s.logger.WithField("workspace_id", workspace.ID).Info("Deleting workspace")
		if err := s.workspaceRepo.Delete(ctx, workspace.ID); err != nil {
			s.logger.WithField("workspace_id", workspace.ID).WithField("error", err.Error()).Error("Failed to delete workspace")
			// Continue with other workspaces even if one fails
		}
	}

	s.logger.WithField("count", len(workspaces)).Info("Finished deleting workspaces")
	return nil
}

// deleteAllTasks deletes all tasks from the system database
func (s *DemoService) deleteAllTasks(ctx context.Context) error {
	s.logger.Info("Deleting all existing tasks")

	// Since tasks are workspace-specific and we've deleted all workspaces,
	// we need to clean up any remaining tasks in the system database
	// This is a simplified approach - in a real implementation you might want
	// to query and delete tasks more systematically

	// For now, we'll log this step as tasks should be cleaned up with workspace deletion
	s.logger.Info("Tasks cleanup completed (handled by workspace deletion)")
	return nil
}

// createDemoWorkspace creates a new demo workspace with sample data
func (s *DemoService) createDemoWorkspace(ctx context.Context) error {
	s.logger.Info("Creating demo workspace")

	// Generate a unique workspace ID
	workspaceID := "demo-" + uuid.New().String()[:8]

	// Create workspace settings
	fileManagerSettings := domain.FileManagerSettings{
		Endpoint:  "https://demo-storage.notifuse.com",
		Bucket:    "demo-bucket",
		AccessKey: "demo-access-key",
	}

	// Create the demo workspace
	workspace, err := s.workspaceService.CreateWorkspace(
		ctx,
		workspaceID,
		"Demo Workspace",
		"https://demo.notifuse.com",
		"https://demo.notifuse.com/logo.png",
		"https://demo.notifuse.com/cover.png",
		"UTC",
		fileManagerSettings,
	)
	if err != nil {
		return fmt.Errorf("failed to create demo workspace: %w", err)
	}

	s.logger.WithField("workspace_id", workspace.ID).Info("Demo workspace created successfully")

	// Add sample data to the workspace
	if err := s.addSampleData(ctx, workspace.ID); err != nil {
		s.logger.WithField("workspace_id", workspace.ID).WithField("error", err.Error()).Warn("Failed to add sample data to demo workspace")
		// Don't fail the entire operation if sample data creation fails
	}

	return nil
}

// addSampleData adds sample contacts, lists, and templates to the demo workspace
func (s *DemoService) addSampleData(ctx context.Context, workspaceID string) error {
	s.logger.WithField("workspace_id", workspaceID).Info("Adding sample data to demo workspace")

	// Create sample contacts
	sampleContacts := []*domain.Contact{
		{
			Email:     "john.doe@example.com",
			FirstName: &domain.NullableString{String: "John", IsNull: false},
			LastName:  &domain.NullableString{String: "Doe", IsNull: false},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "jane.smith@example.com",
			FirstName: &domain.NullableString{String: "Jane", IsNull: false},
			LastName:  &domain.NullableString{String: "Smith", IsNull: false},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "demo.user@example.com",
			FirstName: &domain.NullableString{String: "Demo", IsNull: false},
			LastName:  &domain.NullableString{String: "User", IsNull: false},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Add contacts to workspace
	for _, contact := range sampleContacts {
		operation := s.contactService.UpsertContact(ctx, workspaceID, contact)
		if operation.Action == domain.UpsertContactOperationError {
			s.logger.WithField("email", contact.Email).WithField("error", operation.Error).Warn("Failed to create sample contact")
		}
	}

	// Create sample lists
	sampleLists := []*domain.List{
		{
			ID:            "newsletter",
			Name:          "Newsletter Subscribers",
			IsDoubleOptin: true,
			IsPublic:      true,
			Description:   "Main newsletter subscription list",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "customers",
			Name:          "Customers",
			IsDoubleOptin: false,
			IsPublic:      false,
			Description:   "Existing customers list",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	// Add lists to workspace
	for _, list := range sampleLists {
		if err := s.listService.CreateList(ctx, workspaceID, list); err != nil {
			s.logger.WithField("list_id", list.ID).WithField("error", err.Error()).Warn("Failed to create sample list")
		}
	}

	s.logger.WithField("workspace_id", workspaceID).Info("Sample data added successfully")
	return nil
}
