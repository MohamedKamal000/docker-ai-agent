# Docker AI Agent

A Go-based AI agent that helps you inspect and operate local Docker environments through a CLI (and planned TUI/web frontends). It combines a Docker SDK wrapper with LLM-driven planning to propose the next action, then executes tooling in a controlled loop.

## Status

This project is in active development. Core features are in place, including a Docker SDK wrapper, prompt templates, a Genkit client for LLM integration, a tool registry, and a tool execution loop. The CLI is functional for agent-based chat. Next steps involve adding more tools, implementing a user confirmation flow, and building out planned commands like `containerize` and `diagnose`.

## Architecture

![Architecture](./project-arch.png)

At a high level:

- CLI (Cobra) parses commands/flags and passes the user goal to the agent runtime.
- The AI agent builds prompts from the current Docker context and user intent.
- A main loop selects the next action, maps it to a tool call from the **Tool Registry**, and stores results in memory.
- The **Tool Registry** discovers and manages available tools (`internal/tools`).
- The Docker SDK wrapper provides structured access to containers, images, volumes, and networks.
- LLM providers are pluggable via Genkit (Gemini, OpenAI, Anthropic).

## Features

- Docker environment snapshotting via the SDK wrapper (`internal/docker`).
- Prompt templates for system + user execution context (`internal/core/prompts.go`).
- Genkit-backed LLM client for multiple providers (`internal/core/genkit_client.go`).
- **Tool Registry and Executor** for dynamically calling Docker commands (`internal/core/tool_registry.go`, `internal/tools/docker_commands_tool.go`).
- CLI commands for agent chat and planned docker workflows (`cmd/*`).

## Prerequisites

- Go 1.25.0
- Docker Engine available locally for Docker SDK calls.
- One of the supported LLM provider API keys.

## Configuration

The agent can be configured via a `config.json` file or environment variables.

### Config File

Create a `config.json` file in the root of the project. You can use the `-c` or `--config` flag to specify a different path.

```json
{
  "provider": "Gemini",
  "model-name": "gemini-1.5-flash",
  "temperature": 0.7,
  "max-tokens": 1024,
  "api-key": "your_api_key_here",
  "max-iterations": 10
}
```

Note: If `api-key` is not provided, the agent will look for environment variables: `GEMINI_API_KEY`, `ANTHROPIC_API_KEY`, or `OPENAI_API_KEY` based on the provider.

### Environment Variables

If you prefer not to include the API key in the config file, you can set it as an environment variable.

Example for Gemini:

```bash
export GEMINI_API_KEY=your_key_here
```

## How to Use

Run the `agent-chat` command with a request message.

### Flags

- `-c`, `--config`: Path to the configuration file (default: `config.json`).
- `-m`, `--request`: The request/question for the AI agent.

### Example

```bash
go run . agent-chat -m "How do I view running containers?"
```

## CLI commands (current)

- `agent-chat` (`ac`): Ask Docker-related questions via LLM (Gemini example wired).
- `containerize` (`c`): Planned – generate a Dockerfile for the current directory.
- `diagnose` (`d`): Planned – analyze container logs for failures.
- `optimize` (`o`): Planned – suggest improvements for Dockerfiles.

## Project layout

```
cmd/                 Cobra CLI commands
internal/core/       Agent loop, prompts, Genkit client, tool registry
internal/docker/     Docker SDK wrapper + exec helpers
internal/models/     Shared data models
internal/tools/      Tool definitions (e.g., Docker commands)
```

## Contributing

Issues and PRs are welcome. Please open an issue to discuss larger changes before submitting a PR.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE).
