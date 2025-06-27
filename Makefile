
# Run tests with coverage
.PHONY: test
test: install
	PATH=$(HOME)/.local/bin:$(PATH) go test ./... -coverprofile=coverage.out

# Run tests with coverage, continue on failure
.PHONY: test-with-coverage
test-with-coverage:
	go test ./... -coverprofile=coverage.out || true

# Check coverage targets
.PHONY: coverage-check
coverage-check:
	cd scripts && go run coverage_check.go

# Run tests and check coverage
.PHONY: test-ci
test-ci: test coverage-check

# Makefile for uzi.go

# Variables
# Main source file
MAIN_FILE := uzi.go
BINARY := uzi

# Default target
all: build

# Build the Go binary
build:
	go build -o $(BINARY) .

# Run the Go program
run: build
	./$(BINARY)

# Install the binary to ~/.local/bin
install: build
	mkdir -p ~/.local/bin
	cp $(BINARY) ~/.local/bin/

# Clean up build artifacts
clean:
	rm -f $(BINARY)

.PHONY: all build run clean install