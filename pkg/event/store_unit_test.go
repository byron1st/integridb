package event_test

import (
	"context"
	"errors"
	"testing"

	"github.com/byron1st/integridb/internal/mocks"
	"github.com/byron1st/integridb/pkg/event"
	"go.uber.org/mock/gomock"
)

// TestStore_AppendEvent_WithMockAdapter tests AppendEvent using gomock.
func TestStore_AppendEvent_WithMockAdapter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	payload := event.EventPayload{
		After: map[string]any{"id": "123", "name": "John"},
	}
	metadata := map[string]any{"user_id": "admin"}

	// Expect GetLastEvent to be called and return "not found" (first event)
	mockAdapter.EXPECT().
		GetLastEvent(ctx, "users", "123").
		Return(nil, event.ErrRowNotFound{TableName: "users", RowID: "123"})

	// Expect InsertEvent to be called once
	mockAdapter.EXPECT().
		InsertEvent(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, ev *event.Event) error {
			// Verify event properties
			if ev.Version != 1 {
				t.Errorf("Expected version 1, got %d", ev.Version)
			}
			if ev.PrevChecksum != "" {
				t.Errorf("Expected empty prev_checksum, got %q", ev.PrevChecksum)
			}
			if ev.TableName != "users" {
				t.Errorf("Expected table name 'users', got %q", ev.TableName)
			}
			if ev.RowID != "123" {
				t.Errorf("Expected row ID '123', got %q", ev.RowID)
			}
			if ev.EventType != event.EventTypeInsert {
				t.Errorf("Expected event type INSERT, got %v", ev.EventType)
			}
			return nil
		})

	ev, err := store.AppendEvent(ctx, "users", "123", event.EventTypeInsert, payload, metadata)
	if err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	if ev == nil {
		t.Fatal("AppendEvent() returned nil event")
	}
}

// TestStore_AppendEvent_SecondEvent tests appending a second event with hash chain linking.
func TestStore_AppendEvent_SecondEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	payload := event.EventPayload{
		Before:         map[string]any{"id": "123", "name": "John"},
		After:          map[string]any{"id": "123", "name": "Jane"},
		ChangedColumns: []string{"name"},
	}

	// Mock the last event
	lastEvent := &event.Event{
		ID:           "event-1",
		TableName:    "users",
		RowID:        "123",
		EventType:    event.EventTypeInsert,
		Version:      1,
		Checksum:     "abc123",
		PrevChecksum: "",
	}

	// Expect GetLastEvent to return the first event
	mockAdapter.EXPECT().
		GetLastEvent(ctx, "users", "123").
		Return(lastEvent, nil)

	// Expect InsertEvent to be called with version 2 and prev_checksum set
	mockAdapter.EXPECT().
		InsertEvent(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, ev *event.Event) error {
			if ev.Version != 2 {
				t.Errorf("Expected version 2, got %d", ev.Version)
			}
			if ev.PrevChecksum != "abc123" {
				t.Errorf("Expected prev_checksum 'abc123', got %q", ev.PrevChecksum)
			}
			return nil
		})

	ev, err := store.AppendEvent(ctx, "users", "123", event.EventTypeUpdate, payload, nil)
	if err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	if ev.Version != 2 {
		t.Errorf("Version = %d, want 2", ev.Version)
	}
	if ev.PrevChecksum != "abc123" {
		t.Errorf("PrevChecksum = %q, want 'abc123'", ev.PrevChecksum)
	}
}

// TestStore_AppendEvent_GetLastEventError tests error handling when GetLastEvent fails.
func TestStore_AppendEvent_GetLastEventError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	payload := event.EventPayload{
		After: map[string]any{"id": "123", "name": "John"},
	}

	// Simulate database error (not ErrRowNotFound)
	dbError := errors.New("database connection error")
	mockAdapter.EXPECT().
		GetLastEvent(ctx, "users", "123").
		Return(nil, dbError)

	_, err := store.AppendEvent(ctx, "users", "123", event.EventTypeInsert, payload, nil)
	if err == nil {
		t.Fatal("AppendEvent() expected error, got nil")
	}
}

