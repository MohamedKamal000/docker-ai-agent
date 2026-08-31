package core

import (
	"fmt"
	"sync"
	"time"
)

type TaskStatus string

const (
	TaskQueued    TaskStatus = "queued"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
	TaskCanceled  TaskStatus = "canceled"
)

type TaskRecord struct {
	ID         string     `json:"id"`
	Tool       string     `json:"tool"`
	Input      string     `json:"input"`
	Status     TaskStatus `json:"status"`
	Result     string     `json:"result,omitempty"`
	Error      string     `json:"error,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt time.Time  `json:"finished_at"`
}

type TaskRegistry struct {
	mu    sync.RWMutex
	tasks map[string]*TaskRecord
}

func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{tasks: make(map[string]*TaskRecord)}
}

func (r *TaskRegistry) Register(rec *TaskRecord) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	rec.Status = TaskRunning
	rec.StartedAt = time.Now()
	r.tasks[rec.ID] = rec
	return rec.ID
}

func (r *TaskRegistry) Get(id string) (*TaskRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, ok
}

func (r *TaskRegistry) Update(id string, status TaskStatus, result, errStr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.tasks[id]; ok {
		t.Status = status
		t.Result = result
		t.Error = errStr
		t.FinishedAt = time.Now()
	}
}

func (r *TaskRegistry) List() []*TaskRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*TaskRecord, 0, len(r.tasks))
	for _, t := range r.tasks {
		cp := *t
		out = append(out, &cp)
	}
	return out
}
