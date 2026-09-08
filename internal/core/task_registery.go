package core

import (
	"fmt"
	"sync"
	"time"

	"docker-cli/internal/models"
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
	ID         string             `json:"id"`
	Tool       string             `json:"tool"`
	Input      string             `json:"input"`
	Status     TaskStatus         `json:"status"`
	StartedAt  time.Time          `json:"started_at"`
	FinishedAt time.Time          `json:"finished_at"`
	Reported   bool               `json:"reported"`
	Result     *models.ExecResult `json:"result,omitempty"`
	Error      string             `json:"error,omitempty"`
}

func NewTaskRecord(tool string, command string) *TaskRecord {
	t := TaskRecord{
		Tool:  tool,
		Input: command,
	}

	t.ID = fmt.Sprintf("task-%s-%d", command, time.Now().Unix())
	t.Status = TaskRunning
	t.StartedAt = time.Now()
	return &t
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

func (r *TaskRegistry) UpdateComplete(id string, status TaskStatus, result *models.ExecResult, errStr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.tasks[id]; ok {
		t.Result = result
		t.Error = errStr
		t.Status = status
		t.FinishedAt = time.Now()
	}
}

func (r *TaskRegistry) pullWithCondition(isSatisfied func(t *TaskRecord) bool) []*TaskRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*TaskRecord
	for _, t := range r.tasks {
		if isSatisfied(t) {
			out = append(out, t)
		}
	}
	return out
}

func (r *TaskRegistry) PullRunning() []*TaskRecord {
	return r.pullWithCondition(func(t *TaskRecord) bool {
		if t.Status == TaskRunning {
			return true
		}
		return false
	})
}

func (r *TaskRegistry) PullCompleted() []*TaskRecord {
	return r.pullWithCondition(func(t *TaskRecord) bool {
		if !t.Reported && (t.Status == TaskSucceeded || t.Status == TaskFailed || t.Status == TaskCanceled) {
			t.Reported = true
			return true
		}
		return false
	})
}

func (r *TaskRegistry) List() []*TaskRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*TaskRecord, 0, len(r.tasks))
	for _, t := range r.tasks {
		out = append(out, t)
	}
	return out
}
