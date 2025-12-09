package integridb

import (
	"fmt"

	"github.com/byron1st/integridb/pkg/event"
)

// Re-export event-related errors for convenience
type (
	ErrRowNotFound      = event.ErrRowNotFound
	ErrEventNotFound    = event.ErrEventNotFound
	ErrVersionMismatch  = event.ErrVersionMismatch
	ErrChecksumMismatch = event.ErrChecksumMismatch
)

// ErrTableNotTracked indicates an operation was attempted on a table that is not configured for tracking.
type ErrTableNotTracked struct {
	TableName string
}

func (e ErrTableNotTracked) Error() string {
	return fmt.Sprintf("table %q is not tracked", e.TableName)
}

// ErrIntegrityViolation indicates a hash chain integrity check failed.
type ErrIntegrityViolation struct {
	TableName string
	RowID     string
	EventID   string
	Reason    string
}

func (e ErrIntegrityViolation) Error() string {
	return fmt.Sprintf("integrity violation in table %q, row %q, event %q: %s",
		e.TableName, e.RowID, e.EventID, e.Reason)
}

// ErrInvalidSQL indicates the SQL statement could not be parsed or is not supported.
type ErrInvalidSQL struct {
	SQL    string
	Reason string
}

func (e ErrInvalidSQL) Error() string {
	return fmt.Sprintf("invalid SQL: %s (statement: %q)", e.Reason, e.SQL)
}

// ErrInvalidConfig indicates the configuration is invalid.
type ErrInvalidConfig struct {
	Field  string
	Reason string
	Index  *int // Optional: for array/slice fields
}

func (e ErrInvalidConfig) Error() string {
	if e.Index != nil {
		return fmt.Sprintf("invalid config field %q[%d]: %s", e.Field, *e.Index, e.Reason)
	}
	return fmt.Sprintf("invalid config field %q: %s", e.Field, e.Reason)
}

