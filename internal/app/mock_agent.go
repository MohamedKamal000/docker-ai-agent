package app

import (
	"context"
	"docker-cli/internal/models"
	"fmt"
	"strings"
	"time"

	"docker-cli/internal/core"
)

type MockStepType int

const (
	MockThought MockStepType = iota
	MockWarning
	MockFinal
)

type MockStep struct {
	Type    MockStepType
	Message string
}

type MockAgentLoop struct {
	Steps []MockStep
	Delay time.Duration
}

func NewMockAgent(config core.AppConfig, ctx context.Context, toolsToRegister []string) *Agent {
	chatSession := core.NewStaticMemoryStore()

	agentLoop := &MockAgentLoop{Delay: 1 * time.Second}

	return &Agent{
		AgentLoop: agentLoop,
		SessionContext: &core.LoopContext{
			Memory: chatSession,
			Tools:  nil,
		},
	}
}

func (mal *MockAgentLoop) Run(ctx context.Context, userGoal string, comm *core.AgentCommunication) error {
	defer close(comm.ToUser)

	steps := mal.Steps
	if len(steps) == 0 {
		steps = defaultMockSteps(userGoal)
	}

	delay := mal.Delay
	if delay == 0 {
		delay = 1 * time.Second
	}

	for _, step := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch step.Type {
		case MockThought:
			comm.ToUser <- core.NewThought(models.AgentResult{
				IsStructured: false,
				Raw:          step.Message,
			})
		case MockWarning:
			comm.ToUser <- core.NewWarning(step.Message)
			var cmd core.UserCommand
			select {
			case cmd = <-comm.FromUser:
			case <-ctx.Done():
				return ctx.Err()
			}
			if !cmd.ShouldContinue {
				comm.ToUser <- core.NewFinal(models.AgentResult{
					IsStructured: false,
					Raw:          "Execution cancelled by user.",
				})
				return nil
			}
		case MockFinal:
			comm.ToUser <- core.NewFinal(models.AgentResult{
				IsStructured: false,
				Raw:          step.Message,
			})
		}

		if err := core.SleepCancellable(ctx, delay); err != nil {
			return err
		}
	}

	return nil
}

func defaultMockSteps(userGoal string) []MockStep {
	goal := strings.TrimSpace(userGoal)
	if goal == "" {
		goal = "(no goal provided)"
	}

	return []MockStep{
		{Type: MockThought, Message: fmt.Sprintf("Thought: I will outline a quick plan for: %s", goal)},
		{Type: MockWarning, Message: "This action will stop container 'my-container'"},
		{Type: MockThought, Message: "Action: Identify required containers and images"},
		{Type: MockThought, Message: "Observation: A Dockerfile and compose file are likely needed"},
		{Type: MockThought, Message: "Action: Draft steps for building and running the service"},
		{Type: MockFinal, Message: "Here is a short, actionable plan you can follow"},
	}
}
