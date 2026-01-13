package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hades/podx/crypto"
	"github.com/hades/podx/keygen"
	"github.com/hades/podx/project"
)

// EncryptMethod represents the encryption method
type EncryptMethod int

const (
	MethodPassword EncryptMethod = iota
	MethodAgeKey
)

// DialogState represents the current state of the dialog
type DialogState int

const (
	StateSelectMethod DialogState = iota
	StatePasswordInput
	StateAgeKeyConfirm
	StateAddRecipient  // New state for adding recipient
	StateGenerateKey   // State for generating a new key pair
	StateProcessing
	StateComplete
	StateError
)

// EncryptDialogModel represents the encryption dialog
type EncryptDialogModel struct {
	visible      bool
	state        DialogState
	method       EncryptMethod
	selectedIdx  int
	files        []FileInfo
	password     textinput.Model
	confirmPass  textinput.Model
	focusedInput int // 0 = password, 1 = confirm
	outputPath   string
	project      *project.Project
	errorMsg     string
	successMsg   string
	width        int
	height       int
	isDecrypt    bool // true for decryption mode

	// Add recipient fields
	recipientName textinput.Model
	recipientKey  textinput.Model

	// Generated key fields
	generatedPublicKey  string
	generatedPrivateKey string
	generatingKey       bool
}

// encryptCompleteMsg is sent when encryption is complete
type encryptCompleteMsg struct {
	success bool
	message string
	files   int
}

// keygenCompleteMsg is sent when key generation completes
type keygenCompleteMsg struct {
	success    bool
	publicKey  string
	privateKey string
	keyFile    string
	errorMsg   string
}

// NewEncryptDialogModel creates a new encrypt dialog model
func NewEncryptDialogModel() EncryptDialogModel {
	password := textinput.New()
	password.Placeholder = "Enter password..."
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.CharLimit = 128

	confirmPass := textinput.New()
	confirmPass.Placeholder = "Confirm password..."
	confirmPass.EchoMode = textinput.EchoPassword
	confirmPass.EchoCharacter = '•'
	confirmPass.CharLimit = 128

	recipientName := textinput.New()
	recipientName.Placeholder = "e.g. John Doe"
	recipientName.CharLimit = 64

	recipientKey := textinput.New()
	recipientKey.Placeholder = "age1xxxxxxxxxx..."
	recipientKey.CharLimit = 256

	return EncryptDialogModel{
		visible:       false,
		state:         StateSelectMethod,
		method:        MethodPassword,
		password:      password,
		confirmPass:   confirmPass,
		recipientName: recipientName,
		recipientKey:  recipientKey,
	}
}

// Show displays the dialog for encryption
func (m *EncryptDialogModel) Show(files []FileInfo, proj *project.Project, width, height int) tea.Cmd {
	m.visible = true
	m.state = StateSelectMethod
	m.method = MethodPassword
	m.selectedIdx = 0
	m.files = files
	m.project = proj
	m.errorMsg = ""
	m.successMsg = ""
	m.width = width
	m.height = height
	m.isDecrypt = false
	m.password.SetValue("")
	m.confirmPass.SetValue("")
	m.focusedInput = 0

	// Generate output path for display
	if len(files) == 1 {
		m.outputPath = files[0].Path + ".podx"
	} else {
		m.outputPath = fmt.Sprintf("%d files → *.podx", len(files))
	}

	return nil
}

// ShowDecrypt displays the dialog for decryption
func (m *EncryptDialogModel) ShowDecrypt(files []FileInfo, proj *project.Project, width, height int) tea.Cmd {
	m.visible = true
	m.state = StateSelectMethod
	m.method = MethodPassword
	m.selectedIdx = 0
	m.files = files
	m.project = proj
	m.errorMsg = ""
	m.successMsg = ""
	m.width = width
	m.height = height
	m.isDecrypt = true
	m.password.SetValue("")
	m.confirmPass.SetValue("")
	m.focusedInput = 0

	// Generate output path for display
	if len(files) == 1 {
		m.outputPath = strings.TrimSuffix(files[0].Path, ".podx")
	} else {
		m.outputPath = fmt.Sprintf("%d files → original names", len(files))
	}

	return nil
}

// Hide hides the dialog
func (m *EncryptDialogModel) Hide() {
	m.visible = false
	m.password.SetValue("")
	m.confirmPass.SetValue("")
	m.errorMsg = ""
}