// TestStore_AppendEvent_InsertEventError tests error handling when InsertEvent fails.
func TestStore_AppendEvent_InsertEventError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	payload := event.EventPayload{
		After: map[string]any{"id": "123", "name": "John"},
	}

	// First event, so GetLastEvent returns not found
	mockAdapter.EXPECT().
		GetLastEvent(ctx, "users", "123").
		Return(nil, event.ErrRowNotFound{TableName: "users", RowID: "123"})

	// Simulate insert error
	insertError := errors.New("insert failed")
	mockAdapter.EXPECT().
		InsertEvent(ctx, gomock.Any()).
		Return(insertError)

	_, err := store.AppendEvent(ctx, "users", "123", event.EventTypeInsert, payload, nil)
	if err == nil {
		t.Fatal("AppendEvent() expected error, got nil")
	}
}

// TestStore_GetEvents_WithMockAdapter tests GetEvents using gomock.
func TestStore_GetEvents_WithMockAdapter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	expectedEvents := []*event.Event{
		{ID: "1", Version: 1},
		{ID: "2", Version: 2},
		{ID: "3", Version: 3},
	}

	mockAdapter.EXPECT().
		GetEvents(ctx, "users", "123").
		Return(expectedEvents, nil)

	events, err := store.GetEvents(ctx, "users", "123")
	if err != nil {
		t.Fatalf("GetEvents() error = %v", err)
	}

	if len(events) != len(expectedEvents) {
		t.Errorf("GetEvents() returned %d events, want %d", len(events), len(expectedEvents))
	}
}

// TestStore_GetEvents_Error tests error handling in GetEvents.
func TestStore_GetEvents_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	dbError := errors.New("database error")
	mockAdapter.EXPECT().
		GetEvents(ctx, "users", "123").
		Return(nil, dbError)

	_, err := store.GetEvents(ctx, "users", "123")
	if err == nil {
		t.Fatal("GetEvents() expected error, got nil")
	}
}

// TestStore_GetLastEvent_WithMockAdapter tests GetLastEvent using gomock.
func TestStore_GetLastEvent_WithMockAdapter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	expectedEvent := &event.Event{
		ID:      "event-5",
		Version: 5,
	}

	mockAdapter.EXPECT().
		GetLastEvent(ctx, "users", "123").
		Return(expectedEvent, nil)

	ev, err := store.GetLastEvent(ctx, "users", "123")
	if err != nil {
		t.Fatalf("GetLastEvent() error = %v", err)
	}

	if ev.ID != expectedEvent.ID {
		t.Errorf("GetLastEvent() ID = %q, want %q", ev.ID, expectedEvent.ID)
	}
	if ev.Version != expectedEvent.Version {
		t.Errorf("GetLastEvent() Version = %d, want %d", ev.Version, expectedEvent.Version)
	}
}

// TestStore_GetLastEvent_NotFound tests GetLastEvent when row doesn't exist.
func TestStore_GetLastEvent_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	mockAdapter.EXPECT().
		GetLastEvent(ctx, "users", "123").
		Return(nil, event.ErrRowNotFound{TableName: "users", RowID: "123"})

	_, err := store.GetLastEvent(ctx, "users", "123")
	if err == nil {
		t.Fatal("GetLastEvent() expected error, got nil")
	}

	var errNotFound event.ErrRowNotFound
	if !errors.As(err, &errNotFound) {
		t.Errorf("GetLastEvent() error type = %T, want ErrRowNotFound", err)
	}
}

// TestStore_GetEventsByTable_WithMockAdapter tests GetEventsByTable using gomock.
func TestStore_GetEventsByTable_WithMockAdapter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAdapter := mocks.NewMockAdapter(ctrl)
	store := event.NewStore(mockAdapter)
	ctx := context.Background()

	expectedEvents := []*event.Event{
		{ID: "1", RowID: "123"},
		{ID: "2", RowID: "456"},
		{ID: "3", RowID: "789"},
	}

	mockAdapter.EXPECT().
		GetEventsByTable(ctx, "users").
		Return(expectedEvents, nil)

	events, err := store.GetEventsByTable(ctx, "users")
	if err != nil {
		t.Fatalf("GetEventsByTable() error = %v", err)
	}

	if len(events) != len(expectedEvents) {
		t.Errorf("GetEventsByTable() returned %d events, want %d", len(events), len(expectedEvents))
	}
}
