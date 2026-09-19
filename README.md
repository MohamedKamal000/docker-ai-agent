<div align="center">
<img src="assets/logo-image.png" alt="docker-ai logo" width="1388"/>

</div>

[![codecov](https://codecov.io/gh/MohamedKamal000/docker-ai-agent/graph/badge.svg)](https://codecov.io/gh/MohamedKamal000/docker-ai-agent)
[![License](https://img.shields.io/github/license/MohamedKamal000/docker-ai-agent)](https://github.com/MohamedKamal000/docker-ai-agent/blob/main/LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/MohamedKamal000/docker-ai-agent)](https://github.com/MohamedKamal000/docker-ai-agent)
[![Go CI](https://github.com/MohamedKamal000/docker-ai-agent/actions/workflows/Go.yml/badge.svg)](https://github.com/MohamedKamal000/docker-ai-agent/actions/workflows/Go.yml)

# Docker AI Agent (docker-ai)

An AI agent that helps you inspect and operate local Docker environments through a terminal UI, without the need to install Docker Desktop. It combines a Docker SDK wrapper with LLM-driven planning to propose the next action, then executes tooling in a controlled loop.

## Demo

![Demo](assets/demo.gif)

## Features

- Docker environment snapshotting via the SDK wrapper (`internal/docker`).
- Prompt templates for system + user execution context (`internal/core/prompts.go`).
- Genkit-backed LLM client for multiple providers (`internal/core/genkit_client.go`).
- **Tool Registry and Executor** for dynamically calling Docker commands (`internal/core/tool_registry.go`, `internal/tools/docker_commands_tool.go`).
- TUI chat interface for interactive agent conversations (`tui/`).
- CLI commands for agent chat and planned docker workflows (`cmd/*`).

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
docker-ai-agent agent-chat
# or
docker-ai-agent ac
```

The TUI provides a chat interface where you can type messages, view agent responses, and interact with the agent in real time.

### Flags

- `-c`, `--config`: Path to the configuration file (default: `config.json`).

### Example

```bash
docker-ai-agent agent-chat -c ./my-config.json
```

## CLI commands (current)

- `agent-chat` (`ac`): Launch the interactive TUI chat session to ask Docker-related questions.
- `containerize` (`c`): Planned – let the AI scan your current directory and generate a Dockerfile.
- `initialize-rag` (`ir`): Planned – set up retrieval augmented generation so the agent can ground answers in indexed documentation.

## Documentation

Full documentation is built with [mdBook](https://rust-lang.github.io/mdBook/) and lives in the [`docs/`](./docs) directory, covering getting started guides (including how to obtain and store API keys), usage, architecture, and a configuration reference.

## Contributing

Issues and PRs are welcome. Please open an issue to discuss larger changes before submitting a PR.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE).
