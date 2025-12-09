# Task 3.2: Implement INSERT Event Capture

## Description
Capture INSERT operations as events with after state.

## Actions

1. Update `db.ExecContext()` to detect INSERT operations:
   - Parse SQL query using the query parser
   - Check if operation is INSERT and table is tracked
   - Delegate to `execInsert()` if conditions met

2. Implement `execInsert()` method:
   - Execute original INSERT statement
   - Extract row ID from result or RETURNING clause
   - Query after state for the inserted row
   - Create event with `before: nil, after: state`
   - Append event to store

3. Handle RETURNING clause:
   - Parse RETURNING to get row ID
   - Support `INSERT ... RETURNING id`
   - Handle cases without RETURNING by using `LastInsertId()`

## Acceptance Criteria

- [ ] INSERT creates event with correct after state
- [ ] Row ID extracted correctly
- [ ] Event has correct version (1 for new row)
- [ ] Checksum calculated correctly
- [ ] Integration tests pass
