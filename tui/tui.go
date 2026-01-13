package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab constants
const (
	TabDashboard = iota
	TabCommands
	TabSecurity
	TabFiles
)

// Tab names with icons
var tabNames = []string{"Dashboard", "Commands", "Security", "Files"}
var tabIcons = []string{" ", " ", " ", " "}

// Model is the main TUI model
type Model struct {
	activeTab int
	width     int
	height    int
	showHelp  bool
	statusMsg string
	keys      KeyMap

	// Sub-models
	dashboard DashboardModel
	commands  CommandsModel
	security  SecurityModel
	files     FilesModel
}

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		activeTab: TabDashboard,
		showHelp:  false,
		statusMsg: "Ready",
		keys:      DefaultKeyMap(),
		dashboard: NewDashboardModel(),
		commands:  NewCommandsModel(),
		security:  NewSecurityModel(),
		files:     NewFilesModel(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Initialize sub-models and collect their init commands
	return tea.Batch(
		m.dashboard.Init(),
		m.security.Init(),
		m.files.Init(),
	)
}

// Update handles all messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if any dialog is active - if so, pass keys directly to sub-model
		// This allows dialogs to receive number keys, tab, etc. without global interception
		if m.hasActiveDialog() {
			var cmd tea.Cmd
			switch m.activeTab {
			case TabDashboard:
				m.dashboard, cmd = m.dashboard.Update(msg)
			case TabCommands:
				m.commands, cmd = m.commands.Update(msg)
			case TabSecurity:
				m.security, cmd = m.security.Update(msg)
			case TabFiles:
				m.files, cmd = m.files.Update(msg)
			}
			return m, cmd
		}

		// Global key handling (only when no dialog is active)
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.Tab):
			m.activeTab = (m.activeTab + 1) % len(tabNames)
			m.statusMsg = fmt.Sprintf("Switched to %s", tabNames[m.activeTab])
			return m, nil

		case key.Matches(msg, m.keys.ShiftTab):
			m.activeTab = (m.activeTab - 1 + len(tabNames)) % len(tabNames)
			m.statusMsg = fmt.Sprintf("Switched to %s", tabNames[m.activeTab])
			return m, nil

		case key.Matches(msg, m.keys.Tab1):
			m.activeTab = TabDashboard
			m.statusMsg = "Switched to Dashboard"
			return m, nil

		case key.Matches(msg, m.keys.Tab2):
			m.activeTab = TabCommands
			m.statusMsg = "Switched to Commands"
			return m, nil

		case key.Matches(msg, m.keys.Tab3):
			m.activeTab = TabSecurity
			m.statusMsg = "Switched to Security"
			return m, nil

		case key.Matches(msg, m.keys.Tab4):
			m.activeTab = TabFiles
			m.statusMsg = "Switched to Files"
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			m.statusMsg = "Refreshed"
			return m, nil
		}

		// Pass key messages to active sub-model only
		var cmd tea.Cmd
		switch m.activeTab {
		case TabDashboard:
			m.dashboard, cmd = m.dashboard.Update(msg)
		case TabCommands:
			m.commands, cmd = m.commands.Update(msg)
		case TabSecurity:
			m.security, cmd = m.security.Update(msg)
		case TabFiles:
			m.files, cmd = m.files.Update(msg)
		}
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update sub-model sizes (accounting for header, tabs, and status bar)
		contentHeight := m.height - 6
		m.dashboard.SetSize(m.width, contentHeight)
		m.commands.SetSize(m.width, contentHeight)
		m.security.SetSize(m.width, contentHeight)
		m.files.SetSize(m.width, contentHeight)

		return m, nil

	default:
		// Pass all other messages (async results) to all sub-models
		// since we don't know which sub-model the message is for
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.dashboard, cmd = m.dashboard.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.commands, cmd = m.commands.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.security, cmd = m.security.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.files, cmd = m.files.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}
}

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Help overlay takes precedence
	if m.showHelp {
		return m.renderHelp()
	}

	// If a fullscreen dialog is active, render only the dialog with full terminal dimensions
	if m.hasFullscreenDialog() {
		return m.renderFullscreenDialog()
	}

	var s strings.Builder

	// Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n")

	// Tabs
	s.WriteString(m.renderTabs())
	s.WriteString("\n")

	// Content
	s.WriteString(m.renderContent())

	// Status bar (at the bottom)
	s.WriteString("\n")
	s.WriteString(m.renderStatusBar())

	return s.String()
}

// renderFullscreenDialog renders fullscreen dialogs with proper terminal dimensions
func (m Model) renderFullscreenDialog() string {
	var dialog string

	// Check Commands tab dialogs
	if m.activeTab == TabCommands {
		if m.commands.formDialog != nil && m.commands.formDialog.IsVisible() {
			dialog = m.commands.formDialog.View()
		} else if m.commands.confirmDialog != nil && m.commands.confirmDialog.IsVisible() {
			dialog = m.commands.confirmDialog.View()
		} else if m.commands.progressView != nil && m.commands.progressView.IsVisible() {
			dialog = m.commands.progressView.View()
		}
	}

	// Check Files tab encrypt dialog
	if m.activeTab == TabFiles {
		if m.files.encryptDialog.IsVisible() {
			dialog = m.files.encryptDialog.View()
		}
	}

	if dialog == "" {
		return m.renderContent()
	}

	// Use full terminal dimensions for centering
	return CenterDialog(dialog, m.width, m.height)
}

