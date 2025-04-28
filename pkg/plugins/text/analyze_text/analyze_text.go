package analyze_text

import (
	_ "embed"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

type input struct {
	Path string `json:"path"`
}

type analysis struct {
	LineCount      int `json:"line_count"`
	WordCount      int `json:"word_count"`
	CharCount      int `json:"char_count"`
	NonSpaceCount  int `json:"non_space_count"`
	AlphaCount     int `json:"alpha_count"`
	DigitCount     int `json:"digit_count"`
	PunctuationCount int `json:"punctuation_count"`
}

func Definition() agent.ToolDefinition {
	var sch map[string]interface{}
	_ = json.Unmarshal(schema, &sch)
	return agent.ToolDefinition{
		Name:        "analyze_text",
		Description: "Analyze text for patterns or statistics",
		InputSchema: sch,
		Function:    run,
	}
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
	var in input
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		return agent.NewToolResult("analyze_text", err.Error(), true), nil
	}

	if err := json.Unmarshal(inputBytes, &in); err != nil {
		return agent.NewToolResult("analyze_text", err.Error(), true), nil
	}

	if in.Path == "" {
		return agent.NewToolResult("analyze_text", "path is required", true), nil
	}

	if !filepath.IsLocal(in.Path) {
		return agent.NewToolResult("analyze_text", "path must be local", true), nil
	}

	content, err := os.ReadFile(in.Path)
	if err != nil {
		return agent.NewToolResult("analyze_text", err.Error(), true), nil
	}

	text := string(content)
	stats := analysis{
		LineCount:      strings.Count(text, "\n") + 1,
		WordCount:      len(strings.Fields(text)),
		CharCount:      len(text),
		NonSpaceCount:  len(strings.ReplaceAll(text, " ", "")),
		AlphaCount:     countRunes(text, unicode.IsLetter),
		DigitCount:     countRunes(text, unicode.IsDigit),
		PunctuationCount: countRunes(text, unicode.IsPunct),
	}

	result, err := json.Marshal(stats)
	if err != nil {
		return agent.NewToolResult("analyze_text", err.Error(), true), nil
	}

	return agent.NewToolResult("analyze_text", string(result), false), nil
}

func countRunes(s string, f func(rune) bool) int {
	count := 0
	for _, r := range s {
		if f(r) {
			count++
		}
	}
	return count
} 