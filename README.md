# Magikarp

Magikarp is a language-agnostic tool system for LLM agents with a modular architecture. This allows tools to be developed in any language and used seamlessly within LLM agents.

## Features

- Language-agnostic tool system (Go, Python, etc.)
- Modular architecture with tool registry
- Plugin-based tool discovery and registration
- Command-line interface for managing tools and agents
- Support for multiple LLM providers (OpenAI, Anthropic, etc.)

## Directory Structure

```
magikarp/
├── cmd/
│   └── mgk/                  # CLI entry-point
│       └── main.go
├── internal/                 # shared, non-public helpers
│   └── registry/
│       ├── registry.go       # runtime registry for tools & agents
│       └── loader.go         # loads JSON descriptors
├── models/                   # LLM wrappers
│   ├── client.go             # common interface
│   ├── openai/
│   │   └── openai.go
│   ├── anthropic/
│   │   └── anthropic.go
│   └── ollama.go
├── tools/                    # one dir per tool
│   ├── filesystem/
│   │   ├── filesystem.go
│   │   ├── read_file.go
│   │   └── list_files.go
│   ├── execution/
│   │   └── shell_exec.go
│   └── text/
│       └── rot13.go
├── agents/                   # orchestration logic
│   ├── base.go
│   ├── coding/
│   │   └── coding.go
│   └── planner/
│       └── planner.go
└── utils/
    └── history/
        └── history.go
```

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/magikarp.git
cd magikarp

# Build the main application and CLI
make build

# Alternatively, install the CLI globally
make install-cli
```

## Usage

### Command-Line Interface

Magikarp comes with a command-line interface (CLI) named `mgk` for managing tools and agents:

```bash
# List available tools and agents
mgk list

# Create a new tool
mgk create tool mytool --language go --description "My awesome tool"

# Create a new agent
mgk create agent myagent --description "My awesome agent"

# Run an agent
mgk run myagent
```

### Tool Development

Tools can be developed in any language by implementing the tool interface:

#### Go Tool Example

```go
package tool_example

import (
	"encoding/json"
	"fmt"
	
	"github.com/pprunty/magikarp/internal/registry"
)

func init() {
	registry.RegisterGoTool(
		"my_tool",
		"Description of my tool",
		// Input schema
		map[string]interface{}{
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"input"},
			"type": "object",
		},
		// Output schema
		map[string]interface{}{
			"properties": map[string]interface{}{
				"output": map[string]interface{}{
					"type": "string",
				},
			},
			"type": "object",
		},
		// Implementation
		func(args json.RawMessage) (json.RawMessage, error) {
			// Parse input
			var input struct {
				Input string `json:"input"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, err
			}
			
			// Process input
			result := "Processed: " + input.Input
			
			// Return output
			return json.Marshal(map[string]string{
				"output": result,
			})
		},
	)
}
```

#### Python Tool Example

```python
#!/usr/bin/env python3
import json
import sys

def main():
    # Read input JSON from stdin
    args = json.load(sys.stdin)
    
    # Process input
    result = "Processed: " + args.get("input", "")
    
    # Write result as JSON to stdout
    json.dump({"output": result}, sys.stdout)

if __name__ == "__main__":
    main()
```

With a corresponding `tool.json`:

```json
{
  "name": "my_tool",
  "description": "Description of my tool",
  "language": "python",
  "entrypoint": "main.py",
  "run": "python3 main.py",
  "inputs": {
    "properties": {
      "input": {
        "type": "string",
        "description": "Input to process"
      }
    },
    "required": ["input"],
    "type": "object"
  },
  "outputs": {
    "properties": {
      "output": {
        "type": "string",
        "description": "Processed output"
      }
    },
    "type": "object"
  }
}
```

## Agent Development

Agents are responsible for orchestrating the interaction between LLMs and tools:

```go
package myagent

import (
	"context"
	
	"github.com/pprunty/magikarp/agents"
)

type MyAgent struct {
	*agents.BaseAgent
}

func NewMyAgent(getUserMessage func() (string, bool)) *MyAgent {
	baseAgent := agents.NewBaseAgent("myagent", "My custom agent", getUserMessage)
	return &MyAgent{
		BaseAgent: baseAgent,
	}
}

func (a *MyAgent) Run(ctx context.Context) error {
	// Implement agent logic here
	return nil
}
```

## License

MIT 