// renderHeader renders the header section
func (m Model) renderHeader() string {
	// ASCII art logo - simple and clean
	logo := LogoStyle.Render("PODX")
	subtitle := MutedStyle.Render(" Secure Encryption Tool")
	version := MutedStyle.Render(" v1.0")

	header := lipgloss.JoinHorizontal(lipgloss.Center, logo, subtitle, version)

	// Full width header bar
	headerBar := lipgloss.NewStyle().
		Background(ColorBgDark).
		Padding(0, 2).
		Width(m.width).
		Render(header)

	return headerBar
}

// renderTabs renders the tab bar
func (m Model) renderTabs() string {
	var tabs []string

	for i, name := range tabNames {
		// Tab number indicator
		tabNum := fmt.Sprintf("%d", i+1)

		if i == m.activeTab {
			// Active tab - highlighted
			tabLabel := fmt.Sprintf(" %s %s ", tabNum, name)
			tabs = append(tabs, TabActiveStyle.Render(tabLabel))
		} else {
			// Inactive tab
			tabLabel := fmt.Sprintf(" %s %s ", tabNum, name)
			tabs = append(tabs, TabInactiveStyle.Render(tabLabel))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Add separator line below tabs
	separator := DividerStyle.Render(strings.Repeat("─", m.width))

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, separator)
}

// renderContent renders the active tab content
func (m Model) renderContent() string {
	switch m.activeTab {
	case TabDashboard:
		return m.dashboard.View()
	case TabCommands:
		return m.commands.View()
	case TabSecurity:
		return m.security.View()
	case TabFiles:
		return m.files.View()
	default:
		return ""
	}
}

// renderHelp renders the help overlay
func (m Model) renderHelp() string {
	helpText := []string{
		TitleStyle.Render("Keyboard Shortcuts"),
		"",
		CardTitleStyle.Render("Navigation:"),
		"  up/k        Move up",
		"  down/j      Move down",
		"  left/h      Move left / Back",
		"  right/l     Move right / Select",
		"  Enter       Confirm action",
		"",
		CardTitleStyle.Render("Tabs:"),
		"  Tab         Next tab",
		"  Shift+Tab   Previous tab",
		"  1/2/3/4     Jump to tab directly",
		"",
		CardTitleStyle.Render("Files Tab:"),
		"  e           Encrypt selected file(s)",
		"  d           Decrypt selected file(s)",
		"  Space       Toggle file selection",
		"  a           Select/deselect all",
		"  /           Filter files",
		"  g           Go to path",
		"  p           Toggle preview panel",
		"",
		CardTitleStyle.Render("General:"),
		"  r           Refresh current view",
		"  ?           Toggle this help",
		"  q/Esc       Quit",
		"",
		MutedStyle.Render("Press any key to close..."),
	}

	// Center the help dialog
	helpContent := strings.Join(helpText, "\n")
	helpBox := BoxStyle.Copy().
		BorderForeground(ColorPrimary).
		Width(50).
		Render(helpContent)

	return CenterDialog(helpBox, m.width, m.height)
}

// renderStatusBar renders the status bar at the bottom
func (m Model) renderStatusBar() string {
	// Left side - status message
	statusLeft := fmt.Sprintf(" %s", m.statusMsg)

	// Right side - help hint
	statusRight := "? Help  q Quit "

	// Calculate padding
	padding := m.width - lipgloss.Width(statusLeft) - lipgloss.Width(statusRight)
	if padding < 0 {
		padding = 0
	}

	status := statusLeft + strings.Repeat(" ", padding) + statusRight
	return StatusBarStyle.Width(m.width).Render(status)
}

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// hasActiveDialog checks if any dialog is currently active
// This is used to prevent global key bindings from intercepting keys meant for dialogs
func (m Model) hasActiveDialog() bool {
	// Check Commands tab dialogs
	if m.activeTab == TabCommands {
		if m.commands.formDialog != nil && m.commands.formDialog.IsVisible() {
			return true
		}
		if m.commands.confirmDialog != nil && m.commands.confirmDialog.IsVisible() {
			return true
		}
		if m.commands.progressView != nil && m.commands.progressView.IsVisible() {
			return true
		}
	}

	// Check Files tab dialogs
	if m.activeTab == TabFiles {
		if m.files.encryptDialog.IsVisible() {
			return true
		}
		if m.files.filtering || m.files.showGoto {
			return true
		}
	}

	return false
}

// hasFullscreenDialog checks if any fullscreen dialog overlay is currently active
// Fullscreen dialogs should hide the header, tabs, and status bar
func (m Model) hasFullscreenDialog() bool {
	// Check Commands tab dialogs
	if m.activeTab == TabCommands {
		if m.commands.formDialog != nil && m.commands.formDialog.IsVisible() {
			return true
		}
		if m.commands.confirmDialog != nil && m.commands.confirmDialog.IsVisible() {
			return true
		}
		if m.commands.progressView != nil && m.commands.progressView.IsVisible() {
			return true
		}
	}

	// Check Files tab encrypt dialog
	if m.activeTab == TabFiles {
		if m.files.encryptDialog.IsVisible() {
			return true
		}
	}

	return false
}
