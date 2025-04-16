#!/bin/bash

# Check if Ollama server is running
if ! curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "Error: Ollama server is not running"
    echo "Please start the Ollama server in a separate terminal with:"
    echo "make ollama"
    exit 1
fi

# Create models directory if it doesn't exist
mkdir -p models

# Pull the model using Ollama
echo "Pulling llama3.2 model..."
if ! ollama pull llama3.2; then
    echo "Error: Failed to pull llama3.2 model"
    exit 1
fi

# Get the model path from Ollama
MODEL_PATH=$(ollama show llama3.2 | grep -i "path:" | awk '{print $2}')

if [ -z "$MODEL_PATH" ]; then
    echo "Error: Could not determine model path"
    exit 1
fi

# Create a symbolic link to the model in our models directory
echo "Creating symbolic link to model..."
ln -sf "$MODEL_PATH" models/llama3.2

echo "Model has been downloaded and linked to models/llama3.2"
echo "You can now use the model with: ollama run llama3.2"
echo ""
echo "Note: Keep the Ollama server running in a separate terminal with 'make ollama'" 