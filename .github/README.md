# GitHub Actions

## Code Quality Workflow

The `lint.yml` workflow runs on every pull request targeting the `main` branch and performs the following checks:

### Lint Job
- **go vet**: Examines Go source code and reports suspicious constructs
- **golangci-lint** (v2.x): Runs multiple linters with default configuration
- **staticcheck**: Advanced static analysis for Go code
- **modernize**: Suggests modern Go idioms and improvements

### Test Job
- Runs all tests with race detector enabled
- Generates code coverage report

## Branch Protection

To require these checks to pass before merging, configure branch protection rules on GitHub:

1. Go to your repository **Settings** → **Branches**
2. Add a branch protection rule for `main`
3. Enable:
   - ✅ **Require status checks to pass before merging**
   - Select required checks:
     - `Lint and Static Analysis`
     - `Run Tests`
   - ✅ **Require branches to be up to date before merging** (optional but recommended)

## Local Development

Run the same checks locally before pushing:

```bash
# Run go vet
go vet ./...

# Run tests with race detector
go test -race ./...

# Run golangci-lint (requires installation)
golangci-lint run

# Run staticcheck (requires installation)
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# Run modernize (requires installation)
go install github.com/Crocmagnon/go-modernize/cmd/modernize@latest
modernize ./...
```

## Installing Tools

### golangci-lint

```bash
# macOS/Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or with Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### staticcheck

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### modernize

```bash
go install github.com/Crocmagnon/go-modernize/cmd/modernize@latest
```
