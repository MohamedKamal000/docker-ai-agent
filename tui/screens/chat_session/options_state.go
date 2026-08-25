package chat_session

import (
	"docker-cli/tui/common"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func OptionsMenuStateExecute(s *common.StateManager[*ChatSessionModel], c *ChatSessionModel, msg tea.Msg) (*ChatSessionModel, tea.Cmd) {
	msg_str, ok := msg.(tea.KeyPressMsg)
	if ok && (msg_str.String() == "q" || msg_str.String() == "esc") {
		s.SwitchToPreviousState()
		return c, nil
	}
	var cmd tea.Cmd
	c.OptionsMenu, cmd = c.OptionsMenu.Update(msg)
	return c, cmd
}

func OptionsMenuStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
	content := s.RenderPrevious(m).Content

	listView := m.OptionsMenu.View()
	overlayContent := listView.Content
	overlayLayer := lipgloss.NewLayer(overlayContent).
		X((m.width - lipgloss.Width(overlayContent)) / 2).
		Y((m.height - lipgloss.Height(overlayContent)) / 2).
		Z(1)
	baseLayer := lipgloss.NewLayer(content)
	content = lipgloss.NewCompositor(baseLayer, overlayLayer).Render()
	return tea.NewView(content)
}
