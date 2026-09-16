package core

import (
	"context"
	"encoding/json"

	"docker-cli/internal/models"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type EvaluatorInput struct {
	Goal    string              `json:"goal"`
	History []models.HistoryEntry `json:"history"`
}

type EvaluatorResult struct {
	GoalAccomplished bool   `json:"goal_accomplished"`
	FinalResponse    string `json:"final_response,omitempty"`
}

type EvaluatorFlow = *core.Flow[EvaluatorInput, *ai.ModelResponse, struct{}]

type GoalEvaluator interface {
	Evaluate(ctx context.Context, input EvaluatorInput) (EvaluatorResult, error)
}

type GenkitGoalEvaluator struct {
	Client       GenkitClient
	Flow         EvaluatorFlow
	SystemPrompt string
}

func NewGoalEvaluator(client GenkitClient, registry ToolRegistry, systemPrompt string) *GenkitGoalEvaluator {
	refs := make([]ai.ToolRef, 0)
	for _, t := range registry.List() {
		refs = append(refs, ai.ToolName(t.Name()))
	}

	flow := genkit.DefineFlow(client.G, "GoalEvaluator",
		func(ctx context.Context, input EvaluatorInput) (*ai.ModelResponse, error) {
			parsedPrompt, err := ParsePrompt(Evaluator_Prompt_Template, input)
			if err != nil {
				return nil, err
			}

			resp, err := genkit.Generate(ctx, client.G,
				ai.WithModelName(client.Config.ModelName),
				ai.WithSystem(systemPrompt),
				ai.WithPrompt(parsedPrompt),
				ai.WithTools(refs...),
			)
			if err != nil {
				return nil, err
			}
			return resp, nil
		})

	return &GenkitGoalEvaluator{
		Client:       client,
		Flow:         flow,
		SystemPrompt: systemPrompt,
	}
}

func (e *GenkitGoalEvaluator) Evaluate(ctx context.Context, input EvaluatorInput) (EvaluatorResult, error) {
	resp, err := e.Flow.Run(ctx, input)
	if err != nil {
		return EvaluatorResult{GoalAccomplished: false}, err
	}

	raw := resp.Text()
	jsonStr := extractJSON(raw)
	if jsonStr != "" {
		var result EvaluatorResult
		if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
			return result, nil
		}
	}

	return EvaluatorResult{GoalAccomplished: false}, nil
}
