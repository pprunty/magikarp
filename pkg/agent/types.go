package agent

import "github.com/invopop/jsonschema"

// ToolDefinition defines a tool that can be used by the agent
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	Function    func(input []byte) (string, error)
}

// GenerateSchema generates a JSON schema for the given type
func GenerateSchema[T any]() map[string]interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:           true,
	}
	var v T
	schema := reflector.Reflect(v)
	return map[string]interface{}{
		"type":       "object",
		"properties": schema.Properties,
	}
} 