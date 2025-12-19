package tui

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hades/podx/crypto"
	"github.com/hades/podx/keygen"
	"github.com/hades/podx/parser"
	"github.com/hades/podx/updater"
)

// AppState represents the current state of the TUI app
type AppState int

const (
	StateMenu AppState = iota
	StateFileBrowser
	StateAlgoSelect
	StateKeygenSelect
	StateEncryptPassword
	StateDecryptPassword
	StateAgeEncrypt
	StateAgeDecrypt
	StateEnvAlgoSelect
	StateEnvFileBrowser
	StateEnvPassword
	StateEnvDecryptBrowser
	StateEnvDecryptPassword
	StateKeygenForm
	StateProjectMenu
	StateProcessing
	StateResult
)

// App is the main TUI application
type App struct {
	state        AppState
	prevState    AppState
	menu         MenuModel
	fileBrowser  FileBrowserModel
	passwordForm PasswordFormModel
	form         FormModel
	result       ResultModel
	projectMenu  MenuModel
	spinner      spinner.Model
	width        int
	height       int
	version      string

	// Temp data for encryption flow
	inputFile    string
	outputFile   string
	isEncrypting bool
	useAgeKey    bool   // Use Age key instead of password
	selectedAlgo string // Selected algorithm
	algoMenu     MenuModel
	keygenMenu   MenuModel
	envAlgoMenu  MenuModel
	isEnvMode    bool // .env format-preserving mode
}

// PasswordFormModel is a simple password input
type PasswordFormModel struct {
	input      textinput.Model
	inputFile  string
	outputFile string
	algo       string
	done       bool
	cancelled  bool
	isEncrypt  bool
}

func NewPasswordForm(inputFile, outputFile, algo string, isEncrypt bool) PasswordFormModel {
	ti := textinput.New()
	ti.Placeholder = "Enter password..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Width = 40
	ti.Focus()

	return PasswordFormModel{
		input:      ti,
		inputFile:  inputFile,
		outputFile: outputFile,
		algo:       algo,
		isEncrypt:  isEncrypt,
	}
}

