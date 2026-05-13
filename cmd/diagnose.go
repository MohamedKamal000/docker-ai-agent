package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var diagnoseCmd = &cobra.Command{
	Use:     "diagnose",
	Aliases: []string{"d"},
	Short:   "let the ai analyse the container using its logs and see why it is likely failing",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Not made yet")
	},
}

func init() {
	rootCmd.AddCommand(diagnoseCmd)
}
