# Task 5.4: Implement Event History Retrieval

## Description
Expose event history to users for audit and analysis purposes.

## Actions

1. Add to `pkg/integridb/db.go`:
   - `GetHistory()` method to retrieve all events for a row
   - `GetHistoryRange()` method to retrieve events in version range

2. Implementation:
   - Delegate to event store's query methods
   - Return events in chronological order (ordered by version)
   - Include all event details (payload, metadata, checksums, timestamps)

3. Write tests:
   - Test full history retrieval
   - Test range queries with various parameters
   - Test ordering of returned events
   - Test large history retrieval performance

## Acceptance Criteria

- [ ] Full history returned correctly
- [ ] Range queries work correctly
- [ ] Events ordered by version
- [ ] Integration tests pass
