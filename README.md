# IntegriDB

[![Go Reference](https://pkg.go.dev/badge/github.com/byron1st/integridb.svg)](https://pkg.go.dev/github.com/byron1st/integridb)
[![Go Report Card](https://goreportcard.com/badge/github.com/byron1st/integridb)](https://goreportcard.com/report/github.com/byron1st/integridb)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**IntegriDB** is a Go library that wraps `database/sql` to provide built-in data integrity guarantees through event sourcing. It records every database change as an immutable event with cryptographic verification, giving you complete audit trails and tamper detection out of the box.

## Why IntegriDB?

Traditional databases only store the current state of your data. IntegriDB gives you:

- **Complete Audit Trail**: Know exactly who changed what, when, and why
- **Tamper Detection**: Cryptographic hash chains detect unauthorized modifications
- **Time Travel**: Reconstruct your database state at any point in history
- **Zero Schema Changes**: Works with your existing tables and migrations
- **Familiar API**: Drop-in replacement for `database/sql` - no query rewriting needed

Perfect for applications requiring compliance (HIPAA, SOX, GDPR), financial record-keeping, or any system where data integrity is critical.

## Quick Start

### Installation

```bash
go get github.com/byron1st/integridb
```

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/byron1st/integridb/pkg/integridb"
    _ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
    // Configure IntegriDB
    config := &integridb.Config{
        Driver: "postgres",
        DSN:    "postgres://user:pass@localhost/mydb?sslmode=disable",
        TrackedTables: []integridb.TableConfig{
            {
                Name:       "users",
                PrimaryKey: []string{"id"},
            },
            {
                Name:       "orders",
                PrimaryKey: []string{"id"},
            },
        },
    }

    // Open database connection
    db, err := integridb.Open(config)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Initialize IntegriDB schema (run once)
    ctx := context.Background()
    if err := db.Migrate(ctx); err != nil {
        log.Fatal(err)
    }

    // Use like normal database/sql - mutations are automatically tracked!
    _, err = db.ExecContext(ctx,
        "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
        "user_123", "Alice", "alice@example.com",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Updates are captured with before/after states
    _, err = db.ExecContext(ctx,
        "UPDATE users SET name = $1 WHERE id = $2",
        "Alice Smith", "user_123",
    )

    log.Println("All changes tracked automatically!")
}
```

That's it! Every `INSERT`, `UPDATE`, and `DELETE` on tracked tables is now recorded as an immutable event.

## How It Works

IntegriDB transparently intercepts mutations and stores them in an append-only event log alongside your normal tables:

```
Your Database (PostgreSQL)
├── users              (your table - unchanged)
├── orders             (your table - unchanged)
├── integridb_events   (event store - managed by IntegriDB)
└── integridb_tracked_tables (registry - managed by IntegriDB)
```

**Write Path:**
1. You execute: `UPDATE users SET name = 'Bob' WHERE id = 123`
2. IntegriDB captures the "before" state
3. IntegriDB executes your UPDATE
4. IntegriDB captures the "after" state
5. IntegriDB stores an event with both states + cryptographic hash

**Read Path:**
- `SELECT` queries go directly to your tables (zero overhead)
- History queries replay events from the event store

## Core Features

### 1. Automatic Event Capture

All mutations on tracked tables generate immutable events:

```go
// INSERT: captures "after" state
db.ExecContext(ctx,
    "INSERT INTO users (id, name) VALUES ($1, $2)",
    "user_456", "Bob",
)

// UPDATE: captures "before" and "after" states
db.ExecContext(ctx,
    "UPDATE users SET name = $1 WHERE department = $2",
    "Engineering", "Sales", // Updates ALL users in Sales department
)

// DELETE: captures "before" state
db.ExecContext(ctx,
    "DELETE FROM users WHERE id = $1",
    "user_789",
)
```

IntegriDB automatically handles:
- Single-row operations
- Multi-row updates and deletes
- Composite primary keys
- Complex WHERE clauses

### 2. Transaction Support

Events are buffered and committed atomically with your data:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    log.Fatal(err)
}

// Both operations recorded as events
tx.ExecContext(ctx, "INSERT INTO orders (id, user_id) VALUES ($1, $2)", "ord_1", "user_123")
tx.ExecContext(ctx, "UPDATE users SET last_order = $1 WHERE id = $2", "ord_1", "user_123")

// Events persist atomically on commit
if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

### 3. Metadata Injection

Attach custom metadata to events (user IDs, request IDs, reasons for change):

```go
config := &integridb.Config{
    Driver: "postgres",
    DSN:    "...",
    TrackedTables: []integridb.TableConfig{
        {Name: "users", PrimaryKey: []string{"id"}},
    },
    MetadataFunc: func(ctx context.Context) map[string]any {
        // Extract from context (middleware can populate this)
        userID := ctx.Value("user_id")
        requestID := ctx.Value("request_id")

        return map[string]any{
            "user_id":    userID,
            "request_id": requestID,
            "ip_address": ctx.Value("ip"),
        }
    },
}

// Now every event includes this metadata automatically
db.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", "Charlie", "user_123")
```

### 4. Selective Tracking

Only tables you explicitly configure are tracked. Other tables work normally:

```go
config := &integridb.Config{
    Driver: "postgres",
    DSN:    "...",
    TrackedTables: []integridb.TableConfig{
        {Name: "users", PrimaryKey: []string{"id"}},      // Tracked
        {Name: "orders", PrimaryKey: []string{"id"}},     // Tracked
        // "sessions" table is NOT listed - not tracked
    },
}

db.ExecContext(ctx, "INSERT INTO users ...", ...)    // ✅ Event created
db.ExecContext(ctx, "INSERT INTO sessions ...", ...) // ✅ Executes normally, no event
```

### 5. Composite Primary Keys

Full support for tables with multi-column primary keys:

```go
config := &integridb.Config{
    Driver: "postgres",
    DSN:    "...",
    TrackedTables: []integridb.TableConfig{
        {
            Name:       "order_items",
            PrimaryKey: []string{"order_id", "product_id"}, // Composite key
        },
    },
}
```

### 6. Integrity Verification (Planned)

Verify hash chain integrity to detect tampering:

```go
// Verify entire table's event history
report, err := db.VerifyIntegrity(ctx, "users")
if err != nil {
    log.Fatal(err)
}

if !report.Valid {
    log.Printf("Tampering detected at version %d", report.FirstInvalidVersion)
}
```

### 7. Historical State Reconstruction (Planned)

Reconstruct database state at any point in time:

```go
// Get user state at version 42
state, err := db.ReplayToVersion(ctx, "users", "user_123", 42)

// Get user state at specific timestamp
timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
state, err := db.ReplayToTime(ctx, "users", "user_123", timestamp)
```

### 8. Event History Access (Planned)

Query the complete change history:

```go
// Get all events for a specific row
events, err := db.GetEventHistory(ctx, "users", "user_123")

for _, event := range events {
    fmt.Printf("Version %d at %s: %s\n",
        event.Version,
        event.Timestamp,
        event.Type,
    )
    fmt.Printf("  Before: %+v\n", event.Payload.Before)
    fmt.Printf("  After:  %+v\n", event.Payload.After)
}
```

## Configuration

### Config Options

```go
type Config struct {
    // Database connection
    Driver string  // "postgres" (MySQL and SQLite coming soon)
    DSN    string  // Connection string

    // Event sourcing
    TrackedTables []TableConfig  // Tables to track
    MetadataFunc  func(context.Context) map[string]any  // Optional metadata extraction

    // Connection pool (same as database/sql)
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

type TableConfig struct {
    Name       string    // Table name (case-sensitive)
    PrimaryKey []string  // Column(s) forming the primary key
}
```

### Environment-Specific Configuration

```go
// Development
devConfig := &integridb.Config{
    Driver: "postgres",
    DSN:    "postgres://localhost/dev?sslmode=disable",
    TrackedTables: []integridb.TableConfig{
        {Name: "users", PrimaryKey: []string{"id"}},
    },
}

// Production
prodConfig := &integridb.Config{
    Driver:          "postgres",
    DSN:             os.Getenv("DATABASE_URL"),
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    TrackedTables: []integridb.TableConfig{
        {Name: "users", PrimaryKey: []string{"id"}},
        {Name: "orders", PrimaryKey: []string{"id"}},
        {Name: "transactions", PrimaryKey: []string{"id"}},
    },
    MetadataFunc: extractMetadataFromContext,
}
```

## Database Support

| Database   | Status         | Notes                          |
|------------|----------------|--------------------------------|
| PostgreSQL | ✅ Supported   | Primary target for MVP         |
| MySQL      | 🚧 Planned     | Phase 2                        |
| SQLite     | 🚧 Planned     | Phase 2                        |

## Event Store Schema

IntegriDB creates two system tables:

```sql
-- Event store (append-only)
CREATE TABLE integridb_events (
    id            BIGSERIAL PRIMARY KEY,
    aggregate_id  TEXT NOT NULL,           -- Table name
    row_id        TEXT NOT NULL,           -- Primary key value(s)
    version       BIGINT NOT NULL,         -- Version number within aggregate
    event_type    TEXT NOT NULL,           -- INSERT, UPDATE, DELETE
    payload       JSONB NOT NULL,          -- before/after states
    metadata      JSONB,                   -- User-defined metadata
    checksum      TEXT NOT NULL,           -- SHA-256 hash
    prev_checksum TEXT,                    -- Previous event's hash (hash chain)
    timestamp     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tracked tables registry
CREATE TABLE integridb_tracked_tables (
    table_name   TEXT PRIMARY KEY,
    primary_key  TEXT[] NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Performance Considerations

IntegriDB adds overhead to write operations (INSERT/UPDATE/DELETE) to capture state and persist events:

- **Reads (SELECT)**: No overhead - queries pass through directly
- **Writes**: 2-3x slower due to state capture and event persistence
- **Transactions**: Events buffered in memory, persisted on commit

**Optimization Tips:**

1. **Track selectively**: Only track tables requiring audit trails
2. **Use transactions**: Batch multiple operations to amortize overhead
3. **Index carefully**: Add indexes to your tables' primary keys
4. **Consider async writes** (Phase 2): Event persistence in background workers

Benchmarks (preliminary):

```
Direct INSERT:          1000 ops in 45ms   (22,222 ops/sec)
IntegriDB INSERT:       1000 ops in 120ms  (8,333 ops/sec)
IntegriDB UPDATE:       1000 ops in 180ms  (5,555 ops/sec)
```

## Architecture

```
┌─────────────────────┐
│   Your Application  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   IntegriDB (Wrapper)│
│  ┌───────────────┐  │
│  │ SQL Parser    │  │ Detects mutations
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │ State Capture │  │ Reads before/after
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │ Event Store   │  │ Persists events
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │ Hash Chain    │  │ Links events
│  └───────────────┘  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  PostgreSQL Database│
│  ├─ users          │ (your tables)
│  ├─ orders         │ (your tables)
│  ├─ integridb_events │ (event store)
│  └─ integridb_tracked_tables │ (registry)
└─────────────────────┘
```

## Comparison with Alternatives

| Approach | Pros | Cons |
|----------|------|------|
| **Database Triggers** | Native, automatic | Hard to manage, database-specific, no version control |
| **CDC (Change Data Capture)** | Captures everything | Complex setup, vendor-specific, eventual consistency |
| **Application-Level Logging** | Full control | Easy to miss, inconsistent, hard to enforce |
| **IntegriDB** | Transparent, portable, Go-native | Write overhead, storage cost |

IntegriDB is ideal when you want event sourcing without changing your application logic or database schema.

## Roadmap

### Phase 1: MVP (Current)
- [x] PostgreSQL support
- [x] Event capture for INSERT/UPDATE/DELETE
- [x] Hash chain integrity
- [x] Transaction support
- [ ] Integrity verification
- [ ] State replay (by version and time)
- [ ] Event history API

### Phase 2: Advanced Features
- [ ] Merkle tree for O(log n) verification
- [ ] Snapshot support for faster replay
- [ ] MySQL and SQLite support
- [ ] External storage backends (S3, GCS)
- [ ] Event compaction and archival
- [ ] Async event persistence

### Phase 3: Enterprise Features
- [ ] Event streaming (Kafka, NATS)
- [ ] Multi-database sharding
- [ ] Encryption at rest
- [ ] RBAC for event access

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository
git clone https://github.com/byron1st/integridb.git
cd integridb

# Run tests
make test

# Run linters
make lint

# Generate mocks (after interface changes)
make mockgen

# Full verification (before committing)
make mockgen && make fmt && make lint && make test-ci
```

## License

IntegriDB is released under the [MIT License](LICENSE).

## Support

- **Documentation**: [https://pkg.go.dev/github.com/byron1st/integridb](https://pkg.go.dev/github.com/byron1st/integridb)
- **Issues**: [https://github.com/byron1st/integridb/issues](https://github.com/byron1st/integridb/issues)
- **Discussions**: [https://github.com/byron1st/integridb/discussions](https://github.com/byron1st/integridb/discussions)

## Acknowledgments

IntegriDB is inspired by event sourcing patterns from:
- [Event Sourcing by Martin Fowler](https://martinfowler.com/eaaDev/EventSourcing.html)
- [CQRS Journey by Microsoft](https://docs.microsoft.com/en-us/previous-versions/msp-n-p/jj554200(v=pandp.10))
- [EventStoreDB](https://www.eventstore.com/)
