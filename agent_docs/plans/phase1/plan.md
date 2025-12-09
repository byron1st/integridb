# Phase 1 Implementation Plan

This document provides a detailed, actionable plan for implementing IntegriDB Phase 1 (MVP).

## Overview

**Goal**: Build a Go library that wraps `database/sql` to provide transparent event sourcing with hash chain integrity for PostgreSQL.

**Key Deliverables**:
- `database/sql` wrapper with familiar API
- SQL parser for mutation detection
- Event store with hash chain integrity
- State replay and integrity verification
- PostgreSQL support

---

## Week 1: Project Setup & Foundation

### Task 1.1: Initialize Go Module and Project Structure

**Description**: Set up the Go module and create the directory structure.

**Actions**:
1. Initialize Go module: `go mod init github.com/byron1st/integridb`
2. Create directory structure:
   ```
   integridb/
   ├── pkg/
   │   ├── integridb/      # Public API
   │   ├── event/          # Event store
   │   ├── projection/     # State replay
   │   ├── query/          # SQL parsing
   │   ├── integrity/      # Verification
   │   └── adapter/
   │       └── postgres/   # PostgreSQL adapter
   ├── internal/
   │   ├── hash/           # Hashing utilities
   │   └── serialize/      # JSON serialization
   ├── examples/
   └── docs/
   ```
3. Add initial dependencies:
   - `github.com/lib/pq` (PostgreSQL driver)
   - `github.com/stretchr/testify` (testing, optional)

**Acceptance Criteria**:
- [ ] `go mod tidy` runs without errors
- [ ] Directory structure exists
- [ ] Empty placeholder files compile: `go build ./...`

---

### Task 1.2: Define Core Types

**Description**: Create the foundational type definitions in `pkg/integridb/`.

**Actions**:
1. Create `pkg/integridb/types.go`:
   ```go
   // Event represents an immutable record of a state change
   type Event struct {
       ID            int64
       TableName     string
       RowID         string
       EventType     EventType  // INSERT, UPDATE, DELETE
       Version       int64
       Payload       EventPayload
       Metadata      map[string]any
       Checksum      string
       PrevChecksum  string
       CreatedAt     time.Time
   }

   // EventType enumeration
   type EventType string
   const (
       EventTypeInsert EventType = "INSERT"
       EventTypeUpdate EventType = "UPDATE"
       EventTypeDelete EventType = "DELETE"
   )

   // EventPayload contains before/after states
   type EventPayload struct {
       Before         map[string]any `json:"before,omitempty"`
       After          map[string]any `json:"after,omitempty"`
       ChangedColumns []string       `json:"changed_columns,omitempty"`
   }
   ```

2. Create `pkg/integridb/config.go`:
   ```go
   // Config holds IntegriDB configuration
   type Config struct {
       Driver          string
       DSN             string
       TrackedTables   []TableConfig
       MaxOpenConns    int
       MaxIdleConns    int
       ConnMaxLifetime time.Duration
       MetadataFunc    func(ctx context.Context) map[string]any
   }

   // TableConfig defines tracking configuration for a table
   type TableConfig struct {
       Name       string
       PrimaryKey string
   }
   ```

3. Create `pkg/integridb/errors.go`:
   ```go
   var (
       ErrTableNotTracked    = errors.New("table is not tracked")
       ErrRowNotFound        = errors.New("row not found")
       ErrIntegrityViolation = errors.New("integrity violation detected")
       ErrInvalidSQL         = errors.New("invalid SQL statement")
   )
   ```

4. Create `pkg/integridb/report.go`:
   ```go
   // IntegrityReport contains verification results
   type IntegrityReport struct {
       Valid          bool
       TableName      string
       RowID          string
       EventsVerified int64
       FirstEventID   int64
       LastEventID    int64
       BrokenAt       *BrokenChainInfo
       Errors         []IntegrityError
       VerifiedAt     time.Time
       Duration       time.Duration
   }

   type BrokenChainInfo struct {
       EventID          int64
       ExpectedChecksum string
       ActualChecksum   string
   }

   type IntegrityError struct {
       EventID int64
       Message string
   }
   ```

**Acceptance Criteria**:
- [ ] All types compile without errors
- [ ] Types are documented with godoc comments
- [ ] `go vet ./...` passes

