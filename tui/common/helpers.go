package common

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func RenderWithBorderForDebug(text string) string {
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Margin(0).Padding(0).Render(text)
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

// renderGutterLines prefixes each line of an already-rendered body with a ┃
// gutter.
func renderGutterLines(body string) string {
	lines := strings.Split(body, "\n")

	var b strings.Builder
	for _, line := range lines {
		b.WriteString(FMobyBlue.Render("┃"))
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// renderGutterBlock renders a message in a padded background block with a
// ┃ gutter prefix on each line.
func renderGutterBlock(message string, width int, bg lipgloss.Style) string {
	style := bg.Padding(1).
		Width(width - 3). // 1 for ┃ + 2 spaces
		MaxWidth(width - 3)

	return renderGutterLines(style.Render(message))
}

func RenderUserMessageWithBackground(message string, width int) string {
	return renderGutterBlock(message, width, BBlack)
}

func RenderWarningBody(message string, width int) string {
	return renderGutterBlock(message, width, BRed)
}

func RenderConfirmedMessage(message string, width int) string {
	return renderGutterBlock(message, width, BEmerald)
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

	message = lipgloss.JoinVertical(lipgloss.Left, BRed.Padding(1).
		Width(width - 3). // 1 for ┃ + 2 spaces
		MaxWidth(width - 3).
		Render(message), status)

	return renderGutterLines(message)
}
