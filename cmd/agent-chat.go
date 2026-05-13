package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/spf13/cobra"
)

type DockerUserQuestion struct {
	UserQuestion string `json:"docker-user-question"`
}

type DockerResponse struct {
	UserDockerAnswer string `json:"user-docker-answer"`
}

var agent_chat = &cobra.Command{
	Use:     "agent-chat",
	Aliases: []string{"ac"},
	Short:   "Agent chat mode, ask it anything related to docker",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		gemeni_key := os.Getenv("GEMINI_KEY")
		g := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: gemeni_key}), genkit.WithDefaultModel("googleai/gemini-2.5-flash"))

		questionFlow := genkit.DefineFlow(g, "DockerQuestion", func(ctx context.Context, input DockerUserQuestion) (DockerResponse, error) {
			resp, err := genkit.Generate(ctx, g, ai.WithPrompt("Answer the user docker related question, user question: %s", input.UserQuestion))

			if err != nil {
				return DockerResponse{UserDockerAnswer: "ai didn't respond"}, err
			}
			return DockerResponse{UserDockerAnswer: resp.Text()}, nil
		})

		res, err := questionFlow.Run(ctx, DockerUserQuestion{UserQuestion: args[0]})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Ai Response %s\n", res.UserDockerAnswer)
	},
}

func init() {
	rootCmd.AddCommand(agent_chat)
}
