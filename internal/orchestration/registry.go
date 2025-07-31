package orchestration

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/pprunty/magikarp/internal/config"
	"github.com/pprunty/magikarp/internal/providers"
	"github.com/pprunty/magikarp/internal/providers/anthropic"
	"github.com/pprunty/magikarp/internal/providers/gemini"
	"github.com/pprunty/magikarp/internal/providers/openai"
)

var (
	modelToProvider   = make(map[string]providers.Provider)
	registryInitOnce  sync.Once
	registryInitError error
)

// Init builds the provider registry from configuration. Safe for concurrent use.
func Init(cfg *config.Config) error {
	registryInitOnce.Do(func() {
		registryInitError = build(cfg)
	})
	return registryInitError
}

func build(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("nil config passed to registry")
	}

	var initErrors []string

	// OpenAI provider
	if pCfg, ok := cfg.Providers["openai"]; ok {
		if pCfg.Key != "" && pCfg.Key != "${OPENAI_API_KEY}" {
			for _, m := range pCfg.Models {
				client := openai.New(pCfg.Key, []string{m})
				modelToProvider[m] = client
			}
		} else {
			initErrors = append(initErrors, "OpenAI: API key not set (OPENAI_API_KEY environment variable)")
		}
	}

	// Anthropic provider
	if pCfg, ok := cfg.Providers["anthropic"]; ok {
		if pCfg.Key != "" && pCfg.Key != "${ANTHROPIC_API_KEY}" {
			for _, m := range pCfg.Models {
				client := anthropic.New(pCfg.Key, []string{m})
				modelToProvider[m] = client
			}
		} else {
			initErrors = append(initErrors, "Anthropic: API key not set (ANTHROPIC_API_KEY environment variable)")
		}
	}

	// Gemini provider
	if pCfg, ok := cfg.Providers["gemini"]; ok {
		if pCfg.Key != "" && pCfg.Key != "${GEMINI_API_KEY}" {
			client, err := gemini.New(pCfg.Key, pCfg.Models)
			if err != nil {
				initErrors = append(initErrors, fmt.Sprintf("Gemini: failed to create client: %v", err))
			} else {
				for _, m := range pCfg.Models {
					modelToProvider[m] = client
				}
			}
		} else {
			initErrors = append(initErrors, "Gemini: API key not set (GEMINI_API_KEY environment variable)")
		}
	}

	if len(modelToProvider) == 0 {
		msg := "No providers initialized. Please set at least one API key:\n"
		for _, e := range initErrors {
			msg += "  - " + e + "\n"
		}
		return errors.New(msg)
	}

	// Print info about initialized providers
	if len(initErrors) > 0 {
		fmt.Printf("Warning: Some providers not initialized:\n")
		for _, err := range initErrors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Printf("\n")
	}

	initializedCount := 0
	for providerName := range cfg.Providers {
		hasModels := false
		for _, m := range cfg.Providers[providerName].Models {
			if _, exists := modelToProvider[m]; exists {
				hasModels = true
				break
			}
		}
		if hasModels {
			initializedCount++
		}
	}

	return nil
}

// ProviderFor returns the provider responsible for the specified model.
func ProviderFor(model string) (providers.Provider, error) {
	p, ok := modelToProvider[model]
	if !ok {
		return nil, fmt.Errorf("no provider registered for model %s", model)
	}
	return p, nil
}

// FirstModel returns an arbitrary model that has a registered provider.
func FirstModel() (string, error) {
	if len(modelToProvider) == 0 {
		return "", fmt.Errorf("no model available")
	}
	names := make([]string, 0, len(modelToProvider))
	for m := range modelToProvider {
		names = append(names, m)
	}
	sort.Strings(names) // deterministic order
	return names[0], nil
}

// Models returns the list of model names currently registered.
func Models() []string {
	names := make([]string, 0, len(modelToProvider))
	for m := range modelToProvider {
		names = append(names, m)
	}
	return names
}
