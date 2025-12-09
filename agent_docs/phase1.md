# 1. MVP (Phase 1)

## 1.1 Scope

**Included:**
- Wrapper for Go's standard `database/sql` package
- Transparent event capture for INSERT, UPDATE, DELETE operations
- Event store with hash chain integrity
- Standard SQL interface (users write normal SQL)
- Transaction support via wrapped `sql.Tx`
- State replay by version and by time
- Integrity verification
- PostgreSQL driver

**Not Included:**
- Merkle tree (Phase 2)
- Snapshots (Phase 2)
- External storage (Phase 2)
- MySQL, SQLite3 drivers (Future)
- Java SDK (Future)

## 1.2 Design Philosophy

IntegriDB wraps Go's standard `database/sql` package rather than providing ORM-like methods. This approach offers several benefits:

- **Familiar API**: Developers use standard SQL they already know
- **No learning curve**: Same `Query`, `Exec`, `QueryRow` methods as `database/sql`
- **Full SQL power**: No limitations on complex queries, joins, or database-specific features
- **Easy adoption**: Minimal code changes to integrate into existing projects
- **Transparent capture**: Event logging happens automatically without changing query patterns

## 1.3 Database Schema

### 1.3.1 Events Table

Stores all state changes as immutable records.

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| id | BIGSERIAL | No | Auto-incrementing primary key |
| table_name | VARCHAR(255) | No | Name of the affected table |
| row_id | VARCHAR(255) | No | Primary key value of the affected row |
| event_type | VARCHAR(20) | No | Type of change: INSERT, UPDATE, or DELETE |
| event_version | BIGINT | No | Sequential version number per row |
| payload | JSONB | No | Event data containing before/after states and changed fields |
| metadata | JSONB | Yes | Additional context (user_id, correlation_id, etc.) |
| checksum | VARCHAR(64) | No | SHA-256 hash of payload + metadata + prev_checksum |
| prev_checksum | VARCHAR(64) | Yes | Checksum of previous event (NULL for first event) |
| created_at | TIMESTAMPTZ | No | Timestamp of event creation |

**Constraints:**
- Unique constraint on (table_name, row_id, event_version)

**Indexes:**
- (table_name, row_id, event_version) for event lookups
- (created_at) for time-based queries
- (checksum) for integrity verification

### 1.3.2 Tracked Tables Registry

Registry of tables being tracked by IntegriDB.

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| id | BIGSERIAL | No | Auto-incrementing primary key |
| table_name | VARCHAR(255) | No | Name of the tracked table |
| primary_key | VARCHAR(255) | No | Primary key column name |
| created_at | TIMESTAMPTZ | No | When tracking started |

### 1.3.3 Event Payload Structure

```
{
    "before": { ... },         // Row state before change (null for INSERT)
    "after": { ... },          // Row state after change (null for DELETE)
    "changed_columns": [ ... ] // List of modified column names (for UPDATE)
}
```

## 1.4 Hash Chain Specification

### 1.4.1 Checksum Calculation

For each event, the checksum is calculated as:

```
checksum = SHA256(JSON(payload) + JSON(metadata) + prev_checksum)
```

Where:
- `payload` is the event payload (before, after, changed_fields)
- `metadata` is the event metadata (user_id, correlation_id, etc.)
- `prev_checksum` is the checksum of the previous event (empty string for first event)
- JSON serialization must be deterministic (sorted keys)

### 1.4.2 Chain Verification

To verify integrity:

1. Load all events for an aggregate, ordered by version
2. For each event starting from the first:
   - Recalculate checksum from payload, metadata, and previous event's checksum
   - Compare with stored checksum
   - Verify prev_checksum matches the previous event's checksum
3. Report any mismatches as integrity violations

## 1.5 Go SDK Specification

### 1.5.1 Overview

IntegriDB provides a drop-in replacement for `database/sql`. Users interact with familiar interfaces (`DB`, `Tx`, `Rows`, `Result`) while IntegriDB transparently captures mutations.

```
Standard database/sql:          IntegriDB wrapper:
--------------------            ------------------
sql.Open(driver, dsn)     →     integridb.Open(driver, dsn, opts)
db.Exec(query, args...)   →     db.Exec(query, args...)      // Same API
db.Query(query, args...)  →     db.Query(query, args...)     // Same API
db.Begin()                →     db.Begin()                   // Same API
tx.Exec(query, args...)   →     tx.Exec(query, args...)      // Same API
tx.Commit()               →     tx.Commit()                  // Same API
```

### 1.5.2 Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| Driver | string | Yes | - | Database driver name (e.g., "postgres") |
| DSN | string | Yes | - | Database connection string |
| TrackedTables | []TableConfig | Yes | - | Tables to track for event capture |
| MaxOpenConns | int | No | 25 | Maximum open connections |
| MaxIdleConns | int | No | 5 | Maximum idle connections |
| ConnMaxLifetime | duration | No | 5m | Maximum connection lifetime |
| MetadataFunc | func() map | No | nil | Function to inject metadata into events |

**TableConfig:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Name | string | Yes | Table name to track |
| PrimaryKey | string | Yes | Primary key column name |

### 1.5.3 Wrapped Interfaces

**integridb.DB** (wraps `*sql.DB`):

| Method | Description |
|--------|-------------|
| `Exec(query, args...) (Result, error)` | Execute INSERT/UPDATE/DELETE with event capture |
| `Query(query, args...) (*Rows, error)` | Execute SELECT query (no event capture) |
| `QueryRow(query, args...) *Row` | Execute SELECT for single row |
| `Begin() (*Tx, error)` | Start transaction with event capture |
| `BeginTx(ctx, opts) (*Tx, error)` | Start transaction with context and options |
| `Prepare(query) (*Stmt, error)` | Prepare statement |
| `Close() error` | Close database connection |
| `Ping() error` | Ping database |
| `PingContext(ctx) error` | Ping database with context |

