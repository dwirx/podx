package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a single form field
type FormField struct {
	Label       string
	Placeholder string
	Value       string
	Required    bool
	Password    bool // Mask input
	input       textinput.Model
}

// FormDialogResult is sent when form is submitted or cancelled
type FormDialogResult struct {
	Submitted bool
	Values    map[string]string
	CommandID string
}

// FormDialogModel represents a form dialog overlay
type FormDialogModel struct {
	Title       string
	Description string
	Fields      []FormField
	CommandID   string

	focused   int
	submitted bool
	cancelled bool
	width     int
	height    int
	visible   bool
	keys      KeyMap
}

// NewFormDialog creates a new form dialog
func NewFormDialog(title, description string, fields []FormField, commandID string) FormDialogModel {
	// Initialize text inputs for each field
	for i := range fields {
		ti := textinput.New()
		ti.Placeholder = fields[i].Placeholder
		ti.CharLimit = 256
		ti.Width = 40

		if fields[i].Password {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '*'
		}

		if fields[i].Value != "" {
			ti.SetValue(fields[i].Value)
		}

		if i == 0 {
			ti.Focus()
		}

		fields[i].input = ti
	}

	return FormDialogModel{
		Title:       title,
		Description: description,
		Fields:      fields,
		CommandID:   commandID,
		focused:     0,
		visible:     true,
		keys:        DefaultKeyMap(),
	}
}

