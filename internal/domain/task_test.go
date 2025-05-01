package domain

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskState_Value(t *testing.T) {
	t.Run("empty state", func(t *testing.T) {
		state := TaskState{}
		value, err := state.Value()
		require.NoError(t, err)

		// Convert value back to bytes for assertion
		bytes, ok := value.([]byte)
		require.True(t, ok, "Expected []byte, got %T", value)

		// Validate JSON
		var m map[string]interface{}
		err = json.Unmarshal(bytes, &m)
		require.NoError(t, err)

		// Should be an empty object
		assert.Empty(t, m)
	})

	t.Run("with common fields", func(t *testing.T) {
		state := TaskState{
			Progress: 50.5,
			Message:  "Half done",
		}
		value, err := state.Value()
		require.NoError(t, err)

		bytes, ok := value.([]byte)
		require.True(t, ok)

		var m map[string]interface{}
		err = json.Unmarshal(bytes, &m)
		require.NoError(t, err)

		assert.Equal(t, 50.5, m["progress"])
		assert.Equal(t, "Half done", m["message"])
	})

	t.Run("with specialized fields", func(t *testing.T) {
		state := TaskState{
			Progress: 75.0,
			Message:  "Processing broadcast",
			SendBroadcast: &SendBroadcastState{
				BroadcastID:     "broadcast-123",
				TotalRecipients: 1000,
				SentCount:       750,
				FailedCount:     10,
				ChannelType:     "email",
				RecipientOffset: 750,
				EndOffset:       1000,
			},
		}
		value, err := state.Value()
		require.NoError(t, err)

		bytes, ok := value.([]byte)
		require.True(t, ok)

		var m map[string]interface{}
		err = json.Unmarshal(bytes, &m)
		require.NoError(t, err)

		assert.Equal(t, 75.0, m["progress"])
		assert.Equal(t, "Processing broadcast", m["message"])

		// Check specialized fields
		broadcastMap, ok := m["send_broadcast"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "broadcast-123", broadcastMap["broadcast_id"])
		assert.Equal(t, float64(1000), broadcastMap["total_recipients"])
		assert.Equal(t, float64(750), broadcastMap["sent_count"])
		assert.Equal(t, float64(10), broadcastMap["failed_count"])
		assert.Equal(t, "email", broadcastMap["channel_type"])
		assert.Equal(t, float64(750), broadcastMap["recipient_offset"])
		assert.Equal(t, float64(1000), broadcastMap["end_offset"])
	})
}

func TestTaskState_Scan(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		var state TaskState
		err := state.Scan(nil)
		require.NoError(t, err)

		// Should result in empty task state
		assert.Equal(t, 0.0, state.Progress)
		assert.Equal(t, "", state.Message)
		assert.Nil(t, state.SendBroadcast)
	})

	t.Run("empty json", func(t *testing.T) {
		var state TaskState
		err := state.Scan([]byte(`{}`))
		require.NoError(t, err)

		// Should result in empty task state
		assert.Equal(t, 0.0, state.Progress)
		assert.Equal(t, "", state.Message)
		assert.Nil(t, state.SendBroadcast)
	})

	t.Run("with common fields", func(t *testing.T) {
		var state TaskState
		data := []byte(`{"progress": 42.5, "message": "Working on it"}`)

		err := state.Scan(data)
		require.NoError(t, err)

		assert.Equal(t, 42.5, state.Progress)
		assert.Equal(t, "Working on it", state.Message)
		assert.Nil(t, state.SendBroadcast)
	})

	t.Run("with specialized fields", func(t *testing.T) {
		var state TaskState
		data := []byte(`{
			"progress": 60.0, 
			"message": "Sending emails", 
			"send_broadcast": {
				"broadcast_id": "broadcast-456",
				"total_recipients": 500,
				"sent_count": 300,
				"failed_count": 5,
				"channel_type": "email",
				"recipient_offset": 300,
				"end_offset": 500
			}
		}`)

		err := state.Scan(data)
		require.NoError(t, err)

		assert.Equal(t, 60.0, state.Progress)
		assert.Equal(t, "Sending emails", state.Message)
		assert.NotNil(t, state.SendBroadcast)
		assert.Equal(t, "broadcast-456", state.SendBroadcast.BroadcastID)
		assert.Equal(t, 500, state.SendBroadcast.TotalRecipients)
		assert.Equal(t, 300, state.SendBroadcast.SentCount)
		assert.Equal(t, 5, state.SendBroadcast.FailedCount)
		assert.Equal(t, "email", state.SendBroadcast.ChannelType)
		assert.Equal(t, int64(300), state.SendBroadcast.RecipientOffset)
		assert.Equal(t, int64(500), state.SendBroadcast.EndOffset)
	})

	t.Run("invalid type", func(t *testing.T) {
		var state TaskState
		err := state.Scan(123)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected []byte")
	})

	t.Run("invalid json", func(t *testing.T) {
		var state TaskState
		err := state.Scan([]byte(`{not valid json`))
		require.Error(t, err)
	})
}

