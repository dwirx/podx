package tui

import (
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommandItem represents a podx command
type CommandItem struct {
	id           string // Unique identifier for the command
	name         string
	description  string
	category     string
	args         []string
	needsForm    bool
	needsConfirm bool
}

// Title returns the command name (implements list.Item)
func (c CommandItem) Title() string {
	return c.name
}

// Description returns the command description (implements list.Item)
func (c CommandItem) Description() string {
	return c.description
}

// FilterValue returns the value to filter on (implements list.Item)
func (c CommandItem) FilterValue() string {
	return c.name + " " + c.description + " " + c.category
}

// ID returns the command ID
func (c CommandItem) ID() string {
	return c.id
}

// commandOutputMsg is sent when command execution completes
type commandOutputMsg struct {
	output  string
	err     error
	command string
}

// CommandsModel represents the commands tab content
type CommandsModel struct {
	list          list.Model
	width         int
	height        int
	keys          KeyMap
	showOutput    bool
	output        string
	outputErr     error
	outputCommand string
	running       bool

	// Dialogs
	formDialog    *FormDialogModel
	confirmDialog *ConfirmDialogModel
	progressView  *ProgressViewModel
	pendingAction string
}

// GetAllCommands returns all available commands grouped by category
func GetAllCommands() []list.Item {
	return []list.Item{
		// Project Commands
		CommandItem{
			id: "init", name: "init",
			description: "Initialize PODX project",
			category:    "Project", args: []string{"init"},
		},
		CommandItem{
			id: "add-recipient", name: "add-recipient",
			description: "Add team member (requires name & key)",
			category:    "Project", args: []string{"add-recipient"},
			needsForm: true,
		},
		CommandItem{
			id: "encrypt-all", name: "encrypt-all",
			description: "Encrypt all secrets (deletes originals!)",
			category:    "Project", args: []string{"encrypt-all"},
			needsConfirm: true,
		},
		CommandItem{
			id: "decrypt-all", name: "decrypt-all",
			description: "Decrypt all secrets",
			category:    "Project", args: []string{"decrypt-all"},
			needsConfirm: true,
		},
		CommandItem{
			id: "status", name: "status",
			description: "Show project status",
			category:    "Project", args: []string{"status"},
		},
		CommandItem{
			id: "sync", name: "sync",
			description: "Encrypt, commit and push (safe git workflow)",
			category:    "Project", args: []string{"sync"},
			needsForm: true,
		},

		// File Commands
		CommandItem{
			id: "encrypt", name: "encrypt",
			description: "Encrypt single file with password",
			category:    "File", args: []string{"encrypt"},
			needsForm: true,
		},
		CommandItem{
			id: "decrypt", name: "decrypt",
			description: "Decrypt single file",
			category:    "File", args: []string{"decrypt"},
			needsForm: true,
		},
		CommandItem{
			id: "env-encrypt", name: "env encrypt",
			description: "Encrypt .env file (format-preserving)",
			category:    "File", args: []string{"env", "encrypt"},
			needsForm: true,
		},
		CommandItem{
			id: "env-decrypt", name: "env decrypt",
			description: "Decrypt .env file",
			category:    "File", args: []string{"env", "decrypt"},
			needsForm: true,
		},

		// Key Commands
		CommandItem{
			id: "keygen-age", name: "keygen (Age)",
			description: "Generate Age X25519 key pair",
			category:    "Keys", args: []string{"keygen", "-t", "age"},
		},
		CommandItem{
			id: "keygen-gpg", name: "keygen (GPG)",
			description: "Generate GPG key pair",
			category:    "Keys", args: []string{"keygen", "-t", "gpg"},
			needsForm: true,
		},

		// Security Commands
		CommandItem{
			id: "check", name: "check",
			description: "Run security checks",
			category:    "Security", args: []string{"check"},
		},
		CommandItem{
			id: "check-fix", name: "check --fix",
			description: "Fix gitignore issues automatically",
			category:    "Security", args: []string{"check", "--fix"},
		},
		CommandItem{
			id: "hook-install", name: "hook install",
			description: "Install pre-commit hook",
			category:    "Security", args: []string{"hook", "install"},
		},
		CommandItem{
			id: "hook-uninstall", name: "hook uninstall",
			description: "Remove pre-commit hook",
			category:    "Security", args: []string{"hook", "uninstall"},
			needsConfirm: true,
		},
		CommandItem{
			id: "hook-status", name: "hook status",
			description: "Check if hook is installed",
			category:    "Security", args: []string{"hook", "status"},
		},

		// Update Commands
		CommandItem{
			id: "update", name: "update",
			description: "Update PODX to latest version",
			category:    "System", args: []string{"update"},
			needsConfirm: true,
		},
		CommandItem{
			id: "update-beta", name: "update --beta",
			description: "Update to beta version (pre-release)",
			category:    "System", args: []string{"update", "--beta"},
			needsConfirm: true,
		},
		CommandItem{
			id: "rollback", name: "rollback",
			description: "Rollback to previous version",
			category:    "System", args: []string{"rollback"},
			needsConfirm: true,
		},
		CommandItem{
			id: "version", name: "version",
			description: "Show version info",
			category:    "System", args: []string{"version"},
		},
		CommandItem{
			id: "key-info", name: "key-info",
			description: "Show your Age public key",
			category:    "Keys", args: []string{"key-info"},
		},
	}
}

// NewCommandsModel creates a new commands model
func NewCommandsModel() CommandsModel {
	items := GetAllCommands()

	// Create list with custom delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedStyle
	delegate.Styles.SelectedDesc = MutedStyle

	l := list.New(items, delegate, 0, 0)
	l.Title = "📋 Available Commands"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	// Customize key bindings for vim navigation
	l.KeyMap.CursorUp = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "up"),
	)
	l.KeyMap.CursorDown = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "down"),
	)

	// Style the list
	l.Styles.Title = TitleStyle

	return CommandsModel{
		list: l,
		keys: DefaultKeyMap(),
	}
}

