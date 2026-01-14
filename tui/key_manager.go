package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hades/podx/keygen"
)

// KeyManagerState represents the current state of the key manager
type KeyManagerState int

const (
	KeyManagerStateList KeyManagerState = iota
	KeyManagerStateImport
	KeyManagerStateImportName
	KeyManagerStateGenerate
	KeyManagerStateGenerateName
	KeyManagerStateConfirmDelete
	KeyManagerStateRename
)

// KeyManagerModel represents the key management dialog
type KeyManagerModel struct {
	visible    bool
	state      KeyManagerState
	keys       []keygen.AgeKeyEntry
	cursor     int
	width      int
	height     int
	errorMsg   string
	successMsg string

	// Import key fields
	importInput textinput.Model

	// Name input field
	nameInput textinput.Model

	// Generated key fields
	generatedPublicKey  string
	generatedPrivateKey string
	generatedKeyName    string
	generatingKey       bool

	// Temporary storage for imported key before naming
	pendingImportKey string

	// Default key tracking
	defaultKey string
}

// keyManagerResultMsg is sent when an operation completes
type keyManagerResultMsg struct {
	success bool
	message string
}

// NewKeyManagerModel creates a new key manager model
func NewKeyManagerModel() KeyManagerModel {
	importInput := textinput.New()
	importInput.Placeholder = "AGE-SECRET-KEY-..."
	importInput.CharLimit = 256
	importInput.Width = 60

	nameInput := textinput.New()
	nameInput.Placeholder = "e.g., John Doe, Team Lead, Production Key..."
	nameInput.CharLimit = 50
	nameInput.Width = 50

	return KeyManagerModel{
		visible:     false,
		state:       KeyManagerStateList,
		importInput: importInput,
		nameInput:   nameInput,
	}
}

// Show displays the key manager dialog
func (m *KeyManagerModel) Show(width, height int) tea.Cmd {
	m.visible = true
	m.state = KeyManagerStateList
	m.width = width
	m.height = height
	m.errorMsg = ""
	m.successMsg = ""
	m.cursor = 0

	// Load keys
	m.loadKeys()

	return nil
}

// Hide hides the dialog
func (m *KeyManagerModel) Hide() {
	m.visible = false
	m.importInput.SetValue("")
	m.nameInput.SetValue("")
	m.pendingImportKey = ""
	m.generatedKeyName = ""
	m.errorMsg = ""
	m.successMsg = ""
}

// IsVisible returns whether the dialog is visible
func (m KeyManagerModel) IsVisible() bool {
	return m.visible
}

// loadKeys loads all keys from the key store
func (m *KeyManagerModel) loadKeys() {
	keys, err := keygen.ListAgeKeys()
	if err != nil {
		m.errorMsg = err.Error()
		return
	}
	m.keys = keys

	// Get default key
	if pubKey, err := keygen.LoadAgeRecipient(); err == nil {
		m.defaultKey = pubKey
	}
}

// Update handles messages for the dialog
func (m KeyManagerModel) Update(msg tea.Msg) (KeyManagerModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case keyManagerResultMsg:
		if msg.success {
			m.successMsg = msg.message
			m.errorMsg = ""
		} else {
			m.errorMsg = msg.message
			m.successMsg = ""
		}
		m.loadKeys()
		return m, nil

	case keygenCompleteMsg:
		m.generatingKey = false
		if msg.success {
			m.generatedPublicKey = msg.publicKey
			m.generatedPrivateKey = msg.privateKey
			m.successMsg = "Key generated successfully!"
			m.loadKeys()
		} else {
			m.errorMsg = msg.errorMsg
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case KeyManagerStateList:
			return m.updateList(msg)
		case KeyManagerStateImport:
			return m.updateImport(msg)
		case KeyManagerStateImportName:
			return m.updateImportName(msg)
		case KeyManagerStateGenerate:
			return m.updateGenerate(msg)
		case KeyManagerStateGenerateName:
			return m.updateGenerateName(msg)
		case KeyManagerStateConfirmDelete:
			return m.updateConfirmDelete(msg)
		case KeyManagerStateRename:
			return m.updateRename(msg)
		}
	}

	return m, nil
}

