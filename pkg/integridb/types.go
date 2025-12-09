package integridb

import "github.com/byron1st/integridb/pkg/event"

// Re-export event types for convenience
type (
	Event        = event.Event
	EventType    = event.EventType
	EventPayload = event.EventPayload
)

// Re-export event type constants
const (
	EventTypeInsert = event.EventTypeInsert
	EventTypeUpdate = event.EventTypeUpdate
	EventTypeDelete = event.EventTypeDelete
)
