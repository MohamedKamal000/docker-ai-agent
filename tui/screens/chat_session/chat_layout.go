package chat_session

import (
	"docker-cli/tui/common"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func renderChatScreen(m *ChatSessionModel, aboveInput ...string) tea.View {
	line := common.RenderStatusLineBorder(m.width, "model name")
	chatBox := lipgloss.Place(
		m.width,
		m.height-m.viewPort.Height(),
		lipgloss.Left,
		lipgloss.Bottom,
		lipgloss.JoinVertical(lipgloss.Left, append(aboveInput, m.ta.View(), line)...),
	)

	content := m.viewPort.View() + "\n" + chatBox
	return tea.NewView(content)
}