---

### Task 1.3: Implement PostgreSQL Schema Migration

**Description**: Create schema for events table and tracked tables registry.

**Actions**:
1. Create `pkg/adapter/postgres/schema.go`:
   ```go
   const createEventsTable = `
   CREATE TABLE IF NOT EXISTS integridb_events (
       id BIGSERIAL PRIMARY KEY,
       table_name VARCHAR(255) NOT NULL,
       row_id VARCHAR(255) NOT NULL,
       event_type VARCHAR(20) NOT NULL,
       event_version BIGINT NOT NULL,
       payload JSONB NOT NULL,
       metadata JSONB,
       checksum VARCHAR(64) NOT NULL,
       prev_checksum VARCHAR(64),
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
       CONSTRAINT unique_table_row_version UNIQUE (table_name, row_id, event_version)
   );

   CREATE INDEX IF NOT EXISTS idx_events_table_row_version
       ON integridb_events (table_name, row_id, event_version);
   CREATE INDEX IF NOT EXISTS idx_events_created_at
       ON integridb_events (created_at);
   CREATE INDEX IF NOT EXISTS idx_events_checksum
       ON integridb_events (checksum);
   `

   const createTrackedTablesTable = `
   CREATE TABLE IF NOT EXISTS integridb_tracked_tables (
       id BIGSERIAL PRIMARY KEY,
       table_name VARCHAR(255) NOT NULL UNIQUE,
       primary_key VARCHAR(255) NOT NULL,
       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );
   `
   ```

2. Create `pkg/adapter/postgres/adapter.go`:
   ```go
   type PostgresAdapter struct {
       db *sql.DB
   }

   func NewPostgresAdapter(db *sql.DB) *PostgresAdapter
   func (a *PostgresAdapter) Migrate(ctx context.Context) error
   func (a *PostgresAdapter) InsertEvent(ctx context.Context, event *Event) error
   func (a *PostgresAdapter) GetEvents(ctx context.Context, table, rowID string) ([]Event, error)
   func (a *PostgresAdapter) GetLastEvent(ctx context.Context, table, rowID string) (*Event, error)
   ```

3. Write unit tests in `pkg/adapter/postgres/adapter_test.go`

**Acceptance Criteria**:
- [ ] Schema creates tables successfully on PostgreSQL
- [ ] Migration is idempotent (can run multiple times)
- [ ] Unit tests pass
- [ ] Integration test with real PostgreSQL passes

---

### Task 1.4: Implement database/sql Wrapper Skeleton

**Description**: Create the wrapper types that mirror `database/sql` interfaces.

**Actions**:
1. Create `pkg/integridb/db.go`:
   ```go
   // DB wraps *sql.DB with event capture capabilities
   type DB struct {
       inner         *sql.DB
       adapter       Adapter
       trackedTables map[string]TableConfig
       metadataFunc  func(ctx context.Context) map[string]any
   }

   // Open creates a new IntegriDB connection
   func Open(driver, dsn string, config Config) (*DB, error)

   // Standard sql.DB methods (pass-through initially)
   func (db *DB) Exec(query string, args ...any) (sql.Result, error)
   func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
   func (db *DB) Query(query string, args ...any) (*sql.Rows, error)
   func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
   func (db *DB) QueryRow(query string, args ...any) *sql.Row
   func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
   func (db *DB) Begin() (*Tx, error)
   func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error)
   func (db *DB) Prepare(query string) (*sql.Stmt, error)
   func (db *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
   func (db *DB) Close() error
   func (db *DB) Ping() error
   func (db *DB) PingContext(ctx context.Context) error
   ```

2. Create `pkg/integridb/tx.go`:
   ```go
   // Tx wraps *sql.Tx with event capture capabilities
   type Tx struct {
       inner         *sql.Tx
       db            *DB
       pendingEvents []*Event  // Events to commit
   }

   func (tx *Tx) Exec(query string, args ...any) (sql.Result, error)
   func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
   func (tx *Tx) Query(query string, args ...any) (*sql.Rows, error)
   func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
   func (tx *Tx) QueryRow(query string, args ...any) *sql.Row
   func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
   func (tx *Tx) Prepare(query string) (*sql.Stmt, error)
   func (tx *Tx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
   func (tx *Tx) Commit() error
   func (tx *Tx) Rollback() error
   ```