// updateList handles key presses in list state
func (m KeyManagerModel) updateList(msg tea.KeyMsg) (KeyManagerModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.Hide()
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.keys)-1 {
			m.cursor++
		}
		return m, nil

	case "enter", " ":
		// Set as default key
		if len(m.keys) > 0 && m.cursor < len(m.keys) {
			if err := keygen.SetDefaultKey(m.cursor); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.successMsg = "Key set as default"
				m.defaultKey = m.keys[m.cursor].PublicKey
			}
		}
		return m, nil

	case "i", "I":
		// Import key
		m.state = KeyManagerStateImport
		m.importInput.SetValue("")
		m.importInput.Focus()
		m.errorMsg = ""
		m.successMsg = ""
		return m, textinput.Blink

	case "g", "G":
		// Generate new key - first ask for name
		m.state = KeyManagerStateGenerateName
		m.nameInput.SetValue("")
		m.nameInput.Focus()
		m.generatedPublicKey = ""
		m.generatedPrivateKey = ""
		m.generatedKeyName = ""
		m.generatingKey = false
		m.errorMsg = ""
		m.successMsg = ""
		return m, textinput.Blink

	case "r", "R":
		// Rename key
		if len(m.keys) > 0 && m.cursor < len(m.keys) {
			m.state = KeyManagerStateRename
			m.nameInput.SetValue(m.keys[m.cursor].Name)
			m.nameInput.Focus()
			m.errorMsg = ""
			m.successMsg = ""
			return m, textinput.Blink
		}
		return m, nil

	case "d", "D", "delete", "backspace":
		// Delete key
		if len(m.keys) > 0 && m.cursor < len(m.keys) {
			m.state = KeyManagerStateConfirmDelete
			m.errorMsg = ""
			m.successMsg = ""
		}
		return m, nil
	}

	return m, nil
}

// updateImport handles key presses in import state
func (m KeyManagerModel) updateImport(msg tea.KeyMsg) (KeyManagerModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = KeyManagerStateList
		m.importInput.Blur()
		return m, nil

	case "enter":
		privateKey := strings.TrimSpace(m.importInput.Value())
		if privateKey == "" {
			m.errorMsg = "Please enter a private key"
			return m, nil
		}

		// Validate key format first
		if !strings.HasPrefix(privateKey, "AGE-SECRET-KEY-") {
			m.errorMsg = "Invalid key format (must start with 'AGE-SECRET-KEY-')"
			return m, nil
		}

		// Store the key and go to name entry
		m.pendingImportKey = privateKey
		m.importInput.Blur()
		m.state = KeyManagerStateImportName
		m.nameInput.SetValue("")
		m.nameInput.Focus()
		m.errorMsg = ""
		return m, textinput.Blink

	default:
		var cmd tea.Cmd
		m.importInput, cmd = m.importInput.Update(msg)
		m.errorMsg = "" // Clear error on typing
		return m, cmd
	}
}

// updateGenerate handles key presses in generate state
func (m KeyManagerModel) updateGenerate(msg tea.KeyMsg) (KeyManagerModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.generatingKey {
			return m, nil
		}
		m.state = KeyManagerStateList
		m.generatedPublicKey = ""
		m.generatedPrivateKey = ""
		return m, nil

	case "enter", "y", "Y":
		if m.generatedPublicKey != "" {
			// Key already generated, go back to list
			m.state = KeyManagerStateList
			return m, nil
		}
		if !m.generatingKey {
			m.generatingKey = true
			m.errorMsg = ""
			return m, m.doGenerateKey()
		}

	case "n", "N":
		if m.generatedPublicKey == "" && !m.generatingKey {
			m.state = KeyManagerStateList
		}
		return m, nil
	}

	return m, nil
}

// updateImportName handles key presses in import name state
func (m KeyManagerModel) updateImportName(msg tea.KeyMsg) (KeyManagerModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = KeyManagerStateImport
		m.nameInput.Blur()
		m.nameInput.SetValue("")
		m.importInput.Focus()
		return m, textinput.Blink

	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		// Name is optional, so we can proceed even if empty

		result, err := keygen.ImportAgeKeyWithName(m.pendingImportKey, name)
		if err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		displayName := result.Name
		if displayName == "" {
			displayName = truncateKey(result.PublicKey)
		}
		m.successMsg = fmt.Sprintf("Imported key: %s", displayName)
		m.state = KeyManagerStateList
		m.nameInput.Blur()
		m.nameInput.SetValue("")
		m.importInput.SetValue("")
		m.pendingImportKey = ""
		m.loadKeys()
		return m, nil

	default:
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		m.errorMsg = "" // Clear error on typing
		return m, cmd
	}
}

// updateGenerateName handles key presses in generate name state
func (m KeyManagerModel) updateGenerateName(msg tea.KeyMsg) (KeyManagerModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = KeyManagerStateList
		m.nameInput.Blur()
		m.nameInput.SetValue("")
		return m, nil

	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		m.generatedKeyName = name
		m.nameInput.Blur()
		m.state = KeyManagerStateGenerate
		return m, nil

	default:
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}
}

