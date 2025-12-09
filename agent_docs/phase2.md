# 1. Phase 2

## 1.1 Scope

**New Features:**
- Merkle tree for efficient verification
- Snapshot feature for storage optimization
- S3 external storage for snapshots
- Event compaction

**Migration:**
- Backward compatible with MVP
- Migration tool to add Merkle tree to existing events

## 1.2 Schema Additions

### 1.2.1 Events Table Additions

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| leaf_hash | VARCHAR(64) | Yes | Merkle tree leaf hash |
| tree_position | BIGINT | Yes | Position in Merkle tree |
| archived | BOOLEAN | No | Whether event is archived (default: false) |
| archived_at | TIMESTAMPTZ | Yes | When event was archived |

### 1.2.2 Merkle Nodes Table

Stores internal nodes of Merkle trees.

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| id | BIGSERIAL | No | Auto-incrementing primary key |
| table_name | VARCHAR(255) | No | Table name |
| row_id | VARCHAR(255) | Yes | Row ID (NULL for table-wide tree) |
| tree_level | INT | No | Level in tree (0 = leaves) |
| node_index | BIGINT | No | Position at this level |
| hash | VARCHAR(64) | No | Node hash value |
| left_child | VARCHAR(64) | Yes | Left child hash |
| right_child | VARCHAR(64) | Yes | Right child hash |
| updated_at | TIMESTAMPTZ | No | Last update timestamp |

**Constraints:**
- Unique constraint on (table_name, row_id, tree_level, node_index)

### 1.2.3 Merkle Roots Table

History of Merkle root hashes for audit trail.

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| id | BIGSERIAL | No | Auto-incrementing primary key |
| table_name | VARCHAR(255) | No | Table name |
| row_id | VARCHAR(255) | Yes | Row ID |
| root_hash | VARCHAR(64) | No | Root hash value |
| event_count | BIGINT | No | Number of events at this root |
| created_at | TIMESTAMPTZ | No | Creation timestamp |

### 1.2.4 Snapshots Table

| Column | Type | Nullable | Description |
|--------|------|----------|-------------|
| id | BIGSERIAL | No | Auto-incrementing primary key |
| table_name | VARCHAR(255) | No | Table name |
| row_id | VARCHAR(255) | Yes | Row ID (NULL for full table snapshot) |
| snapshot_version | BIGINT | No | Sequential snapshot version |
| last_event_id | BIGINT | No | Last event included in snapshot |
| last_event_version | BIGINT | No | Version of last event |
| state | JSONB | Yes | Inline state (if not external) |
| storage_type | VARCHAR(20) | No | "database" or "s3" |
| storage_ref | VARCHAR(512) | Yes | External storage URI |
| checksum | VARCHAR(64) | No | SHA-256 of state |
| merkle_root | VARCHAR(64) | Yes | Merkle root at snapshot time |
| event_count | BIGINT | No | Total events up to snapshot |
| size_bytes | BIGINT | Yes | Size of snapshot data |
| created_at | TIMESTAMPTZ | No | Creation timestamp |

**Constraints:**
- Unique constraint on (table_name, row_id, snapshot_version)

## 1.3 Merkle Tree Specification

### 1.3.1 Tree Structure

- Binary tree where leaves are event hashes
- Internal nodes: `hash = SHA256(left_child_hash + right_child_hash)`
- Tree is padded to next power of 2 with empty hashes
- Root hash represents integrity of entire event set

### 1.3.2 Leaf Hash Calculation

```
leaf_hash = SHA256(aggregate_type + aggregate_id + event_type + version + JSON(payload) + JSON(metadata))
```

### 1.3.3 Inclusion Proof

A proof that a specific event exists in the tree contains:
- Leaf hash of the event
- Leaf index (position in tree)
- Proof path (sibling hashes from leaf to root)
- Directions (whether each sibling is left or right)
- Root hash

Verification: Starting from leaf, hash with each sibling in order. Result must equal root hash.

### 1.3.4 Benefits Over Hash Chain

| Aspect | Hash Chain | Merkle Tree |
|--------|------------|-------------|
| Verify single event | O(n) | O(log n) |
| Proof size | O(n) | O(log n) |
| Third-party verification | Requires all events | Requires only proof path |
| Detect tampering location | Sequential scan | Binary search |

## 1.4 Snapshot Specification

### 1.4.1 Snapshot Policy Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Enabled | bool | false | Enable automatic snapshots |
| EventThreshold | int | 100 | Create snapshot after N events |
| TimeInterval | duration | 24h | Create snapshot after duration |
| RetainEvents | int | 10 | Events to keep after compaction |
| AutoCompact | bool | false | Automatically compact after snapshot |

### 1.4.2 Snapshot Creation Process

1. Load latest snapshot for aggregate (if exists)
2. Load all events since last snapshot
1. Apply events to build current state
4. Calculate state checksum
5. Get current Merkle root
6. Store snapshot (database or S3)
7. Optionally compact old events

