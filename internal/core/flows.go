package core

import (
	"context"
	"docker-cli/internal/models"
	"log"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

func NewDockerAgentFlow(client GenkitClient) *core.Flow[models.UserInputPrompt, models.ContextStep, struct{}] {
	agentFlow := genkit.DefineFlow[models.UserInputPrompt, models.ContextStep](client.G, "DockerAgent", func(ctx context.Context, input models.UserInputPrompt) (models.ContextStep, error) {
		// need to initialize the system prompt here as well
		parsedPrompt, err := ParsePrompt(User_Prompt_Template, input)
		if err != nil {
			log.Fatal(err)
		}
		// don't forget to use define tools here and pass all tools to it
		// we will also use this https://genkit.dev/docs/go/tool-calling/#explicitly-handling-tool-calls
		// to explicitly handle tool calls
		resp, err := genkit.Generate(ctx, client.G, ai.WithPrompt(parsedPrompt), ai.WithOutputType(models.ContextAction{}))

		if err != nil {
			return models.ContextStep{}, err
		}

		var action *models.ContextAction
		err = resp.Output(action)
		if err != nil {
			return models.ContextStep{}, err
		}

		return models.ContextStep{
			Result:      "leave empty for now",
			TakenAction: *action,
		}, nil
	})

	return agentFlow
}