**Acceptance Criteria**:
- [ ] `integridb.Open()` successfully connects to PostgreSQL
- [ ] Basic pass-through operations work (SELECT, INSERT without event capture)
- [ ] `db.Close()` properly closes connection
- [ ] Unit tests pass

---

## Week 2: SQL Parser & Event Store

### Task 2.1: Implement SQL Parser for Mutation Detection

**Description**: Parse SQL statements to identify INSERT, UPDATE, DELETE operations and extract table names.

**Actions**:
1. Create `pkg/query/parser.go`:
   ```go
   // ParsedQuery contains parsed SQL information
   type ParsedQuery struct {
       Type      QueryType  // SELECT, INSERT, UPDATE, DELETE, OTHER
       TableName string
       Columns   []string   // For INSERT: columns being inserted
                            // For UPDATE: columns being updated
       WhereClause string   // For UPDATE/DELETE: WHERE clause for before state
   }

   type QueryType int
   const (
       QueryTypeSelect QueryType = iota
       QueryTypeInsert
       QueryTypeUpdate
       QueryTypeDelete
       QueryTypeOther
   )

   // Parse analyzes a SQL statement
   func Parse(query string) (*ParsedQuery, error)
   ```

2. Implement parsing logic:
   - Use regex or simple tokenization (avoid heavy parser libraries)
   - Handle basic cases: `INSERT INTO table`, `UPDATE table SET`, `DELETE FROM table`
   - Extract table name reliably
   - Handle quoted identifiers
   - Return `QueryTypeOther` for unsupported/complex queries

3. Write comprehensive tests in `pkg/query/parser_test.go`:
   - Test INSERT with various formats
   - Test UPDATE with WHERE clauses
   - Test DELETE with WHERE clauses
   - Test SELECT (should return QueryTypeSelect)
   - Test edge cases (quoted names, schemas, etc.)

**Acceptance Criteria**:
- [ ] Parser correctly identifies INSERT, UPDATE, DELETE, SELECT
- [ ] Parser extracts table names accurately
- [ ] Parser handles PostgreSQL-specific syntax
- [ ] Unit tests cover >90% of parser code
- [ ] Parser returns clear errors for unparseable queries

---

### Task 2.2: Implement Hash Chain Logic

**Description**: Implement deterministic checksum calculation for events.

**Actions**:
1. Create `internal/serialize/json.go`:
   ```go
   // SerializeDeterministic converts a value to JSON with sorted keys
   func SerializeDeterministic(v any) ([]byte, error)
   ```

2. Create `internal/hash/checksum.go`:
   ```go
   // Calculate computes SHA-256 checksum for an event
   // checksum = SHA256(JSON(payload) + JSON(metadata) + prev_checksum)
   func Calculate(payload EventPayload, metadata map[string]any, prevChecksum string) (string, error)

   // Verify checks if a checksum is valid
   func Verify(payload EventPayload, metadata map[string]any, prevChecksum, expectedChecksum string) (bool, error)
   ```

3. Write tests in `internal/hash/checksum_test.go`:
   - Test deterministic output (same input = same hash)
   - Test chain linking (prev_checksum included)
   - Test empty metadata handling
   - Test nil values in payload

**Acceptance Criteria**:
- [ ] Same payload always produces same checksum
- [ ] Different payloads produce different checksums
- [ ] Checksum includes prev_checksum in calculation
- [ ] JSON serialization is deterministic (sorted keys)
- [ ] Unit tests pass

---

### Task 2.3: Implement Event Store

**Description**: Build the event store component for persisting and querying events.