func (m PasswordFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PasswordFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, nil
		case "enter":
			if m.input.Value() != "" {
				m.done = true
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m PasswordFormModel) View() string {
	var b strings.Builder

	action := "🔐 Encrypt"
	if !m.isEncrypt {
		action = "🔓 Decrypt"
	}

	b.WriteString(TitleStyle.Render(action))
	b.WriteString("\n\n")

	// Show file info
	b.WriteString(LabelStyle.Render("📄 Input:  "))
	b.WriteString(InfoStyle.Render(m.inputFile))
	b.WriteString("\n")
	b.WriteString(LabelStyle.Render("💾 Output: "))
	b.WriteString(InfoStyle.Render(m.outputFile))
	b.WriteString("\n")
	if m.isEncrypt && m.algo != "" {
		b.WriteString(LabelStyle.Render("🔧 Algo:   "))
		b.WriteString(InfoStyle.Render(m.algo))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Password input
	b.WriteString(FocusedLabelStyle.Render("🔑 Password"))
	b.WriteString("\n")
	b.WriteString(FocusedInputStyle.Render(m.input.View()))
	b.WriteString("\n\n")

	// Buttons
	b.WriteString(ActiveButtonStyle.Render("[ Submit ]"))
	b.WriteString("  ")
	b.WriteString(ButtonStyle.Render("[ Cancel ]"))
	b.WriteString("\n\n")

	// Help
	help := []string{
		RenderHelp("Enter", "submit"),
		RenderHelp("Esc", "cancel"),
	}
	b.WriteString(strings.Join(help, HelpSeparator))

	return FocusedBoxStyle.Render(b.String())
}

func (m PasswordFormModel) Password() string { return m.input.Value() }
func (m PasswordFormModel) Done() bool       { return m.done }
func (m PasswordFormModel) Cancelled() bool  { return m.cancelled }

// NewApp creates a new TUI app
func NewApp(version string) App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SuccessStyle

	return App{
		state:   StateMenu,
		menu:    NewMenuModel(),
		version: version,
		spinner: s,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}

	switch a.state {
	case StateMenu:
		return a.updateMenu(msg)
	case StateFileBrowser, StateEnvFileBrowser, StateEnvDecryptBrowser:
		return a.updateFileBrowser(msg)
	case StateAlgoSelect:
		return a.updateAlgoSelect(msg)
	case StateEnvAlgoSelect:
		return a.updateEnvAlgoSelect(msg)
	case StateKeygenSelect:
		return a.updateKeygenSelect(msg)
	case StateEncryptPassword:
		return a.updatePasswordForm(msg, true)
	case StateDecryptPassword:
		return a.updatePasswordForm(msg, false)
	case StateEnvPassword:
		return a.updateEnvPasswordForm(msg, true)
	case StateEnvDecryptPassword:
		return a.updateEnvPasswordForm(msg, false)
	case StateKeygenForm:
		return a.updateGenericForm(msg)
	case StateProjectMenu:
		return a.updateProjectMenu(msg)
	case StateResult:
		return a.updateResult(msg)
	}

	return a, nil
}

func (a App) View() string {
	switch a.state {
	case StateMenu:
		return a.menu.View()
	case StateFileBrowser, StateEnvFileBrowser, StateEnvDecryptBrowser:
		return "\n" + a.fileBrowser.View()
	case StateAlgoSelect:
		return a.algoMenu.View()
	case StateEnvAlgoSelect:
		return a.envAlgoMenu.View()
	case StateKeygenSelect:
		return a.keygenMenu.View()
	case StateEncryptPassword, StateDecryptPassword, StateEnvPassword, StateEnvDecryptPassword:
		return "\n" + a.passwordForm.View()
	case StateKeygenForm:
		return a.form.View()
	case StateProjectMenu:
		return a.projectMenu.View()
	case StateProcessing:
		return fmt.Sprintf("\n\n  %s Processing...\n", a.spinner.View())
	case StateResult:
		return a.result.View()
	}
	return ""
}

// Menu handling
func (a App) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	menuModel, cmd := a.menu.Update(msg)
	a.menu = menuModel.(MenuModel)

	switch a.menu.Selected() {
	case "encrypt":
		// Show algorithm selection menu
		a.state = StateAlgoSelect
		a.isEncrypting = true
		a.algoMenu = NewAlgoMenuModel()
		a.menu = NewMenuModel()
		return a, nil

	case "decrypt":
		a.state = StateFileBrowser
		a.isEncrypting = false
		a.useAgeKey = false
		a.fileBrowser = NewFileBrowser("🔓 Select File to Decrypt", []string{".enc", ".podx", ".age"})
		a.menu = NewMenuModel()
		return a, nil

	case "env-encrypt":
		// Show env algorithm selection
		a.state = StateEnvAlgoSelect
		a.isEnvMode = true
		a.isEncrypting = true
		a.useAgeKey = false
		a.envAlgoMenu = NewEnvAlgoMenuModel()
		a.menu = NewMenuModel()
		return a, nil

	case "env-decrypt":
		// Show .env.podx files
		a.state = StateEnvDecryptBrowser
		a.isEnvMode = true
		a.isEncrypting = false
		a.useAgeKey = false
		a.fileBrowser = NewFileBrowser("📂 Select encrypted .env to decrypt", []string{".podx", ".enc"})
		a.menu = NewMenuModel()
		return a, nil

	case "keygen":
		// Show key type selection menu
		a.state = StateKeygenSelect
		a.keygenMenu = NewKeygenMenuModel()
		a.menu = NewMenuModel()
		return a, nil

	case "project":
		a.state = StateProjectMenu
		a.projectMenu = NewProjectMenuModel()
		a.menu = NewMenuModel()
		return a, nil

	case "age-encrypt":
		a.state = StateFileBrowser
		a.isEncrypting = true
		a.useAgeKey = true
		a.fileBrowser = NewFileBrowser("🔑 Select File to Encrypt (Age)", nil)
		a.fileBrowser.SetExcludeSuffixes([]string{".enc", ".podx", ".age"})
		a.menu = NewMenuModel()
		return a, nil

	case "age-decrypt":
		a.state = StateFileBrowser
		a.isEncrypting = false
		a.useAgeKey = true
		a.fileBrowser = NewFileBrowser("🗝️ Select File to Decrypt (Age)", []string{".age"})
		a.menu = NewMenuModel()
		return a, nil

	case "update":
		if newVersion, available := updater.CheckUpdate(a.version); available {
			a.result = NewResultModel(
				"Update Available",
				fmt.Sprintf("New version: %s\n\nRun 'podx update' from terminal to install.", newVersion),
				true,
			)
		} else {
			a.result = NewResultModel(
				"Up to Date",
				fmt.Sprintf("You are running the latest version (%s)", a.version),
				true,
			)
		}
		a.state = StateResult
		a.menu = NewMenuModel()
		return a, nil
	}

	return a, cmd
}

