package models

import "time"

type ExecResult struct {
	Command  string        `json:"command"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

func (r ExecResult) Succeeded() bool { return r.ExitCode == 0 }
