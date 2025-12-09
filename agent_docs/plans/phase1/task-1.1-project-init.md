# Task 1.1: Initialize Go Module and Project Structure

## Description
Set up the Go module and create the directory structure for IntegriDB.

## Actions

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

## Acceptance Criteria

- [ ] `go mod tidy` runs without errors
- [ ] Directory structure exists
- [ ] Empty placeholder files compile: `go build ./...`
