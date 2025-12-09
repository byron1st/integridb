# Testing

## Requirements

When implementing features, ensure:

1. **Unit Tests**: Each component (parser, hasher, ledger) has isolated tests
2. **Integration Tests**: Test with real PostgreSQL instances (consider using testcontainers-go)
3. **Concurrency Tests**: Verify advisory lock behavior with parallel transactions
4. **Hash Consistency Tests**: Verify that identical queries produce identical hashes

## Writing tests

- Use the Go's standard testing library
- Prefer the "table" style (not required)
- If tests share some common setup, use `t.Run`s inside a `Test_Xxx` function.
- Add the `_test` suffix to a package name and test only exported (public) functions/methods

## Running

```bash
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -race ./...              # Race condition detection
go test -cover ./...             # Coverage report
```
