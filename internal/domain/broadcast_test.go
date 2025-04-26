package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
			Type: "conditions",
			SegmentConditions: domain.MapOfAny{
				"operator": "and",
				"conditions": []interface{}{
					map[string]interface{}{
						"field":    "email",
						"operator": "not_blank",
					},
				},
			},
			ExcludeUnsubscribed: true,
			SkipDuplicateEmails: true,
		},
		Schedule: domain.ScheduleSettings{
			SendImmediately: true,
		},
		TestSettings: domain.BroadcastTestSettings{
			Enabled: false,
		},
		TrackingEnabled: true,
		CreatedAt:       now,
		UpdatedAt:       now,
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
				ID:              "variation1",
				Name:            "Variation A",
				TemplateID:      "template123",
				TemplateVersion: 1,
				Subject:         "Test Subject A",
				FromName:        "Sender A",
				FromEmail:       "sender@example.com",
			},
			{
				ID:              "variation2",
				Name:            "Variation B",
				TemplateID:      "template123",
				TemplateVersion: 1,
				Subject:         "Test Subject B",
				FromName:        "Sender B",
				FromEmail:       "sender@example.com",
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
			name: "missing audience type",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Audience.Type = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "invalid audience type",
		},
		{
			name: "invalid audience type",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Audience.Type = "invalid"
				return b
			}(),
			wantErr: true,
			errMsg:  "invalid audience type",
		},
		{
			name: "audience type conditions without segment conditions",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Audience.Type = "conditions"
				b.Audience.SegmentConditions = nil
				return b
			}(),
			wantErr: true,
			errMsg:  "segment conditions are required",
		},
		{
			name: "audience type import without recipients",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Audience.Type = "import"
				b.Audience.ImportedRecipients = []string{}
				return b
			}(),
			wantErr: true,
			errMsg:  "imported recipients are required",
		},
		{
			name: "audience type individual without recipient",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Audience.Type = "individual"
				b.Audience.IndividualRecipient = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "individual recipient is required",
		},
		{
			name: "scheduled time required when not sending immediately",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Schedule.SendImmediately = false
				return b
			}(),
			wantErr: true,
			errMsg:  "scheduled time is required",
		},
		{
			name: "invalid time window format",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcast()
				b.Schedule.UseRecipientTimezone = true
				b.Schedule.TimeWindowStart = "9:00" // Missing leading zero
				b.Schedule.TimeWindowEnd = "17:00"
				return b
			}(),
			wantErr: true,
			errMsg:  "time window must be in HH:MM format",
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
						ID:              "variation" + string(rune(i+49)),
						Name:            "Variation " + string(rune(i+65)),
						TemplateID:      "template123",
						TemplateVersion: 1,
						Subject:         "Test Subject " + string(rune(i+65)),
						FromName:        "Sender " + string(rune(i+65)),
						FromEmail:       "sender@example.com",
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
			name: "missing subject in variation",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.Variations[0].Subject = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "subject is required for variation",
		},
		{
			name: "missing from name in variation",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.Variations[0].FromName = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "from_name is required for variation",
		},
		{
			name: "missing from email in variation",
			broadcast: func() domain.Broadcast {
				b := createValidBroadcastWithTest()
				b.TestSettings.Variations[0].FromEmail = ""
				return b
			}(),
			wantErr: true,
			errMsg:  "from_email is required for variation",
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
					Type: "conditions",
					SegmentConditions: domain.MapOfAny{
						"operator": "and",
						"conditions": []interface{}{
							map[string]interface{}{
								"field":    "email",
								"operator": "not_blank",
							},
						},
					},
					ExcludeUnsubscribed: true,
				},
				Schedule: domain.ScheduleSettings{
					SendImmediately: true,
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
					Type: "conditions",
					SegmentConditions: domain.MapOfAny{
						"operator": "and",
						"conditions": []interface{}{
							map[string]interface{}{
								"field":    "email",
								"operator": "not_blank",
							},
						},
					},
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
					Type: "conditions",
					SegmentConditions: domain.MapOfAny{
						"operator": "and",
						"conditions": []interface{}{
							map[string]interface{}{
								"field":    "email",
								"operator": "not_blank",
							},
						},
					},
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
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
				ScheduledAt: time.Now().Add(time.Hour),
				SendNow:     false,
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
				ID:          "broadcast123",
				ScheduledAt: time.Now().Add(time.Hour),
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing broadcast ID",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID: "workspace123",
				ScheduledAt: time.Now().Add(time.Hour),
			},
			wantErr: true,
			errMsg:  "broadcast id is required",
		},
		{
			name: "missing scheduled time when not sending now",
			request: domain.ScheduleBroadcastRequest{
				WorkspaceID: "workspace123",
				ID:          "broadcast123",
				SendNow:     false,
			},
			wantErr: true,
			errMsg:  "scheduled_at is required",
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
				WorkspaceID: "workspace123",
				BroadcastID: "broadcast123",
				RecipientID: "recipient123",
			},
			wantErr: false,
		},
		{
			name: "valid request with variation",
			request: domain.SendToIndividualRequest{
				WorkspaceID: "workspace123",
				BroadcastID: "broadcast123",
				RecipientID: "recipient123",
				VariationID: "variation1",
			},
			wantErr: false,
		},
		{
			name: "missing workspace ID",
			request: domain.SendToIndividualRequest{
				BroadcastID: "broadcast123",
				RecipientID: "recipient123",
			},
			wantErr: true,
			errMsg:  "workspace_id is required",
		},
		{
			name: "missing broadcast ID",
			request: domain.SendToIndividualRequest{
				WorkspaceID: "workspace123",
				RecipientID: "recipient123",
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
			errMsg:  "recipient_id is required",
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
