package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileEntry represents a file or directory
type FileEntry struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

func (f FileEntry) Title() string {
	if f.IsDir {
		return "📁 " + f.Name
	}
	return "📄 " + f.Name
}

func (f FileEntry) Description() string {
	if f.IsDir {
		return "Directory"
	}
	return formatSize(f.Size)
}

func (f FileEntry) FilterValue() string { return f.Name }

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// FileBrowserModel is an advanced file browser
type FileBrowserModel struct {
	entries     []FileEntry
	filtered    []FileEntry
	cursor      int
	offset      int
	height      int
	currentDir  string
	searchQuery string
	selected    string
	extensions  []string // Filter by extensions
	excludeSufs []string // Hide files with these suffixes
	done        bool
	cancelled   bool
	showHidden  bool
	title       string
}

// NewFileBrowser creates a new file browser
func NewFileBrowser(title string, extensions []string) FileBrowserModel {
	cwd, err := os.Getwd()
	if err != nil {
		cwd, _ = os.UserHomeDir()
	}

	m := FileBrowserModel{
		currentDir: cwd,
		height:     12,
		extensions: extensions,
		title:      title,
	}
	m.loadEntries()
	return m
}

func (m *FileBrowserModel) loadEntries() {
	entries, err := os.ReadDir(m.currentDir)
	if err != nil {
		return
	}

	m.entries = nil

	// Add parent directory option if not at root
	if m.currentDir != "/" {
		m.entries = append(m.entries, FileEntry{
			Name:  "..",
			Path:  filepath.Dir(m.currentDir),
			IsDir: true,
		})
	}

	// Separate dirs and files
	var dirs, files []FileEntry

	for _, entry := range entries {
		name := entry.Name()

		isHidden := strings.HasPrefix(name, ".")
		if isHidden && !m.showHidden {
			if entry.IsDir() || len(m.extensions) == 0 || !m.matchesExtensions(name) {
				continue
			}
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fe := FileEntry{
			Name:  name,
			Path:  filepath.Join(m.currentDir, name),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}

		if fe.IsDir {
			dirs = append(dirs, fe)
		} else {
			if m.isExcluded(name) {
				continue
			}
			// Filter by extension if specified
			if !m.matchesExtensions(name) {
				continue
			}
			files = append(files, fe)
		}
	}

	// Sort alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	m.entries = append(m.entries, dirs...)
	m.entries = append(m.entries, files...)
	m.filterEntries()
}

func (m *FileBrowserModel) filterEntries() {
	if m.searchQuery == "" {
		m.filtered = m.entries
		return
	}

	query := strings.ToLower(m.searchQuery)
	m.filtered = nil
	for _, e := range m.entries {
		if strings.Contains(strings.ToLower(e.Name), query) {
			m.filtered = append(m.filtered, e)
		}
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
	}
}

func (m *FileBrowserModel) matchesExtensions(name string) bool {
	if len(m.extensions) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(name))
	nameLower := strings.ToLower(name)
	for _, e := range m.extensions {
		// Match exact extension or filename contains pattern (for .env files)
		if ext == e || strings.Contains(nameLower, e) {
			return true
		}
	}
	return false
}

func (m *FileBrowserModel) isExcluded(name string) bool {
	if len(m.excludeSufs) == 0 {
		return false
	}
	nameLower := strings.ToLower(name)
	for _, suf := range m.excludeSufs {
		if strings.HasSuffix(nameLower, strings.ToLower(suf)) {
			return true
		}
	}
	return false
}

func (m FileBrowserModel) Init() tea.Cmd {
	return nil
}

func (m FileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.height {
					m.offset = m.cursor - m.height + 1
				}
			}

		case "pgup":
			m.cursor -= m.height
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.offset = m.cursor

		case "pgdown":
			m.cursor += m.height
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			if m.cursor >= m.offset+m.height {
				m.offset = m.cursor - m.height + 1
			}

		case "home":
			m.cursor = 0
			m.offset = 0

		case "end":
			m.cursor = len(m.filtered) - 1
			if m.cursor >= m.height {
				m.offset = m.cursor - m.height + 1
			}

		case "h":
			m.showHidden = !m.showHidden
			m.loadEntries()

		case "enter":
			if len(m.filtered) == 0 {
				return m, nil
			}
			entry := m.filtered[m.cursor]
			if entry.IsDir {
				m.currentDir = entry.Path
				m.cursor = 0
				m.offset = 0
				m.searchQuery = ""
				m.loadEntries()
			} else {
				m.selected = entry.Path
				m.done = true
			}
			return m, nil

		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterEntries()
			}

		default:
			// Type to search
			if len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				m.filterEntries()
			}
		}
	}

	return m, nil
}

func (m FileBrowserModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Current path
	dir := m.currentDir
	if home, err := os.UserHomeDir(); err == nil {
		dir = strings.Replace(dir, home, "~", 1)
	}
	b.WriteString(InfoStyle.Render("📂 " + dir))
	b.WriteString("\n")

	// Search query
	if m.searchQuery != "" {
		b.WriteString(SubtitleStyle.Render("🔍 " + m.searchQuery))
	}
	b.WriteString("\n\n")

	// File list
	if len(m.filtered) == 0 {
		b.WriteString(TextDimStyle().Render("  No matching files"))
		b.WriteString("\n")
	} else {
		visible := m.filtered[m.offset:]
		if len(visible) > m.height {
			visible = visible[:m.height]
		}

		for i, entry := range visible {
			idx := m.offset + i
			icon := "  "
			style := FileItemStyle

			if entry.IsDir {
				style = DirItemStyle
			}

			if idx == m.cursor {
				icon = "▸ "
				style = SelectedItemStyle
			}

			line := icon + entry.Name
			if !entry.IsDir {
				line += " " + TextDimStyle().Render("("+formatSize(entry.Size)+")")
			}

			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	// Scrollbar indicator
	if len(m.filtered) > m.height {
		progress := float64(m.cursor) / float64(len(m.filtered)-1) * 100
		b.WriteString(fmt.Sprintf("\n%s %d/%d (%.0f%%)",
			TextDimStyle().Render("↕"),
			m.cursor+1,
			len(m.filtered),
			progress))
	}

	b.WriteString("\n\n")

	// Help
	help := []string{
		RenderHelp("↑↓", "navigate"),
		RenderHelp("Enter", "select"),
		RenderHelp("h", "hidden"),
		RenderHelp("Esc", "cancel"),
	}
	b.WriteString(strings.Join(help, HelpSeparator))

	return FocusedBoxStyle.Render(b.String())
}

// Helper style function
func TextDimStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(TextDim)
}

func (m *FileBrowserModel) SetExcludeSuffixes(suffixes []string) {
	m.excludeSufs = suffixes
	m.loadEntries()
}

// Getters
func (m FileBrowserModel) Selected() string   { return m.selected }
func (m FileBrowserModel) Done() bool         { return m.done }
func (m FileBrowserModel) Cancelled() bool    { return m.cancelled }
func (m FileBrowserModel) CurrentDir() string { return m.currentDir }
