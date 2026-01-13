package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchResultType identifies what type of result was found
type SearchResultType int

const (
	SearchResultFile SearchResultType = iota
	SearchResultCommand
	SearchResultAction
)

// SearchResult represents a search result item
type SearchResult struct {
	Type        SearchResultType
	Title       string
	Description string
	Path        string
	ActionID    string
}

// GlobalSearchModel manages the global search overlay
type GlobalSearchModel struct {
	input     textinput.Model
	results   []SearchResult
	selected  int
	visible   bool
	width     int
	height    int
	files     []FileInfo
	commands  []CommandItem
	searching bool
}

// NewGlobalSearchModel creates a new global search model
func NewGlobalSearchModel() GlobalSearchModel {
	input := textinput.New()
	input.Placeholder = "Search files, commands..."
	input.CharLimit = 100
	input.Width = 50

	return GlobalSearchModel{
		input:    input,
		results:  make([]SearchResult, 0),
		selected: 0,
	}
}

// Init initializes the model
func (m GlobalSearchModel) Init() tea.Cmd {
	return nil
}

// Show shows the search overlay
func (m *GlobalSearchModel) Show() tea.Cmd {
	m.visible = true
	m.input.SetValue("")
	m.input.Focus()
	m.results = nil
	m.selected = 0
	return textinput.Blink
}

// Hide hides the search overlay
func (m *GlobalSearchModel) Hide() {
	m.visible = false
	m.input.Blur()
	m.input.SetValue("")
	m.results = nil
}

// IsVisible returns whether search is visible
func (m GlobalSearchModel) IsVisible() bool {
	return m.visible
}

// Update handles messages
func (m GlobalSearchModel) Update(msg tea.Msg) (GlobalSearchModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			return m, nil
		case "enter":
			if len(m.results) > 0 && m.selected < len(m.results) {
				// Return selected result
				m.Hide()
				return m, m.handleSelection(m.results[m.selected])
			}
			return m, nil
		case "up", "ctrl+p":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "ctrl+n":
			if m.selected < len(m.results)-1 {
				m.selected++
			}
			return m, nil
		case "tab":
			// Cycle through results
			if len(m.results) > 0 {
				m.selected = (m.selected + 1) % len(m.results)
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.search(m.input.Value())
			return m, cmd
		}
	}

	return m, nil
}

// search performs the search
func (m *GlobalSearchModel) search(query string) {
	if query == "" {
		m.results = nil
		m.selected = 0
		return
	}

	query = strings.ToLower(query)
	m.results = make([]SearchResult, 0)

	// Search commands
	commands := GetAllCommands()
	for _, item := range commands {
		cmd := item.(CommandItem)
		if strings.Contains(strings.ToLower(cmd.name), query) ||
			strings.Contains(strings.ToLower(cmd.description), query) {
			m.results = append(m.results, SearchResult{
				Type:        SearchResultCommand,
				Title:       cmd.name,
				Description: cmd.description,
				ActionID:    cmd.id,
			})
		}
	}

	// Search files in current directory
	cwd, _ := os.Getwd()
	filepath.WalkDir(cwd, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Skip hidden directories and common ignores
		name := d.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.Contains(strings.ToLower(name), query) {
			relPath, _ := filepath.Rel(cwd, path)
			typeDesc := "File"
			if d.IsDir() {
				typeDesc = "Directory"
			}
			m.results = append(m.results, SearchResult{
				Type:        SearchResultFile,
				Title:       name,
				Description: typeDesc + ": " + relPath,
				Path:        path,
			})
		}

		// Limit results
		if len(m.results) >= 15 {
			return filepath.SkipAll
		}
		return nil
	})

	// Add quick actions
	quickActions := []struct {
		name string
		desc string
		id   string
	}{
		{"Encrypt All", "Encrypt all secret files", "encrypt-all"},
		{"Decrypt All", "Decrypt all secret files", "decrypt-all"},
		{"Run Check", "Run security checks", "check"},
		{"Generate Key", "Generate new Age key", "keygen-age"},
	}

	for _, action := range quickActions {
		if strings.Contains(strings.ToLower(action.name), query) ||
			strings.Contains(strings.ToLower(action.desc), query) {
			m.results = append(m.results, SearchResult{
				Type:        SearchResultAction,
				Title:       action.name,
				Description: action.desc,
				ActionID:    action.id,
			})
		}
	}

	// Reset selection if out of bounds
	if m.selected >= len(m.results) {
		m.selected = 0
	}
}

// handleSelection handles when a result is selected
func (m GlobalSearchModel) handleSelection(result SearchResult) tea.Cmd {
	switch result.Type {
	case SearchResultFile:
		// Return a message to navigate to the file
		return func() tea.Msg {
			return SearchNavigateMsg{Path: result.Path}
		}
	case SearchResultCommand, SearchResultAction:
		// Return a message to execute the command
		return func() tea.Msg {
			return SearchExecuteMsg{CommandID: result.ActionID}
		}
	}
	return nil
}

