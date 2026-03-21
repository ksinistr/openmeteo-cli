# Makefile for openmeteo-cli
# Build, test, and format the Go CLI

.PHONY: all build test fmt lint clean

all: build test

# Build the binary
build:
	go build -o bin/openmeteo-cli ./cmd/openmeteo-cli

# Run all tests
test:
	go test -v ./...

# Format Go source files
fmt:
	go fmt ./...

# Run linting
lint:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf gocache/
	rm -f go.sum
