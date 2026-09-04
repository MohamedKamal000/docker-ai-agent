package cmd

import (
	"context"
	"log"

	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"docker-cli/internal/rag"
	"docker-cli/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

var (
	configPath string
	useRAG     bool
)

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

		var agent *app.Agent
		if useRAG {
			ret, err := rag.NewRetriever(ctx, config)
			if err != nil {
				log.Fatal(err)
			}

			agent, err = app.NewAgent(config, ctx, []string{
				"docker_command_tool",
			},
				app.WithRagOption(ret),
			)
		} else {
			agent, err = app.NewAgent(config, ctx, []string{
				"docker_command_tool",
			})
		}

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

	agent_chat.Flags().BoolVar(&useRAG, "rag", false, "use RAG for agent chat")
}
