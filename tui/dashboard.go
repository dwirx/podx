package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hades/podx/project"
	"github.com/hades/podx/security"
)

// Action identifiers
const (
	ActionEncryptAll = iota
	ActionDecryptAll
	ActionCheck
	ActionHookInstall
)

// DashboardModel represents the dashboard tab content
type DashboardModel struct {
	project     *project.Project
	checkResult *security.CheckResult
	loading     bool
	err         error
	selected    int
	actions     []string
	width       int
	height      int
	cwd         string
	keys        KeyMap
}

// projectLoadedMsg is sent when project loading completes
type projectLoadedMsg struct {
	project *project.Project
	err     error
}

// checkCompletedMsg is sent when security check completes
type checkCompletedMsg struct {
	result *security.CheckResult
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
		},
		selected: 0,
		cwd:      cwd,
		keys:     DefaultKeyMap(),
	}
}

// Init initializes the dashboard model
func (m DashboardModel) Init() tea.Cmd {
	return m.loadProject
}

// loadProject loads the project configuration
func (m DashboardModel) loadProject() tea.Msg {
	proj, err := project.Load(m.cwd)
	return projectLoadedMsg{project: proj, err: err}
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
		if m.project != nil {
			// Project loaded, run security check
			return m, m.runSecurityCheck
		}
		return m, nil

	case checkCompletedMsg:
		m.checkResult = msg.result
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
		return BoxStyle.Render("Loading...")
	}

	if m.err != nil {
		return m.renderNoProject()
	}

	return m.renderDashboard()
}

// renderNoProject renders the view when no project is found
func (m DashboardModel) renderNoProject() string {
	content := []string{
		ErrorStyle.Render("No project found"),
		"",
		MutedStyle.Render("This directory is not a PODX project."),
		MutedStyle.Render("Run 'podx init' to initialize a new project."),
		"",
		MutedStyle.Render(fmt.Sprintf("Path: %s", m.cwd)),
	}
	return BoxStyle.Render(strings.Join(content, "\n"))
}

// renderDashboard renders the full dashboard view
func (m DashboardModel) renderDashboard() string {
	var sections []string

	// Project Info Section
	sections = append(sections, m.renderProjectInfo())

	// Security Status Section
	sections = append(sections, m.renderSecurityStatus())

	// Quick Actions Section
	sections = append(sections, m.renderQuickActions())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderProjectInfo renders project information
func (m DashboardModel) renderProjectInfo() string {
	var lines []string

	lines = append(lines, TitleStyle.Render("Project Info"))
	lines = append(lines, "")

	// Path
	lines = append(lines, fmt.Sprintf("  Path:    %s", m.project.RootDir))

	// Backend
	lines = append(lines, fmt.Sprintf("  Backend: %s", m.project.Config.Backend))

	// Recipients
	recipientCount := len(m.project.Config.Recipients)
	lines = append(lines, fmt.Sprintf("  Recipients: %d", recipientCount))

	for _, r := range m.project.Config.Recipients {
		keyPreview := r.Key
		if len(keyPreview) > 20 {
			keyPreview = keyPreview[:20] + "..."
		}
		lines = append(lines, MutedStyle.Render(fmt.Sprintf("    - %s (%s)", r.Name, keyPreview)))
	}

	// Secrets patterns
	secretCount := len(m.project.Config.Secrets)
	lines = append(lines, fmt.Sprintf("  Secrets: %d patterns", secretCount))
	for _, s := range m.project.Config.Secrets {
		lines = append(lines, MutedStyle.Render(fmt.Sprintf("    - %s", s)))
	}

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// renderSecurityStatus renders the security check status
func (m DashboardModel) renderSecurityStatus() string {
	var lines []string

	lines = append(lines, TitleStyle.Render("Security Status"))
	lines = append(lines, "")

	if m.checkResult == nil {
		lines = append(lines, MutedStyle.Render("  Checking..."))
		return BoxStyle.Render(strings.Join(lines, "\n"))
	}

	// Overall status
	if m.checkResult.Passed {
		lines = append(lines, SuccessStyle.Render("  [PASS] All checks passed"))
	} else {
		lines = append(lines, ErrorStyle.Render("  [FAIL] Some checks failed"))
	}

	lines = append(lines, "")

	// Encryption Issues
	if len(m.checkResult.EncryptionIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  Encryption: OK"))
	} else {
		lines = append(lines, ErrorStyle.Render("  Encryption: Issues found"))
		for _, issue := range m.checkResult.EncryptionIssues {
			lines = append(lines, WarningStyle.Render(fmt.Sprintf("    - %s", issue)))
		}
	}

	// Gitignore Issues
	if len(m.checkResult.GitignoreIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  Gitignore:  OK"))
	} else {
		lines = append(lines, WarningStyle.Render("  Gitignore:  Missing entries"))
		for _, pattern := range m.checkResult.GitignoreIssues {
			lines = append(lines, WarningStyle.Render(fmt.Sprintf("    - %s", pattern)))
		}
	}

	// Secret Findings
	if len(m.checkResult.SecretFindings) == 0 {
		lines = append(lines, SuccessStyle.Render("  Scan:       No secrets detected"))
	} else {
		lines = append(lines, ErrorStyle.Render(fmt.Sprintf("  Scan:       %d files with secrets", len(m.checkResult.SecretFindings))))
	}

	// Hook status
	hookInstalled := security.IsHookInstalled(m.cwd)
	if hookInstalled {
		lines = append(lines, SuccessStyle.Render("  Hook:       Installed"))
	} else {
		lines = append(lines, WarningStyle.Render("  Hook:       Not installed"))
	}

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// renderQuickActions renders the quick actions menu
func (m DashboardModel) renderQuickActions() string {
	var lines []string

	lines = append(lines, TitleStyle.Render("Quick Actions"))
	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("  Use j/k to navigate, Enter/l to execute"))
	lines = append(lines, "")

	for i, action := range m.actions {
		prefix := "  "
		if i == m.selected {
			// Selected item
			lines = append(lines, SelectedStyle.Render(fmt.Sprintf("> [%d] %s", i+1, action)))
		} else {
			lines = append(lines, fmt.Sprintf("%s [%d] %s", prefix, i+1, action))
		}
	}

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// SetSize updates the model dimensions
func (m *DashboardModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