// Init initializes the commands model
func (m CommandsModel) Init() tea.Cmd {
	return nil
}

// executeCommand runs the selected command and returns a tea.Cmd
func (m CommandsModel) executeCommand(args []string, commandName string) tea.Cmd {
	return func() tea.Msg {
		// Get the podx binary path
		executable, err := os.Executable()
		if err != nil {
			executable = "podx"
		}

		// Create command
		cmd := exec.Command(executable, args...)
		cmd.Dir, _ = os.Getwd()

		// Capture combined output
		output, err := cmd.CombinedOutput()

		return commandOutputMsg{
			output:  string(output),
			err:     err,
			command: commandName,
		}
	}
}

// Update handles messages for the commands model
func (m CommandsModel) Update(msg tea.Msg) (CommandsModel, tea.Cmd) {
	// Handle form dialog if visible (only for key messages)
	if m.formDialog != nil && m.formDialog.IsVisible() {
		if _, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			*m.formDialog, cmd = m.formDialog.Update(msg)
			return m, cmd
		}
	}

	// Handle confirm dialog if visible (only for key messages)
	if m.confirmDialog != nil && m.confirmDialog.IsVisible() {
		if _, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			*m.confirmDialog, cmd = m.confirmDialog.Update(msg)
			return m, cmd
		}
	}

	// Handle progress view if visible (only for key messages)
	// commandOutputMsg must be processed below, not intercepted here
	if m.progressView != nil && m.progressView.IsVisible() {
		if _, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			*m.progressView, cmd = m.progressView.Update(msg)
			if !m.progressView.IsVisible() {
				m.progressView = nil
			}
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case FormDialogResult:
		if msg.Submitted {
			// Validate and execute command
			args := BuildCommandArgs(msg.CommandID, msg.Values)
			if len(args) > 0 {
				m.running = true
				// Show progress view
				pv := NewProgressView("Running: " + msg.CommandID)
				m.progressView = &pv
				return m, m.executeCommand(args, msg.CommandID)
			}
		}
		m.formDialog = nil
		return m, nil

	case ConfirmDialogResult:
		if msg.Confirmed {
			// Find and execute the command
			for _, item := range GetAllCommands() {
				if cmdItem, ok := item.(CommandItem); ok && cmdItem.id == msg.ActionID {
					m.running = true
					// Show progress view
					pv := NewProgressView("Running: " + cmdItem.name)
					m.progressView = &pv
					return m, m.executeCommand(cmdItem.args, cmdItem.name)
				}
			}
		}
		m.confirmDialog = nil
		return m, nil

	case commandOutputMsg:
		m.running = false

		// Update progress view with result
		if m.progressView != nil {
			m.progressView.Complete(msg.err == nil, "")
			m.progressView.Output = msg.output
		} else {
			m.showOutput = true
			m.output = msg.output
			m.outputErr = msg.err
			m.outputCommand = msg.command
		}
		return m, nil

	case tea.KeyMsg:
		// If showing output, handle back navigation
		if m.showOutput {
			switch {
			case key.Matches(msg, m.keys.Left), msg.String() == "h", msg.String() == "esc", msg.String() == "q":
				m.showOutput = false
				m.output = ""
				m.outputErr = nil
				return m, nil
			}
			return m, nil
		}

		// If running, don't process input
		if m.running {
			return m, nil
		}

		// Handle enter/l to execute selected command
		switch {
		case key.Matches(msg, m.keys.Enter), msg.String() == "l":
			// Don't trigger if filtering is active
			if m.list.FilterState() == list.Filtering {
				break
			}

			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				if item, ok := selectedItem.(CommandItem); ok {
					return m.handleCommandSelection(item)
				}
			}
			return m, nil
		}

		// Pass other keys to the list
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	// Handle other messages (like window size)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// handleCommandSelection handles when a command is selected
func (m CommandsModel) handleCommandSelection(item CommandItem) (CommandsModel, tea.Cmd) {
	// Check if command needs form input
	if item.needsForm {
		form := CreateFormForCommand(item.id)
		if form != nil {
			m.formDialog = form
			return m, form.Init()
		}
	}

	// Check if command needs confirmation
	if item.needsConfirm {
		title, message := GetConfirmMessage(item.id)
		confirm := NewConfirmDialog(title, message, item.id)
		m.confirmDialog = &confirm
		return m, nil
	}

	// Execute directly
	m.running = true
	// Show progress view
	pv := NewProgressView("Running: " + item.name)
	m.progressView = &pv
	return m, m.executeCommand(item.args, item.name)
}