**Actions**:
1. Create `pkg/event/store.go`:
   ```go
   // Store handles event persistence and retrieval
   type Store struct {
       adapter Adapter
   }

   func NewStore(adapter Adapter) *Store

   // AppendEvent adds a new event with checksum calculation
   func (s *Store) AppendEvent(ctx context.Context, tableName, rowID string,
       eventType EventType, payload EventPayload, metadata map[string]any) (*Event, error)

   // GetEvents retrieves all events for a row
   func (s *Store) GetEvents(ctx context.Context, tableName, rowID string) ([]Event, error)

   // GetEventsInRange retrieves events within version range
   func (s *Store) GetEventsInRange(ctx context.Context, tableName, rowID string,
       fromVersion, toVersion int64) ([]Event, error)

   // GetEventsByTime retrieves events up to a timestamp
   func (s *Store) GetEventsByTime(ctx context.Context, tableName, rowID string,
       until time.Time) ([]Event, error)

   // GetLastEvent retrieves the most recent event for a row
   func (s *Store) GetLastEvent(ctx context.Context, tableName, rowID string) (*Event, error)
   ```

2. Implement version numbering:
   - Get last event version for row
   - Increment version for new event
   - Handle first event (version = 1, prev_checksum = "")

3. Write tests in `pkg/event/store_test.go`

**Acceptance Criteria**:
- [ ] Events are persisted correctly
- [ ] Version numbers auto-increment per row
- [ ] Hash chain links events correctly
- [ ] Queries return events in correct order
- [ ] Integration tests with PostgreSQL pass

---

## Week 3: Event Capture Implementation

### Task 3.1: Implement Before/After State Capture

**Description**: Query row state before and after mutations.

**Actions**:
1. Create `pkg/integridb/capture.go`:
   ```go
   // captureBeforeState queries current row state before mutation
   func (db *DB) captureBeforeState(ctx context.Context, tx *sql.Tx,
       tableName, primaryKey string, whereClause string, args []any) (map[string]any, error)

   // captureAfterState queries row state after mutation
   func (db *DB) captureAfterState(ctx context.Context, tx *sql.Tx,
       tableName, primaryKey string, rowID string) (map[string]any, error)

   // extractRowID extracts the primary key value from args or RETURNING clause
   func extractRowID(parsedQuery *ParsedQuery, args []any, result sql.Result) (string, error)
   ```

2. Handle column discovery:
   - Query `information_schema.columns` to get column list
   - Cache column metadata per table
   - Build SELECT query dynamically

3. Handle multiple rows affected:
   - For UPDATE/DELETE: capture all affected row IDs first
   - Create separate events for each row
   - Use transactions to ensure atomicity

**Acceptance Criteria**:
- [ ] Before state captured accurately for UPDATE/DELETE
- [ ] After state captured accurately for INSERT/UPDATE
- [ ] Works with various column types (string, int, timestamp, etc.)
- [ ] Handles NULL values correctly
- [ ] Integration tests pass

---

### Task 3.2: Implement INSERT Event Capture

**Description**: Capture INSERT operations as events.

**Actions**:
1. Update `db.ExecContext()` to detect INSERT:
   ```go
   func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
       parsed, err := query.Parse(query)
       if err != nil {
           return db.inner.ExecContext(ctx, query, args...)
       }

       if parsed.Type == query.QueryTypeInsert && db.isTracked(parsed.TableName) {
           return db.execInsert(ctx, parsed, query, args)
       }

       return db.inner.ExecContext(ctx, query, args...)
   }
   ```

2. Implement `execInsert()`:
   - Execute original INSERT
   - Extract row ID from result or RETURNING clause
   - Query after state
   - Create event with `before: nil, after: state`
   - Append event to store

3. Handle RETURNING clause:
   - Parse RETURNING to get row ID
   - Support `INSERT ... RETURNING id`

**Acceptance Criteria**:
- [ ] INSERT creates event with correct after state
- [ ] Row ID extracted correctly
- [ ] Event has correct version (1 for new row)
- [ ] Checksum calculated correctly
- [ ] Integration tests pass

---

### Task 3.3: Implement UPDATE Event Capture

**Description**: Capture UPDATE operations as events.

**Actions**:
1. Implement `execUpdate()`:
   - Parse WHERE clause to identify affected rows
   - Query before state for all affected rows
   - Execute original UPDATE
   - Query after state for all affected rows
   - Detect changed columns by comparing before/after
   - Create events for each affected row

2. Handle changed columns detection:
   ```go
   func detectChangedColumns(before, after map[string]any) []string {
       var changed []string
       for col, afterVal := range after {
           if beforeVal, ok := before[col]; ok {
               if !reflect.DeepEqual(beforeVal, afterVal) {
                   changed = append(changed, col)
               }
           }
       }
       return changed
   }
   ```

