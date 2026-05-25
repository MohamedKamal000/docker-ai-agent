package models

type LLMProvider int

const (
	Gemini LLMProvider = iota

	OpenAi

	Anthropic
)

type SystemPromptInput struct {
	Containers []Container `json:"Containers"`
}

func NewSystemPromptInput(containers []Container) *SystemPromptInput {
	return &SystemPromptInput{Containers: containers}
}

type ModelConfig struct {
	Provider    LLMProvider `json:"provider"`
	ModelName   string      `json:"model-model-name"`
	Temperature float32     `json:"temperature"`
	MaxTokens   uint32      `json:"max-tokens"`
	ApiKey      string      `json:"api-key"`
}

func NewModelConfig(provider LLMProvider, modelName string, temperature float32, maxTokens uint32, apiKey string) *ModelConfig {
	return &ModelConfig{Provider: provider, ModelName: modelName, Temperature: temperature, MaxTokens: maxTokens, ApiKey: apiKey}
}

type ContextAction struct {
	Thought string `json:"thought" description:"the ai thought about the next action"`
	Action  string `json:"action" description:"the final action the ai going to take"`
	Done    bool
}

// genkit tool parsing, we will parse the ai request for the tool, run it after confirming
// then take the result and make one of these structs and append it to the context list
type ContextStep struct {
	TakenAction ContextAction `json:"action"` // ai action that he will do next, this is the ai output without the tool request
	// add a tool call type here with tool name and inputs, outputs (the result of the context step can be the tool output)
	Result string `json:"result"`
}

func NewContextStep(action ContextAction, result string) *ContextStep {
	return &ContextStep{TakenAction: action, Result: result}
}

type UserInputPrompt struct {
	Goal    string        `json:"goal"`
	Context []ContextStep `json:"context"`
}
