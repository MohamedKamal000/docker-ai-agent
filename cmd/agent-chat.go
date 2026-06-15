package cmd

import (
	"context"
	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var configPath string
var userRequest string

type DockerResponse struct {
	UserDockerAnswer string `json:"user-docker-answer"`
}

func ReadAgentOutput(ctx context.Context, agent *app.Agent) {
	outputChannel := make(chan string)
	go func() {
		err := agent.AgentLoop.Run(ctx, userRequest, outputChannel)
		if err != nil {
			log.Fatal(err)
		}
	}()

	for output := range outputChannel {
		fmt.Printf("Received from agent: %s\n", output)
	}
}

var agent_chat = &cobra.Command{
	Use:     "agent-chat",
	Aliases: []string{"ac"},
	Short:   "Agent chat mode, ask it anything related to docker",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if userRequest == "" {
			fmt.Println("no request was provided, request can't be empty")
			return
		}
		ctx := context.Background()

		config, err := core.ModelConfigFromJsonFile(configPath)
		if err != nil {
			log.Fatal(err)
		}

		// needs a way to make the user choose the tools to use before running this command
		agent, err := app.NewAgent(config, ctx, []string{
			"docker_command_tool",
		})

		if err != nil {
			log.Fatalf("failed to initalize agent, Err:%v", err)
		}
		ReadAgentOutput(ctx, agent)
	},
}

func init() {
	rootCmd.AddCommand(agent_chat)

	agent_chat.Flags().StringVarP(&configPath, "config", "c", "current directory with file name being config.json", "path to config file")
	agent_chat.Flags().StringVarP(&userRequest, "request", "m", "", "docker ai request to make")
}
