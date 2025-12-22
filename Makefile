.PHONY: build test test-coverage clean deps help

# Build the application
build:
	@echo "Building reconciliation service..."
	@go build -o bin/recon ./cmd/recon
	@echo "Build complete! Binary: bin/recon"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@go mod vendor

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  help          - Show this help message"
