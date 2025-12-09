package event

import "fmt"

// ErrRowNotFound indicates a requested row does not exist in the event store.
type ErrRowNotFound struct {
	TableName string
	RowID     string
}

func (e ErrRowNotFound) Error() string {
	return fmt.Sprintf("row %q not found in table %q", e.RowID, e.TableName)
}

// ErrEventNotFound indicates a requested event does not exist.
type ErrEventNotFound struct {
	EventID string
}

func (e ErrEventNotFound) Error() string {
	return fmt.Sprintf("event %q not found", e.EventID)
}

// ErrVersionMismatch indicates an unexpected version number in the event chain.
type ErrVersionMismatch struct {
	Expected int64
	Actual   int64
	EventID  string
}

func (e ErrVersionMismatch) Error() string {
	return fmt.Sprintf("version mismatch in event %q: expected %d, got %d",
		e.EventID, e.Expected, e.Actual)
}

// ErrChecksumMismatch indicates a hash chain checksum mismatch.
type ErrChecksumMismatch struct {
	EventID  string
	Expected string
	Actual   string
}

func (e ErrChecksumMismatch) Error() string {
	return fmt.Sprintf("checksum mismatch in event %q: expected %q, got %q",
		e.EventID, e.Expected, e.Actual)
}
