package tools

import (
	"context"
	"docker-cli/internal/core"
	"encoding/json"
	"fmt"
)

type TaskStatusTool struct {
	InputSchema map[string]any
	Tasks       *core.TaskRegistry
}

func NewTaskStatusTool(tasks *core.TaskRegistry) *TaskStatusTool {
	return &TaskStatusTool{
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task_id": map[string]any{
					"type":        "string",
					"description": "Optional task ID to check status for (e.g., task-12345678). If omitted, returns all tasks.",
				},
			},
			"additionalProperties": false,
		},
		Tasks: tasks,
	}
}

func (t *TaskStatusTool) Name() string {
	return "task_status_tool"
}

func (t *TaskStatusTool) Description() string {
	return `Checks the status, logs, and docker_object_id of background tasks.
Provide a task_id to get detailed status for a specific task, or omit task_id to get a list of all recorded background tasks.`
}

func (t *TaskStatusTool) Call(ctx context.Context, input any) (string, error) {
	if t.Tasks == nil {
		return "", fmt.Errorf("task registry is not initialized")
	}

	taskID := ""
	if m, ok := input.(map[string]any); ok {
		if v, ok := m["task_id"].(string); ok {
			taskID = v
		}
	}

	if taskID != "" {
		rec, found := t.Tasks.Get(taskID)
		if !found {
			return fmt.Sprintf("task with ID %q not found", taskID), nil
		}
		b, err := json.MarshalIndent(rec, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	tasksList := t.Tasks.List()
	b, err := json.MarshalIndent(tasksList, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *TaskStatusTool) GetInputSchema() map[string]any {
	return t.InputSchema
}

func (t *TaskStatusTool) ShouldRaiseWarning(input any) (string, bool) {
	return "", false
}
