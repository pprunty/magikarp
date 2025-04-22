# magikarp

A flexible plugin-based agentic LLM framework built with Go. Magikarp provides a foundation for building AI-powered agents that can interact with the user using _"**tools**"_—custom code which the agent can execute to help them complete their task.
These tools form the basis for the plugin system. 

A plugin is nothing but a "tool", where the tool is a predefined 
block of code the agent knows how to execute, such as `read_file`, `edit_file`, `execute_command`, `trigger_api`, etc.

![shiny_magikarp.png](assets/shiny_magikarp.png)

## Prerequisites

Before you begin, ensure you have the following installed:

1. **Go 1.21 or later**
   - Option 1: Download and install from [golang.org](https://golang.org/dl/)
   - Option 2: Install via Homebrew:
     ```bash
     brew install go
     ```
   - Verify installation: `go version`

2. **Anthropic API Key** (if using Claude integration)
   - Sign up for an API key at [Anthropic's website](https://console.anthropic.com/)
   - Set your API key as an environment variable:
     ```bash
     export ANTHROPIC_API_KEY="your-api-key-here"
     ```

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/pprunty/magikarp.git
   cd magikarp
   ```

2. Install dependencies and build:
   ```bash
   make install
   make build
   ```

## Quick Start

Run the agent:
```bash
make run
```

## Model Selection

magikarp supports multiple LLM providers and models. When you start the application, you'll be presented with a selection menu to choose your preferred provider and model.

### Available Providers

1. **Auto (Smart Model Selection)**
   - Automatically selects the best model based on the task
   - No configuration required

2. **Anthropic (Claude)**
   - Requires `ANTHROPIC_API_KEY` environment variable
   - Models:
     - `claude-3-opus`: Most capable model for highly complex tasks
     - `claude-3-sonnet`: Balanced model for most tasks

3. **OpenAI**
   - Requires `OPENAI_API_KEY` environment variable
   - Models:
     - `gpt-4`: Most capable model for complex tasks
     - `gpt-3.5-turbo`: Fast and efficient model for most tasks

4. **Google Gemini**
   - Requires `GEMINI_API_KEY` environment variable
   - Models:
     - `gemini-pro`: Most capable Gemini model

5. **Ollama**
   - Local model running on your machine
   - Models:
     - `llama3.2`: Local Llama 3.2 model
   - Setup Instructions:
     1. Install Ollama:
        - macOS: `brew install ollama`
        - Linux: `curl -fsSL https://ollama.com/install.sh | sh`
        - Windows: Download from [Ollama website](https://ollama.com/download)
     2. Start the Ollama server:
        ```bash
        ollama serve
        ```
     3. Pull the model you want to use:
        ```bash
        ollama pull llama3.2
        ```
     4. Verify the model is available:
        ```bash
        ollama list
        ```
   - Documentation:
     - [Ollama Documentation](https://github.com/ollama/ollama/blob/main/docs/README.md)
     - [Available Models](https://github.com/ollama/ollama/blob/main/docs/models.md)
     - [API Documentation](https://github.com/ollama/ollama/blob/main/docs/api.md)

### Selecting a Model

1. Start the application:
   ```bash
   make run
   ```

2. Use the arrow keys to navigate through the provider list
   - Press `↑` or `↓` to move between options
   - Press `←` to go back to previous menu
   - Press `/` to search

3. After selecting a provider, choose your preferred model
   - Each model includes a description of its capabilities
   - The selection menu shows both model name and description

4. For API-based providers (Anthropic, OpenAI, Gemini), ensure you have set the required environment variables before starting the application.

## Tool Configuration

magikarp uses a `tools.json` file to configure tool behavior and trigger keywords. This file defines how tools are presented to the LLM and when they should be used.

### tools.json Structure

```json
{
  "tools": [
    {
      "name": "tool_name",
      "description": "Description of what the tool does",
      "category": "Category name (e.g., filesystem, execution, text)",
      "trigger_keywords": ["keyword1", "keyword2"],
      "input_schema": {
        "type": "object",
        "properties": {
          "property1": {
            "type": "string",
            "description": "Description of the property"
          }
        }
      }
    }
  ]
}
```

### Configuration Fields

- `name`: The name of the tool (must match the tool's implementation)
- `description`: A clear description of what the tool does
- `category`: The category the tool belongs to
- `trigger_keywords`: Keywords that hint when the tool should be used
- `input_schema`: JSON schema defining the tool's input parameters

### Example Configuration

```json
{
  "tools": [
    {
      "name": "read_file",
      "description": "Read the contents of a file",
      "category": "filesystem",
      "trigger_keywords": ["read", "file", "contents", "show"],
      "input_schema": {
        "type": "object",
        "properties": {
          "path": {
            "type": "string",
            "description": "Path to the file to read"
          }
        }
      }
    },
    {
      "name": "execute_command",
      "description": "Execute a shell command",
      "category": "execution",
      "trigger_keywords": ["run", "execute", "command", "shell"],
      "input_schema": {
        "type": "object",
        "properties": {
          "command": {
            "type": "string",
            "description": "Command to execute"
          },
          "args": {
            "type": "array",
            "items": {
              "type": "string"
            },
            "description": "Command arguments"
          }
        }
      }
    }
  ]
}
```

### Using tools.json

1. Create a `tools.json` file in your project root
2. Define your tools with appropriate trigger keywords
3. The agent will automatically load this configuration when starting
4. Tools will be triggered when user input contains any of the defined keywords or the model infers the tool should be used

### Best Practices

- Use clear, descriptive tool names
- Provide detailed descriptions
- Choose relevant trigger keywords
- Keep input schemas simple and well-documented
- Group related tools in the same category
- Test trigger keywords to ensure they don't conflict

## Development

### Adding New Plugins

magikarp's plugin architecture makes it easy to add new capabilities. To create a new plugin:

1. Create a new package in `pkg/plugins/your-plugin-name`
2. Implement the `Plugin` interface:
   ```go
   type Plugin interface {
       Name() string
       Description() string
       Tools() []ToolDefinition
       Initialize() error
       Cleanup() error
   }
   ```

3. Use the `BasePlugin` for common functionality:
   ```go
   type YourPlugin struct {
       *BasePlugin
   }

   func New() *YourPlugin {
       return &YourPlugin{
           BasePlugin: NewBasePlugin("your-plugin", "Description of what your plugin does"),
       }
   }
   ```

4. Add tools in the `Initialize` method:
   ```go
   func (p *YourPlugin) Initialize() error {
       p.AddTool("tool_name", "Description of the tool", 
           map[string]interface{}{
               "type": "object",
               "properties": {
                   "input": {
                       "type": "string",
                       "description": "Tool input"
                   }
               }
           },
           func(input []byte) (string, error) {
               result := NewToolResult(true, "Success", data)
               return result.ToJSON()
           })
       return nil
   }
   ```

5. Register your plugin in `main.go`

Example plugin structure:
```go
package yourplugin

type YourPlugin struct {
    *BasePlugin
}

func New() *YourPlugin {
    return &YourPlugin{
        BasePlugin: NewBasePlugin("your-plugin", "Description of what your plugin does"),
    }
}

func (p *YourPlugin) Initialize() error {
    p.AddTool("example_tool", "Example tool description",
        map[string]interface{}{
            "type": "object",
            "properties": {
                "input": {
                    "type": "string",
                    "description": "Example input"
                }
            }
        },
        func(input []byte) (string, error) {
            var data struct {
                Input string `json:"input"`
            }
            if err := json.Unmarshal(input, &data); err != nil {
                return "", err
            }
            result := NewToolResult(true, "Processed successfully", data.Input)
            return result.ToJSON()
        })
    return nil
}

func (p *YourPlugin) Cleanup() error {
    return nil
}
```

### Tool Definition

Each tool must implement the `ToolDefinition` interface:
```go
type ToolDefinition struct {
    Name        string                 // Name of the tool
    Description string                 // Description of what the tool does
    InputSchema map[string]interface{} // JSON schema for tool input
    Function    func(input []byte) (string, error) // Tool implementation
}
```

### Tool Results

Tools should return results using the `ToolResult` struct:
```go
type ToolResult struct {
    Success bool        // Whether the tool execution was successful
    Message string      // Status message
    Data    interface{} // Result data
}
```

Use `NewToolResult` to create results and `ToJSON` to convert them to JSON:
```go
result := NewToolResult(true, "Operation successful", data)
jsonResult, err := result.ToJSON()
```

### Available Commands

- `make install` - Install dependencies
- `make build` - Build the binary
- `make run` - Build and run the binary
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make lint` - Run linter
- `make tools` - Install development tools

## Project Structure

```
magikarp/
├── pkg/
│   ├── agent/         # Core agent implementation
│   ├── llm/           # LLM provider implementations
│   │   ├── anthropic.go  # Anthropic (Claude) client
│   │   ├── auto.go       # Auto model selection client
│   │   ├── gemini.go     # Google Gemini client
│   │   ├── ollama.go     # Ollama local model client
│   │   └── openai.go     # OpenAI client
│   ├── plugins/       # Plugin implementations
│   │   ├── execution/ # Command execution plugin
│   │   ├── filesystem/# File system operations plugin
│   │   └── text/      # Text manipulation plugin
├── main.go           # Entry point
├── Makefile         # Build and development commands
└── README.md        # This file
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details 