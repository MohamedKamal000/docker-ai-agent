package core

import (
	"docker-cli/internal/models"
	"testing"
	"time"
)

func TestNewTaskRecord(t *testing.T) {
	tool := "test_tool"
	command := "test command"

	record := NewTaskRecord(tool, command)

	if record.Tool != tool {
		t.Errorf("Tool = %q, want %q", record.Tool, tool)
	}
	if record.Input != command {
		t.Errorf("Input = %q, want %q", record.Input, command)
	}
	if record.Status != TaskRunning {
		t.Errorf("Status = %v, want %v", record.Status, TaskRunning)
	}
	if record.ID == "" {
		t.Error("ID should not be empty")
	}
	if record.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}
	if record.Reported {
		t.Error("Reported should be false initially")
	}
	if record.Result != nil {
		t.Error("Result should be nil initially")
	}
	if record.Error != "" {
		t.Error("Error should be empty initially")
	}
	if record.FinishedAt.IsZero() == false {
		t.Error("FinishedAt should be zero initially")
	}
}

func TestTaskRegistry_Register(t *testing.T) {
	registry := NewTaskRegistry()

	record := NewTaskRecord("tool1", "command1")
	id := registry.Register(record)

	if id == "" {
		t.Error("Register() returned empty ID")
	}

	if id != record.ID {
		t.Errorf("Register() returned ID %q, want %q", id, record.ID)
	}

	retrieved, ok := registry.Get(id)
	if !ok {
		t.Fatal("Get() failed to find registered task")
	}
	if retrieved.Tool != "tool1" {
		t.Errorf("Retrieved task Tool = %q, want %q", retrieved.Tool, "tool1")
	}
}

func TestTaskRegistry_Get(t *testing.T) {
	registry := NewTaskRegistry()

	record := NewTaskRecord("tool1", "command1")
	registry.Register(record)

	retrieved, ok := registry.Get(record.ID)
	if !ok {
		t.Fatal("Get() failed to find existing task")
	}
	if retrieved.ID != record.ID {
		t.Errorf("Get() returned task with ID %q, want %q", retrieved.ID, record.ID)
	}

	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for nonexistent task")
	}
}

func TestTaskRegistry_UpdateComplete(t *testing.T) {
	registry := NewTaskRegistry()

	record := NewTaskRecord("tool1", "command1")
	registry.Register(record)

	result := &models.ExecResult{
		Command:  "command1",
		Stdout:   "output",
		Stderr:   "",
		ExitCode: 0,
		Duration: 100 * time.Millisecond,
	}

	registry.UpdateComplete(record.ID, TaskSucceeded, result, "")

	updated, ok := registry.Get(record.ID)
	if !ok {
		t.Fatal("Get() failed after UpdateComplete")
	}

	if updated.Status != TaskSucceeded {
		t.Errorf("Status = %v, want %v", updated.Status, TaskSucceeded)
	}
	if updated.Result == nil {
		t.Fatal("Result should not be nil after UpdateComplete")
	}
	if updated.Result.Stdout != "output" {
		t.Errorf("Result.Stdout = %q, want %q", updated.Result.Stdout, "output")
	}
	if updated.Error != "" {
		t.Errorf("Error = %q, want empty", updated.Error)
	}
	if updated.FinishedAt.IsZero() {
		t.Error("FinishedAt should not be zero after UpdateComplete")
	}
}

func TestTaskRegistry_UpdateCompleteWithError(t *testing.T) {
	registry := NewTaskRegistry()

	record := NewTaskRecord("tool1", "command1")
	registry.Register(record)

	result := &models.ExecResult{
		Command:  "command1",
		Stdout:   "",
		Stderr:   "error output",
		ExitCode: 1,
		Duration: 50 * time.Millisecond,
	}

	registry.UpdateComplete(record.ID, TaskFailed, result, "command failed")

	updated, ok := registry.Get(record.ID)
	if !ok {
		t.Fatal("Get() failed after UpdateComplete")
	}

	if updated.Status != TaskFailed {
		t.Errorf("Status = %v, want %v", updated.Status, TaskFailed)
	}
	if updated.Error != "command failed" {
		t.Errorf("Error = %q, want %q", updated.Error, "command failed")
	}
}

