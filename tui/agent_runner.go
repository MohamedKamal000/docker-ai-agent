package tui

import (
	"context"
	"errors"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"docker-cli/tui/screens"
)

// Sink receives the runner's UI events. *tea.Program satisfies it.
type Sink interface {
	Send(tea.Msg)
}

// runPhase tracks where the active run is inside the confirmation flow.
type runPhase int

const (
	runRunning runPhase = iota
	runAwaitingConfirmation
)

type activeRun struct {
	ctx    context.Context
	cancel context.CancelFunc
	comm   *core.AgentCommunication
	phase  runPhase
}

// AgentRunner owns the full lifecycle of the conversation between the user and
// the agent: starting runs, cancelation, confirmations, and teardown. It
// enforces that at most one agent loop is alive at any time, and that the UI
// only learns a run is over (terminal RunEvent) after the run has been fully
// torn down.
//
// Contract with AgentLoop implementations: Run must close comm.ToUser when it
// returns; FromUser is never closed by the runner so late sends cannot panic.
type AgentRunner struct {
	agent *app.Agent
	sink  Sink

	mu  sync.Mutex
	run *activeRun // nil when idle; non-nil for the entire run incl. teardown
}

func NewAgentRunner(agent *app.Agent, sink Sink) *AgentRunner {
	return &AgentRunner{
		agent: agent,
		sink:  sink,
	}
}

// Start launches a new agent loop for userMessage. It refuses if a previous
// run has not been fully torn down yet: only one loop may exist at a time.
func (r *AgentRunner) Start(userMessage string) {
	comm := &core.AgentCommunication{
		ToUser:   make(chan core.AiResponse),
		FromUser: make(chan core.UserCommand),
	}
	ctx, cancel := context.WithCancel(context.Background())

	r.mu.Lock()
	if r.run != nil {
		r.mu.Unlock()
		cancel() // a loop is already active; refuse and release the context
		return
	}
	r.run = &activeRun{
		ctx:    ctx,
		cancel: cancel,
		comm:   comm,
		phase:  runRunning,
	}
	r.mu.Unlock()

	runErr := make(chan error, 1)
	go func() {
		runErr <- r.agent.AgentLoop.Run(ctx, userMessage, comm)
	}()

	go func() {
		for output := range comm.ToUser {
			if output.Type == core.Warning {
				r.setPhase(runAwaitingConfirmation)
			}
			r.sink.Send(screens.RunEvent{Kind: screens.RunResponse, Data: output})
		}
		// ToUser is closed by the agent when Run returns: everything has been
		// drained at this point, so it is safe to tear down and report.
		err := <-runErr
		r.teardown(err)
	}()
}

// Cancel aborts the active run, if any. It is a no-op when idle.
func (r *AgentRunner) Cancel() {
	r.mu.Lock()
	var cancel context.CancelFunc
	if r.run != nil {
		cancel = r.run.cancel
	}
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// Confirm answers a pending tool-execution warning. It only acts while a
// warning is actually awaiting confirmation and is ignored otherwise. The
// delivery happens off the caller's thread and gives up if the run is canceled
// or the agent doesn't consume the answer in time.
func (r *AgentRunner) Confirm(shouldContinue bool) {
	r.mu.Lock()
	var (
		comm *core.AgentCommunication
		ctx  context.Context
	)
	if r.run != nil && r.run.phase == runAwaitingConfirmation {
		comm = r.run.comm
		ctx = r.run.ctx
		r.run.phase = runRunning
	}
	r.mu.Unlock()

	if comm == nil {
		return
	}

	cmd := core.UserCommand{ShouldContinue: shouldContinue}
	go func() {
		select {
		case comm.FromUser <- cmd:
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}
	}()
}

func (r *AgentRunner) setPhase(phase runPhase) {
	r.mu.Lock()
	if r.run != nil {
		r.run.phase = phase
	}
	r.mu.Unlock()
}

// teardown clears the run and emits exactly one terminal event, but only once
// the pump goroutine above has fully exited.
func (r *AgentRunner) teardown(err error) {
	r.mu.Lock()
	if r.run == nil {
		r.mu.Unlock()
		return
	}
	r.run.cancel()
	r.run = nil
	r.mu.Unlock()

	switch {
	case err == nil:
		r.sink.Send(screens.RunEvent{Kind: screens.RunFinished})
	case errors.Is(err, context.Canceled):
		r.sink.Send(screens.RunEvent{Kind: screens.RunCanceled})
	default:
		r.sink.Send(screens.RunEvent{Kind: screens.RunFailed, Text: err.Error()})
	}
}
