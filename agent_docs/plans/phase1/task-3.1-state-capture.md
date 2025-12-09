# Task 3.1: Implement Before/After State Capture

## Description
Query row state before and after mutations to create complete event payloads.

## Actions

1. Create `pkg/integridb/capture.go` with:
   - `captureBeforeState()` method to query current row state before mutation
   - `captureAfterState()` method to query row state after mutation
   - `extractRowID()` function to extract primary key value from args or RETURNING clause

2. Implement column discovery:
   - Query `information_schema.columns` to get column list for tables
   - Cache column metadata per table to avoid repeated queries
   - Build SELECT query dynamically based on discovered columns

3. Handle multiple rows affected:
   - For UPDATE/DELETE: capture all affected row IDs first using WHERE clause
   - Create separate events for each row
   - Use transactions to ensure atomicity

## Acceptance Criteria

- [ ] Before state captured accurately for UPDATE/DELETE
- [ ] After state captured accurately for INSERT/UPDATE
- [ ] Works with various column types (string, int, timestamp, etc.)
- [ ] Handles NULL values correctly
- [ ] Integration tests pass
