# Speech-to-Text Mode Architecture

This document describes how **Magikarp** will support voice input inside the TUI via a `/speech` toggle. The goal is to let users dictate prompts and submit them by saying the keyword **“over”**.

---

## 1. Technology choice

### Whisper.cpp (recommended)
* Offline, open-source, MIT licence.
* High accuracy for English and many other languages.
* Supports streaming / real-time transcription.
* Go bindings available: `github.com/ggerganov/whisper.cpp/bindings/go` **or** the higher-level wrapper `github.com/mutablelogic/go-whisper` (preferred).
* Relies on **PortAudio** for microphone capture (`github.com/gordonklaus/portaudio`).

### Alternatives considered
| Engine | Pros | Cons |
| ------ | ---- | ---- |
| **Vosk** | Very light models (<60 MB). Simple Go API. | Lower accuracy on casual speech. |
| Google / AssemblyAI API | Great quality, no local CPU usage. | Requires internet & paid after free tier. |

We choose **Whisper.cpp** for quality + offline capability.

---

## 2. High-level flow

```text
User presses ↵ on “/speech” command
└── TUI toggles SpeechMode = ON
    └── Spawn goroutine speech.Listen(ctx, outCh)
        └── PortAudio stream → Whisper recognizer → text chunks
            ├── Partial transcripts → (optional) live preview in input box
            └── Final transcript containing "over" → stripped keyword → send to input handler
                └── Same path as typed input (dispatch to LLM etc.)
```

* Speech mode stays enabled until user types `/speech` again or presses `Esc`.
* Background goroutine is cancelled via context.

---

## 3. Package layout

```
internal/
  speech/            ← new
    recognizer.go    (wrap Whisper.cpp)
    listener.go      (microphone → recognizer → channel)
  terminal/
    menu.go
    ui.go
    input.go         ← (text box Bubble Tea model)
```

### internal/speech
* `func Listen(ctx context.Context, out chan<- string, cfg Config) error`
  * opens PortAudio, feeds audio to Whisper.
  * Detects boundary word **"over"** (case-insensitive, trimmed).
  * Sends transcript without the keyword on `out`.
* `type Config struct { ModelPath string; SampleRate int }`

### terminal/input.go
Bubble Tea model that renders a single-line prompt input.
* Receives keystrokes _or_ messages of type `speechMsg { text string }`.
* On receive `speechMsg`, append text to the input buffer and simulate Enter.

---

## 4. Command handling
Add a **slash command** parser in the input model:
* `/speech` – toggle speech mode.
* Optional `/speech off` to force disable.

The input model maintains `speechOn bool` and emits a `toggleSpeechMsg` so that `ui.go` can start/stop the listener goroutine.

---

## 5. Concurrency sketch
```go
// ui.go
listenerCtx, cancel := context.WithCancel(ctx)
var speechCh = make(chan string)

if toggleSpeechMsg.On {
    go speech.Listen(listenerCtx, speechCh, cfg.Speech)
}

for {
    select {
    case <-ctx.Done():
        cancel()
        return
    case txt := <-speechCh:
        teaProgram.Send(speechMsg{text: txt})
    }
}
```

---

## 6. Configuration
Extend `config.yaml`:
```yaml
speech:
  model_path: ~/.magikarp/models/ggml-small.en.bin
  sample_rate: 16000
  keyword: "over"
```
Load this into `Config` and pass to `speech.Listen`.

---

## 7. Makefile additions
```makefile
whisper-models:
	mkdir -p $(HOME)/.magikarp/models && \
	curl -L -o $(HOME)/.magikarp/models/ggml-small.en.bin \
	    https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin
```

---

## 8. Milestones
1. 🎙️ **MVP**: microphone → Whisper → print transcript in console.
2. 🔗 **TUI integration**: `/speech` toggles listening; transcripts fill input.
3. 🚦 **Keyword detection**: submit on "over".
4. 🧪 **Tests**: unit test recognizer with WAV fixtures.
5. ⚙️ **Configurable models** & automatic download.
6. 🐳 **Cross-platform**: verify on macOS, Linux, Windows.

