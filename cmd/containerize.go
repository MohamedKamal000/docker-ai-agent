package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var containerizeCmd = &cobra.Command{
	Use:     "containerize",
	Aliases: []string{"c"},
	Short:   "let the ai scan your current directory and generate a docker file for you",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Not made yet")
	},
}

func init() {
	rootCmd.AddCommand(containerizeCmd)
}