// View renders the commands model
func (m CommandsModel) View() string {
	// Render form dialog overlay
	if m.formDialog != nil && m.formDialog.IsVisible() {
		dialog := m.formDialog.View()
		// Center the dialog
		centered := CenterDialog(dialog, m.width, m.height)
		return centered
	}

	// Render confirm dialog overlay
	if m.confirmDialog != nil && m.confirmDialog.IsVisible() {
		dialog := m.confirmDialog.View()
		centered := CenterDialog(dialog, m.width, m.height)
		return centered
	}

	// Render progress view overlay
	if m.progressView != nil && m.progressView.IsVisible() {
		dialog := m.progressView.View()
		centered := CenterDialog(dialog, m.width, m.height)
		return centered
	}

	if m.running {
		return BoxStyle.Render("⏳ Running command...")
	}

	if m.showOutput {
		return m.renderOutput()
	}

	// Render help bar at bottom
	helpBar := m.renderHelpBar()
	listView := m.list.View()

	return lipgloss.JoinVertical(lipgloss.Left, listView, helpBar)
}

// renderOutput renders the command output view
func (m CommandsModel) renderOutput() string {
	var lines []string

	// Header
	header := TitleStyle.Render("📋 Command Output: " + m.outputCommand)
	lines = append(lines, header)
	lines = append(lines, "")

	// Status
	if m.outputErr != nil {
		lines = append(lines, ErrorStyle.Render("❌ Error: "+m.outputErr.Error()))
		lines = append(lines, "")
	} else {
		lines = append(lines, SuccessStyle.Render("✓ Command completed successfully"))
		lines = append(lines, "")
	}

	// Separator
	lines = append(lines, strings.Repeat("─", 50))
	lines = append(lines, "")

	// Render output with scrolling
	if m.output != "" {
		outputLines := strings.Split(strings.TrimSpace(m.output), "\n")
		maxLines := m.height - 12
		if maxLines < 5 {
			maxLines = 5
		}

		// Show last N lines if output is too long
		if len(outputLines) > maxLines {
			lines = append(lines, MutedStyle.Render(strings.Repeat("... ", 10)))
			outputLines = outputLines[len(outputLines)-maxLines:]
		}

		for _, line := range outputLines {
			// Colorize output
			coloredLine := m.colorizeOutput(line)
			lines = append(lines, "  "+coloredLine)
		}
	} else {
		lines = append(lines, MutedStyle.Render("  (No output)"))
	}

	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", 50))
	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("Press h/←/Esc to go back"))

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// colorizeOutput adds colors to output based on content
func (m CommandsModel) colorizeOutput(line string) string {
	lowerLine := strings.ToLower(line)

	// Success indicators
	if strings.Contains(lowerLine, "✓") || strings.Contains(lowerLine, "success") ||
		strings.Contains(lowerLine, "passed") || strings.Contains(lowerLine, "ok") {
		return SuccessStyle.Render(line)
	}

	// Error indicators
	if strings.Contains(lowerLine, "✗") || strings.Contains(lowerLine, "error") ||
		strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "fail") {
		return ErrorStyle.Render(line)
	}

	// Warning indicators
	if strings.Contains(lowerLine, "warning") || strings.Contains(lowerLine, "⚠") {
		return WarningStyle.Render(line)
	}

	// Keys and important values
	if strings.HasPrefix(line, "  age1") || strings.HasPrefix(line, "age1") {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(line)
	}

	return line
}

// renderHelpBar renders the help bar at the bottom
func (m CommandsModel) renderHelpBar() string {
	helpItems := []string{
		"j/k: navigate",
		"/: filter",
		"Enter/l: execute",
		"Esc: back",
	}

	help := strings.Join(helpItems, " │ ")
	return MutedStyle.Render("\n" + help)
}

// SetSize updates the model dimensions
func (m *CommandsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Account for box padding and help bar
	m.list.SetSize(width-4, height-6)
}
