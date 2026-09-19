# Born ML Framework - Makefile

.PHONY: help build test bench lint clean install run-mnist

# Default target
help:
	@echo "Born ML Framework - Make targets:"
	@echo ""
	@echo "  build        - Build all binaries"
	@echo "  test         - Run all tests"
	@echo "  test-race    - Run tests with race detector"
	@echo "  bench        - Run benchmarks"
	@echo "  lint         - Run linter (golangci-lint)"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install CLI tools"
	@echo "  run-mnist    - Run MNIST example"
	@echo "  coverage     - Generate coverage report"
	@echo ""

# Build all binaries
build:
	@echo "Building Born CLI..."
	@go build -o bin/born ./cmd/born
	@echo "Building Born Benchmark..."
	@go build -o bin/born-bench ./cmd/born-bench
	@echo "Building Born Convert..."
	@go build -o bin/born-convert ./cmd/born-convert
	@echo "Build complete!"

# Run all tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./benchmarks/...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache
	@echo "Clean complete!"

# Install CLI tools
install:
	@echo "Installing Born CLI..."
	@go install ./cmd/born
	@echo "Install complete!"

# Run MNIST example
run-mnist:
	@echo "Running MNIST example..."
	@go run ./examples/mnist

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# All checks (test + lint + vet)
check: test lint vet
	@echo "All checks passed!"

# CI target (used by GitHub Actions)
ci: test-race lint vet
	@echo "CI checks passed!"
