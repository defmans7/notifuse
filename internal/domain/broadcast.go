package domain

import (
	"context"
	"fmt"
	"time"
)

//go:generate mockgen -destination mocks/mock_broadcast_service.go -package mocks github.com/Notifuse/notifuse/internal/domain BroadcastService
//go:generate mockgen -destination mocks/mock_broadcast_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain BroadcastRepository

// BroadcastStatus defines the current status of a broadcast
type BroadcastStatus string

const (
	BroadcastStatusDraft     BroadcastStatus = "draft"
	BroadcastStatusScheduled BroadcastStatus = "scheduled"
	BroadcastStatusSending   BroadcastStatus = "sending"
	BroadcastStatusPaused    BroadcastStatus = "paused"
	BroadcastStatusSent      BroadcastStatus = "sent"
	BroadcastStatusCancelled BroadcastStatus = "cancelled"
	BroadcastStatusFailed    BroadcastStatus = "failed"
)

// TestWinnerMetric defines the metric used to determine the winning A/B test variation
type TestWinnerMetric string

const (
	TestWinnerMetricOpenRate  TestWinnerMetric = "open_rate"
	TestWinnerMetricClickRate TestWinnerMetric = "click_rate"
)

// BroadcastTestSettings contains configuration for A/B testing
type BroadcastTestSettings struct {
	Enabled              bool                 `json:"enabled"`
	SamplePercentage     int                  `json:"sample_percentage"`
	AutoSendWinner       bool                 `json:"auto_send_winner"`
	AutoSendWinnerMetric TestWinnerMetric     `json:"auto_send_winner_metric,omitempty"`
	TestDurationHours    int                  `json:"test_duration_hours,omitempty"`
	Variations           []BroadcastVariation `json:"variations"`
}

// BroadcastVariation represents a single variation in an A/B test
type BroadcastVariation struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	TemplateID      string            `json:"template_id"`
	TemplateVersion int64             `json:"template_version"`
	Subject         string            `json:"subject"`
	PreviewText     string            `json:"preview_text,omitempty"`
	FromName        string            `json:"from_name"`
	FromEmail       string            `json:"from_email"`
	ReplyTo         string            `json:"reply_to,omitempty"`
	Metrics         *VariationMetrics `json:"metrics,omitempty"`
}

// VariationMetrics contains performance metrics for a variation
type VariationMetrics struct {
	Recipients int     `json:"recipients"`
	Delivered  int     `json:"delivered"`
	Opens      int     `json:"opens"`
	Clicks     int     `json:"clicks"`
	OpenRate   float64 `json:"open_rate"`
	ClickRate  float64 `json:"click_rate"`
}

// AudienceSettings defines how recipients are determined for a broadcast
type AudienceSettings struct {
	Type                string   `json:"type"` // "conditions", "import", "individual"
	SegmentConditions   MapOfAny `json:"segment_conditions,omitempty"`
	ImportedRecipients  []string `json:"imported_recipients,omitempty"`
	IndividualRecipient string   `json:"individual_recipient,omitempty"`
	ExcludeUnsubscribed bool     `json:"exclude_unsubscribed"`
	SkipDuplicateEmails bool     `json:"skip_duplicate_emails"`
	RateLimitPerMinute  int      `json:"rate_limit_per_minute,omitempty"`
}

// ScheduleSettings defines when a broadcast will be sent
type ScheduleSettings struct {
	SendImmediately      bool      `json:"send_immediately"`
	ScheduledTime        time.Time `json:"scheduled_time,omitempty"`
	UseRecipientTimezone bool      `json:"use_recipient_timezone"`
	TimeWindowStart      string    `json:"time_window_start,omitempty"` // HH:MM format
	TimeWindowEnd        string    `json:"time_window_end,omitempty"`   // HH:MM format
}

