package chat_session

import (
	"fmt"

	"docker-cli/internal/core"
	"docker-cli/tui/common"
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (c *ChatSessionModel) sendAiMessage(message string, responseType core.ResponseType) {
	switch responseType {
	case core.FinalResponse:
		// Input is unlocked by the terminal RunFinished event, not here.
		message = common.FMobyBlue.Render("Agent: ") + message
	case core.Thoughts:
		message = common.FMobyBlue.Render("Thoughts: ") + message
	case core.Warning:
		c.pendingWarningMessage = message
		message = common.RenderWarningMessage(message, c.viewPort.Width())
		c.stateManager.SwitchTo(ShowWarningState.Value())
	case core.Retrying:
		// do nothing for now, we might add some ui for it later, but the current spinner does the job
		return
	}
	c.appendNewMessage(message)
}

func AgentRunningStateExecute(s *common.StateManager[*ChatSessionModel], c *ChatSessionModel, msg tea.Msg) (*ChatSessionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "esc" {
			return c, func() tea.Msg {
				return screens.CancelAgentRequest{}
			}
		}
	case screens.RunEvent:
		switch msg.Kind {
		case screens.RunResponse:
			c.sendAiMessage(msg.Data.Message, msg.Data.Type)
			return c, c.spinner.Tick
		case screens.RunFailed:
			c.appendNewMessage(common.RenderWarningBody(msg.Text, c.viewPort.Width()))
			c.stateManager.SwitchTo(NormalState.Value())
		case screens.RunCanceled:
			c.appendNewMessage(common.FMobyBlue.Render("Agent: ") + "Request canceled.")
			c.stateManager.SwitchTo(NormalState.Value())
		case screens.RunFinished:
			c.stateManager.SwitchTo(NormalState.Value())
		}
	}

	return c, c.updateChildren(msg)
}

func AgentRunningStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
	left := fmt.Sprintf("%s Agent is Running", m.spinner.View())
	right := common.FCyan.Render("esc Cancel")
	status := lipgloss.JoinHorizontal(
		lipgloss.Left,
		left,
		lipgloss.PlaceHorizontal(
			m.width-(lipgloss.Width(left)+lipgloss.Width(right)),
			lipgloss.Right,
			right,
		),
	)

	return renderChatScreen(m, lipgloss.NewStyle().PaddingBottom(1).Render(status))
}