func TestGetTaskRequest_FromURLParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		values := url.Values{
			"workspace_id": []string{"ws-123"},
			"id":           []string{"task-456"},
		}

		req := &GetTaskRequest{}
		err := req.FromURLParams(values)
		require.NoError(t, err)

		assert.Equal(t, "ws-123", req.WorkspaceID)
		assert.Equal(t, "task-456", req.ID)
	})

	t.Run("missing params", func(t *testing.T) {
		values := url.Values{
			"workspace_id": []string{"ws-123"},
			// Missing ID
		}

		req := &GetTaskRequest{}
		err := req.FromURLParams(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})
}

func TestDeleteTaskRequest_FromURLParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		values := url.Values{
			"workspace_id": []string{"ws-123"},
			"id":           []string{"task-456"},
		}

		req := &DeleteTaskRequest{}
		err := req.FromURLParams(values)
		require.NoError(t, err)

		assert.Equal(t, "ws-123", req.WorkspaceID)
		assert.Equal(t, "task-456", req.ID)
	})

	t.Run("missing workspace", func(t *testing.T) {
		values := url.Values{
			// Missing workspace_id
			"id": []string{"task-456"},
		}

		req := &DeleteTaskRequest{}
		err := req.FromURLParams(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "workspace_id is required")
	})
}

func TestListTasksRequest_FromURLParams(t *testing.T) {
	t.Run("full params", func(t *testing.T) {
		values := url.Values{
			"workspace_id":   []string{"ws-123"},
			"status":         []string{"pending,running"},
			"type":           []string{"broadcast,import"},
			"created_after":  []string{"2023-01-01T00:00:00Z"},
			"created_before": []string{"2023-12-31T23:59:59Z"},
			"limit":          []string{"50"},
			"offset":         []string{"10"},
		}

		req := &ListTasksRequest{}
		err := req.FromURLParams(values)
		require.NoError(t, err)

		assert.Equal(t, "ws-123", req.WorkspaceID)
		assert.Equal(t, []string{"pending", "running"}, req.Status)
		assert.Equal(t, []string{"broadcast", "import"}, req.Type)
		assert.Equal(t, "2023-01-01T00:00:00Z", req.CreatedAfter)
		assert.Equal(t, "2023-12-31T23:59:59Z", req.CreatedBefore)
		assert.Equal(t, 50, req.Limit)
		assert.Equal(t, 10, req.Offset)
	})

	t.Run("minimal params", func(t *testing.T) {
		values := url.Values{
			"workspace_id": []string{"ws-123"},
			// No optional params
		}

		req := &ListTasksRequest{}
		err := req.FromURLParams(values)
		require.NoError(t, err)

		assert.Equal(t, "ws-123", req.WorkspaceID)
		assert.Empty(t, req.Status)
		assert.Empty(t, req.Type)
		assert.Empty(t, req.CreatedAfter)
		assert.Empty(t, req.CreatedBefore)
		assert.Equal(t, 0, req.Limit)  // Default value
		assert.Equal(t, 0, req.Offset) // Default value
	})

	t.Run("invalid limit/offset", func(t *testing.T) {
		values := url.Values{
			"workspace_id": []string{"ws-123"},
			"limit":        []string{"not-a-number"},
			"offset":       []string{"also-not-a-number"},
		}

		req := &ListTasksRequest{}
		err := req.FromURLParams(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid limit")
	})
}

func TestListTasksRequest_ToFilter(t *testing.T) {
	t.Run("convert all fields", func(t *testing.T) {
		// Create a request with all fields populated
		req := &ListTasksRequest{
			WorkspaceID:   "ws-123",
			Status:        []string{"pending", "running"},
			Type:          []string{"broadcast", "import"},
			CreatedAfter:  "2023-01-01T00:00:00Z",
			CreatedBefore: "2023-12-31T23:59:59Z",
			Limit:         50,
			Offset:        10,
		}

		filter := req.ToFilter()

		// Check that statuses were converted properly
		assert.Len(t, filter.Status, 2)
		assert.Contains(t, filter.Status, TaskStatus("pending"))
		assert.Contains(t, filter.Status, TaskStatus("running"))

		// Check other fields
		assert.Equal(t, []string{"broadcast", "import"}, filter.Type)
		assert.Equal(t, 50, filter.Limit)
		assert.Equal(t, 10, filter.Offset)

		// Check time conversions
		require.NotNil(t, filter.CreatedAfter)
		require.NotNil(t, filter.CreatedBefore)

		expectedStartTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		expectedEndTime, _ := time.Parse(time.RFC3339, "2023-12-31T23:59:59Z")

		assert.Equal(t, expectedStartTime, *filter.CreatedAfter)
		assert.Equal(t, expectedEndTime, *filter.CreatedBefore)
	})

	t.Run("minimal fields", func(t *testing.T) {
		// Create a request with minimal fields
		req := &ListTasksRequest{
			WorkspaceID: "ws-123",
			// No optional params
		}

		filter := req.ToFilter()

		// Check defaults
		assert.Empty(t, filter.Status)
		assert.Empty(t, filter.Type)
		assert.Nil(t, filter.CreatedAfter)
		assert.Nil(t, filter.CreatedBefore)
		assert.Equal(t, 100, filter.Limit)
		assert.Equal(t, 0, filter.Offset)
	})

	t.Run("invalid time format", func(t *testing.T) {
		// Create a request with invalid time format
		req := &ListTasksRequest{
			WorkspaceID:  "ws-123",
			CreatedAfter: "not-a-valid-time",
		}

		filter := req.ToFilter()

		// Time parsing should fail silently, returning nil
		assert.Nil(t, filter.CreatedAfter)
	})
}

func TestSplitAndTrim(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		result := splitAndTrim("")
		assert.Empty(t, result)
	})

	t.Run("single value", func(t *testing.T) {
		result := splitAndTrim("value")
		assert.Equal(t, []string{"value"}, result)
	})

	t.Run("multiple values", func(t *testing.T) {
		result := splitAndTrim("one,two,three")
		assert.Equal(t, []string{"one", "two", "three"}, result)
	})

	t.Run("with spaces", func(t *testing.T) {
		result := splitAndTrim(" one , two , three ")
		assert.Equal(t, []string{"one", "two", "three"}, result)
	})

	t.Run("with empty segments", func(t *testing.T) {
		result := splitAndTrim("one,,three")
		assert.Equal(t, []string{"one", "three"}, result)
	})
}

