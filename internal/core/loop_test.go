package core

import (
	"docker-cli/internal/models"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON in markdown code block",
			input:    "```json\n{\"plan\": \"test\", \"step_summary\": \"done\"}\n```",
			expected: "{\"plan\": \"test\", \"step_summary\": \"done\"}",
		},
		{
			name:     "JSON in markdown with spaces",
			input:    "  ```json  \n  {\"key\": \"value\"}  \n  ```  ",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "Plain JSON object",
			input:    "{\"plan\": \"test\", \"step_summary\": \"done\"}",
			expected: "{\"plan\": \"test\", \"step_summary\": \"done\"}",
		},
		{
			name:     "JSON with surrounding text",
			input:    "Here is the result: {\"plan\": \"test\", \"step_summary\": \"done\"} end",
			expected: "{\"plan\": \"test\", \"step_summary\": \"done\"}",
		},
		{
			name:     "Multiple JSON objects - matches greedy (entire string)",
			input:    "{\"first\": 1} and {\"second\": 2}",
			expected: "{\"first\": 1} and {\"second\": 2}",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No JSON",
			input:    "just plain text",
			expected: "",
		},
		{
			name:     "JSON with newlines",
			input:    "```json\n{\n  \"plan\": \"multi\\nline\",\n  \"step_summary\": \"done\"\n}\n```",
			expected: "{\n  \"plan\": \"multi\\nline\",\n  \"step_summary\": \"done\"\n}",
		},
		{
			name:     "Nested braces",
			input:    "{\"outer\": {\"inner\": \"value\"}}",
			expected: "{\"outer\": {\"inner\": \"value\"}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractResult(t *testing.T) {
	tests := []struct {
		name         string
		resp         *ai.ModelResponse
		expectedRaw  string
		expectStruct bool
	}{
		{
			name:         "Nil response",
			resp:         nil,
			expectedRaw:  "empty response",
			expectStruct: false,
		},
		{
			name: "Structured JSON response",
			resp: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: "```json\n{\"plan\": \"test plan\", \"step_summary\": \"test summary\"}\n```"},
					},
				},
			},
			expectedRaw:  "```json\n{\"plan\": \"test plan\", \"step_summary\": \"test summary\"}\n```",
			expectStruct: true,
		},
		{
			name: "Plain text response",
			resp: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: "Just a text response"},
					},
				},
			},
			expectedRaw:  "Just a text response",
			expectStruct: false,
		},
		{
			name: "Response with invalid JSON",
			resp: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: "```json\n{\"invalid json}\n```"},
					},
				},
			},
			expectedRaw:  "```json\n{\"invalid json}\n```",
			expectStruct: false,
		},
		{
			name: "Empty message",
			resp: &ai.ModelResponse{
				Message: &ai.Message{},
			},
			expectedRaw:  "",
			expectStruct: false,
		},
		{
			name: "Response with multiple parts",
			resp: &ai.ModelResponse{
				Message: &ai.Message{
					Content: []*ai.Part{
						{Kind: ai.PartText, Text: "Some text"},
						{Kind: ai.PartText, Text: "```json\n{\"plan\": \"test\", \"step_summary\": \"done\"}\n```"},
					},
				},
			},
			expectStruct: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractResult(tt.resp)

			if tt.expectStruct {
				if !result.IsStructured {
					t.Errorf("extractResult() IsStructured = false, want true")
				}
				if result.Structured == nil {
					t.Error("extractResult() Structured = nil")
				} else {
					if result.Structured.Plan != "test plan" && result.Structured.Plan != "test" {
						t.Errorf("Plan = %q", result.Structured.Plan)
					}
				}
			} else {
				if result.IsStructured {
					t.Errorf("extractResult() IsStructured = true, want false")
				}
				if result.Raw != tt.expectedRaw {
					t.Errorf("Raw = %q, want %q", result.Raw, tt.expectedRaw)
				}
			}
		})
	}
}

func TestFormatTaskResult(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		task     *TaskRecord
		expected string
	}{
		{
			name: "Succeeded task with result",
			task: &TaskRecord{
				ID:        "task-1",
				Tool:      "docker_command_tool",
				Input:     "docker ps",
				Status:    TaskSucceeded,
				StartedAt: now,
				FinishedAt: now,
				Reported:  true,
				Result: &models.ExecResult{
					Command:  "docker ps",
					Stdout:   "CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES",
					Stderr:   "",
					ExitCode: 0,
				},
				Error: "",
			},
			expected: "Status: succeeded\nResult: ExitCode: 0\nStdout: CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES",
		},
		{
			name: "Failed task with error",
			task: &TaskRecord{
				ID:        "task-2",
				Tool:      "docker_command_tool",
				Input:     "docker rm container1",
				Status:    TaskFailed,
				StartedAt: now,
				FinishedAt: now,
				Reported:  true,
				Result: &models.ExecResult{
					Command:  "docker rm container1",
					Stdout:   "",
					Stderr:   "Error: No such container",
					ExitCode: 1,
				},
				Error: "exit 1: Error: No such container",
			},
			expected: "Status: failed\nError: exit 1: Error: No such container\nResult: ExitCode: 1\nStderr: Error: No such container",
		},
		{
			name: "Task with only error, no result",
			task: &TaskRecord{
				ID:        "task-3",
				Tool:      "docker_command_tool",
				Input:     "docker invalid",
				Status:    TaskFailed,
				StartedAt: now,
				FinishedAt: now,
				Reported:  true,
				Result:    nil,
				Error:     "command not found",
			},
			expected: "Status: failed\nError: command not found",
		},
		{
			name: "Canceled task",
			task: &TaskRecord{
				ID:        "task-4",
				Tool:      "docker_command_tool",
				Input:     "docker ps",
				Status:    TaskCanceled,
				StartedAt: now,
				FinishedAt: now,
				Reported:  true,
				Result:    nil,
				Error:     "context canceled",
			},
			expected: "Status: canceled\nError: context canceled",
		},
		{
			name: "Running task (no result or error)",
			task: &TaskRecord{
				ID:        "task-5",
				Tool:      "docker_command_tool",
				Input:     "docker ps",
				Status:    TaskRunning,
				StartedAt: now,
				FinishedAt: time.Time{},
				Reported:  false,
				Result:    nil,
				Error:     "",
			},
			expected: "Status: running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTaskResult(tt.task)
			if result != tt.expected {
				t.Errorf("formatTaskResult() = %q, want %q", result, tt.expected)
			}
		})
	}
}