// File browser handling
func (a App) updateFileBrowser(msg tea.Msg) (tea.Model, tea.Cmd) {
	browserModel, cmd := a.fileBrowser.Update(msg)
	a.fileBrowser = browserModel.(FileBrowserModel)

	if a.fileBrowser.Cancelled() {
		a.state = StateMenu
		a.menu = NewMenuModel()
		a.useAgeKey = false
		return a, nil
	}

	if a.fileBrowser.Done() {
		a.inputFile = a.fileBrowser.Selected()

		// .env mode - format preserving
		if a.isEnvMode {
			if a.isEncrypting {
				a.outputFile = a.inputFile + ".podx"
				// Password encryption for .env
				a.state = StateEnvPassword
				a.passwordForm = NewPasswordForm(a.inputFile, a.outputFile, a.selectedAlgo, true)
				return a, a.passwordForm.Init()
			} else {
				// Decrypt .env
				a.outputFile = strings.TrimSuffix(a.inputFile, ".podx")
				a.outputFile = strings.TrimSuffix(a.outputFile, ".enc")
				if a.outputFile == a.inputFile {
					a.outputFile = a.inputFile + ".dec"
				}
				// Password decrypt for .env
				a.state = StateEnvDecryptPassword
				a.passwordForm = NewPasswordForm(a.inputFile, a.outputFile, "", false)
				return a, a.passwordForm.Init()
			}
		}

		// Age key encryption (no password needed)
		if a.useAgeKey {
			if a.isEncrypting {
				a.outputFile = a.inputFile + ".age"
				err := encryptWithAge(a.inputFile, a.outputFile)
				if err != nil {
					a.result = NewResultModel("Age Encryption Failed", err.Error(), false)
				} else {
					a.result = NewResultModel(
						"Age Encryption Successful",
						fmt.Sprintf("✓ %s\n  → %s\n\nEncrypted with your Age public key\nOriginal deleted", a.inputFile, a.outputFile),
						true,
					)
				}
			} else {
				// Remove .age extension
				a.outputFile = strings.TrimSuffix(a.inputFile, ".age")
				if a.outputFile == a.inputFile {
					a.outputFile = a.inputFile + ".dec"
				}
				err := decryptWithAge(a.inputFile, a.outputFile)
				if err != nil {
					a.result = NewResultModel("Age Decryption Failed", err.Error(), false)
				} else {
					a.result = NewResultModel(
						"Age Decryption Successful",
						fmt.Sprintf("✓ %s\n  → %s\n\nDecrypted with your Age private key", a.inputFile, a.outputFile),
						true,
					)
				}
			}
			a.state = StateResult
			a.useAgeKey = false
			return a, nil
		}

		// Password-based encryption
		if a.isEncrypting {
			a.outputFile = a.inputFile + ".enc"
			a.state = StateEncryptPassword
			a.passwordForm = NewPasswordForm(a.inputFile, a.outputFile, a.selectedAlgo, true)
		} else {
			// Auto-detect Age files
			if strings.HasSuffix(a.inputFile, ".age") {
				// Age decryption (no password needed)
				a.outputFile = strings.TrimSuffix(a.inputFile, ".age")
				if a.outputFile == a.inputFile {
					a.outputFile = a.inputFile + ".dec"
				}
				err := decryptWithAge(a.inputFile, a.outputFile)
				if err != nil {
					a.result = NewResultModel("Age Decryption Failed", err.Error(), false)
				} else {
					a.result = NewResultModel(
						"Age Decryption Successful",
						fmt.Sprintf("✓ %s\n  → %s\n\nDecrypted with your Age private key", a.inputFile, a.outputFile),
						true,
					)
				}
				a.state = StateResult
				return a, nil
			}

			// Password-based decryption
			a.outputFile = strings.TrimSuffix(a.inputFile, ".enc")
			a.outputFile = strings.TrimSuffix(a.outputFile, ".podx")
			if a.outputFile == a.inputFile {
				a.outputFile = a.inputFile + ".dec"
			}
			a.state = StateDecryptPassword
			a.passwordForm = NewPasswordForm(a.inputFile, a.outputFile, "", false)
		}
		return a, a.passwordForm.Init()
	}

	return a, cmd
}

