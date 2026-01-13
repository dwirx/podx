package tui

import "github.com/charmbracelet/lipgloss"

// Colors - PODX color palette (Modern dark theme)
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#00D4FF") // Bright cyan
	ColorSecondary = lipgloss.Color("#BD93F9") // Purple
	ColorAccent    = lipgloss.Color("#FF79C6") // Pink
	ColorSuccess   = lipgloss.Color("#50FA7B") // Bright green
	ColorWarning   = lipgloss.Color("#F1FA8C") // Yellow
	ColorError     = lipgloss.Color("#FF5555") // Red
	ColorMuted     = lipgloss.Color("#6272A4") // Muted blue-gray
	ColorBg        = lipgloss.Color("#1E1E2E") // Dark background (Catppuccin Mocha)
	ColorBgLight   = lipgloss.Color("#313244") // Lighter background
	ColorBgDark    = lipgloss.Color("#11111B") // Darker background
	ColorBorder    = lipgloss.Color("#45475A") // Border color

	// Legacy aliases for compatibility
	ColorCyan   = ColorPrimary
	ColorGreen  = ColorSuccess
	ColorYellow = ColorWarning
	ColorRed    = ColorError
	ColorGray   = ColorMuted
	ColorWhite  = lipgloss.Color("#CDD6F4") // Catppuccin text
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
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(ColorPrimary)

	// BoxStyle for bordered containers
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBg).
			Padding(1, 2)

	// TabActiveStyle for active tab
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBg).
			Background(ColorPrimary).
			Padding(0, 2)

	// TabInactiveStyle for inactive tabs
	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2)

	// StatusBarStyle for the status bar at the bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorBgDark).
			Padding(0, 1)

	// HeaderStyle for the main header
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorBgDark).
			Padding(0, 2)

	// CardStyle for dashboard cards
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBg).
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
				Foreground(lipgloss.Color("#1E1E2E")).
				Background(ColorSuccess).
				Padding(0, 1).
				Bold(true)

	// BadgeWarningStyle for warning badges
	BadgeWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1E1E2E")).
				Background(ColorWarning).
				Padding(0, 1).
				Bold(true)

	// BadgeErrorStyle for error badges
	BadgeErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff")).
			Background(ColorError).
			Padding(0, 1).
			Bold(true)

	// SectionStyle for content sections
	SectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBg).
			Padding(1, 2).
			MarginBottom(1)

	// OverlayBgStyle for dialog overlays - dark semi-transparent background
	OverlayBgStyle = lipgloss.NewStyle().
			Background(ColorBgDark)

	// TabBarStyle for the tab bar container
	TabBarStyle = lipgloss.NewStyle().
			Background(ColorBgLight).
			Padding(0, 1).
			MarginBottom(1)

	// LogoStyle for the PODX logo
	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	// DividerStyle for horizontal dividers
	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorBorder)
)
