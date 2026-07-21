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
		message = common.FMobyBlue.Render("Agent: ") + message
		c.stateManager.SwitchTo(NormalState.Value())
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
	case screens.ErrorMessage:
		c.appendNewMessage(common.RenderWarningBody(msg.Message, c.viewPort.Width()))
		c.stateManager.SwitchTo(NormalState.Value())
	case screens.ReceiveMessageResponse:
		c.sendAiMessage(msg.Response.Message, msg.Response.Type)
		return c, c.spinner.Tick
	}

	return c, c.updateChildren(msg)
}

func AgentRunningStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
	line := common.RenderStatusLineBorder(m.width, "model name")
	spin := fmt.Sprintf("%s Agent is Running", m.spinner.View())
	chatBox := lipgloss.Place(
		m.width,
		m.height-m.viewPort.Height(),
		lipgloss.Left,
		lipgloss.Bottom,
		lipgloss.JoinVertical(lipgloss.Left, lipgloss.NewStyle().PaddingBottom(1).Render(spin), m.ta.View(), line),
	)

	content := m.viewPort.View() + "\n" + chatBox

	return tea.NewView(content)
}
