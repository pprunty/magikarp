.PHONY: install build run test clean lint download-model ollama

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

all: install build

install:
	$(GOMOD) tidy

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)

run: build
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
	@echo "Available commands:"
	@echo "  make install    - Install dependencies"
	@echo "  make build     - Build the binary"
	@echo "  make run       - Build and run the binary"
	@echo "  make test      - Run tests"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make lint      - Run linter"
	@echo "  make tools     - Install development tools"
	@echo "  make download-model - Download and link Ollama model" 