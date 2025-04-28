package agent

import (
	"context"
)

type Agent struct {
	llm    LLM
	memory *Memory        // pluggable memory impl
	tools  map[string]ToolDefinition
	exec   *Executor     // executes ToolCalls
}

func New(llm LLM, toolDefs []ToolDefinition) *Agent {
	t := make(map[string]ToolDefinition, len(toolDefs))
	for _, td := range toolDefs {
		t[td.Name] = td
	}
	return &Agent{
		llm:    llm,
		memory: NewMemory(),
		tools:  t,
		exec:   NewExecutor(t),
	}
}

func (a *Agent) toolList() []ToolDefinition {
	tools := make([]ToolDefinition, 0, len(a.tools))
	for _, tool := range a.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (a *Agent) Loop(ctx context.Context, userIn <-chan string, uiOut chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		/* 1️⃣ user input */
		case text := <-userIn:
			a.memory.Append(ChatMessage{Role: "user", Content: text})

		/* 2️⃣ ask the LLM */
		default:
			assistant, calls, err := a.llm.Chat(ctx,
				a.memory.Messages(),
				a.toolList(),
			)
			if err != nil {
				return err
			}
			a.memory.Append(assistant...)

			/* 3️⃣ run tools if requested */
			if len(calls) > 0 {
				results := a.exec.Run(ctx, calls)      // may run in parallel
				a.memory.ToolResults(results)

				assistant, err = a.llm.SendToolResults(ctx,
					a.memory.Messages(), results)
				if err != nil {
					return err
				}
				a.memory.Append(assistant...)
			}

			/* 4️⃣ stream or print the assistant's final message */
			if msg := a.memory.LastAssistant(); msg != nil {
				uiOut <- msg.Content
			}
		}
	}
} 