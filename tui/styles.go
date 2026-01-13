package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors - Dracula color palette
// https://draculatheme.com/contribute
var (
	// Dracula primary colors
	ColorPrimary   = lipgloss.Color("#BD93F9") // Purple
	ColorSecondary = lipgloss.Color("#8BE9FD") // Cyan
	ColorAccent    = lipgloss.Color("#FF79C6") // Pink
	ColorSuccess   = lipgloss.Color("#50FA7B") // Green
	ColorWarning   = lipgloss.Color("#F1FA8C") // Yellow
	ColorError     = lipgloss.Color("#FF5555") // Red
	ColorMuted     = lipgloss.Color("#6272A4") // Comment (muted purple-gray)
	ColorBg        = lipgloss.Color("#282A36") // Background
	ColorBgLight   = lipgloss.Color("#44475A") // Current Line / Selection
	ColorBgDark    = lipgloss.Color("#21222C") // Darker background
	ColorBorder    = lipgloss.Color("#6272A4") // Border (comment color)
	ColorFg        = lipgloss.Color("#F8F8F2") // Foreground
	ColorOrange    = lipgloss.Color("#FFB86C") // Orange

	// Legacy aliases for compatibility
	ColorCyan   = ColorSecondary
	ColorGreen  = ColorSuccess
	ColorYellow = ColorWarning
	ColorRed    = ColorError
	ColorGray   = ColorMuted
	ColorWhite  = ColorFg
)

// Styles - Reusable lipgloss styles with Dracula theme
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
			Foreground(ColorBg).
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
			Foreground(ColorFg).
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
				Foreground(ColorBg).
				Background(ColorSuccess).
				Padding(0, 1).
				Bold(true)

	// BadgeWarningStyle for warning badges
	BadgeWarningStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorWarning).
				Padding(0, 1).
				Bold(true)

	// BadgeErrorStyle for error badges
	BadgeErrorStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
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

	// OverlayBgStyle for dialog overlays
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

	// BreadcrumbStyle for navigation breadcrumbs
	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// BreadcrumbActiveStyle for active breadcrumb
	BreadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true)

	// ToastStyle for notification toasts
	ToastStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Background(ColorBg).
			Padding(0, 2)

	// ToastSuccessStyle for success toasts
	ToastSuccessStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Background(ColorBg).
				Padding(0, 2)

	// ToastErrorStyle for error toasts
	ToastErrorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Background(ColorBg).
			Padding(0, 2)

	// LogEntryStyle for activity log entries
	LogEntryStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	// LogTimestampStyle for log timestamps
	LogTimestampStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// StatValueStyle for statistic values
	StatValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	// StatLabelStyle for statistic labels
	StatLabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// SearchStyle for global search input
	SearchStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Background(ColorBg).
			Padding(0, 1)

	// HighlightStyle for highlighted text
	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorOrange).
			Bold(true)

	// LinkStyle for links and interactive elements
	LinkStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Underline(true)
)

// SmallLogo for header
const SmallLogo = "🔐 PODX"

// TerminalBorder is a custom border style for terminal-like appearance
var TerminalBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "┌",
	TopRight:    "┐",
	BottomLeft:  "└",
	BottomRight: "┘",
}

// RenderHorizontalDivider renders a horizontal divider
func RenderHorizontalDivider(width int) string {
	return DividerStyle.Render(strings.Repeat("─", width))
}

// RenderDoubleDivider renders a double-line divider
func RenderDoubleDivider(width int) string {
	return DividerStyle.Render(strings.Repeat("═", width))
}

// CenterDialog centers a dialog within the terminal dimensions
func CenterDialog(dialog string, termWidth, termHeight int) string {
	// Ensure minimum dimensions
	if termWidth < 80 {
		termWidth = 80
	}
	if termHeight < 24 {
		termHeight = 24
	}

	// Use lipgloss.Place to center the dialog with a background
	overlay := lipgloss.Place(
		termWidth,
		termHeight,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceBackground(ColorBgDark),
		lipgloss.WithWhitespaceForeground(ColorBgDark),
	)

	return overlay
}

// RenderSpinner returns an animated spinner frame
func RenderSpinner(frame int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return lipgloss.NewStyle().Foreground(ColorPrimary).Render(spinners[frame%len(spinners)])
}

// RenderStatusIndicator returns a status indicator
func RenderStatusIndicator(ok bool) string {
	if ok {
		return SuccessStyle.Render("✓")
	}
	return ErrorStyle.Render("✗")
}

// RenderMenuItem renders a menu item
func RenderMenuItem(label string, shortcut string, selected bool) string {
	if selected {
		return SelectedStyle.Render(" > [" + shortcut + "] " + label + " ")
	}
	return "   [" + MutedStyle.Render(shortcut) + "] " + label
}

// RenderProgressBar renders a progress bar
func RenderProgressBar(current, total, width int) string {
	if total == 0 || width < 5 {
		return "░░░░░"
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percent := int(float64(current) / float64(total) * 100)

	return lipgloss.NewStyle().Foreground(ColorPrimary).Render(bar) + " " +
		StatValueStyle.Render(strings.Repeat(" ", 3-len(string(rune(percent/10)))+1)) +
		StatValueStyle.Render(string(rune('0'+percent/100%10))) +
		StatValueStyle.Render(string(rune('0'+percent/10%10))) +
		StatValueStyle.Render(string(rune('0'+percent%10))) +
		StatValueStyle.Render("%")
}

// RenderKeyHint renders a keyboard shortcut hint
func RenderKeyHint(key, description string) string {
	return lipgloss.NewStyle().
		Foreground(ColorBg).
		Background(ColorMuted).
		Padding(0, 1).
		Render(key) + " " +
		MutedStyle.Render(description)
}

// RenderFileIcon returns an icon for file types
func RenderFileIcon(isDir, isEncrypted bool, ext string) string {
	if isDir {
		return IconStyle.Render("📁")
	}
	if isEncrypted {
		return WarningStyle.Render("🔒")
	}

	switch strings.ToLower(ext) {
	case ".go":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Render("⚙")
	case ".py":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3776AB")).Render("🐍")
	case ".js", ".ts":
		return lipgloss.NewStyle().Foreground(ColorWarning).Render("⚡")
	case ".json", ".yaml", ".yml":
		return SuccessStyle.Render("📋")
	case ".md", ".txt":
		return MutedStyle.Render("📝")
	case ".env":
		return WarningStyle.Render("🔐")
	default:
		return MutedStyle.Render("📄")
	}
}

// RenderTag renders a colored tag/badge
func RenderTag(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(ColorBg).
		Background(color).
		Padding(0, 1).
		Bold(true).
		Render(text)
}

// RenderBox renders content in a styled box
func RenderBox(title, content string, width int) string {
	if width < 20 {
		width = 40
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Background(ColorBg).
		Padding(1, 2).
		Width(width)

	if title != "" {
		titleStyle := CardTitleStyle.Copy().MarginBottom(1)
		content = titleStyle.Render(title) + "\n" + content
	}

	return style.Render(content)
}
