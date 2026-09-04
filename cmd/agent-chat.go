package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"docker-cli/internal/rag"
	"docker-cli/tui"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

var (
	configPath string
	oneShotRun string
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
				"task_status_tool",
			},
				app.WithRagOption(ret),
			)
		} else {
			agent, err = app.NewAgent(config, ctx, []string{
				"docker_command_tool",
				"task_status_tool",
			})
			if err != nil {
				log.Fatalf("failed to initalize agent, Err:%v", err)
			}
		}

		rootModel := tui.NewRootModel(agent)
		p := tea.NewProgram(rootModel)
		rootModel.SetProgram(p)

		if oneShotRun != "" {
			fmt.Println(oneShotRun)
			comm := &core.AgentCommunication{
				ToUser:   make(chan core.AiResponse),
				FromUser: make(chan core.UserCommand),
			}
			go func() {
				err := agent.AgentLoop.Run(ctx, oneShotRun, comm)
				if err != nil {
					log.Fatal(err)
				}
			}()
			for message := range comm.ToUser {
				log.Println(message.Message)
			}

			signal := make(chan os.Signal, 1)
			<-signal
		} else {
			rootModel := tui.NewRootModel(agent)
			p := tea.NewProgram(rootModel)
			rootModel.SetProgram(p)

			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(agent_chat)

	agent_chat.Flags().StringVarP(&configPath, "config", "c", "current directory with file name being config.json", "path to config file")
	agent_chat.Flags().StringVarP(&oneShotRun, "one-shot", "o", "", "run the agent once with a goal, print the output then exit")
	agent_chat.Flags().BoolVar(&useRAG, "rag", false, "use RAG for agent chat")
}
