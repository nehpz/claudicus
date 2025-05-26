# Makefile for uzi.go

# Variables
GO_FILES := $(wildcard *.go)
BINARY := uzi

# Default target
all: build

# Build the Go binary
build:
	go build -o $(BINARY) $(GO_FILES)

# Run the Go program
run: build
	./$(BINARY)

# Clean up build artifacts
clean:
	rm -f $(BINARY)

.PHONY: all build run clean