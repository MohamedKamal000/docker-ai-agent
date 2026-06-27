package common

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func RenderWithBorderForDebug(text string) string {
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(text)
}

// needs to be refactored later in a dedicated widget
// since we will have tokens usage and context
func RenderStatusLineBorder(width int, model string) string {
	left := FCyan.Render("Ctrl+O Options")
	right := FCyan.Render(fmt.Sprintf("● %s", model))

	status := lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		lipgloss.PlaceHorizontal(
			width-(lipgloss.Width(left)+lipgloss.Width(right)),
			lipgloss.Right,
			right,
		),
	)

	statusStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false)

	return statusStyle.Width(width).Render(status)
}

func AdjustTextForBackground(text string, width, height int) string {
	var stringBuilder strings.Builder
	textLength := len(text)
	stringBuilder.WriteString(text)
	stringBuilder.WriteString(strings.Repeat(" ", width-(textLength%width)))
	remainingHeight := stringBuilder.Len() % width
	line := strings.Repeat(" ", width)
	stringBuilder.WriteString(strings.Repeat("\n"+line, height-remainingHeight))
	return stringBuilder.String()
}
