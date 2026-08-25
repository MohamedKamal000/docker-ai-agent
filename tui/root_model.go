package tui

import (
	tea "charm.land/bubbletea/v2"
	"docker-cli/internal/app"
	"docker-cli/tui/screens"
	"docker-cli/tui/screens/chat_session"
)

type RootModel struct {
	current tea.Model
	agent   *app.Agent
	runner  *AgentRunner
	width   int
	height  int
}

func NewRootModel(agent *app.Agent) *RootModel {
	return &RootModel{
		current: screens.NewHomeModel(),
		agent:   agent,
	}
}

// SetProgram wires the program handle; the agent runner needs it to push
// events into the main loop from its goroutines.
func (m *RootModel) SetProgram(program *tea.Program) {
	m.runner = NewAgentRunner(m.agent, program)
}

func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestWindowSize,
		m.current.Init(),
	)
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.current = chat_session.NewChatSessionModel()
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			}
		}
	case screens.SendMessageRequest:
		m.runner.Start(msg.UserMessage)
	case screens.ConfirmMessage:
		m.runner.Confirm(msg.ShouldContinue)
	case screens.CancelAgentRequest:
		m.runner.Cancel()
	}

	updatedModel, cmd := m.current.Update(msg)
	m.current = updatedModel
	return m, cmd
}

func (m *RootModel) View() tea.View {
	return m.current.View()
}
