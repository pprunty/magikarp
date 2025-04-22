.PHONY: install build run test clean lint download-model ollama help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=magikarp

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.DEFAULT_GOAL := run

all: install build

install:
	$(GOMOD) tidy

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)

run: install build
	@echo "\033[1;32mStarting Magikarp...\033[0m"
	./$(BINARY_NAME)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

lint:
	golangci-lint run

# Development tools
tools:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Download model
download-model:
	./scripts/download_model.sh

ollama:
	export OLLAMA_MODELS=$(shell pwd)/models  && ollama serve

# Help
help:
	@echo "Magikarp\n\nA flexible plugin-based agentic LLM framework"
	@echo "\n\033[1;33mUsage:\033[0m"
	@echo "  make [target]"
	@echo "\n\033[1;33mTargets:\033[0m"
	@echo "  \033[1;32mhelp\033[0m            Display this help message"
	@echo "  \033[1;32minstall\033[0m         Install project dependencies"
	@echo "  \033[1;32mbuild\033[0m           Build the binary"
	@echo "  \033[1;32mrun\033[0m             Install, build and run the binary"
	@echo "  \033[1;32mtest\033[0m            Run all tests"
	@echo "  \033[1;32mclean\033[0m           Remove build artifacts"
	@echo "  \033[1;32mlint\033[0m            Run the linter"
	@echo "  \033[1;32mtools\033[0m           Install development tools"
	@echo "  \033[1;32mdownload-model\033[0m  Download and link Ollama model"
	@echo "  \033[1;32mollama\033[0m          Start Ollama server"
	@echo "\n\033[1;33mExamples:\033[0m"
	@echo "  make install    # Install dependencies"
	@echo "  make run       # Install, build and run the application"
	@echo "  make test      # Run tests"
	@echo "\n\033[1;33mEnvironment:\033[0m"
	@echo "  OLLAMA_MODELS  Path to Ollama models directory" 