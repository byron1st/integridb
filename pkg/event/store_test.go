package event

import (
	"context"
	"testing"
	"time"

	"github.com/byron1st/integridb/internal/hash"
)

// mockAdapter is a simple in-memory adapter for testing.
type mockAdapter struct {
	events map[string][]*Event // key: tableName:rowID, value: events for that row
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{
		events: make(map[string][]*Event),
	}
}

func (m *mockAdapter) InsertEvent(_ context.Context, event *Event) error {
	key := event.TableName + ":" + event.RowID
	m.events[key] = append(m.events[key], event)
	return nil
}

func (m *mockAdapter) GetEvents(_ context.Context, tableName, rowID string) ([]*Event, error) {
	key := tableName + ":" + rowID
	events := m.events[key]
	if events == nil {
		return []*Event{}, nil
	}
	return events, nil
}

func (m *mockAdapter) GetLastEvent(_ context.Context, tableName, rowID string) (*Event, error) {
	key := tableName + ":" + rowID
	events := m.events[key]
	if len(events) == 0 {
		return nil, ErrRowNotFound{TableName: tableName, RowID: rowID}
	}
	return events[len(events)-1], nil
}

func (m *mockAdapter) GetEventsByTable(_ context.Context, tableName string) ([]*Event, error) {
	var result []*Event
	for key, events := range m.events {
		// Simple check if key starts with tableName:
		if len(key) > len(tableName) && key[:len(tableName)] == tableName && key[len(tableName)] == ':' {
			result = append(result, events...)
		}
	}
	return result, nil
}

func (m *mockAdapter) GetEventByID(_ context.Context, eventID string) (*Event, error) {
	for _, events := range m.events {
		for _, event := range events {
			if event.ID == eventID {
				return event, nil
			}
		}
	}
	return nil, ErrEventNotFound{EventID: eventID}
}

func TestStore_AppendEvent_FirstEvent(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	payload := EventPayload{
		After: map[string]any{"id": "123", "name": "John"},
	}
	metadata := map[string]any{"user_id": "admin"}

	event, err := store.AppendEvent(ctx, "users", "123", EventTypeInsert, payload, metadata)
	if err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	// Check version is 1 for first event
	if event.Version != 1 {
		t.Errorf("Version = %d, want 1", event.Version)
	}

	// Check prev_checksum is empty for first event
	if event.PrevChecksum != "" {
		t.Errorf("PrevChecksum = %q, want empty", event.PrevChecksum)
	}

	// Check checksum is not empty
	if event.Checksum == "" {
		t.Error("Checksum is empty")
	}

	// Check event type
	if event.EventType != EventTypeInsert {
		t.Errorf("EventType = %v, want %v", event.EventType, EventTypeInsert)
	}

	// Check table and row
	if event.TableName != "users" {
		t.Errorf("TableName = %q, want %q", event.TableName, "users")
	}
	if event.RowID != "123" {
		t.Errorf("RowID = %q, want %q", event.RowID, "123")
	}
}

func TestStore_AppendEvent_MultipleEvents(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	// First event (INSERT)
	payload1 := EventPayload{
		After: map[string]any{"id": "123", "name": "John"},
	}
	event1, err := store.AppendEvent(ctx, "users", "123", EventTypeInsert, payload1, nil)
	if err != nil {
		t.Fatalf("AppendEvent() event1 error = %v", err)
	}

	// Second event (UPDATE)
	payload2 := EventPayload{
		Before:         map[string]any{"id": "123", "name": "John"},
		After:          map[string]any{"id": "123", "name": "Jane"},
		ChangedColumns: []string{"name"},
	}
	event2, err := store.AppendEvent(ctx, "users", "123", EventTypeUpdate, payload2, nil)
	if err != nil {
		t.Fatalf("AppendEvent() event2 error = %v", err)
	}

	// Check version incremented
	if event2.Version != 2 {
		t.Errorf("Event2 Version = %d, want 2", event2.Version)
	}

	// Check prev_checksum links to first event
	if event2.PrevChecksum != event1.Checksum {
		t.Errorf("Event2 PrevChecksum = %q, want %q", event2.PrevChecksum, event1.Checksum)
	}

	// Third event (DELETE)
	payload3 := EventPayload{
		Before: map[string]any{"id": "123", "name": "Jane"},
	}
	event3, err := store.AppendEvent(ctx, "users", "123", EventTypeDelete, payload3, nil)
	if err != nil {
		t.Fatalf("AppendEvent() event3 error = %v", err)
	}

	// Check version incremented
	if event3.Version != 3 {
		t.Errorf("Event3 Version = %d, want 3", event3.Version)
	}

	// Check prev_checksum links to second event
	if event3.PrevChecksum != event2.Checksum {
		t.Errorf("Event3 PrevChecksum = %q, want %q", event3.PrevChecksum, event2.Checksum)
	}
}

