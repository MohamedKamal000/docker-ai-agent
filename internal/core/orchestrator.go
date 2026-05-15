package core

import (
	"context"
	"docker-cli/internal/models"
	"log"
)

// we can stream ai thoughts and also add in the prompt instructions that the ai need to write final summary as the response
func RunAgent(ctx context.Context, client GenkitClient, userGoal string) error {
	flow := NewDockerAgentFlow(client)
	// we can make the memory this way for now, we need to think in the future on how to make it persistance
	memory := make([]models.ContextStep, 0)
	userInput := models.UserInputPrompt{
		Goal:    userGoal,
		Context: memory,
	}

	for {
		aiStep, err := flow.Run(ctx, userInput)
		if err != nil {
			log.Fatal(err)
		}

		// parse tool execution and use the tool adapter that needs to be implemented
		// the parsing should determine if we will continue the loop or not

		memory = append(memory, aiStep)

	}

}
