package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hades/podx/keygen"
	"github.com/hades/podx/project"
	"github.com/hades/podx/security"
	"github.com/hades/podx/updater"
)

// Action identifiers
const (
	ActionEncryptAll = iota
	ActionDecryptAll
	ActionCheck
	ActionHookInstall
	ActionKeygen
	ActionManageKeys
	ActionInitProject
)

// DashboardModel represents the dashboard tab content
type DashboardModel struct {
	project     *project.Project
	checkResult *security.CheckResult
	updateInfo  *updater.UpdateInfo
	loading     bool
	err         error
	selected    int
	actions     []string
	width       int
	height      int
	cwd         string
	keys        KeyMap
	version     string
	keyInfo     keygen.KeyInfo
	keyManager  KeyManagerModel
}

// projectLoadedMsg is sent when project loading completes
type projectLoadedMsg struct {
	project *project.Project
	err     error
	keyInfo keygen.KeyInfo
}

// checkCompletedMsg is sent when security check completes
type checkCompletedMsg struct {
	result *security.CheckResult
}

// updateCheckMsg is sent when update check completes
type updateCheckMsg struct {
	info *updater.UpdateInfo
}

// actionResultMsg is sent when an action completes
type actionResultMsg struct {
	action  int
	success bool
	message string
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() DashboardModel {
	cwd, _ := os.Getwd()
	return DashboardModel{
		loading: true,
		actions: []string{
			"Encrypt All",
			"Decrypt All",
			"Run Check",
			"Install Hook",
			"Generate Key",
			"Manage Keys",
			"Init Project",
		},
		selected:   0,
		cwd:        cwd,
		keys:       DefaultKeyMap(),
		version:    "1.0.0", // Default version, will be set by TUI
		keyManager: NewKeyManagerModel(),
	}
}

// SetVersion sets the current version for update checking
func (m *DashboardModel) SetVersion(version string) {
	m.version = version
}

// Init initializes the dashboard model
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(m.loadProject, m.checkForUpdates)
}

// checkForUpdates checks for available updates in background
func (m DashboardModel) checkForUpdates() tea.Msg {
	info := updater.CheckUpdate(m.version)
	return updateCheckMsg{info: info}
}

// loadProject loads the project configuration
func (m DashboardModel) loadProject() tea.Msg {
	proj, err := project.Load(m.cwd)
	keyInfo := keygen.GetAgeKeyInfo()
	return projectLoadedMsg{project: proj, err: err, keyInfo: keyInfo}
}

// runSecurityCheck performs a security check on the project
func (m DashboardModel) runSecurityCheck() tea.Msg {
	result := security.CheckProject(m.cwd, false)
	return checkCompletedMsg{result: &result}
}

// executeAction runs the selected action
func (m DashboardModel) executeAction(action int) tea.Cmd {
	return func() tea.Msg {
		switch action {
		case ActionEncryptAll:
			if m.project == nil {
				return actionResultMsg{action: action, success: false, message: "No project loaded"}
			}
			count, err := m.project.EncryptAll()
			if err != nil {
				return actionResultMsg{action: action, success: false, message: err.Error()}
			}
			return actionResultMsg{action: action, success: true, message: fmt.Sprintf("Encrypted %d files", count)}

		case ActionDecryptAll:
			if m.project == nil {
				return actionResultMsg{action: action, success: false, message: "No project loaded"}
			}
			count, err := m.project.DecryptAll()
			if err != nil {
				return actionResultMsg{action: action, success: false, message: err.Error()}
			}
			return actionResultMsg{action: action, success: true, message: fmt.Sprintf("Decrypted %d files", count)}

		case ActionCheck:
			result := security.CheckProject(m.cwd, false)
			if result.Passed {
				return actionResultMsg{action: action, success: true, message: "All checks passed"}
			}
			return actionResultMsg{action: action, success: false, message: "Some checks failed"}

		case ActionHookInstall:
			err := security.InstallHook(m.cwd)
			if err != nil {
				return actionResultMsg{action: action, success: false, message: err.Error()}
			}
			return actionResultMsg{action: action, success: true, message: "Hook installed successfully"}

		case ActionKeygen:
			result, err := keygen.GenerateAge()
			if err != nil {
				return actionResultMsg{action: action, success: false, message: err.Error()}
			}
			// Show truncated public key
			pubKey := result.PublicKey
			if len(pubKey) > 30 {
				pubKey = pubKey[:30] + "..."
			}
			return actionResultMsg{action: action, success: true, message: fmt.Sprintf("Generated key: %s", pubKey)}

		case ActionInitProject:
			proj, err := project.Init(m.cwd)
			if err != nil {
				return actionResultMsg{action: action, success: false, message: err.Error()}
			}
			return actionResultMsg{action: action, success: true, message: fmt.Sprintf("Project initialized with %d recipient(s)", len(proj.Config.Recipients))}
		}
		return nil
	}
}

