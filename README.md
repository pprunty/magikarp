# magikarp

A flexible plugin-based agent framework built with Go. magikarp provides a foundation for building AI-powered agents that can interact with various tools and services through a plugin architecture.

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
   }
   ```
3. Register your plugin in `main.go`

Example plugin structure:
```go
package yourplugin

type YourPlugin struct{}

func New() *YourPlugin {
    return &YourPlugin{}
}

func (p *YourPlugin) Name() string {
    return "your-plugin"
}

func (p *YourPlugin) Description() string {
    return "Description of what your plugin does"
}

func (p *YourPlugin) Tools() []agent.ToolDefinition {
    return []agent.ToolDefinition{
        // Define your tools here
    }
}
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