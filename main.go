package main

import (
	"docker-cli/tui"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	p := tea.NewProgram(tui.NewRootModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("bye")
		os.Exit(1)
	}
	// cmd.Execute()
}
