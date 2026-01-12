package tui

import (
	"fmt"
	"strings"

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

	inputs    []textinput.Model // Separate slice for inputs
	focused   int
	submitted bool
	cancelled bool
	width     int
	height    int
	visible   bool
	keys      KeyMap
	errMsg    string
}

// NewFormDialog creates a new form dialog
func NewFormDialog(title, description string, fields []FormField, commandID string) FormDialogModel {
	inputs := make([]textinput.Model, len(fields))

	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		ti.CharLimit = 256
		ti.Width = 40

		if field.Password {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}

		if field.Value != "" {
			ti.SetValue(field.Value)
		}

		if i == 0 {
			ti.Focus()
		}

		inputs[i] = ti
	}

	return FormDialogModel{
		Title:       title,
		Description: description,
		Fields:      fields,
		CommandID:   commandID,
		inputs:      inputs,
		focused:     0,
		visible:     true,
		keys:        DefaultKeyMap(),
	}
}

// Init initializes the form dialog
func (m FormDialogModel) Init() tea.Cmd {
	if len(m.inputs) > 0 {
		return m.inputs[0].Focus()
	}
	return nil
}

// Update handles form dialog messages
func (m FormDialogModel) Update(msg tea.Msg) (FormDialogModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Cancel
		case "esc":
			m.cancelled = true
			m.visible = false
			return m, func() tea.Msg {
				return FormDialogResult{Submitted: false, CommandID: m.CommandID}
			}

		// Submit with Ctrl+S
		case "ctrl+s":
			return m.submit()

		// Enter - next field or submit
		case "enter":
			if m.focused >= len(m.inputs) {
				// On submit button
				return m.submit()
			}
			// Move to next field
			return m.nextField()

		// Tab to next field
		case "tab":
			return m.nextField()

		// Shift+Tab to previous field
		case "shift+tab":
			return m.prevField()

		// Down arrow to next field
		case "down":
			if m.focused < len(m.inputs) {
				return m.nextField()
			}

		// Up arrow to previous field
		case "up":
			return m.prevField()
		}
	}

	// Update the focused input
	if m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		return m, cmd
	}

	return m, nil
}

// nextField moves focus to next field
func (m FormDialogModel) nextField() (FormDialogModel, tea.Cmd) {
	// Blur current input
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Blur()
	}

	m.focused++
	if m.focused > len(m.inputs) {
		m.focused = 0
	}

	// Focus new input
	if m.focused < len(m.inputs) {
		return m, m.inputs[m.focused].Focus()
	}

	return m, nil
}

// prevField moves focus to previous field
func (m FormDialogModel) prevField() (FormDialogModel, tea.Cmd) {
	// Blur current input
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Blur()
	}

	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs)
	}

	// Focus new input
	if m.focused < len(m.inputs) {
		return m, m.inputs[m.focused].Focus()
	}

	return m, nil
}

