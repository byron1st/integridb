# Task 3.4: Implement DELETE Event Capture

## Description
Capture DELETE operations as events with before state.

## Actions

1. Implement `execDelete()` method:
   - Parse WHERE clause to identify affected rows
   - Query before state for all affected rows
   - Execute original DELETE statement
   - Create events with `before: state, after: nil`

2. Handle edge cases:
   - DELETE with no rows affected
   - DELETE that removes multiple rows
   - DELETE with complex WHERE clauses
   - DELETE without WHERE clause (delete all rows)

3. Ensure proper event creation:
   - Events created before deletion occurs
   - Row IDs captured correctly
   - Version numbers increment properly

## Acceptance Criteria

- [ ] DELETE creates event with correct before state
- [ ] Works with multiple rows affected
- [ ] Version increments correctly
- [ ] Hash chain maintained
- [ ] Integration tests pass
