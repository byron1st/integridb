# Task 5.3: Implement State Replay by Time

## Description
Reconstruct row state at a specific timestamp by replaying events up to that time.

## Actions

1. Add to `pkg/projection/replay.go`:
   - `ReplayToTime()` method to reconstruct state at specific timestamp

2. Implement time-based replay logic:
   - Query events up to specified timestamp using `GetEventsByTime()`
   - Apply same replay logic as version-based replay
   - Handle timezone considerations correctly

3. Expose via `DB` public API:
   - `db.ReplayToTime()` method for time-based replay

4. Write tests:
   - Test correct state returned for any timestamp
   - Test timezone handling (UTC, local time, etc.)
   - Test edge cases (time before first event, time after last event)

## Acceptance Criteria

- [ ] Correct state returned for any timestamp
- [ ] Works with timezone handling
- [ ] Integration tests pass
