.PHONY: test test-verbose test-coverage build run clean

# Run all tests
test:
	@echo "Running backend tests..."
	CGO_ENABLED=1 go test ./... -v

# Run tests with verbose output
test-verbose:
	@echo "Running backend tests (verbose)..."
	CGO_ENABLED=1 go test ./... -v -count=1

# Run tests with coverage
test-coverage:
	@echo "Running backend tests with coverage..."
	CGO_ENABLED=1 go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the application
build:
	@echo "Building backend..."
	go build -o bin/api ./cmd/api

# Run the application
run:
	@echo "Running backend..."
	go run ./cmd/api

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ coverage.out coverage.html

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Pre-deployment checks
pre-deploy: test lint
	@echo "✓ All pre-deployment checks passed!"
