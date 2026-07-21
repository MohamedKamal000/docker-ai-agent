package chat_session

import (
	"docker-cli/tui/common"
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func NormalStateExecute(s *common.StateManager[*ChatSessionModel], c *ChatSessionModel, msg tea.Msg) (*ChatSessionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if c.ta.Value() == "" {
				return c, nil
			}

			sent := c.ta.Value()
			c.sendUserMessage(sent)
			s.SwitchTo(AgentRunningState.Value())
			return c, tea.Batch(
				c.spinner.Tick,
				c.updateChildren(msg),
				func() tea.Msg {
					return screens.SendMessageRequest{
						UserMessage: sent,
					}
				},
			)
		case "ctrl+o":
			c.showMenu = true
			// s.SwitchTo(OptionsMenuState.Value()) // needs to make an overlay models or something for this so we make it appear
			return c, nil
		}
	}
	return c, c.updateChildren(msg)
}

func NormalStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
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
