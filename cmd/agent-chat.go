package cmd

import (
	"context"
	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"docker-cli/tui"
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

var configPath string
var userRequest string

func ReadAgentOutput(ctx context.Context, agent *app.Agent) {
	commChannel := &core.AgentCommunication{
		ToUser:   make(chan core.AiResponse),
		FromUser: make(chan core.UserCommand),
	}
	defer close(commChannel.FromUser)
	go func() {
		if err := agent.AgentLoop.Run(ctx, userRequest, commChannel); err != nil {
			fmt.Printf("Agent error: %v\n", err)
		}
	}()

	for output := range commChannel.ToUser {
		switch output.Type {
		case core.Thoughts:
			fmt.Printf("Agent:%s\n", output.Message)
		case core.FinalResponse:
			fmt.Printf("Agent Final Response:%s\n", output.Message)
		case core.Warning:
			fmt.Printf("WARNING \n %s", output.Message)
			fmt.Printf("Do you wish to continue ? (y,n)\n")
			var conf string
			fmt.Scan(&conf)
			commChannel.FromUser <- core.UserCommand{ShouldContinue: conf == "yes" || conf == "y"}
		}
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
		rootModel := tui.NewRootModel(agent)
		p := tea.NewProgram(rootModel)
		rootModel.SetProgram(p)

		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(agent_chat)

	agent_chat.Flags().StringVarP(&configPath, "config", "c", "current directory with file name being config.json", "path to config file")
	agent_chat.Flags().StringVarP(&userRequest, "request", "m", "", "docker ai request to make")
}
