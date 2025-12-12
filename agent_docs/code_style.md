# Code Style Guidelines

- Follow standard Go conventions (effective Go, Go Code Review Comments)
- Use meaningful variable names (avoid single-letter except for common idioms like `i`, `err`)
- Document exported functions and types with godoc comments
- Keep functions focused and testable
- Prefer explicit error handling over panic
- Prefer the Go standard library over 3rd party libraries

# Error Handling Patterns

## Deferred Cleanup with Error Returns

When using `defer` for cleanup operations that return errors (e.g., `Close()`, `Rollback()`), explicitly ignore the error to satisfy `errcheck` linter.

```go
defer func() { _ = rows.Close() }()
defer func() { _ = tx.Rollback() }()
```

# Linting and Formatting

```bash
gofmt -w .                       # Format code
go vet ./...                     # Static analysis
golangci-lint run                # Comprehensive linting
modernize -fix ./...             # Modernize Go code
```
