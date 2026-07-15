package tui

import (
	"context"

	"docker-cli/internal/app"
	"docker-cli/internal/core"
	"docker-cli/tui/screens"

	tea "charm.land/bubbletea/v2"
)

type RootModel struct {
	current    tea.Model
	agent      *app.Agent
	program    *tea.Program // a reference to the program to send messages from the agent to the main loop
	comm       *core.AgentCommunication
	cancelFunc context.CancelFunc
	width      int
	height     int
}

func NewRootModel(agent *app.Agent) *RootModel {
	return &RootModel{
		current: screens.NewHomeModel(),
		agent:   agent,
	}
}

func (m *RootModel) SetProgram(program *tea.Program) {
	m.program = program
}

func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestWindowSize,
		m.current.Init(),
	)
}

func (m *RootModel) sendMessageToAgent(ctx context.Context, userRequest string) {
	m.comm = &core.AgentCommunication{
		ToUser:   make(chan core.AiResponse),
		FromUser: make(chan core.UserCommand),
	}
	defer func() {
		close(m.comm.FromUser)
		m.comm = nil
		m.cancelFunc = nil
	}()
	go func() {
		if err := m.agent.AgentLoop.Run(ctx, userRequest, m.comm); err != nil {
			m.program.Send(screens.ErrorMessage{Message: err.Error()})
		}
	}()

	for output := range m.comm.ToUser {
		m.program.Send(screens.ReceiveMessageResponse{Response: output})
	}
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
		m.current = screens.NewChatSessionModel()
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height,
			}
		}
	case screens.SendMessageRequest:
		var ctx context.Context
		ctx, m.cancelFunc = context.WithCancel(context.Background()) // make a switch message late for cancelation
		go m.sendMessageToAgent(ctx, msg.UserMessage)
	case screens.ConfirmMessage:
		if m.comm != nil {
			m.comm.FromUser <- core.UserCommand{ShouldContinue: msg.ShouldContinue}
		}
	}

	updatedModel, cmd := m.current.Update(msg)
	m.current = updatedModel
	return m, cmd
}

func (m *RootModel) View() tea.View {
	return m.current.View()
}
