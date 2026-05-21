package main

import (
	"context"
	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"fmt"
)

func ReadAgentOutput(ctx context.Context, agent *app.Agent) {
	outputChannel := make(chan string)
	go func() {
		err := agent.AgentLoop.Run(ctx, "List all running containers and their statuses.", outputChannel)
		if err != nil {
			// handle error
			return
		}
	}()

	for output := range outputChannel {
		fmt.Printf("Received from agent: %s\n", output)
	}
}

func main() {
	ctx := context.Background()

	config := core.ModelConfig{
		ModelName: "gpt-4",
		// other config fields
	}
	agent := app.NewMockAgent(config, ctx)
	ReadAgentOutput(ctx, agent)
}
