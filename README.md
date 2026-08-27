<div align="center">
<img src="logo-image.png" alt="docker-ai logo" width="1388"/>

</div>

# Docker AI Agent

A Go-based AI agent that helps you inspect and operate local Docker environments through a terminal UI, without the need to install Docker Desktop. It combines a Docker SDK wrapper with LLM-driven planning to propose the next action, then executes tooling in a controlled loop.

## Status

This project is in active development. Core features are in place, including a Docker SDK wrapper, prompt templates, a Genkit client for LLM integration, a tool registry, and a tool execution loop. The CLI is functional for agent-based chat. Next steps involve adding more tools, implementing a user confirmation flow, and building out planned commands like `initialize-rag` and `enable-rag`.

## Architecture

![Architecture](./docs/src/images/project-arch.png)

At a high level:

- TUI (Bubble Tea) provides a chat interface where users can interact with the agent in real time.
- CLI (Cobra) parses commands/flags and launches the TUI.
- The AI agent builds prompts from the current Docker context and user intent.
- A main loop selects the next action, maps it to a tool call from the **Tool Registry**, and stores results in memory.
- The **Tool Registry** discovers and manages available tools (`internal/tools`).
- The Docker SDK wrapper provides structured access to containers, images, volumes, and networks.
- LLM providers are pluggable via Genkit (Gemini, OpenAI, Anthropic, Ollama, DeepSeek, Kimi, Qwen, Grok).

## Features

- Docker environment snapshotting via the SDK wrapper (`internal/docker`).
- Prompt templates for system + user execution context (`internal/core/prompts.go`).
- Genkit-backed LLM client for multiple providers (`internal/core/genkit_client.go`).
- **Tool Registry and Executor** for dynamically calling Docker commands (`internal/core/tool_registry.go`, `internal/tools/docker_commands_tool.go`).
- TUI chat interface for interactive agent conversations (`tui/`).
- CLI commands for agent chat and planned docker workflows (`cmd/*`).

## Prerequisites

- Go 1.25.0
- Docker Engine available locally for Docker SDK calls.
- One of the supported LLM provider API keys.

## Configuration

The agent can be configured via a `config.json` file or environment variables.

### Config File

Create a `config.json` file in the root of the project. You can use the `-c` or `--config` flag to specify a different path.

**Gemini example:**

```json
{
  "provider": "Gemini",
  "model-name": "gemini-2.5-flash-lite",
  "temperature": 0.7,
  "max-tokens": 1024,
  "max-iterations": 10
}
```

**OpenAI example:**

```json
{
  "provider": "OpenAi",
  "model-name": "gpt-4o-mini",
  "temperature": 0.7,
  "max-tokens": 1024,
  "max-iterations": 10
}
```

**Ollama example (local, no API key needed):**

```json
{
  "provider": "Ollama",
  "model-name": "llama3",
  "server-address": "http://localhost:11434",
  "temperature": 0.7,
  "max-tokens": 1024,
  "max-iterations": 10
}
```

We support 8 providers total: Gemini, OpenAI, Anthropic, Ollama, DeepSeek, Kimi, Qwen, and Grok. See the [Reference](./docs/src/reference.md) page for the full list and API key setup instructions.

Note: If `api-key` is not provided, the agent will look for the appropriate environment variable based on the provider.

### Environment Variables

If you prefer not to include the API key in the config file, you can set it as an environment variable.

Examples:

```bash
export GEMINI_API_KEY=your_key_here
export OPENAI_API_KEY=your_key_here
export ANTHROPIC_API_KEY=your_key_here
export DEEPSEEK_API_KEY=your_key_here
export KIMI_API_KEY=your_key_here
export QWEN_API_KEY=your_key_here
export GROK_API_KEY=your_key_here
```

Ollama does not require an API key.

## How to Use

Run `agent-chat` (or its shortcut `ac`) to launch the interactive TUI session:

```bash
go run . agent-chat
# or
go run . ac
```

The TUI provides a chat interface where you can type messages, view agent responses, and interact with the agent in real time.

### Flags

- `-c`, `--config`: Path to the configuration file (default: `config.json`).

### Example

```bash
go run . agent-chat -c ./my-config.json
```

## CLI commands (current)

- `agent-chat` (`ac`): Launch the interactive TUI chat session to ask Docker-related questions.
- `initialize-rag`: Planned – set up retrieval augmented generation so the agent can ground answers in indexed documentation.
- `enable-rag`: Planned – activate RAG mode for the agent session.

## Documentation

Full documentation is built with [mdBook](https://rust-lang.github.io/mdBook/) and lives in the [`docs/`](./docs) directory, covering getting started guides (including how to obtain and store API keys), usage, architecture, and a configuration reference.

## Project layout

```
cmd/                 Cobra CLI commands
internal/core/       Agent loop, prompts, Genkit client, tool registry
internal/docker/     Docker SDK wrapper + exec helpers
internal/models/     Shared data models
internal/tools/      Tool definitions (e.g., Docker commands)
tui/                 Bubble Tea TUI (chat interface, screens, widgets)
```

## Contributing

Issues and PRs are welcome. Please open an issue to discuss larger changes before submitting a PR.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE).
