package cmd

import (
	"context"
	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"docker-cli/tui"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

var configPath string

var agent_chat = &cobra.Command{
	Use:     "agent-chat",
	Aliases: []string{"ac"},
	Short:   "Agent chat mode, ask it anything related to docker",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
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
}
