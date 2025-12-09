# Task 6.4: Example Applications

## Description
Create working example applications demonstrating IntegriDB features.

## Actions

1. Create `examples/basic/main.go`:
   - Open IntegriDB connection
   - Create tracked table
   - Perform simple CRUD operations
   - Retrieve event history
   - Verify integrity
   - Display results

2. Create `examples/transactions/main.go`:
   - Demonstrate transaction usage
   - Show commit and rollback behavior
   - Multiple operations within transaction
   - Event buffering demonstration

3. Create `examples/metadata/main.go`:
   - Show metadata injection via context
   - Configure MetadataFunc
   - Demonstrate user ID tracking
   - Show metadata in event history

4. Create `examples/replay/main.go`:
   - Insert/update/delete sequence
   - Replay state at different versions
   - Replay state at different times
   - Show state evolution over time

5. Add README.md for each example:
   - Explain what the example demonstrates
   - How to run the example
   - Expected output

## Acceptance Criteria

- [ ] All examples compile
- [ ] All examples run successfully with PostgreSQL
- [ ] Examples are well-commented
- [ ] Each example has a README
