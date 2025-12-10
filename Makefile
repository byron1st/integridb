.PHONY: mockgen test lint clean help

# Generate mocks for all interfaces
mockgen:
	@echo "Generating mocks..."
	@command -v mockgen >/dev/null 2>&1 || go install go.uber.org/mock/mockgen@latest
	@rm -rf internal/mocks && mkdir -p internal/mocks
	@go generate ./...
	@echo "Mocks generated successfully"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Run linters and static analysis
lint:
	@echo "Running linters..."
	@gofmt -d .
	@go vet ./...
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -w .

# Clean generated files
clean:
	@echo "Cleaning..."
	@rm -rf internal/mocks
	@rm -f coverage.out

# Show help
help:
	@echo "Available targets:"
	@echo "  mockgen        - Generate mocks for all interfaces"
	@echo "  test           - Run tests"
	@echo "  test-race      - Run tests with race detector"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linters and static analysis"
	@echo "  fmt            - Format code"
	@echo "  clean          - Clean generated files"
	@echo "  help           - Show this help message"
