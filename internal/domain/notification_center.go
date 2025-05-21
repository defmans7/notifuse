package domain

import (
	"context"
	"errors"
	"net/url"
)

//go:generate mockgen -destination mocks/mock_notification_center_service.go -package mocks github.com/Notifuse/notifuse/internal/domain NotificationCenterService

type NotificationCenterService interface {
	// GetContactPreferences returns public lists and notifications for a contact
	GetContactPreferences(ctx context.Context, workspaceID string, email string, emailHMAC string) (*ContactPreferencesResponse, error)
}

type NotificationCenterRequest struct {
	Email       string `json:"email"`
	EmailHMAC   string `json:"email_hmac"`
	WorkspaceID string `json:"workspace_id"`
}

func (r *NotificationCenterRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.EmailHMAC == "" {
		return errors.New("email_hmac is required")
	}
	if r.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	return nil
}

func (r *NotificationCenterRequest) FromURLValues(values url.Values) error {
	r.Email = values.Get("email")
	r.EmailHMAC = values.Get("email_hmac")
	r.WorkspaceID = values.Get("workspace_id")
	return r.Validate()
}

// ContactPreferencesResponse contains the response data for the notification center
type ContactPreferencesResponse struct {
	Contact      *Contact       `json:"contact"`
	PublicLists  []*List        `json:"public_lists"`
	ContactLists []*ContactList `json:"contact_lists"`
	LogoURL      string         `json:"logo_url"`
	WebsiteURL   string         `json:"website_url"`
}
