package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var optimizeCmd = &cobra.Command{
	Use:     "optimize",
	Aliases: []string{"o"},
	Short:   "use static analysis for docker files to see what options you can do to make your docker file better",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Not made yet")
	},
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
}
