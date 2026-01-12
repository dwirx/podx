package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hades/podx/project"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name        string
	Path        string
	IsDir       bool
	Size        int64
	ModTime     time.Time
	IsEncrypted bool
}

// FilesModel represents the files tab content
type FilesModel struct {
	cwd       string
	files     []FileInfo
	selected  int
	width     int
	height    int
	keys      KeyMap
	loading   bool
	message   string
	msgStyle  lipgloss.Style
	filter    textinput.Model
	filtering bool
	err       error
	offset    int // For scrolling
	project   *project.Project
}

// fileLoadedMsg is sent when file list is loaded
type fileLoadedMsg struct {
	files []FileInfo
	err   error
}

// fileOperationMsg is sent when a file operation completes
type fileOperationMsg struct {
	success bool
	message string
}

// NewFilesModel creates a new files model
func NewFilesModel() FilesModel {
	cwd, _ := os.Getwd()
	filter := textinput.New()
	filter.Placeholder = "Type to filter..."
	filter.CharLimit = 100

	return FilesModel{
		cwd:      cwd,
		selected: 0,
		keys:     DefaultKeyMap(),
		loading:  true,
		filter:   filter,
		msgStyle: SuccessStyle,
	}
}

// Init initializes the files model
func (m FilesModel) Init() tea.Cmd {
	return m.loadFiles
}

// loadFiles loads the file list from the current directory
func (m FilesModel) loadFiles() tea.Msg {
	files, err := readDirectory(m.cwd)
	return fileLoadedMsg{files: files, err: err}
}

