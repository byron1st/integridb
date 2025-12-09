# Task 2.3: Implement Event Store

## Description
Build the event store component for persisting and querying events with automatic version numbering and hash chain linking.

## Actions

1. Create `pkg/event/store.go` with:
   - `Store` struct with adapter interface
   - `NewStore()` constructor
   - `AppendEvent()` method to add events with checksum calculation
   - `GetEvents()` method to retrieve all events for a row
   - `GetEventsInRange()` method to retrieve events within version range
   - `GetEventsByTime()` method to retrieve events up to a timestamp
   - `GetLastEvent()` method to retrieve the most recent event for a row

2. Implement version numbering:
   - Get last event version for row
   - Increment version for new event
   - Handle first event (version = 1, prev_checksum = "")

3. Write tests in `pkg/event/store_test.go`:
   - Test event persistence
   - Test version auto-increment
   - Test hash chain linking
   - Test query methods

## Acceptance Criteria

- [ ] Events are persisted correctly
- [ ] Version numbers auto-increment per row
- [ ] Hash chain links events correctly
- [ ] Queries return events in correct order
- [ ] Integration tests with PostgreSQL pass
