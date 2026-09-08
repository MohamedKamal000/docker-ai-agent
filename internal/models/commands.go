package models

import (
	"fmt"
	"strings"
	"time"
)

type ExecResult struct {
	Command  string        `json:"command"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

func (r ExecResult) Succeeded() bool { return r.ExitCode == 0 }

func (r ExecResult) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "ExitCode: %d\n", r.ExitCode)
	if r.Stdout != "" {
		fmt.Fprintf(&b, "Stdout: %s\n", r.Stdout)
	}
	if r.Stderr != "" {
		fmt.Fprintf(&b, "Stderr: %s\n", r.Stderr)
	}
	return b.String()
}
