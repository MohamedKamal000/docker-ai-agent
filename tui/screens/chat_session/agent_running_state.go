package chat_session

import (
	"fmt"

	"docker-cli/internal/core"
	"docker-cli/tui/common"
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (c *ChatSessionModel) sendAiMessage(message string, responseType core.ResponseType, toolData *core.ToolExecutionData) {
	switch responseType {
	case core.FinalResponse:
		c.appendNewMessage(common.RenderAiMessage(message, c.viewPort.Width()))
	case core.Thoughts:
		c.appendNewMessage(common.RenderThought(message, c.viewPort.Width()))
	case core.Warning:
		c.pendingWarningMessage = message
		c.appendNewMessage(common.RenderWarning(message, c.viewPort.Width()))
		c.stateManager.SwitchTo(ShowWarningState.Value())
	case core.Retrying:
		c.appendNewMessage(common.StyleWarning.Render("⟳ " + message))
	case core.Error:
		c.appendNewMessage(common.RenderError(message, c.viewPort.Width()))
	case core.ToolExecution:
		if toolData != nil {
			c.trackToolMessage(toolData)
			return
		}
	}
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
			c.sendAiMessage(msg.Data.Message, msg.Data.Type, msg.Data.ToolData)
			return c, c.spinner.Tick
		case screens.RunFailed:
			c.appendNewMessage(common.RenderWarning(msg.Text, c.viewPort.Width()))
			c.stateManager.SwitchTo(NormalState.Value())
		case screens.RunCanceled:
			c.appendNewMessage(common.RenderAiMessage("Request canceled.", c.viewPort.Width()))
			c.stateManager.SwitchTo(NormalState.Value())
		case screens.RunFinished:
			c.stateManager.SwitchTo(NormalState.Value())
		}
	}

	return c, c.updateChildren(msg)
}

func AgentRunningStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
	left := fmt.Sprintf("%s %s", m.spinner.View(), common.StyleTextPrimary.Render("Thinking..."))
	right := common.StyleTextMuted.Render("esc to cancel")

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
