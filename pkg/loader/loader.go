package loader

import (
	"github.com/pprunty/magikarp/pkg/agent"
)

var registry []agent.Plugin

// Register adds a plugin to the registry
func Register(p agent.Plugin) {
	registry = append(registry, p)
}

// All returns all registered plugins
func All() []agent.Plugin {
	return registry
} 