# Architecture

This page describes how the project is put together: the major components, how data flows through them, and what each package is responsible for. It is written for contributors who want to modify the agent or add new capabilities.

## Component map

<a href="images/project-arch.png" target="_blank">
  <img src="images/project-arch.png" alt="Architecture diagram">
</a>

## Data flow for a user message

1. The TUI root model spawns the agent loop in a goroutine, passing the user message over the `AgentCommunication` channels defined in `internal/core/agent_communication.go`.
2. Earlier, at startup, the agent captured a Docker context snapshot through the SDK. The snapshot is formatted into text and embedded in the system prompt.
3. The intent classifier labels the message as a question, an action request, or ambiguous.
4. Action requests run the plan and execute loop. Each iteration may invoke `docker_command_tool` through the tool executor.
5. Tools shell out to the local `docker` binary. Destructive commands route back through the TUI for confirmation before running.
6. Thoughts, retry notices, warnings, and final responses stream to the TUI over the channel, which closes when the run completes.

## Package layout

```
main.go              Entry point, calls cmd.Execute()
cmd/                 Cobra commands: agent-chat plus planned initialize-rag / enable-rag
internal/app         Agent assembly and a mock agent loop for testing
internal/core        Config, Genkit client, agent loop, classifier, prompts,
                     tool registry and executor, memory, communication channels
internal/docker      Docker SDK client, exec helpers, context formatting
internal/models      Shared models: agent steps and results, docker summaries
internal/tools       Concrete tool implementations
tui/                 Bubble Tea root model, screens, common styles, widgets
```

The module itself is named `docker-cli` (visible in `go.mod`) even though the repository is called docker-ai-agent.

## Agent core (`internal/core`)

### Genkit client

`NewGenkitClient` initializes Firebase Genkit with the provider plugin selected by configuration:

| Provider    | Plugin                     |
| ----------- | -------------------------- |
| `Gemini`    | `googlegenai.GoogleAI`     |
| `OpenAi`    | `compat_oai/openai.OpenAI` |
| `Anthropic` | `anthropic.Anthropic`      |

The configured `model-name` becomes Genkit's default model so every subsequent generate call uses it without repetition.

### Configuration

`ModelConfigFromJsonFile` loads and validates the JSON config into a `ModelConfig` struct covering provider, model name, temperature, max tokens, API key, and iteration cap. When no key is present in the file, it falls back to provider specific environment variables. Validation rules are documented on the [Reference](./reference.md) page.

### Intent classifier

A dedicated LLM call labels each user message as `general_question`, `action_request`, or `ambiguous`. Ambiguous messages short circuit with a clarification question so the agent never guesses at destructive intent. General questions bypass tools entirely and get answered with a system prompt that positions the model as a concise Docker expert.

### Prompts

System prompts live in `prompts.go` as Go `text/template` templates. The template embeds the Docker environment snapshot produced by `docker.FormatContextPrompt`, along with the current goal, accumulated progress, previous chat history, and results of already executed commands. The instructions require strict JSON output of `{thought, finalResponse, done}` plus any tool calls needed for the next step.

### Agent loop

`GenkitAgentLoop.Run` orchestrates everything:

1. Classify intent and dispatch accordingly.
2. For action requests, iterate up to `MaxIterations`. Each pass runs the flow, extracts structured JSON from the model response using regex fallbacks for fenced or loosely wrapped output, executes requested tools via the `ToolExecutor`, appends their outputs to progress, and repeats until the model signals `done`.
3. Handle transport errors: a 503 triggers a retry after five seconds with a notice to the user; a 429 aborts with a rate limit error; other errors surface immediately.
4. Persist the completed goal, executed commands, and intermediate steps into memory once finished.

### Tool registry and executor

Tools implement a small interface: name, description, JSON input schema, a `Call` method, and an optional `ShouldRaiseWarning` guard. They are registered with Genkit so the model can discover and invoke them by name. The executor parses model output, routes any warnings through the communication channel first, and only then executes.

