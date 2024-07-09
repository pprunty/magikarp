### Makefile for FastAPI Python service with Poetry and Docker ###

# Global Variables
SERVICE_NAME := magikarp-service
MODEL_NAME := magikarp
DOCKERFILE := Dockerfile
DOCKER_IMAGE := $(SERVICE_NAME)
DOCKER_TAG := latest
PORT := 80

.PHONY: install update lock docker-build docker-run

# Poetry commands
install:
	poetry install

update:
	poetry update

lock:
	poetry lock

# FastAPI commands
run:
	poetry run fastapi dev magikarp/main.py

## ollama commands
model:
	ollama create $(MODEL_NAME) -f ./Modelfile

model-run:
	ollama run $(MODEL_NAME)

model-delete:
	ollama rm $(MODEL_NAME)

# Docker commands
docker-build:
	docker pull ollama/ollama && docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .  --progress=plain

docker-run:
	docker run -d  $(DOCKER_IMAGE):$(DOCKER_TAG)

