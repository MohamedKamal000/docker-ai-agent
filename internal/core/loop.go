package core

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"docker-cli/internal/models"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type AgentLoop interface {
	Run(ctx context.Context, userGoal string, comm *AgentCommunication) error
}
type LoopContext struct {
	Memory MemoryStore
	Tools  ToolRegistry
	Tasks  *TaskRegistry
}

type GenkitAgentLoop struct {
	Client         GenkitClient
	Flow           AgentFlow
	SessionContext *LoopContext
	Classifier     IntentClassifier
}

func NewGenkitAgentLoop(client GenkitClient, sessionContext *LoopContext, systemPrompt string, classifier IntentClassifier) *GenkitAgentLoop {
	flow := NewDockerAgentFlow(client, sessionContext.Tools, systemPrompt)
	return &GenkitAgentLoop{Client: client, SessionContext: sessionContext, Flow: flow, Classifier: classifier}
}

var (
	jsonBlockRe = regexp.MustCompile("(?s)```json\\s*(\\{.*?})\\s*```")
	jsonLooseRe = regexp.MustCompile("(?s)(\\{.*})")
)

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

	classification, err := gal.Classifier.Classify(ctx, userGoal)
	if err != nil {
		comm.ToUser <- NewError(err.Error())
		return nil
	}

	switch classification.Intent {
	case IntentAmbiguous:
		comm.ToUser <- NewFinal(models.AgentResult{
			Structured: &models.AgentExecutionStep{
				FinalResponse: "Your request is unclear. Are you asking for information (e.g., 'how do I...') or wanting me to perform an action on your Docker environment (e.g., 'run nginx')? Please clarify.",
				Done:          true,
			},
			IsStructured: true,
		})
		return nil

	case IntentGeneralQuestion:
		return gal.answerGeneralQuestion(ctx, classification.RewrittenPrompt, comm)

	case IntentActionRequest:
		userGoal = classification.RewrittenPrompt
	}

	toolExec := NewToolExecutor(gal.SessionContext.Tools)
	previousHistory, err := gal.SessionContext.Memory.Load()
	if err != nil {
		return err
	}

	history := make([]models.HistoryEntry, len(previousHistory))
	copy(history, previousHistory)
	nextRun := len(history) + 1

	step := 0
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

		completedTasks := map[string]*TaskRecord{}
		if gal.SessionContext.Tasks != nil {
			for _, t := range gal.SessionContext.Tasks.PullCompleted() {
				completedTasks[t.ID] = t
			}
		}

		userInput := models.UserInputPrompt{
			Goal:    userGoal,
			History: history,
		}

		aiStep, err := gal.Flow.Run(ctx, userInput)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				comm.ToUser <- NewRetryingMessage("retrying in 5 second ...")
				if err := SleepCancellable(ctx, 5*time.Second); err != nil {
					return err
				}
				continue
			} else if strings.Contains(err.Error(), "429") {
				comm.ToUser <- NewError("Rate limit exceeded. Please wait a moment and try again.")
				return nil
			}
			comm.ToUser <- NewError(err.Error())
			return nil
		}

		agentOutput := extractResult(aiStep)

		if agentOutput.IsStructured && agentOutput.Structured.Done {
			entry := models.HistoryEntry{
				Run:           nextRun,
				Goal:          userGoal,
				FinalResponse: agentOutput.Structured.FinalResponse,
				Done:          true,
			}
			if agentOutput.Structured.Thought != "" {
				entry.Thought = agentOutput.Structured.Thought
			}
			history = append(history, entry)
			comm.ToUser <- NewFinal(agentOutput)
			err = gal.SessionContext.Memory.Save(history)
			break
		}

		_, err = toolExec.ExecuteGenkitTool(ctx, aiStep, comm)
		if err != nil {
			return err
		}

		entry := models.HistoryEntry{
			Run:  nextRun,
			Goal: userGoal,
		}
		if agentOutput.IsStructured && agentOutput.Structured.Thought != "" {
			entry.Thought = agentOutput.Structured.Thought
			comm.ToUser <- NewThought(agentOutput)
		}

		if gal.SessionContext.Tasks != nil {
			for _, t := range gal.SessionContext.Tasks.PullCompleted() {
				entry.ToolCalls = append(entry.ToolCalls, models.ToolCallInfo{
					ToolName: t.Tool,
					Command:  t.Input,
					Status:   string(t.Status),
					Result:   formatTaskResult(t),
				})
			}
			for _, t := range gal.SessionContext.Tasks.PullRunning() {
				entry.ToolCalls = append(entry.ToolCalls, models.ToolCallInfo{
					ToolName: t.Tool,
					Command:  t.Input,
					Status:   string(t.Status),
				})
			}
		}

		history = append(history, entry)
		nextRun++
	}

	return nil
}

func (gal *GenkitAgentLoop) answerGeneralQuestion(ctx context.Context, prompt string, comm *AgentCommunication) error {
	resp, err := genkit.Generate(ctx, gal.Client.G,
		ai.WithModelName(gal.Client.Config.ModelName),
		ai.WithSystem("You are a Docker expert. Answer the user's question clearly and concisely. No tools."),
		ai.WithPrompt(prompt))
	if err != nil {
		comm.ToUser <- NewError(err.Error())
		return nil
	}
	comm.ToUser <- NewFinal(models.AgentResult{
		Structured: &models.AgentExecutionStep{
			FinalResponse: resp.Text(),
			Done:          true,
		},
		IsStructured: true,
	})
	return nil
}

func formatTaskResult(t *TaskRecord) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Status: %s\n", t.Status)
	if t.Error != "" {
		fmt.Fprintf(&b, "Error: %s\n", t.Error)
	}
	if t.Result != nil {
		fmt.Fprintf(&b, "Result: %s\n", t.Result)
	}

	return strings.TrimRight(b.String(), "\n")
}
