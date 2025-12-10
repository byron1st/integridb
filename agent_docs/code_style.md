# Code Style Guidelines

- Follow standard Go conventions (effective Go, Go Code Review Comments)
- Use meaningful variable names (avoid single-letter except for common idioms like `i`, `err`)
- Document exported functions and types with godoc comments
- Keep functions focused and testable
- Prefer explicit error handling over panic
- Prefer the Go standard library over 3rd party libraries

# Error Handling Patterns

## Deferred Cleanup with Error Returns

When using `defer` for cleanup operations that return errors (e.g., `Close()`, `Rollback()`), explicitly handle or ignore the error to satisfy `errcheck` linter.

**Pattern: Explicitly ignore errors in defer**

```go
// GOOD: Explicitly ignore error with anonymous function
defer func() { _ = rows.Close() }()
defer func() { _ = tx.Rollback() }()

// BAD: Unchecked error (fails errcheck)
defer rows.Close()
defer tx.Rollback()
```

**Rationale:**
- `tx.Rollback()` in defer is safe to ignore because:
  - If `tx.Commit()` succeeds, `Rollback()` is a no-op
  - If we return early with an error, the rollback succeeds and we're already handling the error
- `rows.Close()` in defer is safe to ignore because:
  - We check `rows.Err()` for iteration errors
  - Close errors are typically not actionable in defer context
  - The connection pool handles cleanup

**When to check errors:**
- For critical cleanup where failure needs logging or handling
- Use a named return value or closure that can handle the error properly

```go
// Example: Check and log defer errors when critical
defer func() {
    if err := rows.Close(); err != nil {
        log.Printf("failed to close rows: %v", err)
    }
}()
```

# Linting and Formatting

```bash
gofmt -w .                       # Format code
go vet ./...                     # Static analysis
golangci-lint run                # Comprehensive linting
modernize -fix ./...             # Modernize Go code
```

# Verification & Validation

When verifying changes:

## Always

1. Verify code formatting: `gofmt -d .` (should have no output)
2. Modernize code: `modernize -fix ./...`
3. Run a linter: `golangci-lint run`
4. Run some tests related to changed code: `go test`

## Before committing

1. Run static analysis: `go vet ./...`
2. Run full test suite: `go test ./...`
3. Check for race conditions: `go test -race ./...`
4. Test with actual PostgreSQL instance
5. Verify hash chain integrity after test transactions
