# TUI Responsive Design Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the TUI responsive and compatible with different terminal sizes (small, medium, large) for optimal user experience on any terminal.

**Architecture:** Add responsive breakpoints, dynamic layout calculations, and adaptive content rendering based on terminal dimensions. Implement minimum size validation and graceful degradation for very small terminals.

**Tech Stack:** Go, Bubble Tea (bubbletea), Lipgloss

---

## Task 1: Add Responsive Constants and Helpers

**Files:**
- Modify: `tui/styles.go`

**Step 1: Add responsive breakpoints and helpers**

Add constants for terminal size breakpoints and helper functions for responsive calculations at the end of `tui/styles.go`:

```go
// Terminal size breakpoints
const (
	MinTerminalWidth  = 60
	MinTerminalHeight = 20
	SmallWidth        = 80
	MediumWidth       = 120
	LargeWidth        = 160
)

// TerminalSize represents the terminal size category
type TerminalSize int

const (
	TerminalSmall TerminalSize = iota
	TerminalMedium
	TerminalLarge
)

// GetTerminalSize returns the terminal size category based on width
func GetTerminalSize(width int) TerminalSize {
	if width < SmallWidth {
		return TerminalSmall
	}
	if width < MediumWidth {
		return TerminalMedium
	}
	return TerminalLarge
}

// ResponsiveWidth calculates width based on percentage and constraints
func ResponsiveWidth(totalWidth, percentage, minWidth, maxWidth int) int {
	width := (totalWidth * percentage) / 100
	if width < minWidth {
		return minWidth
	}
	if maxWidth > 0 && width > maxWidth {
		return maxWidth
	}
	return width
}

// TruncateText truncates text to fit within width, adding ellipsis if needed
func TruncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}

// AdaptiveColumns returns the number of columns based on terminal width
func AdaptiveColumns(width int) int {
	size := GetTerminalSize(width)
	switch size {
	case TerminalSmall:
		return 1
	case TerminalMedium:
		return 2
	default:
		return 3
	}
}
```

**Step 2: Build and verify**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add tui/styles.go
git commit -m "feat(tui): add responsive breakpoints and helper functions

Add terminal size categories (Small/Medium/Large), responsive width
calculation, text truncation, and adaptive column helpers.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: Improve Dashboard Responsive Layout

**Files:**
- Modify: `tui/dashboard.go`

**Step 1: Update renderDashboard() to be fully responsive**

Find the `renderDashboard()` function (around line 255) and update it:

```go
// renderDashboard renders the full dashboard view
func (m DashboardModel) renderDashboard() string {
	var sections []string

	// Get terminal size category
	termSize := GetTerminalSize(m.width)

	// Calculate responsive card widths based on terminal size
	var cardWidth int
	switch termSize {
	case TerminalSmall:
		// Single column layout for small terminals
		cardWidth = m.width - 6
		if cardWidth < 30 {
			cardWidth = 30
		}
	case TerminalMedium:
		// Two column layout
		cardWidth = (m.width - 8) / 2
		if cardWidth < 35 {
			cardWidth = 35
		}
	default:
		// Large terminal - comfortable card widths
		cardWidth = (m.width - 10) / 2
		if cardWidth > 60 {
			cardWidth = 60
		}
	}

	// Update notification at top if available
	if m.updateInfo != nil && m.updateInfo.Available {
		updateWidth := m.width - 6
		if updateWidth > 100 {
			updateWidth = 100
		}
		updateCard := m.renderUpdateNotification(updateWidth)
		sections = append(sections, updateCard)
	}

	// Create horizontal layout with project info and security status
	projectCard := m.renderProjectInfoWithWidth(cardWidth)
	securityCard := m.renderSecurityStatusWithWidth(cardWidth)

	// Layout based on terminal size
	if termSize == TerminalSmall {
		// Stack vertically on small terminals
		sections = append(sections, projectCard)
		sections = append(sections, securityCard)
	} else {
		// Side by side on medium/large terminals
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, projectCard, securityCard)
		sections = append(sections, topRow)
	}

	// Quick actions - always full width but constrained
	actionsWidth := m.width - 6
	if actionsWidth > 100 {
		actionsWidth = 100
	}
	actionsCard := m.renderQuickActionsWithWidth(actionsWidth)
	sections = append(sections, actionsCard)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
```