// Broadcast represents a one-time communication to multiple recipients (newsletter)
type Broadcast struct {
	ID               string                `json:"id"`
	WorkspaceID      string                `json:"workspace_id"`
	Name             string                `json:"name"`
	Status           BroadcastStatus       `json:"status"`
	Audience         AudienceSettings      `json:"audience"`
	Schedule         ScheduleSettings      `json:"schedule"`
	TestSettings     BroadcastTestSettings `json:"test_settings"`
	GoalID           string                `json:"goal_id,omitempty"`
	TrackingEnabled  bool                  `json:"tracking_enabled"`
	UTMParameters    *UTMParameters        `json:"utm_parameters,omitempty"`
	Metadata         MapOfAny              `json:"metadata,omitempty"`
	SentCount        int                   `json:"sent_count"`
	DeliveredCount   int                   `json:"delivered_count"`
	FailedCount      int                   `json:"failed_count"`
	WinningVariation string                `json:"winning_variation,omitempty"`
	TestSentAt       *time.Time            `json:"test_sent_at,omitempty"`
	WinnerSentAt     *time.Time            `json:"winner_sent_at,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	ScheduledAt      *time.Time            `json:"scheduled_at,omitempty"`
	StartedAt        *time.Time            `json:"started_at,omitempty"`
	CompletedAt      *time.Time            `json:"completed_at,omitempty"`
	CancelledAt      *time.Time            `json:"cancelled_at,omitempty"`
	PausedAt         *time.Time            `json:"paused_at,omitempty"`
}

// UTMParameters contains UTM tracking parameters for the broadcast
type UTMParameters struct {
	Source   string `json:"source,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Campaign string `json:"campaign,omitempty"`
	Term     string `json:"term,omitempty"`
	Content  string `json:"content,omitempty"`
}

// Validate validates the broadcast struct
func (b *Broadcast) Validate() error {
	if b.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if b.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(b.Name) > 255 {
		return fmt.Errorf("name must be less than 255 characters")
	}

	// Validate status
	switch b.Status {
	case BroadcastStatusDraft, BroadcastStatusScheduled, BroadcastStatusSending,
		BroadcastStatusPaused, BroadcastStatusSent, BroadcastStatusCancelled,
		BroadcastStatusFailed:
		// Valid status
	default:
		return fmt.Errorf("invalid broadcast status: %s", b.Status)
	}

	// Validate test settings if enabled
	if b.TestSettings.Enabled {
		if b.TestSettings.SamplePercentage <= 0 || b.TestSettings.SamplePercentage > 100 {
			return fmt.Errorf("test sample percentage must be between 1 and 100")
		}

		if len(b.TestSettings.Variations) < 2 {
			return fmt.Errorf("at least 2 variations are required for A/B testing")
		}

		if len(b.TestSettings.Variations) > 8 {
			return fmt.Errorf("maximum 8 variations are allowed for A/B testing")
		}

		if b.TestSettings.AutoSendWinner {
			switch b.TestSettings.AutoSendWinnerMetric {
			case TestWinnerMetricOpenRate, TestWinnerMetricClickRate:
				// Valid metric
			default:
				return fmt.Errorf("invalid test winner metric: %s", b.TestSettings.AutoSendWinnerMetric)
			}

			if b.TestSettings.TestDurationHours <= 0 {
				return fmt.Errorf("test duration must be greater than 0 hours")
			}
		}

		// Validate variations
		for i, variation := range b.TestSettings.Variations {
			if variation.TemplateID == "" {
				return fmt.Errorf("template_id is required for variation %d", i+1)
			}

			if variation.Subject == "" {
				return fmt.Errorf("subject is required for variation %d", i+1)
			}

			if variation.FromName == "" {
				return fmt.Errorf("from_name is required for variation %d", i+1)
			}

			if variation.FromEmail == "" {
				return fmt.Errorf("from_email is required for variation %d", i+1)
			}
		}
	}

	// Validate audience settings
	switch b.Audience.Type {
	case "conditions":
		if b.Audience.SegmentConditions == nil {
			return fmt.Errorf("segment conditions are required when audience type is 'conditions'")
		}
	case "import":
		if len(b.Audience.ImportedRecipients) == 0 {
			return fmt.Errorf("imported recipients are required when audience type is 'import'")
		}
	case "individual":
		if b.Audience.IndividualRecipient == "" {
			return fmt.Errorf("individual recipient is required when audience type is 'individual'")
		}
	default:
		return fmt.Errorf("invalid audience type: %s", b.Audience.Type)
	}

	// Validate schedule settings
	if !b.Schedule.SendImmediately && b.Schedule.ScheduledTime.IsZero() {
		return fmt.Errorf("scheduled time is required when not sending immediately")
	}

	if b.Schedule.UseRecipientTimezone {
		// If using recipient timezone with a delivery window, validate time window format
		if b.Schedule.TimeWindowStart != "" || b.Schedule.TimeWindowEnd != "" {
			// Time window should be in HH:MM format
			// This is a basic check, more thorough validation could be done
			if len(b.Schedule.TimeWindowStart) != 5 || len(b.Schedule.TimeWindowEnd) != 5 {
				return fmt.Errorf("time window must be in HH:MM format")
			}
		}
	}

	return nil
}

