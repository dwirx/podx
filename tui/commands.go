package tui

import (
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandItem represents a podx command
type CommandItem struct {
	name        string
	description string
	args        []string
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
	return c.name + " " + c.description
}

// commandOutputMsg is sent when command execution completes
type commandOutputMsg struct {
	output string
	err    error
}

// CommandsModel represents the commands tab content
type CommandsModel struct {
	list       list.Model
	width      int
	height     int
	keys       KeyMap
	showOutput bool
	output     string
	outputErr  error
	running    bool
}

// NewCommandsModel creates a new commands model
func NewCommandsModel() CommandsModel {
	items := []list.Item{
		CommandItem{"init", "Initialize PODX project", []string{"init"}},
		CommandItem{"add-recipient", "Add team member", []string{"add-recipient"}},
		CommandItem{"encrypt-all", "Encrypt all secrets", []string{"encrypt-all"}},
		CommandItem{"decrypt-all", "Decrypt all secrets", []string{"decrypt-all"}},
		CommandItem{"status", "Show project status", []string{"status"}},
		CommandItem{"check", "Run security checks", []string{"check"}},
		CommandItem{"check --fix", "Fix gitignore issues", []string{"check", "--fix"}},
		CommandItem{"hook install", "Install pre-commit hook", []string{"hook", "install"}},
		CommandItem{"hook uninstall", "Remove pre-commit hook", []string{"hook", "uninstall"}},
		CommandItem{"hook status", "Check hook status", []string{"hook", "status"}},
		CommandItem{"keygen", "Generate Age key pair", []string{"keygen", "-t", "age"}},
		CommandItem{"update", "Update PODX", []string{"update"}},
	}

	// Create list with custom delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedStyle
	delegate.Styles.SelectedDesc = MutedStyle

	l := list.New(items, delegate, 0, 0)
	l.Title = "Available Commands"
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
func (m CommandsModel) executeCommand(item CommandItem) tea.Cmd {
	return func() tea.Msg {
		// Get the podx binary path
		executable, err := os.Executable()
		if err != nil {
			executable = "podx"
		}

		// Create command
		cmd := exec.Command(executable, item.args...)
		cmd.Dir, _ = os.Getwd()

		// Capture combined output
		output, err := cmd.CombinedOutput()

		return commandOutputMsg{
			output: string(output),
			err:    err,
		}
	}
}

// Update handles messages for the commands model
func (m CommandsModel) Update(msg tea.Msg) (CommandsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case commandOutputMsg:
		m.running = false
		m.showOutput = true
		m.output = msg.output
		m.outputErr = msg.err
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
					m.running = true
					return m, m.executeCommand(item)
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

// View renders the commands model
func (m CommandsModel) View() string {
	if m.running {
		return BoxStyle.Render("Running command...")
	}

	if m.showOutput {
		return m.renderOutput()
	}

	return m.list.View()
}

// renderOutput renders the command output view
func (m CommandsModel) renderOutput() string {
	var lines []string

	lines = append(lines, TitleStyle.Render("Command Output"))
	lines = append(lines, "")

	if m.outputErr != nil {
		lines = append(lines, ErrorStyle.Render("Error: "+m.outputErr.Error()))
		lines = append(lines, "")
	}

	// Render output
	if m.output != "" {
		outputLines := strings.Split(strings.TrimSpace(m.output), "\n")
		for _, line := range outputLines {
			lines = append(lines, "  "+line)
		}
	} else {
		lines = append(lines, MutedStyle.Render("  (No output)"))
	}

	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("Press h/left/esc to go back"))

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// SetSize updates the model dimensions
func (m *CommandsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Account for box padding
	m.list.SetSize(width-4, height-4)
}
