# Introduction

**Docker AI Agent** (`docker-ai`) is a Go-based AI agent that helps you inspect and operate your local Docker environment through an interactive terminal UI without the need for docker desktop.
Instead of memorizing Docker commands or digging through `docker inspect` output, you simply describe what you want in plain language. The agent understands the current state of your Docker setup, including containers, images, volumes, and networks, and it can execute real Docker commands on your behalf inside a controlled loop.

The project combines two things that work well together: a thin wrapper around the official Docker SDK for reading environment state,
and an LLM-driven planning loop built on Firebase Genkit. The model proposes the next action,
the proposed command is checked for safety,
and then it is executed through the local `docker` CLI. Results are fed back to the model so it can reason over them and decide what to do next,
repeating until your goal is complete.

## How it works

When you start the agent, it connects to your local Docker daemon and captures a snapshot of the environment: running and stopped containers, images with their tags and sizes, volumes, and networks. This snapshot is embedded into the system prompt so the model always has accurate context about your machine.

Each message you send goes through a short pipeline:

1. An LLM-based classifier decides whether your message is a general question, an action request, or something ambiguous.
2. General questions are answered directly by the model without touching Docker at all.
3. Action requests enter an iterative plan and execute loop. On every iteration the model explains its thought, calls a tool (currently the `docker_command_tool`, which runs any Docker CLI command), receives the output, and continues until it considers the goal done.
4. Thoughts, progress updates, and the final answer are streamed back into the terminal chat while the agent works.

## Features

- **Terminal chat interface** built with the Charm stack (Bubble Tea), designed for interactive back and forth conversation.
- **Docker awareness**: containers, images, volumes, and networks are snapshotted via the Docker SDK and injected into every prompt.
- **Real tool execution**: the agent does not just suggest commands, it runs them locally and reads their output.
- **Safety checks**: destructive commands such as `rm`, `prune`, or `stop` trigger a confirmation warning in the UI before anything executes.
- **Pluggable LLM providers** through Genkit: Google Gemini, OpenAI, Anthropic, Ollama (local), DeepSeek, Kimi, Qwen, and Grok are supported out of the box.
- **Session memory**: previous goals and executed commands are stored and included in later prompts, giving the agent continuity across messages.

## Project status

The project is in active development. The core chat experience is functional today:

| Command             | Status  |
| ------------------- | ------- |
| `agent-chat` (`ac`) | Working |
| `initialize-rag`    | Planned |
| `enable-rag`        | Planned |

The planned RAG commands will extend the agent with retrieval augmented generation, letting it ground its answers in indexed documentation and project-specific knowledge.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](https://github.com/your-org/docker-ai-agent/blob/main/LICENSE).
