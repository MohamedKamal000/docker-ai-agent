package main

import (
	"context"
	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"fmt"
	"log"
	"os"
)

func ReadAgentOutput(ctx context.Context, agent *app.Agent) {
	outputChannel := make(chan string)
	go func() {
		err := agent.AgentLoop.Run(ctx, "can you list all the containers i have ?", outputChannel)
		if err != nil {
			log.Fatal(err)
		}
	}()

	for output := range outputChannel {
		fmt.Printf("Received from agent: %s\n", output)
	}
}

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	config := core.ModelConfig{
		Provider:  core.Gemini,
		ModelName: "googleai/gemini-2.5-flash-lite",
		ApiKey:    apiKey,
	}
	agent := app.NewAgent(config, ctx, []string{
		"docker_command_tool",
	})
	ReadAgentOutput(ctx, agent)

}