// Update handles messages for the dashboard model
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	// Handle key manager dialog first
	if m.keyManager.IsVisible() {
		var cmd tea.Cmd
		m.keyManager, cmd = m.keyManager.Update(msg)
		// If key manager was closed, reload project/key info
		if !m.keyManager.IsVisible() {
			m.keyInfo = keygen.GetAgeKeyInfo()
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case projectLoadedMsg:
		m.loading = false
		m.project = msg.project
		m.err = msg.err
		m.keyInfo = msg.keyInfo
		if m.project != nil {
			// Project loaded, run security check
			return m, m.runSecurityCheck
		}
		return m, nil

	case checkCompletedMsg:
		m.checkResult = msg.result
		return m, nil

	case updateCheckMsg:
		m.updateInfo = msg.info
		return m, nil

	case actionResultMsg:
		// Could update status message here if needed
		// For now, just trigger a reload
		m.loading = true
		return m, m.loadProject

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			if m.selected < len(m.actions)-1 {
				m.selected++
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter), key.Matches(msg, m.keys.Right):
			// Handle special actions that don't require async execution
			if m.selected == ActionManageKeys {
				// Open key manager dialog
				return m, m.keyManager.Show(m.width, m.height)
			}
			if m.selected == ActionInitProject {
				// Init project action - works even without existing project
				return m, m.executeAction(m.selected)
			}
			if m.project != nil {
				return m, m.executeAction(m.selected)
			}
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, m.loadProject
		}
	}

	return m, nil
}

// View renders the dashboard model
func (m DashboardModel) View() string {
	// Show key manager dialog if visible
	if m.keyManager.IsVisible() {
		return CenterDialog(m.keyManager.View(), m.width, m.height)
	}

	if m.loading {
		return BoxStyle.Render(RenderSpinner(0) + " Loading...")
	}

	if m.err != nil {
		return m.renderNoProject()
	}

	return m.renderDashboard()
}

// renderNoProject renders the view when no project is found
func (m DashboardModel) renderNoProject() string {
	content := []string{
		"",
		ErrorStyle.Render("[!!] No project found"),
		"",
		MutedStyle.Render("This directory is not a PODX project."),
		MutedStyle.Render("Run 'podx init' to initialize a new project."),
		"",
		MutedStyle.Render(fmt.Sprintf("Path: %s", m.cwd)),
		"",
	}
	return BoxStyle.Render(strings.Join(content, "\n"))
}

// renderDashboard renders the full dashboard view
func (m DashboardModel) renderDashboard() string {
	var sections []string

	// Get terminal size category
	termSize := GetTerminalSize(m.width)

	// Calculate responsive card widths based on terminal size
	var cardWidth int
	switch termSize {
	case TerminalSmall:
		// Single column layout for small terminals
		cardWidth = m.width - 6
		if cardWidth < 30 {
			cardWidth = 30
		}
	case TerminalMedium:
		// Two column layout
		cardWidth = (m.width - 8) / 2
		if cardWidth < 35 {
			cardWidth = 35
		}
	default:
		// Large terminal - comfortable card widths
		cardWidth = (m.width - 10) / 2
		if cardWidth > 55 {
			cardWidth = 55
		}
	}

	// Update notification at top if available
	if m.updateInfo != nil && m.updateInfo.Available {
		updateWidth := m.width - 6
		if updateWidth > 100 {
			updateWidth = 100
		}
		updateCard := m.renderUpdateNotification(updateWidth)
		sections = append(sections, updateCard)
	}

	// Create horizontal layout with project info and security status
	projectCard := m.renderProjectInfoWithWidth(cardWidth)
	securityCard := m.renderSecurityStatusWithWidth(cardWidth)

	// Layout based on terminal size
	if termSize == TerminalSmall {
		// Stack vertically on small terminals
		sections = append(sections, projectCard)
		sections = append(sections, "")
		sections = append(sections, securityCard)
	} else {
		// Side by side on medium/large terminals
		// Use lipgloss.JoinHorizontal with Top alignment
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, projectCard, "   ", securityCard)
		sections = append(sections, topRow)
	}

	// Add spacing before quick actions
	sections = append(sections, "")

	// Quick actions - use same width as combined cards for consistency
	actionsWidth := cardWidth*2 + 3 // Two cards + spacing between them
	if actionsWidth > m.width-6 {
		actionsWidth = m.width - 6
	}
	if actionsWidth < 50 {
		actionsWidth = 50
	}
	actionsCard := m.renderQuickActionsWithWidth(actionsWidth)
	sections = append(sections, actionsCard)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderUpdateNotification renders the update notification banner
