package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialogResult is sent when confirm dialog is answered
type ConfirmDialogResult struct {
	Confirmed bool
	ActionID  string
}

// ConfirmDialogModel represents a confirmation dialog
type ConfirmDialogModel struct {
	Title    string
	Message  string
	ActionID string

	confirmed bool
	selected  int // 0 = Cancel, 1 = Confirm
	visible   bool
	width     int
	height    int
	keys      KeyMap
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(title, message, actionID string) ConfirmDialogModel {
	return ConfirmDialogModel{
		Title:    title,
		Message:  message,
		ActionID: actionID,
		selected: 0, // Default to Cancel for safety
		visible:  true,
		keys:     DefaultKeyMap(),
	}
}

// Init initializes the confirm dialog
func (m ConfirmDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles confirm dialog messages
func (m ConfirmDialogModel) Update(msg tea.Msg) (ConfirmDialogModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// Cancel with escape
		case msg.String() == "esc", msg.String() == "n":
			m.visible = false
			return m, func() tea.Msg {
				return ConfirmDialogResult{Confirmed: false, ActionID: m.ActionID}
			}

		// Confirm with y
		case msg.String() == "y":
			m.confirmed = true
			m.visible = false
			return m, func() tea.Msg {
				return ConfirmDialogResult{Confirmed: true, ActionID: m.ActionID}
			}

		// Enter to confirm selection
		case msg.String() == "enter":
			m.confirmed = m.selected == 1
			m.visible = false
			return m, func() tea.Msg {
				return ConfirmDialogResult{Confirmed: m.confirmed, ActionID: m.ActionID}
			}

		// Left/Right or h/l to switch selection
		case key.Matches(msg, m.keys.Left), msg.String() == "h":
			m.selected = 0
			return m, nil

		case key.Matches(msg, m.keys.Right), msg.String() == "l":
			m.selected = 1
			return m, nil

		// Tab to switch
		case msg.String() == "tab":
			m.selected = (m.selected + 1) % 2
			return m, nil
		}
	}

	return m, nil
}

// View renders the confirm dialog
func (m ConfirmDialogModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title with warning icon using PODX colors
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorWarning).
		MarginBottom(1)
	content.WriteString(titleStyle.Render("⚠️  " + m.Title))
	content.WriteString("\n\n")

	// Message
	msgStyle := lipgloss.NewStyle().
		Foreground(ColorWhite).
		Width(50)
	content.WriteString(msgStyle.Render(m.Message))
	content.WriteString("\n\n")

	// Buttons
	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(ColorWhite)

	confirmStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(ColorError)

	cancelBtn := "[ Cancel ]"
	confirmBtn := "[ Confirm ]"

	if m.selected == 0 {
		cancelBtn = lipgloss.NewStyle().
			Background(ColorBgLight).
			Foreground(ColorWhite).
			Padding(0, 1).
			Render(" Cancel ")
		confirmBtn = confirmStyle.Render(confirmBtn)
	} else {
		cancelBtn = cancelStyle.Render(cancelBtn)
		confirmBtn = lipgloss.NewStyle().
			Background(ColorError).
			Foreground(ColorWhite).
			Padding(0, 1).
			Render(" Confirm ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "    ", confirmBtn)
	buttonContainer := lipgloss.NewStyle().
		Width(56).
		Align(lipgloss.Center).
		Render(buttons)

	content.WriteString(buttonContainer)
	content.WriteString("\n\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Align(lipgloss.Center)
	help := "y: confirm | n/Esc: cancel | Tab/h/l: switch | Enter: select"
	content.WriteString(helpStyle.Render(help))

	// Dialog box style with warning border and solid background
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Background(ColorBg).
		Padding(1, 2).
		Width(60)

	return dialogStyle.Render(content.String())
}

// IsVisible returns whether dialog is visible
func (m ConfirmDialogModel) IsVisible() bool {
	return m.visible
}

// Show makes the dialog visible
func (m *ConfirmDialogModel) Show() {
	m.visible = true
	m.confirmed = false
	m.selected = 0 // Default to Cancel
}

// Hide hides the dialog
func (m *ConfirmDialogModel) Hide() {
	m.visible = false
}

// SetSize updates dialog dimensions
func (m *ConfirmDialogModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// DangerousActions returns action IDs that need confirmation
func DangerousActions() map[string]bool {
	return map[string]bool{
		"decrypt-all":    true,
		"encrypt-all":    true,
		"hook-uninstall": true,
		"update":         true,
		"update-beta":    true,
		"rollback":       true,
	}
}

// GetConfirmMessage returns confirmation message for dangerous actions
func GetConfirmMessage(actionID string) (title, message string) {
	switch actionID {
	case "decrypt-all":
		return "Decrypt All Secrets", "This will decrypt all encrypted files in this project.\n\nAre you sure you want to continue?"

	case "encrypt-all":
		return "Encrypt All Secrets", "This will encrypt all secret files and DELETE the originals.\n\nMake sure you have the password/key to decrypt them later.\n\nAre you sure?"

	case "hook-uninstall":
		return "Uninstall Pre-commit Hook", "This will remove the security check hook.\n\nYou may accidentally commit secrets without this protection.\n\nContinue?"

	case "update":
		return "Update PODX", "This will download and install the latest version.\n\nContinue with update?"

	case "update-beta":
		return "Update to Beta", "This will install the BETA version which may have bugs.\n\nContinue with beta update?"

	case "rollback":
		return "Rollback Update", "This will restore the previous version of PODX.\n\nContinue with rollback?"

	default:
		return "Confirm Action", "Are you sure you want to perform this action?"
	}
}
