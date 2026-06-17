package app

import (
	"context"
	"docker-cli/internal/models"
	"fmt"
	"strings"
	"time"

	"docker-cli/internal/core"
)

// MockAgent mirrors Agent for frontend development without external dependencies.

func NewMockAgent(config core.ModelConfig, ctx context.Context) *Agent {
	_ = config
	_ = ctx

	chatSession := core.NewStaticMemoryStore()

	agentLoop := &MockAgentLoop{}

	return &Agent{
		AgentLoop: agentLoop,
		SessionContext: &core.LoopContext{
			Memory: chatSession,
			Tools:  nil,
		},
	}
}

type MockAgentLoop struct {
	Steps []string
}

func (mal *MockAgentLoop) Run(ctx context.Context, userGoal string, comm *core.AgentCommunication) error {
	defer close(comm.ToUser)

	steps := mal.Steps
	if len(steps) == 0 {
		steps = defaultMockSteps(userGoal)
	}

	for _, step := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		comm.ToUser <- core.NewThought(models.AgentResult{
			IsStructured: false,
			Raw:          step,
		})
		time.Sleep(3 * time.Second) // Simulate processing time
	}

	comm.ToUser <- core.NewFinal(models.AgentResult{
		IsStructured: false,
		Raw:          "Agent has completed the goal.",
	})
	return nil
}

func defaultMockSteps(userGoal string) []string {
	goal := strings.TrimSpace(userGoal)
	if goal == "" {
		goal = "(no goal provided)"
	}

	return []string{
		fmt.Sprintf("Thought: I will outline a quick plan for: %s", goal),
		"Action: Identify required containers and images",
		"Observation: A Dockerfile and compose file are likely needed",
		"Action: Draft steps for building and running the service",
		"Final: Here is a short, actionable plan you can follow",
	}
}