// View renders the search overlay
func (m GlobalSearchModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Header
	header := TitleStyle.Render("[ SEARCH ]")
	content.WriteString(header)
	content.WriteString("\n\n")

	// Search input
	content.WriteString("  > ")
	content.WriteString(m.input.View())
	content.WriteString("\n\n")

	// Divider
	content.WriteString(RenderHorizontalDivider(50))
	content.WriteString("\n\n")

	// Results
	if len(m.results) == 0 {
		if m.input.Value() != "" {
			content.WriteString(MutedStyle.Render("  No results found"))
		} else {
			content.WriteString(MutedStyle.Render("  Type to search..."))
		}
	} else {
		for i, result := range m.results {
			prefix := "   "
			if i == m.selected {
				prefix = " > "
			}

			// Type icon
			var icon string
			switch result.Type {
			case SearchResultFile:
				icon = "[F]"
			case SearchResultCommand:
				icon = "[C]"
			case SearchResultAction:
				icon = "[A]"
			}

			line := fmt.Sprintf("%s%s %s", prefix, icon, result.Title)

			if i == m.selected {
				content.WriteString(SelectedStyle.Render(line))
			} else {
				content.WriteString(line)
			}
			content.WriteString("\n")

			// Show description for selected item
			if i == m.selected && result.Description != "" {
				content.WriteString(MutedStyle.Render("       " + result.Description))
				content.WriteString("\n")
			}
		}
	}

	content.WriteString("\n")
	content.WriteString(RenderHorizontalDivider(50))
	content.WriteString("\n\n")

	// Help
	help := MutedStyle.Render("  [Enter] Select  [Tab] Next  [Esc] Close")
	content.WriteString(help)

	// Box it up
	dialogStyle := lipgloss.NewStyle().
		Border(TerminalBorder).
		BorderForeground(ColorSecondary).
		Background(ColorBg).
		Padding(1, 2).
		Width(60)

	dialog := dialogStyle.Render(content.String())

	return CenterDialog(dialog, m.width, m.height)
}

// SetSize updates dimensions
func (m *GlobalSearchModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SearchNavigateMsg tells the app to navigate to a path
type SearchNavigateMsg struct {
	Path string
}

// SearchExecuteMsg tells the app to execute a command
type SearchExecuteMsg struct {
	CommandID string
}

// GlobalSearchKeyMap defines search-related keys
type GlobalSearchKeyMap struct {
	Toggle key.Binding
}

// DefaultGlobalSearchKeyMap returns default search keys
func DefaultGlobalSearchKeyMap() GlobalSearchKeyMap {
	return GlobalSearchKeyMap{
		Toggle: key.NewBinding(
			key.WithKeys("ctrl+f", "ctrl+/"),
			key.WithHelp("Ctrl+F", "search"),
		),
	}
}

// Statistics holds project statistics
type Statistics struct {
	TotalFiles         int
	EncryptedFiles     int
	UnencryptedSecrets int
	LastEncryption     time.Time
	LastDecryption     time.Time
	LastCheck          time.Time
	ChecksPassed       int
	ChecksFailed       int
	Recipients         int
}

// GlobalStats holds global statistics
var GlobalStats = &Statistics{}

// UpdateStats updates the global statistics
func UpdateStats(s Statistics) {
	GlobalStats = &s
}

// RenderStatistics renders statistics in a compact format
func RenderStatistics(stats *Statistics, width int) string {
	var lines []string

	// Header
	lines = append(lines, CardTitleStyle.Render("STATISTICS"))
	lines = append(lines, "")

	// File stats
	lines = append(lines, fmt.Sprintf("  Files:      %s total, %s encrypted",
		StatValueStyle.Render(fmt.Sprintf("%d", stats.TotalFiles)),
		StatValueStyle.Render(fmt.Sprintf("%d", stats.EncryptedFiles))))

	// Security stats
	if stats.ChecksPassed > 0 || stats.ChecksFailed > 0 {
		passedStyle := SuccessStyle
		failedStyle := ErrorStyle
		lines = append(lines, fmt.Sprintf("  Checks:     %s passed, %s failed",
			passedStyle.Render(fmt.Sprintf("%d", stats.ChecksPassed)),
			failedStyle.Render(fmt.Sprintf("%d", stats.ChecksFailed))))
	}

	// Recipients
	lines = append(lines, fmt.Sprintf("  Recipients: %s",
		StatValueStyle.Render(fmt.Sprintf("%d", stats.Recipients))))

	// Last activity
	if !stats.LastEncryption.IsZero() {
		lines = append(lines, fmt.Sprintf("  Last Enc:   %s",
			MutedStyle.Render(stats.LastEncryption.Format("2006-01-02 15:04"))))
	}

	if !stats.LastCheck.IsZero() {
		lines = append(lines, fmt.Sprintf("  Last Check: %s",
			MutedStyle.Render(stats.LastCheck.Format("2006-01-02 15:04"))))
	}

	return strings.Join(lines, "\n")
}
