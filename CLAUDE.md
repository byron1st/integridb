# IntegriDB

A database abstraction layer that ensures data integrity through event sourcing patterns.

## 1.1 What is IntegriDB?

IntegriDB is a Go library that wraps SQL databases to provide built-in data integrity guarantees. Instead of storing only the current state, it records every change as an immutable event, creating an auditable history that can detect tampering and reconstruct any past state.

## 1.2 Why IntegriDB?

Traditional databases store only current state, making it difficult to:

- **Audit changes**: Who changed what, and when?
- **Detect tampering**: Has data been modified outside the application?
- **Reconstruct history**: What was the state at a specific point in time?
- **Prove integrity**: Can you cryptographically verify data hasn't been altered?

IntegriDB solves these problems by:

- Recording all mutations as immutable events
- Linking events with cryptographic hashes to detect tampering
- Providing state replay to reconstruct any historical point
- Supporting snapshots to optimize storage while preserving full history

## 1.3 Architecture

IntegriDB wraps Go's `database/sql` package, transparently capturing mutations as immutable events while maintaining your existing tables and schema.

**Key Design Principles:**

- **Dual Storage**: Event Store (immutable audit log) + User Tables (current state)
- **Non-invasive**: No modifications to your existing database schema
- **Transparent**: Use familiar SQL statements; IntegriDB handles event capture
- **Selective Tracking**: Only tables you explicitly configure are tracked

**Data Flow:**

- **Writes**: Parse SQL → Capture before/after state → Execute mutation → Record event with hash
- **Reads**: SELECT queries pass through directly to user tables
- **History**: Replay events from Event Store to reconstruct past states
- **Verification**: Validate hash chain integrity to detect tampering

For detailed architectural diagrams and explanations, see [agent_docs/architecture.md](agent_docs/architecture.md).

## 1.4 Development Guidelines

When writing code or tests for this project:

- **Code Style**: Follow Go conventions and project standards - see [agent_docs/code_style.md](agent_docs/code_style.md)
- **Testing**: Write comprehensive unit and integration tests - see [agent_docs/test_style.md](agent_docs/test_style.md)

## 1.5 Core Concepts

| Concept | Description |
|---------|-------------|
| **Event** | Immutable record of a state change (INSERT, UPDATE, DELETE) with cryptographic hash |
| **Aggregate** | Logical grouping of related events, typically corresponding to a database table |
| **Projection** | Materialized current state derived from events, optimized for querying |
| **Hash Chain** | Sequential linking of event hashes for tamper detection (MVP) |
| **Merkle Tree** | Tree-based hash structure for efficient O(log n) verification (Phase 2) |
| **Snapshot** | Point-in-time state capture to optimize storage and replay performance (Phase 2) |

## 1.6 Project Structure

```
integridb/
├── cmd/
│   └── integridb-cli/          # CLI tool for migrations and verification
│
├── pkg/
│   ├── integridb/              # Public API (client, config, types)
│   ├── event/                  # Event store and hash chain logic
│   ├── merkle/                 # Merkle tree implementation (Phase 2)
│   ├── snapshot/               # Snapshot management (Phase 2)
│   ├── projection/             # Projection building and sync
│   ├── query/                  # Query builder
│   ├── integrity/              # Integrity verification
│   ├── adapter/                # Database adapters
│   │   ├── postgres/           # PostgreSQL adapter (MVP)
│   │   ├── mysql/              # MySQL adapter (Future)
│   │   └── sqlite/             # SQLite3 adapter (Future)
│   └── storage/                # External storage backends (Phase 2)
│       ├── s3/                 # AWS S3 adapter
│       └── gcs/                # Google Cloud Storage (Future)
│
├── internal/                   # Internal utilities (hash, serialize)
├── examples/                   # Example applications
└── docs/                       # Documentation
```

## 1.7 Glossary

| Term | Definition |
|------|------------|
| Aggregate | A cluster of domain objects treated as a single unit for data changes |
| Compaction | Archiving old events after snapshot creation to reduce storage |
| Event Sourcing | Pattern of storing state changes as a sequence of events |
| Hash Chain | Sequential linking of cryptographic hashes for tamper detection |
| Merkle Tree | Binary tree of hashes enabling efficient O(log n) verification |
| Projection | Read-optimized materialized view derived from events |
| Snapshot | Point-in-time capture of aggregate state for optimization |
