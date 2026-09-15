package screens

import (
	"context"
	"fmt"
	"strings"

	"docker-cli/internal/docker"
	"docker-cli/tui/common"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type DockerContextLoadedMsg struct {
	Context *docker.Context
	Err     error
}

type HomeModel struct {
	options       []string
	currentOption int
	width         int
	height        int
	dockerCtx     *docker.Context
	loading       bool
}

func NewHomeModel() *HomeModel {
	return &HomeModel{
		options: []string{
			"Start a new session",
			"Resume last session",
		},
		currentOption: 0,
		loading:       true,
	}
}

func (m *HomeModel) Init() tea.Cmd {
	return func() tea.Msg {
		ctx, err := docker.GetContext(context.Background())
		return DockerContextLoadedMsg{Context: ctx, Err: err}
	}
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case DockerContextLoadedMsg:
		m.dockerCtx = msg.Context
		m.loading = false
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

func (m *HomeModel) renderDockerContext() string {
	if m.loading {
		return common.StyleTextMuted.Render("Loading Docker context...")
	}

	if m.dockerCtx == nil {
		return common.StyleError.Render("Docker daemon not connected")
	}

	var sb strings.Builder

	// Container summary
	running := 0
	stopped := 0
	for _, c := range m.dockerCtx.Containers {
		if c.State == "running" {
			running++
		} else {
			stopped++
		}
	}
	fmt.Fprintf(&sb, "Containers:  %s running,  %s stopped\n",
		common.StyleSuccess.Render(fmt.Sprintf("%d", running)),
		common.StyleTextMuted.Render(fmt.Sprintf("%d", stopped)))

	// Image summary
	totalImages := len(m.dockerCtx.Images)
	totalSize := int64(0)
	for _, img := range m.dockerCtx.Images {
		totalSize += img.Size
	}
	fmt.Fprintf(&sb, "Images:      %s total,   %s\n",
		common.StyleTextSecondary.Render(fmt.Sprintf("%d", totalImages)),
		common.StyleTextMuted.Render(formatBytes(totalSize)))

	// Volume summary
	fmt.Fprintf(&sb, "Volumes:     %s active\n",
		common.StyleTextSecondary.Render(fmt.Sprintf("%d", len(m.dockerCtx.Volumes))))

	// Network summary
	fmt.Fprintf(&sb, "Networks:    %s\n",
		common.StyleTextSecondary.Render(fmt.Sprintf("%d", len(m.dockerCtx.Networks))))

	// Daemon status
	fmt.Fprintf(&sb, "Daemon:      %s Connected",
		common.StyleSuccess.Render("●"))

	return sb.String()
}

func (m *HomeModel) View() tea.View {
	// Logo
	logo := common.StylePrimary.Render(common.Logo)
	logo = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logo)

	// Subtitle
	subtitle := common.StyleTextSecondary.Render("Docker Environment AI Assistant")
	subtitle = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, subtitle)

	// Docker context card
	dockerContext := m.renderDockerContext()
	card := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(common.ColorDockerBlue)).
		Width(min(50, m.width-4)).
		Padding(1, 2).
		Render(common.StylePrimary.Render("Docker Context") + "\n\n" + dockerContext)
	card = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, card)

	// Menu
	var menu strings.Builder
	for i, opt := range m.options {
		prefix := "  "
		if i == m.currentOption {
			prefix = common.StylePrimary.Render("> ")
			opt = common.StyleAccent.Render(opt)
		}
		menu.WriteString(prefix)
		menu.WriteString(opt)
		menu.WriteString("\n")
	}
	menuStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(menu.String())

	// Footer
	footer := common.StyleTextMuted.Render("Press Enter to start • ↑↓ to navigate")
	footer = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, footer)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		"\n",
		logo,
		subtitle,
		"\n",
		card,
		"\n",
		menuStyle,
		"\n",
		footer,
	)

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

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
