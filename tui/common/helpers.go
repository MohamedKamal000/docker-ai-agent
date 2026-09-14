package common

import (
	"fmt"
	"regexp"
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
		Width(width-3). // 1 for ┃ + 2 spaces
		MaxWidth(width-3).
		Render(message), status)

	return renderGutterLines(message)
}

func RenderToolExecution(toolName, command, output string, expanded bool, width int) string {
	indicator := FGray.Render("▶")
	toolLabel := FYellow.Render(toolName)
	summary := fmt.Sprintf("%s %s: %s", indicator, toolLabel, command)

	if !expanded {
		return renderGutterBlock(summary, width, BTool)
	}

	detailLines := []string{summary}
	if output != "" {
		detailLines = append(detailLines, FGray.Render(output))
	}

	body := strings.Join(detailLines, "\n")
	style := BTool.Padding(1).
		Width(width - 3).
		MaxWidth(width - 3)

	return renderGutterLines(style.Render(body))
}

var (
	mdCodeBlockRe  = regexp.MustCompile("(?s)```[a-zA-Z]*\\s*\\n?(.*?)\\n?```")
	mdInlineCodeRe = regexp.MustCompile("`([^`]+)`")
	mdBoldRe       = regexp.MustCompile(`\*\*(.+?)\*\*`)
	mdItalicRe     = regexp.MustCompile(`\*(.+?)\*`)
	mdHeaderRe     = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	mdHrRe         = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
)

func StripMarkdown(text string) string {
	text = mdCodeBlockRe.ReplaceAllString(text, "$1")
	text = mdInlineCodeRe.ReplaceAllString(text, "$1")
	text = mdBoldRe.ReplaceAllString(text, "$1")
	text = mdItalicRe.ReplaceAllString(text, "$1")
	text = mdHeaderRe.ReplaceAllString(text, "")
	text = mdHrRe.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}
