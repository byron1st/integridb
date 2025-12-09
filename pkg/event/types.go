package event

import (
	"encoding/json"
	"time"
)

// EventType represents the type of database mutation that occurred.
type EventType string

const (
	// EventTypeInsert represents an INSERT operation.
	EventTypeInsert EventType = "INSERT"
	// EventTypeUpdate represents an UPDATE operation.
	EventTypeUpdate EventType = "UPDATE"
	// EventTypeDelete represents a DELETE operation.
	EventTypeDelete EventType = "DELETE"
)

// Event represents an immutable record of a state change in the database.
// Each event is cryptographically linked to the previous event via hash chaining.
type Event struct {
	// ID is the unique identifier for this event.
	ID string `json:"id"`

	// TableName is the name of the table this event belongs to.
	TableName string `json:"table_name"`

	// RowID is the primary key value(s) identifying the affected row.
	// For composite keys, this is a JSON-serialized map.
	RowID string `json:"row_id"`

	// EventType indicates the type of mutation (INSERT, UPDATE, DELETE).
	EventType EventType `json:"event_type"`

	// Version is the sequential version number for this row.
	// Starts at 1 for INSERT and increments with each UPDATE/DELETE.
	Version int64 `json:"version"`

	// Payload contains the before/after state and changed columns.
	Payload EventPayload `json:"payload"`

	// Metadata contains user-defined metadata (e.g., user_id, request_id).
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Checksum is the SHA-256 hash of this event's content.
	Checksum string `json:"checksum"`

	// PrevChecksum is the SHA-256 hash of the previous event for this row.
	// Empty for the first event (INSERT).
	PrevChecksum string `json:"prev_checksum"`

	// CreatedAt is the timestamp when this event was persisted.
	CreatedAt time.Time `json:"created_at"`
}

// EventPayload contains the state data for an event.
type EventPayload struct {
	// Before contains the row state before the mutation.
	// Nil for INSERT operations.
	Before map[string]interface{} `json:"before,omitempty"`

	// After contains the row state after the mutation.
	// Nil for DELETE operations.
	After map[string]interface{} `json:"after,omitempty"`

	// ChangedColumns lists the columns modified in an UPDATE operation.
	// Empty for INSERT and DELETE operations.
	ChangedColumns []string `json:"changed_columns,omitempty"`
}

// MarshalJSON implements custom JSON marshaling to ensure deterministic output.
func (p EventPayload) MarshalJSON() ([]byte, error) {
	type Alias EventPayload
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&p),
	})
}
