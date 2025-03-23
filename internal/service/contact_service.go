package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/google/uuid"
)

type ContactService struct {
	repo   domain.ContactRepository
	logger logger.Logger
}

func NewContactService(repo domain.ContactRepository, logger logger.Logger) *ContactService {
	return &ContactService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ContactService) CreateContact(ctx context.Context, contact *domain.Contact) error {
	if contact.UUID == "" {
		contact.UUID = uuid.New().String()
	}

	now := time.Now().UTC()
	contact.CreatedAt = now
	contact.UpdatedAt = now

	if err := contact.Validate(); err != nil {
		return fmt.Errorf("invalid contact: %w", err)
	}

	if err := s.repo.CreateContact(ctx, contact); err != nil {
		s.logger.WithField("contact_uuid", contact.UUID).Error(fmt.Sprintf("Failed to create contact: %v", err))
		return fmt.Errorf("failed to create contact: %w", err)
	}

	return nil
}

func (s *ContactService) GetContactByUUID(ctx context.Context, uuid string) (*domain.Contact, error) {
	contact, err := s.repo.GetContactByUUID(ctx, uuid)
	if err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			return nil, err
		}
		s.logger.WithField("contact_uuid", uuid).Error(fmt.Sprintf("Failed to get contact: %v", err))
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return contact, nil
}

func (s *ContactService) GetContactByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	contact, err := s.repo.GetContactByEmail(ctx, email)
	if err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			return nil, err
		}
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to get contact by email: %v", err))
		return nil, fmt.Errorf("failed to get contact by email: %w", err)
	}

	return contact, nil
}

func (s *ContactService) GetContactByExternalID(ctx context.Context, externalID string) (*domain.Contact, error) {
	contact, err := s.repo.GetContactByExternalID(ctx, externalID)
	if err != nil {
		if _, ok := err.(*domain.ErrContactNotFound); ok {
			return nil, err
		}
		s.logger.WithField("external_id", externalID).Error(fmt.Sprintf("Failed to get contact by external ID: %v", err))
		return nil, fmt.Errorf("failed to get contact by external ID: %w", err)
	}

	return contact, nil
}

func (s *ContactService) GetContacts(ctx context.Context) ([]*domain.Contact, error) {
	contacts, err := s.repo.GetContacts(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get contacts: %v", err))
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	return contacts, nil
}

func (s *ContactService) UpdateContact(ctx context.Context, contact *domain.Contact) error {
	contact.UpdatedAt = time.Now().UTC()

	if err := contact.Validate(); err != nil {
		return fmt.Errorf("invalid contact: %w", err)
	}

	if err := s.repo.UpdateContact(ctx, contact); err != nil {
		s.logger.WithField("contact_uuid", contact.UUID).Error(fmt.Sprintf("Failed to update contact: %v", err))
		return fmt.Errorf("failed to update contact: %w", err)
	}

	return nil
}

func (s *ContactService) DeleteContact(ctx context.Context, uuid string) error {
	if err := s.repo.DeleteContact(ctx, uuid); err != nil {
		s.logger.WithField("contact_uuid", uuid).Error(fmt.Sprintf("Failed to delete contact: %v", err))
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}

func (s *ContactService) BatchImportContacts(ctx context.Context, contacts []*domain.Contact) error {
	// Validate all contacts first
	for i, contact := range contacts {
		if contact.UUID == "" {
			contact.UUID = uuid.New().String()
		}

		now := time.Now().UTC()
		contact.CreatedAt = now
		contact.UpdatedAt = now

		if err := contact.Validate(); err != nil {
			return fmt.Errorf("invalid contact at index %d: %w", i, err)
		}
	}

	// Process the batch
	if err := s.repo.BatchImportContacts(ctx, contacts); err != nil {
		s.logger.WithField("contacts_count", len(contacts)).Error(fmt.Sprintf("Failed to batch import contacts: %v", err))
		return fmt.Errorf("failed to batch import contacts: %w", err)
	}

	return nil
}
