package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// DashboardModel represents the dashboard tab content
type DashboardModel struct {
	width  int
	height int
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() DashboardModel {
	return DashboardModel{}
}

// Init initializes the dashboard model
func (m DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the dashboard model
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	return m, nil
}

// View renders the dashboard model
func (m DashboardModel) View() string {
	return BoxStyle.Render("Dashboard - Project status will appear here")
}

// SetSize updates the model dimensions
func (m *DashboardModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