// updateRename handles key presses in rename state
func (m KeyManagerModel) updateRename(msg tea.KeyMsg) (KeyManagerModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = KeyManagerStateList
		m.nameInput.Blur()
		m.nameInput.SetValue("")
		return m, nil

	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			m.errorMsg = "Name cannot be empty"
			return m, nil
		}

		if err := keygen.RenameKey(m.cursor, name); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}

		m.successMsg = fmt.Sprintf("Key renamed to: %s", name)
		m.state = KeyManagerStateList
		m.nameInput.Blur()
		m.nameInput.SetValue("")
		m.loadKeys()
		return m, nil

	default:
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		m.errorMsg = "" // Clear error on typing
		return m, cmd
	}
}

// updateConfirmDelete handles key presses in confirm delete state
func (m KeyManagerModel) updateConfirmDelete(msg tea.KeyMsg) (KeyManagerModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		m.state = KeyManagerStateList
		return m, nil

	case "y", "Y":
		if err := keygen.DeleteKey(m.cursor); err != nil {
			m.errorMsg = err.Error()
		} else {
			m.successMsg = "Key deleted"
			if m.cursor >= len(m.keys)-1 && m.cursor > 0 {
				m.cursor--
			}
		}
		m.loadKeys()
		m.state = KeyManagerStateList
		return m, nil
	}

	return m, nil
}

// doGenerateKey generates a new Age key pair
func (m KeyManagerModel) doGenerateKey() tea.Cmd {
	name := m.generatedKeyName
	return func() tea.Msg {
		result, err := keygen.GenerateAgeWithName(name)
		if err != nil {
			return keygenCompleteMsg{
				success:  false,
				errorMsg: err.Error(),
			}
		}

		return keygenCompleteMsg{
			success:    true,
			publicKey:  result.PublicKey,
			privateKey: result.PrivateKey,
			keyFile:    result.KeyFile,
		}
	}
}

// View renders the dialog
func (m KeyManagerModel) View() string {
	if !m.visible {
		return ""
	}

	var content string

	switch m.state {
	case KeyManagerStateList:
		content = m.renderList()
	case KeyManagerStateImport:
		content = m.renderImport()
	case KeyManagerStateImportName:
		content = m.renderImportName()
	case KeyManagerStateGenerate:
		content = m.renderGenerate()
	case KeyManagerStateGenerateName:
		content = m.renderGenerateName()
	case KeyManagerStateConfirmDelete:
		content = m.renderConfirmDelete()
	case KeyManagerStateRename:
		content = m.renderRename()
	}

	// Dialog box styling
	dialogWidth := 65
	if m.width > 0 {
		dialogWidth = ResponsiveWidth(m.width, 70, 55, 80)
	}

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Background(ColorBg).
		Padding(1, 2).
		Width(dialogWidth)

	return dialogStyle.Render(content)
}

