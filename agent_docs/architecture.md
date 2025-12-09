# IntegriDB Architecture

This document provides detailed architectural information about IntegriDB's design and implementation.

## 1. High-Level Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Application   │────▶│    IntegriDB    │────▶│    Database     │
│                 │     │ (sql.DB wrapper)│     │   (PostgreSQL)  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

IntegriDB wraps Go's standard `database/sql` package. Applications use familiar SQL statements while IntegriDB transparently captures all mutations as immutable events.

## 2. Database Structure

The database contains two categories of tables: **IntegriDB system tables** (for event storage) and **User tables** (your application's tables). They coexist in the same database.

```
┌─────────────────────────────────────────────────────────┐
│                      Database                           │
│                    (PostgreSQL)                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │              IntegriDB Tables                   │   │
│  │  ┌─────────────────┐  ┌─────────────────────┐   │   │
│  │  │  integridb_     │  │  integridb_         │   │   │
│  │  │  events         │  │  tracked_tables     │   │   │
│  │  │  (Event Store)  │  │  (Registry)         │   │   │
│  │  └─────────────────┘  └─────────────────────┘   │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │          User Tables (Your Schema)              │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │   │
│  │  │   users     │  │   orders    │  │ products│  │   │
│  │  │  (tracked)  │  │  (tracked)  │  │  (...)  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────┘  │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

User tables remain unchanged. IntegriDB only adds its own system tables and does not modify your schema.

## 3. Event Store vs User Tables

| Aspect | Event Store (`integridb_events`) | User Tables (e.g., `users`) |
|--------|----------------------------------|---------------------------|
| **Purpose** | Audit log, source of truth for history | Application data storage |
| **Data** | All historical changes | Current state only |
| **Mutability** | Immutable (append-only) | Mutable (normal CRUD) |
| **Schema** | Fixed (payload is JSONB) | User-defined |
| **Managed by** | IntegriDB | User/Application |

**Event Store** records every change that has ever occurred to tracked tables, linked together with cryptographic hashes to detect tampering.

**User Tables** are your normal application tables. IntegriDB does not modify them; it only observes changes through the SQL wrapper.

## 4. Data Flow

### 4.1 Write Path

When you execute `db.Exec("UPDATE users SET name = $1 WHERE id = $2", "John", "user_123")`:

1. IntegriDB parses SQL to detect it's an UPDATE on tracked table `users`
2. IntegriDB queries current state of the row (`before`)
3. IntegriDB executes the original UPDATE statement
4. IntegriDB queries new state of the row (`after`)
5. IntegriDB calculates checksum and appends event to `integridb_events`

```
integridb_events table:
┌────┬───────────┬───────────┬────────┬─────────────────────────────────┐
│ id │ table     │ row_id    │ type   │ payload                         │
├────┼───────────┼───────────┼────────┼─────────────────────────────────┤
│ 1  │ users     │ user_123  │ INSERT │ {"after": {"name": "Jane"}}     │
│ 2  │ users     │ user_123  │ UPDATE │ {"before": {"name": "Jane"},    │
│    │           │           │        │  "after": {"name": "John"}}     │
└────┴───────────┴───────────┴────────┴─────────────────────────────────┘

users table (unchanged - your schema):
┌──────────┬──────┬─────────────────────┐
│ id       │ name │ updated_at          │
├──────────┼──────┼─────────────────────┤
│ user_123 │ John │ 2025-01-15 10:30:00 │
└──────────┴──────┴─────────────────────┘
```

### 4.2 Read Path (normal queries)

```
┌─────────┐    ┌──────────┐    ┌─────────────┐
│  App    │───▶│ IntegriDB│───▶│ User Tables │  ← Direct pass-through
│         │    │ (wrapper)│    │             │
└─────────┘    └──────────┘    └─────────────┘
```

SELECT queries pass through directly to user tables with no interception.

### 4.3 Read Path (history/replay)

```
┌─────────┐    ┌──────────┐    ┌─────────────────┐
│  App    │───▶│ IntegriDB│───▶│ Event Store     │  ← Replay events
│         │    │          │    │                 │
└─────────┘    └──────────┘    └─────────────────┘
```

History queries and state replay read from the Event Store and reconstruct state by applying events sequentially.

### 4.4 Verification Path

Load event chain → Recalculate hashes → Compare with stored hashes → Report integrity status

## 5. Why This Separation?

**Event Store provides:**

- Complete audit trail (who changed what, when)
- Tamper detection via hash chain
- Ability to reconstruct any past state
- Compliance and regulatory requirements

**User Tables provide:**

- Normal application data storage
- Fast queries with your existing indexes
- Familiar schema you control
- No changes required to existing table structures

## 6. Summary

| Question | Answer |
|----------|--------|
| Does IntegriDB modify my tables? | No, it only adds its own system tables |
| Can I use normal SQL? | Yes, IntegriDB wraps `database/sql` |
| Which tables are tracked? | Only tables you explicitly configure |
| Which is the source of truth? | User tables for current state, Event Store for history |
| Can state be reconstructed? | Yes, by replaying events from Event Store |
