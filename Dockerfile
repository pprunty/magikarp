# Use the ollama/ollama base image
FROM ollama/ollama:latest

# Working directory
WORKDIR /app

COPY . /app

# Install dependencies
CMD ["ollama", "serve"]