---

## 9. Risks & mitigations
* **CPU load** – start with `tiny` or `small` models; allow user-selectable size.
* **PortAudio device issues** – enumerate devices and pick default; expose `--mic` flag.
* **Latency** – chunk size tuning; maybe switch to faster Whisper `int8` models.

---

Happy coding! 🐟🎤

---

## 10. Claude Code Agents for Implementation

To efficiently implement the speech-to-text feature, create specialized agents using `/agents` in Claude Code. Each agent should have focused expertise and specific instructions.

### Agent 1: Go Dependencies Manager
```
/agents create go-deps-agent

Name: Go Dependencies Manager
Instructions: You are an expert in Go module management and CGO integration. Focus on:
- Adding whisper.cpp Go bindings to go.mod 
- Setting up PortAudio dependencies for cross-platform audio capture
- Handling CGO compilation flags and library linking
- Resolving build issues with C dependencies
- Ensuring cross-platform compatibility (macOS, Linux, Windows)

Context: Working on github.com/pprunty/magikarp - a CLI tool that needs speech-to-text via Whisper.cpp. The project structure uses internal/ packages and follows Go CLI best practices.
```

### Agent 2: Speech Package Architect  
```
/agents create speech-architect

Name: Speech Package Architect
Instructions: You are an expert in Go concurrency and audio processing. Focus on:
- Implementing internal/speech package with recognizer.go and listener.go
- Designing clean APIs for microphone capture → Whisper → text output
- Handling goroutine lifecycle and context cancellation
- Implementing keyword detection ("over") and text preprocessing
- Error handling for audio device failures and model loading

Context: Building speech-to-text for Magikarp CLI. Must integrate with Bubble Tea TUI and be callable from terminal UI models. Follow the architecture in docs/speech_mode.md.
```

### Agent 3: Bubble Tea UI Integration Expert
```
/agents create bubbletea-integration

Name: Bubble Tea UI Integration Expert  
Instructions: You are an expert in Bubble Tea framework and terminal UIs. Focus on:
- Creating input.go model for text input with speech message handling
- Implementing slash command parsing (/speech toggle)
- Managing UI state transitions between typing and speech modes  
- Integrating speech transcription into existing terminal/menu.go flow
- Handling real-time UI updates during speech recognition

Context: Extending Magikarp's existing Bubble Tea UI (internal/terminal/) to support voice input. Must maintain existing menu structure while adding speech capabilities.
```

### Agent 4: Configuration & Build Systems
```
/agents create config-build-expert

Name: Configuration & Build Systems Expert
Instructions: You are an expert in Go configuration management and build automation. Focus on:
- Extending internal/config/config.go to support speech settings
- Adding Whisper model management to Makefile
- Implementing model download and caching logic
- Adding speech configuration to config.yaml schema
- Setting up proper build flags for CGO dependencies

Context: Adding speech-to-text config to Magikarp. Must integrate with existing YAML config system and Makefile build process. Support model path configuration and audio device selection.
```

### Agent 5: Testing & Documentation
```
/agents create test-doc-specialist

Name: Testing & Documentation Specialist
Instructions: You are an expert in Go testing and technical documentation. Focus on:
- Writing unit tests for speech package with audio fixtures
- Creating integration tests for TUI speech mode
- Documenting speech configuration options
- Writing troubleshooting guides for common audio issues
- Updating README.md with speech mode usage examples

Context: Ensuring speech-to-text feature in Magikarp is well-tested and documented. Must cover cross-platform audio testing and provide clear user instructions.
```

### Usage Strategy
1. **Start with Agent 1** to handle dependencies and build setup
2. **Use Agent 2** to implement core speech recognition logic  
3. **Engage Agent 3** for TUI integration and user experience
4. **Consult Agent 4** for configuration and automation
5. **Finish with Agent 5** for testing and documentation

Each agent should reference `docs/speech_mode.md` and understand the existing codebase structure under `internal/`. Coordinate between agents by sharing relevant code snippets and discussing integration points. 