package core

type LLMProvider int

const (
	Gemini LLMProvider = iota

	OpenAi

	Anthropic
)

type ModelConfig struct {
	Provider    LLMProvider `json:"provider"`
	ModelName   string      `json:"model-model-name"`
	Temperature float32     `json:"temperature"`
	MaxTokens   uint32      `json:"max-tokens"`
	ApiKey      string      `json:"api-key"`
}
