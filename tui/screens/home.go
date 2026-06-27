package screens

import (
	"docker-cli/tui/common"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type HomeModel struct {
	options       []string
	currentOption int
	width         int
	height        int
}

func NewHomeModel() *HomeModel {
	return &HomeModel{
		options: []string{
			"1) Start a new session",
			"2) Choose an already exist session",
		},
		currentOption: 0,
	}
}

func (m *HomeModel) Init() tea.Cmd {
	return nil
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up":
			if m.currentOption > 0 {
				m.currentOption--
			}
		case "down":
			if m.currentOption < len(m.options)-1 {
				m.currentOption++
			}
		case "enter":
			if m.currentOption == 0 {
				return m, func() tea.Msg {
					return SwitchToNewSessionMessage{}
				}
			}
		}
	}

	return m, nil
}

func (m *HomeModel) View() tea.View {
	box := common.FWhiteBlue.
		Margin(lipgloss.Height(common.Logo)/2, m.width/2-lipgloss.Width(common.Logo)/2, 0)

	styled := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	LogoStyle := box.Render(common.Logo)
	var menu strings.Builder
	for i, opt := range m.options {
		prefix := "  "
		if i == m.currentOption {
			prefix = common.FMobyBlue.Render("> ")
			opt = common.FGreenish.Render(opt)
		}
		menu.WriteString(prefix + opt + "\n")
	}
	menuStyle := styled.Render(menu.String())

	content := LogoStyle + "\n\n" + menuStyle
	v := tea.NewView(
		lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Top,
			content,
		),
	)
	v.AltScreen = true
	return v
}
