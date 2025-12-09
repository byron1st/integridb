# Task 2.2: Implement Hash Chain Logic

## Description
Implement deterministic checksum calculation for events to create a tamper-evident hash chain.

## Actions

1. Create `internal/serialize/json.go` with:
   - `SerializeDeterministic()` function that converts values to JSON with sorted keys
   - Ensures consistent JSON output for identical inputs

2. Create `internal/hash/checksum.go` with:
   - `Calculate()` function to compute SHA-256 checksum for an event
   - Checksum formula: `SHA256(JSON(payload) + JSON(metadata) + prev_checksum)`
   - `Verify()` function to check if a checksum is valid

3. Write tests in `internal/hash/checksum_test.go`:
   - Test deterministic output (same input = same hash)
   - Test chain linking (prev_checksum included in calculation)
   - Test empty metadata handling
   - Test nil values in payload

## Acceptance Criteria

- [ ] Same payload always produces same checksum
- [ ] Different payloads produce different checksums
- [ ] Checksum includes prev_checksum in calculation
- [ ] JSON serialization is deterministic (sorted keys)
- [ ] Unit tests pass