func (m DashboardModel) renderUpdateNotification(width int) string {
	var lines []string

	lines = append(lines, WarningStyle.Render("[!] Update Available"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  New version: %s -> %s", m.updateInfo.CurrentVersion, m.updateInfo.LatestVersion))
	if m.updateInfo.DownloadSize > 0 {
		lines = append(lines, fmt.Sprintf("  Size: %s", updater.FormatSize(m.updateInfo.DownloadSize)))
	}
	if m.updateInfo.IsBeta {
		lines = append(lines, WarningStyle.Render("  This is a beta release"))
	}
	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("  Run 'podx update' to upgrade"))

	style := CardStyle.Copy().Width(width)
	return style.Render(strings.Join(lines, "\n"))
}

// renderProjectInfo renders project information
func (m DashboardModel) renderProjectInfo() string {
	return m.renderProjectInfoWithWidth(0)
}

// renderProjectInfoWithWidth renders project information with specified width
func (m DashboardModel) renderProjectInfoWithWidth(width int) string {
	var lines []string

	lines = append(lines, CardTitleStyle.Render("Project Info"))
	lines = append(lines, "")

	// Path - truncate based on available width
	path := m.project.RootDir
	maxPathLen := width - 12 // Account for "Path: " and padding
	if maxPathLen < 20 {
		maxPathLen = 20
	}
	path = TruncateText(path, maxPathLen)
	lines = append(lines, fmt.Sprintf("  Path:       %s", lipgloss.NewStyle().Foreground(ColorSecondary).Render(path)))

	// Backend with icon
	backendIcon := "🔐"
	if m.project.Config.Backend == "gpg" {
		backendIcon = "🔑"
	}
	lines = append(lines, fmt.Sprintf("  Backend:    %s %s", backendIcon, SuccessStyle.Render(m.project.Config.Backend)))

	// Recipients with count badge
	recipientCount := len(m.project.Config.Recipients)
	countBadge := lipgloss.NewStyle().Foreground(ColorBg).Background(ColorSecondary).Padding(0, 1).Bold(true).Render(fmt.Sprintf("%d", recipientCount))
	lines = append(lines, fmt.Sprintf("  Recipients: %s", countBadge))

	for _, r := range m.project.Config.Recipients {
		keyPreview := r.Key
		if len(keyPreview) > 20 {
			keyPreview = keyPreview[:20] + "..."
		}
		lines = append(lines, MutedStyle.Render(fmt.Sprintf("    - %s (%s)", r.Name, keyPreview)))
	}

	// Secrets patterns with count
	secretCount := len(m.project.Config.Secrets)
	secretBadge := lipgloss.NewStyle().Foreground(ColorBg).Background(ColorWarning).Padding(0, 1).Bold(true).Render(fmt.Sprintf("%d", secretCount))
	lines = append(lines, fmt.Sprintf("  Secrets:    %s patterns", secretBadge))
	for _, s := range m.project.Config.Secrets {
		lines = append(lines, MutedStyle.Render(fmt.Sprintf("    - %s", s)))
	}

	style := CardStyle.Copy()
	if width > 0 {
		style = style.Width(width)
	}
	return style.Render(strings.Join(lines, "\n"))
}

// renderSecurityStatus renders the security check status
func (m DashboardModel) renderSecurityStatus() string {
	return m.renderSecurityStatusWithWidth(0)
}

// renderSecurityStatusWithWidth renders the security check status with specified width
func (m DashboardModel) renderSecurityStatusWithWidth(width int) string {
	var lines []string

	lines = append(lines, CardTitleStyle.Render("Security Status"))
	lines = append(lines, "")

	if m.checkResult == nil {
		lines = append(lines, MutedStyle.Render("  "+RenderSpinner(0)+" Checking..."))
		style := CardStyle.Copy()
		if width > 0 {
			style = style.Width(width)
		}
		return style.Render(strings.Join(lines, "\n"))
	}

	// Overall status with badge
	if m.checkResult.Passed {
		lines = append(lines, "  Status: "+BadgeSuccessStyle.Render(" PASS "))
	} else {
		lines = append(lines, "  Status: "+BadgeErrorStyle.Render(" FAIL "))
	}

	lines = append(lines, "")

	// Encryption Issues
	if len(m.checkResult.EncryptionIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] ")+"Encryption")
	} else {
		lines = append(lines, ErrorStyle.Render("  [!!] ")+"Encryption: "+ErrorStyle.Render("Issues found"))
		for _, issue := range m.checkResult.EncryptionIssues {
			lines = append(lines, WarningStyle.Render(fmt.Sprintf("       - %s", issue)))
		}
	}

	// Gitignore Issues
	if len(m.checkResult.GitignoreIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] ")+"Gitignore")
	} else {
		lines = append(lines, WarningStyle.Render("  [!!] ")+"Gitignore: "+WarningStyle.Render("Missing entries"))
		for _, pattern := range m.checkResult.GitignoreIssues {
			lines = append(lines, WarningStyle.Render(fmt.Sprintf("       - %s", pattern)))
		}
	}

	// Secret Findings
	if len(m.checkResult.SecretFindings) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] ")+"No secrets detected")
	} else {
		lines = append(lines, ErrorStyle.Render("  [!!] ")+ErrorStyle.Render(fmt.Sprintf("%d files with secrets", len(m.checkResult.SecretFindings))))
	}

	// Hook status
	hookInstalled := security.IsHookInstalled(m.cwd)
	if hookInstalled {
		lines = append(lines, SuccessStyle.Render("  [OK] ")+"Pre-commit hook")
	} else {
		lines = append(lines, MutedStyle.Render("  [--] ")+"No pre-commit hook")
	}

	// Local key status
	if m.keyInfo.HasKey {
		keyPreview := m.keyInfo.PublicKey
		if len(keyPreview) > 20 {
			keyPreview = keyPreview[:20] + "..."
		}
		lines = append(lines, SuccessStyle.Render("  [OK] ")+"Local Age key")
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorSecondary).Render(fmt.Sprintf("       %s", keyPreview)))
	} else {
		lines = append(lines, MutedStyle.Render("  [--] ")+"No local Age key")
		lines = append(lines, MutedStyle.Render("       Press [K] to generate"))
	}

	style := CardStyle.Copy()
	if width > 0 {
		style = style.Width(width)
	}
	return style.Render(strings.Join(lines, "\n"))
}