**Step 2: Update renderProjectInfoWithWidth() for better text handling**

Find `renderProjectInfoWithWidth()` and update the path truncation:

```go
// Path - truncate based on available width
path := m.project.RootDir
maxPathLen := width - 12 // Account for "Path: " and padding
if maxPathLen < 20 {
	maxPathLen = 20
}
path = TruncateText(path, maxPathLen)
```

**Step 3: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add tui/dashboard.go
git commit -m "feat(tui): improve dashboard responsive layout

Use terminal size categories for adaptive card layout:
- Small terminals: single column stacked layout
- Medium terminals: two column side by side
- Large terminals: comfortable card widths with max constraints

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: Improve Files Tab Responsive Layout

**Files:**
- Modify: `tui/files.go`

**Step 1: Update renderFileBrowser() for responsive panels**

Find `renderFileBrowser()` (around line 843) and update:

```go
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
		// Force disable preview on small screens
		if m.showPreview && m.width < SmallWidth {
			// Still use full width, preview will be hidden
			fileListWidth = totalWidth
		}
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
```

**Step 2: Update footer keybindings to be responsive**

In `renderFileListPanel()`, update the footer section:

```go
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
```

**Step 3: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add tui/files.go
git commit -m "feat(tui): improve files tab responsive layout

- Auto-hide preview panel on small terminals
- Responsive panel width calculations
- Adaptive keybinding hints based on available width
- Better vertical separator for panels

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: Improve Encrypt Dialog Responsive Layout

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Make dialog width responsive**

Find the `View()` function (around line 830) and update the dialog style:

```go
// Dialog box styling with responsive width
dialogWidth := 60
if m.width > 0 {
	// Calculate responsive dialog width
	dialogWidth = ResponsiveWidth(m.width, 70, 50, 80)
}

dialogStyle := lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorPrimary).
	Background(ColorBg).
	Padding(1, 2).
	Width(dialogWidth)

return dialogStyle.Render(content)
```

**Step 2: Update renderAgeKeyConfirm() for responsive file list**

In `renderAgeKeyConfirm()`, update the file list display:

```go
// File info with responsive display
s.WriteString(lipgloss.NewStyle().Bold(true).Render("Files:"))
s.WriteString("\n")

// Show fewer files on smaller terminals
maxFilesToShow := 3
if m.width < SmallWidth {
	maxFilesToShow = 2
}

for i, file := range m.files {
	if i >= maxFilesToShow {
		s.WriteString(MutedStyle.Render(fmt.Sprintf("  ... and %d more\n", len(m.files)-maxFilesToShow)))
		break
	}
	fileIcon := "📄"
	if file.IsEncrypted {
		fileIcon = "🔒"
	}
	// Truncate filename if needed
	fileName := file.Name
	maxNameLen := 40
	if m.width < SmallWidth {
		maxNameLen = 25
	}
	fileName = TruncateText(fileName, maxNameLen)
	s.WriteString(fmt.Sprintf("  %s %s\n", fileIcon, fileName))
}
```

**Step 3: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): improve encrypt dialog responsive layout

- Responsive dialog width (50-80 chars based on terminal)
- Adaptive file list display for small terminals
- Text truncation for long filenames

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 5: Add Minimum Size Warning

**Files:**
- Modify: `tui/tui.go`

**Step 1: Add minimum size check in View()**

Update the `View()` function to show a warning for very small terminals:

```go
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show warning for very small terminals
	if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
		return m.renderSizeWarning()
	}

	// ... rest of existing View() code
}
```

**Step 2: Add renderSizeWarning() function**

Add this new function:

```go
// renderSizeWarning renders a warning for terminals that are too small
func (m Model) renderSizeWarning() string {
	var s strings.Builder

	s.WriteString(WarningStyle.Render("⚠ Terminal Too Small"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("Current: %dx%d\n", m.width, m.height))
	s.WriteString(fmt.Sprintf("Minimum: %dx%d\n", MinTerminalWidth, MinTerminalHeight))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("Please resize your terminal"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("or press 'q' to quit"))

	return BoxStyle.Render(s.String())
}
```

