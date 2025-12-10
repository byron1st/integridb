package serialize

import (
	"testing"
)

func TestSerializeDeterministic(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:    "nil value",
			input:   nil,
			want:    "null",
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   map[string]any{},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "simple map",
			input: map[string]any{
				"name": "John",
				"age":  30,
			},
			want:    `{"age":30,"name":"John"}`,
			wantErr: false,
		},
		{
			name: "map with keys in different order",
			input: map[string]any{
				"z": "last",
				"a": "first",
				"m": "middle",
			},
			want:    `{"a":"first","m":"middle","z":"last"}`,
			wantErr: false,
		},
		{
			name: "nested map",
			input: map[string]any{
				"user": map[string]any{
					"name": "John",
					"age":  30,
				},
				"id": 123,
			},
			want:    `{"id":123,"user":{"age":30,"name":"John"}}`,
			wantErr: false,
		},
		{
			name: "map with array",
			input: map[string]any{
				"name":  "John",
				"tags":  []any{"admin", "user"},
				"count": 2,
			},
			want:    `{"count":2,"name":"John","tags":["admin","user"]}`,
			wantErr: false,
		},
		{
			name: "complex nested structure",
			input: map[string]any{
				"before": map[string]any{
					"id":   "123",
					"name": "Old Name",
				},
				"after": map[string]any{
					"id":   "123",
					"name": "New Name",
				},
				"changed_columns": []any{"name"},
			},
			want:    `{"after":{"id":"123","name":"New Name"},"before":{"id":"123","name":"Old Name"},"changed_columns":["name"]}`,
			wantErr: false,
		},
		{
			name: "map with null values",
			input: map[string]any{
				"before": nil,
				"after": map[string]any{
					"id": "123",
				},
			},
			want:    `{"after":{"id":"123"},"before":null}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SerializeDeterministic(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeDeterministic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SerializeDeterministic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSerializeDeterministic_Consistency(t *testing.T) {
	// Test that the same input always produces the same output
	input := map[string]any{
		"z": "last",
		"a": "first",
		"m": "middle",
		"nested": map[string]any{
			"y": 2,
			"x": 1,
		},
	}

	result1, err1 := SerializeDeterministic(input)
	result2, err2 := SerializeDeterministic(input)
	result3, err3 := SerializeDeterministic(input)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("SerializeDeterministic() errors: %v, %v, %v", err1, err2, err3)
	}

	if result1 != result2 || result2 != result3 {
		t.Errorf("SerializeDeterministic() not consistent:\n%s\n%s\n%s", result1, result2, result3)
	}
}

func TestSerializeDeterministic_DifferentInputsDifferentOutputs(t *testing.T) {
	input1 := map[string]any{"a": 1, "b": 2}
	input2 := map[string]any{"a": 1, "b": 3}

	result1, err1 := SerializeDeterministic(input1)
	result2, err2 := SerializeDeterministic(input2)

	if err1 != nil || err2 != nil {
		t.Fatalf("SerializeDeterministic() errors: %v, %v", err1, err2)
	}

	if result1 == result2 {
		t.Errorf("SerializeDeterministic() produced same output for different inputs")
	}
}
