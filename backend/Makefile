.PHONY: build run test clean docker docker-run lint fmt help test-integration test-e2e test-all test-bench

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
BINARY_NAME=server
BINARY_PATH=bin/$(BINARY_NAME)

# Build the server
build:
	@echo "Building server..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) ./cmd/server

# Run the server locally
run:
	@echo "Starting server..."
	$(GOCMD) run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Run end-to-end tests (alias for integration tests)
test-e2e:
	@echo "Running end-to-end tests..."
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Run all tests including integration tests
test-all: test test-integration
	@echo "All tests completed."

# Run integration benchmarks
test-bench:
	@echo "Running integration benchmarks..."
	$(GOTEST) -v -tags=integration -bench=. -benchmem ./tests/integration/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet instead"; \
		$(GOCMD) vet ./...; \
	fi

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t elite-agent-collective .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 elite-agent-collective

# Run with docker-compose
compose-up:
	@echo "Starting with docker-compose..."
	docker-compose up -d

# Stop docker-compose
compose-down:
	@echo "Stopping docker-compose..."
	docker-compose down

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build the server binary"
	@echo "  run              - Run the server locally"
	@echo "  test             - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e         - Run end-to-end tests (alias for integration)"
	@echo "  test-all         - Run all tests (unit + integration)"
	@echo "  test-bench       - Run integration benchmarks"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  clean            - Remove build artifacts"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint code"
	@echo "  docker           - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  compose-up       - Start with docker-compose"
	@echo "  compose-down     - Stop docker-compose"
	@echo "  help             - Show this help"
