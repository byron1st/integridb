# Task 3.3: Implement UPDATE Event Capture

## Description
Capture UPDATE operations as events with before and after states.

## Actions

1. Implement `execUpdate()` method:
   - Parse WHERE clause to identify affected rows
   - Query before state for all affected rows
   - Execute original UPDATE statement
   - Query after state for all affected rows
   - Detect changed columns by comparing before/after
   - Create events for each affected row

2. Implement `detectChangedColumns()` function:
   - Compare before and after states field by field
   - Use reflection or map comparison to identify changes
   - Return list of changed column names
   - Include in event payload for audit trail

3. Handle edge cases:
   - UPDATE with no rows affected
   - UPDATE that changes multiple rows
   - UPDATE with complex WHERE clauses

## Acceptance Criteria

- [ ] UPDATE creates event with before and after states
- [ ] Changed columns accurately detected
- [ ] Works with multiple rows affected
- [ ] Version increments correctly
- [ ] Hash chain maintained
- [ ] Integration tests pass
