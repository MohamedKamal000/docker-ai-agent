package chat_session

import (
	"docker-cli/tui/common"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func SidebarStateExecute(s *common.StateManager[*ChatSessionModel], c *ChatSessionModel, msg tea.Msg) (*ChatSessionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+b":
			c.sidebar.Toggle()
			s.SwitchToPreviousState()
			return c, nil
		}
	case SidebarDockerContextLoadedMsg:
		c.sidebar.HandleMessage(msg)
		return c, nil
	}
	return c, nil
}

func SidebarStateRender(s *common.StateManager[*ChatSessionModel], m *ChatSessionModel) tea.View {
	content := s.RenderPrevious(m).Content

	sidebarView := m.sidebar.Render()
	overlayLayer := lipgloss.NewLayer(sidebarView).
		X(m.width - lipgloss.Width(sidebarView) - 2).
		Y(1).
		Z(1)
	baseLayer := lipgloss.NewLayer(content)
	content = lipgloss.NewCompositor(baseLayer, overlayLayer).Render()
	return tea.NewView(content)
}
