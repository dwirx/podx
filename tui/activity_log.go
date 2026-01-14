package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogLevel represents the severity of a log entry
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogWarning
	LogError
)

// LogEntry represents a single activity log entry
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Details   string
}

// ActivityLog manages the global activity log
type ActivityLog struct {
	entries    []LogEntry
	maxEntries int
	mu         sync.RWMutex
}

// Global activity log instance
var globalLog = &ActivityLog{
	entries:    make([]LogEntry, 0, 100),
	maxEntries: 100,
}

// GetActivityLog returns the global activity log
func GetActivityLog() *ActivityLog {
	return globalLog
}

// Add adds a new entry to the activity log
func (l *ActivityLog) Add(level LogLevel, message string, details string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Details:   details,
	}

	l.entries = append(l.entries, entry)

	// Trim to max entries
	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[len(l.entries)-l.maxEntries:]
	}
}

// Info logs an info message
func (l *ActivityLog) Info(message string) {
	l.Add(LogInfo, message, "")
}

// Success logs a success message
func (l *ActivityLog) Success(message string) {
	l.Add(LogSuccess, message, "")
}

// Warning logs a warning message
func (l *ActivityLog) Warning(message string) {
	l.Add(LogWarning, message, "")
}

// Error logs an error message
func (l *ActivityLog) Error(message string) {
	l.Add(LogError, message, "")
}

// GetEntries returns the log entries (most recent first)
func (l *ActivityLog) GetEntries(limit int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}

	// Return in reverse order (newest first)
	result := make([]LogEntry, limit)
	for i := 0; i < limit; i++ {
		result[i] = l.entries[len(l.entries)-1-i]
	}
	return result
}

// Count returns the total number of entries
func (l *ActivityLog) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// Clear clears all log entries
func (l *ActivityLog) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]LogEntry, 0, l.maxEntries)
}

// ActivityLogModel is a bubbletea model for displaying the activity log
type ActivityLogModel struct {
	width    int
	height   int
	offset   int
	expanded bool
}

// NewActivityLogModel creates a new activity log model
func NewActivityLogModel() ActivityLogModel {
	return ActivityLogModel{
		expanded: false,
	}
}

// Init initializes the model
func (m ActivityLogModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ActivityLogModel) Update(msg tea.Msg) (ActivityLogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			entries := GetActivityLog().GetEntries(0)
			if m.offset < len(entries)-1 {
				m.offset++
			}
		case "g":
			m.offset = 0
		case "G":
			entries := GetActivityLog().GetEntries(0)
			if len(entries) > 0 {
				m.offset = len(entries) - 1
			}
		}
	}
	return m, nil
}

// View renders the activity log
func (m ActivityLogModel) View() string {
	cardWidth := m.width - 6
	if cardWidth > 100 {
		cardWidth = 100
	}
	if cardWidth < 50 {
		cardWidth = 50
	}

	var sections []string

	// Title card
	titleContent := TitleStyle.Render("Activity Log")
	sections = append(sections, CardStyle.Copy().Width(cardWidth).Render(titleContent))

	// Log entries card
	sections = append(sections, "")
	logContent := m.RenderLogCard(m.height-8, cardWidth)
	sections = append(sections, logContent)

	// Help card
	sections = append(sections, "")
	helpContent := MutedStyle.Render("  j/k: scroll | g: top | G: bottom")
	sections = append(sections, CardStyle.Copy().Width(cardWidth).Render(helpContent))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// RenderLogCard renders the log entries in a card
func (m ActivityLogModel) RenderLogCard(maxLines int, width int) string {
	if maxLines < 1 {
		maxLines = 10
	}

	var lines []string
	lines = append(lines, CardTitleStyle.Render("Recent Activity"))
	lines = append(lines, "")

	entries := GetActivityLog().GetEntries(maxLines)
	if len(entries) == 0 {
		lines = append(lines, MutedStyle.Render("  No activity yet"))
		lines = append(lines, "")
		lines = append(lines, MutedStyle.Render("  Activity will appear here as you use PODX"))
	} else {
		for _, entry := range entries {
			line := m.formatEntryWithWidth(entry, width-8)
			lines = append(lines, line)
		}
	}

	return CardStyle.Copy().Width(width).Render(strings.Join(lines, "\n"))
}

// formatEntryWithWidth formats a log entry with width constraint
func (m ActivityLogModel) formatEntryWithWidth(entry LogEntry, maxMsgWidth int) string {
	// Time format: HH:MM:SS
	timeStr := entry.Timestamp.Format("15:04:05")

	// Level indicator and style
	var levelIndicator string
	var msgStyle lipgloss.Style

	switch entry.Level {
	case LogSuccess:
		levelIndicator = "[+]"
		msgStyle = SuccessStyle
	case LogWarning:
		levelIndicator = "[!]"
		msgStyle = WarningStyle
	case LogError:
		levelIndicator = "[X]"
		msgStyle = ErrorStyle
	default:
		levelIndicator = "[*]"
		msgStyle = LogEntryStyle
	}

	// Truncate message if needed
	msg := entry.Message
	if len(msg) > maxMsgWidth-15 {
		msg = msg[:maxMsgWidth-18] + "..."
	}

	return fmt.Sprintf("  %s %s %s",
		LogTimestampStyle.Render(timeStr),
		msgStyle.Render(levelIndicator),
		msgStyle.Render(msg))
}

// SetSize updates dimensions
func (m *ActivityLogModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// LogActivityMsg is a message to add a log entry
type LogActivityMsg struct {
	Level   LogLevel
	Message string
	Details string
}

// LogActivity creates a command to log an activity
func LogActivity(level LogLevel, message string) tea.Cmd {
	return func() tea.Msg {
		GetActivityLog().Add(level, message, "")
		return LogActivityMsg{Level: level, Message: message}
	}
}
