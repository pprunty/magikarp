package agent

import (
	"context"
	"sync"
)

type Executor struct {
	tools map[string]ToolDefinition
}

func NewExecutor(td map[string]ToolDefinition) *Executor {
	return &Executor{tools: td}
}

func (x *Executor) Run(ctx context.Context, calls []ToolCall) []ToolResult {
	wg := sync.WaitGroup{}
	res := make([]ToolResult, len(calls))
	for i, call := range calls {
		wg.Add(1)
		go func(i int, c ToolCall) {
			defer wg.Done()
			td, ok := x.tools[c.Name]
			if !ok {
				res[i] = ToolResult{ID: c.ID, Content: "tool not found", IsError: true}
				return
			}
			// timeout / cancellation per-tool
			r, err := td.Fn(ctx, c.RawInput)
			res[i] = ToolResult{ID: c.ID, Content: r, IsError: err != nil}
		}(i, call)
	}
	wg.Wait()
	return res
} 