# Task 1.4: Implement database/sql Wrapper Skeleton

## Description
Create the wrapper types that mirror `database/sql` interfaces to provide a familiar API while enabling event capture.

## Actions

1. Create `pkg/integridb/db.go` with:
   - `DB` struct wrapping `*sql.DB` with event capture infrastructure
   - `Open()` function to create new IntegriDB connection
   - Standard `sql.DB` methods (Exec, Query, Begin, Prepare, Close, Ping, etc.)
   - Initially implement as pass-through to underlying `*sql.DB`

2. Create `pkg/integridb/tx.go` with:
   - `Tx` struct wrapping `*sql.Tx` with event buffering
   - Transaction methods (Exec, Query, Commit, Rollback, etc.)
   - `pendingEvents` slice to buffer events before commit

## Acceptance Criteria

- [ ] `integridb.Open()` successfully connects to PostgreSQL
- [ ] Basic pass-through operations work (SELECT, INSERT without event capture)
- [ ] `db.Close()` properly closes connection
- [ ] Unit tests pass
