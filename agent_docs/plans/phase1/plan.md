# Phase 1 Implementation Plan

## Overview

**Goal**: Build a Go library that wraps `database/sql` to provide transparent event sourcing with hash chain integrity for PostgreSQL.

**Key Deliverables**:
- `database/sql` wrapper with familiar API
- SQL parser for mutation detection
- Event store with hash chain integrity
- State replay and integrity verification
- PostgreSQL support

---

## Implementation Stages

### Stage 1: Project Setup & Foundation

**Goal**: Establish project structure, core types, and database infrastructure.

**Tasks**:
- [Task 1.1: Initialize Go Module and Project Structure](./task-1.1-project-init.md)
- [Task 1.2: Define Core Types](./task-1.2-core-types.md)
- [Task 1.3: Implement PostgreSQL Schema Migration](./task-1.3-postgres-schema.md)
- [Task 1.4: Implement database/sql Wrapper Skeleton](./task-1.4-sql-wrapper.md)

**Acceptance Criteria**:
- [ ] Project compiles without errors
- [ ] Core types defined and documented
- [ ] Database schema created and migrated
- [ ] Basic SQL wrapper connects to PostgreSQL

---

### Stage 2: SQL Parser & Event Store

**Goal**: Parse SQL mutations and persist events with hash chain integrity.

**Tasks**:
- [Task 2.1: Implement SQL Parser for Mutation Detection](./task-2.1-sql-parser.md)
- [Task 2.2: Implement Hash Chain Logic](./task-2.2-hash-chain.md)
- [Task 2.3: Implement Event Store](./task-2.3-event-store.md)

**Acceptance Criteria**:
- [ ] SQL parser identifies INSERT/UPDATE/DELETE operations
- [ ] Hash chain links events cryptographically
- [ ] Events persist correctly with version numbering

---

### Stage 3: Event Capture Implementation

**Goal**: Transparently capture INSERT/UPDATE/DELETE operations as events.

**Tasks**:
- [Task 3.1: Implement Before/After State Capture](./task-3.1-state-capture.md)
- [Task 3.2: Implement INSERT Event Capture](./task-3.2-insert-capture.md)
- [Task 3.3: Implement UPDATE Event Capture](./task-3.3-update-capture.md)
- [Task 3.4: Implement DELETE Event Capture](./task-3.4-delete-capture.md)

**Acceptance Criteria**:
- [ ] INSERT operations create events with after state
- [ ] UPDATE operations create events with before/after states
- [ ] DELETE operations create events with before state
- [ ] Multi-row operations handled correctly

---

### Stage 4: Transaction Support & Metadata

**Goal**: Support transactions and user-defined metadata injection.

**Tasks**:
- [Task 4.1: Implement Transaction Wrapper](./task-4.1-transactions.md)
- [Task 4.2: Implement Metadata Injection](./task-4.2-metadata.md)
- [Task 4.3: Implement Table Management](./task-4.3-table-management.md)

**Acceptance Criteria**:
- [ ] Events persist atomically within transactions
- [ ] Metadata injected via context
- [ ] Tables can be tracked/untracked dynamically

---

### Stage 5: Verification & Replay

**Goal**: Verify data integrity and reconstruct historical states.

**Tasks**:
- [Task 5.1: Implement Integrity Verification](./task-5.1-verification.md)
- [Task 5.2: Implement State Replay by Version](./task-5.2-replay-version.md)
- [Task 5.3: Implement State Replay by Time](./task-5.3-replay-time.md)
- [Task 5.4: Implement Event History Retrieval](./task-5.4-history.md)

**Acceptance Criteria**:
- [ ] Hash chain integrity verified correctly
- [ ] State reconstructed at any version
- [ ] State reconstructed at any timestamp
- [ ] Event history accessible via API

---

### Stage 6: Testing, Documentation & Examples

**Goal**: Ensure quality, provide documentation, and create examples.

**Tasks**:
- [Task 6.1: Comprehensive Unit Tests](./task-6.1-unit-tests.md)
- [Task 6.2: Integration Tests](./task-6.2-integration-tests.md)
- [Task 6.3: Documentation](./task-6.3-documentation.md)
- [Task 6.4: Example Applications](./task-6.4-examples.md)

**Acceptance Criteria**:
- [ ] >80% code coverage
- [ ] All integration tests pass
- [ ] Complete user documentation
- [ ] Working example applications

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