// submit validates and submits the form
func (m FormDialogModel) submit() (FormDialogModel, tea.Cmd) {
	// Validate required fields
	for i, field := range m.Fields {
		if field.Required && strings.TrimSpace(m.inputs[i].Value()) == "" {
			m.errMsg = fmt.Sprintf("Field '%s' is required", field.Label)
			return m, nil
		}
	}

	// Custom validation
	err := m.validate()
	if err != nil {
		m.errMsg = err.Error()
		return m, nil
	}

	// Collect values
	values := make(map[string]string)
	for i, field := range m.Fields {
		values[field.Label] = m.inputs[i].Value()
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

// validate performs custom validation based on command
func (m FormDialogModel) validate() error {
	switch m.CommandID {
	case "encrypt", "env-encrypt":
		// Check password match (assuming Confirm is last field)
		if len(m.inputs) >= 4 {
			pass := m.inputs[2].Value()
			confirm := m.inputs[3].Value()
			if pass != confirm {
				return fmt.Errorf("Passwords do not match")
			}
			if len(pass) < 6 {
				return fmt.Errorf("Password must be at least 6 characters")
			}
		}
	case "add-recipient":
		if len(m.inputs) >= 2 {
			key := m.inputs[1].Value()
			if !strings.HasPrefix(key, "age1") {
				return fmt.Errorf("Invalid Age key (must start with 'age1')")
			}
		}
	}
	return nil
}

// View renders the form dialog
func (m FormDialogModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title using PODX color palette
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		MarginBottom(1)
	content.WriteString(titleStyle.Render("📝 " + m.Title))
	content.WriteString("\n\n")

	// Description
	if m.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
		content.WriteString(descStyle.Render(m.Description))
		content.WriteString("\n\n")
	}

	// Fields
	labelWidth := 12
	for _, field := range m.Fields {
		if len(field.Label) > labelWidth {
			labelWidth = len(field.Label)
		}
	}

	for i, field := range m.Fields {
		label := field.Label
		if field.Required {
			label = label + " *"
		}

		// Style for label
		labelStyle := lipgloss.NewStyle().
			Width(labelWidth + 2).
			Align(lipgloss.Right).
			Foreground(ColorWhite)

		// Highlight focused field
		if i == m.focused {
			labelStyle = labelStyle.Foreground(ColorPrimary).Bold(true)
		}

		// Input styling
		inputView := m.inputs[i].View()

		content.WriteString(labelStyle.Render(label + ": "))
		content.WriteString(inputView)
		content.WriteString("\n")
	}

	// Error message
	if m.errMsg != "" {
		content.WriteString("\n")
		errStyle := lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
		content.WriteString(errStyle.Render("⚠ " + m.errMsg))
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Buttons
	cancelBtn := "[ Cancel ]"
	submitBtn := "[ Submit ]"

	btnInactive := lipgloss.NewStyle().
		Foreground(ColorMuted)

	btnActive := lipgloss.NewStyle().
		Background(ColorSuccess).
		Foreground(lipgloss.Color("#000")).
		Bold(true).
		Padding(0, 1)

	if m.focused == len(m.inputs) {
		submitBtn = btnActive.Render("Submit")
		cancelBtn = btnInactive.Render(cancelBtn)
	} else {
		submitBtn = btnInactive.Render(submitBtn)
		cancelBtn = btnInactive.Render(cancelBtn)
	}

	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "   ", submitBtn)
	buttonContainer := lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		MarginTop(1)
	content.WriteString(buttonContainer.Render(buttonRow))

	content.WriteString("\n\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(ColorMuted)
	help := "Tab/↓: next • Shift+Tab/↑: prev • Enter: next/submit • Esc: cancel"
	content.WriteString(helpStyle.Render(help))

	// Dialog box style with PODX theme
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
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
	m.errMsg = ""
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
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
	for i, field := range m.Fields {
		values[field.Label] = m.inputs[i].Value()
	}
	return values
}

// CenterDialog centers a dialog in the terminal with a proper background overlay
func CenterDialog(dialog string, termWidth, termHeight int) string {
	// Ensure minimum dimensions
	if termWidth < 80 {
		termWidth = 80
	}
	if termHeight < 24 {
		termHeight = 24
	}

	// Use lipgloss.Place to center the dialog with a background
	// This properly handles ANSI escape codes and centering
	overlay := lipgloss.Place(
		termWidth,
		termHeight,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceBackground(ColorBg),
		lipgloss.WithWhitespaceForeground(ColorBg),
	)

	return overlay
}

// CreateFormForCommand creates appropriate form for a command
func CreateFormForCommand(commandID string) *FormDialogModel {
	switch commandID {
	case "add-recipient":
		form := NewFormDialog(
			"Add Recipient",
			"Add a team member who can decrypt secrets",
			[]FormField{
				{Label: "Name", Placeholder: "e.g. John Doe", Required: true},
				{Label: "Key", Placeholder: "age1xxxxxxxxx...", Required: true},
			},
			commandID,
		)
		return &form

	case "encrypt":
		form := NewFormDialog(
			"Encrypt File",
			"Encrypt a single file with password (AES-256-GCM)",
			[]FormField{
				{Label: "Input", Placeholder: "path/to/file.txt", Required: true},
				{Label: "Output", Placeholder: "path/to/file.enc", Required: true},
				{Label: "Password", Placeholder: "Enter password", Required: true, Password: true},
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
				{Label: "Input", Placeholder: "path/to/file.enc", Required: true},
				{Label: "Output", Placeholder: "path/to/file.txt", Required: true},
				{Label: "Password", Placeholder: "Enter password", Required: true, Password: true},
			},
			commandID,
		)
		return &form

	case "env-encrypt":
		form := NewFormDialog(
			"Encrypt .env File",
			"Format-preserving encryption for .env files",
			[]FormField{
				{Label: "Input", Placeholder: ".env", Required: true, Value: ".env"},
				{Label: "Output", Placeholder: ".env.podx", Required: true, Value: ".env.podx"},
				{Label: "Password", Placeholder: "Enter password", Required: true, Password: true},
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
				{Label: "Input", Placeholder: ".env.podx", Required: true, Value: ".env.podx"},
				{Label: "Output", Placeholder: ".env", Required: true, Value: ".env"},
				{Label: "Password", Placeholder: "Enter password", Required: true, Password: true},
			},
			commandID,
		)
		return &form

	case "keygen-age":
		// No form needed, execute directly
		return nil

	case "keygen-gpg":
		form := NewFormDialog(
			"Generate GPG Key",
			"Generate a new GPG key pair",
			[]FormField{
				{Label: "Name", Placeholder: "Your Full Name", Required: true},
				{Label: "Email", Placeholder: "you@example.com", Required: true},
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