// Password form handling
func (a App) updatePasswordForm(msg tea.Msg, isEncrypt bool) (tea.Model, tea.Cmd) {
	formModel, cmd := a.passwordForm.Update(msg)
	a.passwordForm = formModel.(PasswordFormModel)

	if a.passwordForm.Cancelled() {
		a.state = StateMenu
		a.menu = NewMenuModel()
		return a, nil
	}

	if a.passwordForm.Done() {
		password := a.passwordForm.Password()
		var err error

		if isEncrypt {
			algo := a.passwordForm.algo
			if algo == "" {
				algo = "aes-gcm"
			}
			err = encryptFile(a.inputFile, a.outputFile, algo, password)
			if err != nil {
				a.result = NewResultModel("Encryption Failed", err.Error(), false)
			} else {
				a.result = NewResultModel(
					"Encryption Successful",
					fmt.Sprintf("✓ %s\n  → %s\n\nAlgorithm: %s\nOriginal deleted", a.inputFile, a.outputFile, algo),
					true,
				)
			}
		} else {
			err = decryptFile(a.inputFile, a.outputFile, password)
			if err != nil {
				a.result = NewResultModel("Decryption Failed", err.Error(), false)
			} else {
				a.result = NewResultModel(
					"Decryption Successful",
					fmt.Sprintf("✓ %s\n  → %s", a.inputFile, a.outputFile),
					true,
				)
			}
		}

		a.state = StateResult
		return a, nil
	}

	return a, cmd
}

// Generic form handling
func (a App) updateGenericForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	formModel, cmd := a.form.Update(msg)
	a.form = formModel.(FormModel)

	if a.form.Cancelled() {
		a.state = StateMenu
		a.menu = NewMenuModel()
		return a, nil
	}

	if a.form.Submitted() {
		switch a.state {
		case StateKeygenForm:
			values := a.form.Values()
			keyType := values[0]
			if keyType == "" {
				keyType = "age"
			}

			if keyType == "age" {
				result, err := keygen.GenerateAge()
				if err != nil {
					a.result = NewResultModel("Key Generation Failed", err.Error(), false)
				} else {
					a.result = NewResultModel(
						"Age Key Generated",
						fmt.Sprintf("Public Key:\n%s\n\nSaved to:\n%s", result.PublicKey, result.KeyFile),
						true,
					)
				}
			} else if keyType == "gpg" {
				name := values[1]
				email := values[2]
				if name == "" || email == "" {
					a.result = NewResultModel("Key Generation Failed", "Name and email are required for GPG", false)
				} else {
					result, err := keygen.GenerateGPG(name, email)
					if err != nil {
						a.result = NewResultModel("Key Generation Failed", err.Error(), false)
					} else {
						a.result = NewResultModel(
							"GPG Key Generated",
							fmt.Sprintf("Key ID: %s\nEmail: %s", result.PublicKey, result.Email),
							true,
						)
					}
				}
			} else {
				a.result = NewResultModel("Key Generation Failed", "Unknown key type. Use 'age' or 'gpg'.", false)
			}
		}

		a.state = StateResult
		return a, nil
	}

	return a, cmd
}

