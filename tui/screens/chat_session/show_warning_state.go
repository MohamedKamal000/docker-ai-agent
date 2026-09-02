package chat_session

import (
	"docker-cli/tui/common"
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
)

func ShowWarningStateExecute(s *common.StateManager[*ChatSessionModel], c *ChatSessionModel, msg tea.Msg) (*ChatSessionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.RunEvent:
		switch msg.Kind {
		case screens.RunFailed, screens.RunCanceled:
			c.appendNewMessage(common.RenderWarningBody(msg.Text, c.viewPort.Width()))
			c.stateManager.SwitchTo(NormalState.Value())
		case screens.RunFinished:
			c.stateManager.SwitchTo(NormalState.Value())
		}
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
	return renderChatScreen(m)
}
