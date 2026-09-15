package models

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

type ToolCallInfo struct {
	ToolName string `json:"tool_name"`
	Command  string `json:"command"`
	Status   string `json:"status"`
	Result   string `json:"result,omitempty"`
}

type HistoryEntry struct {
	Run           int            `json:"run"`
	Goal          string         `json:"goal"`
	Thought       string         `json:"thought,omitempty"`
	ToolCalls     []ToolCallInfo `json:"tool_calls,omitempty"`
	FinalResponse string         `json:"final_response,omitempty"`
	Done          bool           `json:"done,omitempty"`
}

type UserInputPrompt struct {
	Goal    string         `json:"goal"`
	History []HistoryEntry `json:"history"`
}
