package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FormField represents a form input field
type FormField struct {
	Label       string
	Placeholder string
	IsPassword  bool
	Value       string
}

// FormModel is a generic form model
type FormModel struct {
	title      string
	fields     []textinput.Model
	labels     []string
	focusIndex int
	submitted  bool
	cancelled  bool
	err        error
}

// NewFormModel creates a new form
func NewFormModel(title string, fields []FormField) FormModel {
	inputs := make([]textinput.Model, len(fields))
	labels := make([]string, len(fields))

	for i, f := range fields {
		t := textinput.New()
		t.Placeholder = f.Placeholder
		t.CharLimit = 256
		t.Width = 40

		if f.IsPassword {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		if f.Value != "" {
			t.SetValue(f.Value)
		}

		if i == 0 {
			t.Focus()
		}

		inputs[i] = t
		labels[i] = f.Label
	}

	return FormModel{
		title:  title,
		fields: inputs,
		labels: labels,
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, nil

		case "tab", "down":
			m.focusIndex++
			if m.focusIndex >= len(m.fields) {
				m.focusIndex = 0
			}
			return m, m.updateFocus()

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.fields) - 1
			}
			return m, m.updateFocus()

		case "enter":
			if m.focusIndex == len(m.fields)-1 {
				// Last field, submit
				m.submitted = true
				return m, nil
			}
			// Move to next field
			m.focusIndex++
			return m, m.updateFocus()
		}
	}

	// Update the focused input
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *FormModel) updateFocus() tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.fields {
		if i == m.focusIndex {
			cmds = append(cmds, m.fields[i].Focus())
		} else {
			m.fields[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m *FormModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.fields {
		var cmd tea.Cmd
		m.fields[i], cmd = m.fields[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Fields
	for i, field := range m.fields {
		label := m.labels[i]
		if i == m.focusIndex {
			b.WriteString(FocusedLabelStyle.Render(label))
		} else {
			b.WriteString(LabelStyle.Render(label))
		}
		b.WriteString("\n")
		b.WriteString(field.View())
		b.WriteString("\n\n")
	}

	// Buttons
	submitBtn := ButtonStyle.Render("Submit")
	cancelBtn := ButtonStyle.Render("Cancel")
	if m.focusIndex == len(m.fields)-1 {
		submitBtn = ActiveButtonStyle.Render("Submit")
	}

	b.WriteString("\n")
	b.WriteString(submitBtn)
	b.WriteString(cancelBtn)
	b.WriteString("\n\n")

	// Help
	b.WriteString(HelpStyle.Render("tab/↓: next • shift+tab/↑: prev • enter: submit • esc: cancel"))

	return BoxStyle.Render(b.String())
}

// Values returns all field values
func (m FormModel) Values() []string {
	values := make([]string, len(m.fields))
	for i, f := range m.fields {
		values[i] = f.Value()
	}
	return values
}

// Submitted returns true if form was submitted
func (m FormModel) Submitted() bool {
	return m.submitted
}

// Cancelled returns true if form was cancelled
func (m FormModel) Cancelled() bool {
	return m.cancelled
}

// SetError sets an error message
func (m *FormModel) SetError(err error) {
	m.err = err
}

// Error returns current error
func (m FormModel) Error() error {
	return m.err
}

// ResultModel shows operation result
type ResultModel struct {
	title   string
	message string
	success bool
	done    bool
}

// NewResultModel creates a result display
func NewResultModel(title, message string, success bool) ResultModel {
	return ResultModel{
		title:   title,
		message: message,
		success: success,
	}
}

func (m ResultModel) Init() tea.Cmd {
	return nil
}

func (m ResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q":
			m.done = true
			return m, nil
		}
	}
	return m, nil
}

func (m ResultModel) View() string {
	var b strings.Builder

	// Icon and title
	var icon, titleStyled string
	if m.success {
		icon = "✅"
		titleStyled = SuccessStyle.Render(m.title)
	} else {
		icon = "❌"
		titleStyled = ErrorStyle.Render(m.title)
	}

	b.WriteString(fmt.Sprintf("\n  %s %s\n\n", icon, titleStyled))
	b.WriteString(fmt.Sprintf("  %s\n\n", m.message))
	b.WriteString(HelpStyle.Render("  Press Enter or Esc to continue"))

	return BoxStyle.Render(b.String())
}

// Done returns true if user acknowledged result
func (m ResultModel) Done() bool {
	return m.done
}
