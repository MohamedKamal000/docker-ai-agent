# Getting Started

This page walks you through everything needed to go from a clean machine to a working chat session with the agent: installing the prerequisites, obtaining and storing an API key, writing a configuration file, and launching the program.

## Prerequisites

| Requirement | Notes |
|---|---|
| **Go 1.25+** | Needed to build and run from source |
| **Docker Engine** | Must be installed and running locally; the agent talks to the daemon through the Docker SDK |
| **`docker` binary in PATH** | Tool execution shells out to the local CLI, so it must be resolvable |
| **An LLM API key** | From Gemini, OpenAI, or Anthropic. You only need one of them |

Before anything else, confirm that Docker is reachable on your machine:

```bash
docker info
```

The agent pings the daemon during startup and exits immediately with an error if it cannot be reached, so it is worth checking now rather than later.

## Building from source

Clone the repository and compile the binary:

```bash
git clone https://github.com/your-org/docker-ai-agent.git
cd docker-ai-agent
go build -o docker-ai .
```

You can also skip building entirely and run directly with Go:

```bash
go run .
```

Verify that everything is wired up correctly:

```bash
./docker-ai --help
```

You should see the `docker-ai` help text along with the available commands, including `agent-chat`.

## API keys

The agent needs an API key from one of the supported providers. There are two ways to provide it, and you should pick exactly one:

1. An environment variable, which is the recommended option.
2. The `api-key` field inside `config.json`, which is quick but easier to leak by accident.

If the `api-key` field is absent or empty, the agent automatically falls back to the matching environment variable for whichever provider you configured.

### Getting a key

#### Google Gemini

1. Open [Google AI Studio](https://aistudio.google.com/apikey) and sign in with your Google account.
2. Click **Create API key** and copy the generated value.
3. The free tier is enough to run `gemini-2.5-flash-lite`, which makes Gemini the cheapest way to try the project.

Environment variable name: `GEMINI_API_KEY`

#### OpenAI

1. Open [platform.openai.com/api-keys](https://platform.openai.com/api-keys) and sign in.
2. Click **Create new secret key** and copy the value. It starts with `sk-` and is shown only once.
3. OpenAI is a paid API, so make sure your account has billing credit before using it.

Environment variable name: `OPENAI_API_KEY`

#### Anthropic

1. Open [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys) and sign in.
2. Click **Create Key** and copy the value. It starts with `sk-ant-`.

Environment variable name: `ANTHROPIC_API_KEY`

### Storing the key as an environment variable

For the current shell session only, export the variable before running the agent:

```bash
export GEMINI_API_KEY="your_key_here"
```

To make it persistent across sessions, add the same line to your shell profile (`~/.bashrc` if you use bash, `~/.zshrc` if you use zsh) and then reload it:

```bash
source ~/.bashrc
```

A common pattern is to keep all your local secrets in a dedicated `.env` file in the repository root and source it right before running:

```bash
# .env  (make sure this file is listed in .gitignore)
export GEMINI_API_KEY="your_key_here"
```

```bash
source .env && go run . agent-chat
```

Whichever approach you choose, double check that files containing real keys are covered by `.gitignore` so they can never be committed.

### Storing the key inside config.json

You can also place the key directly in the configuration file:

```json
{
  "provider": "Gemini",
  "model-name": "googleai/gemini-2.5-flash-lite",
  "api-key": "your_key_here"
}
```

This is convenient for local experiments but risky in practice. If you share the folder, copy the config between machines, or accidentally commit it, the key leaks. If you use this option, never commit a `config.json` containing a real key to version control.

### Security tips

- Never commit API keys to git. If a key has been exposed, rotate it immediately from the provider console.
- Prefer restricted or scoped keys where the provider supports them.
- Remember which provider maps to which variable: `Gemini` uses `GEMINI_API_KEY`, `OpenAi` uses `OPENAI_API_KEY`, and `Anthropic` uses `ANTHROPIC_API_KEY`. The resolution rules are described in detail on the [Reference](./reference.md) page.

## Configuration

The agent reads its settings from a JSON file. By default it looks for `config.json` in the current working directory, and you can point at any other file with the `-c` flag when running the command.

### Minimal example

```json
{
  "provider": "Gemini",
  "model-name": "googleai/gemini-2.5-flash-lite",
  "temperature": 0.7,
  "max-tokens": 1024,
  "max-iterations": 10
}
```

With this config and no `api-key` field present, the agent will read the key from `GEMINI_API_KEY` at startup.

### Fields explained

| Field | Type | Required | Description |
|---|---|---|---|
| `provider` | string | yes | Which LLM backend to use: `Gemini`, `OpenAi`, or `Anthropic` |
| `model-name` | string | yes | The provider specific model ID, for example `googleai/gemini-2.5-flash-lite` |
| `api-key` | string | no | API key used for authentication; falls back to the environment when omitted |
| `temperature` | float | no | Sampling temperature; higher values make output more creative, lower values more deterministic |
| `max-tokens` | int | no | Upper limit on tokens generated per model response |
| `max-iterations` | int | no | Maximum number of plan and execute loop iterations per user request |

The loader validates this data and refuses to start when something important is wrong. It rejects an unknown provider name, an empty model name, a negative iteration count, and a setup where no key could be found in either the file or the environment. The full list of validation errors lives on the [Reference](./reference.md) page.

## Running the agent

Launch the interactive chat with the main command or its short alias:

```bash
go run . agent-chat
# or
go run . ac
```

If you compiled the binary earlier, use it instead:

```bash
./docker-ai agent-chat
```

The `-c` / `--config` flag selects a config file at a custom location:

```bash
go run . agent-chat -c ./my-config.json
```

At runtime three things must hold: the Docker daemon has to be running because the agent pings it on startup, a valid API key must be resolvable from the config file or the environment, and the `docker` binary must exist in your `PATH` since tool execution shells out to it. If any of these fail, the command exits early with a descriptive error rather than misbehaving later.

The planned `initialize-rag` and `enable-rag` commands do not exist yet. Once implemented they will set up and activate retrieval augmented generation so the agent can ground its answers in indexed documentation.