// Project menu handling
func (a App) updateProjectMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	menuModel, cmd := a.projectMenu.Update(msg)
	a.projectMenu = menuModel.(MenuModel)

	switch a.projectMenu.Selected() {
	case "back":
		a.state = StateMenu
		a.menu = NewMenuModel()
		return a, nil
	case "init", "status", "encrypt-all", "decrypt-all":
		a.result = NewResultModel(
			"Project Command",
			fmt.Sprintf("Run this command:\n\npodx %s", a.projectMenu.Selected()),
			true,
		)
		a.state = StateResult
		a.projectMenu = MenuModel{}
		return a, nil
	}

	return a, cmd
}

// Result handling
func (a App) updateResult(msg tea.Msg) (tea.Model, tea.Cmd) {
	resultModel, _ := a.result.Update(msg)
	a.result = resultModel.(ResultModel)

	if a.result.Done() {
		a.state = StateMenu
		a.menu = NewMenuModel()
		return a, nil
	}

	return a, nil
}

// NewProjectMenuModel creates project submenu
func NewProjectMenuModel() MenuModel {
	items := []list.Item{
		MenuItem{title: "📁 Init Project", description: "Initialize .podx.yaml", action: "init"},
		MenuItem{title: "📊 Status", description: "Show project status", action: "status"},
		MenuItem{title: "🔐 Encrypt All", description: "Encrypt all project secrets", action: "encrypt-all"},
		MenuItem{title: "🔓 Decrypt All", description: "Decrypt all project secrets", action: "decrypt-all"},
		MenuItem{title: "⬅️  Back", description: "Return to main menu", action: "back"},
	}

	l := list.New(items, menuItemDelegate{}, 55, 12)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return MenuModel{list: l}
}

// NewAlgoMenuModel creates algorithm selection menu
func NewAlgoMenuModel() MenuModel {
	items := []list.Item{
		MenuItem{title: "🔐 AES-GCM + Password", description: "Encrypt with password (AES-256-GCM)", action: "aes-gcm"},
		MenuItem{title: "⚡ ChaCha20 + Password", description: "Encrypt with password (ChaCha20-Poly1305)", action: "chacha20"},
		MenuItem{title: "🔑 Age Key", description: "Encrypt with Age public key (no password)", action: "age"},
		MenuItem{title: "⬅️  Back", description: "Return to main menu", action: "back"},
	}

	l := list.New(items, menuItemDelegate{}, 55, 10)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return MenuModel{list: l}
}

// NewKeygenMenuModel creates keygen type selection menu
func NewKeygenMenuModel() MenuModel {
	items := []list.Item{
		MenuItem{title: "🔑 Age Key", description: "Generate Age X25519 key pair (recommended)", action: "age"},
		MenuItem{title: "🗝️  GPG Key", description: "Generate GPG/PGP key pair", action: "gpg"},
		MenuItem{title: "⬅️  Back", description: "Return to main menu", action: "back"},
	}

	l := list.New(items, menuItemDelegate{}, 55, 8)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return MenuModel{list: l}
}

