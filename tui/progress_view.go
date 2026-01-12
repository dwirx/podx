package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressUpdateMsg updates the progress bar
type ProgressUpdateMsg struct {
	Current int
	Total   int
	Message string
}

// ProgressCompleteMsg signals progress is complete
type ProgressCompleteMsg struct {
	Success bool
	Message string
	Output  string
}

// ProgressViewModel represents a progress bar view
type ProgressViewModel struct {
	Title   string
	Message string
	Current int
	Total   int
	Output  string

	complete bool
	success  bool
	visible  bool
	width    int
	height   int
}

// NewProgressView creates a new progress view
func NewProgressView(title string) ProgressViewModel {
	return ProgressViewModel{
		Title:   title,
		visible: true,
	}
}

// Init initializes the progress view
func (m ProgressViewModel) Init() tea.Cmd {
	return nil
}

// Update handles progress view messages
func (m ProgressViewModel) Update(msg tea.Msg) (ProgressViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressUpdateMsg:
		m.Current = msg.Current
		m.Total = msg.Total
		m.Message = msg.Message
		return m, nil

	case ProgressCompleteMsg:
		m.complete = true
		m.success = msg.Success
		m.Message = msg.Message
		m.Output = msg.Output
		return m, nil

	case tea.KeyMsg:
		// If complete, allow closing
		if m.complete {
			switch msg.String() {
			case "enter", "esc", "q":
				m.visible = false
				return m, nil
			}
		}
	}

	return m, nil
}

// View renders the progress view
func (m ProgressViewModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginBottom(1)
	content.WriteString(titleStyle.Render(m.Title))
	content.WriteString("\n\n")

	// Progress bar
	if !m.complete && m.Total > 0 {
		barWidth := 40
		filled := int(float64(m.Current) / float64(m.Total) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}

		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		percent := int(float64(m.Current) / float64(m.Total) * 100)

		barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
		content.WriteString(barStyle.Render(bar))
		content.WriteString(fmt.Sprintf(" %d%%\n", percent))
	} else if !m.complete {
		// Indeterminate spinner
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		// Use a simple animation based on message length
		idx := len(m.Message) % len(spinner)
		content.WriteString(spinner[idx] + " Processing...\n")
	}

	// Message
	if m.Message != "" {
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		content.WriteString("\n")
		content.WriteString(msgStyle.Render(m.Message))
		content.WriteString("\n")
	}

	// Output (if any)
	if m.Output != "" {
		content.WriteString("\n")
		outputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MaxHeight(10)
		lines := strings.Split(m.Output, "\n")
		if len(lines) > 10 {
			lines = lines[len(lines)-10:]
		}
		content.WriteString(outputStyle.Render(strings.Join(lines, "\n")))
		content.WriteString("\n")
	}

	// Complete status
	if m.complete {
		content.WriteString("\n")
		if m.success {
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")). // Green
				Bold(true)
			content.WriteString(successStyle.Render("✓ Complete"))
		} else {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")). // Red
				Bold(true)
			content.WriteString(errorStyle.Render("✗ Failed"))
		}
		content.WriteString("\n\n")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		content.WriteString(helpStyle.Render("Press Enter or Esc to close"))
	}

	// Box style
	borderColor := lipgloss.Color("86")
	if m.complete && !m.success {
		borderColor = lipgloss.Color("196")
	}

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(60)

	return dialogStyle.Render(content.String())
}

// IsVisible returns whether the view is visible
func (m ProgressViewModel) IsVisible() bool {
	return m.visible
}

// IsComplete returns whether progress is complete
func (m ProgressViewModel) IsComplete() bool {
	return m.complete
}

// Show makes the view visible
func (m *ProgressViewModel) Show() {
	m.visible = true
	m.complete = false
	m.Current = 0
	m.Total = 0
	m.Output = ""
}

// Hide hides the view
func (m *ProgressViewModel) Hide() {
	m.visible = false
}

// SetSize updates view dimensions
func (m *ProgressViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetProgress updates progress
func (m *ProgressViewModel) SetProgress(current, total int, message string) {
	m.Current = current
	m.Total = total
	m.Message = message
}

// Complete marks progress as complete
func (m *ProgressViewModel) Complete(success bool, message string) {
	m.complete = true
	m.success = success
	m.Message = message
}

// AppendOutput adds output text
func (m *ProgressViewModel) AppendOutput(text string) {
	if m.Output != "" {
		m.Output += "\n"
	}
	m.Output += text
}
