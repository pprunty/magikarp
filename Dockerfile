# Stage 1: Use the ollama/ollama base image to set up the model
FROM ollama/ollama:latest AS ollama

# Working directory for ollama
WORKDIR /app

COPY ./Modelfile /app

# Start ollama serve and create the model
RUN nohup sh -c "ollama serve &" && sleep 5 && ollama create magikarp -f ./Modelfile

# Expose the port for the ollama app (if needed)
EXPOSE 8000

# Command to run the ollama server
CMD ["ollama", "serve"]