// renderList renders the key list view
func (m KeyManagerModel) renderList() string {
	var s strings.Builder

	// Title
	boxStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	s.WriteString(boxStyle.Render("╔═══════════════════════════════════════════╗"))
	s.WriteString("\n")
	s.WriteString(boxStyle.Render("║       🔑 Key Manager                      ║"))
	s.WriteString("\n")
	s.WriteString(boxStyle.Render("╚═══════════════════════════════════════════╝"))
	s.WriteString("\n\n")

	// Key count
	keyCountStyle := lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)
	s.WriteString(keyCountStyle.Render(fmt.Sprintf("You have %d key(s)", len(m.keys))))
	s.WriteString("\n\n")

	if len(m.keys) == 0 {
		s.WriteString(MutedStyle.Render("  No keys found."))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Press [G] to generate a new key"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Press [I] to import a key"))
		s.WriteString("\n\n")
	} else {
		// Key list
		for i, k := range m.keys {
			// Cursor
			cursor := "  "
			if i == m.cursor {
				cursor = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("> ")
			}

			// Default indicator
			isDefault := k.PublicKey == m.defaultKey
			defaultBadge := ""
			if isDefault {
				defaultBadge = BadgeSuccessStyle.Render(" DEFAULT ")
			}

			// Name style (prominent)
			nameStyle := lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)
			if i == m.cursor {
				nameStyle = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
			}

			// Key preview (truncated)
			keyPreview := truncateKey(k.PublicKey)
			keyStyle := MutedStyle
			if i == m.cursor {
				keyStyle = lipgloss.NewStyle().Foreground(ColorFg)
			}

			// Display name prominently
			displayName := k.Name
			if displayName == "" {
				displayName = fmt.Sprintf("Key #%d", i+1)
			}

			s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, nameStyle.Render(displayName), defaultBadge))
			s.WriteString(keyStyle.Render(fmt.Sprintf("     %s\n", keyPreview)))

			// Show created date if available
			if k.CreatedAt != "" {
				createdStr := k.CreatedAt
				if len(createdStr) > 10 {
					createdStr = createdStr[:10]
				}
				s.WriteString(MutedStyle.Render(fmt.Sprintf("     Created: %s\n", createdStr)))
			}
		}
		s.WriteString("\n")
	}

	// Success/error messages
	if m.successMsg != "" {
		s.WriteString(SuccessStyle.Render("✓ " + m.successMsg))
		s.WriteString("\n\n")
	}
	if m.errorMsg != "" {
		s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
		s.WriteString("\n\n")
	}

	// Actions
	s.WriteString(CardTitleStyle.Render("Actions:"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  [Enter/Space] Set as default"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  [G] Generate new key"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  [I] Import key"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  [R] Rename key"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  [D] Delete key"))
	s.WriteString("\n\n")

	// Footer
	s.WriteString(MutedStyle.Render("[Esc] Close"))

	return s.String()
}

// renderImport renders the import key view
func (m KeyManagerModel) renderImport() string {
	var s strings.Builder

	// Title
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("📥 Import Private Key"))
	s.WriteString("\n\n")

	// Description
	s.WriteString(MutedStyle.Render("Paste your Age private key below."))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("The key should start with 'AGE-SECRET-KEY-'"))
	s.WriteString("\n\n")

	// Input field
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("Private Key:"))
	s.WriteString("\n")
	s.WriteString(m.importInput.View())
	s.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
		s.WriteString("\n\n")
	}

	// Help
	s.WriteString(MutedStyle.Render("The public key will be derived automatically."))
	s.WriteString("\n\n")

	// Footer
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter]"))
	s.WriteString(MutedStyle.Render(" Import  "))
	s.WriteString(MutedStyle.Render("[Esc] Cancel"))

	return s.String()
}

// renderImportName renders the name input view for imported key
func (m KeyManagerModel) renderImportName() string {
	var s strings.Builder

	// Title
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("📥 Import Private Key - Name"))
	s.WriteString("\n\n")

	// Description
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("✓ Key validated!"))
	s.WriteString("\n\n")

	s.WriteString(MutedStyle.Render("Give this key a name to identify it (e.g., team member name)."))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("Leave empty for an auto-generated name."))
	s.WriteString("\n\n")

	// Input field
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("Key Name:"))
	s.WriteString("\n")
	s.WriteString(m.nameInput.View())
	s.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
		s.WriteString("\n\n")
	}

	// Footer
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter]"))
	s.WriteString(MutedStyle.Render(" Confirm  "))
	s.WriteString(MutedStyle.Render("[Esc] Back"))

	return s.String()
}

// renderGenerateName renders the name input view for generating key
func (m KeyManagerModel) renderGenerateName() string {
	var s strings.Builder

	// Title
	boxStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	s.WriteString(boxStyle.Render("╔════════════════════════════════════════╗"))
	s.WriteString("\n")
	s.WriteString(boxStyle.Render("║     🔑 Generate New Key Pair           ║"))
	s.WriteString("\n")
	s.WriteString(boxStyle.Render("╚════════════════════════════════════════╝"))
	s.WriteString("\n\n")

	// Description
	s.WriteString(MutedStyle.Render("Give this key a name to identify it later."))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("For example: Your name, team role, or purpose."))
	s.WriteString("\n\n")

	// Input field
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("Key Name:"))
	s.WriteString("\n")
	s.WriteString(m.nameInput.View())
	s.WriteString("\n\n")

	// Help
	s.WriteString(MutedStyle.Render("Leave empty for an auto-generated name."))
	s.WriteString("\n\n")

	// Footer
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter]"))
	s.WriteString(MutedStyle.Render(" Continue  "))
	s.WriteString(MutedStyle.Render("[Esc] Cancel"))

	return s.String()
}

