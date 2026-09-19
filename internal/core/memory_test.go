package core

import (
	"docker-cli/internal/models"
	"testing"
)

func TestStaticMemoryStore_Save(t *testing.T) {
	store := NewStaticMemoryStore()

	entries := []models.HistoryEntry{
		{
			Run:  1,
			Goal: "test goal 1",
			ToolCalls: []models.ToolCallInfo{
				{ToolName: "tool1", Command: "cmd1", Status: "completed", Result: "result1"},
			},
		},
		{
			Run:  2,
			Goal: "test goal 2",
			ToolCalls: []models.ToolCallInfo{
				{ToolName: "tool2", Command: "cmd2", Status: "completed", Result: "result2"},
			},
		},
	}

	err := store.Save(entries)
	if err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("Load() returned %d entries, want 2", len(loaded))
	}

	if loaded[0].Goal != "test goal 1" {
		t.Errorf("First entry Goal = %q, want %q", loaded[0].Goal, "test goal 1")
	}
	if loaded[1].Goal != "test goal 2" {
		t.Errorf("Second entry Goal = %q, want %q", loaded[1].Goal, "test goal 2")
	}
}

func TestStaticMemoryStore_SaveMultipleTimes(t *testing.T) {
	store := NewStaticMemoryStore()

	entries1 := []models.HistoryEntry{
		{Run: 1, Goal: "goal 1"},
	}
	entries2 := []models.HistoryEntry{
		{Run: 2, Goal: "goal 2"},
	}
	entries3 := []models.HistoryEntry{
		{Run: 3, Goal: "goal 3"},
	}

	if err := store.Save(entries1); err != nil {
		t.Fatalf("Save() first call error: %v", err)
	}
	if err := store.Save(entries2); err != nil {
		t.Fatalf("Save() second call error: %v", err)
	}
	if err := store.Save(entries3); err != nil {
		t.Fatalf("Save() third call error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("Load() returned %d entries, want 3", len(loaded))
	}
}

func TestStaticMemoryStore_LoadEmpty(t *testing.T) {
	store := NewStaticMemoryStore()

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() on empty store returned error: %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("Load() on empty store returned %d entries, want 0", len(loaded))
	}
}

func TestStaticMemoryStore_Clear(t *testing.T) {
	store := NewStaticMemoryStore()

	entries := []models.HistoryEntry{
		{Run: 1, Goal: "test goal"},
	}

	if err := store.Save(entries); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear() returned error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() after Clear() error: %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("Load() after Clear() returned %d entries, want 0", len(loaded))
	}
}

func TestStaticMemoryStore_SaveAndLoadPreservesOrder(t *testing.T) {
	store := NewStaticMemoryStore()

	entries := []models.HistoryEntry{
		{Run: 1, Goal: "first"},
		{Run: 2, Goal: "second"},
		{Run: 3, Goal: "third"},
	}

	if err := store.Save(entries); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for i, entry := range loaded {
		if entry.Goal != entries[i].Goal {
			t.Errorf("Entry %d Goal = %q, want %q", i, entry.Goal, entries[i].Goal)
		}
	}
}

func TestStaticMemoryStore_SaveWithToolCalls(t *testing.T) {
	store := NewStaticMemoryStore()

	entries := []models.HistoryEntry{
		{
			Run:  1,
			Goal: "test with tool calls",
			ToolCalls: []models.ToolCallInfo{
				{ToolName: "docker_command_tool", Command: "docker ps", Status: "completed", Result: "container list"},
				{ToolName: "task_status_tool", Command: "task-123", Status: "running", Result: ""},
			},
		},
	}

	if err := store.Save(entries); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("Load() returned %d entries, want 1", len(loaded))
	}

	if len(loaded[0].ToolCalls) != 2 {
		t.Errorf("Load() returned %d tool calls, want 2", len(loaded[0].ToolCalls))
	}

	if loaded[0].ToolCalls[0].ToolName != "docker_command_tool" {
		t.Errorf("First tool call name = %q, want %q", loaded[0].ToolCalls[0].ToolName, "docker_command_tool")
	}
	if loaded[0].ToolCalls[1].ToolName != "task_status_tool" {
		t.Errorf("Second tool call name = %q, want %q", loaded[0].ToolCalls[1].ToolName, "task_status_tool")
	}
}