// renderQuickActions renders the quick actions menu
func (m DashboardModel) renderQuickActions() string {
	return m.renderQuickActionsWithWidth(0)
}

// renderQuickActionsWithWidth renders the quick actions menu with specified width
func (m DashboardModel) renderQuickActionsWithWidth(width int) string {
	var lines []string

	lines = append(lines, CardTitleStyle.Render("Quick Actions"))
	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("  Use j/k to navigate, Enter to execute"))
	lines = append(lines, "")

	actionIcons := []string{"E", "D", "C", "H", "K", "M", "I"}
	actionEmojis := []string{"🔐", "🔓", "🛡️", "🪝", "🔑", "📋", "🚀"}
	actionDescs := []string{
		"Encrypt all secret files",
		"Decrypt all secret files",
		"Run security checks",
		"Install pre-commit hook",
		"Generate new Age key pair",
		"Manage keys (import, select, delete)",
		"Initialize new PODX project",
	}

	for i, action := range m.actions {
		icon := actionIcons[i]
		emoji := actionEmojis[i]
		desc := actionDescs[i]
		if i == m.selected {
			// Selected item with highlight and emoji
			selectedLine := fmt.Sprintf("  > [%s] %s %s", icon, emoji, action)
			lines = append(lines, SelectedStyle.Render(selectedLine))
			lines = append(lines, MutedStyle.Render("       "+desc))
		} else {
			// Unselected item with dimmed icon
			keyStyle := lipgloss.NewStyle().Foreground(ColorMuted).Bold(true)
			lines = append(lines, fmt.Sprintf("    [%s] %s %s", keyStyle.Render(icon), emoji, action))
		}
	}

	style := CardStyle.Copy()
	if width > 0 {
		style = style.Width(width)
	}
	return style.Render(strings.Join(lines, "\n"))
}

// SetSize updates the model dimensions
func (m *DashboardModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
