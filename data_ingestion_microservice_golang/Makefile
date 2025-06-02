.PHONY: format lint check test run build clean install-deps

# Format the code
format:
	go fmt ./...

# Lint the code
lint:
	golangci-lint run

# Check if the code compiles
check:
	go build -o /tmp/data-ingestion-microservice .

# Run tests
test:
	go test -v ./...

# Run the application
run:
	go run main.go

# Build the application
build:
	go build -o bin/data-ingestion-microservice main.go

# Build for production (with optimizations)
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/data-ingestion-microservice main.go

# Install dependencies
install-deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Install Go dependencies and tools
setup:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks
all: format lint check test build 