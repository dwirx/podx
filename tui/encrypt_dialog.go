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

// EncryptModeOption represents the encryption mode (normal/paranoid)
type EncryptModeOption int

const (
	EncryptModeNormal EncryptModeOption = iota
	EncryptModeParanoid
)

// DialogState represents the current state of the dialog
type DialogState int

const (
	StateSelectMethod DialogState = iota
	StatePasswordInput
	StateAgeKeyConfirm
	StateSelectRecipients // New state for selecting recipients
	StateAddRecipient     // State for adding recipient
	StateGenerateKey      // State for generating a new key pair
	StateProcessing
	StateComplete
	StateError
)

// EncryptDialogModel represents the encryption dialog
type EncryptDialogModel struct {
	visible      bool
	state        DialogState
	method       EncryptMethod
	encryptMode  EncryptModeOption // normal or paranoid
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

	// Recipient selection fields
	selectedRecipients []bool // track which recipients are selected
	recipientCursor    int    // cursor position in recipient list
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

	case keygenCompleteMsg:
		m.generatingKey = false
		if msg.success {
			m.generatedPublicKey = msg.publicKey
			m.generatedPrivateKey = msg.privateKey
			// Auto-fill the recipient key field
			m.recipientKey.SetValue(msg.publicKey)
			m.successMsg = "Key generated successfully!"
			m.state = StateAddRecipient
			m.focusedInput = 0 // Focus on name field
			m.recipientName.Focus()
			return m, textinput.Blink
		} else {
			m.errorMsg = msg.errorMsg
			m.state = StateAddRecipient
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
		case StateSelectRecipients:
			return m.updateSelectRecipients(msg)
		case StateAddRecipient:
			return m.updateAddRecipient(msg)
		case StateGenerateKey:
			return m.updateGenerateKey(msg)
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
	case "n", "N":
		// Switch to normal mode
		m.encryptMode = EncryptModeNormal
	case "p", "P":
		// Switch to paranoid mode
		m.encryptMode = EncryptModeParanoid
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
	case "s", "S":
		// Select recipients - go to recipient selection state
		if m.project != nil && len(m.project.Config.Recipients) > 0 {
			m.state = StateSelectRecipients
			m.recipientCursor = 0
			// Initialize all recipients as selected
			m.selectedRecipients = make([]bool, len(m.project.Config.Recipients))
			for i := range m.selectedRecipients {
				m.selectedRecipients[i] = true
			}
			m.errorMsg = ""
			m.successMsg = ""
			return m, nil
		}
		m.errorMsg = "No recipients available. Add one first."
		return m, nil
	case "a", "A":
		// Add recipient shortcut
		m.state = StateAddRecipient
		m.recipientName.SetValue("")
		m.recipientKey.SetValue("")
		m.focusedInput = 0
		m.recipientName.Focus()
		m.errorMsg = ""
		return m, textinput.Blink
	case "m", "M":
		// Load and use my own key as recipient
		pubKey, err := keygen.LoadAgeRecipient()
		if err != nil {
			m.errorMsg = "No key found. Press [A] then [G] to generate one."
			return m, nil
		}
		// Check project exists
		if m.project == nil {
			m.errorMsg = "No PODX project found. Run 'podx init' first"
			return m, nil
		}
		// Check if already added
		for _, r := range m.project.Config.Recipients {
			if r.Key == pubKey {
				m.successMsg = "Your key is already added as recipient"
				return m, nil
			}
		}
		m.project.Config.Recipients = append(m.project.Config.Recipients, project.Recipient{
			Name: "Me (local)",
			Key:  pubKey,
		})
		if err := m.project.Save(); err != nil {
			m.errorMsg = "Failed to save: " + err.Error()
			return m, nil
		}
		m.successMsg = "Added your local key as recipient"
		return m, nil
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

// updateSelectRecipients handles recipient selection state
func (m EncryptDialogModel) updateSelectRecipients(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	if m.project == nil || len(m.project.Config.Recipients) == 0 {
		m.state = StateAgeKeyConfirm
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.state = StateAgeKeyConfirm
		m.errorMsg = ""
		return m, nil

	case "up", "k":
		if m.recipientCursor > 0 {
			m.recipientCursor--
		}
		return m, nil

	case "down", "j":
		if m.recipientCursor < len(m.project.Config.Recipients)-1 {
			m.recipientCursor++
		}
		return m, nil

	case " ", "space":
		// Toggle selection
		if m.recipientCursor < len(m.selectedRecipients) {
			m.selectedRecipients[m.recipientCursor] = !m.selectedRecipients[m.recipientCursor]
		}
		return m, nil

	case "a":
		// Select all
		for i := range m.selectedRecipients {
			m.selectedRecipients[i] = true
		}
		m.successMsg = "All recipients selected"
		return m, nil

	case "n":
		// Deselect all
		for i := range m.selectedRecipients {
			m.selectedRecipients[i] = false
		}
		m.successMsg = "All recipients deselected"
		return m, nil

	case "enter":
		// Check if at least one recipient is selected
		hasSelected := false
		for _, selected := range m.selectedRecipients {
			if selected {
				hasSelected = true
				break
			}
		}
		if !hasSelected {
			m.errorMsg = "Select at least one recipient"
			return m, nil
		}
		m.state = StateProcessing
		if m.isDecrypt {
			return m, m.doDecryptAge()
		}
		return m, m.doEncryptAgeWithSelectedRecipients()
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

	case "ctrl+g":
		// Switch to key generation state
		m.state = StateGenerateKey
		m.generatedPublicKey = ""
		m.generatedPrivateKey = ""
		m.generatingKey = false
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

// updateGenerateKey handles the key generation state
func (m EncryptDialogModel) updateGenerateKey(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.generatingKey {
			// Can't cancel during generation
			return m, nil
		}
		m.state = StateAddRecipient
		m.generatedPublicKey = ""
		m.generatedPrivateKey = ""
		return m, nil

	case "enter", "y", "Y":
		if m.generatedPublicKey != "" {
			// Key already generated, use it
			m.recipientKey.SetValue(m.generatedPublicKey)
			m.state = StateAddRecipient
			m.focusedInput = 0
			m.recipientName.Focus()
			m.recipientName.SetValue("Me (local)")
			return m, textinput.Blink
		}
		if !m.generatingKey {
			m.generatingKey = true
			m.errorMsg = ""
			return m, m.doGenerateKey()
		}

	case "n", "N":
		m.state = StateAddRecipient
		return m, nil
	}

	return m, nil
}

// doEncryptPassword encrypts files with password
func (m EncryptDialogModel) doEncryptPassword() tea.Cmd {
	return func() tea.Msg {
		password := m.password.Value()

		// Determine encryption mode and cipher
		mode := crypto.ModeNormal
		cipher := crypto.CipherAESGCM
		if m.encryptMode == EncryptModeParanoid {
			mode = crypto.ModeParanoid
			cipher = crypto.CipherCascade
		}

		opts := &crypto.EncryptOptions{
			Mode:   mode,
			Cipher: cipher,
		}

		successCount := 0
		var lastErr error
		var processedFiles []string

		for _, file := range m.files {
			if file.IsDir || file.Name == ".." || file.IsEncrypted {
				continue
			}

			// Read file content
			content, err := os.ReadFile(file.Path)
			if err != nil {
				lastErr = fmt.Errorf("failed to read %s: %v", file.Name, err)
				continue
			}

			// Encrypt using v2 format
			ciphertext, err := crypto.EncryptV2(content, []byte(password), opts)
			if err != nil {
				lastErr = fmt.Errorf("failed to encrypt %s: %v", file.Name, err)
				continue
			}

			// Write encrypted file
			outPath := file.Path + ".podx"
			if err := os.WriteFile(outPath, ciphertext, 0600); err != nil {
				lastErr = fmt.Errorf("failed to write %s: %v", file.Name+".podx", err)
				continue
			}

			// Only remove original after successful encryption
			if err := os.Remove(file.Path); err != nil {
				// Non-fatal: encryption succeeded but cleanup failed
				lastErr = fmt.Errorf("encrypted but failed to remove original: %v", err)
			}

			processedFiles = append(processedFiles, file.Name+" → "+file.Name+".podx")
			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Encryption failed: %v", lastErr)}
		}

		modeStr := "AES-256-GCM"
		if m.encryptMode == EncryptModeParanoid {
			modeStr = "XChaCha20+Serpent (paranoid)"
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("Encrypted %d file(s) with %s", successCount, modeStr),
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
		var processedFiles []string

		for _, file := range m.files {
			if file.IsDir || file.Name == ".." || !file.IsEncrypted {
				continue
			}

			// Read encrypted file
			content, err := os.ReadFile(file.Path)
			if err != nil {
				lastErr = fmt.Errorf("failed to read %s: %v", file.Name, err)
				continue
			}

			// Use auto-detection to decrypt (supports v1 and v2 formats)
			plaintext, err := crypto.DetectAndDecrypt(content, []byte(password))
			if err != nil {
				lastErr = fmt.Errorf("wrong password or corrupted file: %s", file.Name)
				continue
			}

			// Write decrypted file (remove .podx extension)
			outPath := strings.TrimSuffix(file.Path, ".podx")
			outName := strings.TrimSuffix(file.Name, ".podx")
			if err := os.WriteFile(outPath, plaintext, 0644); err != nil {
				lastErr = fmt.Errorf("failed to write %s: %v", outName, err)
				continue
			}

			// Remove encrypted file after successful decryption
			if err := os.Remove(file.Path); err != nil {
				// Non-fatal: decryption succeeded but cleanup failed
				lastErr = fmt.Errorf("decrypted but failed to remove encrypted file: %v", err)
			}

			processedFiles = append(processedFiles, file.Name+" → "+outName)
			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Decryption failed: %v", lastErr)}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("Decrypted %d file(s) successfully", successCount),
			files:   successCount,
		}
	}
}

// doEncryptAge encrypts files with Age key
func (m EncryptDialogModel) doEncryptAge() tea.Cmd {
	return func() tea.Msg {
		if m.project == nil {
			return encryptCompleteMsg{success: false, message: "No PODX project found. Run 'podx init' first"}
		}

		if len(m.project.Config.Recipients) == 0 {
			return encryptCompleteMsg{success: false, message: "No recipients configured. Add a recipient first"}
		}

		var recipientKeys []string
		for _, r := range m.project.Config.Recipients {
			recipientKeys = append(recipientKeys, r.Key)
		}

		successCount := 0
		var lastErr error
		var processedFiles []string

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
				lastErr = fmt.Errorf("failed to encrypt %s: %v", file.Name, err)
				continue
			}

			// Only delete original after successful encryption
			if err := os.Remove(file.Path); err != nil {
				// Non-fatal: encryption succeeded but cleanup failed
				lastErr = fmt.Errorf("encrypted but failed to remove original %s: %v", file.Name, err)
			}

			processedFiles = append(processedFiles, file.Name+" → "+file.Name+".podx")
			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Encryption failed: %v", lastErr)}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("Encrypted %d file(s) with Age key", successCount),
			files:   successCount,
		}
	}
}

// doDecryptAge decrypts files with Age key
func (m EncryptDialogModel) doDecryptAge() tea.Cmd {
	return func() tea.Msg {
		if m.project == nil {
			return encryptCompleteMsg{success: false, message: "No PODX project found. Run 'podx init' first"}
		}

		successCount := 0
		var lastErr error
		var processedFiles []string

		for _, file := range m.files {
			if file.IsDir || file.Name == ".." || !file.IsEncrypted {
				continue
			}

			decPath := strings.TrimSuffix(file.Path, ".podx")
			decName := strings.TrimSuffix(file.Name, ".podx")
			err := m.project.DecryptFile(file.Path, decPath)

			if err != nil {
				lastErr = fmt.Errorf("failed to decrypt %s: %v", file.Name, err)
				continue
			}

			// Only delete the encrypted .podx file after successful decryption
			if err := os.Remove(file.Path); err != nil {
				// Non-fatal: decryption succeeded but cleanup failed
				lastErr = fmt.Errorf("decrypted but failed to remove %s: %v", file.Name, err)
			}

			processedFiles = append(processedFiles, file.Name+" → "+decName)
			successCount++
		}

		if successCount == 0 {
			if lastErr != nil {
				return encryptCompleteMsg{success: false, message: fmt.Sprintf("Decryption failed: %v", lastErr)}
			}
			return encryptCompleteMsg{success: false, message: "No encrypted files to decrypt"}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("Decrypted %d file(s) with Age key", successCount),
			files:   successCount,
		}
	}
}

// doEncryptAgeWithSelectedRecipients encrypts files with selected Age recipients
func (m EncryptDialogModel) doEncryptAgeWithSelectedRecipients() tea.Cmd {
	return func() tea.Msg {
		if m.project == nil {
			return encryptCompleteMsg{success: false, message: "No PODX project found. Run 'podx init' first"}
		}

		// Build list of selected recipient keys
		var recipientKeys []string
		var selectedNames []string
		for i, r := range m.project.Config.Recipients {
			if i < len(m.selectedRecipients) && m.selectedRecipients[i] {
				recipientKeys = append(recipientKeys, r.Key)
				selectedNames = append(selectedNames, r.Name)
			}
		}

		if len(recipientKeys) == 0 {
			return encryptCompleteMsg{success: false, message: "No recipients selected"}
		}

		successCount := 0
		var lastErr error
		var processedFiles []string

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
				lastErr = fmt.Errorf("failed to encrypt %s: %v", file.Name, err)
				continue
			}

			// Only delete original after successful encryption
			if err := os.Remove(file.Path); err != nil {
				lastErr = fmt.Errorf("encrypted but failed to remove original %s: %v", file.Name, err)
			}

			processedFiles = append(processedFiles, file.Name+" → "+file.Name+".podx")
			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return encryptCompleteMsg{success: false, message: fmt.Sprintf("Encryption failed: %v", lastErr)}
		}

		return encryptCompleteMsg{
			success: true,
			message: fmt.Sprintf("Encrypted %d file(s) for %d recipient(s)", successCount, len(recipientKeys)),
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
	case StateSelectRecipients:
		content = m.renderSelectRecipients()
	case StateAddRecipient:
		content = m.renderAddRecipient()
	case StateGenerateKey:
		content = m.renderGenerateKey()
	case StateProcessing:
		content = m.renderProcessing()
	case StateComplete:
		content = m.renderComplete()
	case StateError:
		content = m.renderError()
	}

	// Dialog box styling with responsive width
	dialogWidth := 60
	if m.width > 0 {
		// Calculate responsive dialog width
		dialogWidth = ResponsiveWidth(m.width, 70, 50, 80)
	}

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Background(ColorBg).
		Padding(1, 2).
		Width(dialogWidth)

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

	// Encryption Mode selection (only for encryption)
	if !m.isDecrypt {
		s.WriteString("Security Mode:\n")

		// Normal mode option
		normalLabel := "  🔰 Normal"
		if m.encryptMode == EncryptModeNormal {
			normalLabel = "  ✓ 🔰 Normal"
			s.WriteString(SuccessStyle.Render(normalLabel))
		} else {
			s.WriteString(MutedStyle.Render(normalLabel))
		}
		s.WriteString("  ")

		// Paranoid mode option
		paranoidLabel := "🛡️  Paranoid"
		if m.encryptMode == EncryptModeParanoid {
			paranoidLabel = "✓ 🛡️  Paranoid"
			s.WriteString(SuccessStyle.Render(paranoidLabel))
		} else {
			s.WriteString(MutedStyle.Render(paranoidLabel))
		}
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("     Press [N] Normal  [P] Paranoid"))
		s.WriteString("\n\n")
	}

	s.WriteString("Choose encryption method:\n\n")

	// Password option
	passwordLabel := "  🔐 Password (AES-GCM)"
	if m.encryptMode == EncryptModeParanoid && !m.isDecrypt {
		passwordLabel = "  🔐 Password (XChaCha20+Serpent)"
	}
	if m.selectedIdx == 0 {
		passwordLabel = "> " + passwordLabel[2:]
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
	actionVerb := "Encrypting"
	if m.isDecrypt {
		action = "Decrypt"
		icon = "🔓"
		actionVerb = "Decrypting"
	}

	s.WriteString(TitleStyle.Render(fmt.Sprintf("%s Age Key %s", icon, action)))
	s.WriteString("\n\n")

	// File info with responsive display
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Files:"))
	s.WriteString("\n")

	// Show fewer files on smaller terminals
	maxFilesToShow := 3
	if m.width > 0 && m.width < SmallWidth {
		maxFilesToShow = 2
	}

	for i, file := range m.files {
		if i >= maxFilesToShow {
			s.WriteString(MutedStyle.Render(fmt.Sprintf("  ... and %d more\n", len(m.files)-maxFilesToShow)))
			break
		}
		fileIcon := "📄"
		if file.IsEncrypted {
			fileIcon = "🔒"
		}
		// Truncate filename if needed
		fileName := file.Name
		maxNameLen := 40
		if m.width > 0 && m.width < SmallWidth {
			maxNameLen = 25
		}
		fileName = TruncateText(fileName, maxNameLen)
		s.WriteString(fmt.Sprintf("  %s %s\n", fileIcon, fileName))
	}
	s.WriteString("\n")

	// Recipients
	if m.project != nil && len(m.project.Config.Recipients) > 0 {
		s.WriteString(lipgloss.NewStyle().Bold(true).Render("Recipients:"))
		s.WriteString("\n")
		for _, r := range m.project.Config.Recipients {
			keyPreview := r.Key
			if len(keyPreview) > 20 {
				keyPreview = keyPreview[:20] + "..."
			}
			s.WriteString(SuccessStyle.Render("  ✓ "))
			s.WriteString(fmt.Sprintf("%s (%s)\n", r.Name, keyPreview))
		}
		s.WriteString("\n")
	} else {
		s.WriteString(WarningStyle.Render("⚠ No recipients configured"))
		s.WriteString("\n\n")
		s.WriteString("  Press ")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[A]"))
		s.WriteString(" to add a recipient now\n")
		s.WriteString("  Press ")
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[M]"))
		s.WriteString(" to use your own local key\n\n")
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

	// Output info
	s.WriteString(MutedStyle.Render(fmt.Sprintf("%s → %s", actionVerb, m.outputPath)))
	s.WriteString("\n\n")

	// Footer with clear action
	if m.project != nil && len(m.project.Config.Recipients) > 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(fmt.Sprintf("[Enter] %s Now", action)))
		s.WriteString(MutedStyle.Render("  [S] Select  [A] Add  [M] My key  [Esc] Back"))
	} else {
		s.WriteString(MutedStyle.Render("[M] My key  [A] Add recipient  [Esc] Back"))
	}

	return s.String()
}

