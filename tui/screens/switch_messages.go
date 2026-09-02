package screens

import "docker-cli/internal/core"

// SwitchToNewSessionMessage is emitted by the home screen to open a new chat
// session.
type SwitchToNewSessionMessage struct{}

// SendMessageRequest asks the agent runner to start a run for a user message.
type SendMessageRequest struct {
	UserMessage string
}

// CancelAgentRequest asks the agent runner to abort the active run.
type CancelAgentRequest struct{}

// ConfirmMessage answers a pending tool-execution warning (y/n).
type ConfirmMessage struct {
	ShouldContinue bool
}

type RunEventKind int

const (
	// RunResponse carries an incremental agent response (thought, final text,
	// retrying notice, warning).
	RunResponse RunEventKind = iota
	// RunFailed is terminal: the run ended with an error (Text holds it).
	RunFailed
	// RunCanceled is terminal: the user aborted the run.
	RunCanceled
	// RunFinished is terminal: the run completed cleanly. Terminal events are
	// only sent after the runner has fully torn the run down, so the UI may
	// safely allow a new message once it receives one.
	RunFinished
)

// RunEvent is the single event type flowing from the agent runner to the UI.
// Exactly one terminal event (RunFailed, RunCanceled or RunFinished) is
// delivered per run.
type RunEvent struct {
	Kind RunEventKind
	Text string          // for RunFailed
	Data core.AiResponse // for RunResponse
}
