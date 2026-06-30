package common

import (
	_ "embed"

	"charm.land/lipgloss/v2"
)

//go:embed logo.txt
var Logo string

var (
	FWhiteBlue = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B9BCC"))
	FGreenish  = lipgloss.NewStyle().Foreground(lipgloss.Color("#4EC7C4"))
	BGreenish  = lipgloss.NewStyle().Background(lipgloss.Color("#4EC7C4"))

	BEmerald = lipgloss.NewStyle().Background(lipgloss.Color("#50C878"))

	FMobyBlue = lipgloss.NewStyle().Foreground(lipgloss.Color("#1D63ED"))

	BMobyBlue = lipgloss.NewStyle().Background(lipgloss.Color("#1D63ED"))

	FCyan = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FD7AF"))

	BBlack = lipgloss.NewStyle().Background(lipgloss.Color("#131313"))

	BRed = lipgloss.NewStyle().Background(lipgloss.Color("#e84c3d"))
	FRed = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4545"))

	FGray = lipgloss.NewStyle().Foreground(lipgloss.Color("#384d54"))
)
