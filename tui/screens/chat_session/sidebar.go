package chat_session

import (
	"context"
	"fmt"
	"strings"

	"docker-cli/internal/docker"
	"docker-cli/tui/common"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SidebarDockerContextLoadedMsg is sent when Docker context finishes loading.
type SidebarDockerContextLoadedMsg struct {
	Context *docker.Context
	Err     error
}

// SidebarModel manages the Docker context sidebar.
type SidebarModel struct {
	dockerCtx *docker.Context
	width     int
	height    int
	visible   bool
	loading   bool
}

// NewSidebarModel creates a new sidebar model.
func NewSidebarModel() *SidebarModel {
	return &SidebarModel{
		visible: false,
		width:   30,
		height:  40,
	}
}

// SetSize updates the sidebar dimensions.
func (s *SidebarModel) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// IsVisible returns whether the sidebar is currently visible.
func (s *SidebarModel) IsVisible() bool {
	return s.visible
}

// Width returns the sidebar width when visible.
func (s *SidebarModel) Width() int {
	if !s.visible {
		return 0
	}
	if s.width == 0 {
		return 28
	}
	return s.width
}

// Toggle toggles the sidebar visibility and returns a command to load Docker context.
func (s *SidebarModel) Toggle() tea.Cmd {
	s.visible = !s.visible
	if s.visible && s.dockerCtx == nil && !s.loading {
		s.loading = true
		return func() tea.Msg {
			ctx, err := docker.GetContext(context.Background())
			return SidebarDockerContextLoadedMsg{Context: ctx, Err: err}
		}
	}
	return nil
}

// HandleMessage handles messages for the sidebar.
func (s *SidebarModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case SidebarDockerContextLoadedMsg:
		s.loading = false
		if msg.Err == nil {
			s.dockerCtx = msg.Context
		}
		return nil
	}
	return nil
}

// Render renders the sidebar content.
func (s *SidebarModel) Render() string {
	if !s.visible {
		return ""
	}

	var sb strings.Builder

	// Header
	header := common.StylePrimary.Render("Docker Context")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	if s.loading {
		sb.WriteString(common.StyleTextMuted.Render("Loading..."))
		return s.wrapInBorder(sb.String())
	}

	if s.dockerCtx == nil {
		sb.WriteString(common.StyleError.Render("Daemon not connected"))
		return s.wrapInBorder(sb.String())
	}

	// Containers section
	sb.WriteString(common.StyleAccent.Render("Containers"))
	sb.WriteString("\n")
	running := 0
	stopped := 0
	for _, c := range s.dockerCtx.Containers {
		if c.State == "running" {
			running++
		} else {
			stopped++
		}
	}
	sb.WriteString(fmt.Sprintf("  %s running\n", common.StyleSuccess.Render(fmt.Sprintf("%d", running))))
	sb.WriteString(fmt.Sprintf("  %s stopped\n", common.StyleTextMuted.Render(fmt.Sprintf("%d", stopped))))
	sb.WriteString("\n")

	// Images section
	sb.WriteString(common.StyleAccent.Render("Images"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s total\n", common.StyleTextSecondary.Render(fmt.Sprintf("%d", len(s.dockerCtx.Images)))))
	sb.WriteString("\n")

	// Volumes section
	sb.WriteString(common.StyleAccent.Render("Volumes"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s active\n", common.StyleTextSecondary.Render(fmt.Sprintf("%d", len(s.dockerCtx.Volumes)))))
	sb.WriteString("\n")

	// Networks section
	sb.WriteString(common.StyleAccent.Render("Networks"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s\n", common.StyleTextSecondary.Render(fmt.Sprintf("%d", len(s.dockerCtx.Networks)))))

	return s.wrapInBorder(sb.String())
}

// wrapInBorder wraps content in a bordered container.
func (s *SidebarModel) wrapInBorder(content string) string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(common.ColorDockerBlue)).
		Width(s.width-2).
		Padding(0, 1)

	return borderStyle.Render(content)
}
