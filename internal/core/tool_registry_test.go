package core

import (
	"context"
	"testing"

	"github.com/firebase/genkit/go/genkit"
)

type mockTool struct {
	name        string
	description string
	inputSchema map[string]any
	callResult  string
	callError   error
	shouldWarn  bool
	warnMsg     string
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Call(ctx context.Context, input any) (string, error) {
	return m.callResult, m.callError
}

func (m *mockTool) GetInputSchema() map[string]any {
	return m.inputSchema
}

func (m *mockTool) ShouldRaiseWarning(input any) (string, bool) {
	if m.shouldWarn {
		return m.warnMsg, true
	}
	return "", false
}

func TestNewGenkitToolRegistry(t *testing.T) {
	registry := NewGenkitToolRegistry()

	if registry == nil {
		t.Fatal("NewGenkitToolRegistry() returned nil")
	}

	tools := registry.List()
	if len(tools) != 0 {
		t.Errorf("List() on new registry returned %d tools, want 0", len(tools))
	}
}

func TestGenkitToolRegistry_RegisterAndGet(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool := &mockTool{
		name:        "test_tool",
		description: "A test tool",
		inputSchema: map[string]any{"type": "object"},
		callResult:  "success",
	}

	registry.Register(tool, g)

	retrieved, ok := registry.Get("test_tool")
	if !ok {
		t.Fatal("Get() failed to find registered tool")
	}

	if retrieved.Name() != "test_tool" {
		t.Errorf("Retrieved tool Name() = %q, want %q", retrieved.Name(), "test_tool")
	}
	if retrieved.Description() != "A test tool" {
		t.Errorf("Retrieved tool Description() = %q, want %q", retrieved.Description(), "A test tool")
	}
}

func TestGenkitToolRegistry_GetNonExistent(t *testing.T) {
	registry := NewGenkitToolRegistry()

	_, ok := registry.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for nonexistent tool")
	}
}

func TestGenkitToolRegistry_List(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool1 := &mockTool{name: "tool1", description: "Tool 1", inputSchema: map[string]any{}}
	tool2 := &mockTool{name: "tool2", description: "Tool 2", inputSchema: map[string]any{}}

	registry.Register(tool1, g)
	registry.Register(tool2, g)

	tools := registry.List()
	if len(tools) != 2 {
		t.Errorf("List() returned %d tools, want 2", len(tools))
	}

	found1 := false
	found2 := false
	for _, t := range tools {
		if t.Name() == "tool1" {
			found1 = true
		}
		if t.Name() == "tool2" {
			found2 = true
		}
	}
	if !found1 {
		t.Error("List() missing tool1")
	}
	if !found2 {
		t.Error("List() missing tool2")
	}
}

func TestGenkitToolRegistry_RegisterMultiple(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	for i := 0; i < 5; i++ {
		tool := &mockTool{
			name:        "tool_" + string(rune('0'+i)),
			description: "Tool",
			inputSchema: map[string]any{},
		}
		registry.Register(tool, g)
	}

	tools := registry.List()
	if len(tools) != 5 {
		t.Errorf("List() returned %d tools, want 5", len(tools))
	}
}

func TestGenkitToolRegistry_RegisterOverwrites(t *testing.T) {
	ctx := context.Background()
	g := genkit.Init(ctx)
	registry := NewGenkitToolRegistry()

	tool1 := &mockTool{
		name:        "tool_1",
		description: "First",
		inputSchema: map[string]any{},
		callResult:  "first",
	}
	registry.Register(tool1, g)

	tool2 := &mockTool{
		name:        "tool_2",
		description: "Second",
		inputSchema: map[string]any{},
		callResult:  "second",
	}
	registry.Register(tool2, g)

	tools := registry.List()
	if len(tools) != 2 {
		t.Errorf("List() returned %d tools after two registers, want 2", len(tools))
	}
}