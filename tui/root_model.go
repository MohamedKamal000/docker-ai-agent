package tui

import (
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
)

type RootModel struct {
	current tea.Model
	width   int
	height  int
}

func NewRootModel() *RootModel {
	return &RootModel{current: screens.NewHomeModel()}
}

func (m RootModel) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestWindowSize,
		m.current.Init(),
	)
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case screens.SwitchToNewSessionMessage:
		m.current = screens.NewChatSessionModel()

		m.current, _ = m.current.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})

		return m, nil
	}

	updatedModel, cmd := m.current.Update(msg)
	m.current = updatedModel
	return m, cmd
}

func (m RootModel) View() tea.View {

	return m.current.View()
}
