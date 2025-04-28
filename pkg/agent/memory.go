package agent

import "sync"

type Memory struct {
	mu   sync.Mutex
	msgs []ChatMessage
}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Append(msgs ...ChatMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msgs = append(m.msgs, msgs...)
}

func (m *Memory) Messages() []ChatMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ChatMessage(nil), m.msgs...)
}

func (m *Memory) ToolResults(r []ToolResult) {
	for _, tr := range r {
		m.Append(ChatMessage{
			Role:    "tool",
			Content: tr.Content,
		})
	}
}

func (m *Memory) LastAssistant() *ChatMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.msgs) - 1; i >= 0; i-- {
		if m.msgs[i].Role == "assistant" {
			return &m.msgs[i]
		}
	}
	return nil
} 