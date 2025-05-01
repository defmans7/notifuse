package broadcast

import (
	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

// Factory creates and wires together all the broadcast components
type Factory struct {
	broadcastService domain.BroadcastSender
	templateService  domain.TemplateService
	emailService     domain.EmailServiceInterface
	contactRepo      domain.ContactRepository
	taskRepo         domain.TaskRepository
	logger           logger.Logger
	config           *Config
}

// NewFactory creates a new factory for broadcast components
func NewFactory(
	broadcastService domain.BroadcastSender,
	templateService domain.TemplateService,
	emailService domain.EmailServiceInterface,
	contactRepo domain.ContactRepository,
	taskRepo domain.TaskRepository,
	logger logger.Logger,
	config *Config,
) *Factory {
	if config == nil {
		config = DefaultConfig()
	}

	return &Factory{
		broadcastService: broadcastService,
		templateService:  templateService,
		emailService:     emailService,
		contactRepo:      contactRepo,
		taskRepo:         taskRepo,
		logger:           logger,
		config:           config,
	}
}

// CreateMessageSender creates a new message sender
func (f *Factory) CreateMessageSender() MessageSender {
	return NewMessageSender(
		f.broadcastService,
		f.templateService,
		f.emailService,
		f.logger,
		f.config,
	)
}

// CreateOrchestrator creates a new broadcast orchestrator
func (f *Factory) CreateOrchestrator() BroadcastOrchestratorInterface {
	messageSender := f.CreateMessageSender()
	timeProvider := NewRealTimeProvider()

	return NewBroadcastOrchestrator(
		messageSender,
		f.broadcastService,
		f.templateService,
		f.contactRepo,
		f.taskRepo,
		f.logger,
		f.config,
		timeProvider,
	)
}

// RegisterWithTaskService registers the orchestrator with the task service
func (f *Factory) RegisterWithTaskService(taskService domain.TaskService) {
	orchestrator := f.CreateOrchestrator()
	taskService.RegisterProcessor(orchestrator)

	f.logger.Info("Broadcast orchestrator registered with task service")
}
