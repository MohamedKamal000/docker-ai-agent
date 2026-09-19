package core

import (
	"docker-cli/internal/models"
	"encoding/json"
	"testing"
	"time"
)

func TestEvaluatorInput(t *testing.T) {
	history := []models.HistoryEntry{
		{Run: 1, Goal: "test goal"},
	}
	
	input := EvaluatorInput{
		Goal:    "test goal",
		History: history,
	}
	
	if input.Goal != "test goal" {
		t.Errorf("Goal = %q, want %q", input.Goal, "test goal")
	}
	if len(input.History) != 1 {
		t.Errorf("History length = %d, want 1", len(input.History))
	}
}

func TestEvaluatorResult(t *testing.T) {
	tests := []struct {
		name           string
		result         EvaluatorResult
		expectAccomplished bool
	}{
		{
			name: "Goal accomplished with final response",
			result: EvaluatorResult{
				GoalAccomplished: true,
				FinalResponse:    "Task completed successfully",
			},
			expectAccomplished: true,
		},
		{
			name: "Goal not accomplished with feedback",
			result: EvaluatorResult{
				GoalAccomplished: false,
				Feedback:         "Need to run more commands",
			},
			expectAccomplished: false,
		},
		{
			name: "Goal accomplished without final response",
			result: EvaluatorResult{
				GoalAccomplished: true,
			},
			expectAccomplished: true,
		},
		{
			name: "Goal not accomplished without feedback",
			result: EvaluatorResult{
				GoalAccomplished: false,
			},
			expectAccomplished: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.GoalAccomplished != tt.expectAccomplished {
				t.Errorf("GoalAccomplished = %v, want %v", tt.result.GoalAccomplished, tt.expectAccomplished)
			}
		})
	}
}

func TestEvaluatorResultJSONMarshal(t *testing.T) {
	// Test JSON marshaling for accomplished goal
	result := EvaluatorResult{
		GoalAccomplished: true,
		FinalResponse:    "Container started successfully",
	}
	
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	var parsed EvaluatorResult
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	
	if parsed.GoalAccomplished != true {
		t.Errorf("Parsed GoalAccomplished = %v, want true", parsed.GoalAccomplished)
	}
	if parsed.FinalResponse != "Container started successfully" {
		t.Errorf("Parsed FinalResponse = %q, want %q", parsed.FinalResponse, "Container started successfully")
	}

	// Test JSON marshaling for unaccomplished goal
	result = EvaluatorResult{
		GoalAccomplished: false,
		Feedback:         "Container failed to start",
	}
	
	jsonData, err = json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	parsed = EvaluatorResult{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	
	if parsed.GoalAccomplished != false {
		t.Errorf("Parsed GoalAccomplished = %v, want false", parsed.GoalAccomplished)
	}
	if parsed.Feedback != "Container failed to start" {
		t.Errorf("Parsed Feedback = %q, want %q", parsed.Feedback, "Container failed to start")
	}
}

func TestExtractJSONForEvaluator(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedGoal bool
		expectedFeedback string
		expectedResponse string
	}{
		{
			name: "Accomplished goal with final response",
			input: `{"goal_accomplished": true, "final_response": "Done"}`,
			expectedGoal: true,
			expectedResponse: "Done",
		},
		{
			name: "Not accomplished with feedback",
			input: `{"goal_accomplished": false, "feedback": "Need to retry"}`,
			expectedGoal: false,
			expectedFeedback: "Need to retry",
		},
		{
			name: "Accomplished without response",
			input: `{"goal_accomplished": true}`,
			expectedGoal: true,
		},
		{
			name: "Not accomplished without feedback",
			input: `{"goal_accomplished": false}`,
			expectedGoal: false,
		},
		{
			name: "JSON in markdown",
			input: "```json\n{\"goal_accomplished\": true, \"final_response\": \"Success\"}\n```",
			expectedGoal: true,
			expectedResponse: "Success",
		},
		{
			name: "Invalid JSON",
			input: "not json",
			expectedGoal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr := extractJSON(tt.input)
			if jsonStr == "" {
				if tt.expectedGoal {
					t.Errorf("extractJSON returned empty for %q", tt.input)
				}
				return
			}
			
			var result EvaluatorResult
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				if tt.expectedGoal || tt.expectedFeedback != "" || tt.expectedResponse != "" {
					t.Errorf("Failed to parse JSON: %v", err)
				}
				return
			}
			
			if result.GoalAccomplished != tt.expectedGoal {
				t.Errorf("GoalAccomplished = %v, want %v", result.GoalAccomplished, tt.expectedGoal)
			}
			if result.Feedback != tt.expectedFeedback {
				t.Errorf("Feedback = %q, want %q", result.Feedback, tt.expectedFeedback)
			}
			if result.FinalResponse != tt.expectedResponse {
				t.Errorf("FinalResponse = %q, want %q", result.FinalResponse, tt.expectedResponse)
			}
		})
	}
}

func TestFormatTaskResultForEvaluator(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	
	task := &TaskRecord{
		ID:        "eval-task-1",
		Tool:      "docker_command_tool",
		Input:     "docker run nginx",
		Status:    TaskSucceeded,
		StartedAt: now,
		FinishedAt: now,
		Reported:  true,
		Result: &models.ExecResult{
			Command:  "docker run nginx",
			Stdout:   "Container started",
			Stderr:   "",
			ExitCode: 0,
		},
		Error: "",
	}
	
	result := formatTaskResult(task)
	
	if !containsString(result, "Status: succeeded") {
		t.Errorf("formatTaskResult missing status: %q", result)
	}
	if !containsString(result, "ExitCode: 0") {
		t.Errorf("formatTaskResult missing exit code: %q", result)
	}
	if !containsString(result, "Container started") {
		t.Errorf("formatTaskResult missing stdout: %q", result)
	}
}