func TestStore_AppendEvent_HashChainIntegrity(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	// Create multiple events
	events := make([]*Event, 5)
	for i := range 5 {
		payload := EventPayload{
			After: map[string]any{"id": "123", "value": i},
		}
		event, err := store.AppendEvent(ctx, "test", "123", EventTypeUpdate, payload, nil)
		if err != nil {
			t.Fatalf("AppendEvent() event %d error = %v", i, err)
		}
		events[i] = event
	}

	// Verify hash chain
	for i := range events {
		if i == 0 {
			// First event should have empty prev_checksum
			if events[i].PrevChecksum != "" {
				t.Errorf("Event 0 PrevChecksum = %q, want empty", events[i].PrevChecksum)
			}
		} else {
			// Each event should link to previous
			if events[i].PrevChecksum != events[i-1].Checksum {
				t.Errorf("Event %d PrevChecksum = %q, want %q",
					i, events[i].PrevChecksum, events[i-1].Checksum)
			}
		}

		// Verify checksum is correct
		expectedChecksum, err := hash.Calculate(
			events[i].Payload,
			events[i].Metadata,
			events[i].PrevChecksum,
		)
		if err != nil {
			t.Fatalf("Calculate checksum for event %d error = %v", i, err)
		}
		if events[i].Checksum != expectedChecksum {
			t.Errorf("Event %d Checksum = %q, want %q", i, events[i].Checksum, expectedChecksum)
		}
	}
}

func TestStore_GetEvents(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	// Create multiple events
	for i := range 3 {
		payload := EventPayload{
			After: map[string]any{"id": "123", "value": i},
		}
		_, err := store.AppendEvent(ctx, "users", "123", EventTypeUpdate, payload, nil)
		if err != nil {
			t.Fatalf("AppendEvent() error = %v", err)
		}
	}

	// Get all events
	events, err := store.GetEvents(ctx, "users", "123")
	if err != nil {
		t.Fatalf("GetEvents() error = %v", err)
	}

	if len(events) != 3 {
		t.Errorf("GetEvents() returned %d events, want 3", len(events))
	}

	// Verify ordering by version
	for i, event := range events {
		if event.Version != int64(i+1) {
			t.Errorf("Event %d Version = %d, want %d", i, event.Version, i+1)
		}
	}
}

func TestStore_GetEventsInRange(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	// Create 10 events
	for i := range 10 {
		payload := EventPayload{
			After: map[string]any{"id": "123", "value": i},
		}
		_, err := store.AppendEvent(ctx, "users", "123", EventTypeUpdate, payload, nil)
		if err != nil {
			t.Fatalf("AppendEvent() error = %v", err)
		}
	}

	tests := []struct {
		name        string
		fromVersion int64
		toVersion   int64
		wantCount   int
		wantFirst   int64
		wantLast    int64
	}{
		{
			name:        "range 1-3",
			fromVersion: 1,
			toVersion:   3,
			wantCount:   3,
			wantFirst:   1,
			wantLast:    3,
		},
		{
			name:        "range 5-7",
			fromVersion: 5,
			toVersion:   7,
			wantCount:   3,
			wantFirst:   5,
			wantLast:    7,
		},
		{
			name:        "single version",
			fromVersion: 5,
			toVersion:   5,
			wantCount:   1,
			wantFirst:   5,
			wantLast:    5,
		},
		{
			name:        "all events",
			fromVersion: 1,
			toVersion:   10,
			wantCount:   10,
			wantFirst:   1,
			wantLast:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := store.GetEventsInRange(ctx, "users", "123", tt.fromVersion, tt.toVersion)
			if err != nil {
				t.Fatalf("GetEventsInRange() error = %v", err)
			}

			if len(events) != tt.wantCount {
				t.Errorf("GetEventsInRange() returned %d events, want %d", len(events), tt.wantCount)
			}

			if len(events) > 0 {
				if events[0].Version != tt.wantFirst {
					t.Errorf("First event Version = %d, want %d", events[0].Version, tt.wantFirst)
				}
				if events[len(events)-1].Version != tt.wantLast {
					t.Errorf("Last event Version = %d, want %d", events[len(events)-1].Version, tt.wantLast)
				}
			}
		})
	}
}

