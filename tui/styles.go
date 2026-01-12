package tui

import "github.com/charmbracelet/lipgloss"

// Colors - PODX color palette
var (
	ColorCyan   = lipgloss.Color("#00FFFF")
	ColorGreen  = lipgloss.Color("#00FF00")
	ColorYellow = lipgloss.Color("#FFFF00")
	ColorRed    = lipgloss.Color("#FF0000")
	ColorGray   = lipgloss.Color("#808080")
	ColorWhite  = lipgloss.Color("#FFFFFF")
)

// Styles - Reusable lipgloss styles
var (
	// TitleStyle for headers and titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			MarginBottom(1)

	// SuccessStyle for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	// WarningStyle for warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	// MutedStyle for secondary/muted text
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	// SelectedStyle for selected items in lists
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Background(lipgloss.Color("#333333"))

	// BoxStyle for bordered containers
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2)

	// TabActiveStyle for active tab
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Background(lipgloss.Color("#333333")).
			Padding(0, 2)

	// TabInactiveStyle for inactive tabs
	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorGray).
				Padding(0, 2)

	// StatusBarStyle for the status bar at the bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)
)
