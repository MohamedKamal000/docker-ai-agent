package tools

import (
	"context"
	"encoding/json"
	"testing"

	"docker-cli/internal/core"
	"docker-cli/internal/models"
)

func TestNewTaskStatusTool(t *testing.T) {
	tasks := core.NewTaskRegistry()
	tool := NewTaskStatusTool(tasks)

	if tool == nil {
		t.Fatal("NewTaskStatusTool() returned nil")
	}

	if tool.Name() != "task_status_tool" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "task_status_tool")
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

	taskIDProp, ok := props["task_id"].(map[string]any)
	if !ok {
		t.Fatal("task_id property not found or not a map")
	}

	if taskIDProp["type"] != "string" {
		t.Errorf("task_id type = %v, want string", taskIDProp["type"])
	}

	additional, ok := schema["additionalProperties"].(bool)
	if !ok || additional {
		t.Error("additionalProperties should be false")
	}
}

func TestTaskStatusTool_Call_SingleTask(t *testing.T) {
	ctx := context.Background()
	tasks := core.NewTaskRegistry()

	// Register a task
	record := core.NewTaskRecord("docker_command_tool", "docker ps")
	tasks.Register(record)
	tasks.UpdateComplete(record.ID, core.TaskSucceeded, &models.ExecResult{
		Command:  "docker ps",
		Stdout:   "CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES",
		Stderr:   "",
		ExitCode: 0,
	}, "")

	tool := NewTaskStatusTool(tasks)

	result, err := tool.Call(ctx, map[string]any{
		"task_id": record.ID,
	})

	if err != nil {
		t.Fatalf("Call() returned error: %v", err)
	}

	if result == "" {
		t.Fatal("Call() returned empty result")
	}

	// Parse the JSON result
	var parsed core.TaskRecord
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	if parsed.ID != record.ID {
		t.Errorf("Parsed task ID = %q, want %q", parsed.ID, record.ID)
	}
	if parsed.Tool != "docker_command_tool" {
		t.Errorf("Parsed tool = %q, want docker_command_tool", parsed.Tool)
	}
	if parsed.Status != core.TaskSucceeded {
		t.Errorf("Parsed status = %v, want %v", parsed.Status, core.TaskSucceeded)
	}
}

func TestTaskStatusTool_Call_AllTasks(t *testing.T) {
	ctx := context.Background()
	tasks := core.NewTaskRegistry()

	// Register multiple tasks
	record1 := core.NewTaskRecord("docker_command_tool", "docker ps")
	tasks.Register(record1)

	record2 := core.NewTaskRecord("docker_command_tool", "docker images")
	tasks.Register(record2)

	tool := NewTaskStatusTool(tasks)

	result, err := tool.Call(ctx, map[string]any{}) // No task_id

	if err != nil {
		t.Fatalf("Call() returned error: %v", err)
	}

	if result == "" {
		t.Fatal("Call() returned empty result")
	}

	// Parse the JSON result
	var parsed []core.TaskRecord
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Failed to parse result JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("Parsed %d tasks, want 2", len(parsed))
	}
}

func TestTaskStatusTool_Call_TaskNotFound(t *testing.T) {
	ctx := context.Background()
	tasks := core.NewTaskRegistry()
	tool := NewTaskStatusTool(tasks)

	result, err := tool.Call(ctx, map[string]any{
		"task_id": "nonexistent-task-id",
	})

	if err != nil {
		t.Fatalf("Call() returned error: %v", err)
	}

	if result == "" {
		t.Fatal("Call() returned empty result")
	}

	expected := "task with ID \"nonexistent-task-id\" not found"
	if result != expected {
		t.Errorf("Call() result = %q, want %q", result, expected)
	}
}

func TestTaskStatusTool_Call_NilTaskRegistry(t *testing.T) {
	ctx := context.Background()
	tool := NewTaskStatusTool(nil)

	_, err := tool.Call(ctx, map[string]any{})
	if err == nil {
		t.Error("Call() expected error for nil task registry")
	}
	if err.Error() != "task registry is not initialized" {
		t.Errorf("Call() error = %q, want %q", err.Error(), "task registry is not initialized")
	}
}

func TestTaskStatusTool_ShouldRaiseWarning(t *testing.T) {
	tasks := core.NewTaskRegistry()
	tool := NewTaskStatusTool(tasks)

	warn, shouldWarn := tool.ShouldRaiseWarning(map[string]any{})
	if shouldWarn {
		t.Error("ShouldRaiseWarning() should return false for task_status_tool")
	}
	if warn != "" {
		t.Errorf("ShouldRaiseWarning() warning = %q, want empty", warn)
	}
}

func TestTaskStatusTool_GetInputSchema(t *testing.T) {
	tasks := core.NewTaskRegistry()
	tool := NewTaskStatusTool(tasks)

	schema := tool.GetInputSchema()

	if schema == nil {
		t.Fatal("GetInputSchema() returned nil")
	}

	// Check structure
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties missing")
	}

	if _, ok := props["task_id"]; !ok {
		t.Error("task_id property missing")
	}

	additional, ok := schema["additionalProperties"].(bool)
	if !ok || additional {
		t.Error("additionalProperties should be false")
	}
}