// IsVisible returns whether the dialog is visible
func (m EncryptDialogModel) IsVisible() bool {
	return m.visible
}

// Update handles messages for the dialog
func (m EncryptDialogModel) Update(msg tea.Msg) (EncryptDialogModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case encryptCompleteMsg:
		if msg.success {
			m.state = StateComplete
			m.successMsg = msg.message
		} else {
			m.state = StateError
			m.errorMsg = msg.message
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case StateSelectMethod:
			return m.updateMethodSelection(msg)
		case StatePasswordInput:
			return m.updatePasswordInput(msg)
		case StateAgeKeyConfirm:
			return m.updateAgeKeyConfirm(msg)
		case StateAddRecipient:
			return m.updateAddRecipient(msg)
		case StateComplete, StateError:
			// Any key closes the dialog
			if msg.String() == "enter" || msg.String() == "esc" || msg.String() == " " {
				m.Hide()
			}
			return m, nil
		}
	}

	return m, nil
}

// updateMethodSelection handles method selection state
func (m EncryptDialogModel) updateMethodSelection(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
	case "down", "j":
		if m.selectedIdx < 1 {
			m.selectedIdx++
		}
	case "enter", "l":
		if m.selectedIdx == 0 {
			m.method = MethodPassword
			m.state = StatePasswordInput
			m.password.Focus()
			return m, textinput.Blink
		} else {
			m.method = MethodAgeKey
			m.state = StateAgeKeyConfirm
		}
	case "esc", "q":
		m.Hide()
	case "1":
		m.method = MethodPassword
		m.state = StatePasswordInput
		m.password.Focus()
		return m, textinput.Blink
	case "2":
		m.method = MethodAgeKey
		m.state = StateAgeKeyConfirm
	}
	return m, nil
}

// updatePasswordInput handles password input state
func (m EncryptDialogModel) updatePasswordInput(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateSelectMethod
		m.password.Blur()
		m.confirmPass.Blur()
		m.password.SetValue("")
		m.confirmPass.SetValue("")
		return m, nil

	case "tab", "down":
		if !m.isDecrypt { // Only toggle focus for encryption (has confirm field)
			if m.focusedInput == 0 {
				m.focusedInput = 1
				m.password.Blur()
				m.confirmPass.Focus()
			} else {
				m.focusedInput = 0
				m.confirmPass.Blur()
				m.password.Focus()
			}
			return m, textinput.Blink
		}

	case "shift+tab", "up":
		if !m.isDecrypt {
			if m.focusedInput == 1 {
				m.focusedInput = 0
				m.confirmPass.Blur()
				m.password.Focus()
				return m, textinput.Blink
			}
		}

	case "enter":
		password := m.password.Value()

		if m.isDecrypt {
			// Decrypt mode - just need password
			if password == "" {
				m.errorMsg = "Password required"
				return m, nil
			}
			m.state = StateProcessing
			return m, m.doDecryptPassword()
		} else {
			// Encrypt mode - need password confirmation
			confirm := m.confirmPass.Value()
			if password == "" {
				m.errorMsg = "Password required"
				return m, nil
			}
			if password != confirm {
				m.errorMsg = "Passwords do not match"
				return m, nil
			}
			if len(password) < 8 {
				m.errorMsg = "Password must be at least 8 characters"
				return m, nil
			}
			m.state = StateProcessing
			return m, m.doEncryptPassword()
		}

	default:
		var cmd tea.Cmd
		if m.focusedInput == 0 || m.isDecrypt {
			m.password, cmd = m.password.Update(msg)
		} else {
			m.confirmPass, cmd = m.confirmPass.Update(msg)
		}
		m.errorMsg = "" // Clear error on typing
		return m, cmd
	}

	return m, nil
}

// updateAgeKeyConfirm handles age key confirmation state
func (m EncryptDialogModel) updateAgeKeyConfirm(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateSelectMethod
	case "a", "A":
		// Add recipient shortcut
		m.state = StateAddRecipient
		m.recipientName.SetValue("")
		m.recipientKey.SetValue("")
		m.focusedInput = 0
		m.recipientName.Focus()
		m.errorMsg = ""
		return m, textinput.Blink
	case "enter":
		if m.project == nil {
			m.errorMsg = "No PODX project found. Run 'podx init' first"
			m.state = StateError
			return m, nil
		}
		if len(m.project.Config.Recipients) == 0 {
			// No recipients - offer to add one
			m.state = StateAddRecipient
			m.recipientName.SetValue("")
			m.recipientKey.SetValue("")
			m.focusedInput = 0
			m.recipientName.Focus()
			m.errorMsg = ""
			return m, textinput.Blink
		}
		m.state = StateProcessing
		if m.isDecrypt {
			return m, m.doDecryptAge()
		}
		return m, m.doEncryptAge()
	}
	return m, nil
}

