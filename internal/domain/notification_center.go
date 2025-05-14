package domain

import (
	"context"
	"errors"
	"net/url"

	"github.com/Notifuse/notifuse/pkg/crypto"
)

//go:generate mockgen -destination mocks/mock_notification_center_service.go -package mocks github.com/Notifuse/notifuse/internal/domain NotificationCenterService

type NotificationCenterService interface {
	// GetNotificationCenter returns public lists and notifications for a contact
	GetNotificationCenter(ctx context.Context, workspaceID string, email string, emailHMAC string) (*NotificationCenterResponse, error)
	SubscribeToList(ctx context.Context, workspaceID string, email string, listID string, emailHMAC *string) error
	UnsubscribeFromList(ctx context.Context, workspaceID string, email string, emailHMAC string, listID string) error
}

type SubscribeToListRequest struct {
	Email       string  `json:"email"`
	EmailHMAC   *string `json:"email_hmac,omitempty"`
	WorkspaceID string  `json:"workspace_id"`
	ListID      string  `json:"list_id"`
}

func (r *SubscribeToListRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if r.ListID == "" {
		return errors.New("list_id is required")
	}
	return nil
}

type UnsubscribeFromListRequest struct {
	Email       string `json:"email"`
	EmailHMAC   string `json:"email_hmac"`
	WorkspaceID string `json:"workspace_id"`
	ListID      string `json:"list_id"`
}

func (r *UnsubscribeFromListRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if r.ListID == "" {
		return errors.New("list_id is required")
	}
	return nil
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

// NotificationCenterResponse contains the response data for the notification center
type NotificationCenterResponse struct {
	Contact      *Contact       `json:"contact"`
	PublicLists  []*List        `json:"public_lists"`
	ContactLists []*ContactList `json:"contact_lists"`
	LogoURL      string         `json:"logo_url"`
	WebsiteURL   string         `json:"website_url"`
}

// VerifyEmailHMAC verifies if the provided HMAC for an email is valid
func VerifyEmailHMAC(email string, providedHMAC string, secretKey string) bool {
	// Use the crypto package to verify the HMAC
	computedHMAC := ComputeEmailHMAC(email, secretKey)
	return computedHMAC == providedHMAC
}

// ComputeEmailHMAC computes an HMAC for an email address using the workspace secret key
func ComputeEmailHMAC(email string, secretKey string) string {
	return crypto.ComputeHMAC256([]byte(email), secretKey)
}