// renderRename renders the rename key view
func (m KeyManagerModel) renderRename() string {
	var s strings.Builder

	// Title
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("✏️ Rename Key"))
	s.WriteString("\n\n")

	// Current key info
	if m.cursor < len(m.keys) {
		keyPreview := truncateKey(m.keys[m.cursor].PublicKey)
		s.WriteString(MutedStyle.Render("Key: "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render(keyPreview))
		s.WriteString("\n\n")
	}

	// Description
	s.WriteString(MutedStyle.Render("Enter a new name for this key:"))
	s.WriteString("\n\n")

	// Input field
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("New Name:"))
	s.WriteString("\n")
	s.WriteString(m.nameInput.View())
	s.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
		s.WriteString("\n\n")
	}

	// Footer
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter]"))
	s.WriteString(MutedStyle.Render(" Rename  "))
	s.WriteString(MutedStyle.Render("[Esc] Cancel"))

	return s.String()
}

// renderGenerate renders the key generation view
func (m KeyManagerModel) renderGenerate() string {
	var s strings.Builder

	// Title
	boxStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	s.WriteString(boxStyle.Render("╔════════════════════════════════════════╗"))
	s.WriteString("\n")
	s.WriteString(boxStyle.Render("║     🔑 Generate New Key Pair           ║"))
	s.WriteString("\n")
	s.WriteString(boxStyle.Render("╚════════════════════════════════════════╝"))
	s.WriteString("\n\n")

	if m.generatingKey {
		s.WriteString(RenderSpinner(0))
		s.WriteString(" ")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("Generating key pair..."))
		s.WriteString("\n\n")
		s.WriteString(MutedStyle.Render("Creating X25519 key pair with secure random..."))
		return s.String()
	}

	if m.generatedPublicKey != "" {
		// Success header
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("✓ Key Generated Successfully!"))
		s.WriteString("\n\n")

		// Public Key box
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("📤 Public Key:"))
		s.WriteString("\n")
		pubKeyBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(0, 1).
			Render(m.generatedPublicKey)
		s.WriteString(pubKeyBox)
		s.WriteString("\n\n")

		// Private Key box
		s.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("🔐 Private Key:"))
		s.WriteString("\n")
		privKeyBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Padding(0, 1).
			Render(m.generatedPrivateKey)
		s.WriteString(privKeyBox)
		s.WriteString("\n\n")

		// Storage info
		s.WriteString(MutedStyle.Render("Saved to: ~/.config/podx/age-keys.txt"))
		s.WriteString("\n\n")

		// Warning
		s.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render("⚠️  Never share your private key!"))
		s.WriteString("\n\n")

		// Footer
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter]"))
		s.WriteString(MutedStyle.Render(" Done"))
	} else {
		// Confirmation prompt - show name if provided
		if m.generatedKeyName != "" {
			s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render("Name: "))
			s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Render(m.generatedKeyName))
			s.WriteString("\n\n")
		}

		s.WriteString(MutedStyle.Render("This will generate a new Age X25519 key pair."))
		s.WriteString("\n\n")

		s.WriteString(CardTitleStyle.Render("📁 Keys will be saved to:"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  • "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render("~/.config/podx/age-keys.txt"))
		s.WriteString("\n\n")

		if m.errorMsg != "" {
			s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
			s.WriteString("\n\n")
		}

		// Confirmation buttons
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("Generate new key pair?"))
		s.WriteString("\n\n")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Y/Enter]"))
		s.WriteString(MutedStyle.Render(" Generate  "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("[N/Esc]"))
		s.WriteString(MutedStyle.Render(" Cancel"))
	}

	return s.String()
}

// renderConfirmDelete renders the delete confirmation view
func (m KeyManagerModel) renderConfirmDelete() string {
	var s strings.Builder

	// Title
	s.WriteString(lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("⚠️  Delete Key"))
	s.WriteString("\n\n")

	if m.cursor < len(m.keys) {
		key := m.keys[m.cursor]
		keyPreview := truncateKey(key.PublicKey)

		s.WriteString(MutedStyle.Render("Are you sure you want to delete this key?"))
		s.WriteString("\n\n")

		// Show key name prominently
		displayName := key.Name
		if displayName == "" {
			displayName = fmt.Sprintf("Key #%d", m.cursor+1)
		}
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render(displayName))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render(keyPreview))
		s.WriteString("\n\n")

		s.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render("⚠️  This action cannot be undone!"))
		s.WriteString("\n\n")

		if m.errorMsg != "" {
			s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
			s.WriteString("\n\n")
		}
	}

	// Buttons
	s.WriteString(lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("[Y]"))
	s.WriteString(MutedStyle.Render(" Yes, delete  "))
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[N/Esc]"))
	s.WriteString(MutedStyle.Render(" Cancel"))

	return s.String()
}

// truncateKey truncates a key for display
func truncateKey(key string) string {
	if len(key) > 40 {
		return key[:20] + "..." + key[len(key)-10:]
	}
	return key
}