// updateAddRecipient handles the add recipient state
func (m EncryptDialogModel) updateAddRecipient(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateAgeKeyConfirm
		m.recipientName.Blur()
		m.recipientKey.Blur()
		m.errorMsg = ""
		return m, nil

	case "tab", "down":
		if m.focusedInput == 0 {
			m.focusedInput = 1
			m.recipientName.Blur()
			m.recipientKey.Focus()
		} else {
			m.focusedInput = 0
			m.recipientKey.Blur()
			m.recipientName.Focus()
		}
		return m, textinput.Blink

	case "shift+tab", "up":
		if m.focusedInput == 1 {
			m.focusedInput = 0
			m.recipientKey.Blur()
			m.recipientName.Focus()
			return m, textinput.Blink
		}

	case "enter":
		name := strings.TrimSpace(m.recipientName.Value())
		key := strings.TrimSpace(m.recipientKey.Value())

		// Validate
		if name == "" {
			m.errorMsg = "Name is required"
			return m, nil
		}
		if key == "" {
			m.errorMsg = "Age public key is required"
			return m, nil
		}
		if !strings.HasPrefix(key, "age1") {
			m.errorMsg = "Invalid Age key (must start with 'age1')"
			return m, nil
		}

		// Add recipient to project
		if m.project != nil {
			m.project.Config.Recipients = append(m.project.Config.Recipients, project.Recipient{
				Name: name,
				Key:  key,
			})
			// Save config
			if err := m.project.Save(); err != nil {
				m.errorMsg = "Failed to save: " + err.Error()
				return m, nil
			}
		}

		m.successMsg = fmt.Sprintf("Added recipient: %s", name)
		m.state = StateAgeKeyConfirm
		m.recipientName.Blur()
		m.recipientKey.Blur()
		m.errorMsg = ""
		return m, nil

	default:
		var cmd tea.Cmd
		if m.focusedInput == 0 {
			m.recipientName, cmd = m.recipientName.Update(msg)
		} else {
			m.recipientKey, cmd = m.recipientKey.Update(msg)
		}
		m.errorMsg = "" // Clear error on typing
		return m, cmd
	}

	return m, nil
}

// doEncryptPassword encrypts files with password
func (m EncryptDialogModel) doEncryptPassword() tea.Cmd {
	return func() tea.Msg {
		password := m.password.Value()

		// Derive key with new salt
		key, salt, err := crypto.DeriveKey([]byte(password), nil)
		if err != nil {
			return encryptCompleteMsg{success: false, message: err.Error()}
		}

		enc, err := crypto.NewEncryptor(crypto.AlgoAESGCM)
		if err != nil {
			return encryptCompleteMsg{success: false, message: err.Error()}
		}

		successCount := 0
		var lastErr error

		for _, file := range m.files {
			if file.IsDir || file.Name == ".." || file.IsEncrypted {
				continue
			}

			// Read file content
			content, err := os.ReadFile(file.Path)
			if err != nil {
				lastErr = err
				continue
			}

			// Encrypt
			ciphertext, err := enc.Encrypt(content, key)
			if err != nil {
				lastErr = err
				continue
			}

			// Format: [salt (16 bytes)][algo (1 byte = 0x01 for AES-GCM)][ciphertext]
			output := make([]byte, 0, 17+len(ciphertext))
			output = append(output, salt...)
			output = append(output, 0x01) // AES-GCM marker
			output = append(output, ciphertext...)

			// Write encrypted file
			outPath := file.Path + ".podx"
			if err := os.WriteFile(outPath, output, 0600); err != nil {
				lastErr = err
				continue
			}

			// Remove original
			if err := os.Remove(file.Path); err != nil {
				lastErr = err
			}

			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Encryption failed: %v", lastErr)}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("🔐 Encrypted %d file(s) with AES-GCM", successCount),
			files:   successCount,
		}
	}
}

