package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// CommandsModel represents the commands tab content
type CommandsModel struct {
	width  int
	height int
}

// NewCommandsModel creates a new commands model
func NewCommandsModel() CommandsModel {
	return CommandsModel{}
}

// Init initializes the commands model
func (m CommandsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the commands model
func (m CommandsModel) Update(msg tea.Msg) (CommandsModel, tea.Cmd) {
	return m, nil
}

// View renders the commands model
func (m CommandsModel) View() string {
	return BoxStyle.Render("Commands - Available commands will appear here")
}

// SetSize updates the model dimensions
func (m *CommandsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
