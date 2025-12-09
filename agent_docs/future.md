# 1. Future

## 1.1 Scope

| Feature | Priority | Description |
|---------|----------|-------------|
| MySQL Adapter | High | Support for MySQL 8.0+ |
| SQLite3 Adapter | Medium | Support for embedded/local databases |
| GCS Storage | Medium | Google Cloud Storage backend |
| Azure Blob Storage | Low | Azure storage backend |
| MinIO Support | Medium | Self-hosted S3-compatible storage |
| Java SDK | High | Java client library for enterprise adoption |

## 1.2 MySQL Adapter Specification

### 1.2.1 Dialect Differences

| Feature | PostgreSQL | MySQL |
|---------|------------|-------|
| Auto-increment | BIGSERIAL | BIGINT AUTO_INCREMENT |
| JSON type | JSONB | JSON |
| UUID type | UUID | CHAR(36) |
| Timestamp type | TIMESTAMPTZ | DATETIME(6) |
| Upsert syntax | ON CONFLICT | ON DUPLICATE KEY UPDATE |
| Placeholder | $1, $2, ... | ?, ?, ... |

### 1.2.2 Requirements

- MySQL 8.0+ (for native JSON support)
- InnoDB storage engine
- UTF8MB4 character set

## 1.3 SQLite3 Adapter Specification

### 1.3.1 Dialect Differences

| Feature | PostgreSQL | SQLite3 |
|---------|------------|---------|
| Auto-increment | BIGSERIAL | INTEGER PRIMARY KEY AUTOINCREMENT |
| JSON type | JSONB | TEXT (JSON stored as string) |
| UUID type | UUID | TEXT |
| Timestamp type | TIMESTAMPTZ | TEXT (ISO8601 format) |
| Placeholder | $1, $2, ... | ?, ?, ... |

### 1.3.2 Requirements

- WAL mode enabled for concurrency
- Single writer limitation (MaxOpenConns = 1 for writes)
- JSON operations handled in application layer

## 1.4 Additional Storage Backends

### 1.4.1 Google Cloud Storage

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Bucket | string | Yes | GCS bucket name |
| Project | string | Yes | GCP project ID |
| Prefix | string | No | Object key prefix |

### 1.4.2 Azure Blob Storage

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Container | string | Yes | Blob container name |
| AccountName | string | Yes | Storage account name |
| Prefix | string | No | Blob name prefix |

### 1.4.3 MinIO

Uses S3 adapter with custom endpoint configuration.

## 1.5 Java SDK Specification

### 1.5.1 Package Structure

```
com.yourorg.integridb
├── IntegriDBClient          # Main client interface
├── IntegriDBClientBuilder   # Builder for client configuration
├── Record                   # Record type
├── Event                    # Event type
├── MerkleProof              # Merkle proof type
├── Snapshot                 # Snapshot type
├── QueryBuilder             # Query builder interface
├── Transaction              # Transaction interface
└── config/
    ├── Config               # Client configuration
    ├── StorageConfig        # External storage configuration
    └── SnapshotPolicy       # Snapshot policy configuration
```

### 1.5.2 API Parity

Java SDK should provide equivalent functionality to Go SDK:
- All CRUD operations
- Query builder with same capabilities
- Transaction support
- Event sourcing operations
- Integrity verification
- Merkle tree operations (Phase 2 features)
- Snapshot operations (Phase 2 features)

### 1.5.3 Java-Specific Considerations

- Use Builder pattern for configuration
- Implement AutoCloseable for resource management
- Use Optional for nullable returns
- Support both sync and async operations
- Compatible with Spring Framework

## 1.6 Future Timeline (Tentative)

| Quarter | Focus |
|---------|-------|
| Q1 | MySQL adapter, MinIO support |
| Q2 | SQLite3 adapter, GCS storage |
| Q3 | Java SDK development |
| Q4 | Java SDK release, Azure Blob storage |
