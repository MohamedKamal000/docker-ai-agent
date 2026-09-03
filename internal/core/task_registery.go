package core

import (
	"context"
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
	ID             string     `json:"id"`
	Tool           string     `json:"tool"`
	Input          string     `json:"input"`
	DockerObjectID string     `json:"docker_object_id,omitempty"`
	Status         TaskStatus `json:"status"`
	Result         string     `json:"result,omitempty"`
	Error          string     `json:"error,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	FinishedAt     time.Time  `json:"finished_at"`
	Reported       bool       `json:"reported"`
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
	r.UpdateComplete(id, status, result, errStr, "")
}

func (r *TaskRegistry) UpdateComplete(id string, status TaskStatus, result, errStr, objectID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.tasks[id]; ok {
		t.Status = status
		t.Result = result
		t.Error = errStr
		if objectID != "" {
			t.DockerObjectID = objectID
		}
		t.FinishedAt = time.Now()
	}
}

func (r *TaskRegistry) PullCompleted() []*TaskRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*TaskRecord
	for _, t := range r.tasks {
		if !t.Reported && (t.Status == TaskSucceeded || t.Status == TaskFailed || t.Status == TaskCanceled) {
			t.Reported = true
			cp := *t
			out = append(out, &cp)
		}
	}
	return out
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

func (r *TaskRegistry) Wait(ctx context.Context, id string) (*TaskRecord, error) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if rec, ok := r.Get(id); ok {
				if rec.Status == TaskSucceeded || rec.Status == TaskFailed || rec.Status == TaskCanceled {
					return rec, nil
				}
			} else {
				return nil, fmt.Errorf("task %s not found", id)
			}
		}
	}
}
