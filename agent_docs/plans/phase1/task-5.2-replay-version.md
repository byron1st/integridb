# Task 5.2: Implement State Replay by Version

## Description
Reconstruct row state at a specific version number by replaying events.

## Actions

1. Create `pkg/projection/replay.go` with:
   - `Replayer` struct wrapping event store
   - `NewReplayer()` constructor
   - `ReplayToVersion()` method to reconstruct state at specific version

2. Implement replay logic:
   - Retrieve events from version 1 to target version
   - Apply events sequentially:
     - INSERT: Set state to after state
     - UPDATE: Update state to after state
     - DELETE: Set state to nil
   - Return final reconstructed state

3. Expose via `DB` public API:
   - `db.ReplayTo()` method for version-based replay

4. Write tests:
   - Test correct state returned for any version
   - Test returns nil for deleted rows
   - Test error for non-existent rows
   - Test replay performance

## Acceptance Criteria

- [ ] Correct state returned for any version
- [ ] Returns nil for deleted rows
- [ ] Error for non-existent rows
- [ ] Integration tests pass
