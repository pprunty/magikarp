package terminal

import "sync/atomic"

var currentModel atomic.Value // string

// SetCurrentModel stores the model name selected by the user/UI.
func SetCurrentModel(name string) {
	currentModel.Store(name)
}

// CurrentModel returns the currently selected model (or empty string if unknown).
func CurrentModel() string {
	if v := currentModel.Load(); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// speech mode global flag
var speechEnabled atomic.Bool

// SetSpeechModeEnabled sets global speech mode flag
func SetSpeechModeEnabled(enabled bool) {
	speechEnabled.Store(enabled)
}

// SpeechModeEnabled returns whether speech mode is globally enabled
func SpeechModeEnabled() bool {
	return speechEnabled.Load()
}
