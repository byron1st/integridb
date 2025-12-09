# Task 1.2: Define Core Types

## Description
Create the foundational type definitions in `pkg/integridb/` that will be used throughout the library.

## Actions

1. Create `pkg/integridb/types.go` with:
   - `Event` struct (ID, TableName, RowID, EventType, Version, Payload, Metadata, Checksum, PrevChecksum, CreatedAt)
   - `EventType` enum (INSERT, UPDATE, DELETE)
   - `EventPayload` struct (Before, After, ChangedColumns)

2. Create `pkg/integridb/config.go` with:
   - `Config` struct (Driver, DSN, TrackedTables, connection pool settings, MetadataFunc)
   - `TableConfig` struct (Name, PrimaryKey)

3. Create `pkg/integridb/errors.go` with:
   - `ErrTableNotTracked`
   - `ErrRowNotFound`
   - `ErrIntegrityViolation`
   - `ErrInvalidSQL`

4. Create `pkg/integridb/report.go` with:
   - `IntegrityReport` struct (verification results)
   - `BrokenChainInfo` struct (details about broken hash chains)
   - `IntegrityError` struct (individual integrity errors)

## Acceptance Criteria

- [ ] All types compile without errors
- [ ] Types are documented with godoc comments
- [ ] `go vet ./...` passes
