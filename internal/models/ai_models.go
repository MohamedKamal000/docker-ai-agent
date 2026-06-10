package models

type ChatInteraction struct {
	UserRequest        string            `json:"user_request"`
	LLMThoughts        []string          `json:"llm_thoughts"`
	FinalResponse      string            `json:"final_response,omitempty"`
	ToolsExecuted      map[string]string `json:"tools_executed,omitempty"`
	UnstructuredOutput []string          `json:"unstructured_output"`
	IsStructured       bool              `json:"is_structured"`
}

type AgentExecutionStep struct {
	Thought       string `json:"thought,omitempty" description:"the ai thought about the next action"`
	FinalResponse string `json:"finalResponse,omitempty" description:"ai final response after finishing execution"`
	Done          bool   `json:"done,omitempty" description:"a flag for the ai to set that he finished the execution"`
}

type AgentResult struct {
	Structured   *AgentExecutionStep `json:"structured,omitempty"`
	Raw          string              `json:"raw,omitempty"`
	IsStructured bool                `json:"is_structured"`
}

type UserInputPrompt struct {
	Goal                string            `json:"goal"`
	CurrentGoalProgress []AgentResult     `json:"current-goal-progress"`
	PreviousChat        []ChatInteraction `json:"previous-chat"`
	ToolsExecuted       map[string]string `json:"tools-executed"`
}
