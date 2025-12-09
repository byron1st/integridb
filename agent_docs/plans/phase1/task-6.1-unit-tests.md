# Task 6.1: Comprehensive Unit Tests

## Description
Achieve >80% code coverage with comprehensive unit tests across all packages.

## Actions

1. Review and add tests for all packages:
   - `pkg/integridb/` - API tests, DB wrapper tests
   - `pkg/event/` - Event store tests
   - `pkg/query/` - SQL parser tests
   - `pkg/integrity/` - Verification tests
   - `pkg/projection/` - Replay tests
   - `pkg/adapter/postgres/` - Adapter tests
   - `internal/hash/` - Checksum tests
   - `internal/serialize/` - JSON serialization tests

2. Add edge case tests:
   - NULL values in columns
   - Unicode characters in data
   - Large payloads
   - Concurrent operations
   - Error conditions and error handling
   - Boundary conditions

3. Run coverage analysis:
   - Execute `go test -cover ./...`
   - Identify untested code paths
   - Add tests to reach >80% coverage

4. Run race detection:
   - Execute `go test -race ./...`
   - Fix any race conditions

## Acceptance Criteria

- [ ] >80% code coverage
- [ ] All edge cases covered
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes
