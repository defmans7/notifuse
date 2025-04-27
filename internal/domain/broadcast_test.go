package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Notifuse/notifuse/internal/domain"
)

func TestBroadcastStatus_Values(t *testing.T) {
	// Verify all status constants are defined
	assert.Equal(t, domain.BroadcastStatus("draft"), domain.BroadcastStatusDraft)
	assert.Equal(t, domain.BroadcastStatus("scheduled"), domain.BroadcastStatusScheduled)
	assert.Equal(t, domain.BroadcastStatus("sending"), domain.BroadcastStatusSending)
	assert.Equal(t, domain.BroadcastStatus("paused"), domain.BroadcastStatusPaused)
	assert.Equal(t, domain.BroadcastStatus("sent"), domain.BroadcastStatusSent)
	assert.Equal(t, domain.BroadcastStatus("cancelled"), domain.BroadcastStatusCancelled)
	assert.Equal(t, domain.BroadcastStatus("failed"), domain.BroadcastStatusFailed)
}

func TestTestWinnerMetric_Values(t *testing.T) {
	// Verify all metric constants are defined
	assert.Equal(t, domain.TestWinnerMetric("open_rate"), domain.TestWinnerMetricOpenRate)
	assert.Equal(t, domain.TestWinnerMetric("click_rate"), domain.TestWinnerMetricClickRate)
}

