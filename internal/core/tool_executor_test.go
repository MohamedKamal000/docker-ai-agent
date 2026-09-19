package core

import (
	"context"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type mockToolExecutor struct {
	name        string
	description string
	inputSchema map[string]any
	callResult  string
	callError   error
	shouldWarn  bool
	warnMsg     string
}

func (m *mockToolExecutor) Name() string {
	return m.name
}

func (m *mockToolExecutor) Description() string {
	return m.description
}

func (m *mockToolExecutor) Call(ctx context.Context, input any) (string, error) {
	return m.callResult, m.callError
}

func (m *mockToolExecutor) GetInputSchema() map[string]any {
	return m.inputSchema
}

func (m *mockToolExecutor) ShouldRaiseWarning(input any) (string, bool) {
	if m.shouldWarn {
		return m.warnMsg, true
	}
	return "", false
}

func TestNewToolExecutor(t *testing.T) {
	registry := NewGenkitToolRegistry()
	executor := NewToolExecutor(registry)

	if executor == nil {
		t.Fatal("NewToolExecutor() returned nil")
	}
}

func createModelResponseWithToolRequest(toolName string, input map[string]any) *ai.ModelResponse {
	return &ai.ModelResponse{
		Message: &ai.Message{
			Content: []*ai.Part{
				ai.NewToolRequestPart(&ai.ToolRequest{
					Name:  toolName,
					Input: input,
				}),
			},
		},
	}
}

func TestToolExecutor_ExecuteGenkitTool(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool := &mockToolExecutor{
		name:        "test_tool",
		description: "Test tool",
		inputSchema: map[string]any{"type": "object"},
		callResult:  "tool output",
	}
	registry.Register(tool, g)

	executor := NewToolExecutor(registry)

	resp := createModelResponseWithToolRequest("test_tool", map[string]any{
		"command": "test command",
	})

	comm := &AgentCommunication{
		ToUser:   make(chan AiResponse, 10),
		FromUser: make(chan UserCommand, 10),
	}

	output, err := executor.ExecuteGenkitTool(ctx, resp, comm)
	if err != nil {
		t.Fatalf("ExecuteGenkitTool() returned error: %v", err)
	}

	if output == nil {
		t.Fatal("ExecuteGenkitTool() returned nil output")
	}

	result, ok := output["test_tool"]
	if !ok {
		t.Fatal("Output missing test_tool result")
	}

	if result != "tool output" {
		t.Errorf("Result = %q, want %q", result, "tool output")
	}
}

func TestToolExecutor_ExecuteGenkitTool_NotFound(t *testing.T) {
	ctx := context.Background()
	registry := NewGenkitToolRegistry()
	executor := NewToolExecutor(registry)

	resp := createModelResponseWithToolRequest("nonexistent", map[string]any{})

	comm := &AgentCommunication{
		ToUser:   make(chan AiResponse, 10),
		FromUser: make(chan UserCommand, 10),
	}

	_, err := executor.ExecuteGenkitTool(ctx, resp, comm)
	if err == nil {
		t.Fatal("ExecuteGenkitTool() expected error for nonexistent tool")
	}
	if err.Error() != "tool nonexistent not found" {
		t.Errorf("Error = %q, want %q", err.Error(), "tool nonexistent not found")
	}
}

func TestToolExecutor_ExecuteGenkitTool_WithWarning(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool := &mockToolExecutor{
		name:        "warning_tool",
		description: "Tool with warning",
		inputSchema: map[string]any{"type": "object"},
		callResult:  "tool output",
		shouldWarn:  true,
		warnMsg:     "This is a warning",
	}
	registry.Register(tool, g)

	executor := NewToolExecutor(registry)

	resp := createModelResponseWithToolRequest("warning_tool", map[string]any{})

	comm := &AgentCommunication{
		ToUser:   make(chan AiResponse, 10),
		FromUser: make(chan UserCommand, 10),
	}

	go func() {
		comm.FromUser <- UserCommand{ShouldContinue: true}
	}()

	output, err := executor.ExecuteGenkitTool(ctx, resp, comm)
	if err != nil {
		t.Fatalf("ExecuteGenkitTool() returned error: %v", err)
	}

	if output == nil {
		t.Fatal("ExecuteGenkitTool() returned nil output")
	}

	result, ok := output["warning_tool"]
	if !ok {
		t.Fatal("Output missing warning_tool result")
	}

	if result != "tool output" {
		t.Errorf("Result = %q, want %q", result, "tool output")
	}
}

func TestToolExecutor_ExecuteGenkitTool_WarningCancelled(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool := &mockToolExecutor{
		name:        "warning_tool",
		description: "Tool with warning",
		inputSchema: map[string]any{"type": "object"},
		callResult:  "tool output",
		shouldWarn:  true,
		warnMsg:     "This is a warning",
	}
	registry.Register(tool, g)

	executor := NewToolExecutor(registry)

	resp := createModelResponseWithToolRequest("warning_tool", map[string]any{})

	comm := &AgentCommunication{
		ToUser:   make(chan AiResponse, 10),
		FromUser: make(chan UserCommand, 10),
	}

	go func() {
		comm.FromUser <- UserCommand{ShouldContinue: false}
	}()

	_, err := executor.ExecuteGenkitTool(ctx, resp, comm)
	if err == nil {
		t.Fatal("ExecuteGenkitTool() expected error when cancelled")
	}
	if err.Error() != "execution is canceled" {
		t.Errorf("Error = %q, want %q", err.Error(), "execution is canceled")
	}
}

func TestToolExecutor_ExecuteGenkitTool_ToolError(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool := &mockToolExecutor{
		name:        "error_tool",
		description: "Tool that returns error",
		inputSchema: map[string]any{"type": "object"},
		callResult:  "",
		callError:   &testErrorExecutor{msg: "tool failed"},
	}
	registry.Register(tool, g)

	executor := NewToolExecutor(registry)

	resp := createModelResponseWithToolRequest("error_tool", map[string]any{})

	comm := &AgentCommunication{
		ToUser:   make(chan AiResponse, 10),
		FromUser: make(chan UserCommand, 10),
	}

	_, err := executor.ExecuteGenkitTool(ctx, resp, comm)
	if err == nil {
		t.Fatal("ExecuteGenkitTool() expected error from tool")
	}
	if err.Error() != "tool failed" {
		t.Errorf("Error = %q, want %q", err.Error(), "tool failed")
	}
}

type testErrorExecutor struct {
	msg string
}

func (e *testErrorExecutor) Error() string {
	return e.msg
}

func TestExtractCommandFromInput(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "Valid map with command",
			input:    map[string]any{"command": "docker ps"},
			expected: "docker ps",
		},
		{
			name:     "Valid map with command and other fields",
			input:    map[string]any{"command": "docker run nginx", "other": "value"},
			expected: "docker run nginx",
		},
		{
			name:     "Map without command key",
			input:    map[string]any{"other": "value"},
			expected: "unknown",
		},
		{
			name:     "Non-map input",
			input:    "not a map",
			expected: "unknown",
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: "unknown",
		},
		{
			name:     "Command is not string",
			input:    map[string]any{"command": 123},
			expected: "unknown",
		},
		{
			name:     "Empty map",
			input:    map[string]any{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCommandFromInput(tt.input)
			if result != tt.expected {
				t.Errorf("extractCommandFromInput(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToolExecutor_ExecuteGenkitTool_MultipleTools(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool1 := &mockToolExecutor{
		name:        "tool1",
		description: "First tool",
		inputSchema: map[string]any{"type": "object"},
		callResult:  "output1",
	}
	tool2 := &mockToolExecutor{
		name:        "tool2",
		description: "Second tool",
		inputSchema: map[string]any{"type": "object"},
		callResult:  "output2",
	}
	registry.Register(tool1, g)
	registry.Register(tool2, g)

	executor := NewToolExecutor(registry)

	resp := &ai.ModelResponse{
		Message: &ai.Message{
			Content: []*ai.Part{
				ai.NewToolRequestPart(&ai.ToolRequest{
					Name:  "tool1",
					Input: map[string]any{"command": "cmd1"},
				}),
				ai.NewToolRequestPart(&ai.ToolRequest{
					Name:  "tool2",
					Input: map[string]any{"command": "cmd2"},
				}),
			},
		},
	}

	comm := &AgentCommunication{
		ToUser:   make(chan AiResponse, 10),
		FromUser: make(chan UserCommand, 10),
	}

	output, err := executor.ExecuteGenkitTool(ctx, resp, comm)
	if err != nil {
		t.Fatalf("ExecuteGenkitTool() returned error: %v", err)
	}

	if len(output) != 2 {
		t.Errorf("ExecuteGenkitTool() returned %d results, want 2", len(output))
	}

	if output["tool1"] != "output1" {
		t.Errorf("tool1 result = %q, want %q", output["tool1"], "output1")
	}
	if output["tool2"] != "output2" {
		t.Errorf("tool2 result = %q, want %q", output["tool2"], "output2")
	}
}

func TestToolExecutor_ExecuteGenkitTool_NoToolRequests(t *testing.T) {
	ctx := context.Background()
	registry := NewGenkitToolRegistry()
	executor := NewToolExecutor(registry)

	resp := &ai.ModelResponse{
		Message: &ai.Message{
			Content: []*ai.Part{
				{Kind: ai.PartText, Text: "Just text response"},
			},
		},
	}

	comm := &AgentCommunication{
		ToUser:   make(chan AiResponse, 10),
		FromUser: make(chan UserCommand, 10),
	}

	output, err := executor.ExecuteGenkitTool(ctx, resp, comm)
	if err != nil {
		t.Fatalf("ExecuteGenkitTool() returned error: %v", err)
	}

	if output == nil {
		t.Fatal("ExecuteGenkitTool() returned nil output")
	}

	if len(output) != 0 {
		t.Errorf("ExecuteGenkitTool() returned %d results, want 0", len(output))
	}
}