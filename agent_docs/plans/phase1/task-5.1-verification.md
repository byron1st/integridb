# Task 5.1: Implement Integrity Verification

## Description
Verify hash chain integrity for events to detect tampering or data corruption.

## Actions

1. Create `pkg/integrity/verifier.go` with:
   - `Verifier` struct wrapping event store
   - `NewVerifier()` constructor
   - `VerifyRow()` method to verify all events for a single row
   - `VerifyTable()` method to verify all rows in a table

2. Implement verification logic:
   - Retrieve events in order for a row/table
   - Verify prev_checksum linkage between consecutive events
   - Recalculate checksum for each event and compare with stored value
   - Identify broken chain location and specific integrity violations
   - Generate detailed integrity report

3. Expose via `DB` public API:
   - `db.VerifyIntegrity()` for single row verification
   - `db.VerifyTable()` for table-wide verification

4. Write comprehensive tests:
   - Test valid chains report as valid
   - Test tampered events are detected
   - Test broken chain location identified
   - Test performance with large chains

## Acceptance Criteria

- [ ] Valid chains report as valid
- [ ] Tampered events detected
- [ ] Broken chain location identified
- [ ] Performance acceptable for large chains
- [ ] Integration tests pass
