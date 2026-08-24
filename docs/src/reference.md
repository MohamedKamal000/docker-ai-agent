# Reference

This page is a lookup reference for everything configurable: the full `config.json` schema with its validation rules, and a comparison of the supported LLM providers.

## Configuration schema

The config file is loaded by `ModelConfigFromJsonFile` in `internal/core/config.go`.

| Field | Type | Required | Description |
|---|---|---|---|
| `provider` | string | yes | LLM provider: `Gemini`, `OpenAi`, or `Anthropic`. Values are case sensitive |
| `model-name` | string | yes | Provider specific model ID, for example `googleai/gemini-2.5-flash-lite` |
| `api-key` | string | no | API key. When empty, the agent falls back to the provider's environment variable |
| `temperature` | float32 | no | Sampling temperature, for example `0.7` |
| `max-tokens` | uint32 | no | Maximum tokens generated per model response |
| `max-iterations` | int | no | Cap on plan and execute loop iterations per user request |

### Example

```json
{
  "provider": "Anthropic",
  "model-name": "claude-sonnet-4-5",
  "temperature": 0.5,
  "max-tokens": 2048,
  "max-iterations": 15
}
```

### Validation errors

The loader refuses to start under the following conditions:

| Condition | Error message |
|---|---|
| Empty `model-name` | `model name can't be empty` |
| Unrecognized `provider` value | `provider <x> is not supported` |
| Negative `max-iterations` | `max iterations can't be negative` |
| No key found in file or environment | `api key not found` |
| Missing file and no fallback present | `config file not found` |

### File resolution

The config path resolves in this order:

1. The explicit path passed via `-c` / `--config`, if that file exists.
2. Otherwise `config.json` in the current working directory as a fallback.

### API key resolution order

1. The `api-key` field in the JSON file.
2. The environment variable matching the configured provider:

| Provider | Environment variable |
|---|---|
| `Gemini` | `GEMINI_API_KEY` |
| `OpenAi` | `OPENAI_API_KEY` |
| `Anthropic` | `ANTHROPIC_API_KEY` |

## LLM providers

The agent uses Firebase Genkit to talk to three providers. Configure one in `config.json` and supply its key; you never need more than one at a time.

### Google Gemini (default)

- Provider value: `"Gemini"`
- Key variable: `GEMINI_API_KEY`
- Model name format: `googleai/<model>`, for example `googleai/gemini-2.5-flash-lite`
- Where to get a key: [Google AI Studio](https://aistudio.google.com/apikey)
- Notes: the free tier is sufficient for flash models, which makes Gemini the cheapest option for trying out the project.

### OpenAI

- Provider value: `"OpenAi"`
- Key variable: `OPENAI_API_KEY`
- Model name format: plain model IDs such as `gpt-4o` or `gpt-4o-mini`
- Where to get a key: [OpenAI platform](https://platform.openai.com/api-keys)
- Notes: a paid API. Make sure your account has billing credit before running the agent against it.

### Anthropic

- Provider value: `"Anthropic"`
- Key variable: `ANTHROPIC_API_KEY`
- Model name format: plain model IDs such as `claude-sonnet-4-5`
- Where to get a key: [Anthropic Console](https://console.anthropic.com/settings/keys)
- Notes: paid API with a console based usage dashboard.

### Switching providers

Change `provider` and `model-name` together and export the matching environment variable:

```json
{
  "provider": "OpenAi",
  "model-name": "gpt-4o-mini",
  "max-iterations": 10
}
```

```bash
export OPENAI_API_KEY="sk-..."
go run . agent-chat
```

### Choosing a model

A single user request triggers several LLM calls: one for intent classification plus one or more loop iterations. Lighter and faster models such as Gemini Flash or GPT-4o mini keep total latency low, which matters because each iteration waits on a network round trip. If you pick a model with smaller output limits, raise `max-tokens`; if you routinely ask for complex multi step tasks, raise `max-iterations` so the agent has room to finish.
