# Task 1.3: Implement PostgreSQL Schema Migration

## Description
Create database schema for the events table and tracked tables registry, and implement the PostgreSQL adapter for event storage.

## Actions

1. Create `pkg/adapter/postgres/schema.go` with:
   - `integridb_events` table schema with columns for event data, checksums, and indexes
   - `integridb_tracked_tables` table schema for tracking which tables are monitored

2. Create `pkg/adapter/postgres/adapter.go` with:
   - `PostgresAdapter` struct
   - `NewPostgresAdapter()` constructor
   - `Migrate()` method to create tables
   - `InsertEvent()` method to persist events
   - `GetEvents()` method to retrieve events by table and row ID
   - `GetLastEvent()` method to retrieve the most recent event for a row

3. Write unit tests in `pkg/adapter/postgres/adapter_test.go`

## Acceptance Criteria

- [ ] Schema creates tables successfully on PostgreSQL
- [ ] Migration is idempotent (can run multiple times)
- [ ] Unit tests pass
- [ ] Integration test with real PostgreSQL passes