// renderSelectRecipients renders the recipient selection view
func (m EncryptDialogModel) renderSelectRecipients() string {
	var s strings.Builder

	// Title
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("🔑 Select Recipients"))
	s.WriteString("\n\n")

	// Description
	s.WriteString(MutedStyle.Render("Choose who can decrypt these files:"))
	s.WriteString("\n\n")

	// Count selected
	selectedCount := 0
	for _, sel := range m.selectedRecipients {
		if sel {
			selectedCount++
		}
	}

	// Status bar
	statusStyle := lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)
	s.WriteString(statusStyle.Render(fmt.Sprintf("Selected: %d/%d recipients", selectedCount, len(m.project.Config.Recipients))))
	s.WriteString("\n\n")

	// Recipient list
	if m.project != nil {
		for i, r := range m.project.Config.Recipients {
			// Cursor indicator
			cursor := "  "
			if i == m.recipientCursor {
				cursor = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("> ")
			}

			// Checkbox
			checkbox := "[ ]"
			if i < len(m.selectedRecipients) && m.selectedRecipients[i] {
				checkbox = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render("[✓]")
			}

			// Recipient info
			keyPreview := r.Key
			if len(keyPreview) > 20 {
				keyPreview = keyPreview[:20] + "..."
			}

			// Highlight current row
			nameStyle := lipgloss.NewStyle()
			if i == m.recipientCursor {
				nameStyle = nameStyle.Foreground(ColorPrimary).Bold(true)
			}

			s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, nameStyle.Render(r.Name)))
			s.WriteString(MutedStyle.Render(fmt.Sprintf("       %s\n", keyPreview)))
		}
	}
	s.WriteString("\n")

	// Suggestions box
	s.WriteString(CardTitleStyle.Render("Quick Actions:"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  [Space] Toggle selection"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  [a] Select all  [n] Deselect all"))
	s.WriteString("\n\n")

	// Error/success messages
	if m.errorMsg != "" {
		s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
		s.WriteString("\n\n")
	}
	if m.successMsg != "" {
		s.WriteString(SuccessStyle.Render("✓ " + m.successMsg))
		s.WriteString("\n\n")
	}

	// Footer
	if selectedCount > 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter] Encrypt"))
		s.WriteString(MutedStyle.Render(fmt.Sprintf(" for %d recipient(s)  ", selectedCount)))
	}
	s.WriteString(MutedStyle.Render("[Esc] Back"))

	return s.String()
}