func TestTaskRegistry_PullRunning(t *testing.T) {
	registry := NewTaskRegistry()

	record1 := NewTaskRecord("tool1", "command1")
	registry.Register(record1)

	record2 := NewTaskRecord("tool2", "command2")
	registry.Register(record2)

	running := registry.PullRunning()
	if len(running) != 2 {
		t.Errorf("PullRunning() returned %d tasks, want 2", len(running))
	}

	for _, r := range running {
		if r.Status != TaskRunning {
			t.Errorf("PullRunning() returned task with status %v, want %v", r.Status, TaskRunning)
		}
	}
}

func TestTaskRegistry_PullCompleted(t *testing.T) {
	registry := NewTaskRegistry()

	record1 := NewTaskRecord("tool1", "command1")
	registry.Register(record1)

	record2 := NewTaskRecord("tool2", "command2")
	registry.Register(record2)

	result := &models.ExecResult{Command: "command1", ExitCode: 0}

	registry.UpdateComplete(record1.ID, TaskSucceeded, result, "")
	registry.UpdateComplete(record2.ID, TaskFailed, result, "error")

	completed := registry.PullCompleted()
	if len(completed) != 2 {
		t.Errorf("PullCompleted() returned %d tasks, want 2", len(completed))
	}

	for _, c := range completed {
		if c.Status != TaskSucceeded && c.Status != TaskFailed {
			t.Errorf("PullCompleted() returned task with status %v, want succeeded or failed", c.Status)
		}
		if !c.Reported {
			t.Errorf("PullCompleted() task should be marked as reported")
		}
	}

	secondPull := registry.PullCompleted()
	if len(secondPull) != 0 {
		t.Errorf("Second PullCompleted() returned %d tasks, want 0", len(secondPull))
	}
}

func TestTaskRegistry_PullCompleted_OnlyReportsOnce(t *testing.T) {
	registry := NewTaskRegistry()

	record := NewTaskRecord("tool1", "command1")
	registry.Register(record)

	result := &models.ExecResult{Command: "command1", ExitCode: 0}
	registry.UpdateComplete(record.ID, TaskSucceeded, result, "")

	firstPull := registry.PullCompleted()
	if len(firstPull) != 1 {
		t.Errorf("First PullCompleted() returned %d tasks, want 1", len(firstPull))
	}

	secondPull := registry.PullCompleted()
	if len(secondPull) != 0 {
		t.Errorf("Second PullCompleted() returned %d tasks, want 0", len(secondPull))
	}
}

func TestTaskRegistry_PullCompleted_Canceled(t *testing.T) {
	registry := NewTaskRegistry()

	record := NewTaskRecord("tool1", "command1")
	registry.Register(record)

	result := &models.ExecResult{Command: "command1", ExitCode: 0}
	registry.UpdateComplete(record.ID, TaskCanceled, result, "canceled")

	completed := registry.PullCompleted()
	if len(completed) != 1 {
		t.Errorf("PullCompleted() returned %d tasks, want 1", len(completed))
	}
	if completed[0].Status != TaskCanceled {
		t.Errorf("Status = %v, want %v", completed[0].Status, TaskCanceled)
	}
}

func TestTaskRegistry_List(t *testing.T) {
	registry := NewTaskRegistry()

	record1 := NewTaskRecord("tool1", "command1")
	registry.Register(record1)

	record2 := NewTaskRecord("tool2", "command2")
	registry.Register(record2)

	list := registry.List()
	if len(list) != 2 {
		t.Errorf("List() returned %d tasks, want 2", len(list))
	}

	found1 := false
	found2 := false
	for _, t := range list {
		if t.ID == record1.ID {
			found1 = true
		}
		if t.ID == record2.ID {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Error("List() missing registered tasks")
	}
}

func TestTaskRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewTaskRegistry()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			record := NewTaskRecord("tool", "command")
			registry.Register(record)
			registry.UpdateComplete(record.ID, TaskSucceeded, &models.ExecResult{}, "")
			registry.PullCompleted()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	list := registry.List()
	if len(list) < 1 {
		t.Errorf("List() returned %d tasks after concurrent access, want at least 1", len(list))
	}
}