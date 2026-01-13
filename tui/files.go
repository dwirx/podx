package tui

import (
	"bufio"
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
	cwd            string
	files          []FileInfo
	selected       int
	width          int
	height         int
	keys           KeyMap
	loading        bool
	message        string
	msgStyle       lipgloss.Style
	filter         textinput.Model
	filtering      bool
	err            error
	offset         int // For scrolling
	project        *project.Project
	showPreview    bool
	previewContent []string
	selectedFiles  map[string]bool // for multi-select
	gotoInput      textinput.Model
	showGoto       bool
	encryptDialog  EncryptDialogModel
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

	gotoInput := textinput.New()
	gotoInput.Placeholder = "Enter path..."
	gotoInput.CharLimit = 256

	return FilesModel{
		cwd:           cwd,
		selected:      0,
		keys:          DefaultKeyMap(),
		loading:       true,
		filter:        filter,
		msgStyle:      SuccessStyle,
		showPreview:   true,
		selectedFiles: make(map[string]bool),
		gotoInput:     gotoInput,
		showGoto:      false,
		encryptDialog: NewEncryptDialogModel(),
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
	// Handle encrypt dialog first if visible
	if m.encryptDialog.IsVisible() {
		var cmd tea.Cmd
		m.encryptDialog, cmd = m.encryptDialog.Update(msg)

		// Check if dialog was closed with completion
		if !m.encryptDialog.IsVisible() && m.encryptDialog.successMsg != "" {
			m.message = m.encryptDialog.successMsg
			m.msgStyle = SuccessStyle
			m.selectedFiles = make(map[string]bool)
			return m, m.loadFiles
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case encryptCompleteMsg:
		m.loading = false
		if msg.success {
			m.message = msg.message
			m.msgStyle = SuccessStyle
			m.selectedFiles = make(map[string]bool)
		} else {
			m.message = msg.message
			m.msgStyle = ErrorStyle
		}
		return m, m.loadFiles

	case fileLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.files = msg.files
			m.selected = 0
			m.offset = 0
			m.updatePreview()
		}
		// Load project for encryption operations
		m.project, _ = project.Load(m.cwd)
		return m, nil

	case fileOperationMsg:
		m.loading = false
		m.message = msg.message
		if msg.success {
			m.msgStyle = SuccessStyle
			// Clear selections after successful operation
			m.selectedFiles = make(map[string]bool)
		} else {
			m.msgStyle = ErrorStyle
		}
		// Reload files after operation
		return m, m.loadFiles

	case tea.KeyMsg:
		// Handle goto input mode
		if m.showGoto {
			switch msg.String() {
			case "esc":
				m.showGoto = false
				m.gotoInput.SetValue("")
				m.gotoInput.Blur()
				return m, nil
			case "enter":
				path := m.gotoInput.Value()
				if path != "" {
					// Expand ~ to home directory
					if strings.HasPrefix(path, "~") {
						home, _ := os.UserHomeDir()
						path = filepath.Join(home, path[1:])
					}
					// Make path absolute if relative
					if !filepath.IsAbs(path) {
						path = filepath.Join(m.cwd, path)
					}
					// Check if path exists and is a directory
					info, err := os.Stat(path)
					if err == nil && info.IsDir() {
						m.cwd = path
						m.loading = true
						m.showGoto = false
						m.gotoInput.SetValue("")
						m.gotoInput.Blur()
						return m, m.loadFiles
					} else {
						m.message = "Invalid directory path"
						m.msgStyle = ErrorStyle
					}
				}
				m.showGoto = false
				m.gotoInput.SetValue("")
				m.gotoInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.gotoInput, cmd = m.gotoInput.Update(msg)
				return m, cmd
			}
		}

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
				m.updatePreview()
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			filteredFiles := m.getFilteredFiles()
			if m.selected < len(filteredFiles)-1 {
				m.selected++
				m.ensureVisible()
				m.updatePreview()
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

		case msg.String() == "backspace" && !m.filtering && !m.showGoto:
			// Go to parent directory
			if m.cwd != "/" {
				m.cwd = filepath.Dir(m.cwd)
				m.loading = true
				m.filter.SetValue("")
				return m, m.loadFiles
			}
			return m, nil

		case msg.String() == "e":
			// Encrypt selected file(s) - show dialog
			files := m.getFilesToOperate()
			if len(files) == 0 {
				m.message = "No file selected"
				m.msgStyle = ErrorStyle
				return m, nil
			}
			// Filter out already encrypted files
			var toEncrypt []FileInfo
			for _, f := range files {
				if !f.IsEncrypted && !f.IsDir && f.Name != ".." {
					toEncrypt = append(toEncrypt, f)
				}
			}
			if len(toEncrypt) == 0 {
				m.message = "Selected files are already encrypted"
				m.msgStyle = WarningStyle
				return m, nil
			}
			return m, m.encryptDialog.Show(toEncrypt, m.project, m.width, m.height)

		case msg.String() == "d":
			// Decrypt selected file(s) - show dialog
			files := m.getFilesToOperate()
			if len(files) == 0 {
				m.message = "No file selected"
				m.msgStyle = ErrorStyle
				return m, nil
			}
			// Filter for encrypted files only
			var toDecrypt []FileInfo
			for _, f := range files {
				if f.IsEncrypted {
					toDecrypt = append(toDecrypt, f)
				}
			}
			if len(toDecrypt) == 0 {
				m.message = "Selected files are not encrypted"
				m.msgStyle = WarningStyle
				return m, nil
			}
			return m, m.encryptDialog.ShowDecrypt(toDecrypt, m.project, m.width, m.height)

		case msg.String() == "/":
			// Start filtering
			m.filtering = true
			m.filter.Focus()
			return m, textinput.Blink

		case msg.String() == "g":
			// Open goto input
			m.showGoto = true
			m.gotoInput.SetValue("")
			m.gotoInput.Focus()
			return m, textinput.Blink

		case msg.String() == "p":
			// Toggle preview panel
			m.showPreview = !m.showPreview
			if m.showPreview {
				m.updatePreview()
			}
			return m, nil

		case msg.String() == " ":
			// Toggle file selection (multi-select)
			filteredFiles := m.getFilteredFiles()
			if m.selected < len(filteredFiles) {
				file := filteredFiles[m.selected]
				if !file.IsDir && file.Name != ".." {
					if m.selectedFiles[file.Path] {
						delete(m.selectedFiles, file.Path)
					} else {
						m.selectedFiles[file.Path] = true
					}
				}
			}
			return m, nil

		case msg.String() == "a":
			// Select/deselect all files
			filteredFiles := m.getFilteredFiles()
			allSelected := true
			for _, file := range filteredFiles {
				if !file.IsDir && file.Name != ".." && !m.selectedFiles[file.Path] {
					allSelected = false
					break
				}
			}
			if allSelected {
				// Deselect all
				m.selectedFiles = make(map[string]bool)
			} else {
				// Select all files (not directories)
				for _, file := range filteredFiles {
					if !file.IsDir && file.Name != ".." {
						m.selectedFiles[file.Path] = true
					}
				}
			}
			return m, nil

		case msg.String() == "r":
			// Refresh file list
			m.loading = true
			return m, m.loadFiles
		}
	}

	return m, nil
}

// updatePreview loads preview content for the selected file
func (m *FilesModel) updatePreview() {
	m.previewContent = nil

	filteredFiles := m.getFilteredFiles()
	if m.selected >= len(filteredFiles) {
		return
	}

	file := filteredFiles[m.selected]
	if file.IsDir || file.Name == ".." {
		m.previewContent = []string{"[Directory]"}
		return
	}

	// Check file size - don't preview large files
	if file.Size > 1024*1024 { // 1MB limit
		m.previewContent = []string{
			"[Large file - preview disabled]",
			"",
			fmt.Sprintf("Size: %s", formatSize(file.Size)),
			fmt.Sprintf("Modified: %s", file.ModTime.Format("2006-01-02 15:04")),
		}
		return
	}

	// Check if it's a binary file by extension
	ext := strings.ToLower(filepath.Ext(file.Name))
	binaryExts := map[string]bool{
		".exe": true, ".bin": true, ".so": true, ".dylib": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".ico": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".mkv": true,
		".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	}
	if binaryExts[ext] {
		m.previewContent = []string{
			"[Binary file]",
			"",
			fmt.Sprintf("Type: %s", getFileTypeDescription(ext)),
			fmt.Sprintf("Size: %s", formatSize(file.Size)),
			fmt.Sprintf("Modified: %s", file.ModTime.Format("2006-01-02 15:04")),
		}
		return
	}

	// Read first 10 lines of the file
	f, err := os.Open(file.Path)
	if err != nil {
		m.previewContent = []string{"[Cannot read file]", err.Error()}
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	maxLines := 15
	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		// Truncate long lines
		if len(line) > 40 {
			line = line[:37] + "..."
		}
		m.previewContent = append(m.previewContent, line)
		lineCount++
	}

	if lineCount == maxLines {
		m.previewContent = append(m.previewContent, "...")
	}
}

// getFileTypeDescription returns a human-readable description for file types
func getFileTypeDescription(ext string) string {
	types := map[string]string{
		".png": "PNG Image", ".jpg": "JPEG Image", ".jpeg": "JPEG Image",
		".gif": "GIF Image", ".bmp": "Bitmap Image", ".ico": "Icon",
		".mp3": "MP3 Audio", ".mp4": "MP4 Video", ".avi": "AVI Video",
		".mov": "QuickTime Video", ".mkv": "MKV Video",
		".zip": "ZIP Archive", ".tar": "TAR Archive", ".gz": "GZip Archive",
		".7z": "7-Zip Archive", ".rar": "RAR Archive",
		".pdf": "PDF Document", ".doc": "Word Document", ".docx": "Word Document",
		".xls": "Excel Spreadsheet", ".xlsx": "Excel Spreadsheet",
		".exe": "Executable", ".bin": "Binary", ".so": "Shared Library",
		".dylib": "Dynamic Library",
	}
	if desc, ok := types[ext]; ok {
		return desc
	}
	return "Binary file"
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
	return m.height - 8
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

// getSelectedCount returns the number of selected files
func (m FilesModel) getSelectedCount() int {
	return len(m.selectedFiles)
}

// getFilesToOperate returns files to encrypt/decrypt (selected files or current file)
func (m FilesModel) getFilesToOperate() []FileInfo {
	if len(m.selectedFiles) > 0 {
		var files []FileInfo
		for path := range m.selectedFiles {
			for _, f := range m.files {
				if f.Path == path {
					files = append(files, f)
					break
				}
			}
		}
		return files
	}

	// Fall back to currently selected file
	filteredFiles := m.getFilteredFiles()
	if m.selected < len(filteredFiles) {
		file := filteredFiles[m.selected]
		if !file.IsDir && file.Name != ".." {
			return []FileInfo{file}
		}
	}
	return nil
}

// encryptSelected encrypts the currently selected file(s)
func (m FilesModel) encryptSelected() tea.Cmd {
	return func() tea.Msg {
		files := m.getFilesToOperate()
		if len(files) == 0 {
			return fileOperationMsg{success: false, message: "No file selected"}
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

		successCount := 0
		var lastErr error

		for _, file := range files {
			if file.IsDir || file.Name == ".." {
				continue
			}
			if file.IsEncrypted {
				continue
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
				lastErr = err
				continue
			}

			// Delete original file after successful encryption
			if err := os.Remove(file.Path); err != nil {
				lastErr = err
			}
			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return fileOperationMsg{success: false, message: fmt.Sprintf("Encryption failed: %v", lastErr)}
		}

		if successCount == 1 {
			return fileOperationMsg{success: true, message: fmt.Sprintf("Encrypted %d file", successCount)}
		}
		return fileOperationMsg{success: true, message: fmt.Sprintf("Encrypted %d files", successCount)}
	}
}

// decryptSelected decrypts the currently selected file(s)
func (m FilesModel) decryptSelected() tea.Cmd {
	return func() tea.Msg {
		files := m.getFilesToOperate()
		if len(files) == 0 {
			return fileOperationMsg{success: false, message: "No file selected"}
		}

		if m.project == nil {
			return fileOperationMsg{success: false, message: "No PODX project found"}
		}

		successCount := 0
		var lastErr error

		for _, file := range files {
			if file.IsDir || file.Name == ".." {
				continue
			}
			if !file.IsEncrypted {
				continue
			}

			// Decrypt the file
			decPath := strings.TrimSuffix(file.Path, ".podx")
			err := m.project.DecryptFile(file.Path, decPath)

			if err != nil {
				lastErr = err
				continue
			}
			successCount++
		}

		if successCount == 0 && lastErr != nil {
			return fileOperationMsg{success: false, message: fmt.Sprintf("Decryption failed: %v", lastErr)}
		}

		if successCount == 1 {
			return fileOperationMsg{success: true, message: fmt.Sprintf("Decrypted %d file", successCount)}
		}
		return fileOperationMsg{success: true, message: fmt.Sprintf("Decrypted %d files", successCount)}
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

	// Overlay dialog if visible - use CenterDialog for proper fullscreen overlay
	if m.encryptDialog.IsVisible() {
		dialog := m.encryptDialog.View()
		return CenterDialog(dialog, m.width, m.height)
	}

	return m.renderFileBrowser()
}

// getFileIcon returns an appropriate icon for the file type
func getFileIcon(file FileInfo) string {
	if file.Name == ".." {
		return "📁"
	}
	if file.IsDir {
		return "📂"
	}
	if file.IsEncrypted {
		return "🔒"
	}

	ext := strings.ToLower(filepath.Ext(file.Name))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg", ".webp":
		return "🖼️"
	case ".zip", ".tar", ".gz", ".7z", ".rar", ".bz2", ".xz":
		return "📦"
	case ".go":
		return "🔷"
	case ".py":
		return "🐍"
	case ".js", ".ts":
		return "⚡"
	case ".rs":
		return "🦀"
	case ".c", ".cpp", ".h", ".hpp":
		return "⚙️"
	case ".java":
		return "☕"
	case ".rb":
		return "💎"
	case ".md":
		return "📝"
	case ".txt", ".doc", ".docx", ".pdf":
		return "📄"
	case ".json", ".yaml", ".yml", ".toml", ".xml", ".ini", ".conf":
		return "⚙️"
	case ".mp3", ".wav", ".ogg", ".flac":
		return "🎵"
	case ".mp4", ".avi", ".mov", ".mkv", ".webm":
		return "🎬"
	case ".sh", ".bash", ".zsh", ".fish":
		return "💻"
	case ".env":
		return "🔐"
	case ".podx":
		return "🔒"
	case ".html", ".css":
		return "🌐"
	default:
		return "📄"
	}
}

// getFileColor returns a color for the file type
func getFileColor(file FileInfo) lipgloss.Color {
	if file.IsDir {
		return ColorPrimary
	}
	if file.IsEncrypted {
		return ColorWarning
	}

	ext := strings.ToLower(filepath.Ext(file.Name))
	switch ext {
	case ".go":
		return lipgloss.Color("#00ADD8") // Go blue
	case ".py":
		return lipgloss.Color("#3776AB") // Python blue
	case ".js", ".ts":
		return lipgloss.Color("#F7DF1E") // JS yellow
	case ".rs":
		return lipgloss.Color("#DEA584") // Rust orange
	case ".md", ".txt":
		return ColorWhite
	case ".json", ".yaml", ".yml", ".toml":
		return ColorSuccess
	case ".png", ".jpg", ".jpeg", ".gif":
		return lipgloss.Color("#FF69B4") // Pink
	default:
		return ColorWhite
	}
}

// renderBreadcrumbs renders the path as ASCII breadcrumbs
func (m FilesModel) renderBreadcrumbs() string {
	parts := strings.Split(m.cwd, string(filepath.Separator))
	var crumbs []string

	crumbs = append(crumbs, MutedStyle.Render("/"))
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == len(parts)-1 {
			// Current directory - highlight it
			crumbs = append(crumbs, BreadcrumbActiveStyle.Render(part))
		} else {
			crumbs = append(crumbs, BreadcrumbStyle.Render(part))
		}
		if i < len(parts)-1 {
			crumbs = append(crumbs, MutedStyle.Render("/"))
		}
	}

	return strings.Join(crumbs, "")
}

// renderFileBrowser renders the full file browser view
func (m FilesModel) renderFileBrowser() string {
	// Get terminal size category
	termSize := GetTerminalSize(m.width)

	// Calculate widths based on terminal size
	totalWidth := m.width - 6 // Account for borders and padding
	if totalWidth < MinTerminalWidth-6 {
		totalWidth = MinTerminalWidth - 6
	}

	var fileListWidth int
	var previewWidth int

	switch termSize {
	case TerminalSmall:
		// Disable preview on small terminals
		fileListWidth = totalWidth
		previewWidth = 0
	case TerminalMedium:
		if m.showPreview {
			fileListWidth = (totalWidth * 65) / 100
			previewWidth = totalWidth - fileListWidth - 3
		} else {
			fileListWidth = totalWidth
		}
	default:
		// Large terminal
		if m.showPreview {
			fileListWidth = (totalWidth * 60) / 100
			previewWidth = totalWidth - fileListWidth - 3
			// Cap preview width for readability
			if previewWidth > 60 {
				previewWidth = 60
				fileListWidth = totalWidth - previewWidth - 3
			}
		} else {
			fileListWidth = totalWidth
		}
	}

	// Build the file list panel
	filePanel := m.renderFileListPanel(fileListWidth)

	// Build the preview panel if enabled and there's space
	var previewPanel string
	shouldShowPreview := m.showPreview && previewWidth > 15 && termSize != TerminalSmall
	if shouldShowPreview {
		previewPanel = m.renderPreviewPanel(previewWidth)
	}

	// Combine panels
	var mainContent string
	if shouldShowPreview {
		mainContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			filePanel,
			MutedStyle.Render(" │ "), // Better vertical separator
			previewPanel,
		)
	} else {
		mainContent = filePanel
	}

	return BoxStyle.Render(mainContent)
}

// renderFileListPanel renders the file list panel
func (m FilesModel) renderFileListPanel(width int) string {
	var lines []string

	// Header with breadcrumb path
	pathIcon := "[DIR]"
	header := TitleStyle.Render(fmt.Sprintf("%s %s", pathIcon, m.renderBreadcrumbs()))
	lines = append(lines, header)

	// Separator
	separator := strings.Repeat("-", min(width-4, 50))
	lines = append(lines, MutedStyle.Render(separator))

	// Goto input if active
	if m.showGoto {
		gotoLabel := TitleStyle.Render("Go to: ")
		lines = append(lines, gotoLabel+m.gotoInput.View())
		lines = append(lines, "")
	}

	// Filter input if active
	if m.filtering {
		lines = append(lines, fmt.Sprintf("Filter: %s", m.filter.View()))
	}

	// Selection count if files are selected
	if count := m.getSelectedCount(); count > 0 {
		selectionInfo := WarningStyle.Render(fmt.Sprintf("[%d selected]", count))
		lines = append(lines, selectionInfo)
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
		line := m.renderFileLine(file, i == m.selected, width)
		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for i := len(filteredFiles); i < visibleHeight && i-m.offset < visibleHeight; i++ {
		lines = append(lines, "")
	}

	// Separator
	lines = append(lines, MutedStyle.Render(separator))

	// Status message if present
	if m.message != "" {
		lines = append(lines, m.msgStyle.Render(m.message))
	}

	// Footer with keybindings - responsive based on width
	var footer string
	if m.width >= MediumWidth {
		footer = MutedStyle.Render("[j/k] Nav  [Enter] Open  [e] Encrypt  [d] Decrypt  [g] Goto  [/] Filter  [Space] Select  [p] Preview  [q] Back")
	} else if m.width >= SmallWidth {
		footer = MutedStyle.Render("[j/k] Nav [e/d] Enc/Dec [Space] Sel [p] Preview [q] Back")
	} else {
		footer = MutedStyle.Render("[j/k] [e/d] [Space] [q]")
	}
	lines = append(lines, footer)

	return strings.Join(lines, "\n")
}

// renderPreviewPanel renders the preview panel
func (m FilesModel) renderPreviewPanel(width int) string {
	var lines []string

	// Preview header
	header := CardTitleStyle.Render("Preview")
	lines = append(lines, header)

	// Separator
	separator := strings.Repeat("\u2500", min(width-2, 40))
	lines = append(lines, MutedStyle.Render(separator))

	// Preview content
	if len(m.previewContent) == 0 {
		lines = append(lines, MutedStyle.Render("[Select a file to preview]"))
	} else {
		for _, line := range m.previewContent {
			// Truncate lines if too long
			if len(line) > width-2 {
				line = line[:width-5] + "..."
			}
			lines = append(lines, line)
		}
	}

	// Pad to fill height
	visibleHeight := m.getVisibleHeight() + 3
	for len(lines) < visibleHeight {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderFileLine renders a single file line
func (m FilesModel) renderFileLine(file FileInfo, selected bool, maxWidth int) string {
	icon := getFileIcon(file)
	fileColor := getFileColor(file)

	var sizeStr string
	if file.Name == ".." {
		sizeStr = ""
	} else if file.IsDir {
		sizeStr = "<DIR>"
	} else {
		sizeStr = formatSize(file.Size)
	}

	// Truncate name if too long
	name := file.Name
	maxNameLen := maxWidth - 25
	if maxNameLen < 15 {
		maxNameLen = 15
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen-3] + "..."
	}

	// Add directory indicator
	if file.IsDir && file.Name != ".." {
		name = name + "/"
	}

	// Selection indicator
	var selectionMark string
	if m.selectedFiles[file.Path] {
		selectionMark = "[X] " // Checkmark
	} else {
		selectionMark = "[ ] "
	}

	// Cursor indicator
	prefix := "  "
	if selected {
		prefix = "> "
	}

	// Build line with padding for alignment
	line := fmt.Sprintf("%s%s%s %s", prefix, selectionMark, icon, name)

	// Add size aligned to the right
	padding := maxWidth - 12 - len(line)
	if padding < 2 {
		padding = 2
	}
	line = line + strings.Repeat(" ", padding) + sizeStr

	if selected {
		return SelectedStyle.Render(line)
	}

	// Apply file-specific color
	nameStyle := lipgloss.NewStyle().Foreground(fileColor)
	if m.selectedFiles[file.Path] {
		nameStyle = nameStyle.Bold(true)
	}

	coloredLine := fmt.Sprintf("%s%s%s %s", prefix, selectionMark, icon, nameStyle.Render(name))
	coloredLine = coloredLine + strings.Repeat(" ", padding) + MutedStyle.Render(sizeStr)

	return coloredLine
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