// readDirectory reads the contents of a directory
func readDirectory(dir string) ([]FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []FileInfo

	// Add parent directory entry if not at root
	if dir != "/" {
		files = append(files, FileInfo{
			Name:  "..",
			Path:  filepath.Dir(dir),
			IsDir: true,
		})
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		isEncrypted := strings.HasSuffix(entry.Name(), ".podx")

		files = append(files, FileInfo{
			Name:        entry.Name(),
			Path:        path,
			IsDir:       entry.IsDir(),
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			IsEncrypted: isEncrypted,
		})
	}

	// Sort: directories first, then files, alphabetically
	sort.Slice(files, func(i, j int) bool {
		// Parent directory always first
		if files[i].Name == ".." {
			return true
		}
		if files[j].Name == ".." {
			return false
		}
		// Directories before files
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		// Alphabetical within same type
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return files, nil
}

// Update handles messages for the files model
func (m FilesModel) Update(msg tea.Msg) (FilesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case fileLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.files = msg.files
			m.selected = 0
			m.offset = 0
		}
		// Load project for encryption operations
		m.project, _ = project.Load(m.cwd)
		return m, nil

	case fileOperationMsg:
		m.loading = false
		m.message = msg.message
		if msg.success {
			m.msgStyle = SuccessStyle
		} else {
			m.msgStyle = ErrorStyle
		}
		// Reload files after operation
		return m, m.loadFiles

	case tea.KeyMsg:
		// Handle filter input mode
		if m.filtering {
			switch msg.String() {
			case "esc":
				m.filtering = false
				m.filter.SetValue("")
				m.filter.Blur()
				return m, nil
			case "enter":
				m.filtering = false
				m.filter.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.filter, cmd = m.filter.Update(msg)
				return m, cmd
			}
		}

		// Normal mode key handling
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.selected > 0 {
				m.selected--
				m.ensureVisible()
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			filteredFiles := m.getFilteredFiles()
			if m.selected < len(filteredFiles)-1 {
				m.selected++
				m.ensureVisible()
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter), key.Matches(msg, m.keys.Right):
			filteredFiles := m.getFilteredFiles()
			if m.selected < len(filteredFiles) {
				file := filteredFiles[m.selected]
				if file.IsDir {
					m.cwd = file.Path
					m.loading = true
					m.filter.SetValue("")
					return m, m.loadFiles
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Left):
			// Go to parent directory
			if m.cwd != "/" {
				m.cwd = filepath.Dir(m.cwd)
				m.loading = true
				m.filter.SetValue("")
				return m, m.loadFiles
			}
			return m, nil

		case msg.String() == "backspace" && !m.filtering:
			// Go to parent directory
			if m.cwd != "/" {
				m.cwd = filepath.Dir(m.cwd)
				m.loading = true
				m.filter.SetValue("")
				return m, m.loadFiles
			}
			return m, nil

		case msg.String() == "e":
			// Encrypt selected file
			return m, m.encryptSelected()

		case msg.String() == "d":
			// Decrypt selected file
			return m, m.decryptSelected()

		case msg.String() == "/":
			// Start filtering
			m.filtering = true
			m.filter.Focus()
			return m, textinput.Blink
		}
	}

	return m, nil
}

// ensureVisible adjusts the scroll offset to keep the selected item visible
func (m *FilesModel) ensureVisible() {
	visibleHeight := m.getVisibleHeight()
	if visibleHeight <= 0 {
		return
	}

	if m.selected < m.offset {
		m.offset = m.selected
	} else if m.selected >= m.offset+visibleHeight {
		m.offset = m.selected - visibleHeight + 1
	}
}

// getVisibleHeight returns the number of visible file rows
func (m *FilesModel) getVisibleHeight() int {
	// Account for header, separator, footer, and some padding
	return m.height - 6
}

// getFilteredFiles returns files matching the current filter
func (m FilesModel) getFilteredFiles() []FileInfo {
	filterVal := strings.ToLower(m.filter.Value())
	if filterVal == "" {
		return m.files
	}

	var filtered []FileInfo
	for _, file := range m.files {
		if strings.Contains(strings.ToLower(file.Name), filterVal) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// encryptSelected encrypts the currently selected file
func (m FilesModel) encryptSelected() tea.Cmd {
	return func() tea.Msg {
		filteredFiles := m.getFilteredFiles()
		if m.selected >= len(filteredFiles) {
			return fileOperationMsg{success: false, message: "No file selected"}
		}

		file := filteredFiles[m.selected]
		if file.IsDir || file.Name == ".." {
			return fileOperationMsg{success: false, message: "Cannot encrypt a directory"}
		}
		if file.IsEncrypted {
			return fileOperationMsg{success: false, message: "File is already encrypted"}
		}

		if m.project == nil {
			return fileOperationMsg{success: false, message: "No PODX project found. Run 'podx init' first"}
		}

		if len(m.project.Config.Recipients) == 0 {
			return fileOperationMsg{success: false, message: "No recipients configured. Add with 'podx add-recipient'"}
		}

		// Get recipient keys
		var recipientKeys []string
		for _, r := range m.project.Config.Recipients {
			recipientKeys = append(recipientKeys, r.Key)
		}

		// Encrypt based on file type
		baseName := filepath.Base(file.Path)
		var err error
		if strings.HasPrefix(baseName, ".env") || strings.HasSuffix(baseName, ".env") {
			err = m.project.EncryptEnvFile(file.Path, recipientKeys)
		} else {
			err = m.project.EncryptRegularFile(file.Path, recipientKeys)
		}

		if err != nil {
			return fileOperationMsg{success: false, message: fmt.Sprintf("Encryption failed: %v", err)}
		}

		// Delete original file after successful encryption
		if err := os.Remove(file.Path); err != nil {
			return fileOperationMsg{
				success: true,
				message: fmt.Sprintf("Encrypted %s (could not delete original: %v)", file.Name, err),
			}
		}

		return fileOperationMsg{success: true, message: fmt.Sprintf("Encrypted %s", file.Name)}
	}
}

// decryptSelected decrypts the currently selected file
func (m FilesModel) decryptSelected() tea.Cmd {
	return func() tea.Msg {
		filteredFiles := m.getFilteredFiles()
		if m.selected >= len(filteredFiles) {
			return fileOperationMsg{success: false, message: "No file selected"}
		}

		file := filteredFiles[m.selected]
		if file.IsDir || file.Name == ".." {
			return fileOperationMsg{success: false, message: "Cannot decrypt a directory"}
		}
		if !file.IsEncrypted {
			return fileOperationMsg{success: false, message: "File is not encrypted (.podx)"}
		}

		if m.project == nil {
			return fileOperationMsg{success: false, message: "No PODX project found"}
		}

		// Decrypt the file
		decPath := strings.TrimSuffix(file.Path, ".podx")
		err := m.project.DecryptFile(file.Path, decPath)

		if err != nil {
			return fileOperationMsg{success: false, message: fmt.Sprintf("Decryption failed: %v", err)}
		}

		return fileOperationMsg{success: true, message: fmt.Sprintf("Decrypted %s", file.Name)}
	}
}

// View renders the files model
func (m FilesModel) View() string {
	if m.loading {
		return BoxStyle.Render("Loading files...")
	}

	if m.err != nil {
		return BoxStyle.Render(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return m.renderFileBrowser()
}

// renderFileBrowser renders the full file browser view
func (m FilesModel) renderFileBrowser() string {
	var lines []string

	// Header with current path
	headerIcon := lipgloss.NewStyle().SetString("\U0001F4C1").String() // 📁
	header := TitleStyle.Render(fmt.Sprintf("%s Files: %s", headerIcon, m.cwd))
	lines = append(lines, header)

	// Separator
	separator := strings.Repeat("\u2501", min(m.width-4, 60)) // ━
	lines = append(lines, MutedStyle.Render(separator))

	// Filter input if active
	if m.filtering {
		lines = append(lines, fmt.Sprintf("Filter: %s", m.filter.View()))
	}

	// File list
	filteredFiles := m.getFilteredFiles()
	visibleHeight := m.getVisibleHeight()
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	endIdx := m.offset + visibleHeight
	if endIdx > len(filteredFiles) {
		endIdx = len(filteredFiles)
	}

	for i := m.offset; i < endIdx; i++ {
		file := filteredFiles[i]
		line := m.renderFileLine(file, i == m.selected)
		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for i := len(filteredFiles); i < visibleHeight; i++ {
		lines = append(lines, "")
	}

	// Separator
	lines = append(lines, MutedStyle.Render(separator))

	// Status message if present
	if m.message != "" {
		lines = append(lines, m.msgStyle.Render(m.message))
	}

	// Footer with keybindings
	footer := MutedStyle.Render("[e] Encrypt  [d] Decrypt  [/] Filter  [h] Back  [Enter] Open")
	lines = append(lines, footer)

	return BoxStyle.Render(strings.Join(lines, "\n"))
}

// renderFileLine renders a single file line
func (m FilesModel) renderFileLine(file FileInfo, selected bool) string {
	var icon string
	var sizeStr string

	if file.Name == ".." {
		icon = "\U0001F4C1" // 📁
		sizeStr = ""
	} else if file.IsDir {
		icon = "\U0001F4C1" // 📁
		sizeStr = "<DIR>"
	} else if file.IsEncrypted {
		icon = "\U0001F510" // 🔐
		sizeStr = formatSize(file.Size)
	} else {
		icon = "\U0001F4C4" // 📄
		sizeStr = formatSize(file.Size)
	}

	// Truncate name if too long
	name := file.Name
	maxNameLen := m.width - 30
	if maxNameLen < 20 {
		maxNameLen = 20
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen-3] + "..."
	}

	// Add directory indicator
	if file.IsDir && file.Name != ".." {
		name = name + "/"
	}

	// Format the line
	prefix := "  "
	if selected {
		prefix = "> "
	}

	// Build line with padding for alignment
	line := fmt.Sprintf("%s%s %s", prefix, icon, name)

	// Add size aligned to the right
	padding := m.width - 20 - len(line)
	if padding < 2 {
		padding = 2
	}
	line = line + strings.Repeat(" ", padding) + sizeStr

	if selected {
		return SelectedStyle.Render(line)
	}
	return line
}

// formatSize formats a file size in human-readable format
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SetSize updates the model dimensions
func (m *FilesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
