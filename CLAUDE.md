# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magikarp is an open-source coding assistant CLI tool built with Go that provides an interactive terminal interface for AI-powered coding assistance. It supports multiple LLM providers including Claude, GPT, and Gemini.

## Project Structure

The project follows Go CLI best practices with Cobra framework:

```
magikarp/
├── main.go                    # Entry point (calls cmd.Execute())
├── cmd/                       # Cobra commands
│   └── root.go               # Root command setup
├── internal/                  # Private packages
│   ├── config/               # Configuration management
│   ├── providers/            # LLM provider implementations
│   ├── tools/                # Tool integrations  
│   ├── orchestration/        # Request orchestration
│   └── terminal/             # Terminal UI (Bubble Tea)
├── assets/                   # Images and static files
├── docs/                     # Architecture docs
├── config.yaml               # Configuration file
├── Makefile                  # Build automation
└── go.mod                    # Go module definition
```

## Installation

Users can install magikarp via:
```bash
go install github.com/pprunty/magikarp@latest
```

## Development Commands

- `make build` - Build binary to bin/magikarp
- `make run` - Run the application
- `make install` - Download dependencies and build
- `make clean` - Clean build artifacts
- `make fmt` - Format code
- `make release` - Build release with GoReleaser

## Configuration

The project uses a YAML configuration file (`config.yaml`) to manage:
- LLM provider settings (Anthropic, OpenAI, Gemini)
- Model specifications and temperature settings
- API key management through environment variables

Key environment variables:
- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY` 
- `GEMINI_API_KEY`

Configuration is handled by the `internal/config` package which supports:
- Loading from config.yaml or ~/.magikarp.yaml
- Environment variable expansion
- Configuration validation

## Dependencies

Key dependencies:
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - Terminal UI
- `github.com/liushuangls/go-anthropic/v2` - Anthropic API
- `github.com/sashabaranov/go-openai` - OpenAI API
- `github.com/google/generative-ai-go` - Gemini API
- `gopkg.in/yaml.v3` - YAML parsing

## Terminal UI Architecture

The terminal user interface (TUI) lives under `internal/terminal/`. It is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and follows the classic **model–update–view** cycle.

### File overview
- `menu.go` – defines `MenuModel`, the main-menu Bubble Tea model. It contains:
  - `choices []string` – menu labels
  - `cursor int` – which item is highlighted
  - `choice string` – the item that was chosen
  - `quitting bool` – signal to exit the program
- `ui.go` – helper routines that spin up the Bubble Tea program and route the final selection to business logic. It also exposes `StartUIWithoutAltScreen` for easier debugging.

### Core flow
1. **Init** – returns initial commands (none for the menu).
2. **Update** – handles keyboard events (`up`, `down`, `enter`, `q`) and mutates the model.
3. **View** – renders styled strings using Lipgloss.
4. **Bootstrap** – `StartUI()` instantiates the model, runs the program, and invokes provider-specific handlers after the Bubble Tea loop ends.

### Extending the UI
1. **Add a new screen**: create a new Bubble Tea model in `internal/terminal/<feature>.go`.
2. **Expose it in the menu**: append its label to the `choices` slice in `NewMenuModel()`.
3. **Dispatch**: in `ui.go`’s `switch m.choice` block, call a function that starts the new model.
4. **Return or chain**: decide whether control returns to the main menu (`tea.Quit`) or chains into another model.

### Theming tips
- Put shared Lipgloss styles in `style.go` (e.g. titleStyle, itemStyle, focusedStyle).
- Reference styles instead of repeating hex codes.
- Consider adding a `config.yaml` option for theme selection.

### Development workflow
- Use `make run` for a full-screen TUI; use `StartUIWithoutAltScreen()` during quick iterations.
- Run `go test ./...` with Bubble Tea’s test helpers to validate state transitions.
- Keep models small and composable; prefer switching models over monolithic `Update` statements.

> Guideline: Aim for **one responsibility per model**. Composition over conditionals keeps the UI easy to reason about and test.