// NewEnvAlgoMenuModel creates .env algorithm selection menu
func NewEnvAlgoMenuModel() MenuModel {
	items := []list.Item{
		MenuItem{title: "🔐 AES-GCM + Password", description: "Format-preserving with AES-256-GCM", action: "aes-gcm"},
		MenuItem{title: "⚡ ChaCha20 + Password", description: "Format-preserving with ChaCha20", action: "chacha20"},
		MenuItem{title: "⬅️  Back", description: "Return to main menu", action: "back"},
	}

	l := list.New(items, menuItemDelegate{}, 55, 10)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return MenuModel{list: l}
}

// Env algorithm selection handler
func (a App) updateEnvAlgoSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	menuModel, cmd := a.envAlgoMenu.Update(msg)
	a.envAlgoMenu = menuModel.(MenuModel)

	switch a.envAlgoMenu.Selected() {
	case "back":
		a.state = StateMenu
		a.menu = NewMenuModel()
		a.envAlgoMenu = MenuModel{}
		a.isEnvMode = false
		return a, nil

	case "aes-gcm", "chacha20":
		a.selectedAlgo = a.envAlgoMenu.Selected()
		a.useAgeKey = false
		a.state = StateEnvFileBrowser
		a.fileBrowser = NewFileBrowser("📄 Select .env file to encrypt", []string{".env"})
		a.fileBrowser.SetExcludeSuffixes([]string{".enc", ".podx", ".age"})
		a.envAlgoMenu = MenuModel{}
		return a, nil
	}

	return a, cmd
}

// Env password form handler
func (a App) updateEnvPasswordForm(msg tea.Msg, isEncrypt bool) (tea.Model, tea.Cmd) {
	formModel, cmd := a.passwordForm.Update(msg)
	a.passwordForm = formModel.(PasswordFormModel)

	if a.passwordForm.Cancelled() {
		a.state = StateMenu
		a.menu = NewMenuModel()
		a.isEnvMode = false
		return a, nil
	}

	if a.passwordForm.Done() {
		password := a.passwordForm.Password()

		if isEncrypt {
			// Actually encrypt the .env file
			err := envEncryptFile(a.inputFile, a.outputFile, a.selectedAlgo, password)
			if err != nil {
				a.result = NewResultModel("Env Encryption Failed", err.Error(), false)
			} else {
				a.result = NewResultModel(
					"Env Encrypted ✓",
					fmt.Sprintf("✓ %s\n  → %s\n\nFormat-preserving encryption with %s\nOriginal deleted", a.inputFile, a.outputFile, a.selectedAlgo),
					true,
				)
			}
		} else {
			// Actually decrypt the .env file
			err := envDecryptFile(a.inputFile, a.outputFile, password)
			if err != nil {
				a.result = NewResultModel("Env Decryption Failed", err.Error(), false)
			} else {
				a.result = NewResultModel(
					"Env Decrypted ✓",
					fmt.Sprintf("✓ %s\n  → %s\n\nDecrypted successfully", a.inputFile, a.outputFile),
					true,
				)
			}
		}

		a.state = StateResult
		a.isEnvMode = false
		return a, nil
	}

	return a, cmd
}

// Algorithm selection handler
func (a App) updateAlgoSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	menuModel, cmd := a.algoMenu.Update(msg)
	a.algoMenu = menuModel.(MenuModel)

	switch a.algoMenu.Selected() {
	case "back":
		a.state = StateMenu
		a.menu = NewMenuModel()
		a.algoMenu = MenuModel{}
		return a, nil

	case "aes-gcm":
		a.selectedAlgo = "aes-gcm"
		a.useAgeKey = false
		a.state = StateFileBrowser
		a.fileBrowser = NewFileBrowser("🔐 Select File to Encrypt (AES-GCM)", nil)
		a.fileBrowser.SetExcludeSuffixes([]string{".enc", ".podx", ".age"})
		a.algoMenu = MenuModel{}
		return a, nil

	case "chacha20":
		a.selectedAlgo = "chacha20"
		a.useAgeKey = false
		a.state = StateFileBrowser
		a.fileBrowser = NewFileBrowser("⚡ Select File to Encrypt (ChaCha20)", nil)
		a.fileBrowser.SetExcludeSuffixes([]string{".enc", ".podx", ".age"})
		a.algoMenu = MenuModel{}
		return a, nil

	case "age":
		a.useAgeKey = true
		a.state = StateFileBrowser
		a.fileBrowser = NewFileBrowser("🔑 Select File to Encrypt (Age)", nil)
		a.fileBrowser.SetExcludeSuffixes([]string{".enc", ".podx", ".age"})
		a.algoMenu = MenuModel{}
		return a, nil
	}

	return a, cmd
}