**Acceptance Criteria**:
- [ ] UPDATE creates event with before and after states
- [ ] Changed columns accurately detected
- [ ] Works with multiple rows affected
- [ ] Version increments correctly
- [ ] Hash chain maintained
- [ ] Integration tests pass

---

### Task 3.4: Implement DELETE Event Capture

**Description**: Capture DELETE operations as events.

**Actions**:
1. Implement `execDelete()`:
   - Parse WHERE clause to identify affected rows
   - Query before state for all affected rows
   - Execute original DELETE
   - Create events with `before: state, after: nil`

**Acceptance Criteria**:
- [ ] DELETE creates event with correct before state
- [ ] Works with multiple rows affected
- [ ] Version increments correctly
- [ ] Hash chain maintained
- [ ] Integration tests pass

---

## Week 4: Transaction Support & Metadata

### Task 4.1: Implement Transaction Wrapper

**Description**: Wrap `sql.Tx` to capture events within transactions.

**Actions**:
1. Update `Tx` struct to buffer events:
   ```go
   type Tx struct {
       inner         *sql.Tx
       db            *DB
       ctx           context.Context
       pendingEvents []*pendingEvent
   }

   type pendingEvent struct {
       tableName string
       rowID     string
       eventType EventType
       payload   EventPayload
       metadata  map[string]any
   }
   ```

2. Implement transaction methods:
   - `Exec()`: Capture event data but don't persist yet
   - `Commit()`: Persist all pending events, then commit
   - `Rollback()`: Discard pending events, rollback

3. Handle event ordering within transaction:
   - Track operation order
   - Persist events in same order as operations
   - Ensure hash chain integrity across transaction

4. Handle nested operations:
   - INSERT then UPDATE on same row in same transaction
   - Should create two separate events

**Acceptance Criteria**:
- [ ] Events persist only on Commit
- [ ] Events discarded on Rollback
- [ ] Multiple operations in transaction work correctly
- [ ] Hash chain correct across transaction
- [ ] Integration tests pass

---

### Task 4.2: Implement Metadata Injection

**Description**: Allow users to inject metadata into events via context.

**Actions**:
1. Create `pkg/integridb/context.go`:
   ```go
   // MetadataKey is the context key for event metadata
   type metadataKeyType struct{}
   var MetadataKey = metadataKeyType{}

   // WithMetadata adds metadata to context
   func WithMetadata(ctx context.Context, metadata map[string]any) context.Context {
       return context.WithValue(ctx, MetadataKey, metadata)
   }

   // GetMetadata retrieves metadata from context
   func GetMetadata(ctx context.Context) map[string]any {
       if v := ctx.Value(MetadataKey); v != nil {
           return v.(map[string]any)
       }
       return nil
   }
   ```

2. Update event capture to include metadata:
   - Check context for metadata
   - Merge with MetadataFunc output if configured
   - Include in checksum calculation

3. Write tests for metadata injection

**Acceptance Criteria**:
- [ ] Metadata from context included in events
- [ ] Metadata from MetadataFunc included in events
- [ ] Metadata included in checksum calculation
- [ ] Unit tests pass

---

### Task 4.3: Implement Table Management

**Description**: Allow dynamic table tracking management.

**Actions**:
1. Implement in `pkg/integridb/db.go`:
   ```go
   // TrackTable starts tracking a table
   func (db *DB) TrackTable(ctx context.Context, config TableConfig) error

   // UntrackTable stops tracking a table
   func (db *DB) UntrackTable(ctx context.Context, tableName string) error

   // ListTrackedTables returns all tracked tables
   func (db *DB) ListTrackedTables(ctx context.Context) ([]TableConfig, error)

   // isTracked checks if a table is being tracked
   func (db *DB) isTracked(tableName string) bool
   ```

2. Persist tracked tables in `integridb_tracked_tables`

3. Load tracked tables on startup

**Acceptance Criteria**:
- [ ] Tables can be tracked dynamically
- [ ] Tables can be untracked
- [ ] Tracked tables persisted across restarts
- [ ] Integration tests pass

---

## Week 5: Verification & Replay

### Task 5.1: Implement Integrity Verification

**Description**: Verify hash chain integrity for events.

