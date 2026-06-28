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

func RenderUserMessageWithBackground(message string, width int) string {
	style := BBlack.Padding(1).
		Width(width - 3). // 1 for ┃ + 2 spaces
		MaxWidth(width - 3)

	body := style.Render(message)

	lines := strings.Split(body, "\n")

	var b strings.Builder
	for _, line := range lines {
		b.WriteString(FMobyBlue.Render("┃"))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func RenderWarningMessage(message string, width int) string {
	left := FGreenish.Render("Accept (y)")
	right := FRed.Render("Reject (n)")

	status := lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		lipgloss.PlaceHorizontal(
			width-(lipgloss.Width(left)+lipgloss.Width(right)),
			lipgloss.Right,
			right,
		),
	)

	style := BRed.Padding(1).
		Width(width - 3). // 1 for ┃ + 2 spaces
		MaxWidth(width - 3)

	body := style.Render(message)

	body = lipgloss.JoinVertical(lipgloss.Left, body, status)

	lines := strings.Split(body, "\n")

	var b strings.Builder
	for _, line := range lines {
		b.WriteString(FMobyBlue.Render("┃"))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
