#!/bin/bash

# Start Ollama in the background.
/bin/ollama serve &
# Record Process ID.
pid=$!

# Pause for Ollama to start.
sleep 5

echo "🔴 Retrieve LLAMA3 model..."
ollama pull deepseek-r1
echo "🟢 Done!"

## Wait for Ollama process to finish.
#wait $pid

/bin/ollama create magikarp -f ./app/Modelfile
echo "🟢 Magikarp agent created!"

# Wait for Ollama magikarp model creation process to finish.
wait $pid