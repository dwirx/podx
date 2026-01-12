package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hades/podx/security"
)

// Security action identifiers
const (
	SecurityActionRefresh = iota
	SecurityActionFix
	SecurityActionInstallHook
)

// securityCheckMsg is sent when security check completes
type securityCheckMsg struct {
	result *security.CheckResult
}

// securityActionMsg is sent when a security action completes
type securityActionMsg struct {
	action  int
	success bool
	message string
}

// SecurityModel represents the security tab content
type SecurityModel struct {
	result   *security.CheckResult
	loading  bool
	selected int
	actions  []string
	cwd      string
	keys     KeyMap
	width    int
	height   int
	message  string
}

// NewSecurityModel creates a new security model
func NewSecurityModel() SecurityModel {
	cwd, _ := os.Getwd()
	return SecurityModel{
		loading:  true,
		selected: 0,
		actions:  []string{"Refresh", "Fix Issues", "Install Hook"},
		cwd:      cwd,
		keys:     DefaultKeyMap(),
	}
}

// Init initializes the security model
func (m SecurityModel) Init() tea.Cmd {
	return m.runSecurityCheck
}

// runSecurityCheck performs a security check on the project
func (m SecurityModel) runSecurityCheck() tea.Msg {
	result := security.CheckProject(m.cwd, false)
	return securityCheckMsg{result: &result}
}

// executeAction runs the selected action
func (m SecurityModel) executeAction(action int) tea.Cmd {
	return func() tea.Msg {
		switch action {
		case SecurityActionRefresh:
			result := security.CheckProject(m.cwd, false)
			return securityCheckMsg{result: &result}

		case SecurityActionFix:
			// Run check with fix=true to auto-fix gitignore issues
			result := security.CheckProject(m.cwd, true)
			if result.Passed {
				return securityActionMsg{action: action, success: true, message: "All issues fixed"}
			}
			return securityActionMsg{action: action, success: false, message: "Some issues could not be fixed automatically"}

		case SecurityActionInstallHook:
			err := security.InstallHook(m.cwd)
			if err != nil {
				return securityActionMsg{action: action, success: false, message: err.Error()}
			}
			return securityActionMsg{action: action, success: true, message: "Hook installed successfully"}
		}
		return nil
	}
}

// Update handles messages for the security model
func (m SecurityModel) Update(msg tea.Msg) (SecurityModel, tea.Cmd) {
	switch msg := msg.(type) {
	case securityCheckMsg:
		m.loading = false
		m.result = msg.result
		m.message = ""
		return m, nil

	case securityActionMsg:
		m.loading = false
		if msg.success {
			m.message = SuccessStyle.Render(msg.message)
		} else {
			m.message = ErrorStyle.Render(msg.message)
		}
		// Refresh after action
		return m, m.runSecurityCheck

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch {
		// Vim navigation: h/l for left/right action selection
		case key.Matches(msg, m.keys.Left):
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case key.Matches(msg, m.keys.Right):
			if m.selected < len(m.actions)-1 {
				m.selected++
			}
			return m, nil

		// Shortcut keys
		case msg.String() == "r":
			m.loading = true
			m.message = ""
			return m, m.executeAction(SecurityActionRefresh)

		case msg.String() == "f":
			m.loading = true
			m.message = "Fixing issues..."
			return m, m.executeAction(SecurityActionFix)

		case msg.String() == "i":
			m.loading = true
			m.message = "Installing hook..."
			return m, m.executeAction(SecurityActionInstallHook)

		// Enter to execute selected action
		case key.Matches(msg, m.keys.Enter):
			m.loading = true
			m.message = ""
			return m, m.executeAction(m.selected)
		}
	}

	return m, nil
}

