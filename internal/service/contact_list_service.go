package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type ContactListService struct {
	repo        domain.ContactListRepository
	contactRepo domain.ContactRepository
	listRepo    domain.ListRepository
	logger      logger.Logger
}

func NewContactListService(
	repo domain.ContactListRepository,
	contactRepo domain.ContactRepository,
	listRepo domain.ListRepository,
	logger logger.Logger,
) *ContactListService {
	return &ContactListService{
		repo:        repo,
		contactRepo: contactRepo,
		listRepo:    listRepo,
		logger:      logger,
	}
}

func (s *ContactListService) AddContactToList(ctx context.Context, contactList *domain.ContactList) error {
	// Verify contact exists by email
	contact, err := s.contactRepo.GetContactByEmail(ctx, contactList.Email)
	if err != nil {
		return fmt.Errorf("contact not found: %w", err)
	}

	// Use the email as the contact ID
	contactList.Email = contact.Email

	// Verify list exists
	list, err := s.listRepo.GetListByID(ctx, contactList.ListID)
	if err != nil {
		return fmt.Errorf("list not found: %w", err)
	}

	// If the list requires double opt-in, set status to pending
	if list.IsDoubleOptin {
		contactList.Status = domain.ContactListStatusPending
	}

	now := time.Now().UTC()
	contactList.CreatedAt = now
	contactList.UpdatedAt = now

	if err := contactList.Validate(); err != nil {
		return fmt.Errorf("invalid contact list: %w", err)
	}

	if err := s.repo.AddContactToList(ctx, contactList); err != nil {
		s.logger.WithField("email", contactList.Email).
			WithField("list_id", contactList.ListID).
			Error(fmt.Sprintf("Failed to add contact to list: %v", err))
		return fmt.Errorf("failed to add contact to list: %w", err)
	}

	return nil
}

func (s *ContactListService) GetContactListByIDs(ctx context.Context, email, listID string) (*domain.ContactList, error) {
	contactList, err := s.repo.GetContactListByIDs(ctx, email, listID)
	if err != nil {
		if _, ok := err.(*domain.ErrContactListNotFound); ok {
			return nil, err
		}
		s.logger.WithField("email", email).
			WithField("list_id", listID).
			Error(fmt.Sprintf("Failed to get contact list: %v", err))
		return nil, fmt.Errorf("failed to get contact list: %w", err)
	}

	return contactList, nil
}

func (s *ContactListService) GetContactsByListID(ctx context.Context, listID string) ([]*domain.ContactList, error) {
	// Verify list exists
	_, err := s.listRepo.GetListByID(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("list not found: %w", err)
	}

	contactLists, err := s.repo.GetContactsByListID(ctx, listID)
	if err != nil {
		s.logger.WithField("list_id", listID).
			Error(fmt.Sprintf("Failed to get contacts for list: %v", err))
		return nil, fmt.Errorf("failed to get contacts for list: %w", err)
	}

	return contactLists, nil
}

func (s *ContactListService) GetListsByEmail(ctx context.Context, email string) ([]*domain.ContactList, error) {
	// Verify contact exists by email
	_, err := s.contactRepo.GetContactByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}

	contactLists, err := s.repo.GetListsByEmail(ctx, email)
	if err != nil {
		s.logger.WithField("email", email).
			Error(fmt.Sprintf("Failed to get lists for contact: %v", err))
		return nil, fmt.Errorf("failed to get lists for contact: %w", err)
	}

	return contactLists, nil
}

func (s *ContactListService) UpdateContactListStatus(ctx context.Context, email, listID string, status domain.ContactListStatus) error {
	// Verify contact list exists
	_, err := s.repo.GetContactListByIDs(ctx, email, listID)
	if err != nil {
		return fmt.Errorf("contact list not found: %w", err)
	}

	if err := s.repo.UpdateContactListStatus(ctx, email, listID, status); err != nil {
		s.logger.WithField("email", email).
			WithField("list_id", listID).
			Error(fmt.Sprintf("Failed to update contact list status: %v", err))
		return fmt.Errorf("failed to update contact list status: %w", err)
	}

	return nil
}

func (s *ContactListService) RemoveContactFromList(ctx context.Context, email, listID string) error {
	if err := s.repo.RemoveContactFromList(ctx, email, listID); err != nil {
		s.logger.WithField("email", email).
			WithField("list_id", listID).
			Error(fmt.Sprintf("Failed to remove contact from list: %v", err))
		return fmt.Errorf("failed to remove contact from list: %w", err)
	}

	return nil
}
