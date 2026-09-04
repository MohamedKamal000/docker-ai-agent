package cmd

import (
	"context"
	"log"

	"docker-cli/internal/core"
	"docker-cli/internal/rag"

	"github.com/spf13/cobra"
)

var initalizeRag = &cobra.Command{
	Use:     "initalizeRag",
	Aliases: []string{"ir"},
	Short:   "initalizes different depenencies for the rag to work by installing the docker docs then chunking it and embedding it, then store it to an embedded database",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		config, err := core.ModelConfigFromJsonFile(configPath)
		if err != nil {
			log.Fatal(err)
		}

		err = rag.InitalizeRag(ctx, config)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(containerizeCmd)
}