### 1.4.3 Snapshot Storage Options

**Database Storage:**
- State stored in JSONB column
- Suitable for small to medium rows
- No additional infrastructure required

**S3 Storage:**
- State serialized to JSON, compressed with gzip
- Stored at: `{prefix}/{table_name}/{row_id}/v{version}.json.gz`
- URI stored in database for reference
- Suitable for large rows or high-volume systems

### 1.4.4 Event Compaction

After snapshot creation:
1. Identify events older than snapshot's last_event_id
2. Keep most recent N events (per RetainEvents setting)
1. Mark older events as archived
4. Optionally move to archive table or delete

## 1.5 S3 Storage Specification

### 1.5.1 Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Bucket | string | Yes | S3 bucket name |
| Region | string | Yes | AWS region |
| Prefix | string | No | Key prefix for all objects |
| Endpoint | string | No | Custom endpoint (for MinIO) |

### 1.5.2 Object Format

- Content-Type: application/gzip
- Content-Encoding: gzip
- Metadata headers: checksum, merkle-root, event-count

### 1.5.3 Operations

- Upload: Serialize → Compress → Upload → Return URI
- Download: Fetch → Decompress → Deserialize → Return state
- Delete: Remove object from bucket
- List: List objects with prefix

## 1.6 Additional API

### 1.6.1 Merkle Tree Operations

- `GetMerkleRoot(table, rowID) → string`: Get current root hash
- `GenerateProof(table, rowID, eventVersion) → MerkleProof`: Generate inclusion proof
- `VerifyWithProof(event, proof) → bool`: Verify event with proof
- `ExportProof(table, rowID, eventVersion) → ExportableProof`: Export proof for third-party

### 1.6.2 Snapshot Operations

- `CreateSnapshot(table, rowID) → Snapshot`: Create snapshot manually
- `GetLatestSnapshot(table, rowID) → Snapshot`: Get most recent snapshot
- `ReplayFromSnapshot(table, rowID, targetVersion) → map`: Replay from nearest snapshot

### 1.6.3 Storage Management

- `CompactEvents(table, rowID) → CompactionReport`: Compact events for a row
- `GetStorageStats(table) → StorageStats`: Get storage statistics

### 1.6.4 New Types

**MerkleProof:**
| Field | Type | Description |
|-------|------|-------------|
| LeafHash | string | Hash of the event |
| LeafIndex | int64 | Position in tree |
| ProofPath | []string | Sibling hashes |
| Directions | []bool | Sibling positions (true = right) |
| RootHash | string | Expected root hash |
| TreeSize | int64 | Number of leaves in tree |

**Snapshot:**
| Field | Type | Description |
|-------|------|-------------|
| ID | int64 | Snapshot ID |
| TableName | string | Table name |
| RowID | string | Row ID |
| SnapshotVersion | int64 | Sequential version |
| LastEventID | int64 | Last event in snapshot |
| LastEventVersion | int64 | Version of last event |
| State | map | Materialized state |
| StorageType | string | "database" or "s3" |
| StorageRef | string | External storage URI |
| Checksum | string | State hash |
| MerkleRoot | string | Root at snapshot time |
| EventCount | int64 | Total events |
| SizeBytes | int64 | Compressed size |
| CreatedAt | time | Creation timestamp |

**StorageStats:**
| Field | Type | Description |
|-------|------|-------------|
| TableName | string | Table name |
| TotalEvents | int64 | All events ever created |
| ActiveEvents | int64 | Non-archived events |
| ArchivedEvents | int64 | Archived events |
| SnapshotCount | int64 | Number of snapshots |
| TotalSizeBytes | int64 | Total storage used |
| ExternalSizeBytes | int64 | Bytes in external storage |

## 1.7 Migration from MVP

### 1.7.1 Migration Steps

1. Add new columns to events table (leaf_hash, tree_position, archived, archived_at)
2. Create new tables (merkle_nodes, merkle_roots, snapshots)
3. Backfill leaf_hash and tree_position for existing events
4. Build Merkle trees for all existing aggregates
5. Record initial Merkle roots

### 1.7.2 Backward Compatibility

- Existing events remain valid
- Hash chain continues to work alongside Merkle tree
- No changes required to existing application code
- New features available through additional API methods

## 1.8 Phase 2 Deliverables

- [ ] Merkle tree implementation with inclusion proofs
- [ ] Snapshot manager with configurable policies
- [ ] S3 storage backend
- [ ] Event compaction and archival
- [ ] Exportable proofs for third-party verification
- [ ] Migration tool from MVP
- [ ] Performance benchmarks
- [ ] Updated documentation

## 1.9 Phase 2 Timeline

| Week | Focus |
|------|-------|
| 1-2 | Merkle tree implementation and testing |
| 3 | Integrate Merkle tree with event store |
| 4 | Snapshot manager implementation |
| 5 | S3 storage backend |
| 6 | Event compaction and archival |
| 7 | Migration tools from MVP |
| 8 | Testing, documentation |