**integridb.Tx** (wraps `*sql.Tx`):

| Method | Description |
|--------|-------------|
| `Exec(query, args...) (Result, error)` | Execute with event capture |
| `Query(query, args...) (*Rows, error)` | Execute SELECT query |
| `QueryRow(query, args...) *Row` | Execute SELECT for single row |
| `Prepare(query) (*Stmt, error)` | Prepare statement |
| `Commit() error` | Commit transaction and finalize events |
| `Rollback() error` | Rollback transaction and discard events |

### 1.5.4 Event Capture Behavior

IntegriDB parses SQL statements to detect mutations on tracked tables:

| SQL Statement | Event Captured |
|---------------|----------------|
| `INSERT INTO users ...` | INSERT event with `after` state |
| `UPDATE users SET ... WHERE ...` | UPDATE event with `before` and `after` states |
| `DELETE FROM users WHERE ...` | DELETE event with `before` state |
| `SELECT * FROM users ...` | No event (read-only) |
| `INSERT INTO logs ...` (untracked) | No event (table not tracked) |

**Capture Process:**

1. Parse SQL to identify operation type and target table
2. If table is tracked and operation is INSERT/UPDATE/DELETE:
   - For UPDATE/DELETE: Query current state (`before`)
   - Execute the original SQL statement
   - For INSERT/UPDATE: Query new state (`after`)
   - Calculate checksum with hash chain
   - Insert event record
3. If table is not tracked: Execute SQL directly without event capture

### 1.5.5 IntegriDB-Specific Operations

Beyond the standard `database/sql` interface, IntegriDB provides additional methods:

**Event History:**
- `GetHistory(table, rowID) → []Event`: Get all events for a row
- `GetHistoryRange(table, rowID, fromVersion, toVersion) → []Event`: Get events in version range

**State Replay:**
- `ReplayTo(table, rowID, version) → map`: Reconstruct state at specific version
- `ReplayToTime(table, rowID, timestamp) → map`: Reconstruct state at specific time

**Integrity Verification:**
- `VerifyIntegrity(table, rowID) → IntegrityReport`: Verify single row's event chain
- `VerifyTable(table) → IntegrityReport`: Verify all rows in a table

**Table Management:**
- `TrackTable(config TableConfig) error`: Start tracking a table
- `UntrackTable(table string) error`: Stop tracking a table
- `ListTrackedTables() → []TableConfig`: List all tracked tables

### 1.5.6 Context and Metadata

Users can inject metadata into events using context:

```
Context key: integridb.MetadataKey
Value type: map[string]any
```

Metadata is included in event records and in checksum calculation for integrity.

### 1.5.7 Types

**Event:**

| Field | Type | Description |
|-------|------|-------------|
| ID | int64 | Event ID |
| TableName | string | Affected table name |
| RowID | string | Primary key value |
| EventType | string | INSERT, UPDATE, or DELETE |
| Version | int64 | Event version for this row |
| Payload | EventPayload | Before/after states |
| Metadata | map | User-provided context |
| Checksum | string | SHA-256 hash |
| PrevChecksum | string | Previous event's checksum |
| CreatedAt | time | Event timestamp |

**EventPayload:**

| Field | Type | Description |
|-------|------|-------------|
| Before | map | Row state before change (nil for INSERT) |
| After | map | Row state after change (nil for DELETE) |
| ChangedColumns | []string | Modified columns (for UPDATE) |

**IntegrityReport:**

| Field | Type | Description |
|-------|------|-------------|
| Valid | bool | True if integrity verified |
| TableName | string | Table name |
| RowID | string | Row ID (empty for VerifyTable) |
| EventsVerified | int64 | Number of events checked |
| FirstEventID | int64 | First event in chain |
| LastEventID | int64 | Last event in chain |
| BrokenAt | *BrokenChainInfo | Location of break (if any) |
| Errors | []IntegrityError | List of errors found |
| VerifiedAt | time | Verification timestamp |
| Duration | duration | Time taken to verify |

### 1.5.8 Usage Example

```
// Instead of:
db, err := sql.Open("postgres", dsn)

// Use:
db, err := integridb.Open("postgres", dsn, integridb.Config{
    TrackedTables: []integridb.TableConfig{
        {Name: "users", PrimaryKey: "id"},
        {Name: "orders", PrimaryKey: "id"},
    },
})

// Then use standard SQL:
db.Exec("INSERT INTO users (id, name, email) VALUES ($1, $2, $3)", 
    "user_123", "John", "john@example.com")

db.Exec("UPDATE users SET name = $1 WHERE id = $2", 
    "John Doe", "user_123")

// Transactions work the same way:
tx, _ := db.Begin()
tx.Exec("INSERT INTO orders ...")
tx.Exec("UPDATE users SET last_order = $1 WHERE id = $2", ...)
tx.Commit()

// IntegriDB-specific operations:
history, _ := db.GetHistory("users", "user_123")
pastState, _ := db.ReplayTo("users", "user_123", 1)
report, _ := db.VerifyIntegrity("users", "user_123")
```

## 1.6 MVP Deliverables

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

## 1.7 MVP Timeline

| Week | Focus |
|------|-------|
| 1 | Project setup, `database/sql` wrapper skeleton, schema implementation |
| 2 | SQL parser for mutation detection, event store implementation |
| 3 | Hash chain logic, event capture for INSERT/UPDATE/DELETE |
| 4 | Transaction wrapper, metadata injection |
| 5 | Integrity verification, replay functionality |
| 6 | Testing, documentation, examples |
