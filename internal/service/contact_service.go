package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ContactService struct {
	repo          domain.ContactRepository
	authService   domain.AuthService
	workspaceRepo domain.WorkspaceRepository
	logger        logger.Logger
}

func NewContactService(repo domain.ContactRepository, workspaceRepo domain.WorkspaceRepository, authService domain.AuthService, logger logger.Logger) *ContactService {
	return &ContactService{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		authService:   authService,
		logger:        logger,
	}
}

func (s *ContactService) GetContactByEmail(ctx context.Context, workspaceID string, email string) (*domain.Contact, error) {

	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	contact, err := s.repo.GetContactByEmail(ctx, workspaceID, email)
	if err != nil {
		if strings.Contains(err.Error(), "contact not found") {
			return nil, err
		}
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to get contact by email: %v", err))
		return nil, fmt.Errorf("failed to get contact by email: %w", err)
	}

	return contact, nil
}

func (s *ContactService) GetContactByExternalID(ctx context.Context, externalID string, workspaceID string) (*domain.Contact, error) {

	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	contact, err := s.repo.GetContactByExternalID(ctx, externalID, workspaceID)
	if err != nil {
		if strings.Contains(err.Error(), "contact not found") {
			return nil, err
		}
		s.logger.WithField("external_id", externalID).Error(fmt.Sprintf("Failed to get contact by external ID: %v", err))
		return nil, fmt.Errorf("failed to get contact by external ID: %w", err)
	}

	return contact, nil
}

func (s *ContactService) GetContacts(ctx context.Context, req *domain.GetContactsRequest) (*domain.GetContactsResponse, error) {

	// Get the user ID from the context
	_, err := s.authService.AuthenticateUserForWorkspace(ctx, req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	response, err := s.repo.GetContacts(ctx, req)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get contacts: %v", err))
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	return response, nil
}

func (s *ContactService) DeleteContact(ctx context.Context, email string, workspaceID string) error {

	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	if err := s.repo.DeleteContact(ctx, email, workspaceID); err != nil {
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to delete contact: %v", err))
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}

func (s *ContactService) BatchImportContacts(ctx context.Context, workspaceID string, contacts []*domain.Contact) error {

	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Validate all contacts first
	for i, contact := range contacts {
		now := time.Now().UTC()
		contact.CreatedAt = now
		contact.UpdatedAt = now

		if err := contact.Validate(); err != nil {
			return fmt.Errorf("invalid contact at index %d: %w", i, err)
		}
	}

	// Process the batch
	if err := s.repo.BatchImportContacts(ctx, workspaceID, contacts); err != nil {
		s.logger.WithField("contacts_count", len(contacts)).Error(fmt.Sprintf("Failed to batch import contacts: %v", err))
		return fmt.Errorf("failed to batch import contacts: %w", err)
	}

	return nil
}

func (s *ContactService) UpsertContact(ctx context.Context, workspaceID string, contact *domain.Contact) error {

	_, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	now := time.Now().UTC()
	// Only set CreatedAt for new contacts
	if contact.CreatedAt.IsZero() {
		contact.CreatedAt = now
	}
	contact.UpdatedAt = now

	if err := contact.Validate(); err != nil {
		return fmt.Errorf("invalid contact: %w", err)
	}

	err = s.repo.UpsertContact(ctx, workspaceID, contact)
	if err != nil {
		s.logger.WithField("email", contact.Email).Error(fmt.Sprintf("Failed to upsert contact: %v", err))
		return fmt.Errorf("failed to upsert contact: %w", err)
	}

	return nil
}
