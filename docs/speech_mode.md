# Speech-to-Text Mode Architecture

This document describes how **Magikarp** will support voice input inside the TUI via a `/speech` toggle. The goal is to let users dictate prompts and submit them by saying the keyword **“over”**.

---

## 1. Technology choice

### Whisper.cpp (recommended)
* Offline, open-source, MIT licence.
* High accuracy for English and many other languages.
* Supports streaming / real-time transcription.
* Go bindings available: `github.com/ggerganov/whisper.cpp/bindings/go`.
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
    input.go         ← NEW (text box Bubble Tea model)
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