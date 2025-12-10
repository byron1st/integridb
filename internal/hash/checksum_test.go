package hash

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name         string
		payload      any
		metadata     any
		prevChecksum string
		wantErr      bool
	}{
		{
			name: "simple payload",
			payload: map[string]any{
				"before": nil,
				"after": map[string]any{
					"id":   "123",
					"name": "John",
				},
			},
			metadata:     map[string]any{"user_id": "admin"},
			prevChecksum: "",
			wantErr:      false,
		},
		{
			name: "nil metadata",
			payload: map[string]any{
				"before": nil,
				"after": map[string]any{
					"id": "123",
				},
			},
			metadata:     nil,
			prevChecksum: "",
			wantErr:      false,
		},
		{
			name: "with previous checksum",
			payload: map[string]any{
				"before": map[string]any{"id": "123", "name": "Old"},
				"after":  map[string]any{"id": "123", "name": "New"},
			},
			metadata:     map[string]any{"user_id": "admin"},
			prevChecksum: "abc123",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.payload, tt.metadata, tt.prevChecksum)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == "" {
					t.Errorf("Calculate() returned empty checksum")
				}
				if len(got) != 64 { // SHA-256 produces 64 hex characters
					t.Errorf("Calculate() returned checksum of length %d, want 64", len(got))
				}
			}
		})
	}
}

func TestCalculate_Deterministic(t *testing.T) {
	// Test that the same input always produces the same checksum
	payload := map[string]any{
		"before": nil,
		"after": map[string]any{
			"id":   "123",
			"name": "John",
		},
	}
	metadata := map[string]any{"user_id": "admin"}
	prevChecksum := "abc123"

	checksum1, err1 := Calculate(payload, metadata, prevChecksum)
	checksum2, err2 := Calculate(payload, metadata, prevChecksum)
	checksum3, err3 := Calculate(payload, metadata, prevChecksum)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("Calculate() errors: %v, %v, %v", err1, err2, err3)
	}

	if checksum1 != checksum2 || checksum2 != checksum3 {
		t.Errorf("Calculate() not deterministic:\n%s\n%s\n%s", checksum1, checksum2, checksum3)
	}
}

func TestCalculate_DifferentPayloadsDifferentChecksums(t *testing.T) {
	metadata := map[string]any{"user_id": "admin"}
	prevChecksum := ""

	payload1 := map[string]any{
		"after": map[string]any{"id": "123", "name": "John"},
	}
	payload2 := map[string]any{
		"after": map[string]any{"id": "123", "name": "Jane"},
	}

	checksum1, err1 := Calculate(payload1, metadata, prevChecksum)
	checksum2, err2 := Calculate(payload2, metadata, prevChecksum)

	if err1 != nil || err2 != nil {
		t.Fatalf("Calculate() errors: %v, %v", err1, err2)
	}

	if checksum1 == checksum2 {
		t.Errorf("Calculate() produced same checksum for different payloads")
	}
}

func TestCalculate_DifferentMetadataDifferentChecksums(t *testing.T) {
	payload := map[string]any{
		"after": map[string]any{"id": "123", "name": "John"},
	}
	prevChecksum := ""

	metadata1 := map[string]any{"user_id": "admin"}
	metadata2 := map[string]any{"user_id": "user"}

	checksum1, err1 := Calculate(payload, metadata1, prevChecksum)
	checksum2, err2 := Calculate(payload, metadata2, prevChecksum)

	if err1 != nil || err2 != nil {
		t.Fatalf("Calculate() errors: %v, %v", err1, err2)
	}

	if checksum1 == checksum2 {
		t.Errorf("Calculate() produced same checksum for different metadata")
	}
}

func TestCalculate_DifferentPrevChecksumDifferentChecksums(t *testing.T) {
	payload := map[string]any{
		"after": map[string]any{"id": "123", "name": "John"},
	}
	metadata := map[string]any{"user_id": "admin"}

	checksum1, err1 := Calculate(payload, metadata, "")
	checksum2, err2 := Calculate(payload, metadata, "abc123")

	if err1 != nil || err2 != nil {
		t.Fatalf("Calculate() errors: %v, %v", err1, err2)
	}

	if checksum1 == checksum2 {
		t.Errorf("Calculate() produced same checksum for different prevChecksum")
	}
}

func TestCalculate_EmptyMetadataSameAsNil(t *testing.T) {
	payload := map[string]any{
		"after": map[string]any{"id": "123"},
	}
	prevChecksum := ""

	checksum1, err1 := Calculate(payload, nil, prevChecksum)
	checksum2, err2 := Calculate(payload, map[string]any{}, prevChecksum)

	if err1 != nil || err2 != nil {
		t.Fatalf("Calculate() errors: %v, %v", err1, err2)
	}

	if checksum1 != checksum2 {
		t.Errorf("Calculate() produced different checksums for nil and empty metadata")
	}
}

func TestVerify(t *testing.T) {
	payload := map[string]any{
		"before": nil,
		"after": map[string]any{
			"id":   "123",
			"name": "John",
		},
	}
	metadata := map[string]any{"user_id": "admin"}
	prevChecksum := "abc123"

	// Calculate correct checksum
	correctChecksum, err := Calculate(payload, metadata, prevChecksum)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	tests := []struct {
		name         string
		payload      any
		metadata     any
		prevChecksum string
		checksum     string
		wantValid    bool
		wantErr      bool
	}{
		{
			name:         "valid checksum",
			payload:      payload,
			metadata:     metadata,
			prevChecksum: prevChecksum,
			checksum:     correctChecksum,
			wantValid:    true,
			wantErr:      false,
		},
		{
			name:         "invalid checksum",
			payload:      payload,
			metadata:     metadata,
			prevChecksum: prevChecksum,
			checksum:     "invalid_checksum",
			wantValid:    false,
			wantErr:      false,
		},
		{
			name:         "wrong prevChecksum",
			payload:      payload,
			metadata:     metadata,
			prevChecksum: "wrong",
			checksum:     correctChecksum,
			wantValid:    false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := Verify(tt.payload, tt.metadata, tt.prevChecksum, tt.checksum)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("Verify() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestVerify_HashChainLinking(t *testing.T) {
	// Test that changing prev_checksum invalidates verification
	payload := map[string]any{
		"after": map[string]any{"id": "123"},
	}
	metadata := map[string]any{}

	// First event (no previous)
	checksum1, err := Calculate(payload, metadata, "")
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	// Second event (linked to first)
	checksum2, err := Calculate(payload, metadata, checksum1)
	if err != nil {
		t.Fatalf("Calculate() error = %v", err)
	}

	// Verify second event with correct prev_checksum
	valid, err := Verify(payload, metadata, checksum1, checksum2)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !valid {
		t.Errorf("Verify() = false, want true for correct prev_checksum")
	}

	// Verify second event with wrong prev_checksum should fail
	valid, err = Verify(payload, metadata, "wrong_prev", checksum2)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if valid {
		t.Errorf("Verify() = true, want false for wrong prev_checksum")
	}
}
