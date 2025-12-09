package integridb

import (
	"fmt"
	"time"
)

// IntegrityReport contains the results of an integrity verification check.
type IntegrityReport struct {
	// TableName is the table that was verified.
	TableName string `json:"table_name"`

	// RowID is the specific row that was verified (if applicable).
	// Empty string indicates a table-wide verification.
	RowID string `json:"row_id,omitempty"`

	// IsValid indicates whether the integrity check passed.
	IsValid bool `json:"is_valid"`

	// EventsChecked is the number of events verified.
	EventsChecked int `json:"events_checked"`

	// Errors contains details about integrity violations found.
	Errors []IntegrityError `json:"errors,omitempty"`

	// BrokenChains contains information about broken hash chains.
	BrokenChains []BrokenChainInfo `json:"broken_chains,omitempty"`

	// VerifiedAt is the timestamp when the verification was performed.
	VerifiedAt time.Time `json:"verified_at"`

	// Duration is how long the verification took.
	Duration time.Duration `json:"duration"`
}

// BrokenChainInfo contains details about a broken hash chain.
type BrokenChainInfo struct {
	// TableName is the table where the break occurred.
	TableName string `json:"table_name"`

	// RowID identifies the affected row.
	RowID string `json:"row_id"`

	// EventID is the event where the break was detected.
	EventID string `json:"event_id"`

	// Version is the version number of the broken event.
	Version int64 `json:"version"`

	// ExpectedChecksum is what the checksum should have been.
	ExpectedChecksum string `json:"expected_checksum"`

	// ActualChecksum is the checksum found in the event.
	ActualChecksum string `json:"actual_checksum"`

	// Reason describes why the chain is broken.
	Reason string `json:"reason"`
}

// IntegrityError represents an individual integrity violation.
type IntegrityError struct {
	// EventID is the event with the error.
	EventID string `json:"event_id"`

	// Type categorizes the error (e.g., "checksum_mismatch", "version_gap", "missing_prev").
	Type string `json:"type"`

	// Message is a human-readable description of the error.
	Message string `json:"message"`

	// Details contains additional context (optional).
	Details map[string]interface{} `json:"details,omitempty"`
}

// AddError adds an integrity error to the report.
func (r *IntegrityReport) AddError(err IntegrityError) {
	r.Errors = append(r.Errors, err)
	r.IsValid = false
}

// AddBrokenChain adds a broken chain to the report.
func (r *IntegrityReport) AddBrokenChain(info BrokenChainInfo) {
	r.BrokenChains = append(r.BrokenChains, info)
	r.IsValid = false
}

// Summary returns a human-readable summary of the report.
func (r *IntegrityReport) Summary() string {
	if r.IsValid {
		return "Integrity check passed: all events verified successfully"
	}
	return "Integrity check failed: " + r.FailureSummary()
}

// FailureSummary returns a summary of failures (only call if !IsValid).
func (r *IntegrityReport) FailureSummary() string {
	errorCount := len(r.Errors)
	brokenCount := len(r.BrokenChains)

	if errorCount > 0 && brokenCount > 0 {
		return fmt.Sprintf("%d integrity errors, %d broken chains", errorCount, brokenCount)
	}
	if errorCount > 0 {
		return fmt.Sprintf("%d integrity errors", errorCount)
	}
	if brokenCount > 0 {
		return fmt.Sprintf("%d broken chains", brokenCount)
	}
	return "unknown failure"
}
