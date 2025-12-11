package integridb

import "context"

// metadataKey is the context key type for metadata injection.
type metadataKey struct{}

// WithMetadata adds metadata to the context for event capture.
// The metadata will be included in all events generated during operations
// executed with this context.
//
// Example:
//
//	ctx := integridb.WithMetadata(context.Background(), map[string]any{
//	    "user_id": "12345",
//	    "session_id": "abc-def-ghi",
//	    "ip_address": "192.168.1.1",
//	})
//	db.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
func WithMetadata(ctx context.Context, metadata map[string]any) context.Context {
	return context.WithValue(ctx, metadataKey{}, metadata)
}

// GetMetadata retrieves metadata from the context.
// Returns nil if no metadata is present.
func GetMetadata(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}

	metadata, ok := ctx.Value(metadataKey{}).(map[string]any)
	if !ok {
		return nil
	}

	return metadata
}

// extractMetadata extracts metadata from both context and MetadataFunc if configured.
// Context metadata takes precedence and is merged with MetadataFunc output.
// If both sources provide the same key, the context value wins.
func (d *DB) extractMetadata(ctx context.Context) map[string]any {
	var result map[string]any

	// Extract from MetadataFunc if configured
	if d.config.MetadataFunc != nil {
		funcMetadata := d.config.MetadataFunc(ctx)
		if funcMetadata != nil {
			result = make(map[string]any, len(funcMetadata))
			for k, v := range funcMetadata {
				result[k] = v
			}
		}
	}

	// Extract from context and merge (context takes precedence)
	ctxMetadata := GetMetadata(ctx)
	if ctxMetadata != nil {
		if result == nil {
			result = make(map[string]any, len(ctxMetadata))
		}
		for k, v := range ctxMetadata {
			result[k] = v
		}
	}

	return result
}
