package core

import (
	"context"
	"docker-cli/internal/models"
)

// we can stream ai thoughts and also add in the prompt instructions that the ai need to write final summary as the response
type AgentLoop interface {
	Run(ctx context.Context, userGoal string, outputChannel chan<- string) error
}

type GenkitAgentLoop struct {
	Client       GenkitClient
	SystemPrompt string
}

func formatAiAction(action models.ContextAction) string {
	return "Thought: " + action.Thought + "\nAction: " + action.Action
}

func (gal *GenkitAgentLoop) Run(ctx context.Context, userGoal string, writeChannel chan<- string) error {
	defer close(writeChannel)
	flow := NewDockerAgentFlow(gal.Client)
	// we can make the memory this way for now, we need to think in the future on how to make it persistance
	memory := NewStaticMemoryStore()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		previousChat, err := memory.Load()
		if err != nil {
			return err
		}

		userInput := models.UserInputPrompt{
			Goal:    userGoal,
			Context: previousChat,
		}

		aiStep, err := flow.Run(ctx, userInput)
		if err != nil {
			return err
		}

		if aiStep.TakenAction.Done {
			writeChannel <- "Agent has completed the goal."
			break
		}

		writeChannel <- formatAiAction(aiStep.TakenAction)
		// parse tool execution and use the tool adapter that needs to be implemented
		// the parsing should determine if we will continue the loop or not

		err = memory.Save(aiStep)
		if err != nil {
			return err
		}
	}

	return nil
}
