package chat_session

import (
	"docker-cli/tui/common"
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
)

func NormalStateExecute(s *common.StateManager[*ChatSessionModel], c *ChatSessionModel, msg tea.Msg) (*ChatSessionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			idx := c.messageIndexAtClick(msg.Y)
			if idx >= 0 {
				c.toggleToolMessage(idx)
			}
			return c, nil
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
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
			s.SwitchTo(optionsMenuState.Value())
			return c, nil
		case "ctrl+b":
			toggleCmd := c.sidebar.Toggle()
			if c.sidebar.IsVisible() {
				s.SwitchTo(sidebarState.Value())
			}
			return c, toggleCmd
		}
	}
	return c, c.updateChildren(msg)
}

func NormalStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
	return renderChatScreen(m)
}
