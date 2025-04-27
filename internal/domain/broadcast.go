package domain

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
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

// Value implements the driver.Valuer interface for database serialization
func (b BroadcastTestSettings) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan implements the sql.Scanner interface for database deserialization
func (b *BroadcastTestSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	cloned := bytes.Clone(v)
	return json.Unmarshal(cloned, b)
}

// BroadcastVariation represents a single variation in an A/B test
type BroadcastVariation struct {
	ID         string            `json:"id"`
	TemplateID string            `json:"template_id"`
	Metrics    *VariationMetrics `json:"metrics,omitempty"`
	// joined servers-side
	Template *Template `json:"template,omitempty"`
}

// Value implements the driver.Valuer interface for database serialization
func (v BroadcastVariation) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// Scan implements the sql.Scanner interface for database deserialization
func (v *BroadcastVariation) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	cloned := bytes.Clone(b)
	return json.Unmarshal(cloned, v)
}

// VariationMetrics contains performance metrics for a variation
type VariationMetrics struct {
	Recipients   int `json:"recipients"`
	Delivered    int `json:"delivered"`
	Opens        int `json:"opens"`
	Clicks       int `json:"clicks"`
	Bounced      int `json:"bounced"`
	Complained   int `json:"complained"`
	Unsubscribed int `json:"unsubscribed"`
}

// Value implements the driver.Valuer interface for database serialization
func (m VariationMetrics) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database deserialization
func (m *VariationMetrics) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	cloned := bytes.Clone(b)
	return json.Unmarshal(cloned, m)
}

// AudienceSettings defines how recipients are determined for a broadcast
type AudienceSettings struct {
	Lists               []string `json:"lists,omitempty"`
	Segments            []string `json:"segments,omitempty"`
	ExcludeUnsubscribed bool     `json:"exclude_unsubscribed"`
	SkipDuplicateEmails bool     `json:"skip_duplicate_emails"`
	RateLimitPerMinute  int      `json:"rate_limit_per_minute,omitempty"`
}

// Value implements the driver.Valuer interface for database serialization
func (a AudienceSettings) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for database deserialization
func (a *AudienceSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	cloned := bytes.Clone(b)
	return json.Unmarshal(cloned, a)
}

// ScheduleSettings defines when a broadcast will be sent
type ScheduleSettings struct {
	IsScheduled          bool   `json:"is_scheduled"`
	ScheduledDate        string `json:"scheduled_date,omitempty"` // Format: YYYY-MM-dd
	ScheduledTime        string `json:"scheduled_time,omitempty"` // Format: HH:mm
	Timezone             string `json:"timezone,omitempty"`       // IANA timezone format, e.g. "America/New_York"
	UseRecipientTimezone bool   `json:"use_recipient_timezone"`
}

// Value implements the driver.Valuer interface for database serialization
func (s ScheduleSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database deserialization
func (s *ScheduleSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	cloned := bytes.Clone(b)
	return json.Unmarshal(cloned, s)
}

// ParseScheduledDateTime parses the ScheduledDate and ScheduledTime fields and returns a time.Time
func (s *ScheduleSettings) ParseScheduledDateTime() (time.Time, error) {
	if s.ScheduledDate == "" || s.ScheduledTime == "" {
		return time.Time{}, nil
	}

	// Extract current time to preserve seconds and nanoseconds
	now := time.Now()
	seconds := now.Second()
	nanoseconds := now.Nanosecond()

	datetime := fmt.Sprintf("%s %s", s.ScheduledDate, s.ScheduledTime)
	var t time.Time
	var err error

	if s.Timezone == "" {
		t, err = time.Parse("2006-01-02 15:04", datetime)
		if err != nil {
			return time.Time{}, err
		}
	} else {
		loc, err := time.LoadLocation(s.Timezone)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid timezone: %s", err)
		}

		t, err = time.ParseInLocation("2006-01-02 15:04", datetime, loc)
		if err != nil {
			return time.Time{}, err
		}
	}

	// Add seconds and nanoseconds to preserve current time precision
	return t.Add(time.Duration(seconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond), nil
}

// SetScheduledDateTime formats a time.Time as ScheduledDate and ScheduledTime strings
func (s *ScheduleSettings) SetScheduledDateTime(t time.Time, timezone string) error {
	if t.IsZero() {
		s.ScheduledDate = ""
		s.ScheduledTime = ""
		s.Timezone = ""
		return nil
	}

	// If timezone is provided, convert time to that timezone
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return fmt.Errorf("invalid timezone: %s", err)
		}
		t = t.In(loc)
		s.Timezone = timezone
	}

	s.ScheduledDate = t.Format("2006-01-02")
	s.ScheduledTime = t.Format("15:04")
	return nil
}

