package common

import (
	"fmt"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
)

// renderGutterLines prefixes each line with a blue gutter bar.
func renderGutterLines(body string) string {
	lines := strings.Split(body, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(StyleGutter.Render("┃"))
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

// renderBlock renders content inside a bordered box with a header.
func renderBlock(header, content string, width int, borderColor string) string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width-2).
		Padding(0, 1)

	if content == "" {
		return borderStyle.Render(header)
	}
	return borderStyle.Render(header + "\n" + content)
}

// RenderStatusLine renders the bottom status bar.
func RenderStatusLine(width int, modelName string, extras ...string) string {
	parts := []string{
		StyleTextMuted.Render("●"),
		StyleTextSecondary.Render(modelName),
	}
	for _, extra := range extras {
		parts = append(parts, StyleTextMuted.Render("│"), StyleTextSecondary.Render(extra))
	}
	left := strings.Join(parts, " ")

	hints := StyleTextMuted.Render("Ctrl+O options │ Ctrl+B sidebar")
	right := hints

	status := lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		lipgloss.PlaceHorizontal(
			width-lipgloss.Width(left)-lipgloss.Width(right),
			lipgloss.Right,
			right,
		),
	)

	return StyleStatusBar.
		Width(width).
		Render(status)
}

// RenderUserMessage renders a user message with a blue gutter.
func RenderUserMessage(message string, width int) string {
	style := lipgloss.NewStyle().
		Width(width-2).
		Padding(0, 1).
		Foreground(lipgloss.Color(ColorTextPrimary))
	return renderGutterLines(style.Render(message))
}

// RenderAiMessage renders an AI response.
func RenderAiMessage(message string, width int) string {
	stripped := StripMarkdown(message)
	style := lipgloss.NewStyle().
		Width(width).
		Foreground(lipgloss.Color(ColorTextPrimary))
	return style.Render(stripped)
}

// RenderThought renders a collapsible thought block.
func RenderThought(message string, width int) string {
	header := StyleTextMuted.Render("Thinking")
	content := StyleTextSecondary.Render(StripMarkdown(message))
	return renderBlock(header, content, width, ColorDockerDarkBlue)
}

// RenderWarning renders a destructive command warning.
func RenderWarning(message string, width int) string {
	header := StyleWarning.Render("⚠ Destructive Command")
	content := StyleTextPrimary.Render(message)
	block := renderBlock(header, content, width, ColorWarning)

	actions := lipgloss.JoinHorizontal(
		lipgloss.Left,
		StyleSuccess.Render("[Y] Confirm"),
		lipgloss.PlaceHorizontal(20, lipgloss.Right, StyleError.Render("[N] Cancel")),
	)

	return block + "\n" + actions
}

// RenderConfirmed renders a confirmed action.
func RenderConfirmed(message string, width int) string {
	header := StyleSuccess.Render("✓ Confirmed")
	return renderBlock(header, message, width, ColorSuccess)
}

// RenderRejected renders a rejected action.
func RenderRejected(message string, width int) string {
	header := StyleError.Render("✗ Rejected")
	return renderBlock(header, message, width, ColorError)
}

// RenderError renders an error message.
func RenderError(message string, width int) string {
	header := StyleError.Render("✗ Error")
	return renderBlock(header, message, width, ColorError)
}

// RenderToolExecution renders a tool call with collapsible output.
func RenderToolExecution(toolName, command, output string, expanded bool, width int) string {
	statusIcon := StyleTextMuted.Render("▶")
	toolLabel := StylePrimary.Render(toolName)
	summary := fmt.Sprintf("%s %s: %s", statusIcon, toolLabel, command)

	if !expanded {
		return renderBlock(summary, "", width, ColorBorder)
	}

	content := summary
	if output != "" {
		content = summary + "\n" + StyleTextSecondary.Render(output)
	}
	return renderBlock(summary, content, width, ColorDockerBlue)
}

// RenderToolExecutionWithStatus renders a tool call with running/completed status.
func RenderToolExecutionWithStatus(toolName, command, output string, status string, expanded bool, width int) string {
	var statusIcon string
	var borderColor string

	switch status {
	case "running":
		statusIcon = StyleWarning.Render("⠋")
		borderColor = ColorWarning
	case "completed":
		statusIcon = StyleSuccess.Render("✓")
		borderColor = ColorSuccess
	case "failed":
		statusIcon = StyleError.Render("✗")
		borderColor = ColorError
	default:
		statusIcon = StyleTextMuted.Render("▶")
		borderColor = ColorBorder
	}

	toolLabel := StylePrimary.Render(toolName)
	header := fmt.Sprintf("%s %s: %s", statusIcon, toolLabel, command)

	if !expanded && status != "running" {
		return renderBlock(header, "", width, borderColor)
	}

	content := ""
	if output != "" {
		content = StyleTextSecondary.Render(output)
	}
	return renderBlock(header, content, width, borderColor)
}

// RenderInputBox wraps the textarea's own rendered view in a styled border.
func RenderInputBox(textareaView string, focused bool, width int) string {
	borderColor := ColorBorder
	if focused {
		borderColor = ColorBorderFocused
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width - 2).
		Render(textareaView)
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
