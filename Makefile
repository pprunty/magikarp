### Makefile for FastAPI Python service with Poetry and Docker ###

# Global Variables
SERVICE_NAME := magikarp-service
MODEL_NAME := magikarp
DOCKERFILE := Dockerfile
DOCKER_IMAGE := $(SERVICE_NAME)
DOCKER_TAG := latest
PORT := 80
OLLAMA_PORT := 11434

# Default target (runs when just 'make' is called)
.PHONY: help
help:
	@echo "Available commands for $(SERVICE_NAME):"
	@echo ""
	@echo "Poetry Commands:"
	@echo "  make install        Install project dependencies using Poetry"
	@echo "  make update         Update project dependencies"
	@echo "  make lock           Generate Poetry lock file"
	@echo ""
	@echo "API Commands:"
	@echo "  make server         Start FastAPI development server"
	@echo ""
	@echo "Ollama Commands:"
	@echo "  make serve          Start Ollama server with models directory"
	@echo "  make model          Create Ollama model from Modelfile"
	@echo "  make run            Run the Ollama model"
	@echo "  make delete         Delete the Ollama model"
	@echo "  make stop           Stop the Ollama server running in the background"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   Build Docker image"
	@echo "  make docker-run     Run Docker container"
	@echo "  make docker-compose Start with docker-compose"
	@echo ""
	@echo "Development Commands:"
	@echo "  make everything     Install dependencies, start Ollama server, create model, and show next steps"
	@echo ""

# Mark all targets that don't create files as PHONY
.PHONY: install update lock serve server model run delete docker-build docker-run docker-compose everything stop

# Poetry commands
install:
	poetry install

update:
	poetry update

lock:
	poetry lock

# API commands
server:
	poetry run fastapi dev app/main.py

# Ollama commands
serve:
	export OLLAMA_MODELS=$(shell pwd)/models  && ollama serve

model:
	@echo "Creating Ollama model $(MODEL_NAME)..."
	@if ollama list | grep -q "$(MODEL_NAME)"; then \
		echo "Model $(MODEL_NAME) already exists, removing first..."; \
		ollama rm $(MODEL_NAME) || { echo "\033[33mWARNING: Failed to remove existing model\033[0m"; }; \
	fi
	@export OLLAMA_MODELS=$(shell pwd)/models && ollama create $(MODEL_NAME) -f ./Modelfile || { \
		echo "\033[33mWARNING: Failed to create model $(MODEL_NAME)\033[0m"; \
		echo "Continuing with existing model if available..."; \
	}

run:
	export OLLAMA_MODELS=$(shell pwd)/models  && ollama run $(MODEL_NAME)

delete:
	ollama rm $(MODEL_NAME)

# Stop the Ollama server running in the background
stop:
	@echo "Stopping Ollama server..."
	@if pgrep -f "ollama serve" > /dev/null; then \
		pkill -f "ollama serve"; \
		echo "\033[32mOllama server stopped successfully\033[0m"; \
	else \
		echo "\033[33mNo Ollama server process found running\033[0m"; \
	fi

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f $(DOCKERFILE) .

docker-run:
	docker run -p $(PORT):$(PORT) $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose:
	docker-compose up --build

# Development command that runs everything needed for local development
everything:
	@echo "Checking prerequisites for Magikarp development environment..."

	@# Check if Python is installed
	@if ! command -v python3 &> /dev/null; then \
		echo "\033[31mERROR: Python 3 is not installed\033[0m"; \
		echo "Please install Python 3 from https://www.python.org/downloads/"; \
		echo ""; \
		exit 1; \
	else \
		echo "\033[32m✓ Python is installed\033[0m"; \
	fi

	@# Check if Poetry is installed
	@if ! command -v poetry &> /dev/null; then \
		echo "\033[31mERROR: Poetry is not installed\033[0m"; \
		echo "Please install Poetry using:"; \
		echo "curl -sSL https://install.python-poetry.org | python3 -"; \
		echo ""; \
		exit 1; \
	else \
		echo "\033[32m✓ Poetry is installed\033[0m"; \
	fi

	@# Check if Ollama is installed
	@if ! command -v ollama &> /dev/null; then \
		echo "\033[31mERROR: Ollama is not installed\033[0m"; \
		echo "Please install Ollama using:"; \
		echo "- macOS: brew install ollama"; \
		echo "- Other platforms: https://ollama.com/download"; \
		echo ""; \
		exit 1; \
	else \
		echo "\033[32m✓ Ollama is installed\033[0m"; \
	fi

	@echo ""
	@echo "\033[1mSetting up everything for local development...\033[0m"

	@echo ""
	@echo "\033[1mStep 1/3: Installing dependencies\033[0m"
	@make install

	@echo ""
	@echo "\033[1mStep 2/3: Starting Ollama server\033[0m"
	@echo "Starting Ollama server in the background on port $(OLLAMA_PORT)..."
	@export OLLAMA_MODELS=$(shell pwd)/models && ollama serve > /dev/null 2>&1 &
	@echo "Waiting for Ollama server to start..."
	@sleep 3

	@echo ""
	@echo "\033[1mStep 3/3: Creating Ollama model\033[0m"
	@echo "\033[33mNOTE: This may take 10-20 minutes and will use ~4.8GB of disk space\033[0m"
	@echo "Model will be stored at: $(shell pwd)/models"
	@if ! export OLLAMA_MODELS=$(shell pwd)/models && ollama create $(MODEL_NAME) -f ./Modelfile; then \
		echo "\033[31mERROR: Failed to create model. Stopping Ollama server...\033[0m"; \
		make stop; \
		exit 1; \
	fi

	@echo ""
	@echo "\033[32mModel created successfully!\033[0m"
	@echo ""
	@echo "\033[1mNext steps:\033[0m"
	@echo "1. Chat with the model via terminal:"
	@echo "   \033[36mmake run\033[0m"
	@echo ""
	@echo "2. Start the FastAPI server to interact via API:"
	@echo "   \033[36mmake server\033[0m"
	@echo "   (This will start the API server on http://127.0.0.1:8000)"
	@echo ""
	@echo "3. When finished, stop the Ollama server running in background:"
	@echo "   \033[36mmake stop\033[0m"
	@echo "   (Ollama is currently running on port $(OLLAMA_PORT))"
	@echo ""