// View renders the security model
func (m SecurityModel) View() string {
	if m.loading {
		return BoxStyle.Render("Running security check...")
	}

	var sections []string

	// Title
	sections = append(sections, m.renderTitle())

	// Check results
	sections = append(sections, m.renderResults())

	// Detailed findings (if any issues)
	if m.result != nil && !m.result.Passed {
		sections = append(sections, m.renderFindings())
	}

	// Overall result
	sections = append(sections, m.renderOverallResult())

	// Status message (if any)
	if m.message != "" {
		sections = append(sections, m.message)
	}

	// Action bar
	sections = append(sections, m.renderActionBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTitle renders the security check title
func (m SecurityModel) renderTitle() string {
	return BoxStyle.Render(TitleStyle.Render("Security Check"))
}

// renderResults renders the security check results summary
func (m SecurityModel) renderResults() string {
	var lines []string

	if m.result == nil {
		lines = append(lines, MutedStyle.Render("No results available"))
		return BoxStyle.Render(strings.Join(lines, "\n"))
	}

	// Encryption Status
	if len(m.result.EncryptionIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  Encryption Status    All secrets encrypted"))
	} else {
		lines = append(lines, ErrorStyle.Render("  Encryption Status    Issues found"))
	}

	// Gitignore
	if len(m.result.GitignoreIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  Gitignore            Properly configured"))
	} else {
		lines = append(lines, WarningStyle.Render("  Gitignore            Missing entries"))
	}

	// Pattern Scan
	if len(m.result.SecretFindings) == 0 {
		lines = append(lines, SuccessStyle.Render("  Pattern Scan         No secrets detected"))
	} else {
		lines = append(lines, ErrorStyle.Render(fmt.Sprintf("  Pattern Scan         %d files with secrets", len(m.result.SecretFindings))))
	}

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// renderFindings renders detailed findings
func (m SecurityModel) renderFindings() string {
	var lines []string

	lines = append(lines, TitleStyle.Render("Detailed Findings"))
	lines = append(lines, "")

	// Encryption issues
	if len(m.result.EncryptionIssues) > 0 {
		lines = append(lines, ErrorStyle.Render("Encryption Issues:"))
		for _, issue := range m.result.EncryptionIssues {
			lines = append(lines, fmt.Sprintf("    %s", issue))
		}
		lines = append(lines, "")
	}

	// Gitignore issues
	if len(m.result.GitignoreIssues) > 0 {
		lines = append(lines, WarningStyle.Render("Gitignore Issues:"))
		for _, pattern := range m.result.GitignoreIssues {
			lines = append(lines, fmt.Sprintf("    Missing: %s", pattern))
		}
		lines = append(lines, "")
	}

	// Secret findings
	if len(m.result.SecretFindings) > 0 {
		lines = append(lines, ErrorStyle.Render("Secret Patterns Found:"))
		for _, file := range m.result.SecretFindings {
			for _, match := range file.Matches {
				lines = append(lines, fmt.Sprintf("    %s:%d  %s", file.Path, match.Line, match.Content))
			}
		}
	}

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// renderOverallResult renders the overall pass/fail result
func (m SecurityModel) renderOverallResult() string {
	separator := strings.Repeat("-", 40)

	var resultLine string
	if m.result != nil && m.result.Passed {
		resultLine = SuccessStyle.Render("Result: PASSED")
	} else {
		resultLine = ErrorStyle.Render("Result: FAILED")
	}

	return BoxStyle.Render(fmt.Sprintf("%s\n%s", separator, resultLine))
}

// renderActionBar renders the horizontal action bar
func (m SecurityModel) renderActionBar() string {
	var actionButtons []string

	shortcuts := []string{"R", "F", "I"}

	for i, action := range m.actions {
		buttonText := fmt.Sprintf("[%s] %s", shortcuts[i], action)

		if i == m.selected {
			actionButtons = append(actionButtons, SelectedStyle.Render(buttonText))
		} else {
			actionButtons = append(actionButtons, MutedStyle.Render(buttonText))
		}
	}

	actionBar := strings.Join(actionButtons, "  ")
	helpText := MutedStyle.Render("h/l: select action | Enter: execute | r/f/i: quick action")

	return BoxStyle.Render(fmt.Sprintf("%s\n\n%s", actionBar, helpText))
}

// SetSize updates the model dimensions
func (m *SecurityModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
