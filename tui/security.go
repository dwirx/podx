package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SecurityModel represents the security tab content
type SecurityModel struct {
	width  int
	height int
}

// NewSecurityModel creates a new security model
func NewSecurityModel() SecurityModel {
	return SecurityModel{}
}

// Init initializes the security model
func (m SecurityModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the security model
func (m SecurityModel) Update(msg tea.Msg) (SecurityModel, tea.Cmd) {
	return m, nil
}

// View renders the security model
func (m SecurityModel) View() string {
	return BoxStyle.Render("Security - Security status will appear here")
}

// SetSize updates the model dimensions
func (m *SecurityModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
