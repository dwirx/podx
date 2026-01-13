package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastType defines the type of toast notification
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// Toast represents a notification toast
type Toast struct {
	ID        int
	Type      ToastType
	Title     string
	Message   string
	CreatedAt time.Time
	Duration  time.Duration
}

// ToastManager manages toast notifications
type ToastManager struct {
	toasts     []Toast
	nextID     int
	width      int
	maxToasts  int
	defaultTTL time.Duration
}

// NewToastManager creates a new toast manager
func NewToastManager() *ToastManager {
	return &ToastManager{
		toasts:     make([]Toast, 0),
		nextID:     1,
		maxToasts:  3,
		defaultTTL: 3 * time.Second,
	}
}

// Add adds a new toast
func (tm *ToastManager) Add(toastType ToastType, title, message string) int {
	toast := Toast{
		ID:        tm.nextID,
		Type:      toastType,
		Title:     title,
		Message:   message,
		CreatedAt: time.Now(),
		Duration:  tm.defaultTTL,
	}
	tm.nextID++

	tm.toasts = append(tm.toasts, toast)

	// Keep only max toasts
	if len(tm.toasts) > tm.maxToasts {
		tm.toasts = tm.toasts[len(tm.toasts)-tm.maxToasts:]
	}

	return toast.ID
}

// Info adds an info toast
func (tm *ToastManager) Info(title, message string) int {
	return tm.Add(ToastInfo, title, message)
}

// Success adds a success toast
func (tm *ToastManager) Success(title, message string) int {
	return tm.Add(ToastSuccess, title, message)
}

// Warning adds a warning toast
func (tm *ToastManager) Warning(title, message string) int {
	return tm.Add(ToastWarning, title, message)
}

// Error adds an error toast
func (tm *ToastManager) Error(title, message string) int {
	return tm.Add(ToastError, title, message)
}

// Remove removes a toast by ID
func (tm *ToastManager) Remove(id int) {
	for i, t := range tm.toasts {
		if t.ID == id {
			tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
			return
		}
	}
}

// Cleanup removes expired toasts
func (tm *ToastManager) Cleanup() {
	now := time.Now()
	active := make([]Toast, 0)
	for _, t := range tm.toasts {
		if now.Sub(t.CreatedAt) < t.Duration {
			active = append(active, t)
		}
	}
	tm.toasts = active
}

// GetActive returns active toasts
func (tm *ToastManager) GetActive() []Toast {
	tm.Cleanup()
	return tm.toasts
}

// HasToasts returns true if there are active toasts
func (tm *ToastManager) HasToasts() bool {
	tm.Cleanup()
	return len(tm.toasts) > 0
}

// SetWidth sets the width for rendering
func (tm *ToastManager) SetWidth(width int) {
	tm.width = width
}

// View renders all active toasts
func (tm *ToastManager) View() string {
	toasts := tm.GetActive()
	if len(toasts) == 0 {
		return ""
	}

	var views []string
	for _, t := range toasts {
		views = append(views, tm.renderToast(t))
	}

	return lipgloss.JoinVertical(lipgloss.Right, views...)
}

// renderToast renders a single toast
func (tm *ToastManager) renderToast(t Toast) string {
	var style lipgloss.Style
	var icon string

	switch t.Type {
	case ToastSuccess:
		style = ToastSuccessStyle
		icon = "[+]"
	case ToastWarning:
		style = lipgloss.NewStyle().
			Border(TerminalBorder).
			BorderForeground(ColorWarning).
			Background(ColorBg).
			Padding(0, 2)
		icon = "[!]"
	case ToastError:
		style = ToastErrorStyle
		icon = "[X]"
	default:
		style = ToastStyle
		icon = "[*]"
	}

	width := tm.width
	if width < 30 {
		width = 40
	}
	if width > 60 {
		width = 60
	}

	content := fmt.Sprintf("%s %s", icon, t.Title)
	if t.Message != "" {
		content += "\n   " + t.Message
	}

	return style.Width(width).Render(content)
}

// ToastTickMsg is sent to update toast visibility
type ToastTickMsg struct{}

// ToastTick returns a command that ticks the toast manager
func ToastTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return ToastTickMsg{}
	})
}

// ShowToastMsg is a message to show a toast
type ShowToastMsg struct {
	Type    ToastType
	Title   string
	Message string
}

// ShowToast creates a command to show a toast
func ShowToast(toastType ToastType, title, message string) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{
			Type:    toastType,
			Title:   title,
			Message: message,
		}
	}
}
