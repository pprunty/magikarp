### Makefile for FastAPI Python service with Poetry and Docker ###

# Global Variables
SERVICE_NAME := magikarp-service
MODEL_NAME := magikarp
DOCKERFILE := Dockerfile
DOCKER_IMAGE := $(SERVICE_NAME)
DOCKER_TAG := latest
PORT := 80

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
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   Build Docker image"
	@echo "  make docker-run     Run Docker container"
	@echo "  make docker-compose Start with docker-compose"
	@echo ""

# Mark all targets that don't create files as PHONY
.PHONY: install update lock serve server model run delete docker-build docker-run docker-compose

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
	export OLLAMA_MODELS=$(shell pwd)/models && ollama serve

model:
	export OLLAMA_MODELS=$(shell pwd)/models && ollama create $(MODEL_NAME) -f ./Modelfile

run:
	ollama run $(MODEL_NAME)

delete:
	ollama rm $(MODEL_NAME)

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f $(DOCKERFILE) .

docker-run:
	docker run -p $(PORT):$(PORT) $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose:
	docker-compose up --build