func TestStore_GetEventsByTime(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	// Create events with specific timestamps
	baseTime := time.Now().UTC()
	times := make([]time.Time, 5)
	for i := range 5 {
		payload := EventPayload{
			After: map[string]any{"id": "123", "value": i},
		}
		event, err := store.AppendEvent(ctx, "users", "123", EventTypeUpdate, payload, nil)
		if err != nil {
			t.Fatalf("AppendEvent() error = %v", err)
		}
		// Manually set timestamp for testing
		event.CreatedAt = baseTime.Add(time.Duration(i) * time.Hour)
		times[i] = event.CreatedAt
	}

	tests := []struct {
		name      string
		timestamp time.Time
		wantCount int
	}{
		{
			name:      "before first event",
			timestamp: baseTime.Add(-1 * time.Hour),
			wantCount: 0,
		},
		{
			name:      "at second event",
			timestamp: times[1],
			wantCount: 2,
		},
		{
			name:      "between third and fourth",
			timestamp: times[2].Add(30 * time.Minute),
			wantCount: 3,
		},
		{
			name:      "after all events",
			timestamp: baseTime.Add(10 * time.Hour),
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := store.GetEventsByTime(ctx, "users", "123", tt.timestamp)
			if err != nil {
				t.Fatalf("GetEventsByTime() error = %v", err)
			}

			if len(events) != tt.wantCount {
				t.Errorf("GetEventsByTime() returned %d events, want %d", len(events), tt.wantCount)
			}
		})
	}
}

func TestStore_GetLastEvent(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	// Test with no events
	_, err := store.GetLastEvent(ctx, "users", "123")
	if err == nil {
		t.Error("GetLastEvent() expected error for non-existent row, got nil")
	}

	// Create multiple events
	var lastEvent *Event
	for i := range 5 {
		payload := EventPayload{
			After: map[string]any{"id": "123", "value": i},
		}
		event, err := store.AppendEvent(ctx, "users", "123", EventTypeUpdate, payload, nil)
		if err != nil {
			t.Fatalf("AppendEvent() error = %v", err)
		}
		lastEvent = event
	}

	// Get last event
	event, err := store.GetLastEvent(ctx, "users", "123")
	if err != nil {
		t.Fatalf("GetLastEvent() error = %v", err)
	}

	if event.ID != lastEvent.ID {
		t.Errorf("GetLastEvent() ID = %q, want %q", event.ID, lastEvent.ID)
	}
	if event.Version != 5 {
		t.Errorf("GetLastEvent() Version = %d, want 5", event.Version)
	}
}

func TestStore_MultipleRows(t *testing.T) {
	adapter := newMockAdapter()
	store := NewStore(adapter)
	ctx := context.Background()

	// Create events for different rows
	rowIDs := []string{"123", "456", "789"}
	for _, rowID := range rowIDs {
		for i := range 3 {
			payload := EventPayload{
				After: map[string]any{"id": rowID, "value": i},
			}
			_, err := store.AppendEvent(ctx, "users", rowID, EventTypeUpdate, payload, nil)
			if err != nil {
				t.Fatalf("AppendEvent() error = %v", err)
			}
		}
	}

	// Verify each row has correct events
	for _, rowID := range rowIDs {
		events, err := store.GetEvents(ctx, "users", rowID)
		if err != nil {
			t.Fatalf("GetEvents() for row %s error = %v", rowID, err)
		}

		if len(events) != 3 {
			t.Errorf("Row %s has %d events, want 3", rowID, len(events))
		}

		// Verify versions are independent per row
		for i, event := range events {
			if event.Version != int64(i+1) {
				t.Errorf("Row %s Event %d Version = %d, want %d",
					rowID, i, event.Version, i+1)
			}
		}
	}
}