**Actions**:
1. Create `pkg/integrity/verifier.go`:
   ```go
   // Verifier handles integrity checking
   type Verifier struct {
       store *event.Store
   }

   func NewVerifier(store *event.Store) *Verifier

   // VerifyRow verifies all events for a single row
   func (v *Verifier) VerifyRow(ctx context.Context, tableName, rowID string) (*IntegrityReport, error)

   // VerifyTable verifies all rows in a table
   func (v *Verifier) VerifyTable(ctx context.Context, tableName string) (*IntegrityReport, error)
   ```

2. Implement verification logic:
   ```go
   func (v *Verifier) verifyChain(events []Event) (*IntegrityReport, error) {
       report := &IntegrityReport{Valid: true}
       var prevChecksum string

       for i, event := range events {
           // Verify prev_checksum link
           if event.PrevChecksum != prevChecksum {
               report.Valid = false
               report.BrokenAt = &BrokenChainInfo{...}
               break
           }

           // Recalculate checksum
           calculated, _ := hash.Calculate(event.Payload, event.Metadata, prevChecksum)
           if calculated != event.Checksum {
               report.Valid = false
               report.Errors = append(report.Errors, IntegrityError{...})
               break
           }

           prevChecksum = event.Checksum
       }

       return report, nil
   }
   ```

3. Expose via `DB`:
   ```go
   func (db *DB) VerifyIntegrity(ctx context.Context, tableName, rowID string) (*IntegrityReport, error)
   func (db *DB) VerifyTable(ctx context.Context, tableName string) (*IntegrityReport, error)
   ```

**Acceptance Criteria**:
- [ ] Valid chains report as valid
- [ ] Tampered events detected
- [ ] Broken chain location identified
- [ ] Performance acceptable for large chains
- [ ] Integration tests pass

---

### Task 5.2: Implement State Replay by Version

**Description**: Reconstruct row state at a specific version.

**Actions**:
1. Create `pkg/projection/replay.go`:
   ```go
   // Replayer handles state reconstruction
   type Replayer struct {
       store *event.Store
   }

   func NewReplayer(store *event.Store) *Replayer

   // ReplayToVersion reconstructs state at specific version
   func (r *Replayer) ReplayToVersion(ctx context.Context,
       tableName, rowID string, version int64) (map[string]any, error)
   ```

2. Implement replay logic:
   ```go
   func (r *Replayer) ReplayToVersion(...) (map[string]any, error) {
       events, err := r.store.GetEventsInRange(ctx, tableName, rowID, 1, version)
       if err != nil {
           return nil, err
       }

       var state map[string]any
       for _, event := range events {
           switch event.EventType {
           case EventTypeInsert:
               state = event.Payload.After
           case EventTypeUpdate:
               state = event.Payload.After
           case EventTypeDelete:
               state = nil
           }
       }

       return state, nil
   }
   ```

3. Expose via `DB`:
   ```go
   func (db *DB) ReplayTo(ctx context.Context, tableName, rowID string, version int64) (map[string]any, error)
   ```

**Acceptance Criteria**:
- [ ] Correct state returned for any version
- [ ] Returns nil for deleted rows
- [ ] Error for non-existent rows
- [ ] Integration tests pass

---

### Task 5.3: Implement State Replay by Time

**Description**: Reconstruct row state at a specific timestamp.

**Actions**:
1. Add to `pkg/projection/replay.go`:
   ```go
   // ReplayToTime reconstructs state at specific timestamp
   func (r *Replayer) ReplayToTime(ctx context.Context,
       tableName, rowID string, timestamp time.Time) (map[string]any, error)
   ```

2. Implementation:
   - Query events up to timestamp
   - Apply same replay logic as version-based

3. Expose via `DB`:
   ```go
   func (db *DB) ReplayToTime(ctx context.Context, tableName, rowID string, timestamp time.Time) (map[string]any, error)
   ```

**Acceptance Criteria**:
- [ ] Correct state returned for any timestamp
- [ ] Works with timezone handling
- [ ] Integration tests pass

---

### Task 5.4: Implement Event History Retrieval

**Description**: Expose event history to users.