// renderInputCursor returns a cursor indicator for input fields
func (m EncryptDialogModel) renderInputCursor(inputIndex int) string {
	if m.focusedInput == inputIndex {
		return lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(">")
	}
	return " "
}

// renderAddRecipient renders the add recipient form
func (m EncryptDialogModel) renderAddRecipient() string {
	var s strings.Builder

	// Title with primary color and bold
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("Add Recipient"))
	s.WriteString("\n\n")

	// Description
	s.WriteString(MutedStyle.Render("Add a team member who can decrypt your files."))
	s.WriteString("\n\n")

	// Name input with cursor indicator
	nameLabelStyle := lipgloss.NewStyle()
	if m.focusedInput == 0 {
		nameLabelStyle = nameLabelStyle.Foreground(ColorPrimary).Bold(true)
	} else {
		nameLabelStyle = nameLabelStyle.Foreground(ColorMuted)
	}
	s.WriteString(m.renderInputCursor(0))
	s.WriteString(" ")
	s.WriteString(nameLabelStyle.Render("Name:"))
	s.WriteString("     ")
	s.WriteString(m.recipientName.View())
	s.WriteString("\n\n")

	// Key input with cursor indicator
	keyLabelStyle := lipgloss.NewStyle()
	if m.focusedInput == 1 {
		keyLabelStyle = keyLabelStyle.Foreground(ColorPrimary).Bold(true)
	} else {
		keyLabelStyle = keyLabelStyle.Foreground(ColorMuted)
	}
	s.WriteString(m.renderInputCursor(1))
	s.WriteString(" ")
	s.WriteString(keyLabelStyle.Render("Age Key:"))
	s.WriteString("  ")
	s.WriteString(m.recipientKey.View())
	s.WriteString("\n")

	// Success message
	if m.successMsg != "" {
		s.WriteString("\n")
		s.WriteString(SuccessStyle.Render("  [OK] " + m.successMsg))
		s.WriteString("\n")
	}

	// Error message
	if m.errorMsg != "" {
		s.WriteString("\n")
		s.WriteString(ErrorStyle.Render("  [!] " + m.errorMsg))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Help text section
	s.WriteString(MutedStyle.Render("  Age public keys start with 'age1...'"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("  Generate a key with [Ctrl+G] or run: podx keygen -t age"))
	s.WriteString("\n\n")

	// Footer with action hints
	s.WriteString(MutedStyle.Render("[Enter] Add  [Ctrl+G] Generate key  [Tab] Next  [Esc] Cancel"))

	return s.String()
}

// renderGenerateKey renders the key generation view
func (m EncryptDialogModel) renderGenerateKey() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("Generate Age Key Pair"))
	s.WriteString("\n\n")

	if m.generatingKey {
		s.WriteString("Generating key pair...\n")
		s.WriteString(MutedStyle.Render("Please wait..."))
		return s.String()
	}

	if m.generatedPublicKey != "" {
		// Show generated key
		s.WriteString(SuccessStyle.Render("Key generated successfully!"))
		s.WriteString("\n\n")

		s.WriteString(CardTitleStyle.Render("Public Key:"))
		s.WriteString("\n")
		// Wrap long key
		pubKey := m.generatedPublicKey
		if len(pubKey) > 50 {
			s.WriteString("  " + pubKey[:50] + "\n")
			s.WriteString("  " + pubKey[50:])
		} else {
			s.WriteString("  " + pubKey)
		}
		s.WriteString("\n\n")

		s.WriteString(CardTitleStyle.Render("Private Key:"))
		s.WriteString("\n")
		privKey := m.generatedPrivateKey
		if len(privKey) > 50 {
			s.WriteString("  " + privKey[:50] + "\n")
			s.WriteString("  " + privKey[50:])
		} else {
			s.WriteString("  " + privKey)
		}
		s.WriteString("\n\n")

		s.WriteString(WarningStyle.Render("Keys saved to ~/.config/podx/"))
		s.WriteString("\n\n")

		s.WriteString(MutedStyle.Render("[Enter] Use this key  [Esc] Back"))
	} else {
		// Confirmation prompt
		s.WriteString("This will generate a new Age X25519 key pair.\n\n")

		s.WriteString("The keys will be saved to:\n")
		s.WriteString(MutedStyle.Render("  ~/.config/podx/age-keys.txt (private)\n"))
		s.WriteString(MutedStyle.Render("  ~/.config/podx/age-recipients/default.txt (public)\n"))
		s.WriteString("\n")

		if m.errorMsg != "" {
			s.WriteString(ErrorStyle.Render("Error: " + m.errorMsg))
			s.WriteString("\n\n")
		}

		s.WriteString("Generate new key pair?\n\n")
		s.WriteString(MutedStyle.Render("[Y/Enter] Generate  [N/Esc] Cancel"))
	}

	return s.String()
}

