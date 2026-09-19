package tools

import (
	"testing"

	"docker-cli/internal/core"
)

func TestNewDockerCommandsTool(t *testing.T) {
	tasks := core.NewTaskRegistry()
	tool := NewDockerCommandsTool(tasks)

	if tool == nil {
		t.Fatal("NewDockerCommandsTool() returned nil")
	}

	if tool.Name() != "docker_command_tool" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "docker_command_tool")
	}

	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}

	schema := tool.GetInputSchema()
	if schema == nil {
		t.Fatal("GetInputSchema() returned nil")
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("schema properties not a map")
	}

	cmdProp, ok := props["command"].(map[string]any)
	if !ok {
		t.Fatal("command property not found or not a map")
	}

	if cmdProp["type"] != "string" {
		t.Errorf("command type = %v, want string", cmdProp["type"])
	}
}

func TestDockerCommandsTool_ShouldRaiseWarning(t *testing.T) {
	tasks := core.NewTaskRegistry()

	tool := NewDockerCommandsTool(tasks)

	tests := []struct {
		name        string
		input       any
		expectWarn  bool
		expectedMsg string
	}{
		{
			name:        "Destructive rm command",
			input:       map[string]any{"command": "docker rm container1"},
			expectWarn:  true,
			expectedMsg: "destructive command",
		},
		{
			name:        "Destructive rmi command",
			input:       map[string]any{"command": "docker rmi image1"},
			expectWarn:  true,
			expectedMsg: "destructive command",
		},
		{
			name:        "Destructive prune command",
			input:       map[string]any{"command": "docker system prune"},
			expectWarn:  true,
			expectedMsg: "destructive command",
		},
		{
			name:        "Non-destructive ps command",
			input:       map[string]any{"command": "docker ps"},
			expectWarn:  false,
			expectedMsg: "",
		},
		{
			name:        "Non-destructive images command",
			input:       map[string]any{"command": "docker images"},
			expectWarn:  false,
			expectedMsg: "",
		},
		{
			name:        "Invalid input (non-map)",
			input:       "not a map",
			expectWarn:  true,
			expectedMsg: "unknown input",
		},
		{
			name:        "Invalid input (nil)",
			input:       nil,
			expectWarn:  true,
			expectedMsg: "unknown input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warn, shouldWarn := tool.ShouldRaiseWarning(tt.input)
			if shouldWarn != tt.expectWarn {
				t.Errorf("ShouldRaiseWarning(%v) = (%q, %v), want (_, %v)", tt.input, warn, shouldWarn, tt.expectWarn)
			}
			if tt.expectWarn && !containsString(warn, tt.expectedMsg) {
				t.Errorf("ShouldRaiseWarning() warning = %q, want to contain %q", warn, tt.expectedMsg)
			}
		})
	}
}

func TestDockerCommandsTool_GetInputSchema(t *testing.T) {
	tasks := core.NewTaskRegistry()
	tool := NewDockerCommandsTool(tasks)

	schema := tool.GetInputSchema()

	// Check required fields
	requiredRaw, ok := schema["required"]
	if !ok {
		t.Fatal("required field missing")
	}
	required, ok := requiredRaw.([]any)
	if !ok {
		// Try as []string
		requiredStr, ok := requiredRaw.([]string)
		if !ok {
			t.Fatal("required field not array")
		}
		if len(requiredStr) != 1 || requiredStr[0] != "command" {
			t.Errorf("required = %v, want [command]", requiredStr)
		}
	} else {
		if len(required) != 1 || required[0] != "command" {
			t.Errorf("required = %v, want [command]", required)
		}
	}

	// Check additionalProperties is false
	additional, ok := schema["additionalProperties"].(bool)
	if !ok || additional {
		t.Error("additionalProperties should be false")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}