// doDecryptPassword decrypts files with password
func (m EncryptDialogModel) doDecryptPassword() tea.Cmd {
	return func() tea.Msg {
		password := m.password.Value()

		successCount := 0
		var lastErr error

		for _, file := range m.files {
			if file.IsDir || file.Name == ".." || !file.IsEncrypted {
				continue
			}

			// Read encrypted file
			content, err := os.ReadFile(file.Path)
			if err != nil {
				lastErr = err
				continue
			}

			// Check minimum size (16 salt + 1 algo + some ciphertext)
			if len(content) < 18 {
				lastErr = fmt.Errorf("invalid encrypted file format")
				continue
			}

			// Extract salt and derive key
			salt := content[:16]
			algoMarker := content[16]
			ciphertext := content[17:]

			// Derive key with salt
			key, err := crypto.DeriveKeyWithSalt([]byte(password), salt)
			if err != nil {
				lastErr = err
				continue
			}

			// Get encryptor based on algo marker
			var enc crypto.Encryptor
			switch algoMarker {
			case 0x01:
				enc, _ = crypto.NewEncryptor(crypto.AlgoAESGCM)
			case 0x02:
				enc, _ = crypto.NewEncryptor(crypto.AlgoChaCha20)
			default:
				// Try Age decryption instead
				lastErr = fmt.Errorf("not a password-encrypted file (try Age key)")
				continue
			}

			// Decrypt
			plaintext, err := enc.Decrypt(ciphertext, key)
			if err != nil {
				lastErr = fmt.Errorf("wrong password or corrupted file")
				continue
			}

			// Write decrypted file
			outPath := strings.TrimSuffix(file.Path, ".podx")
			if err := os.WriteFile(outPath, plaintext, 0644); err != nil {
				lastErr = err
				continue
			}

			// Remove encrypted file
			if err := os.Remove(file.Path); err != nil {
				lastErr = err
			}

			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Decryption failed: %v", lastErr)}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("🔓 Decrypted %d file(s)", successCount),
			files:   successCount,
		}
	}
}

// doEncryptAge encrypts files with Age key
func (m EncryptDialogModel) doEncryptAge() tea.Cmd {
	return func() tea.Msg {
		if m.project == nil {
			return encryptCompleteMsg{success: false, message: "No PODX project found"}
		}

		var recipientKeys []string
		for _, r := range m.project.Config.Recipients {
			recipientKeys = append(recipientKeys, r.Key)
		}

		successCount := 0
		var lastErr error

		for _, file := range m.files {
			if file.IsDir || file.Name == ".." || file.IsEncrypted {
				continue
			}

			// Use project's encryption methods
			baseName := filepath.Base(file.Path)
			var err error
			if strings.HasPrefix(baseName, ".env") || strings.HasSuffix(baseName, ".env") {
				err = m.project.EncryptEnvFile(file.Path, recipientKeys)
			} else {
				err = m.project.EncryptRegularFile(file.Path, recipientKeys)
			}

			if err != nil {
				lastErr = err
				continue
			}

			// Delete original
			if err := os.Remove(file.Path); err != nil {
				lastErr = err
			}

			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Encryption failed: %v", lastErr)}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("🔑 Encrypted %d file(s) with Age key", successCount),
			files:   successCount,
		}
	}
}

// doDecryptAge decrypts files with Age key
func (m EncryptDialogModel) doDecryptAge() tea.Cmd {
	return func() tea.Msg {
		if m.project == nil {
			return encryptCompleteMsg{success: false, message: "No PODX project found"}
		}

		successCount := 0
		var lastErr error

		for _, file := range m.files {
			if file.IsDir || file.Name == ".." || !file.IsEncrypted {
				continue
			}

			decPath := strings.TrimSuffix(file.Path, ".podx")
			err := m.project.DecryptFile(file.Path, decPath)

			if err != nil {
				lastErr = err
				continue
			}

			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Decryption failed: %v", lastErr)}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("🔑 Decrypted %d file(s) with Age key", successCount),
			files:   successCount,
		}
	}
}