**Step 3: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add tui/tui.go
git commit -m "feat(tui): add minimum terminal size warning

Show friendly warning when terminal is smaller than 60x20
with current and required dimensions.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 6: Improve CenterDialog for Responsive Dialogs

**Files:**
- Modify: `tui/styles.go`

**Step 1: Update CenterDialog function**

Replace the existing `CenterDialog` function:

```go
// CenterDialog centers a dialog within the terminal dimensions
func CenterDialog(dialog string, termWidth, termHeight int) string {
	// Use actual terminal dimensions, with reasonable minimums
	if termWidth < MinTerminalWidth {
		termWidth = MinTerminalWidth
	}
	if termHeight < MinTerminalHeight {
		termHeight = MinTerminalHeight
	}

	// Use lipgloss.Place to center the dialog with a background
	overlay := lipgloss.Place(
		termWidth,
		termHeight,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceBackground(ColorBgDark),
		lipgloss.WithWhitespaceForeground(ColorBgDark),
	)

	return overlay
}
```

**Step 2: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add tui/styles.go
git commit -m "refactor(tui): update CenterDialog to use responsive constants

Use MinTerminalWidth and MinTerminalHeight constants for
consistent minimum dimensions.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 7: Update Tabs Rendering for Responsive Layout

**Files:**
- Modify: `tui/tui.go`

**Step 1: Find and update renderTabs() function**

Make tabs responsive:

```go
// renderTabs renders the tab bar
func (m Model) renderTabs() string {
	tabs := []struct {
		name   string
		short  string
		key    string
	}{
		{"Dashboard", "Dash", "1"},
		{"Commands", "Cmd", "2"},
		{"Security", "Sec", "3"},
		{"Files", "Files", "4"},
		{"Activity", "Log", "5"},
	}

	// Use short names on small terminals
	useShortNames := m.width < SmallWidth

	var renderedTabs []string
	for i, tab := range tabs {
		name := tab.name
		if useShortNames {
			name = tab.short
		}

		label := fmt.Sprintf(" [%s] %s ", tab.key, name)

		if Tab(i) == m.activeTab {
			renderedTabs = append(renderedTabs, TabActiveStyle.Render(label))
		} else {
			renderedTabs = append(renderedTabs, TabInactiveStyle.Render(label))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Center, renderedTabs...)

	// Wrap in tab bar style
	return TabBarStyle.Copy().Width(m.width).Render(tabBar)
}
```

**Step 2: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add tui/tui.go
git commit -m "feat(tui): responsive tab names for small terminals

Use short tab names (Dash, Cmd, Sec, Files, Log) on terminals
narrower than 80 characters.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 8: Final Build, Test, and Install

**Step 1: Full rebuild**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 2: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 3: Test responsive behavior**

Test at different terminal sizes:
1. Small (60x20): `resize -s 20 60` or resize terminal manually
   - Verify single column layout in dashboard
   - Verify preview panel hidden in files tab
   - Verify short keybinding hints
   - Verify short tab names

2. Medium (100x30):
   - Verify two column layout in dashboard
   - Verify preview panel works in files tab
   - Verify full keybinding hints

3. Large (150x40):
   - Verify comfortable card widths with max constraints
   - Verify full features available

4. Very small (50x15):
   - Verify "Terminal Too Small" warning appears

**Step 4: Install to /usr/local/bin**

Run:
```bash
sudo rm -f /usr/local/bin/podx
sudo cp podx /usr/local/bin/podx
```

**Step 5: Verify installation**

Run: `/usr/local/bin/podx version`
Expected: Shows version info

---

## Summary of Changes

1. **Responsive Constants:** Added terminal size breakpoints (Small/Medium/Large) and helper functions
2. **Dashboard:** Adaptive card layout based on terminal size
3. **Files Tab:** Auto-hide preview on small terminals, responsive panel widths
4. **Encrypt Dialog:** Responsive dialog width, adaptive file list
5. **Size Warning:** Friendly message for terminals smaller than 60x20
6. **Tabs:** Short tab names on small terminals
7. **General:** Consistent use of responsive helpers throughout