func createValidBroadcast() domain.Broadcast {
	now := time.Now()
	return domain.Broadcast{
		ID:          "broadcast123",
		WorkspaceID: "workspace123",
		Name:        "Test Newsletter",
		Status:      domain.BroadcastStatusDraft,
		Audience: domain.AudienceSettings{
			Lists:               []string{"list123"},
			ExcludeUnsubscribed: true,
			SkipDuplicateEmails: true,
		},
		Schedule: domain.ScheduleSettings{
			IsScheduled: false,
		},
		TestSettings: domain.BroadcastTestSettings{
			Enabled: false,
		},
		TrackingEnabled:   true,
		TotalSent:         100,
		TotalDelivered:    95,
		TotalFailed:       2,
		TotalBounced:      3,
		TotalComplained:   1,
		TotalOpens:        80,
		TotalClicks:       50,
		TotalUnsubscribed: 5,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func createValidBroadcastWithTest() domain.Broadcast {
	broadcast := createValidBroadcast()
	broadcast.TestSettings = domain.BroadcastTestSettings{
		Enabled:              true,
		SamplePercentage:     20,
		AutoSendWinner:       true,
		AutoSendWinnerMetric: domain.TestWinnerMetricOpenRate,
		TestDurationHours:    24,
		Variations: []domain.BroadcastVariation{
			{
				ID:         "variation1",
				TemplateID: "template123",
				Metrics: &domain.VariationMetrics{
					Recipients:   50,
					Delivered:    48,
					Opens:        40,
					Clicks:       25,
					Bounced:      1,
					Complained:   1,
					Unsubscribed: 2,
				},
			},
			{
				ID:         "variation2",
				TemplateID: "template123",
				Metrics: &domain.VariationMetrics{
					Recipients:   50,
					Delivered:    47,
					Opens:        35,
					Clicks:       20,
					Bounced:      2,
					Complained:   0,
					Unsubscribed: 3,
				},
			},
		},
	}
	return broadcast
}

func TestBroadcast_Validate(t *testing.T) {
	tests := []struct {
		name      string
		broadcast domain.Broadcast
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid broadcast",
			broadcast: createValidBroadcast(),
			wantErr:   false,
		},
		{
			name:      "valid broadcast with A/B test",
			broadcast: createValidBroadcastWithTest(),
			wantErr:   false,
		},
		{
			name: "missing workspace ID",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.WorkspaceID = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing name",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Name = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Name = string(make([]rune, 256))
				return b
			}(),
			wantErr: true,
			errMsg:  "name must be less than 255 characters",
		},
		{
			name: "invalid status",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Status = "invalid"
				return b
			}(),
			wantErr: true,
			errMsg:  "invalid broadcast status",
		},
		{
			name: "missing audience selection",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Audience.Lists = []string{}
				b.Audience.Segments = []string{}
				return b
			}(),
			wantErr: true,
			errMsg:  "either lists or segments must be specified",
		},
		{
			name: "both lists and segments specified",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Audience.Lists = []string{"list1"}
				b.Audience.Segments = []string{"segment1"}
				return b
			}(),
			wantErr: true,
			errMsg:  "both lists and segments are specified",
		},
		{
			name: "scheduled time required when not sending immediately",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Schedule.IsScheduled = true
				return b
			}(),
			wantErr: true,
			errMsg:  "scheduled date and time are required",
		},
		{
			name: "invalid date format",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Schedule.IsScheduled = true
				b.Schedule.ScheduledDate = "05/15/2023" // Wrong format
				b.Schedule.ScheduledTime = "14:30"
				return b
			}(),
			wantErr: true,
			errMsg:  "scheduled date must be in YYYY-MM-DD format",
		},
		{
			name: "invalid time format",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Schedule.IsScheduled = true
				b.Schedule.ScheduledDate = "2023-05-15"
				b.Schedule.ScheduledTime = "2:30" // Missing leading zero
				return b
			}(),
			wantErr: true,
			errMsg:  "scheduled time must be in HH:MM format",
		},
		{
			name: "invalid timezone",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Schedule.IsScheduled = true
				b.Schedule.ScheduledDate = "2023-05-15"
				b.Schedule.ScheduledTime = "14:30"
				b.Schedule.Timezone = "Invalid/Timezone"
				return b
			}(),
			wantErr: true,
			errMsg:  "invalid timezone",
		},
		{
			name: "test percentage too low",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.SamplePercentage = 0
				return b
			}(),
			wantErr: true,
			errMsg:  "test sample percentage must be between 1 and 100",
		},
		{
			name: "test percentage too high",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.SamplePercentage = 101
				return b
			}(),
			wantErr: true,
			errMsg:  "test sample percentage must be between 1 and 100",
		},
		{
			name: "not enough test variations",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.Variations = b.TestSettings.Variations[:1]
				return b
			}(),
			wantErr: true,
			errMsg:  "at least 2 variations are required",
		},
		{
			name: "too many test variations",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				// Create 9 variations (exceeding the 8 maximum)
				variations := make([]domain.BroadcastVariation, 9)
				for i := 0; i < 9; i++ {
					variations[i] = domain.BroadcastVariation{
						ID:         "variation" + string(rune(i+49)),
						TemplateID: "template123",
					}
				}
				b.TestSettings.Variations = variations
				return b
			}(),
			wantErr: true,
			errMsg:  "maximum 8 variations are allowed",
		},
		{
			name: "invalid test winner metric",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.AutoSendWinnerMetric = "invalid"
				return b
			}(),
			wantErr: true,
			errMsg:  "invalid test winner metric",
		},
		{
			name: "test duration must be positive",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.TestDurationHours = 0
				return b
			}(),
			wantErr: true,
			errMsg:  "test duration must be greater than 0 hours",
		},
		{
			name: "missing template ID in variation",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.Variations[0].TemplateID = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "template_id is required for variation",
		},
		{
			name: "tracking must be enabled for auto-send winner",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TrackingEnabled = false
				return b
			}(),
			wantErr: true,
			errMsg:  "tracking must be enabled to use auto-send winner feature",
		},
		{
			name: "valid scheduled broadcast",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Schedule.IsScheduled = true
				b.Schedule.ScheduledDate = "2023-05-15"
				b.Schedule.ScheduledTime = "14:30"
				b.Schedule.Timezone = "America/New_York"
				return b
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.broadcast.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateBroadcastRequest_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		request domain.CreateBroadcastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: domain.CreateBroadcastRequest{
				WorkspaceID: "workspace123",
				Name:        "Test Newsletter",
				Audience: domain.AudienceSettings{
					Lists:               []string{"list123"},
					ExcludeUnsubscribed: true,
				},
				Schedule: domain.ScheduleSettings{
					IsScheduled: false,
				},
				TestSettings: domain.BroadcastTestSettings{
					Enabled: false,
				},
				TrackingEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: domain.CreateBroadcastRequest{
				Name: "Test Newsletter",
				Audience: domain.AudienceSettings{
					Lists: []string{"list123"},
				},
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing name",
			request: domain.CreateBroadcastRequest{
				WorkspaceID: "workspace123",
				Audience: domain.AudienceSettings{
					Lists: []string{"list123"},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcast, err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, broadcast)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, broadcast)
				assert.Equal(t, tt.request.WorkspaceID, broadcast.WorkspaceID)
				assert.Equal(t, tt.request.Name, broadcast.Name)
				assert.Equal(t, domain.BroadcastStatusDraft, broadcast.Status)
				assert.WithinDuration(t, now, broadcast.CreatedAt, 5*time.Second)
				assert.WithinDuration(t, now, broadcast.UpdatedAt, 5*time.Second)
			}
		})
	}
}

