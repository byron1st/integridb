# Task 4.3: Implement Table Management

## Description
Allow dynamic table tracking management so users can enable/disable event capture for specific tables.

## Actions

1. Implement table management methods in `pkg/integridb/db.go`:
   - `TrackTable()` method to start tracking a table
   - `UntrackTable()` method to stop tracking a table
   - `ListTrackedTables()` method to return all tracked tables
   - `isTracked()` internal method to check if a table is being tracked

2. Persist tracked tables in database:
   - Use `integridb_tracked_tables` table for persistence
   - Insert/delete records when tracking status changes
   - Load tracked tables from database on startup

3. Handle edge cases:
   - Track table that's already tracked (idempotent)
   - Untrack table that's not tracked (no-op)
   - Query tracked tables efficiently

## Acceptance Criteria

- [ ] Tables can be tracked dynamically
- [ ] Tables can be untracked
- [ ] Tracked tables persisted across restarts
- [ ] Integration tests pass
