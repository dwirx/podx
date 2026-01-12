# PODX TUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add interactive TUI dashboard using Bubbletea with vim-style navigation.

**Architecture:** New `tui/` package with multi-tab interface integrating with existing project/security packages.

**Tech Stack:** Bubbletea, Lipgloss, Bubbles

---

### Task 1: Add Dependencies and Create Base Structure

**Files:**
- Modify: `go.mod`
- Create: `tui/tui.go`
- Create: `tui/styles.go`
- Create: `tui/keys.go`

**Step 1: Add Bubbletea dependencies**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

**Step 2: Create `tui/styles.go`**

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorCyan    = lipgloss.Color("86")
	ColorGreen   = lipgloss.Color("78")
	ColorYellow  = lipgloss.Color("220")
	ColorRed     = lipgloss.Color("196")
	ColorGray    = lipgloss.Color("241")
	ColorWhite   = lipgloss.Color("255")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	SelectedStyle = lipgloss.NewStyle().
			Background(ColorCyan).
			Foreground(lipgloss.Color("0"))

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGray).
			Padding(0, 1)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Background(lipgloss.Color("236")).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Padding(0, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Padding(0, 1)
)
```

**Step 3: Create `tui/keys.go`**

```go
package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Tab1     key.Binding
	Tab2     key.Binding
	Tab3     key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "back"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "select"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Tab1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "dashboard"),
	),
	Tab2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "commands"),
	),
	Tab3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "security"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
```

**Step 4: Create `tui/tui.go` (base structure)**

```go
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	TabDashboard = iota
	TabCommands
	TabSecurity
)

type Model struct {
	activeTab    int
	width        int
	height       int
	showHelp     bool
	statusMsg    string

	// Sub-models
	dashboard    DashboardModel
	commands     CommandsModel
	security     SecurityModel
}

