package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Colors - Terminal Classic palette (retro terminal aesthetic)
var (
	// Primary colors - Terminal green/cyan theme
	ColorPrimary   = lipgloss.Color("#00FF00") // Bright green (terminal classic)
	ColorSecondary = lipgloss.Color("#00FFFF") // Cyan
	ColorAccent    = lipgloss.Color("#FFFF00") // Yellow
	ColorSuccess   = lipgloss.Color("#00FF00") // Bright green
	ColorWarning   = lipgloss.Color("#FFFF00") // Yellow
	ColorError     = lipgloss.Color("#FF0000") // Red
	ColorMuted     = lipgloss.Color("#808080") // Gray
	ColorBg        = lipgloss.Color("#000000") // Pure black
	ColorBgLight   = lipgloss.Color("#1A1A1A") // Slightly lighter black
	ColorBgDark    = lipgloss.Color("#000000") // Pure black
	ColorBorder    = lipgloss.Color("#00FF00") // Green borders

	// Legacy aliases for compatibility
	ColorCyan   = ColorSecondary
	ColorGreen  = ColorSuccess
	ColorYellow = ColorWarning
	ColorRed    = ColorError
	ColorGray   = ColorMuted
	ColorWhite  = lipgloss.Color("#FFFFFF") // Pure white
)

// ASCII Box drawing characters
const (
	BoxTopLeft     = "+"
	BoxTopRight    = "+"
	BoxBottomLeft  = "+"
	BoxBottomRight = "+"
	BoxHorizontal  = "-"
	BoxVertical    = "|"
	BoxTLeft       = "+"
	BoxTRight      = "+"
	BoxTTop        = "+"
	BoxTBottom     = "+"
	BoxCross       = "+"
)

// TerminalBorder creates a classic ASCII border
var TerminalBorder = lipgloss.Border{
	Top:         BoxHorizontal,
	Bottom:      BoxHorizontal,
	Left:        BoxVertical,
	Right:       BoxVertical,
	TopLeft:     BoxTopLeft,
	TopRight:    BoxTopRight,
	BottomLeft:  BoxBottomLeft,
	BottomRight: BoxBottomRight,
}

// DoubleBorder for emphasis
var DoubleBorder = lipgloss.Border{
	Top:         "=",
	Bottom:      "=",
	Left:        "||",
	Right:       "||",
	TopLeft:     "#",
	TopRight:    "#",
	BottomLeft:  "#",
	BottomRight: "#",
}

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
			Foreground(ColorBg).
			Background(ColorPrimary)

	// BoxStyle for bordered containers
	BoxStyle = lipgloss.NewStyle().
			Border(TerminalBorder).
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
			Foreground(ColorPrimary).
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
			Border(TerminalBorder).
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
			Border(TerminalBorder).
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

	// BreadcrumbStyle for navigation breadcrumbs
	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// BreadcrumbActiveStyle for active breadcrumb
	BreadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	// ToastStyle for notification toasts
	ToastStyle = lipgloss.NewStyle().
			Border(TerminalBorder).
			BorderForeground(ColorPrimary).
			Background(ColorBg).
			Padding(0, 2)

	// ToastSuccessStyle for success toasts
	ToastSuccessStyle = lipgloss.NewStyle().
				Border(TerminalBorder).
				BorderForeground(ColorSuccess).
				Background(ColorBg).
				Padding(0, 2)

	// ToastErrorStyle for error toasts
	ToastErrorStyle = lipgloss.NewStyle().
			Border(TerminalBorder).
			BorderForeground(ColorError).
			Background(ColorBg).
			Padding(0, 2)

	// LogEntryStyle for activity log entries
	LogEntryStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	// LogTimestampStyle for log timestamps
	LogTimestampStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// StatValueStyle for statistic values
	StatValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	// StatLabelStyle for statistic labels
	StatLabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// SearchStyle for global search input
	SearchStyle = lipgloss.NewStyle().
			Border(TerminalBorder).
			BorderForeground(ColorSecondary).
			Background(ColorBg).
			Padding(0, 1)
)

