package models

import (
	"encoding/json"
	"testing"
)

func TestAgentExecutionStep_JSON(t *testing.T) {
	step := AgentExecutionStep{
		Plan:        "Plan to run nginx",
		StepSummary: "Started nginx container",
	}

	jsonData, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed AgentExecutionStep
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Plan != step.Plan {
		t.Errorf("Plan = %q, want %q", parsed.Plan, step.Plan)
	}
	if parsed.StepSummary != step.StepSummary {
		t.Errorf("StepSummary = %q, want %q", parsed.StepSummary, step.StepSummary)
	}
}

func TestAgentExecutionStep_JSON_OmitEmpty(t *testing.T) {
	step := AgentExecutionStep{
		Plan:        "",
		StepSummary: "Only summary",
	}

	jsonData, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	jsonStr := string(jsonData)
	if containsString(jsonStr, "plan") {
		t.Errorf("Empty Plan should be omitted: %s", jsonStr)
	}
}

func TestAgentResult_JSON(t *testing.T) {
	result := AgentResult{
		Structured: &AgentExecutionStep{
			Plan:        "Test plan",
			StepSummary: "Test summary",
		},
		Raw:          "raw response",
		IsStructured: true,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed AgentResult
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.IsStructured != result.IsStructured {
		t.Errorf("IsStructured = %v, want %v", parsed.IsStructured, result.IsStructured)
	}
	if parsed.Raw != result.Raw {
		t.Errorf("Raw = %q, want %q", parsed.Raw, result.Raw)
	}
	if parsed.Structured == nil {
		t.Fatal("Structured should not be nil")
	}
	if parsed.Structured.Plan != "Test plan" {
		t.Errorf("Structured.Plan = %q, want %q", parsed.Structured.Plan, "Test plan")
	}
}

func TestAgentResult_JSON_NilStructured(t *testing.T) {
	result := AgentResult{
		Structured:   nil,
		Raw:          "raw only",
		IsStructured: false,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed AgentResult
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.IsStructured != false {
		t.Errorf("IsStructured = %v, want false", parsed.IsStructured)
	}
	if parsed.Structured != nil {
		t.Error("Structured should be nil")
	}
}

func TestToolCallInfo_JSON(t *testing.T) {
	info := ToolCallInfo{
		ToolName: "docker_command_tool",
		Command:  "docker ps",
		Status:   "completed",
		Result:   "container list",
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed ToolCallInfo
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.ToolName != info.ToolName {
		t.Errorf("ToolName = %q, want %q", parsed.ToolName, info.ToolName)
	}
	if parsed.Command != info.Command {
		t.Errorf("Command = %q, want %q", parsed.Command, info.Command)
	}
	if parsed.Status != info.Status {
		t.Errorf("Status = %q, want %q", parsed.Status, info.Status)
	}
	if parsed.Result != info.Result {
		t.Errorf("Result = %q, want %q", parsed.Result, info.Result)
	}
}

func TestToolCallInfo_JSON_OmitEmptyResult(t *testing.T) {
	info := ToolCallInfo{
		ToolName: "docker_command_tool",
		Command:  "docker ps",
		Status:   "running",
		// Result omitted
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	jsonStr := string(jsonData)
	if containsString(jsonStr, "result") {
		t.Errorf("Empty Result should be omitted: %s", jsonStr)
	}
}

func TestHistoryEntry_JSON(t *testing.T) {
	entry := HistoryEntry{
		Run:  1,
		Goal: "test goal",
		Plan: "test plan",
		ToolCalls: []ToolCallInfo{
			{ToolName: "tool1", Command: "cmd1", Status: "completed", Result: "result1"},
			{ToolName: "tool2", Command: "cmd2", Status: "running"},
		},
		Feedback: "step feedback",
		Done:     false,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed HistoryEntry
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Run != entry.Run {
		t.Errorf("Run = %d, want %d", parsed.Run, entry.Run)
	}
	if parsed.Goal != entry.Goal {
		t.Errorf("Goal = %q, want %q", parsed.Goal, entry.Goal)
	}
	if parsed.Plan != entry.Plan {
		t.Errorf("Plan = %q, want %q", parsed.Plan, entry.Plan)
	}
	if len(parsed.ToolCalls) != 2 {
		t.Errorf("ToolCalls = %d, want 2", len(parsed.ToolCalls))
	}
	if parsed.Feedback != entry.Feedback {
		t.Errorf("Feedback = %q, want %q", parsed.Feedback, entry.Feedback)
	}
	if parsed.Done != entry.Done {
		t.Errorf("Done = %v, want %v", parsed.Done, entry.Done)
	}
}

func TestUserInputPrompt_JSON(t *testing.T) {
	prompt := UserInputPrompt{
		Goal: "user goal",
		History: []HistoryEntry{
			{Run: 1, Goal: "previous goal", Plan: "prev plan"},
		},
	}

	jsonData, err := json.Marshal(prompt)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed UserInputPrompt
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Goal != prompt.Goal {
		t.Errorf("Goal = %q, want %q", parsed.Goal, prompt.Goal)
	}
	if len(parsed.History) != 1 {
		t.Errorf("History = %d, want 1", len(parsed.History))
	}
}