func New() Model {
	return Model{
		activeTab: TabDashboard,
		dashboard: NewDashboard(),
		commands:  NewCommands(),
		security:  NewSecurity(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
		m.security.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch {
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, Keys.Tab):
			m.activeTab = (m.activeTab + 1) % 3
			return m, nil
		case key.Matches(msg, Keys.ShiftTab):
			m.activeTab = (m.activeTab + 2) % 3
			return m, nil
		case key.Matches(msg, Keys.Tab1):
			m.activeTab = TabDashboard
			return m, nil
		case key.Matches(msg, Keys.Tab2):
			m.activeTab = TabCommands
			return m, nil
		case key.Matches(msg, Keys.Tab3):
			m.activeTab = TabSecurity
			return m, nil
		}
	}

	// Delegate to active tab
	switch m.activeTab {
	case TabDashboard:
		newDashboard, cmd := m.dashboard.Update(msg)
		m.dashboard = newDashboard.(DashboardModel)
		cmds = append(cmds, cmd)
	case TabCommands:
		newCommands, cmd := m.commands.Update(msg)
		m.commands = newCommands.(CommandsModel)
		cmds = append(cmds, cmd)
	case TabSecurity:
		newSecurity, cmd := m.security.Update(msg)
		m.security = newSecurity.(SecurityModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	// Header
	header := m.renderHeader()

	// Tabs
	tabs := m.renderTabs()

	// Content based on active tab
	var content string
	switch m.activeTab {
	case TabDashboard:
		content = m.dashboard.View()
	case TabCommands:
		content = m.commands.View()
	case TabSecurity:
		content = m.security.View()
	}

	// Status bar
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		statusBar,
	)
}

func (m Model) renderHeader() string {
	title := TitleStyle.Render("🔐 PODX")
	version := MutedStyle.Render("v1.0.0")
	return lipgloss.JoinHorizontal(lipgloss.Left, title, " ", version)
}

func (m Model) renderTabs() string {
	tabs := []string{"Dashboard", "Commands", "Security"}
	var rendered []string

	for i, tab := range tabs {
		style := TabInactiveStyle
		if i == m.activeTab {
			style = TabActiveStyle
		}
		rendered = append(rendered, style.Render(fmt.Sprintf("[%d] %s", i+1, tab)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, rendered...)
}

func (m Model) renderStatusBar() string {
	help := "↑↓/jk: navigate  Enter/l: select  Tab: switch  ?: help  q: quit"
	return StatusBarStyle.Render(help)
}

func (m Model) renderHelp() string {
	help := `
PODX TUI Help

NAVIGATION
  ↑/k        Move up
  ↓/j        Move down
  ←/h        Go back
  →/l        Select / Enter
  Enter      Confirm action
  Tab        Next tab
  Shift+Tab  Previous tab
  1/2/3      Jump to tab

ACTIONS
  r          Refresh data
  ?          Toggle help
  q/Esc      Quit

Press any key to close help...
`
	return BoxStyle.Render(help)
}

func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

**Step 5: Verify it compiles**

```bash
go build ./tui
```

**Step 6: Commit**

```bash
git add go.mod go.sum tui/
git commit -m "feat(tui): add base TUI structure with styles and keybindings"
```

---

### Task 2: Create Dashboard Tab

**Files:**
- Create: `tui/dashboard.go`

**Step 1: Create `tui/dashboard.go`**

```go
package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hades/podx/project"
	"github.com/hades/podx/security"
)

type DashboardModel struct {
	project      *project.Project
	checkResult  *security.CheckResult
	loading      bool
	err          error
	selected     int
	actions      []string
}

type projectLoadedMsg struct {
	project *project.Project
	check   *security.CheckResult
	err     error
}

func NewDashboard() DashboardModel {
	return DashboardModel{
		loading: true,
		actions: []string{"Encrypt All", "Decrypt All", "Check", "Hook Install"},
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return loadProjectCmd()
}

func loadProjectCmd() tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			return projectLoadedMsg{err: err}
		}

		p, err := project.Load(cwd)
		if err != nil {
			return projectLoadedMsg{err: err}
		}

		check := security.CheckProject(cwd, false)
		return projectLoadedMsg{project: p, check: &check}
	}
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case projectLoadedMsg:
		m.loading = false
		m.project = msg.project
		m.checkResult = msg.check
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, Keys.Down):
			if m.selected < len(m.actions)-1 {
				m.selected++
			}
		case key.Matches(msg, Keys.Refresh):
			m.loading = true
			return m, loadProjectCmd()
		case key.Matches(msg, Keys.Enter), key.Matches(msg, Keys.Right):
			return m, m.executeAction()
		}
	}

	return m, nil
}

func (m DashboardModel) executeAction() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		switch m.selected {
		case 0: // Encrypt All
			p, _ := project.Load(cwd)
			if p != nil {
				p.EncryptAll()
			}
		case 1: // Decrypt All
			p, _ := project.Load(cwd)
			if p != nil {
				p.DecryptAll()
			}
		case 2: // Check
			security.CheckProject(cwd, false)
		case 3: // Hook Install
			security.InstallHook(cwd)
		}
		return loadProjectCmd()()
	}
}

func (m DashboardModel) View() string {
	if m.loading {
		return BoxStyle.Render("Loading...")
	}

	if m.err != nil {
		return m.renderNoProject()
	}

	// Build dashboard layout
	var sections []string

	// Project info
	sections = append(sections, m.renderProjectInfo())

	// Encryption status
	sections = append(sections, m.renderEncryptionStatus())

	// Recipients
	sections = append(sections, m.renderRecipients())

	// Quick actions
	sections = append(sections, m.renderQuickActions())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m DashboardModel) renderNoProject() string {
	content := `
No PODX project found in current directory.

Run 'podx init' to initialize a new project,
or navigate to a directory with .podx.yaml

Press 'i' to initialize here, or 'q' to quit.
`
	return BoxStyle.Width(60).Render(content)
}

func (m DashboardModel) renderProjectInfo() string {
	cwd, _ := os.Getwd()
	content := fmt.Sprintf(`📁 %s
🔐 Backend: %s`, cwd, m.project.Config.Backend)

	return BoxStyle.Width(35).Render(
		TitleStyle.Render("Project") + "\n" + content,
	)
}

func (m DashboardModel) renderEncryptionStatus() string {
	var status string
	if m.checkResult != nil && m.checkResult.Passed {
		status = SuccessStyle.Render("✅ All secrets encrypted")
	} else if m.checkResult != nil {
		issues := len(m.checkResult.EncryptionIssues)
		status = ErrorStyle.Render(fmt.Sprintf("❌ %d issue(s) found", issues))
	} else {
		status = MutedStyle.Render("Unknown")
	}

	return BoxStyle.Width(35).Render(
		TitleStyle.Render("Encryption") + "\n" + status,
	)
}

func (m DashboardModel) renderRecipients() string {
	var lines []string
	if m.project != nil {
		for _, r := range m.project.Config.Recipients {
			keyShort := r.Key
			if len(keyShort) > 20 {
				keyShort = keyShort[:20] + "..."
			}
			lines = append(lines, fmt.Sprintf("• %s (%s)", r.Name, keyShort))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, MutedStyle.Render("No recipients configured"))
	}

	return BoxStyle.Width(60).Render(
		TitleStyle.Render("Recipients") + "\n" + strings.Join(lines, "\n"),
	)
}

func (m DashboardModel) renderQuickActions() string {
	var lines []string
	for i, action := range m.actions {
		style := lipgloss.NewStyle()
		if i == m.selected {
			style = SelectedStyle
		}
		lines = append(lines, style.Render(fmt.Sprintf(" %s ", action)))
	}

	return BoxStyle.Width(60).Render(
		TitleStyle.Render("Quick Actions") + "\n" +
			lipgloss.JoinHorizontal(lipgloss.Left, lines...),
	)
}
```

**Step 2: Add missing import to tui.go**

Add `"github.com/charmbracelet/bubbles/key"` to imports in tui.go.

**Step 3: Verify it compiles**

```bash
go build ./tui
```

**Step 4: Commit**

```bash
git add tui/dashboard.go
git commit -m "feat(tui): add dashboard tab with project info and quick actions"
```

---

### Task 3: Create Commands Tab

**Files:**
- Create: `tui/commands.go`

**Step 1: Create `tui/commands.go`**

```go
package tui

import (
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type CommandItem struct {
	name string
	desc string
	args []string
}

func (i CommandItem) Title() string       { return i.name }
func (i CommandItem) Description() string { return i.desc }
func (i CommandItem) FilterValue() string { return i.name }

type CommandsModel struct {
	list     list.Model
	executed bool
	output   string
}

func NewCommands() CommandsModel {
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

	l := list.New(items, list.NewDefaultDelegate(), 60, 20)
	l.Title = "Commands"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	// Custom key bindings for vim navigation
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up", "k"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down", "j"))

	return CommandsModel{list: l}
}

func (m CommandsModel) Init() tea.Cmd {
	return nil
}

func (m CommandsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Enter), key.Matches(msg, Keys.Right):
			if item, ok := m.list.SelectedItem().(CommandItem); ok {
				return m, m.runCommand(item.args)
			}
		case key.Matches(msg, Keys.Left):
			m.executed = false
			m.output = ""
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type commandOutputMsg struct {
	output string
}

func (m CommandsModel) runCommand(args []string) tea.Cmd {
	return func() tea.Msg {
		exe, _ := os.Executable()
		cmd := exec.Command(exe, args...)
		output, _ := cmd.CombinedOutput()
		return commandOutputMsg{output: string(output)}
	}
}

func (m CommandsModel) View() string {
	if m.executed && m.output != "" {
		return BoxStyle.Width(60).Height(15).Render(
			TitleStyle.Render("Output") + "\n\n" +
				m.output + "\n\n" +
				MutedStyle.Render("Press ← or h to go back"),
		)
	}

	return m.list.View()
}
```

**Step 2: Verify it compiles**

```bash
go build ./tui
```

**Step 3: Commit**

```bash
git add tui/commands.go
git commit -m "feat(tui): add commands tab with interactive menu"
```

---

### Task 4: Create Security Tab

**Files:**
- Create: `tui/security.go`

**Step 1: Create `tui/security.go`**

```go
package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hades/podx/security"
)

type SecurityModel struct {
	result   *security.CheckResult
	loading  bool
	selected int
	actions  []string
}

type securityCheckMsg struct {
	result *security.CheckResult
}

func NewSecurity() SecurityModel {
	return SecurityModel{
		loading: true,
		actions: []string{"Refresh", "Fix Issues", "Install Hook"},
	}
}

func (m SecurityModel) Init() tea.Cmd {
	return runSecurityCheckCmd()
}

func runSecurityCheckCmd() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		result := security.CheckProject(cwd, false)
		return securityCheckMsg{result: &result}
	}
}

func (m SecurityModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case securityCheckMsg:
		m.loading = false
		m.result = msg.result
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Left):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, Keys.Right):
			if m.selected < len(m.actions)-1 {
				m.selected++
			}
		case key.Matches(msg, Keys.Refresh):
			m.loading = true
			return m, runSecurityCheckCmd()
		case key.Matches(msg, Keys.Enter):
			return m, m.executeAction()
		}
	}

	return m, nil
}

func (m SecurityModel) executeAction() tea.Cmd {
	return func() tea.Msg {
		cwd, _ := os.Getwd()
		switch m.selected {
		case 0: // Refresh
			// Already handled by Refresh key
		case 1: // Fix Issues
			security.CheckProject(cwd, true)
		case 2: // Install Hook
			security.InstallHook(cwd)
		}
		return runSecurityCheckCmd()()
	}
}

func (m SecurityModel) View() string {
	if m.loading {
		return BoxStyle.Render("Running security check...")
	}

	var sections []string

	// Header
	sections = append(sections, TitleStyle.Render("🔍 Security Check"))
	sections = append(sections, "")

	// Encryption status
	if len(m.result.EncryptionIssues) == 0 {
		sections = append(sections, SuccessStyle.Render("✅ Encryption Status    All secrets encrypted"))
	} else {
		sections = append(sections, ErrorStyle.Render("❌ Encryption Status"))
		for _, issue := range m.result.EncryptionIssues {
			sections = append(sections, "   "+issue)
		}
	}

	// Gitignore
	if len(m.result.GitignoreIssues) == 0 {
		sections = append(sections, SuccessStyle.Render("✅ Gitignore            Properly configured"))
	} else {
		sections = append(sections, WarningStyle.Render("⚠️  Gitignore Issues"))
		for _, pattern := range m.result.GitignoreIssues {
			sections = append(sections, "   Missing: "+pattern)
		}
	}

	// Pattern scan
	if len(m.result.SecretFindings) == 0 {
		sections = append(sections, SuccessStyle.Render("✅ Pattern Scan         No secrets detected"))
	} else {
		sections = append(sections, ErrorStyle.Render("❌ Sensitive Patterns Found"))
		for _, file := range m.result.SecretFindings {
			for _, match := range file.Matches {
				sections = append(sections, fmt.Sprintf("   %s:%d  %s", file.Path, match.Line, match.Content))
			}
		}
	}

	// Result
	sections = append(sections, "")
	sections = append(sections, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if m.result.Passed {
		sections = append(sections, SuccessStyle.Render("Result: PASSED"))
	} else {
		sections = append(sections, ErrorStyle.Render("Result: FAILED"))
	}

	// Actions
	sections = append(sections, "")
	sections = append(sections, m.renderActions())

	return BoxStyle.Width(60).Render(strings.Join(sections, "\n"))
}

func (m SecurityModel) renderActions() string {
	var actions []string
	for i, action := range m.actions {
		style := MutedStyle
		if i == m.selected {
			style = SelectedStyle
		}
		actions = append(actions, style.Render(fmt.Sprintf(" [%c] %s ", action[0], action)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, actions...)
}
```

**Step 2: Add lipgloss import**

Ensure `"github.com/charmbracelet/lipgloss"` is imported.

**Step 3: Verify it compiles**

```bash
go build ./tui
```

**Step 4: Commit**

```bash
git add tui/security.go
git commit -m "feat(tui): add security tab with check results and actions"
```

---

### Task 5: Integrate TUI into main.go

**Files:**
- Modify: `main.go`

**Step 1: Update main.go to launch TUI when no args**

Change the beginning of main():

```go
func main() {
	if len(os.Args) < 2 {
		// Launch TUI when no arguments
		if err := tui.Run(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]
	// ... rest of switch statement
}
```

**Step 2: Add import**

```go
"github.com/hades/podx/tui"
```

**Step 3: Test the TUI launches**

```bash
go run .
```

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat: launch TUI when podx runs without arguments"
```

---

### Task 6: Polish and Fix Issues

**Files:**
- Modify: `tui/tui.go`
- Modify: `tui/dashboard.go`
- Modify: `tui/commands.go`

**Step 1: Fix any compilation errors**

Review all files for missing imports, type mismatches, etc.

**Step 2: Add proper type assertions and error handling**

**Step 3: Test all navigation**

- Tab switching works
- vim keys (h,j,k,l) work
- Enter/select works
- Quit works

**Step 4: Commit**

```bash
git add tui/
git commit -m "fix(tui): polish navigation and fix issues"
```

---

### Task 7: Update Documentation

**Files:**
- Modify: `CLAUDE.md`
- Modify: `README.md` (optional)

**Step 1: Add TUI section to CLAUDE.md**

```markdown
## TUI (Terminal User Interface)

```bash
podx                          # Launch interactive TUI
```

The TUI provides:
- **Dashboard** - Project status, encryption status, quick actions
- **Commands** - Interactive menu to run any podx command
- **Security** - Live security check with detailed results

Navigation:
- `Tab` / `1,2,3` - Switch tabs
- `↑↓` / `j,k` - Navigate up/down
- `←→` / `h,l` - Back/Select
- `Enter` - Confirm
- `r` - Refresh
- `?` - Help
- `q` - Quit

The `tui/` package contains:
- `tui.go` - Main model and update loop
- `dashboard.go` - Dashboard tab
- `commands.go` - Commands tab
- `security.go` - Security tab
- `styles.go` - Lipgloss styles
- `keys.go` - Key bindings
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add TUI documentation to CLAUDE.md"
```

---

### Task 8: Run Full Test Suite

**Step 1: Run all tests**

```bash
go test ./... -v
```

**Step 2: Build and verify**

```bash
make build
./podx
```

**Step 3: Test TUI manually**

- Launch TUI
- Navigate between tabs
- Run a command
- View security results
- Quit

**Step 4: Final commit if needed**

```bash
git status
# If changes, commit them
```
