# EVERYTHING Autonomous Agent (Magikarp)

![shiny_Magikarp.png](images/shiny_magikarp.png)

## Description

Magikarp is an autonomous agent built with FastAPI and Ollama, designed to provide chat interactions through a simple API interface. It leverages the DeepSeek-R1 model to deliver conversational AI capabilities in a lightweight, self-hosted package.

## How It Works / Key Features

- **Local LLM Integration**: Uses Ollama to run the DeepSeek-R1 model locally, offering privacy and control over your AI interactions
- **FastAPI Backend**: Provides a robust and performant API for chat interactions
- **Docker Support**: Easily deployable via Docker for consistent environments
- **Customizable Context Window**: Configurable parameters for the AI model to balance performance and capability
- **Poetry Dependency Management**: Modern Python dependency management for reproducible builds

## Pre-requisites

- Python 3.12
- Poetry
- ollama (optional / `ollama` image used in docker commands to run the application)

```shell
brew install ollama
```

> [!NOTE]  
> Ollama is required for local model serving but can be skipped if you're using Docker deployment.

## Quickstart

## Quickstart

### Local Development

1. **All-in-one setup**:
   ```shell
   make everything
   ```
   This command will:
   - Check for required prerequisites (Python, Poetry, Ollama)
   - Install dependencies
   - Start the Ollama server (on port 11434)
   - Create the Magikarp model (takes 10-20 minutes, uses ~5GB)
   - Show clear instructions for next steps

   After the setup completes:
   - To chat with the model via terminal: `make run`
   - To start the API server: `make server` (runs on http://127.0.0.1:8000)
   - To stop the Ollama server: `make stop`

   OR follow these steps individually:

1. Install dependencies:
   ```shell
   make install
   ```

2. Start the Ollama server:
   ```shell
   make serve
   ```

3. Create and run the model (in a new terminal):
   ```shell
   make model
   ```

4. (Optional) Interact with the model directly in terminal:
   ```shell
   make run
   ```

5. Start the FastAPI server (in a new terminal):
   ```shell
   make server
   ```

6. When finished, you can stop the Ollama server running in the other terminal via:
   ```shell
   make stop
   ```

8. Access the API at http://127.0.0.1:8000 or the documentation at http://127.0.0.1:8000/docs

> [!TIP]  
> You can use `make help` to see all available commands.

### Docker Deployment

The easiest way to run this application in a production-like environment is via docker compose:

```shell
make docker-compose
```

This will build and start all necessary containers configured in the docker-compose.yml file.

> [!WARNING]  
> Docker deployment requires significant disk space for model storage. When running with Docker on MacOS, ollama does not have access to the native GPUs on device and therefore model responses will be very slow.

## Model Configuration

Magikarp uses a custom configuration of the DeepSeek-R1 model with the following parameters:

```
FROM deepseek-r1

# Set the context window (default: 2048)
PARAMETER num_ctx 6114
PARAMETER num_thread 4

# Set the temperature (higher is more creative, lower is more coherent)
PARAMETER temperature 0.33

# System message
SYSTEM """
Introduce yourself always as 'Magikarp'.
"""
```

You can modify these parameters in the Modelfile to adjust the model's behavior.

> [!NOTE]  
>  For complete documentation on Modelfile configuration:
> - [Ollama Modelfile Documentation](https://github.com/ollama/ollama/blob/main/docs/modelfile.md) - Detailed reference for all available parameters and configuration options
> - [DeepSeek-R1 Whitepaper](https://github.com/deepseek-ai/DeepSeek-R1/blob/main/DeepSeek_R1.pdf) - Technical details about the DeepSeek-R1 model
> - [Ollama Structured Outputs Guide](https://ollama.com/blog/structured-outputs) - Learn how to get structured JSON responses from your model

## Data Integration (Potential Usecase)

One of the key features of Magikarp is its ability to integrate with local data, providing the LLM with contextual awareness about the user. This creates a more personalized AI experience while keeping sensitive data local.

### How Data Integration Works

1. **Data Storage**: The `/data` directory acts as a personal "dropbox" where various data files can be stored, such as calendar events, location history, social media activity, or user preferences.

2. **Data Service**: The `DataService` class loads and manages this data:
   ```python
   class DataService:
       def __init__(self, base_path: str = 'data'):
           self.base_path = os.path.join(os.getcwd(), base_path)
           # Data properties are loaded lazily
           self._file_contents = {}
           
       def load_all_files(self):
           # Dynamically loads all files in the data directory
           
       def get_formatted_data(self):
           # Formats the loaded data for injection into the model context
   ```

3. **Model Integration**: The `TransformerModel` class injects this data into conversations:
   ```python
   def ask_model(self, prompt: str) -> str:
       # Format user data and combine with prompt
       user_message = {'role': 'user', 'content': self.formatted_user_data + prompt}
       self.chat_messages.append(user_message)
       # Send to LLM
       response = ollama.chat(model='magikarp', messages=self.chat_messages)
       return response['message']['content']
   ```

> [!IMPORTANT]  
> All data is processed locally on your device and is never sent to external servers, ensuring privacy and data sovereignty.

### Use Cases & Examples

1. **Personalized Recommendations**:
   ```
   "Based on your recent workout history and location data, I notice you've been jogging in Central Park. Have you considered trying the riverside path for a change of scenery?"
   ```

2. **Contextual Notifications**:
   ```
   "You have a meeting with Alex in 30 minutes, and traffic looks heavy on your usual route. You might want to leave 10 minutes earlier."
   ```

3. **Data-Driven Insights**:
   ```
   "I've noticed your Spotify playlists include a lot of focus music on weekday mornings. Would you like me to automatically queue up your 'Deep Focus' playlist at 9am on workdays?"
   ```

### Getting Started with Your Own Data

To experiment with this feature:

1. Create your own data files in the `/data` directory following the existing formats
2. Modify the `DataService` class if needed to handle new data types
3. Run the application and interact with the AI

> [!TIP]
> Try adding a simple JSON file with your interests to see how the AI incorporates that information into responses.

This approach demonstrates a powerful use case for local LLMs: creating AI assistants that can access and reason with personal data without privacy concerns, as all processing remains on your device.

### Future Possibilities

This pattern opens up numerous possibilities:
- Integration with local health data (from wearables)
- Processing personal document collections
- Analysis of private communications
- Custom data connectors for other services

By keeping sensitive data local while leveraging the power of open-source LLMs, Magikarp provides a foundation for building truly personalized AI assistants without compromising privacy.

## APIs

The application provides the following APIs:

### Chat API

- **Endpoint**: `/chat`
- **Method**: POST
- **Description**: Send prompts to the AI and receive responses
- **Request Body**:
  ```json
  {
    "prompt": "Hello there."
  }
  ```
- **Response**:
  ```json
  {
    "response": "Hey! It's me, Magikarp! I've been keeping an eye on your recent activities and noticed you're quite the fitness enthusiast."
  }
  ```

### Root Endpoint

- **Endpoint**: `/`
- **Method**: GET
- **Description**: Returns a welcome message with instructions to access the documentation

> [!TIP]  
> You can test the API using the interactive Swagger UI at http://127.0.0.1:8000/docs

## Project Structure

```
magikarp/
├── app/                     # Main application code
│   ├── __init__.py
│   ├── dependencies.py      # FastAPI dependencies
│   ├── main.py              # FastAPI application entry point
│   ├── enums/               # Enumeration definitions
│   ├── models/              # Pydantic models for API
│   ├── routers/             # API route definitions
│   ├── services/            # Business logic services
│   └── utils/               # Utility functions
├── data/                    # Application data files (below are examples)
│   ├── calendar.csv
│   ├── location.csv
│   ├── social_media.json
│   ├── spotify_playlists.json
│   └── user_profile.json
├── models/                  # Directory for Ollama models (mostly gitignored)
├── images/                  # Project images
├── Dockerfile.fastapi       # Docker configuration for the FastAPI service
├── docker-compose.yaml      # Docker Compose configuration
├── entrypoint.sh            # Docker entrypoint script
├── Modelfile                # Ollama model configuration
├── pyproject.toml           # Poetry configuration
├── poetry.lock              # Poetry lock file
├── Makefile                 # Build and development tasks
├── README.md                # Project documentation
└── LICENSE                  # License information
```

## Make Commands

The project includes a comprehensive Makefile to simplify common tasks:

## Make Commands

The project includes a comprehensive Makefile to simplify common tasks:

### Poetry Commands
- `make install` - Install project dependencies using Poetry
- `make update` - Update project dependencies
- `make lock` - Generate Poetry lock file

### API Commands
- `make server` - Start FastAPI development server (runs on http://127.0.0.1:8000)

### Ollama Commands
- `make serve` - Start Ollama server with models directory (runs on port 11434)
- `make model` - Create Ollama model from Modelfile
- `make run` - Run the Ollama model in interactive terminal mode
- `make delete` - Delete the Ollama model
- `make stop` - Stop the Ollama server running in the background

### Docker Commands
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make docker-compose` - Start with docker-compose

### Development Commands
- `make everything` - One-command setup that:
   - Checks for required prerequisites
   - Installs dependencies
   - Starts the Ollama server
   - Creates the model (takes 10-20 minutes, uses ~5GB)
   - Provides clear instructions for next steps

## Resource Considerations

When running Ollama models locally, be mindful of:

> [!CAUTION]  
> Large language models can be resource-intensive. Monitor your system's performance and adjust parameters as needed.

- **Context Window Size**: The default is set to 6114 tokens, which should work on most modern machines. Larger context windows (up to 32768) require significantly more RAM and may cause performance issues.
- **Thread Count**: Set to 4 by default, adjust based on your CPU capabilities.
- **Temperature**: Set to 0.33 for more deterministic responses. Increase for more creative outputs.

## Improvements

Potential areas for enhancement:

1. **Structured Output Mode**: Add configuration flag to enable JSON-formatted responses for programmatic consumption
2. **Adaptive Resource Management**: Implement automatic resource detection and optimization when running locally to balance performance and system load
3. **Streaming API Responses**: Enhance the FastAPI `/chat` endpoint with SSE or WebSocket support for real-time streaming responses
4. **Multi-agent Orchestration**: Enable running multiple model instances in parallel to generate, refine, and select optimal responses
5. **Contextual Memory**: Implement persistent chat history that feeds back into the model context for improved conversation coherence
6. **System Integration Capabilities**: Explore secure frameworks for allowing the agent to execute commands, create or modify files based on structured output prompts
7. **Lightweight Web Interface**: Develop a minimal browser-based UI with WebSocket support for direct interactions and streaming responses
8. **User Authentication**: Add multi-user support with personal data isolation
9. **Vector Database Integration**: Implement retrieval-augmented generation for improved factual responses
10. **Extensible Plugin System**: Create a modular architecture that allows for community-developed extensions
11. **Multi-modal Input Processing**: Add support for processing images and other non-text inputs
12. **Comprehensive Observability**: Enhance logging, monitoring, and performance tracking
13. **Fine-tuning Toolkit**: Provide utilities for customizing the base model with domain-specific knowledge
14. **Agent Automation Framework**: Expand capabilities to allow for scheduled or event-triggered autonomous actions

> [!NOTE]  
> Feel free to submit pull requests for any of these improvements or suggest new features!