package tui

import "github.com/charmbracelet/lipgloss"

// Colors - PODX color palette (Dracula-inspired)
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#00D4FF") // Bright cyan
	ColorSecondary = lipgloss.Color("#7B68EE") // Medium slate blue
	ColorAccent    = lipgloss.Color("#FF6B6B") // Coral red
	ColorSuccess   = lipgloss.Color("#50FA7B") // Bright green
	ColorWarning   = lipgloss.Color("#FFB86C") // Orange
	ColorError     = lipgloss.Color("#FF5555") // Red
	ColorMuted     = lipgloss.Color("#6272A4") // Muted blue-gray
	ColorBg        = lipgloss.Color("#282A36") // Dark background
	ColorBgLight   = lipgloss.Color("#44475A") // Lighter background

	// Legacy aliases for compatibility
	ColorCyan   = ColorPrimary
	ColorGreen  = ColorSuccess
	ColorYellow = ColorWarning
	ColorRed    = ColorError
	ColorGray   = ColorMuted
	ColorWhite  = lipgloss.Color("#F8F8F2")
)

// Styles - Reusable lipgloss styles
var (
	// TitleStyle for headers and titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	// SuccessStyle for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	// WarningStyle for warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	// MutedStyle for secondary/muted text
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// SelectedStyle for selected items in lists
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorBgLight)

	// BoxStyle for bordered containers
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	// TabActiveStyle for active tab
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorBgLight).
			Padding(0, 2).
			BorderStyle(lipgloss.Border{Bottom: "─"}).
			BorderBottom(true).
			BorderForeground(ColorPrimary)

	// TabInactiveStyle for inactive tabs
	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2)

	// StatusBarStyle for the status bar at the bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorBgLight).
			Padding(0, 1)

	// HeaderStyle for the main header
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorBgLight).
			Padding(0, 2).
			MarginBottom(1)

	// CardStyle for dashboard cards
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2).
			MarginRight(1)

	// CardTitleStyle for card headers
	CardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			MarginBottom(1)

	// IconStyle for decorative icons
	IconStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	// BadgeSuccessStyle for success badges
	BadgeSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000")).
				Background(ColorSuccess).
				Padding(0, 1)

	// BadgeWarningStyle for warning badges
	BadgeWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000")).
				Background(ColorWarning).
				Padding(0, 1)

	// BadgeErrorStyle for error badges
	BadgeErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Background(ColorError).
			Padding(0, 1)

	// SectionStyle for content sections
	SectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2).
			MarginBottom(1)

	// OverlayBgStyle for dialog overlays - dark semi-transparent background
	OverlayBgStyle = lipgloss.NewStyle().
			Background(ColorBg)
)
