package common

import (
	_ "embed"

	"charm.land/lipgloss/v2"
)

//go:embed logo.txt
var Logo string

// Docker-branded color palette
const (
	// Primary brand colors
	ColorDockerBlue      = "#2496ED"
	ColorDockerDarkBlue  = "#1A5276"
	ColorDockerLightBlue = "#5DADE2"

	// Semantic colors
	ColorSuccess = "#27AE60"
	ColorWarning = "#F39C12"
	ColorError   = "#E74C3C"

	// UI chrome colors
	ColorBackground    = "#0D1117"
	ColorBackgroundAlt = "#161B22"
	ColorSurface       = "#21262D"
	ColorBorder        = "#30363D"
	ColorBorderFocused = ColorDockerBlue

	// Text colors
	ColorTextPrimary   = "#E6EDF3"
	ColorTextSecondary = "#8B949E"
	ColorTextMuted     = "#484F58"
)

// Pre-built styles for common UI elements
var (
	// Primary accent style (Docker blue)
	StylePrimary = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDockerBlue)).
			Bold(true)

	// Secondary accent (darker blue)
	StyleSecondary = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDockerDarkBlue))

	// Light accent
	StyleAccent = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDockerLightBlue))

	// Semantic styles
	StyleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess))
	StyleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning))
	StyleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError))

	// Text styles
	StyleTextPrimary = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextPrimary))
	StyleTextSecondary = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextSecondary))
	StyleTextMuted = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTextMuted))

	// Background styles
	StyleSurface = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorSurface)).
			Foreground(lipgloss.Color(ColorTextPrimary))
	StyleBackground = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBackground)).
			Foreground(lipgloss.Color(ColorTextPrimary))

	// Border styles
	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorBorder))
	StyleBorderFocused = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(ColorBorderFocused))

	// Status bar style
	StyleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBackgroundAlt)).
			Foreground(lipgloss.Color(ColorTextSecondary)).
			Padding(0, 1)

	// Gutter style (left border for messages)
	StyleGutter = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDockerBlue))

	// Input prompt style
	StyleInputPrompt = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDockerBlue)).
				Bold(true)
)
