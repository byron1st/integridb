package serialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// SerializeDeterministic converts a value to JSON with sorted keys.
// This ensures consistent JSON output for identical inputs, which is
// critical for hash chain integrity verification.
//
// The function recursively sorts all maps by their keys before serialization,
// ensuring that the same data structure always produces the same JSON string.
func SerializeDeterministic(v any) (string, error) {
	// Normalize the value to ensure deterministic output
	normalized := normalize(v)

	// Marshal with sorted keys
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false) // Don't escape HTML characters for consistency
	encoder.SetIndent("", "")    // No indentation for compact output

	if err := encoder.Encode(normalized); err != nil {
		return "", fmt.Errorf("failed to serialize: %w", err)
	}

	// Remove trailing newline added by Encode
	result := buf.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}

// normalize recursively processes a value to ensure deterministic serialization.
// It converts maps to ordered structures and recursively processes nested values.
func normalize(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]any:
		// Sort map keys and create ordered representation
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Create ordered map
		ordered := make(map[string]any, len(val))
		for _, k := range keys {
			ordered[k] = normalize(val[k])
		}
		return ordered

	case []any:
		// Recursively normalize slice elements
		normalized := make([]any, len(val))
		for i, item := range val {
			normalized[i] = normalize(item)
		}
		return normalized

	default:
		// For primitive types, return as-is
		return v
	}
}
