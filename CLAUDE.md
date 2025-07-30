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