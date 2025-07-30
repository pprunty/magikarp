# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

# Project variables
BINARY_NAME := magikarp
VERSION := $(shell grep '^version:' config.yaml | sed 's/version: *"*\([^"]*\)"*/\1/' 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

.PHONY: help build run install clean fmt release

## Show this help message
help:
	@echo "$(CYAN)Magikarp - AI Coding Assistant$(RESET)"
	@echo "$(CYAN)==============================$(RESET)"
	@echo ""
	@echo "$(GREEN)Available commands:$(RESET)"
	@echo ""
	@echo "  $(YELLOW)build$(RESET)        Build the binary"
	@echo "  $(YELLOW)run$(RESET)          Run the application"
	@echo "  $(YELLOW)install$(RESET)      Install dependencies and build"
	@echo "  $(YELLOW)clean$(RESET)        Clean build artifacts"
	@echo "  $(YELLOW)fmt$(RESET)          Format code"
	@echo "  $(YELLOW)release$(RESET)      Build release version with GoReleaser"
	@echo ""
	@echo "$(BLUE)Usage: make [command]$(RESET)"

## Build the binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(RESET)"
	mkdir -p bin
	go build -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)" -o bin/$(BINARY_NAME) .
	@echo "$(GREEN)✓ Build complete: bin/$(BINARY_NAME)$(RESET)"

## Run the application
run:
	@echo "$(GREEN)Running $(BINARY_NAME)...$(RESET)"
	go run .

## Install dependencies and build
install:
	@echo "$(GREEN)Downloading dependencies...$(RESET)"
	go mod download
	@echo "$(GREEN)Building $(BINARY_NAME)...$(RESET)"
	mkdir -p bin
	go build -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)" -o bin/$(BINARY_NAME) .
	@echo "$(GREEN)✓ Installation complete: bin/$(BINARY_NAME)$(RESET)"

## Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	rm -rf dist/
	go clean
	@echo "$(GREEN)✓ Clean complete$(RESET)"

## Format code
fmt:
	@echo "$(GREEN)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

## Build release version with GoReleaser
release:
	@echo "$(GREEN)Building release with GoReleaser...$(RESET)"
	goreleaser release --snapshot --clean
	@echo "$(GREEN)✓ Release build complete$(RESET)"

# Default target
.DEFAULT_GOAL := help