func TestExecutePendingTasksRequest_FromURLParams(t *testing.T) {
	t.Run("with max_tasks", func(t *testing.T) {
		values := url.Values{
			"max_tasks": []string{"20"},
		}

		req := &ExecutePendingTasksRequest{}
		err := req.FromURLParams(values)
		require.NoError(t, err)

		assert.Equal(t, 20, req.MaxTasks)
	})

	t.Run("without max_tasks", func(t *testing.T) {
		values := url.Values{}

		req := &ExecutePendingTasksRequest{}
		err := req.FromURLParams(values)
		require.NoError(t, err)

		// The implementation uses a default value of 10
		assert.Equal(t, 10, req.MaxTasks) // Default value is 10 in the implementation
	})

	t.Run("invalid max_tasks", func(t *testing.T) {
		values := url.Values{
			"max_tasks": []string{"not-a-number"},
		}

		req := &ExecutePendingTasksRequest{}
		err := req.FromURLParams(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid max_tasks")
	})

	// The implementation doesn't validate for negative max_tasks values
	t.Run("negative max_tasks", func(t *testing.T) {
		values := url.Values{
			"max_tasks": []string{"-10"},
		}

		req := &ExecutePendingTasksRequest{}
		err := req.FromURLParams(values)

		// There's no validation for negative max_tasks in the implementation
		require.NoError(t, err)
		assert.Equal(t, -10, req.MaxTasks)
	})
}

func TestExecuteTaskRequest_Validate(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := &ExecuteTaskRequest{
			WorkspaceID: "ws-123",
			ID:          "task-456",
		}

		err := req.Validate()
		require.NoError(t, err)
	})

	t.Run("missing workspace_id", func(t *testing.T) {
		req := &ExecuteTaskRequest{
			ID: "task-456",
		}

		err := req.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "workspace_id is required")
	})

	t.Run("missing id", func(t *testing.T) {
		req := &ExecuteTaskRequest{
			WorkspaceID: "ws-123",
		}

		err := req.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task id is required")
	})
}