// CreateBroadcastRequest defines the request to create a new broadcast
type CreateBroadcastRequest struct {
	WorkspaceID     string                `json:"workspace_id"`
	Name            string                `json:"name"`
	Audience        AudienceSettings      `json:"audience"`
	Schedule        ScheduleSettings      `json:"schedule"`
	TestSettings    BroadcastTestSettings `json:"test_settings"`
	GoalID          string                `json:"goal_id,omitempty"`
	TrackingEnabled bool                  `json:"tracking_enabled"`
	UTMParameters   *UTMParameters        `json:"utm_parameters,omitempty"`
	Metadata        MapOfAny              `json:"metadata,omitempty"`
}

// Validate validates the create broadcast request
func (r *CreateBroadcastRequest) Validate() (*Broadcast, error) {
	broadcast := &Broadcast{
		WorkspaceID:     r.WorkspaceID,
		Name:            r.Name,
		Status:          BroadcastStatusDraft,
		Audience:        r.Audience,
		Schedule:        r.Schedule,
		TestSettings:    r.TestSettings,
		GoalID:          r.GoalID,
		TrackingEnabled: r.TrackingEnabled,
		UTMParameters:   r.UTMParameters,
		Metadata:        r.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := broadcast.Validate(); err != nil {
		return nil, err
	}

	return broadcast, nil
}

// UpdateBroadcastRequest defines the request to update an existing broadcast
type UpdateBroadcastRequest struct {
	WorkspaceID     string                `json:"workspace_id"`
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Audience        AudienceSettings      `json:"audience"`
	Schedule        ScheduleSettings      `json:"schedule"`
	TestSettings    BroadcastTestSettings `json:"test_settings"`
	GoalID          string                `json:"goal_id,omitempty"`
	TrackingEnabled bool                  `json:"tracking_enabled"`
	UTMParameters   *UTMParameters        `json:"utm_parameters,omitempty"`
	Metadata        MapOfAny              `json:"metadata,omitempty"`
}

// Validate validates the update broadcast request
func (r *UpdateBroadcastRequest) Validate(existingBroadcast *Broadcast) (*Broadcast, error) {
	if r.WorkspaceID != existingBroadcast.WorkspaceID {
		return nil, fmt.Errorf("workspace_id cannot be changed")
	}

	if r.ID != existingBroadcast.ID {
		return nil, fmt.Errorf("broadcast id cannot be changed")
	}

	// Cannot update a broadcast that is not in draft or scheduled status
	if existingBroadcast.Status != BroadcastStatusDraft &&
		existingBroadcast.Status != BroadcastStatusScheduled &&
		existingBroadcast.Status != BroadcastStatusPaused {
		return nil, fmt.Errorf("cannot update broadcast with status: %s", existingBroadcast.Status)
	}

	// Update the existing broadcast
	existingBroadcast.Name = r.Name
	existingBroadcast.Audience = r.Audience
	existingBroadcast.Schedule = r.Schedule
	existingBroadcast.TestSettings = r.TestSettings
	existingBroadcast.GoalID = r.GoalID
	existingBroadcast.TrackingEnabled = r.TrackingEnabled
	existingBroadcast.UTMParameters = r.UTMParameters
	existingBroadcast.Metadata = r.Metadata
	existingBroadcast.UpdatedAt = time.Now()

	if err := existingBroadcast.Validate(); err != nil {
		return nil, err
	}

	return existingBroadcast, nil
}

// ScheduleBroadcastRequest defines the request to schedule a broadcast
type ScheduleBroadcastRequest struct {
	WorkspaceID string    `json:"workspace_id"`
	ID          string    `json:"id"`
	ScheduledAt time.Time `json:"scheduled_at,omitempty"`
	SendNow     bool      `json:"send_now"`
}

// Validate validates the schedule broadcast request
func (r *ScheduleBroadcastRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if r.ID == "" {
		return fmt.Errorf("broadcast id is required")
	}

	if !r.SendNow && r.ScheduledAt.IsZero() {
		return fmt.Errorf("scheduled_at is required when not sending immediately")
	}

	return nil
}

// PauseBroadcastRequest defines the request to pause a sending broadcast
type PauseBroadcastRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

// Validate validates the pause broadcast request
func (r *PauseBroadcastRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if r.ID == "" {
		return fmt.Errorf("broadcast id is required")
	}

	return nil
}

// ResumeBroadcastRequest defines the request to resume a paused broadcast
type ResumeBroadcastRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

// Validate validates the resume broadcast request
func (r *ResumeBroadcastRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if r.ID == "" {
		return fmt.Errorf("broadcast id is required")
	}

	return nil
}

// CancelBroadcastRequest defines the request to cancel a scheduled broadcast
type CancelBroadcastRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

// Validate validates the cancel broadcast request
func (r *CancelBroadcastRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if r.ID == "" {
		return fmt.Errorf("broadcast id is required")
	}

	return nil
}

// SendToIndividualRequest defines the request to send a broadcast to an individual
type SendToIndividualRequest struct {
	WorkspaceID    string `json:"workspace_id"`
	BroadcastID    string `json:"broadcast_id"`
	RecipientEmail string `json:"recipient_email"`
	VariationID    string `json:"variation_id,omitempty"`
}

// Validate validates the send to individual request
func (r *SendToIndividualRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if r.BroadcastID == "" {
		return fmt.Errorf("broadcast_id is required")
	}

	if r.RecipientEmail == "" {
		return fmt.Errorf("recipient_email is required")
	}

	return nil
}

// ListBroadcastsParams defines parameters for listing broadcasts with pagination
type ListBroadcastsParams struct {
	WorkspaceID string
	Status      BroadcastStatus
	Limit       int
	Offset      int
}

// BroadcastListResponse defines the response for listing broadcasts
type BroadcastListResponse struct {
	Broadcasts []*Broadcast `json:"broadcasts"`
	TotalCount int          `json:"total_count"`
}

// BroadcastService defines the interface for broadcast operations
type BroadcastService interface {
	// CreateBroadcast creates a new broadcast
	CreateBroadcast(ctx context.Context, request *CreateBroadcastRequest) (*Broadcast, error)

	// GetBroadcast retrieves a broadcast by ID
	GetBroadcast(ctx context.Context, workspaceID, id string) (*Broadcast, error)

	// UpdateBroadcast updates an existing broadcast
	UpdateBroadcast(ctx context.Context, request *UpdateBroadcastRequest) (*Broadcast, error)

	// ListBroadcasts retrieves a list of broadcasts with pagination
	ListBroadcasts(ctx context.Context, params ListBroadcastsParams) (*BroadcastListResponse, error)

	// ScheduleBroadcast schedules a broadcast for sending
	ScheduleBroadcast(ctx context.Context, request *ScheduleBroadcastRequest) error

	// PauseBroadcast pauses a sending broadcast
	PauseBroadcast(ctx context.Context, request *PauseBroadcastRequest) error

	// ResumeBroadcast resumes a paused broadcast
	ResumeBroadcast(ctx context.Context, request *ResumeBroadcastRequest) error

	// CancelBroadcast cancels a scheduled broadcast
	CancelBroadcast(ctx context.Context, request *CancelBroadcastRequest) error

	// SendToIndividual sends a broadcast to an individual recipient
	SendToIndividual(ctx context.Context, request *SendToIndividualRequest) error
}

// BroadcastRepository defines the interface for broadcast persistence
type BroadcastRepository interface {
	// CreateBroadcast persists a new broadcast
	CreateBroadcast(ctx context.Context, broadcast *Broadcast) error

	// GetBroadcast retrieves a broadcast by ID
	GetBroadcast(ctx context.Context, workspaceID, id string) (*Broadcast, error)

	// UpdateBroadcast updates an existing broadcast
	UpdateBroadcast(ctx context.Context, broadcast *Broadcast) error

	// ListBroadcasts retrieves a list of broadcasts with pagination
	ListBroadcasts(ctx context.Context, params ListBroadcastsParams) (*BroadcastListResponse, error)
}

// ErrBroadcastNotFound is an error type for when a broadcast is not found
type ErrBroadcastNotFound struct {
	ID string
}

// Error returns the error message
func (e *ErrBroadcastNotFound) Error() string {
	return fmt.Sprintf("Broadcast not found with ID: %s", e.ID)
}
