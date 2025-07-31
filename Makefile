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

.PHONY: help build run install clean fmt release whisper-models speech-deps whisper-cli whisper-model build-speech run-speech install-speech

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
	@echo "  $(YELLOW)whisper-models$(RESET) Download Whisper models for speech recognition"
	@echo "  $(YELLOW)whisper-cli$(RESET)  Install Whisper CLI"
	@echo "  $(YELLOW)whisper-model$(RESET) Download base Whisper model"
	@echo "  $(YELLOW)speech-deps$(RESET)   Setup speech recognition dependencies"
	@echo "  $(YELLOW)build-speech$(RESET) Build binary with speech tag"
	@echo "  $(YELLOW)run-speech$(RESET)   Run with speech tag"
	@echo "  $(YELLOW)install-speech$(RESET) Install deps & build with speech tag"
	@echo ""
	@echo "$(BLUE)Usage: make [command]$(RESET)"

## Build the binary
build:
	@$(MAKE) _build TAGS=""

## Build with speech tag
build-speech:
	@$(MAKE) _build TAGS="speech"

_build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(RESET)"
	mkdir -p bin
	go build $(if $(TAGS),-tags $(TAGS),) -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)" -o bin/$(BINARY_NAME) .
	@echo "$(GREEN)✓ Build complete: bin/$(BINARY_NAME)$(RESET)"

## Run the application
run:
	@$(MAKE) _run TAGS=""

## Run with speech tag
run-speech:
	@$(MAKE) _run TAGS="speech"

_run:
	@echo "$(GREEN)Running $(BINARY_NAME)...$(RESET)"
	go run $(if $(TAGS),-tags $(TAGS),) .

## Install dependencies and build
install:
	@$(MAKE) _install TAGS=""

## Install deps & build with speech tag
install-speech:
	@$(MAKE) _install TAGS="speech"

_install:
	@echo "$(GREEN)Downloading dependencies...$(RESET)"
	go mod download
	@$(MAKE) _build TAGS="$(TAGS)"
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

## Download Whisper models for speech recognition
whisper-models:
	@echo "$(GREEN)Downloading Whisper models...$(RESET)"
	mkdir -p $(HOME)/.magikarp/models
	@if [ ! -f $(HOME)/.magikarp/models/ggml-small.en.bin ]; then \
		echo "$(BLUE)Downloading ggml-small.en.bin model (~461MB)...$(RESET)"; \
		curl -L -o $(HOME)/.magikarp/models/ggml-small.en.bin \
			https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin; \
		echo "$(GREEN)✓ Model downloaded to $(HOME)/.magikarp/models/ggml-small.en.bin$(RESET)"; \
	else \
		echo "$(YELLOW)Model already exists at $(HOME)/.magikarp/models/ggml-small.en.bin$(RESET)"; \
	fi

## Install Whisper CLI
whisper-cli:
	@echo "$(GREEN)Installing Whisper CLI...$(RESET)"
	go install github.com/mutablelogic/go-whisper/cmd/whisper@latest
	@echo "$(GREEN)✓ Whisper CLI installed$(RESET)"

## Download Whisper base model
whisper-model:
	@echo "$(GREEN)Downloading Whisper base model (ggml-base.en.bin)...$(RESET)"
	whisper download ggml-base.en.bin
	@echo "$(GREEN)✓ Model downloaded to $$HOME/.cache/whisper$(RESET)"

## Setup speech recognition dependencies  
speech-deps:
	@echo "$(GREEN)Setting up speech recognition dependencies...$(RESET)"
	@echo "$(BLUE)Ensuring PortAudio is available...$(RESET)"
	@if command -v pkg-config >/dev/null 2>&1 && pkg-config --exists portaudio-2.0; then \
		echo "$(GREEN)✓ PortAudio found$(RESET)"; \
	else \
		echo "$(YELLOW)⚠ PortAudio not found. Please install:$(RESET)"; \
		echo "  $(CYAN)macOS:$(RESET) brew install portaudio"; \
		echo "  $(CYAN)Ubuntu/Debian:$(RESET) sudo apt-get install portaudio19-dev"; \
		echo "  $(CYAN)CentOS/RHEL:$(RESET) sudo yum install portaudio-devel"; \
	fi
	@echo "$(GREEN)✓ Speech dependencies check complete$(RESET)"

# Default target
.DEFAULT_GOAL := help