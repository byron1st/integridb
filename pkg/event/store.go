package event

import (
	"context"
	"fmt"
	"time"

	"github.com/byron1st/integridb/internal/hash"
	"github.com/google/uuid"
)

// Adapter defines the interface for event storage backends.
type Adapter interface {
	// InsertEvent persists a new event to the event store.
	InsertEvent(ctx context.Context, event *Event) error

	// GetEvents retrieves all events for a specific table and row, ordered by version.
	GetEvents(ctx context.Context, tableName, rowID string) ([]*Event, error)

	// GetLastEvent retrieves the most recent event for a specific row.
	GetLastEvent(ctx context.Context, tableName, rowID string) (*Event, error)

	// GetEventsByTable retrieves all events for a table, ordered by row_id and version.
	GetEventsByTable(ctx context.Context, tableName string) ([]*Event, error)

	// GetEventByID retrieves a specific event by its ID.
	GetEventByID(ctx context.Context, eventID string) (*Event, error)
}

// Store manages event persistence and retrieval with hash chain integrity.
type Store struct {
	adapter Adapter
}

// NewStore creates a new event store with the given adapter.
func NewStore(adapter Adapter) *Store {
	return &Store{
		adapter: adapter,
	}
}

// AppendEvent adds a new event to the store with automatic version numbering
// and checksum calculation. It ensures hash chain integrity by linking to
// the previous event.
func (s *Store) AppendEvent(ctx context.Context, tableName, rowID string, eventType EventType, payload EventPayload, metadata map[string]any) (*Event, error) {
	// Get the last event to determine version and prev_checksum
	var version int64 = 1
	var prevChecksum string

	lastEvent, err := s.adapter.GetLastEvent(ctx, tableName, rowID)
	if err != nil {
		// If row not found, this is the first event
		if _, ok := err.(ErrRowNotFound); !ok {
			return nil, fmt.Errorf("get last event: %w", err)
		}
	} else {
		// Increment version from last event
		version = lastEvent.Version + 1
		prevChecksum = lastEvent.Checksum
	}

	// Calculate checksum
	checksum, err := hash.Calculate(payload, metadata, prevChecksum)
	if err != nil {
		return nil, fmt.Errorf("calculate checksum: %w", err)
	}

	// Create the event
	event := &Event{
		ID:           uuid.New().String(),
		TableName:    tableName,
		RowID:        rowID,
		EventType:    eventType,
		Version:      version,
		Payload:      payload,
		Metadata:     metadata,
		Checksum:     checksum,
		PrevChecksum: prevChecksum,
		CreatedAt:    time.Now().UTC(),
	}

	// Persist the event
	if err := s.adapter.InsertEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("insert event: %w", err)
	}

	return event, nil
}

// GetEvents retrieves all events for a specific row, ordered by version.
func (s *Store) GetEvents(ctx context.Context, tableName, rowID string) ([]*Event, error) {
	events, err := s.adapter.GetEvents(ctx, tableName, rowID)
	if err != nil {
		return nil, fmt.Errorf("get events: %w", err)
	}

	return events, nil
}

// GetEventsInRange retrieves events for a row within a specific version range.
// The range is inclusive on both ends [fromVersion, toVersion].
func (s *Store) GetEventsInRange(ctx context.Context, tableName, rowID string, fromVersion, toVersion int64) ([]*Event, error) {
	// Get all events
	allEvents, err := s.adapter.GetEvents(ctx, tableName, rowID)
	if err != nil {
		return nil, fmt.Errorf("get events: %w", err)
	}

	// Filter by version range
	var filtered []*Event
	for _, event := range allEvents {
		if event.Version >= fromVersion && event.Version <= toVersion {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

// GetEventsByTime retrieves all events for a row up to a specific timestamp.
func (s *Store) GetEventsByTime(ctx context.Context, tableName, rowID string, timestamp time.Time) ([]*Event, error) {
	// Get all events
	allEvents, err := s.adapter.GetEvents(ctx, tableName, rowID)
	if err != nil {
		return nil, fmt.Errorf("get events: %w", err)
	}

	// Filter by timestamp
	var filtered []*Event
	for _, event := range allEvents {
		if event.CreatedAt.Before(timestamp) || event.CreatedAt.Equal(timestamp) {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

// GetLastEvent retrieves the most recent event for a row.
func (s *Store) GetLastEvent(ctx context.Context, tableName, rowID string) (*Event, error) {
	event, err := s.adapter.GetLastEvent(ctx, tableName, rowID)
	if err != nil {
		return nil, fmt.Errorf("get last event: %w", err)
	}

	return event, nil
}

// GetEventsByTable retrieves all events for a table, ordered by row_id and version.
func (s *Store) GetEventsByTable(ctx context.Context, tableName string) ([]*Event, error) {
	events, err := s.adapter.GetEventsByTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("get events by table: %w", err)
	}

	return events, nil
}
