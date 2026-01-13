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
		},
		selected: 0,
		cwd:      cwd,
		keys:     DefaultKeyMap(),
		version:  "1.0.0", // Default version, will be set by TUI
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
		}
		return nil
	}
}

// Update handles messages for the dashboard model
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
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

	// Calculate responsive card widths
	cardWidth := (m.width - 6) / 2
	if cardWidth < 35 {
		cardWidth = 35
	}
	if cardWidth > 50 {
		cardWidth = 50
	}

	// Update notification at top if available
	if m.updateInfo != nil && m.updateInfo.Available {
		updateCard := m.renderUpdateNotification(cardWidth * 2)
		sections = append(sections, updateCard)
	}

	// Create horizontal layout with project info and security status side by side
	projectCard := m.renderProjectInfoWithWidth(cardWidth)
	securityCard := m.renderSecurityStatusWithWidth(cardWidth)

	// Join cards horizontally if there's enough width
	if m.width > 80 {
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, projectCard, securityCard)
		sections = append(sections, topRow)
	} else {
		// Stack vertically on narrow terminals
		sections = append(sections, projectCard)
		sections = append(sections, securityCard)
	}

	// Quick actions below
	actionsCard := m.renderQuickActionsWithWidth(cardWidth * 2)
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

	// Path - truncate if too long
	path := m.project.RootDir
	maxPathLen := width - 12
	if maxPathLen > 0 && len(path) > maxPathLen {
		path = "..." + path[len(path)-maxPathLen+3:]
	}
	lines = append(lines, fmt.Sprintf("  Path:       %s", MutedStyle.Render(path)))

	// Backend
	lines = append(lines, fmt.Sprintf("  Backend:    %s", SuccessStyle.Render(m.project.Config.Backend)))

	// Recipients
	recipientCount := len(m.project.Config.Recipients)
	lines = append(lines, fmt.Sprintf("  Recipients: %s", SuccessStyle.Render(fmt.Sprintf("%d", recipientCount))))

	for _, r := range m.project.Config.Recipients {
		keyPreview := r.Key
		if len(keyPreview) > 20 {
			keyPreview = keyPreview[:20] + "..."
		}
		lines = append(lines, MutedStyle.Render(fmt.Sprintf("    - %s (%s)", r.Name, keyPreview)))
	}

	// Secrets patterns
	secretCount := len(m.project.Config.Secrets)
	lines = append(lines, fmt.Sprintf("  Secrets:    %s patterns", SuccessStyle.Render(fmt.Sprintf("%d", secretCount))))
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

	if m.checkResult == nil {
		lines = append(lines, MutedStyle.Render("  Checking..."))
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
		lines = append(lines, SuccessStyle.Render("  [OK] Encryption"))
	} else {
		lines = append(lines, ErrorStyle.Render("  [!!] Encryption: Issues found"))
		for _, issue := range m.checkResult.EncryptionIssues {
			lines = append(lines, WarningStyle.Render(fmt.Sprintf("       - %s", issue)))
		}
	}

	// Gitignore Issues
	if len(m.checkResult.GitignoreIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] Gitignore"))
	} else {
		lines = append(lines, WarningStyle.Render("  [!!] Gitignore: Missing entries"))
		for _, pattern := range m.checkResult.GitignoreIssues {
			lines = append(lines, WarningStyle.Render(fmt.Sprintf("       - %s", pattern)))
		}
	}

	// Secret Findings
	if len(m.checkResult.SecretFindings) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] No secrets detected"))
	} else {
		lines = append(lines, ErrorStyle.Render(fmt.Sprintf("  [!!] %d files with secrets", len(m.checkResult.SecretFindings))))
	}

	// Hook status
	hookInstalled := security.IsHookInstalled(m.cwd)
	if hookInstalled {
		lines = append(lines, SuccessStyle.Render("  [OK] Pre-commit hook"))
	} else {
		lines = append(lines, WarningStyle.Render("  [--] No pre-commit hook"))
	}

	// Local key status
	if m.keyInfo.HasKey {
		keyPreview := m.keyInfo.PublicKey
		if len(keyPreview) > 20 {
			keyPreview = keyPreview[:20] + "..."
		}
		lines = append(lines, SuccessStyle.Render("  [OK] Local Age key"))
		lines = append(lines, MutedStyle.Render(fmt.Sprintf("       %s", keyPreview)))
	} else {
		lines = append(lines, WarningStyle.Render("  [--] No local Age key"))
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
	lines = append(lines, MutedStyle.Render("  Use j/k to navigate, Enter to execute"))
	lines = append(lines, "")

	actionIcons := []string{"E", "D", "C", "H", "K"}
	actionDescs := []string{
		"Encrypt all secret files",
		"Decrypt all secret files",
		"Run security checks",
		"Install pre-commit hook",
		"Generate new Age key pair",
	}

	for i, action := range m.actions {
		icon := actionIcons[i]
		desc := actionDescs[i]
		if i == m.selected {
			// Selected item with highlight
			selectedLine := fmt.Sprintf("  > [%s] %s", icon, action)
			lines = append(lines, SelectedStyle.Render(selectedLine))
			lines = append(lines, MutedStyle.Render("       "+desc))
		} else {
			lines = append(lines, fmt.Sprintf("    [%s] %s", MutedStyle.Render(icon), action))
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
