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
	TabLogs
)

// Tab names with ASCII indicators
var tabNames = []string{"Dashboard", "Commands", "Security", "Files", "Logs"}
var tabShortcuts = []string{"1", "2", "3", "4", "5"}

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
	logs      ActivityLogModel

	// Global components
	search       GlobalSearchModel
	toastManager *ToastManager
	searchKeys   GlobalSearchKeyMap
}

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		activeTab:    TabDashboard,
		showHelp:     false,
		statusMsg:    "Ready",
		keys:         DefaultKeyMap(),
		dashboard:    NewDashboardModel(),
		commands:     NewCommandsModel(),
		security:     NewSecurityModel(),
		files:        NewFilesModel(),
		logs:         NewActivityLogModel(),
		search:       NewGlobalSearchModel(),
		toastManager: NewToastManager(),
		searchKeys:   DefaultGlobalSearchKeyMap(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Log startup
	GetActivityLog().Info("PODX TUI started")

	// Initialize sub-models and collect their init commands
	return tea.Batch(
		m.dashboard.Init(),
		m.security.Init(),
		m.files.Init(),
		ToastTick(),
	)
}

// Update handles all messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global search first if visible
	if m.search.IsVisible() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			return m, cmd
		case SearchNavigateMsg:
			// Navigate to the file in Files tab
			m.activeTab = TabFiles
			m.files.cwd = msg.Path
			m.files.loading = true
			GetActivityLog().Info("Navigated to: " + msg.Path)
			return m, m.files.loadFiles
		case SearchExecuteMsg:
			// Switch to Commands tab and execute
			m.activeTab = TabCommands
			GetActivityLog().Info("Executing: " + msg.CommandID)
			return m, nil
		}
	}

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
			case TabLogs:
				m.logs, cmd = m.logs.Update(msg)
			}
			return m, cmd
		}

		// Global key handling (only when no dialog is active)
		switch {
		case key.Matches(msg, m.keys.Quit):
			GetActivityLog().Info("PODX TUI exited")
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.searchKeys.Toggle):
			// Toggle global search
			m.search.SetSize(m.width, m.height)
			return m, m.search.Show()

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

		case msg.String() == "5":
			m.activeTab = TabLogs
			m.statusMsg = "Switched to Logs"
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			m.statusMsg = "Refreshed"
			GetActivityLog().Info("Refreshed")
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
		case TabLogs:
			m.logs, cmd = m.logs.Update(msg)
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
		m.logs.SetSize(m.width, contentHeight)
		m.search.SetSize(m.width, m.height)
		m.toastManager.SetWidth(m.width)

		return m, nil

	case ToastTickMsg:
		// Cleanup expired toasts and continue ticking
		m.toastManager.Cleanup()
		return m, ToastTick()

	case ShowToastMsg:
		m.toastManager.Add(msg.Type, msg.Title, msg.Message)
		return m, nil

	case LogActivityMsg:
		// Activity already logged, just update status
		m.statusMsg = msg.Message
		return m, nil

	case actionResultMsg:
		// Log action results
		if msg.success {
			GetActivityLog().Success(msg.message)
			m.toastManager.Success("Success", msg.message)
		} else {
			GetActivityLog().Error(msg.message)
			m.toastManager.Error("Error", msg.message)
		}
		// Pass to dashboard
		var cmd tea.Cmd
		m.dashboard, cmd = m.dashboard.Update(msg)
		return m, cmd

	case fileOperationMsg:
		// Log file operations
		if msg.success {
			GetActivityLog().Success(msg.message)
			m.toastManager.Success("File Operation", msg.message)
		} else {
			GetActivityLog().Error(msg.message)
			m.toastManager.Error("Error", msg.message)
		}
		// Pass to files tab
		var cmd tea.Cmd
		m.files, cmd = m.files.Update(msg)
		return m, cmd

	case commandOutputMsg:
		// Log command execution
		if msg.err != nil {
			GetActivityLog().Error("Command failed: " + msg.command)
		} else {
			GetActivityLog().Success("Command completed: " + msg.command)
		}
		// Pass to commands tab
		var cmd tea.Cmd
		m.commands, cmd = m.commands.Update(msg)
		return m, cmd

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

	// Show warning for very small terminals
	if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
		return m.renderSizeWarning()
	}

	// Help overlay takes precedence
	if m.showHelp {
		return m.renderHelp()
	}

	// Global search overlay
	if m.search.IsVisible() {
		return m.search.View()
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

	// Overlay toasts if any
	mainView := s.String()
	if m.toastManager.HasToasts() {
		toastView := m.toastManager.View()
		// Position toasts at top-right
		toastStyle := lipgloss.NewStyle().
			MarginTop(2).
			MarginRight(2)
		mainView = lipgloss.JoinHorizontal(lipgloss.Top,
			mainView,
			toastStyle.Render(toastView))
	}

	return mainView
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

// renderHeader renders the header section with ASCII art style
func (m Model) renderHeader() string {
	// ASCII-style header
	logo := LogoStyle.Render(SmallLogo)
	subtitle := MutedStyle.Render(" Secure Encryption Tool")
	version := MutedStyle.Render(" v1.0")

	header := lipgloss.JoinHorizontal(lipgloss.Center, logo, subtitle, version)

	// Full width header bar with ASCII border feel
	headerBar := lipgloss.NewStyle().
		Background(ColorBgDark).
		Foreground(ColorPrimary).
		Padding(0, 2).
		Width(m.width).
		Render(header)

	return headerBar
}

// renderTabs renders the tab bar with ASCII style
func (m Model) renderTabs() string {
	var tabs []string

	// Short names for small terminals
	shortNames := []string{"Dash", "Cmd", "Sec", "Files", "Log"}
	useShortNames := m.width < SmallWidth

	for i, name := range tabNames {
		// Tab number indicator
		tabNum := tabShortcuts[i]

		// Use short name on small terminals
		displayName := name
		if useShortNames {
			displayName = shortNames[i]
		}

		if i == m.activeTab {
			// Active tab - highlighted with ASCII brackets
			tabLabel := fmt.Sprintf("[%s:%s]", tabNum, displayName)
			tabs = append(tabs, TabActiveStyle.Render(tabLabel))
		} else {
			// Inactive tab
			tabLabel := fmt.Sprintf(" %s:%s ", tabNum, displayName)
			tabs = append(tabs, TabInactiveStyle.Render(tabLabel))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Add ASCII separator line below tabs
	separator := DividerStyle.Render(strings.Repeat("═", m.width))

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
	case TabLogs:
		return m.renderLogsTab()
	default:
		return ""
	}
}

// renderLogsTab renders the activity logs tab
func (m Model) renderLogsTab() string {
	var content strings.Builder

	// Title
	content.WriteString(TitleStyle.Render("[ ACTIVITY LOG ]"))
	content.WriteString("\n\n")

	// Log entries
	content.WriteString(m.logs.RenderCompact(m.height - 10))

	content.WriteString("\n\n")
	content.WriteString(RenderHorizontalDivider(50))
	content.WriteString("\n\n")

	// Help
	content.WriteString(MutedStyle.Render("  [j/k] Navigate  [g] Go to top  [G] Go to bottom"))

	return BoxStyle.Render(content.String())
}

// renderSizeWarning renders a warning for terminals that are too small
func (m Model) renderSizeWarning() string {
	var s strings.Builder

	s.WriteString(WarningStyle.Render("⚠ Terminal Too Small"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("Current: %dx%d\n", m.width, m.height))
	s.WriteString(fmt.Sprintf("Minimum: %dx%d\n", MinTerminalWidth, MinTerminalHeight))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("Please resize your terminal"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("or press 'q' to quit"))

	return BoxStyle.Render(s.String())
}

// renderHelp renders the help overlay with ASCII style
func (m Model) renderHelp() string {
	helpText := []string{
		TitleStyle.Render("[ KEYBOARD SHORTCUTS ]"),
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
		"  1/2/3/4/5   Jump to tab directly",
		"",
		CardTitleStyle.Render("Search:"),
		"  Ctrl+F      Global search",
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
	helpBox := lipgloss.NewStyle().
		Border(TerminalBorder).
		BorderForeground(ColorPrimary).
		Background(ColorBg).
		Padding(1, 2).
		Width(50).
		Render(helpContent)

	return CenterDialog(helpBox, m.width, m.height)
}

// renderStatusBar renders the status bar at the bottom with ASCII style
func (m Model) renderStatusBar() string {
	// Left side - status message
	statusLeft := fmt.Sprintf(" %s", m.statusMsg)

	// Right side - help hint
	statusRight := "? Help | Ctrl+F Search | q Quit "

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
