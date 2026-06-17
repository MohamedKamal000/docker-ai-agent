package core

import (
	"context"
	"docker-cli/internal/models"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
)

type AgentLoop interface {
	Run(ctx context.Context, userGoal string, comm *AgentCommunication) error
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

var jsonBlockRe = regexp.MustCompile("(?s)```json\\s*(\\{.*?})\\s*```")
var jsonLooseRe = regexp.MustCompile("(?s)(\\{.*})")

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)

	// try ones with ``` first
	if match := jsonBlockRe.FindStringSubmatch(raw); len(match) > 1 {
		return match[1]
	}

	// fall back to normal {}
	if match := jsonLooseRe.FindStringSubmatch(raw); len(match) > 1 {
		return match[1]
	}

	return ""
}

func extractResult(resp *ai.ModelResponse) models.AgentResult {
	if resp == nil {
		return models.AgentResult{
			Raw: "empty response",
		}
	}

	raw := resp.Text()

	jsonStr := extractJSON(raw)
	if jsonStr != "" {
		var step models.AgentExecutionStep
		err := json.Unmarshal([]byte(jsonStr), &step)
		if err == nil {
			return models.AgentResult{
				Structured:   &step,
				IsStructured: true,
			}
		}
	}

	return models.AgentResult{
		Raw:          raw,
		IsStructured: false,
	}
}

func (gal *GenkitAgentLoop) Run(ctx context.Context, userGoal string, comm *AgentCommunication) error {
	defer close(comm.ToUser)
	toolExec := NewToolExecutor(gal.SessionContext.Tools)
	previousChat, err := gal.SessionContext.Memory.Load()
	if err != nil {
		return err
	}
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
		if step >= gal.Client.Config.MaxIterations {
			break
		}
		step++
		if step != 0 {
			time.Sleep(3 * time.Second) // waits 3 sec on purpose if not first call
		}
		aiStep, err := gal.Flow.Run(ctx, userInput)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				fmt.Println("retrying in 5 second ...")
				time.Sleep(5 * time.Second)
				continue
			} else if strings.Contains(err.Error(), "429") {
				return fmt.Errorf("you have exceeded the limit")
			}
			return err
		}

		agentOutput := extractResult(aiStep)
		if agentOutput.IsStructured && agentOutput.Structured.Done {
			comm.ToUser <- NewFinal(agentOutput)
			err = gal.SessionContext.Memory.Save(userGoal, userInput.ToolsExecuted, userInput.CurrentGoalProgress)
			break
		}

		toolsResult, err := toolExec.ExecuteGenkitTool(ctx, aiStep, comm)
		if err != nil {
			return err
		}

		comm.ToUser <- NewThought(agentOutput)
		userInput.CurrentGoalProgress = append(userInput.CurrentGoalProgress, agentOutput)
		for k, v := range toolsResult {
			userInput.ToolsExecuted[k] = v
		}
	}

	return nil
}
