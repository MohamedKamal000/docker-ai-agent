package core

import (
	"context"
	"docker-cli/internal/models"
	"encoding/json"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
)

type AgentLoop interface {
	Run(ctx context.Context, userGoal string, outputChannel chan<- string) error
}
type LoopContext struct {
	Memory MemoryStore
	Tools  ToolRegistry
}

type GenkitAgentLoop struct {
	Client         GenkitClient
	Flow           AgentFlow
	SessionContext *LoopContext
}

func NewGenkitAgentLoop(client GenkitClient, sessionContext *LoopContext, systemPrompt string) *GenkitAgentLoop {
	flow := NewDockerAgentFlow(client, sessionContext.Tools, systemPrompt)
	return &GenkitAgentLoop{Client: client, SessionContext: sessionContext, Flow: flow}
}

func formatAiAction(action models.AgentResult) string {
	if action.IsStructured {
		return "Thought: \n" + action.Structured.Thought + "\n"
	}
	return "Unstructured output: \n" + action.Raw + "\n"
}

func extractResult(resp *ai.ModelResponse) models.AgentResult {
	if resp == nil {
		return models.AgentResult{
			Raw: "empty response",
		}
	}

	raw := resp.Text()

	var step models.AgentExecutionStep
	err := json.Unmarshal([]byte(raw), &step)

	if err == nil {
		return models.AgentResult{
			Structured:   &step,
			IsStructured: true,
		}
	}

	return models.AgentResult{
		Raw:          raw,
		IsStructured: false,
	}
}

func (gal *GenkitAgentLoop) Run(ctx context.Context, userGoal string, writeChannel chan<- string) error {
	defer close(writeChannel)
	toolExec := NewToolExecutor(gal.SessionContext.Tools)
	previousChat, err := gal.SessionContext.Memory.Load()
	if err != nil {
		return err
	}
	maxSteps := 3
	step := 0
	userInput := models.UserInputPrompt{
		Goal:                userGoal,
		CurrentGoalProgress: make([]models.AgentResult, 0),
		PreviousChat:        previousChat,
		ToolsExecuted:       map[string]string{},
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if step >= maxSteps {
			break
		}
		step++

		aiStep, err := gal.Flow.Run(ctx, userInput)
		if err != nil {
			slog.Error("failed to run the flow", "err", err)
			return err
		}

		agentOutput := extractResult(aiStep)

		if agentOutput.IsStructured && agentOutput.Structured.Done {
			writeChannel <- "final ai response: " + agentOutput.Structured.FinalResponse + "\n"
			err = gal.SessionContext.Memory.Save(userGoal, userInput.ToolsExecuted, userInput.CurrentGoalProgress)
			break
		}

		toolsResult, err := toolExec.ExecuteGenkitTool(ctx, aiStep)
		if err != nil {
			return err
		}

		writeChannel <- formatAiAction(agentOutput)
		userInput.CurrentGoalProgress = append(userInput.CurrentGoalProgress, agentOutput)
		for k, v := range toolsResult {
			userInput.ToolsExecuted[k] = v
		}
	}

	return nil
}