// Init initializes the form dialog
func (m FormDialogModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles form dialog messages
func (m FormDialogModel) Update(msg tea.Msg) (FormDialogModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// Cancel
		case msg.String() == "esc":
			m.cancelled = true
			m.visible = false
			return m, func() tea.Msg {
				return FormDialogResult{Submitted: false, CommandID: m.CommandID}
			}

		// Submit
		case msg.String() == "enter":
			// If on last field or pressing ctrl+enter, submit
			if m.focused >= len(m.Fields) {
				// Submit button focused
				return m.submit()
			}
			// Move to next field
			return m.nextField()

		// Tab or down to next field
		case msg.String() == "tab", key.Matches(msg, m.keys.Down):
			return m.nextField()

		// Shift+tab or up to previous field
		case msg.String() == "shift+tab", key.Matches(msg, m.keys.Up):
			return m.prevField()

		// Ctrl+Enter to submit from any field
		case msg.String() == "ctrl+s":
			return m.submit()
		}

		// Update focused input
		if m.focused < len(m.Fields) {
			var cmd tea.Cmd
			m.Fields[m.focused].input, cmd = m.Fields[m.focused].input.Update(msg)
			return m, cmd
		}
	}

	// Update focused input for other messages
	if m.focused < len(m.Fields) {
		var cmd tea.Cmd
		m.Fields[m.focused].input, cmd = m.Fields[m.focused].input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// nextField moves focus to next field
func (m FormDialogModel) nextField() (FormDialogModel, tea.Cmd) {
	if m.focused < len(m.Fields) {
		m.Fields[m.focused].input.Blur()
	}

	m.focused++
	if m.focused > len(m.Fields) {
		m.focused = 0
	}

	if m.focused < len(m.Fields) {
		m.Fields[m.focused].input.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

// prevField moves focus to previous field
func (m FormDialogModel) prevField() (FormDialogModel, tea.Cmd) {
	if m.focused < len(m.Fields) {
		m.Fields[m.focused].input.Blur()
	}

	m.focused--
	if m.focused < 0 {
		m.focused = len(m.Fields)
	}

	if m.focused < len(m.Fields) {
		m.Fields[m.focused].input.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

// submit validates and submits the form
func (m FormDialogModel) submit() (FormDialogModel, tea.Cmd) {
	// Validate required fields
	for _, field := range m.Fields {
		if field.Required && strings.TrimSpace(field.input.Value()) == "" {
			// Could show error, for now just don't submit
			return m, nil
		}
	}

	// Collect values
	values := make(map[string]string)
	for _, field := range m.Fields {
		values[field.Label] = field.input.Value()
	}

	m.submitted = true
	m.visible = false

	return m, func() tea.Msg {
		return FormDialogResult{
			Submitted: true,
			Values:    values,
			CommandID: m.CommandID,
		}
	}
}

// View renders the form dialog
func (m FormDialogModel) View() string {
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

	// Description
	if m.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)
		content.WriteString(descStyle.Render(m.Description))
		content.WriteString("\n\n")
	}

	// Fields
	labelStyle := lipgloss.NewStyle().
		Width(15).
		Align(lipgloss.Right).
		MarginRight(2)

	requiredStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	for i, field := range m.Fields {
		label := field.Label
		if field.Required {
			label = label + requiredStyle.Render("*")
		}

		// Highlight focused field
		if i == m.focused {
			label = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Render(label)
		}

		content.WriteString(labelStyle.Render(label))
		content.WriteString(field.input.View())
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Buttons
	cancelBtn := "[ Cancel ]"
	submitBtn := "[ Submit ]"

	if m.focused == len(m.Fields) {
		submitBtn = lipgloss.NewStyle().
			Background(lipgloss.Color("86")).
			Foreground(lipgloss.Color("0")).
			Render("[ Submit ]")
	} else {
		submitBtn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("[ Submit ]")
	}

	cancelBtn = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(cancelBtn)

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", submitBtn)
	buttonContainer := lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Render(buttons)

	content.WriteString(buttonContainer)
	content.WriteString("\n\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center)
	help := "Tab: next field | Shift+Tab: prev | Enter: next/submit | Esc: cancel"
	content.WriteString(helpStyle.Render(help))

	// Dialog box style
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 2).
		Width(70)

	return dialogStyle.Render(content.String())
}

// IsVisible returns whether dialog is visible
func (m FormDialogModel) IsVisible() bool {
	return m.visible
}

// Show makes the dialog visible
func (m *FormDialogModel) Show() {
	m.visible = true
	m.submitted = false
	m.cancelled = false
	m.focused = 0
	if len(m.Fields) > 0 {
		m.Fields[0].input.Focus()
	}
}

// Hide hides the dialog
func (m *FormDialogModel) Hide() {
	m.visible = false
}

// SetSize updates dialog dimensions
func (m *FormDialogModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetValues returns all field values
func (m FormDialogModel) GetValues() map[string]string {
	values := make(map[string]string)
	for _, field := range m.Fields {
		values[field.Label] = field.input.Value()
	}
	return values
}

// CenterDialog centers a dialog in the terminal
func CenterDialog(dialog string, termWidth, termHeight int) string {
	lines := strings.Split(dialog, "\n")
	dialogHeight := len(lines)
	dialogWidth := 0
	for _, line := range lines {
		if len(line) > dialogWidth {
			dialogWidth = len(line)
		}
	}

	// Calculate padding
	topPadding := (termHeight - dialogHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}
	leftPadding := (termWidth - dialogWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	var result strings.Builder
	for i := 0; i < topPadding; i++ {
		result.WriteString("\n")
	}

	leftPad := strings.Repeat(" ", leftPadding)
	for _, line := range lines {
		result.WriteString(leftPad)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// CreateFormForCommand creates appropriate form for a command
func CreateFormForCommand(commandID string) *FormDialogModel {
	switch commandID {
	case "add-recipient":
		form := NewFormDialog(
			"Add Recipient",
			"Add a team member who can decrypt secrets",
			[]FormField{
				{Label: "Name", Placeholder: "Team member name", Required: true},
				{Label: "Key", Placeholder: "age1xxx... (Age public key)", Required: true},
			},
			commandID,
		)
		return &form

	case "encrypt":
		form := NewFormDialog(
			"Encrypt File",
			"Encrypt a single file with password",
			[]FormField{
				{Label: "Input", Placeholder: "Path to file to encrypt", Required: true},
				{Label: "Output", Placeholder: "Path for encrypted file", Required: true},
				{Label: "Password", Placeholder: "Encryption password", Required: true, Password: true},
				{Label: "Confirm", Placeholder: "Confirm password", Required: true, Password: true},
			},
			commandID,
		)
		return &form

	case "decrypt":
		form := NewFormDialog(
			"Decrypt File",
			"Decrypt an encrypted file",
			[]FormField{
				{Label: "Input", Placeholder: "Path to encrypted file", Required: true},
				{Label: "Output", Placeholder: "Path for decrypted file", Required: true},
				{Label: "Password", Placeholder: "Decryption password", Required: true, Password: true},
			},
			commandID,
		)
		return &form

	case "env-encrypt":
		form := NewFormDialog(
			"Encrypt .env File",
			"Encrypt .env file with format-preserving encryption",
			[]FormField{
				{Label: "Input", Placeholder: ".env file path", Required: true, Value: ".env"},
				{Label: "Output", Placeholder: "Output path", Required: true, Value: ".env.podx"},
				{Label: "Password", Placeholder: "Encryption password", Required: true, Password: true},
				{Label: "Confirm", Placeholder: "Confirm password", Required: true, Password: true},
			},
			commandID,
		)
		return &form

	case "env-decrypt":
		form := NewFormDialog(
			"Decrypt .env File",
			"Decrypt encrypted .env file",
			[]FormField{
				{Label: "Input", Placeholder: "Encrypted .env path", Required: true, Value: ".env.podx"},
				{Label: "Output", Placeholder: "Output path", Required: true, Value: ".env"},
				{Label: "Password", Placeholder: "Decryption password", Required: true, Password: true},
			},
			commandID,
		)
		return &form

	case "keygen-age":
		form := NewFormDialog(
			"Generate Age Key",
			"Generate a new Age X25519 key pair",
			[]FormField{},
			commandID,
		)
		return &form

	case "keygen-gpg":
		form := NewFormDialog(
			"Generate GPG Key",
			"Generate a new GPG key pair",
			[]FormField{
				{Label: "Name", Placeholder: "Your full name", Required: true},
				{Label: "Email", Placeholder: "your@email.com", Required: true},
			},
			commandID,
		)
		return &form

	default:
		return nil
	}
}

// CommandNeedsForm returns true if command needs a form dialog
func CommandNeedsForm(commandID string) bool {
	needsForm := map[string]bool{
		"add-recipient": true,
		"encrypt":       true,
		"decrypt":       true,
		"env-encrypt":   true,
		"env-decrypt":   true,
		"keygen-gpg":    true,
	}
	return needsForm[commandID]
}

// BuildCommandArgs builds command arguments from form values
func BuildCommandArgs(commandID string, values map[string]string) []string {
	switch commandID {
	case "add-recipient":
		return []string{"add-recipient", "-n", values["Name"], "-k", values["Key"]}

	case "encrypt":
		return []string{"encrypt", "-i", values["Input"], "-o", values["Output"], "-p", values["Password"]}

	case "decrypt":
		return []string{"decrypt", "-i", values["Input"], "-o", values["Output"], "-p", values["Password"]}

	case "env-encrypt":
		return []string{"env", "encrypt", "-i", values["Input"], "-o", values["Output"], "-p", values["Password"]}

	case "env-decrypt":
		return []string{"env", "decrypt", "-i", values["Input"], "-o", values["Output"], "-p", values["Password"]}

	case "keygen-age":
		return []string{"keygen", "-t", "age"}

	case "keygen-gpg":
		return []string{"keygen", "-t", "gpg", "-n", values["Name"], "-e", values["Email"]}

	default:
		return []string{}
	}
}

// ValidateFormValues validates form values before submission
func ValidateFormValues(commandID string, values map[string]string) error {
	switch commandID {
	case "encrypt", "env-encrypt":
		if values["Password"] != values["Confirm"] {
			return fmt.Errorf("passwords do not match")
		}
		if len(values["Password"]) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}
	case "add-recipient":
		if !strings.HasPrefix(values["Key"], "age1") {
			return fmt.Errorf("invalid Age public key (must start with 'age1')")
		}
	}
	return nil
}
