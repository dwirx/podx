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
		return CardStyle.Copy().Width(m.width - 6).Render(RenderSpinner(0) + " Running security check...")
	}

	var sections []string

	// Get terminal size for responsive layout
	termSize := GetTerminalSize(m.width)
	cardWidth := m.width - 6
	if cardWidth > 100 {
		cardWidth = 100
	}
	if cardWidth < 50 {
		cardWidth = 50
	}

	// Title card
	titleContent := TitleStyle.Render("Security Check")
	sections = append(sections, CardStyle.Copy().Width(cardWidth).Render(titleContent))

	// Check results card
	sections = append(sections, "")
	sections = append(sections, m.renderResultsCard(cardWidth))

	// Detailed findings (if any issues)
	if m.result != nil && !m.result.Passed {
		sections = append(sections, "")
		sections = append(sections, m.renderFindingsCard(cardWidth, termSize))
	}

	// Overall result card
	sections = append(sections, "")
	sections = append(sections, m.renderOverallResultCard(cardWidth))

	// Status message (if any)
	if m.message != "" {
		sections = append(sections, "")
		sections = append(sections, m.message)
	}

	// Action bar
	sections = append(sections, "")
	sections = append(sections, m.renderActionBarCard(cardWidth))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTitle renders the security check title
func (m SecurityModel) renderTitle() string {
	return BoxStyle.Render(TitleStyle.Render("Security Check"))
}

// renderResultsCard renders the security check results summary with card style
func (m SecurityModel) renderResultsCard(width int) string {
	var lines []string

	lines = append(lines, CardTitleStyle.Render("Check Results"))
	lines = append(lines, "")

	if m.result == nil {
		lines = append(lines, MutedStyle.Render("  No results available"))
		return CardStyle.Copy().Width(width).Render(strings.Join(lines, "\n"))
	}

	// Encryption Status
	if len(m.result.EncryptionIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] Encryption       All secrets encrypted"))
	} else {
		lines = append(lines, ErrorStyle.Render("  [!!] Encryption       Issues found"))
	}

	// Gitignore
	if len(m.result.GitignoreIssues) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] Gitignore        Properly configured"))
	} else {
		lines = append(lines, WarningStyle.Render("  [!!] Gitignore        Missing entries"))
	}

	// Pattern Scan
	if len(m.result.SecretFindings) == 0 {
		lines = append(lines, SuccessStyle.Render("  [OK] Pattern Scan     No secrets detected"))
	} else {
		lines = append(lines, ErrorStyle.Render(fmt.Sprintf("  [!!] Pattern Scan     %d files with secrets", len(m.result.SecretFindings))))
	}

	return CardStyle.Copy().Width(width).Render(strings.Join(lines, "\n"))
}

// renderFindingsCard renders detailed findings with card style
func (m SecurityModel) renderFindingsCard(width int, termSize TerminalSize) string {
	var lines []string

	lines = append(lines, CardTitleStyle.Render("Detailed Findings"))
	lines = append(lines, "")

	// Encryption issues
	if len(m.result.EncryptionIssues) > 0 {
		lines = append(lines, ErrorStyle.Render("  Encryption Issues:"))
		for _, issue := range m.result.EncryptionIssues {
			issueText := TruncateText(issue, width-10)
			lines = append(lines, fmt.Sprintf("    %s", issueText))
		}
		lines = append(lines, "")
	}

	// Gitignore issues
	if len(m.result.GitignoreIssues) > 0 {
		lines = append(lines, WarningStyle.Render("  Gitignore Issues:"))
		for _, pattern := range m.result.GitignoreIssues {
			lines = append(lines, fmt.Sprintf("    Missing: %s", pattern))
		}
		lines = append(lines, "")
	}

	// Secret findings
	if len(m.result.SecretFindings) > 0 {
		lines = append(lines, ErrorStyle.Render("  Secret Patterns Found:"))
		maxFindings := 5
		if termSize == TerminalLarge {
			maxFindings = 10
		}
		count := 0
		for _, file := range m.result.SecretFindings {
			if count >= maxFindings {
				remaining := len(m.result.SecretFindings) - count
				lines = append(lines, MutedStyle.Render(fmt.Sprintf("    ... and %d more files", remaining)))
				break
			}
			for _, match := range file.Matches {
				content := TruncateText(match.Content, width-20)
				lines = append(lines, fmt.Sprintf("    %s:%d  %s", file.Path, match.Line, content))
				count++
				if count >= maxFindings {
					break
				}
			}
		}
	}

	return CardStyle.Copy().Width(width).Render(strings.Join(lines, "\n"))
}

// renderOverallResultCard renders the overall pass/fail result with card style
func (m SecurityModel) renderOverallResultCard(width int) string {
	var content string
	if m.result != nil && m.result.Passed {
		content = "  " + BadgeSuccessStyle.Render(" PASSED ") + "  " + SuccessStyle.Render("All security checks passed")
	} else {
		content = "  " + BadgeErrorStyle.Render(" FAILED ") + "  " + ErrorStyle.Render("Some security checks failed")
	}

	return CardStyle.Copy().Width(width).Render(content)
}

// renderActionBarCard renders the horizontal action bar with card style
func (m SecurityModel) renderActionBarCard(width int) string {
	var lines []string

	lines = append(lines, CardTitleStyle.Render("Actions"))
	lines = append(lines, "")

	var actionButtons []string
	shortcuts := []string{"R", "F", "I"}

	for i, action := range m.actions {
		buttonText := fmt.Sprintf("[%s] %s", shortcuts[i], action)

		if i == m.selected {
			actionButtons = append(actionButtons, SelectedStyle.Render(" "+buttonText+" "))
		} else {
			actionButtons = append(actionButtons, MutedStyle.Render(buttonText))
		}
	}

	actionBar := strings.Join(actionButtons, "   ")
	lines = append(lines, "  "+actionBar)
	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("  h/l: select | Enter: execute | r/f/i: quick action"))

	return CardStyle.Copy().Width(width).Render(strings.Join(lines, "\n"))
}

// SetSize updates the model dimensions
func (m *SecurityModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
