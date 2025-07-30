# magikarp

An open-source coding assistant CLI tool built with Go. Magikarp provides an interactive terminal interface for AI-powered coding assistance with support for multiple LLM providers (Claude, GPT, Gemini).

The project's journey can be followed over on Subtack at [Build Your Own Claude Code](https://furrycircuits.io). Stay tuned.

![v0.1.0.png](./assets/images/v0.1.0.webp)

## Prerequisites

Before installing Magikarp, ensure you have the following installed:

- **Go 1.24.2 or later** - [Download Go](https://golang.org/dl/)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** (optional, but recommended) - Usually pre-installed on macOS/Linux, [install on Windows](https://gnuwin32.sourceforge.net/packages/make.htm)

## Quickstart

### Local Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/pprunty/magikarp.git
   cd magikarp
   ```

2. **Install dependencies and build:**
   ```bash
   make install
   ```

3. **Run Magikarp:**
   ```bash
   make run
   ```

### Alternative Commands

- **Build only:** `make build` (creates `bin/magikarp`)
- **Clean build artifacts:** `make clean`
- **Format code:** `make fmt`
- **View all commands:** `make help`

### Configuration

Set up your API keys by creating a `.env` file:

1. **Copy the template:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your API keys:**
   ```bash
   # Anthropic API Key (for Claude models)
   ANTHROPIC_API_KEY=your_anthropic_api_key_here

   # OpenAI API Key (for GPT models)  
   OPENAI_API_KEY=your_openai_api_key_here

   # Google Gemini API Key (for Gemini models)
   GEMINI_API_KEY=your_gemini_api_key_here
   ```

Alternatively, you can set up your API keys as environment variables:

```bash
export ANTHROPIC_API_KEY="your-anthropic-key"
export OPENAI_API_KEY="your-openai-key"
export GEMINI_API_KEY="your-gemini-key"
```