// doGenerateKey generates a new Age key pair
func (m EncryptDialogModel) doGenerateKey() tea.Cmd {
	return func() tea.Msg {
		result, err := keygen.GenerateAge()
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
func (m EncryptDialogModel) View() string {
	if !m.visible {
		return ""
	}

	var content string

	switch m.state {
	case StateSelectMethod:
		content = m.renderMethodSelection()
	case StatePasswordInput:
		content = m.renderPasswordInput()
	case StateAgeKeyConfirm:
		content = m.renderAgeKeyConfirm()
	case StateAddRecipient:
		content = m.renderAddRecipient()
	case StateProcessing:
		content = m.renderProcessing()
	case StateComplete:
		content = m.renderComplete()
	case StateError:
		content = m.renderError()
	}

	// Dialog box styling with solid background
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Background(ColorBg).
		Padding(1, 2).
		Width(60)

	return dialogStyle.Render(content)
}

// renderMethodSelection renders the method selection view
func (m EncryptDialogModel) renderMethodSelection() string {
	var s strings.Builder

	// Title
	action := "Encrypt"
	if m.isDecrypt {
		action = "Decrypt"
	}

	var title string
	if len(m.files) == 1 {
		title = fmt.Sprintf("%s: %s", action, m.files[0].Name)
	} else {
		title = fmt.Sprintf("%s %d files", action, len(m.files))
	}
	s.WriteString(TitleStyle.Render(title))
	s.WriteString("\n\n")

	s.WriteString("Choose encryption method:\n\n")

	// Password option
	passwordLabel := "  🔐 Password (AES-GCM)"
	if m.selectedIdx == 0 {
		passwordLabel = "> 🔐 Password (AES-GCM)"
		s.WriteString(SelectedStyle.Render(passwordLabel))
	} else {
		s.WriteString(passwordLabel)
	}
	s.WriteString("\n")
	if m.selectedIdx == 0 {
		s.WriteString(MutedStyle.Render("     Encrypt with a password you choose"))
	}
	s.WriteString("\n\n")

	// Age key option
	ageLabel := "  🔑 Age Key (public key)"
	if m.selectedIdx == 1 {
		ageLabel = "> 🔑 Age Key (public key)"
		s.WriteString(SelectedStyle.Render(ageLabel))
	} else {
		s.WriteString(ageLabel)
	}
	s.WriteString("\n")
	if m.selectedIdx == 1 {
		s.WriteString(MutedStyle.Render("     Encrypt with configured recipients"))
	}
	s.WriteString("\n\n")

	// Output path
	s.WriteString(MutedStyle.Render("Output: "))
	s.WriteString(m.outputPath)
	s.WriteString("\n\n")

	// Footer
	s.WriteString(MutedStyle.Render("[↑↓/jk] Select  [Enter] Confirm  [Esc] Cancel"))

	return s.String()
}

// renderPasswordInput renders the password input view
func (m EncryptDialogModel) renderPasswordInput() string {
	var s strings.Builder

	action := "Encrypt"
	icon := "🔐"
	if m.isDecrypt {
		action = "Decrypt"
		icon = "🔓"
	}

	s.WriteString(TitleStyle.Render(fmt.Sprintf("%s Password Encryption", icon)))
	s.WriteString("\n\n")

	// File info
	if len(m.files) == 1 {
		s.WriteString(MutedStyle.Render("File: "))
		s.WriteString(m.files[0].Name)
	} else {
		s.WriteString(MutedStyle.Render("Files: "))
		s.WriteString(fmt.Sprintf("%d selected", len(m.files)))
	}
	s.WriteString("\n\n")

	// Password input
	passwordLabel := "Password:  "
	if m.focusedInput == 0 && !m.isDecrypt {
		passwordLabel = "Password:  "
	}
	s.WriteString(passwordLabel)
	s.WriteString(m.password.View())
	s.WriteString("\n")

	// Confirm password (only for encryption)
	if !m.isDecrypt {
		s.WriteString("\n")
		s.WriteString("Confirm:   ")
		s.WriteString(m.confirmPass.View())
		s.WriteString("\n")
	}

	// Error message
	if m.errorMsg != "" {
		s.WriteString("\n")
		s.WriteString(ErrorStyle.Render("⚠ " + m.errorMsg))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Output path
	s.WriteString(MutedStyle.Render("Output: "))
	s.WriteString(m.outputPath)
	s.WriteString("\n\n")

	// Footer
	footer := fmt.Sprintf("[Enter] %s  [Tab] Next field  [Esc] Back", action)
	if m.isDecrypt {
		footer = fmt.Sprintf("[Enter] %s  [Esc] Back", action)
	}
	s.WriteString(MutedStyle.Render(footer))

	return s.String()
}

// renderAgeKeyConfirm renders the Age key confirmation view
func (m EncryptDialogModel) renderAgeKeyConfirm() string {
	var s strings.Builder

	action := "Encrypt"
	icon := "🔑"
	if m.isDecrypt {
		action = "Decrypt"
		icon = "🔓"
	}

	s.WriteString(TitleStyle.Render(fmt.Sprintf("%s Age Key Encryption", icon)))
	s.WriteString("\n\n")

	// File info
	if len(m.files) == 1 {
		s.WriteString(MutedStyle.Render("File: "))
		s.WriteString(m.files[0].Name)
	} else {
		s.WriteString(MutedStyle.Render("Files: "))
		s.WriteString(fmt.Sprintf("%d selected", len(m.files)))
	}
	s.WriteString("\n\n")

	// Recipients
	if m.project != nil && len(m.project.Config.Recipients) > 0 {
		s.WriteString("Recipients:\n")
		for _, r := range m.project.Config.Recipients {
			keyPreview := r.Key
			if len(keyPreview) > 20 {
				keyPreview = keyPreview[:20] + "..."
			}
			s.WriteString(SuccessStyle.Render("  ✓ "))
			s.WriteString(fmt.Sprintf("%s (%s)\n", r.Name, keyPreview))
		}
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Press [A] to add more recipients"))
	} else {
		s.WriteString(WarningStyle.Render("⚠ No recipients configured"))
		s.WriteString("\n\n")
		s.WriteString("  Press ")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[A]"))
		s.WriteString(" to add a recipient now\n")
		s.WriteString(MutedStyle.Render("  or run 'podx add-recipient' from terminal"))
	}
	s.WriteString("\n")

	// Output path
	s.WriteString(MutedStyle.Render("Output: "))
	s.WriteString(m.outputPath)
	s.WriteString("\n\n")

	// Footer
	s.WriteString(MutedStyle.Render(fmt.Sprintf("[Enter] %s  [A] Add recipient  [Esc] Back", action)))

	return s.String()
}

// renderAddRecipient renders the add recipient form
func (m EncryptDialogModel) renderAddRecipient() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("➕ Add Recipient"))
	s.WriteString("\n\n")

	s.WriteString("Add a team member who can decrypt your files.\n\n")

	// Name input
	nameLabel := "Name:     "
	if m.focusedInput == 0 {
		nameLabel = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("Name:     ")
	}
	s.WriteString(nameLabel)
	s.WriteString(m.recipientName.View())
	s.WriteString("\n\n")

	// Key input
	keyLabel := "Age Key:  "
	if m.focusedInput == 1 {
		keyLabel = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("Age Key:  ")
	}
	s.WriteString(keyLabel)
	s.WriteString(m.recipientKey.View())
	s.WriteString("\n")

	// Success message
	if m.successMsg != "" {
		s.WriteString("\n")
		s.WriteString(SuccessStyle.Render("✓ " + m.successMsg))
		s.WriteString("\n")
	}

	// Error message
	if m.errorMsg != "" {
		s.WriteString("\n")
		s.WriteString(ErrorStyle.Render("⚠ " + m.errorMsg))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Help text
	s.WriteString(MutedStyle.Render("Age public keys start with 'age1...'"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("Generate a key pair with: podx keygen -t age"))
	s.WriteString("\n\n")

	// Footer
	s.WriteString(MutedStyle.Render("[Enter] Add  [Tab] Next field  [Esc] Cancel"))

	return s.String()
}

// renderProcessing renders the processing view
func (m EncryptDialogModel) renderProcessing() string {
	action := "Encrypting"
	if m.isDecrypt {
		action = "Decrypting"
	}
	return TitleStyle.Render(fmt.Sprintf("⏳ %s...", action))
}

// renderComplete renders the completion view
func (m EncryptDialogModel) renderComplete() string {
	var s strings.Builder
	s.WriteString(SuccessStyle.Render("✅ " + m.successMsg))
	s.WriteString("\n\n")
	s.WriteString(MutedStyle.Render("[Enter] Close"))
	return s.String()
}

// renderError renders the error view
func (m EncryptDialogModel) renderError() string {
	var s strings.Builder
	s.WriteString(ErrorStyle.Render("❌ " + m.errorMsg))
	s.WriteString("\n\n")
	s.WriteString(MutedStyle.Render("[Enter] Close"))
	return s.String()
}
