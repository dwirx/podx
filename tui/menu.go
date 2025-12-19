package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a menu item
type MenuItem struct {
	title       string
	description string
	action      string
	icon        string
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Description() string { return i.description }
func (i MenuItem) FilterValue() string { return i.title }

// menuItemDelegate renders menu items with custom styling
type menuItemDelegate struct{}

func (d menuItemDelegate) Height() int                             { return 2 }
func (d menuItemDelegate) Spacing() int                            { return 0 }
func (d menuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d menuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	var title, desc string
	if isSelected {
		title = SelectedItemStyle.Render("▸ " + i.title)
		desc = ItemDescStyle.Render("  " + i.description)
	} else {
		title = ItemStyle.Render("  " + i.title)
		desc = ItemDescStyle.Render("  " + i.description)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// MenuModel is the main menu
type MenuModel struct {
	list     list.Model
	quitting bool
	selected string
}

// NewMenuModel creates a new menu
func NewMenuModel() MenuModel {
	items := []list.Item{
		MenuItem{
			title:       "🔐 Encrypt File",
			description: "Choose algorithm → Select file",
			action:      "encrypt",
		},
		MenuItem{
			title:       "🔓 Decrypt File",
			description: "Decrypt any encrypted file",
			action:      "decrypt",
		},
		MenuItem{
			title:       "📄 Encrypt .env",
			description: "Format-preserving .env encryption",
			action:      "env-encrypt",
		},
		MenuItem{
			title:       "📂 Decrypt .env",
			description: "Decrypt .env file",
			action:      "env-decrypt",
		},
		MenuItem{
			title:       "🆕 Generate Keys",
			description: "Generate Age or GPG key pair",
			action:      "keygen",
		},
		MenuItem{
			title:       "⚙️  Project",
			description: "Init, status, encrypt-all, decrypt-all",
			action:      "project",
		},
		MenuItem{
			title:       "⬆️  Check Updates",
			description: "Check and install updates",
			action:      "update",
		},
		MenuItem{
			title:       "❌ Exit",
			description: "Quit PODX",
			action:      "exit",
		},
	}

	const listWidth = 55
	const listHeight = 16

	l := list.New(items, menuItemDelegate{}, listWidth, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle()

	return MenuModel{list: l}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width - 4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if item, ok := m.list.SelectedItem().(MenuItem); ok {
				m.selected = item.action
				if item.action == "exit" {
					m.quitting = true
					return m, tea.Quit
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MenuModel) View() string {
	if m.quitting {
		return SuccessStyle.Render("\n  👋 Goodbye!\n\n")
	}

	logo := LogoStyle.Render(`
  ██████╗  ██████╗ ██████╗ ██╗  ██╗
  ██╔══██╗██╔═══██╗██╔══██╗╚██╗██╔╝
  ██████╔╝██║   ██║██║  ██║ ╚███╔╝ 
  ██╔═══╝ ██║   ██║██║  ██║ ██╔██╗ 
  ██║     ╚██████╔╝██████╔╝██╔╝ ██╗
  ╚═╝      ╚═════╝ ╚═════╝ ╚═╝  ╚═╝`)

	subtitle := SubtitleStyle.Render("  🔐 Encryption Tool — Use ↑↓ to navigate, Enter to select")

	// Help bar
	help := []string{
		RenderHelp("↑/↓", "navigate"),
		RenderHelp("Enter", "select"),
		RenderHelp("q", "quit"),
	}
	helpBar := HelpStyle.Render("  " + strings.Join(help, HelpSeparator))

	var b strings.Builder
	b.WriteString(logo)
	b.WriteString("\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpBar)

	return b.String()
}

// Selected returns the selected action
func (m MenuModel) Selected() string {
	return m.selected
}
