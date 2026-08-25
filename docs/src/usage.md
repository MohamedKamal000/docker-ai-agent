# Usage

This page explains what the agent can do for you day to day, how a request travels through the system, and how the built in safety mechanisms protect your environment. Reading it will help you write better prompts and understand what the agent is doing at each moment.

## The chat interface

Running `agent-chat` opens a Bubble Tea terminal UI where you converse with the agent in plain language. There are no special command syntaxes to learn; you type messages the way you would talk to a colleague.

Typical things to ask:

- Questions about the current state of your machine, such as *"what containers am I running?"* or *"how much space do my images take up?"*
- Actions you want performed, such as *"start an nginx container on port 8080"* or *"restart the redis container"*
- General Docker knowledge questions, such as *"what is the difference between CMD and ENTRYPOINT?"*

### Session flow

1. You type a message and send it.
2. The agent runner starts the agent loop in a background goroutine so the UI stays responsive. Only one run can be active at a time.
3. While the agent works, the chat shows its current thought so you can follow the reasoning live.
4. When it finishes, the final response is rendered into the conversation.

### Canceling a run

While the agent is working you can press `esc` to cancel the request. The run aborts cleanly, a *"Request canceled."* message appears in the conversation, and you get the input back immediately.

### Session states

The chat session moves through a small set of states defined under `tui/screens/chat_session/`:

| State | What you see |
|---|---|
| Normal | Idle chat where you can type a new message |
| Agent running | The agent is thinking or executing tools, with thoughts streaming in |
| Warning | A destructive Docker command was proposed and awaits your approval or rejection |
| Options | Selection lists used for confirmations and choices |

The warning state deserves attention: when the agent proposes anything destructive, nothing executes until you explicitly confirm. This flow is described further below.

## How the agent works

Understanding the pipeline helps you predict behavior and phrase requests that the agent handles well.

### Step 1: Docker context snapshot

On startup the agent connects to the Docker daemon and captures a structured snapshot of the environment through `docker.GetContext`. It records containers with their names, images, states, and port mappings; images with tags and sizes; volumes with drivers; and networks with drivers and scopes. IDs are trimmed to twelve characters to keep prompts compact.

This snapshot is rendered into text by `FormatContextPrompt` and embedded into the system prompt. As a result, the model never has to guess what exists on your machine; it can see your containers and images from the very first token.

### Step 2: Intent classification

Each incoming message first passes through an LLM based classifier that assigns one of three labels:

| Intent | Behavior |
|---|---|
| `general_question` | Answered directly by the model. No tools run, so pure knowledge questions stay fast and side effect free |
| `action_request` | Enters the plan and execute loop described next |
| `ambiguous` | The agent stops and asks whether you wanted information or an action, rather than guessing |

The classifier also rewrites vague phrasing into a clearer instruction before the loop starts, which improves reliability on casually worded requests.

### Step 3: Plan and execute loop

Action requests drive an iterative loop bounded by `max-iterations`, with a short delay inserted between calls to be gentle on provider rate limits.

On every iteration the model receives your goal, the progress made so far, previous chat history loaded from memory, and the outputs of any commands already executed. It must answer with strict JSON of this shape:

```json
{
  "thought": "why I am doing this next",
  "finalResponse": "the answer for the user once done",
  "done": false
}
```

When `done` is `false`, the model also requests a tool call. Today there is one tool available, `docker_command_tool`, which accepts any valid Docker CLI command. The command runs locally, and its stdout, stderr, exit code, and duration come back as structured JSON that gets appended to the loop context. Reasoning then continues from that new evidence.

When `done` becomes `true`, the final response is delivered to the chat and the completed goal plus all executed commands are saved to memory.

Response parsing is forgiving: if the model wraps its JSON in code fences or adds prose around it, regular expressions extract the JSON object before decoding.

### Error handling

Provider hiccups are handled gracefully. On a 503 response the agent tells you it is retrying after five seconds instead of failing outright. On a 429 rate limit response the run stops with a clear message because backing off automatically could burn through your quota. If the iteration limit is reached without completion, the loop halts rather than spinning forever.

### Memory

A `MemoryStore` (in memory implementation today) keeps completed goals, executed commands, and intermediate steps. Future prompts include this history labeled as previous chat, which gives the agent continuity: after helping you start nginx, a later message like *"stop it again"* still makes sense.

## Safety and destructive commands

Because the agent can run arbitrary Docker CLI commands, destructive operations get special treatment.

### How detection works

Before any tool call executes, the proposed command string passes through `docker.IsDestructive`. That function lowercases the command and checks whether it contains any of these verbs as an argument:

```go
var DestructiveVerbs = []string{
    "rm", "rmi", "prune", "stop", "kill", "down", "restart", "pause", "remove",
}
```

Commands such as `docker rm my-container` or `docker system prune` match and raise a warning. Safe read only commands like `docker ps` pass straight through without interrupting you.

### What happens on a match

The TUI switches to the warning state and displays the exact offending command, for example:

```
docker command "docker rm nginx" contains a destructive command
```

You must confirm before execution proceeds. If you reject, the command does not run and the agent is informed that you declined, letting it explain itself or propose an alternative.

### Limitations

Detection is string based, which means it catches known destructive verbs but not every conceivable harmful operation. Commands that cause damage in unusual ways may not be flagged, and heavy but non destructive actions such as starting resource hungry containers execute without confirmation. Keep in mind also that the agent operates with whatever privileges your Docker daemon has, which on most Linux setups is root equivalent on the host.

### Best practices

- Read every warning carefully before confirming. `rm`, `prune`, and `down` can remove data that is difficult or impossible to recover.
- Avoid pointing the agent at production Docker hosts until you have built trust in its behavior.
- Prefer named containers and volumes you recognize over ambiguous short IDs when approving destructive actions.
