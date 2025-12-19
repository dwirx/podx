package tui

import "github.com/charmbracelet/lipgloss"

// Color palette - Modern dark theme
var (
	// Primary colors
	Primary      = lipgloss.Color("#A78BFA") // Violet
	PrimaryDark  = lipgloss.Color("#7C3AED") // Deep violet
	Secondary    = lipgloss.Color("#22D3EE") // Cyan
	Accent       = lipgloss.Color("#FBBF24") // Amber

	// Status colors
	Success = lipgloss.Color("#34D399") // Emerald
	Error   = lipgloss.Color("#F87171") // Red
	Warning = lipgloss.Color("#FBBF24") // Amber
	Info    = lipgloss.Color("#60A5FA") // Blue

	// Neutral colors
	Text       = lipgloss.Color("#F1F5F9") // Slate 100
	TextMuted  = lipgloss.Color("#94A3B8") // Slate 400
	TextDim    = lipgloss.Color("#64748B") // Slate 500
	Background = lipgloss.Color("#0F172A") // Slate 900
	Surface    = lipgloss.Color("#1E293B") // Slate 800
	Border     = lipgloss.Color("#334155") // Slate 700
	BorderFocus = lipgloss.Color("#A78BFA") // Violet
)

// Logo style
var LogoStyle = lipgloss.NewStyle().
	Foreground(Primary).
	Bold(true)

// Title styles
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Italic(true)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(PrimaryDark).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)
)

// Box styles
var (
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderFocus).
			Padding(1, 2)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Background(Surface).
			Padding(1, 2)
)

// Input styles
var (
	InputStyle = lipgloss.NewStyle().
			Foreground(Text).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(Text).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(Primary).
				Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	FocusedLabelStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true)

	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(TextDim).
				Italic(true)
)

// Button styles
var (
	ButtonStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Surface).
			Padding(0, 3).
			MarginRight(1)

	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(Primary).
				Padding(0, 3).
				MarginRight(1).
				Bold(true)

	SuccessButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(Success).
				Padding(0, 3).
				Bold(true)

	DangerButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(Error).
				Padding(0, 3).
				Bold(true)
)

// List item styles
var (
	ItemStyle = lipgloss.NewStyle().
			Foreground(Text).
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true).
				PaddingLeft(0)

	ItemDescStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			PaddingLeft(4)

	FileItemStyle = lipgloss.NewStyle().
			Foreground(Text)

	DirItemStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)
)

// Status message styles
var (
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)
)

// Help styles
var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	HelpSeparator = lipgloss.NewStyle().
			Foreground(Border).
			Render(" • ")
)

// Progress bar styles
var (
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(Primary)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(Border)
)

// Badge styles
var (
	EncryptedBadge = lipgloss.NewStyle().
			Foreground(Success).
			Background(lipgloss.Color("#064E3B")).
			Padding(0, 1).
			Bold(true).
			Render("🔐 ENCRYPTED")

	DecryptedBadge = lipgloss.NewStyle().
			Foreground(Warning).
			Background(lipgloss.Color("#78350F")).
			Padding(0, 1).
			Bold(true).
			Render("🔓 DECRYPTED")
)

// Divider
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(Border).
		Render(lipgloss.NewStyle().Width(width).Render("─────────────────────────────────────────────────────────────"))
}

// RenderKey renders a keyboard shortcut
func RenderKey(key string) string {
	return HelpKeyStyle.Render(key)
}

// RenderHelp renders help text with key
func RenderHelp(key, desc string) string {
	return RenderKey(key) + " " + HelpDescStyle.Render(desc)
}

// Spinner characters for animation
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