**Actions**:
1. Add to `pkg/integridb/db.go`:
   ```go
   // GetHistory retrieves all events for a row
   func (db *DB) GetHistory(ctx context.Context, tableName, rowID string) ([]Event, error)

   // GetHistoryRange retrieves events in version range
   func (db *DB) GetHistoryRange(ctx context.Context, tableName, rowID string,
       fromVersion, toVersion int64) ([]Event, error)
   ```

**Acceptance Criteria**:
- [ ] Full history returned correctly
- [ ] Range queries work correctly
- [ ] Events ordered by version
- [ ] Integration tests pass

---

## Week 6: Testing, Documentation & Examples

### Task 6.1: Comprehensive Unit Tests

**Description**: Achieve >80% code coverage with unit tests.

**Actions**:
1. Review and add tests for all packages:
   - `pkg/integridb/` - API tests
   - `pkg/event/` - Event store tests
   - `pkg/query/` - SQL parser tests
   - `pkg/integrity/` - Verification tests
   - `pkg/projection/` - Replay tests
   - `internal/hash/` - Checksum tests
   - `internal/serialize/` - JSON tests

2. Add edge case tests:
   - NULL values in columns
   - Unicode characters
   - Large payloads
   - Concurrent operations
   - Error conditions

3. Run coverage: `go test -cover ./...`

**Acceptance Criteria**:
- [ ] >80% code coverage
- [ ] All edge cases covered
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes

---

### Task 6.2: Integration Tests

**Description**: Test with real PostgreSQL instances.

**Actions**:
1. Create `integration_test.go` files using testcontainers-go:
   ```go
   func TestIntegration_FullWorkflow(t *testing.T) {
       // Start PostgreSQL container
       // Run migrations
       // Test INSERT/UPDATE/DELETE with event capture
       // Test verification
       // Test replay
   }
   ```

2. Test scenarios:
   - Full CRUD workflow
   - Transaction commit/rollback
   - Concurrent operations
   - Large number of events
   - Integrity verification after tampering

**Acceptance Criteria**:
- [ ] All integration tests pass
- [ ] Tests are reproducible
- [ ] Tests clean up after themselves

---

### Task 6.3: Documentation

**Description**: Create user-facing documentation.

**Actions**:
1. Create `docs/getting-started.md`:
   - Installation instructions
   - Basic usage example
   - Configuration options

2. Create `docs/api-reference.md`:
   - All public types
   - All public methods
   - Code examples

3. Update README.md with:
   - Project description
   - Quick start example
   - Link to documentation

**Acceptance Criteria**:
- [ ] Getting started guide complete
- [ ] API reference complete
- [ ] README updated
- [ ] All examples compile and run

---

### Task 6.4: Example Applications

**Description**: Create working example applications.

**Actions**:
1. Create `examples/basic/main.go`:
   - Simple CRUD operations
   - Event history retrieval
   - Integrity verification

2. Create `examples/transactions/main.go`:
   - Transaction usage
   - Rollback handling

3. Create `examples/metadata/main.go`:
   - Metadata injection via context
   - MetadataFunc configuration

**Acceptance Criteria**:
- [ ] All examples compile
- [ ] All examples run successfully with PostgreSQL
- [ ] Examples are well-commented

---

## Deliverables Checklist

- [ ] `database/sql` wrapper implementation
- [ ] SQL parser for mutation detection (INSERT, UPDATE, DELETE)
- [ ] Event store with hash chain integrity
- [ ] Transparent event capture for tracked tables
- [ ] Transaction support with event batching
- [ ] State replay by version and by time
- [ ] Integrity verification (single row and table-wide)
- [ ] PostgreSQL driver support
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests with PostgreSQL
- [ ] Getting started documentation
- [ ] API reference documentation
- [ ] Example applications

---

## Dependencies

### External Go Packages
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/testcontainers/testcontainers-go` - Integration testing (dev only)

### Development Tools
- Go 1.21+
- PostgreSQL 14+
- Docker (for testcontainers)
- golangci-lint

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| SQL parsing complexity | Start with simple regex, upgrade to parser library if needed |
| Performance overhead | Benchmark early, optimize event insertion path |
| Multi-row updates | Test thoroughly, use transactions for atomicity |
| Hash chain corruption | Implement verification early, test tampering detection |
| Concurrent access | Use database-level locking, test with race detector |