// Broadcast represents a one-time communication to multiple recipients (newsletter)
type Broadcast struct {
	ID                string                `json:"id"`
	WorkspaceID       string                `json:"workspace_id"`
	Name              string                `json:"name"`
	Status            BroadcastStatus       `json:"status"`
	Audience          AudienceSettings      `json:"audience"`
	Schedule          ScheduleSettings      `json:"schedule"`
	TestSettings      BroadcastTestSettings `json:"test_settings"`
	TrackingEnabled   bool                  `json:"tracking_enabled"`
	UTMParameters     *UTMParameters        `json:"utm_parameters,omitempty"`
	Metadata          MapOfAny              `json:"metadata,omitempty"`
	TotalSent         int                   `json:"total_sent"`
	TotalDelivered    int                   `json:"total_delivered"`
	TotalBounced      int                   `json:"total_bounced"`
	TotalComplained   int                   `json:"total_complained"`
	TotalFailed       int                   `json:"total_failed"`
	TotalOpens        int                   `json:"total_opens"`
	TotalClicks       int                   `json:"total_clicks"`
	TotalUnsubscribed int                   `json:"total_unsubscribed"`
	WinningVariation  string                `json:"winning_variation,omitempty"`
	TestSentAt        *time.Time            `json:"test_sent_at,omitempty"`
	WinnerSentAt      *time.Time            `json:"winner_sent_at,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	StartedAt         *time.Time            `json:"started_at,omitempty"`
	CompletedAt       *time.Time            `json:"completed_at,omitempty"`
	CancelledAt       *time.Time            `json:"cancelled_at,omitempty"`
	PausedAt          *time.Time            `json:"paused_at,omitempty"`
}

// UTMParameters contains UTM tracking parameters for the broadcast
type UTMParameters struct {
	Source   string `json:"source,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Campaign string `json:"campaign,omitempty"`
	Term     string `json:"term,omitempty"`
	Content  string `json:"content,omitempty"`
}