// ASCII Art Logo
const ASCIILogo = `
 ____   ___  ______  __
|  _ \ / _ \|  _ \ \/ /
| |_) | | | | | | \  /
|  __/| |_| | |_| /  \
|_|    \___/|____/_/\_\
`

// SmallLogo for header
const SmallLogo = "[ PODX ]"

// RenderASCIIBox renders content in an ASCII box
func RenderASCIIBox(title, content string, width int) string {
	if width < 10 {
		width = 40
	}

	// Top border with title
	titleLen := len(title)
	paddingLeft := (width - titleLen - 4) / 2
	paddingRight := width - titleLen - 4 - paddingLeft

	var sb strings.Builder

	// Top line
	sb.WriteString(BoxTopLeft)
	if title != "" {
		sb.WriteString(strings.Repeat(BoxHorizontal, paddingLeft))
		sb.WriteString("[ ")
		sb.WriteString(title)
		sb.WriteString(" ]")
		sb.WriteString(strings.Repeat(BoxHorizontal, paddingRight))
	} else {
		sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	}
	sb.WriteString(BoxTopRight)
	sb.WriteString("\n")

	// Content lines
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		sb.WriteString(BoxVertical)
		sb.WriteString(" ")
		// Truncate or pad line to fit
		if len(line) > width-4 {
			line = line[:width-7] + "..."
		}
		sb.WriteString(line)
		padding := width - len(line) - 3
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(BoxVertical)
		sb.WriteString("\n")
	}

	// Bottom line
	sb.WriteString(BoxBottomLeft)
	sb.WriteString(strings.Repeat(BoxHorizontal, width-2))
	sb.WriteString(BoxBottomRight)

	return sb.String()
}

// RenderProgressBar renders an ASCII progress bar
func RenderProgressBar(current, total, width int) string {
	if total == 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
	percent := int(float64(current) / float64(total) * 100)

	return "[" + bar + "] " + lipgloss.NewStyle().Foreground(ColorPrimary).Render(strings.Repeat("", 0)+string(rune('0'+percent/100%10))+string(rune('0'+percent/10%10))+string(rune('0'+percent%10))+"%")
}

// RenderHorizontalDivider renders a horizontal divider
func RenderHorizontalDivider(width int) string {
	return DividerStyle.Render(strings.Repeat(BoxHorizontal, width))
}

// RenderDoubleDivider renders a double-line divider
func RenderDoubleDivider(width int) string {
	return DividerStyle.Render(strings.Repeat("=", width))
}

// CenterDialog centers a dialog within the terminal dimensions
func CenterDialog(dialog string, termWidth, termHeight int) string {
	dialogWidth := lipgloss.Width(dialog)
	dialogHeight := lipgloss.Height(dialog)

	// Calculate horizontal padding
	hPadding := (termWidth - dialogWidth) / 2
	if hPadding < 0 {
		hPadding = 0
	}

	// Calculate vertical padding
	vPadding := (termHeight - dialogHeight) / 2
	if vPadding < 0 {
		vPadding = 0
	}

	// Create the centered view
	centeredStyle := lipgloss.NewStyle().
		Width(termWidth).
		Height(termHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Background(ColorBg)

	return centeredStyle.Render(dialog)
}

// RenderSpinner returns an ASCII spinner frame
func RenderSpinner(frame int) string {
	spinners := []string{"|", "/", "-", "\\"}
	return spinners[frame%len(spinners)]
}

// RenderStatusIndicator returns a status indicator
func RenderStatusIndicator(ok bool) string {
	if ok {
		return SuccessStyle.Render("[OK]")
	}
	return ErrorStyle.Render("[!!]")
}

// RenderMenuItem renders a menu item
func RenderMenuItem(label string, shortcut string, selected bool) string {
	if selected {
		return SelectedStyle.Render(" > [" + shortcut + "] " + label + " ")
	}
	return "   [" + MutedStyle.Render(shortcut) + "] " + label
}