// Keygen type selection handler
func (a App) updateKeygenSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	menuModel, cmd := a.keygenMenu.Update(msg)
	a.keygenMenu = menuModel.(MenuModel)

	switch a.keygenMenu.Selected() {
	case "back":
		a.state = StateMenu
		a.menu = NewMenuModel()
		a.keygenMenu = MenuModel{}
		return a, nil

	case "age":
		result, err := keygen.GenerateAge()
		if err != nil {
			a.result = NewResultModel("Key Generation Failed", err.Error(), false)
		} else {
			a.result = NewResultModel(
				"Age Key Generated ✓",
				fmt.Sprintf("Public Key:\n%s\n\nSaved to:\n%s", result.PublicKey, result.KeyFile),
				true,
			)
		}
		a.state = StateResult
		a.keygenMenu = MenuModel{}
		return a, nil

	case "gpg":
		a.state = StateKeygenForm
		a.form = NewFormModel("🗝️ Generate GPG Key", []FormField{
			{Label: "Your Name", Placeholder: "John Doe", Value: ""},
			{Label: "Your Email", Placeholder: "john@example.com", Value: ""},
		})
		a.keygenMenu = MenuModel{}
		return a, a.form.Init()
	}

	return a, cmd
}

// Helper functions for encryption/decryption
func encryptFile(input, output, algo, password string) error {
	if input == "" || output == "" || password == "" {
		return fmt.Errorf("all fields are required")
	}
	if samePath(input, output) {
		return fmt.Errorf("output must be different from input")
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	key, salt, err := crypto.DeriveKey([]byte(password), nil)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	plaintext, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	enc, err := crypto.NewEncryptor(crypto.Algorithm(algo))
	if err != nil {
		return err
	}

	ciphertext, err := enc.Encrypt(plaintext, key)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	algoB := byte(0)
	if algo == "chacha20" {
		algoB = 1
	}

	out := append(salt, algoB)
	out = append(out, ciphertext...)

	if err := os.WriteFile(output, out, 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if err := removeOriginal(input, output); err != nil {
		return fmt.Errorf("failed to delete original: %w", err)
	}

	return nil
}

func decryptFile(input, output, password string) error {
	if input == "" || output == "" || password == "" {
		return fmt.Errorf("all fields are required")
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if len(data) < crypto.SaltSize+1 {
		return fmt.Errorf("file too small or corrupted")
	}

	salt := data[:crypto.SaltSize]
	algoB := data[crypto.SaltSize]
	ciphertext := data[crypto.SaltSize+1:]

	algo := crypto.AlgoAESGCM
	if algoB == 1 {
		algo = crypto.AlgoChaCha20
	}

	key, err := crypto.DeriveKeyWithSalt([]byte(password), salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	enc, err := crypto.NewEncryptor(algo)
	if err != nil {
		return err
	}

	plaintext, err := enc.Decrypt(ciphertext, key)
	if err != nil {
		return fmt.Errorf("decryption failed (wrong password?): %w", err)
	}

	if err := os.WriteFile(output, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Age key encryption/decryption
func encryptWithAge(input, output string) error {
	if input == "" || output == "" {
		return fmt.Errorf("input and output are required")
	}
	if samePath(input, output) {
		return fmt.Errorf("output must be different from input")
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	// Load public key
	recipient, err := keygen.LoadAgeRecipient()
	if err != nil {
		return fmt.Errorf("failed to load Age public key: %w\n\nRun 'podx keygen -t age' first to generate keys", err)
	}

	// Read input
	plaintext, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Encrypt
	ciphertext, err := crypto.AgeEncrypt(plaintext, recipient)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Write output
	if err := os.WriteFile(output, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if err := removeOriginal(input, output); err != nil {
		return fmt.Errorf("failed to delete original: %w", err)
	}

	return nil
}

func decryptWithAge(input, output string) error {
	if input == "" || output == "" {
		return fmt.Errorf("input and output are required")
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	// Load private key
	identity, err := keygen.LoadAgeIdentity()
	if err != nil {
		return fmt.Errorf("failed to load Age private key: %w\n\nRun 'podx keygen -t age' first to generate keys", err)
	}

	// Read input
	ciphertext, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Decrypt
	plaintext, err := crypto.AgeDecrypt(ciphertext, identity)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write output
	if err := os.WriteFile(output, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Env format-preserving encryption
func envEncryptFile(input, output, algo, password string) error {
	if input == "" || output == "" || password == "" {
		return fmt.Errorf("all fields are required")
	}
	if samePath(input, output) {
		return fmt.Errorf("output must be different from input")
	}
	if algo == "" {
		algo = "aes-gcm"
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	// Parse .env file
	entries, err := parser.ParseEnvFile(input)
	if err != nil {
		return fmt.Errorf("failed to parse .env: %w", err)
	}

	// Derive key from password
	key, salt, err := crypto.DeriveKey([]byte(password), nil)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	// Encrypt values
	if err := parser.EncryptEnvValues(entries, key, crypto.Algorithm(algo)); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Write encrypted file
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	saltEntry := parser.EnvEntry{
		IsComment: true,
		Comment:   fmt.Sprintf("# IRONVAULT_SALT=%s", saltB64),
	}
	entries = append([]parser.EnvEntry{saltEntry}, entries...)
	if err := parser.WriteEnvFile(output, entries); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if err := removeOriginal(input, output); err != nil {
		return fmt.Errorf("failed to delete original: %w", err)
	}

	return nil
}

func envDecryptFile(input, output, password string) error {
	if input == "" || output == "" || password == "" {
		return fmt.Errorf("all fields are required")
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	// Parse encrypted .env file
	entries, err := parser.ParseEnvFile(input)
	if err != nil {
		return fmt.Errorf("failed to parse .env: %w", err)
	}

	var salt []byte
	var cleanEntries []parser.EnvEntry

	for _, entry := range entries {
		if entry.IsComment && strings.HasPrefix(entry.Comment, "# IRONVAULT_SALT=") {
			saltB64 := strings.TrimPrefix(entry.Comment, "# IRONVAULT_SALT=")
			salt, err = base64.StdEncoding.DecodeString(saltB64)
			if err != nil {
				return fmt.Errorf("failed to decode salt: %w", err)
			}
			continue
		}
		cleanEntries = append(cleanEntries, entry)
	}

	if len(salt) == 0 {
		return fmt.Errorf("no salt found in encrypted file")
	}

	// Derive key from password
	key, err := crypto.DeriveKeyWithSalt([]byte(password), salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	// Decrypt values
	if err := parser.DecryptEnvValues(cleanEntries, key); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write decrypted file
	if err := parser.WriteEnvFile(output, cleanEntries); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

func samePath(a, b string) bool {
	aAbs, errA := filepath.Abs(a)
	bAbs, errB := filepath.Abs(b)
	if errA == nil && errB == nil {
		return aAbs == bAbs
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func removeOriginal(input, output string) error {
	if samePath(input, output) {
		return fmt.Errorf("output must be different from input")
	}
	return os.Remove(input)
}

// Run starts the TUI application
func Run(version string) error {
	p := tea.NewProgram(NewApp(version), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
