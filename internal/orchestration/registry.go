package orchestration

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/pprunty/magikarp/internal/config"
	"github.com/pprunty/magikarp/internal/providers"
	"github.com/pprunty/magikarp/internal/providers/alibaba"
	"github.com/pprunty/magikarp/internal/providers/anthropic"
	"github.com/pprunty/magikarp/internal/providers/gemini"
	"github.com/pprunty/magikarp/internal/providers/mistral"
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
			temperature := cfg.GetEffectiveTemperature("openai")
			for _, m := range pCfg.Models {
				client := openai.New(pCfg.Key, []string{m}, temperature, cfg.System)
				modelToProvider[m] = client
			}
		} else {
			initErrors = append(initErrors, "OpenAI: API key not set (OPENAI_API_KEY environment variable)")
		}
	}

	// Anthropic provider
	if pCfg, ok := cfg.Providers["anthropic"]; ok {
		if pCfg.Key != "" && pCfg.Key != "${ANTHROPIC_API_KEY}" {
			temperature := cfg.GetEffectiveTemperature("anthropic")
			for _, m := range pCfg.Models {
				client := anthropic.New(pCfg.Key, []string{m}, temperature, cfg.System)
				modelToProvider[m] = client
			}
		} else {
			initErrors = append(initErrors, "Anthropic: API key not set (ANTHROPIC_API_KEY environment variable)")
		}
	}

	// Gemini provider
	if pCfg, ok := cfg.Providers["gemini"]; ok {
		if pCfg.Key != "" && pCfg.Key != "${GEMINI_API_KEY}" {
			temperature := cfg.GetEffectiveTemperature("gemini")
			client, err := gemini.New(pCfg.Key, pCfg.Models, temperature, cfg.System)
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

	// Mistral provider
	if pCfg, ok := cfg.Providers["mistral"]; ok {
		if pCfg.Key != "" && pCfg.Key != "${MISTRAL_API_KEY}" {
			temperature := cfg.GetEffectiveTemperature("mistral")
			client, err := mistral.New(pCfg.Key, pCfg.Models, temperature, cfg.System)
			if err != nil {
				initErrors = append(initErrors, fmt.Sprintf("Mistral: failed to create client: %v", err))
			} else {
				for _, m := range pCfg.Models {
					modelToProvider[m] = client
				}
			}
		} else {
			initErrors = append(initErrors, "Mistral: API key not set (MISTRAL_API_KEY environment variable)")
		}
	}

	// Alibaba provider
	if pCfg, ok := cfg.Providers["alibaba"]; ok {
		if pCfg.Key != "" && pCfg.Key != "${ALIBABA_API_KEY}" {
			temperature := cfg.GetEffectiveTemperature("alibaba")
			client, err := alibaba.New(pCfg.Key, pCfg.Models, temperature, cfg.System)
			if err != nil {
				initErrors = append(initErrors, fmt.Sprintf("Alibaba: failed to create client: %v", err))
			} else {
				for _, m := range pCfg.Models {
					modelToProvider[m] = client
				}
			}
		} else {
			initErrors = append(initErrors, "Alibaba: API key not set (ALIBABA_API_KEY environment variable)")
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

// ModelsByProvider returns a map of provider names to their available models.
func ModelsByProvider(cfg *config.Config) map[string][]string {
	providerModels := make(map[string][]string)

	// Iterate through all configured providers
	for providerName, providerCfg := range cfg.Providers {
		availableModels := make([]string, 0)

		// Check which models from this provider are actually available (have initialized clients)
		for _, model := range providerCfg.Models {
			if _, exists := modelToProvider[model]; exists {
				availableModels = append(availableModels, model)
			}
		}

		// Only include providers that have at least one available model
		if len(availableModels) > 0 {
			providerModels[providerName] = availableModels
		}
	}

	return providerModels
}

// GetInitializedProviders returns a map of provider names to their initialization status.
// Returns true if the provider has at least one successfully initialized model client.
func GetInitializedProviders(cfg *config.Config) map[string]bool {
	providerStatus := make(map[string]bool)
	
	// Check all configured providers
	for providerName, providerCfg := range cfg.Providers {
		hasInitializedClient := false
		
		// Check if any model from this provider has an initialized client
		for _, model := range providerCfg.Models {
			if _, exists := modelToProvider[model]; exists {
				hasInitializedClient = true
				break
			}
		}
		
		providerStatus[providerName] = hasInitializedClient
	}
	
	return providerStatus
}
