# Task 6.2: Integration Tests

## Description
Test with real PostgreSQL instances to validate end-to-end functionality.

## Actions

1. Set up integration test infrastructure:
   - Use testcontainers-go to spin up PostgreSQL containers
   - Create helper functions for test database setup/teardown
   - Ensure tests are isolated and repeatable

2. Create `integration_test.go` files with test scenarios:
   - Full CRUD workflow (INSERT → UPDATE → DELETE with event capture)
   - Transaction commit and rollback behavior
   - Concurrent operations and race conditions
   - Large number of events (performance testing)
   - Integrity verification after tampering
   - State replay accuracy

3. Test error conditions:
   - Connection failures
   - Schema migration errors
   - Constraint violations
   - Transaction deadlocks

4. Ensure test cleanup:
   - All containers stopped after tests
   - No resource leaks
   - Tests can run in parallel

## Acceptance Criteria

- [ ] All integration tests pass
- [ ] Tests are reproducible
- [ ] Tests clean up after themselves
- [ ] Can run in CI/CD environment
