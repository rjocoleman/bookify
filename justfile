# Bookify Tasks
set dotenv-load

# Show available commands
default:
    @just --list

# Build the server binary
build:
    @echo "Building bookify server..."
    go build -o server cmd/server/main.go

# Install dependencies and dev tools
install:
    @echo "Installing dependencies..."
    go mod download
    go mod tidy
    go install github.com/a-h/templ/cmd/templ@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate templ templates
generate:
    @echo "Generating templates..."
    templ generate

# Run all tests
test:
    @echo "Running tests..."
    go test ./...

# Run tests with verbose output
test-verbose:
    @echo "Running tests with verbose output..."
    go test -v ./...

# Run tests with coverage
coverage:
    @echo "Running tests with coverage..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# Show coverage in terminal
coverage-show:
    @echo "Running tests with coverage..."
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run tests with race detection
test-race:
    @echo "Running tests with race detection..."
    go test -race ./...

# Run benchmarks
bench:
    @echo "Running benchmarks..."
    go test -bench=. ./...

# Format code
fmt:
    @echo "Formatting code..."
    go fmt ./...

# Run linter
lint:
    @echo "Running linter..."
    golangci-lint run

# Run all quality checks
check: fmt lint test
    @echo "All checks passed ✓"

# Clean build artifacts
clean:
    @echo "Cleaning..."
    rm -f server coverage.out coverage.html
    rm -rf temp/

# Development server (generate, build, run)
dev: generate build
    @echo "Starting development server..."
    ./server

# Production server (build and run)
run: build
    @echo "Starting server..."
    ./server

# Quick development cycle
quick: generate test build
    @echo "Quick build complete ✓"

# Full CI pipeline
ci: install generate fmt lint test-race coverage
    @echo "CI pipeline completed successfully ✓"

# Reset database
db-reset:
    @echo "Resetting database..."
    rm -f kepub.db
    @echo "Database will be recreated on next run"

# Docker build
docker-build:
    @echo "Building Docker image..."
    docker build -t bookify .

# Docker run
docker-run:
    @echo "Running Docker container..."
    docker run -p 8080:8080 bookify

# Test specific package (usage: just test-pkg internal/db)
test-pkg pkg:
    @echo "Testing {{pkg}}..."
    go test -v ./{{pkg}}/...
