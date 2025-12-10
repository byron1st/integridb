package hash

import (
	"crypto/sha256"
	"fmt"

	"github.com/byron1st/integridb/internal/serialize"
)

// Calculate computes a SHA-256 checksum for an event.
// The checksum is calculated as: SHA256(JSON(payload) + JSON(metadata) + prev_checksum)
//
// This creates a hash chain where each event's checksum depends on the previous
// event's checksum, making tampering detectable.
//
// Parameters:
//   - payload: The event payload containing before/after states
//   - metadata: Additional event metadata
//   - prevChecksum: The checksum of the previous event (empty string for first event)
//
// Returns the hex-encoded SHA-256 hash.
func Calculate(payload any, metadata any, prevChecksum string) (string, error) {
	// Serialize payload deterministically
	payloadJSON, err := serialize.SerializeDeterministic(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize payload: %w", err)
	}

	// Serialize metadata deterministically (use empty object if nil)
	metadataJSON := "{}"
	if metadata != nil {
		metadataJSON, err = serialize.SerializeDeterministic(metadata)
		if err != nil {
			return "", fmt.Errorf("failed to serialize metadata: %w", err)
		}
	}

	// Concatenate: payload + metadata + prev_checksum
	combined := payloadJSON + metadataJSON + prevChecksum

	// Calculate SHA-256 hash
	hash := sha256.Sum256([]byte(combined))

	// Return hex-encoded hash
	return fmt.Sprintf("%x", hash), nil
}

// Verify checks if a given checksum is valid for the provided event data.
// It recalculates the checksum and compares it with the provided value.
//
// Returns true if the checksum is valid, false otherwise.
func Verify(payload any, metadata any, prevChecksum string, checksum string) (bool, error) {
	// Calculate expected checksum
	expected, err := Calculate(payload, metadata, prevChecksum)
	if err != nil {
		return false, err
	}

	// Compare with provided checksum
	return expected == checksum, nil
}
