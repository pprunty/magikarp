.PHONY: all clean build test mgk

# Go parameters
GO := go
BINARY_NAME := magikarp
CLI_BINARY_NAME := mgk

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.DEFAULT_GOAL := run

all: clean build

build: build-lib build-cli

build-lib:
	$(GO) build -o $(BINARY_NAME) main.go

build-cli:
	$(GO) build -o $(CLI_BINARY_NAME) ./cmd/mgk

clean:
	rm -f $(BINARY_NAME) $(CLI_BINARY_NAME)

test:
	$(GO) test ./...

mgk: build-cli
	./$(CLI_BINARY_NAME)

# Build and install the CLI to the Go binary directory
install-cli: build-cli
	cp $(CLI_BINARY_NAME) $(GOPATH)/bin/

# Helper targets for tool and agent creation
create-tool:
	@test -n "$(NAME)" || (echo "NAME is not set. Usage: make create-tool NAME=mytool"; exit 1)
	./$(CLI_BINARY_NAME) create tool $(NAME)

create-agent:
	@test -n "$(NAME)" || (echo "NAME is not set. Usage: make create-agent NAME=myagent"; exit 1)
	./$(CLI_BINARY_NAME) create agent $(NAME)

# Run an agent
run-agent:
	@test -n "$(NAME)" || (echo "NAME is not set. Usage: make run-agent NAME=myagent"; exit 1)
	./$(CLI_BINARY_NAME) run $(NAME)

# Development tools
tools:
	$(GO) get github.com/golangci/golangci-lint/cmd/golangci-lint@latest

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