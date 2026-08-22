package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type Intent string

const (
	IntentGeneralQuestion Intent = "general_question"
	IntentActionRequest   Intent = "action_request"
	IntentAmbiguous       Intent = "ambiguous"
)

type ClassificationResult struct {
	Intent          Intent
	RewrittenPrompt string
}

type ClassificationResponse struct {
	Intent          string `json:"intent"`
	RewrittenPrompt string `json:"rewritten_prompt"`
}

type IntentClassifier interface {
	Classify(ctx context.Context, userInput string) (ClassificationResult, error)
}

type GenkitIntentClassifier struct {
	Client GenkitClient
	Flow   *core.Flow[IntentClassificationInput, *ai.ModelResponse, struct{}]
}

type IntentClassificationInput struct {
	UserInput string `json:"user_input"`
}

func NewIntentClassifier(client GenkitClient) *GenkitIntentClassifier {
	flow := genkit.DefineFlow(client.G, "IntentClassifier", func(ctx context.Context, input IntentClassificationInput) (*ai.ModelResponse, error) {
		parsedPrompt, err := ParsePrompt(Intent_Classification_Template, input)
		if err != nil {
			return nil, err
		}
		resp, err := genkit.Generate(ctx, client.G, ai.WithModelName(client.Config.ModelName),
			ai.WithSystem("You are a request classifier. Output only valid JSON matching the specified schema."),
			ai.WithPrompt(parsedPrompt))
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	return &GenkitIntentClassifier{Client: client, Flow: flow}
}

func (c *GenkitIntentClassifier) Classify(ctx context.Context, userInput string) (ClassificationResult, error) {
	resp, err := c.Flow.Run(ctx, IntentClassificationInput{UserInput: userInput})
	if err != nil {
		return ClassificationResult{}, err
	}

	raw := resp.Text()
	jsonStr := extractJSON(raw)
	if jsonStr == "" {
		return ClassificationResult{}, fmt.Errorf("failed to extract JSON from classification response: %s", raw)
	}

	var classification ClassificationResponse
	if err := json.Unmarshal([]byte(jsonStr), &classification); err != nil {
		return ClassificationResult{}, fmt.Errorf("failed to parse classification response: %w, raw: %s", err, raw)
	}

	intent := Intent(classification.Intent)
	if intent != IntentGeneralQuestion && intent != IntentActionRequest && intent != IntentAmbiguous {
		return ClassificationResult{}, fmt.Errorf("invalid intent: %s", classification.Intent)
	}

	return ClassificationResult{
		Intent:          intent,
		RewrittenPrompt: classification.RewrittenPrompt,
	}, nil
}

