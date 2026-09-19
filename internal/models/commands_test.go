package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestExecResult_JSON(t *testing.T) {
	result := ExecResult{
		Command:  "docker ps",
		Stdout:   "CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES",
		Stderr:   "",
		ExitCode: 0,
		Duration: 100 * time.Millisecond,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var parsed ExecResult
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if parsed.Command != result.Command {
		t.Errorf("Command = %q, want %q", parsed.Command, result.Command)
	}
	if parsed.Stdout != result.Stdout {
		t.Errorf("Stdout = %q, want %q", parsed.Stdout, result.Stdout)
	}
	if parsed.Stderr != result.Stderr {
		t.Errorf("Stderr = %q, want %q", parsed.Stderr, result.Stderr)
	}
	if parsed.ExitCode != result.ExitCode {
		t.Errorf("ExitCode = %d, want %d", parsed.ExitCode, result.ExitCode)
	}
	if parsed.Duration != result.Duration {
		t.Errorf("Duration = %v, want %v", parsed.Duration, result.Duration)
	}
}

func TestExecResult_Succeeded(t *testing.T) {
	tests := []struct {
		name        string
		exitCode    int
		expectSucceed bool
	}{
		{"Exit code 0", 0, true},
		{"Exit code 1", 1, false},
		{"Exit code 127", 127, false},
		{"Exit code -1", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExecResult{ExitCode: tt.exitCode}
			if result.Succeeded() != tt.expectSucceed {
				t.Errorf("Succeeded() = %v, want %v", result.Succeeded(), tt.expectSucceed)
			}
		})
	}
}

func TestExecResult_String(t *testing.T) {
	tests := []struct {
		name     string
		result   ExecResult
		expected string
	}{
		{
			name: "Success with stdout",
			result: ExecResult{
				Command:  "docker ps",
				Stdout:   "container list",
				Stderr:   "",
				ExitCode: 0,
			},
			expected: "ExitCode: 0\nStdout: container list\n",
		},
		{
			name: "Failure with stderr",
			result: ExecResult{
				Command:  "docker rm invalid",
				Stdout:   "",
				Stderr:   "Error: No such container",
				ExitCode: 1,
			},
			expected: "ExitCode: 1\nStderr: Error: No such container\n",
		},
		{
			name: "Success with both stdout and stderr",
			result: ExecResult{
				Command:  "docker build .",
				Stdout:   "Successfully built",
				Stderr:   "warning: something",
				ExitCode: 0,
			},
			expected: "ExitCode: 0\nStdout: Successfully built\nStderr: warning: something\n",
		},
		{
			name: "Only exit code",
			result: ExecResult{
				Command:  "docker version",
				Stdout:   "",
				Stderr:   "",
				ExitCode: 0,
			},
			expected: "ExitCode: 0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}