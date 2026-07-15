package core

import (
	"context"

	"docker-cli/internal/models"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

type AgentFlow = *core.Flow[models.UserInputPrompt, *ai.ModelResponse, struct{}]

func NewDockerAgentFlow(client GenkitClient, registry ToolRegistry, systemPrompt string) AgentFlow {
	refs := make([]ai.ToolRef, 0)
	for _, t := range registry.List() {
		refs = append(refs, ai.ToolName(t.Name()))
	}
	agentFlow := genkit.DefineFlow[models.UserInputPrompt, *ai.ModelResponse](client.G, "DockerAgent", func(ctx context.Context, input models.UserInputPrompt) (*ai.ModelResponse, error) {
		parsedPrompt, err := ParsePrompt(User_Prompt_Template, input)
		if err != nil {
			return nil, err
		}
		resp, err := genkit.Generate(ctx, client.G, ai.WithModelName(client.Config.ModelName),
			ai.WithTools(refs...), ai.WithSystem(systemPrompt), ai.WithPrompt(parsedPrompt), ai.WithReturnToolRequests(true))
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	return agentFlow
}
