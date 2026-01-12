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
var tabIcons = []string{"[*]", "[>]", "[#]", "[@]"}

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

	var s strings.Builder

	// Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n")

	// Tabs
	s.WriteString(m.renderTabs())
	s.WriteString("\n\n")

	// Content
	s.WriteString(m.renderContent())

	// Help overlay
	if m.showHelp {
		s.WriteString("\n\n")
		s.WriteString(m.renderHelp())
	}

	// Status bar (at the bottom)
	s.WriteString("\n")
	s.WriteString(m.renderStatusBar())

	return s.String()
}

// renderHeader renders the header section
func (m Model) renderHeader() string {
	icon := IconStyle.Render("[+]")
	title := TitleStyle.Copy().MarginBottom(0).Render("PODX")
	subtitle := MutedStyle.Render("Secure Encryption Tool")
	version := MutedStyle.Render("v1.0")

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		icon, " ", title, " ", MutedStyle.Render("|"), " ", subtitle, " ", version)

	return HeaderStyle.Render(header)
}

// renderTabs renders the tab bar
func (m Model) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		icon := tabIcons[i]
		tabLabel := fmt.Sprintf(" %s %s ", icon, name)
		if i == m.activeTab {
			tabs = append(tabs, TabActiveStyle.Render(tabLabel))
		} else {
			tabs = append(tabs, TabInactiveStyle.Render(tabLabel))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
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
		"Keyboard Shortcuts:",
		"",
		"  Navigation:",
		"    up/k      - Move up",
		"    down/j    - Move down",
		"    left/h    - Move left",
		"    right/l   - Move right",
		"    enter     - Select",
		"",
		"  Tabs:",
		"    tab       - Next tab",
		"    shift+tab - Previous tab",
		"    1/2/3/4   - Jump to tab",
		"",
		"  Files Tab:",
		"    e         - Encrypt file",
		"    d         - Decrypt file",
		"    /         - Filter files",
		"",
		"  Other:",
		"    r         - Refresh",
		"    ?         - Toggle help",
		"    q/esc     - Quit",
	}

	return BoxStyle.Render(strings.Join(helpText, "\n"))
}

// renderStatusBar renders the status bar at the bottom
func (m Model) renderStatusBar() string {
	status := fmt.Sprintf(" %s | Press ? for help | q to quit ", m.statusMsg)
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