// Value implements the driver.Valuer interface for database serialization
func (u UTMParameters) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// Scan implements the sql.Scanner interface for database deserialization
func (u *UTMParameters) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	cloned := bytes.Clone(b)
	return json.Unmarshal(cloned, u)
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

			// Tracking must be enabled to use auto-send winner feature
			if !b.TrackingEnabled {
				return fmt.Errorf("tracking must be enabled to use auto-send winner feature")
			}
		}

		// Validate variations
		for i, variation := range b.TestSettings.Variations {
			if variation.TemplateID == "" {
				return fmt.Errorf("template_id is required for variation %d", i+1)
			}
		}
	}

	// Validate audience settings
	switch {
	case len(b.Audience.Lists) > 0 && len(b.Audience.Segments) > 0:
		return fmt.Errorf("both lists and segments are specified")
	case len(b.Audience.Lists) > 0:
		// Lists are specified, no need to check segments
	case len(b.Audience.Segments) > 0:
		// Segments are specified, no need to check lists
	default:
		return fmt.Errorf("either lists or segments must be specified")
	}

	// Validate schedule settings
	if b.Schedule.IsScheduled && (b.Schedule.ScheduledDate == "" || b.Schedule.ScheduledTime == "") {
		return fmt.Errorf("scheduled date and time are required when not sending immediately")
	}

	if b.Schedule.IsScheduled {
		// Validate date format (YYYY-MM-DD)
		if len(b.Schedule.ScheduledDate) != 10 || b.Schedule.ScheduledDate[4] != '-' || b.Schedule.ScheduledDate[7] != '-' {
			return fmt.Errorf("scheduled date must be in YYYY-MM-DD format")
		}

		// Validate time format (HH:MM)
		if len(b.Schedule.ScheduledTime) != 5 || b.Schedule.ScheduledTime[2] != ':' {
			return fmt.Errorf("scheduled time must be in HH:MM format")
		}

		// If a timezone is specified, validate it
		if b.Schedule.Timezone != "" {
			_, err := time.LoadLocation(b.Schedule.Timezone)
			if err != nil {
				return fmt.Errorf("invalid timezone: %s", err)
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
		TrackingEnabled: r.TrackingEnabled,
		UTMParameters:   r.UTMParameters,
		Metadata:        r.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Set status to scheduled if the broadcast is scheduled
	if r.Schedule.IsScheduled {
		broadcast.Status = BroadcastStatusScheduled
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

// DeleteBroadcastRequest defines the request to delete a broadcast
type DeleteBroadcastRequest struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
}

// Validate validates the delete broadcast request
func (r *DeleteBroadcastRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if r.ID == "" {
		return fmt.Errorf("broadcast id is required")
	}

	return nil
}

// ListBroadcastsParams defines parameters for listing broadcasts with pagination
type ListBroadcastsParams struct {
	WorkspaceID   string
	Status        BroadcastStatus
	Limit         int
	Offset        int
	WithTemplates bool // Whether to fetch and include template details for each variation
}

// BroadcastListResponse defines the response for listing broadcasts
type BroadcastListResponse struct {
	Broadcasts []*Broadcast `json:"broadcasts"`
	TotalCount int          `json:"total_count"`
}

// SendWinningVariationRequest defines the request to send the winning variation of an A/B test
type SendWinningVariationRequest struct {
	WorkspaceID     string `json:"workspace_id"`
	BroadcastID     string `json:"broadcast_id"`
	VariationID     string `json:"variation_id"`
	TrackingEnabled bool   `json:"tracking_enabled"`
}

// Validate validates the send winning variation request
func (r *SendWinningVariationRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	if r.BroadcastID == "" {
		return fmt.Errorf("broadcast_id is required")
	}

	if r.VariationID == "" {
		return fmt.Errorf("variation_id is required")
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

// GetBroadcastsRequest is used to extract query parameters for listing broadcasts
type GetBroadcastsRequest struct {
	WorkspaceID   string `json:"workspace_id"`
	Status        string `json:"status,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	WithTemplates bool   `json:"with_templates,omitempty"`
}

// FromURLParams parses URL query parameters into the request
func (r *GetBroadcastRequest) FromURLParams(values url.Values) error {
	r.WorkspaceID = values.Get("workspace_id")
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	r.ID = values.Get("id")
	if r.ID == "" {
		return fmt.Errorf("id is required")
	}

	if withTemplatesStr := values.Get("with_templates"); withTemplatesStr != "" {
		var err error
		r.WithTemplates, err = ParseBoolParam(withTemplatesStr)
		if err != nil {
			return fmt.Errorf("invalid with_templates parameter: %w", err)
		}
	}

	return nil
}

// FromURLParams parses URL query parameters into the request
func (r *GetBroadcastsRequest) FromURLParams(values url.Values) error {
	r.WorkspaceID = values.Get("workspace_id")
	if r.WorkspaceID == "" {
		return fmt.Errorf("workspace_id is required")
	}

	r.Status = values.Get("status")

	if limitStr := values.Get("limit"); limitStr != "" {
		var err error
		r.Limit, err = ParseIntParam(limitStr)
		if err != nil {
			return fmt.Errorf("invalid limit parameter: %w", err)
		}
	}

	if offsetStr := values.Get("offset"); offsetStr != "" {
		var err error
		r.Offset, err = ParseIntParam(offsetStr)
		if err != nil {
			return fmt.Errorf("invalid offset parameter: %w", err)
		}
	}

	if withTemplatesStr := values.Get("with_templates"); withTemplatesStr != "" {
		var err error
		r.WithTemplates, err = ParseBoolParam(withTemplatesStr)
		if err != nil {
			return fmt.Errorf("invalid with_templates parameter: %w", err)
		}
	}

	return nil
}

// parseIntParam parses a string to an integer
func ParseIntParam(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// parseBoolParam parses a string to a boolean
func ParseBoolParam(s string) (bool, error) {
	var result bool
	_, err := fmt.Sscanf(s, "%t", &result)
	if err != nil {
		return false, err
	}
	return result, nil
}

// GetBroadcastRequest is used to extract query parameters for getting a single broadcast
type GetBroadcastRequest struct {
	WorkspaceID   string `json:"workspace_id"`
	ID            string `json:"id"`
	WithTemplates bool   `json:"with_templates,omitempty"`
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

	// DeleteBroadcast deletes a broadcast
	DeleteBroadcast(ctx context.Context, request *DeleteBroadcastRequest) error

	// SendToIndividual sends a broadcast to an individual recipient
	SendToIndividual(ctx context.Context, request *SendToIndividualRequest) error

	// SendWinningVariation sends the winning variation of an A/B test to remaining recipients
	SendWinningVariation(ctx context.Context, request *SendWinningVariationRequest) error
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

	// DeleteBroadcast deletes a broadcast
	DeleteBroadcast(ctx context.Context, workspaceID, id string) error
}

// ErrBroadcastNotFound is an error type for when a broadcast is not found
type ErrBroadcastNotFound struct {
	ID string
}

// Error returns the error message
func (e *ErrBroadcastNotFound) Error() string {
	return fmt.Sprintf("Broadcast not found with ID: %s", e.ID)
}

// SetTemplateForVariation assigns a template to a specific variation
func (b *Broadcast) SetTemplateForVariation(variationIndex int, template *Template) {
	if b == nil || variationIndex < 0 || variationIndex >= len(b.TestSettings.Variations) {
		return
	}

	b.TestSettings.Variations[variationIndex].Template = template
}
