# Task 4.2: Implement Metadata Injection

## Description
Allow users to inject metadata into events via context for audit trail purposes (user ID, session ID, etc.).

## Actions

1. Create `pkg/integridb/context.go` with:
   - `MetadataKey` context key type for type safety
   - `WithMetadata()` function to add metadata to context
   - `GetMetadata()` function to retrieve metadata from context

2. Update event capture to include metadata:
   - Check context for metadata in all capture operations
   - Merge context metadata with MetadataFunc output if configured
   - Include metadata in checksum calculation for integrity

3. Write tests for metadata injection:
   - Test context metadata propagation
   - Test MetadataFunc integration
   - Test metadata inclusion in checksums
   - Test metadata persistence and retrieval

## Acceptance Criteria

- [ ] Metadata from context included in events
- [ ] Metadata from MetadataFunc included in events
- [ ] Metadata included in checksum calculation
- [ ] Unit tests pass