// renderProcessing renders the processing view
func (m EncryptDialogModel) renderProcessing() string {
	var s strings.Builder

	action := "Encrypting"
	icon := "🔐"
	if m.isDecrypt {
		action = "Decrypting"
		icon = "🔓"
	}

	// Animated header
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("╔══════════════════════════════════╗"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(fmt.Sprintf("║     %s %s...          ║", icon, action)))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("╚══════════════════════════════════╝"))
	s.WriteString("\n\n")

	// Spinner and message
	s.WriteString(RenderSpinner(0))
	s.WriteString(" Processing ")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true).Render(fmt.Sprintf("%d", len(m.files))))
	s.WriteString(" file(s)...")
	s.WriteString("\n\n")

	// Show encryption mode
	if !m.isDecrypt {
		modeInfo := "AES-256-GCM"
		if m.encryptMode == EncryptModeParanoid {
			modeInfo = "XChaCha20 + Serpent"
		}
		s.WriteString(MutedStyle.Render("Mode: "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render(modeInfo))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("Please wait..."))

	return s.String()
}

// renderComplete renders the completion view
func (m EncryptDialogModel) renderComplete() string {
	var s strings.Builder

	// Success header with nice box
	boxStyle := lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	if m.isDecrypt {
		s.WriteString(boxStyle.Render("╔════════════════════════════════════════╗"))
		s.WriteString("\n")
		s.WriteString(boxStyle.Render("║    🔓 DECRYPTION SUCCESSFUL            ║"))
		s.WriteString("\n")
		s.WriteString(boxStyle.Render("╚════════════════════════════════════════╝"))
	} else {
		s.WriteString(boxStyle.Render("╔════════════════════════════════════════╗"))
		s.WriteString("\n")
		s.WriteString(boxStyle.Render("║    🔐 ENCRYPTION SUCCESSFUL            ║"))
		s.WriteString("\n")
		s.WriteString(boxStyle.Render("╚════════════════════════════════════════╝"))
	}
	s.WriteString("\n\n")

	// Status badge with message
	s.WriteString(BadgeSuccessStyle.Render(" SUCCESS "))
	s.WriteString("  ")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render(m.successMsg))
	s.WriteString("\n\n")

	// Show encryption/decryption details
	if m.isDecrypt {
		s.WriteString(CardTitleStyle.Render("Details:"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Method: "))
		if m.method == MethodPassword {
			s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render("Password-based"))
		} else {
			s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render("Age Key"))
		}
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Status: "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("Original files restored"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Cleanup: "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("Encrypted .podx files removed"))
		s.WriteString("\n\n")
	} else {
		s.WriteString(CardTitleStyle.Render("Encryption Details:"))
		s.WriteString("\n")
		modeInfo := "AES-256-GCM"
		if m.encryptMode == EncryptModeParanoid {
			modeInfo = "XChaCha20-Poly1305 + Serpent"
		}
		s.WriteString(MutedStyle.Render("  Cipher:  "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render(modeInfo))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  KDF:     "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSecondary).Render("Argon2id"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Mode:    "))
		if m.encryptMode == EncryptModeParanoid {
			s.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("PARANOID"))
		} else {
			s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("Normal"))
		}
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Cleanup: "))
		s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("Original files removed"))
		s.WriteString("\n\n")
	}

	// Show files that were processed with transformation arrows
	if len(m.files) > 0 {
		s.WriteString(CardTitleStyle.Render("Files Processed:"))
		s.WriteString("\n")
		for i, file := range m.files {
			if i >= 5 {
				s.WriteString(MutedStyle.Render(fmt.Sprintf("  ... and %d more files", len(m.files)-5)))
				s.WriteString("\n")
				break
			}
			if m.isDecrypt {
				// .env.podx → .env
				origName := strings.TrimSuffix(file.Name, ".podx")
				s.WriteString(SuccessStyle.Render("  ✓ "))
				s.WriteString(MutedStyle.Render(file.Name))
				s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(" → "))
				s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(origName))
				s.WriteString("\n")
			} else {
				// .env → .env.podx
				s.WriteString(SuccessStyle.Render("  ✓ "))
				s.WriteString(MutedStyle.Render(file.Name))
				s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(" → "))
				s.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(file.Name+".podx"))
				s.WriteString("\n")
			}
		}
		s.WriteString("\n")
	}

	// Security reminder
	if !m.isDecrypt {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render("💡 Tip: "))
		s.WriteString(MutedStyle.Render("Commit the .podx files to git, not the originals"))
		s.WriteString("\n\n")
	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render("⚠️  Warning: "))
		s.WriteString(MutedStyle.Render("Do not commit decrypted files to git"))
		s.WriteString("\n\n")
	}

	// Footer with highlighted Enter
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter]"))
	s.WriteString(MutedStyle.Render(" Close"))
	return s.String()
}

