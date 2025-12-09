# Task 4.1: Implement Transaction Wrapper

## Description
Wrap `sql.Tx` to capture events within transactions with proper buffering and atomic commit.

## Actions

1. Update `Tx` struct to buffer events:
   - Add `pendingEvents` slice to store event data before commit
   - Create `pendingEvent` struct to hold event information temporarily
   - Track operation order to maintain hash chain integrity

2. Implement transaction methods:
   - `Exec()` and `ExecContext()`: Capture event data but don't persist yet
   - `Commit()`: Persist all pending events to event store, then commit transaction
   - `Rollback()`: Discard pending events, rollback transaction

3. Handle event ordering within transaction:
   - Track operation order accurately
   - Persist events in same order as operations occurred
   - Ensure hash chain integrity across transaction

4. Handle nested operations:
   - Support INSERT then UPDATE on same row in same transaction
   - Should create two separate events with correct version numbering

## Acceptance Criteria

- [ ] Events persist only on Commit
- [ ] Events discarded on Rollback
- [ ] Multiple operations in transaction work correctly
- [ ] Hash chain correct across transaction
- [ ] Integration tests pass
