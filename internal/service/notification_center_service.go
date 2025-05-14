package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
)

type NotificationCenterService struct {
	contactRepo     domain.ContactRepository
	workspaceRepo   domain.WorkspaceRepository
	listRepo        domain.ListRepository
	contactListRepo domain.ContactListRepository
	logger          logger.Logger
}

func NewNotificationCenterService(
	contactRepo domain.ContactRepository,
	workspaceRepo domain.WorkspaceRepository,
	listRepo domain.ListRepository,
	contactListRepo domain.ContactListRepository,
	logger logger.Logger,
) *NotificationCenterService {
	return &NotificationCenterService{
		contactRepo:     contactRepo,
		workspaceRepo:   workspaceRepo,
		listRepo:        listRepo,
		contactListRepo: contactListRepo,
		logger:          logger,
	}
}

// GetNotificationCenter returns the notification center data for a contact
// It returns public lists and public transactional notifications
func (s *NotificationCenterService) GetNotificationCenter(ctx context.Context, workspaceID string, email string, emailHMAC string) (*domain.NotificationCenterResponse, error) {
	// Verify that the email HMAC is valid
	// This is a simple security measure to verify the request is legitimate
	// The workspace should have a secret key that is used to generate the HMAC
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get workspace: %v", err))
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Using the workspace's settings secret key to verify the HMAC
	secretKey := workspace.Settings.SecretKey
	if !domain.VerifyEmailHMAC(email, emailHMAC, secretKey) {
		return nil, fmt.Errorf("invalid email verification")
	}

	// Get the contact
	contact, err := s.contactRepo.GetContactByEmail(ctx, workspaceID, email)
	if err != nil {
		if strings.Contains(err.Error(), "contact not found") {
			return nil, err
		}
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to get contact: %v", err))
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Get public lists for this workspace
	var publicLists []*domain.List

	// Get lists using the list service

	lists, err := s.listRepo.GetLists(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get lists: %v", err))
	} else {
		// Filter to only include public lists
		for _, list := range lists {
			if list.IsPublic {
				publicLists = append(publicLists, list)
			}
		}
	}

	return &domain.NotificationCenterResponse{
		Contact:      contact,
		PublicLists:  publicLists,
		ContactLists: contact.ContactLists,
		LogoURL:      workspace.Settings.LogoURL,
		WebsiteURL:   workspace.Settings.WebsiteURL,
	}, nil
}

// SubscribeToList subscribes a contact to a list
func (s *NotificationCenterService) SubscribeToList(ctx context.Context, workspaceID string, email string, listID string, emailHMAC *string) error {
	// Verify that the email HMAC is valid
	// This is a simple security measure to verify the request is legitimate
	// The workspace should have a secret key that is used to generate the HMAC
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get workspace: %v", err))
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	isAuthenticated := false

	if emailHMAC != nil {
		secretKey := workspace.Settings.SecretKey
		if !domain.VerifyEmailHMAC(email, *emailHMAC, secretKey) {
			return fmt.Errorf("invalid email verification")
		}
		isAuthenticated = true
	}

	contact := &domain.Contact{
		Email: email,
	}

	// upsert the contact
	_, err = s.contactRepo.UpsertContact(ctx, workspaceID, contact)
	if err != nil {
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to upsert contact: %v", err))
		return fmt.Errorf("failed to upsert contact: %w", err)
	}

	// get the list
	list, err := s.listRepo.GetListByID(ctx, workspaceID, listID)
	if err != nil {
		s.logger.WithField("list_id", listID).Error(fmt.Sprintf("Failed to get list: %v", err))
		return fmt.Errorf("failed to get list: %w", err)
	}

	// reject if the list is not public
	if !list.IsPublic {
		return fmt.Errorf("list is not public")
	}

	contactList := &domain.ContactList{
		Email:     email,
		ListID:    listID,
		ListName:  list.Name,
		Status:    domain.ContactListStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// if the list is double optin and the contact is not authenticated, set the status to pending
	if list.IsDoubleOptin && !isAuthenticated {
		contactList.Status = domain.ContactListStatusPending
	}

	// Subscribe to the list
	err = s.contactListRepo.AddContactToList(ctx, workspaceID, contactList)
	if err != nil {
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to subscribe to list: %v", err))
		return fmt.Errorf("failed to subscribe to list: %w", err)
	}

	return nil
}

// UnsubscribeFromList unsubscribes a contact from a list
func (s *NotificationCenterService) UnsubscribeFromList(ctx context.Context, workspaceID string, email string, emailHMAC string, listID string) error {
	// Verify that the email HMAC is valid
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get workspace: %v", err))
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Using the workspace's settings secret key to verify the HMAC
	secretKey := workspace.Settings.SecretKey
	if !domain.VerifyEmailHMAC(email, emailHMAC, secretKey) {
		return fmt.Errorf("invalid email verification")
	}

	// Check if contact exists
	_, err = s.contactRepo.GetContactByEmail(ctx, workspaceID, email)
	if err != nil {
		if strings.Contains(err.Error(), "contact not found") {
			return err
		}
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to get contact: %v", err))
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Check if list exists
	_, err = s.listRepo.GetListByID(ctx, workspaceID, listID)
	if err != nil {
		s.logger.WithField("list_id", listID).Error(fmt.Sprintf("Failed to get list: %v", err))
		return fmt.Errorf("failed to get list: %w", err)
	}

	// Unsubscribe from the list
	err = s.contactListRepo.UpdateContactListStatus(ctx, workspaceID, email, listID, domain.ContactListStatusUnsubscribed)
	if err != nil {
		s.logger.WithField("email", email).Error(fmt.Sprintf("Failed to unsubscribe from list: %v", err))
		return fmt.Errorf("failed to unsubscribe from list: %w", err)
	}

	return nil
}