// renderError renders the error view
func (m EncryptDialogModel) renderError() string {
	var s strings.Builder

	// Error header with box
	s.WriteString(lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("╔══════════════════════════════════╗"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("║        OPERATION FAILED          ║"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("╚══════════════════════════════════╝"))
	s.WriteString("\n\n")

	// Status badge
	s.WriteString(BadgeErrorStyle.Render(" ERROR "))
	s.WriteString("\n\n")

	// Error message in a highlighted box
	errBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorError).
		Padding(0, 1).
		Render(m.errorMsg)
	s.WriteString(errBox)
	s.WriteString("\n\n")

	// Suggestions based on error type
	if strings.Contains(m.errorMsg, "No Age identity") || strings.Contains(m.errorMsg, "no Age identity") {
		s.WriteString(CardTitleStyle.Render("Suggestion:"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Run 'podx keygen -t age' to generate a key"))
		s.WriteString("\n\n")
	} else if strings.Contains(m.errorMsg, "No PODX project") {
		s.WriteString(CardTitleStyle.Render("Suggestion:"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Run 'podx init' to initialize a project"))
		s.WriteString("\n\n")
	} else if strings.Contains(m.errorMsg, "No recipients") {
		s.WriteString(CardTitleStyle.Render("Suggestion:"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Press [A] to add a recipient"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Press [M] to use your own key"))
		s.WriteString("\n\n")
	} else if strings.Contains(m.errorMsg, "wrong password") || strings.Contains(m.errorMsg, "corrupted") {
		s.WriteString(CardTitleStyle.Render("Suggestion:"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Check that you're using the correct password"))
		s.WriteString("\n")
		s.WriteString(MutedStyle.Render("  Verify the file is not corrupted"))
		s.WriteString("\n\n")
	}

	// Footer
	s.WriteString(lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("[Enter]"))
	s.WriteString(MutedStyle.Render(" Close  "))
	s.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("[Esc]"))
	s.WriteString(MutedStyle.Render(" Back"))
	return s.String()
}
