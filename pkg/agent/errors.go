package agent

import "errors"

var (
	ErrToolNotFound = errors.New("tool not found")
	ErrInvalidInput = errors.New("invalid tool input")
	ErrToolExecution = errors.New("tool execution failed")
	ErrLLMResponse = errors.New("LLM response error")
) 