func TestUpdateBroadcastRequest_Validate(t *testing.T) {
	existingBroadcast := createValidBroadcast()

	tests := []struct {
		name     string
		request  domain.UpdateBroadcastRequest
		existing domain.Broadcast
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid update for draft status",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID:  existingBroadcast.WorkspaceID,
				ID:           existingBroadcast.ID,
				Name:         "Updated Newsletter",
				Audience:     existingBroadcast.Audience,
				Schedule:     existingBroadcast.Schedule,
				TestSettings: existingBroadcast.TestSettings,
			},
			existing: existingBroadcast, // default is draft status
			wantErr:  false,
		},
		{
			name: "valid update for scheduled status",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID:  existingBroadcast.WorkspaceID,
				ID:           existingBroadcast.ID,
				Name:         "Updated Newsletter",
				Audience:     existingBroadcast.Audience,
				Schedule:     existingBroadcast.Schedule,
				TestSettings: existingBroadcast.TestSettings,
			},
			existing: func() domain.Broadcast {
				b := existingBroadcast
				b.Status = domain.BroadcastStatusScheduled
				return b
			}(),
			wantErr: false,
		},
		{
			name: "valid update for paused status",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID:  existingBroadcast.WorkspaceID,
				ID:           existingBroadcast.ID,
				Name:         "Updated Newsletter",
				Audience:     existingBroadcast.Audience,
				Schedule:     existingBroadcast.Schedule,
				TestSettings: existingBroadcast.TestSettings,
			},
			existing: func() domain.Broadcast {
				b := existingBroadcast
				b.Status = domain.BroadcastStatusPaused
				return b
			}(),
			wantErr: false,
		},
		{
			name: "workspace ID mismatch",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID: "different-workspace",
				ID:          existingBroadcast.ID,
				Name:        "Updated Newsletter",
			},
			existing: existingBroadcast,
			wantErr:  true,
			errMsg:   "workspace_id cannot be changed",
		},
		{
			name: "broadcast ID mismatch",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID: existingBroadcast.WorkspaceID,
				ID:          "different-id",
				Name:        "Updated Newsletter",
			},
			existing: existingBroadcast,
			wantErr:  true,
			errMsg:   "broadcast id cannot be changed",
		},
		{
			name: "cannot update sent broadcast",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID: existingBroadcast.WorkspaceID,
				ID:          existingBroadcast.ID,
				Name:        "Updated Newsletter",
			},
			existing: func() domain.Broadcast {
				b := existingBroadcast
				b.Status = domain.BroadcastStatusSent
				return b
			}(),
			wantErr: true,
			errMsg:  "cannot update broadcast with status: sent",
		},
		{
			name: "cannot update sending broadcast",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID: existingBroadcast.WorkspaceID,
				ID:          existingBroadcast.ID,
				Name:        "Updated Newsletter",
			},
			existing: func() domain.Broadcast {
				b := existingBroadcast
				b.Status = domain.BroadcastStatusSending
				return b
			}(),
			wantErr: true,
			errMsg:  "cannot update broadcast with status: sending",
		},
		{
			name: "cannot update cancelled broadcast",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID: existingBroadcast.WorkspaceID,
				ID:          existingBroadcast.ID,
				Name:        "Updated Newsletter",
			},
			existing: func() domain.Broadcast {
				b := existingBroadcast
				b.Status = domain.BroadcastStatusCancelled
				return b
			}(),
			wantErr: true,
			errMsg:  "cannot update broadcast with status: cancelled",
		},
		{
			name: "cannot update failed broadcast",
			request: domain.UpdateBroadcastRequest{
				WorkspaceID: existingBroadcast.WorkspaceID,
				ID:          existingBroadcast.ID,
				Name:        "Updated Newsletter",
			},
			existing: func() domain.Broadcast {
				b := existingBroadcast
				b.Status = domain.BroadcastStatusFailed
				return b
			}(),
			wantErr: true,
			errMsg:  "cannot update broadcast with status: failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcast, err := tt.request.Validate(&tt.existing)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, broadcast)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, broadcast)
				assert.Equal(t, tt.request.Name, broadcast.Name)
				assert.WithinDuration(t, time.Now(), broadcast.UpdatedAt, 5*time.Second)
			}
		})
	}
}

func TestScheduleBroadcastRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request domain.ScheduleBroadcastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with scheduled time",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID:          "workspace123",
				ID:                   "broadcast123",
				IsScheduled:          true,
				ScheduledDate:        "2023-12-31",
				ScheduledTime:        "15:30",
				Timezone:             "UTC",
				UseRecipientTimezone: false,
				SendNow:              false,
			},
			wantErr: false,
		},
		{
			name: "valid request with send now",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
				SendNow:     true,
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: domain.ScheduleBroadcastRequest{
				ID:            "broadcast123",
				IsScheduled:   true,
				ScheduledDate: "2023-12-31",
				ScheduledTime: "15:30",
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing broadcast ID",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID:   "workspace123",
				IsScheduled:   true,
				ScheduledDate: "2023-12-31",
				ScheduledTime: "15:30",
			},
			wantErr: true,
			errMsg:  "broadcast id is required",
		},
		{
			name: "missing scheduled fields when not sending now",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
				SendNow:     false,
			},
			wantErr: true,
			errMsg:  "either send_now or is_scheduled must be true",
		},
		{
			name: "is_scheduled is true but missing date/time",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
				IsScheduled: true,
			},
			wantErr: true,
			errMsg:  "scheduled_date and scheduled_time are required",
		},
		{
			name: "invalid date format",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID:   "workspace123",
				ID:            "broadcast123",
				IsScheduled:   true,
				ScheduledDate: "31-12-2023", // Wrong format, should be YYYY-MM-DD
				ScheduledTime: "15:30",
			},
			wantErr: true,
			errMsg:  "scheduled date must be in YYYY-MM-DD format",
		},
		{
			name: "invalid time format",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID:   "workspace123",
				ID:            "broadcast123",
				IsScheduled:   true,
				ScheduledDate: "2023-12-31",
				ScheduledTime: "3:30PM", // Wrong format, should be HH:MM
			},
			wantErr: true,
			errMsg:  "scheduled time must be in HH:MM format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPauseBroadcastRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request domain.PauseBroadcastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: domain.PauseBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: domain.PauseBroadcastRequest{
				ID: "broadcast123",
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing broadcast ID",
			request: domain.PauseBroadcastRequest{
				WorkspaceID: "workspace123",
			},
			wantErr: true,
			errMsg:  "broadcast id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResumeBroadcastRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request domain.ResumeBroadcastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: domain.ResumeBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: domain.ResumeBroadcastRequest{
				ID: "broadcast123",
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing broadcast ID",
			request: domain.ResumeBroadcastRequest{
				WorkspaceID: "workspace123",
			},
			wantErr: true,
			errMsg:  "broadcast id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCancelBroadcastRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request domain.CancelBroadcastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: domain.CancelBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: domain.CancelBroadcastRequest{
				ID: "broadcast123",
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing broadcast ID",
			request: domain.CancelBroadcastRequest{
				WorkspaceID: "workspace123",
			},
			wantErr: true,
			errMsg:  "broadcast id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSendToIndividualRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request domain.SendToIndividualRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: domain.SendToIndividualRequest{
				WorkspaceID:    "workspace123",
				BroadcastID:    "broadcast123",
				RecipientEmail: "recipient@123.com",
			},
			wantErr: false,
		},
		{
			name: "valid request with variation",
			request: domain.SendToIndividualRequest{
				WorkspaceID:    "workspace123",
				BroadcastID:    "broadcast123",
				RecipientEmail: "recipient@123.com",
				VariationID:    "variation1",
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: domain.SendToIndividualRequest{
				BroadcastID:    "broadcast123",
				RecipientEmail: "recipient@123.com",
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing broadcast ID",
			request: domain.SendToIndividualRequest{
				WorkspaceID:    "workspace123",
				RecipientEmail: "recipient@123.com",
			},
			wantErr: true,
			errMsg:  "broadcast_id is required",
		},
		{
			name: "missing recipient ID",
			request: domain.SendToIndividualRequest{
				WorkspaceID: "workspace123",
				BroadcastID: "broadcast123",
			},
			wantErr: true,
			errMsg:  "recipient_email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestErrBroadcastNotFound_Error(t *testing.T) {
	err := &domain.ErrBroadcastNotFound{ID: "broadcast123"}
	assert.Equal(t, "Broadcast not found with ID: broadcast123", err.Error())
}

// TestDeleteBroadcastRequestValidate tests the validation of DeleteBroadcastRequest
func TestDeleteBroadcastRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request domain.DeleteBroadcastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Request",
			request: domain.DeleteBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
			},
			wantErr: false,
		},
		{
			name: "Missing WorkspaceID",
			request: domain.DeleteBroadcastRequest{
				ID: "broadcast123",
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "Missing ID",
			request: domain.DeleteBroadcastRequest{
				WorkspaceID: "workspace123",
			},
			wantErr: true,
			errMsg:  "broadcast id is required",
		},
		{
			name:    "Empty Request",
			request: domain.DeleteBroadcastRequest{},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestScheduleSettings_ParseScheduledDateTime tests the ParseScheduledDateTime method
func TestScheduleSettings_ParseScheduledDateTime(t *testing.T) {
	tests := []struct {
		name     string
		settings domain.ScheduleSettings
		wantErr  bool
	}{
		{
			name: "basic date and time",
			settings: domain.ScheduleSettings{
				ScheduledDate: "2023-05-15",
				ScheduledTime: "14:30",
			},
			wantErr: false,
		},
		{
			name: "with timezone",
			settings: domain.ScheduleSettings{
				ScheduledDate: "2023-05-15",
				ScheduledTime: "14:30",
				Timezone:      "America/New_York",
			},
			wantErr: false,
		},
		{
			name: "empty date and time",
			settings: domain.ScheduleSettings{
				ScheduledDate: "",
				ScheduledTime: "",
			},
			wantErr: false,
		},
		{
			name: "invalid timezone",
			settings: domain.ScheduleSettings{
				ScheduledDate: "2023-05-15",
				ScheduledTime: "14:30",
				Timezone:      "Invalid/Timezone",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.settings.ParseScheduledDateTime()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if !got.IsZero() {
				// Check that hours and minutes match what was specified
				if tt.settings.ScheduledTime != "" {
					assert.Equal(t, tt.settings.ScheduledTime[:2], got.Format("15"))
					assert.Equal(t, tt.settings.ScheduledTime[3:], got.Format("04"))
				}

				// Verify that seconds are not zero in the parsed time
				assert.NotEqual(t, 0, got.Second()+got.Nanosecond(),
					"Expected non-zero seconds or nanoseconds in parsed time")
			}
		})
	}
}