### Memory

The `MemoryStore` interface currently has an in memory implementation. It records previous goals, tool outputs, and intermediate reasoning steps, all of which are injected into later prompts as conversation history.

### Communication

`agent_communication.go` defines typed messages flowing between the loop goroutine and the UI over two channels. Thoughts, retry notices, warnings requiring confirmation, and final responses travel on `ToUser`; confirmations return on `FromUser`. The loop closes its channel when done, which returns the UI to the idle state.

### Mock agent

`MockAgentLoop` in `internal/app/mock_agent.go` implements the same `AgentLoop` interface with scriptable thought, warning, and final steps. It lets you exercise the full TUI without an API key or a running Docker daemon.

## Docker integration (`internal/docker`)

### Client lifecycle

`Init` runs exactly once thanks to `sync.Once`. It builds an SDK client from environment variables (so standard `DOCKER_HOST` and TLS settings work), enables API version negotiation, pings the daemon so startup fails fast when Docker is unreachable, and locates the `docker` binary in `PATH` because CLI execution needs it.

### Context snapshots

`GetContext` collects container, image, volume, and network summaries plus a capture timestamp into a single `Context`. Summaries keep only essentials such as twelve character IDs and the first container name, which keeps prompt size small while remaining unambiguous. `FormatContextPrompt` renders this structure as aligned plain text sections that slot directly into the system prompt.

### CLI execution

- `Exec(ctx, command)` shell splits the command string, correctly handling quotes, escapes, and whitespace, strips a leading `docker` token if present, then runs it with `exec.CommandContext` and captures stdout, stderr, exit code, and duration into an `ExecResult`.
- `ExecMany(ctx, commands, continueOnError)` runs a batch sequentially, optionally stopping at the first failure.
- `ParseCommands(block)` turns multi line script blocks into individual command strings, skipping blanks and `#` comments.

### Safety

`IsDestructive` checks whether a command contains one of the known destructive verbs as an argument. Tools call it through `ShouldRaiseWarning`, which pauses execution until the user confirms in the TUI.

## Terminal UI (`tui`)

The interface is built on Bubble Tea v2 with Bubbles v2 components and Lipgloss v2 styling.

`tui.NewRootModel(agent)` constructs the root model from any `AgentLoop`. On each submitted message it spawns the loop in a background goroutine wired to the communication channels, transitions the chat session into the running state, and renders incoming thoughts and results as they arrive.

The chat session is a small state machine under `screens/chat_session/` with states for normal chatting, agent running, warning confirmation, and option selection. Because rendering is driven entirely by channel messages, the same UI works unchanged against the mock agent during development.

## Tools (`internal/tools`)

### Built-in: docker_command_tool

Currently the only registered tool. Its schema accepts one required string:

```json
{
  "type": "object",
  "properties": {
    "command": {
      "type": "string",
      "description": "A valid Docker CLI command to execute."
    }
  },
  "required": ["command"],
  "additionalProperties": false
}
```

Its description tells the model it can inspect, create, modify, start, stop, restart, and remove Docker resources, and that it should be used only when direct Docker interaction is required. On invocation the command executes through `docker.Exec` and the structured result returns as pretty printed JSON for the next loop iteration. `ShouldRaiseWarning` delegates to `docker.IsDestructive` so dangerous commands pause for confirmation.

### Adding a new tool

1. Create a struct under `internal/tools/` implementing the tool interface.
2. Give it a unique name and a precise description. The model relies heavily on this text when deciding whether and when to use the tool, so invest in wording it well.
3. Define an accurate JSON input schema. Setting `additionalProperties: false` keeps model output predictable.
4. Register the tool in the list passed to `app.NewAgent` in `cmd/agent-chat.go` so it becomes available to the loop.
5. If the tool can cause harm, implement `ShouldRaiseWarning` to gate execution behind explicit user confirmation.
