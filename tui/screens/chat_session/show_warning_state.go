package chat_session

import (
	"docker-cli/tui/common"
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func ShowWarningStateExecute(s *common.StateManager[*ChatSessionModel], c *ChatSessionModel, msg tea.Msg) (*ChatSessionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		confirmation := msg.String() == "y"
		if msg.String() == "y" || msg.String() == "n" {
			s.SwitchTo(AgentRunningState.Value())
			c.confirmMessage(confirmation)
			return c, func() tea.Msg {
				return screens.ConfirmMessage{
					ShouldContinue: confirmation,
				}
			}
		}
	}
	return c, c.updateChildren(msg)
}

func ShowWarningStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
	line := common.RenderStatusLineBorder(m.width, "model name")
	chatBox := lipgloss.Place(
		m.width,
		m.height-m.viewPort.Height(),
		lipgloss.Left,
		lipgloss.Bottom,
		lipgloss.JoinVertical(lipgloss.Left, m.ta.View(), line),
	)

	content := m.viewPort.View() + "\n" + chatBox
	return tea.NewView(content)
}
