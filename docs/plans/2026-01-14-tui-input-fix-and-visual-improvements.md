# TUI Input Fix and Visual Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix key binding conflict in Add Recipient dialog where 'g' triggers Generate Key instead of typing, and improve overall TUI visual appearance with the Dracula theme.

**Architecture:** Fix the input handling in encrypt_dialog.go by checking if a text input is focused before processing shortcut keys. Improve visual styling throughout the TUI for better contrast and readability.

**Tech Stack:** Go, Bubble Tea (bubbletea), Lipgloss

---

## Task 1: Fix 'g' Key Conflict in Add Recipient Dialog

**Problem:** In the Add Recipient dialog (`StateAddRecipient`), pressing 'g' or 'G' triggers the Generate Key action (line 492-499 in encrypt_dialog.go) instead of typing the letter into the Name or Age Key text input field.

**Files:**
- Modify: `tui/encrypt_dialog.go:424-510`

**Step 1: Identify the bug location**

The issue is in `updateAddRecipient()`. The switch statement handles 'g' and 'G' keys BEFORE they can reach the default case that passes characters to the focused text input.

Current code (lines 492-499):
```go
case "g", "G":
	// Switch to key generation state
	m.state = StateGenerateKey
	m.generatedPublicKey = ""
	m.generatedPrivateKey = ""
	m.generatingKey = false
	m.errorMsg = ""
	return m, nil
```

**Step 2: Fix the key handling logic**

The fix requires:
1. Only trigger Generate Key with 'g' when no text input is focused
2. Use Ctrl+G or a different key combination for Generate Key when inputs are focused

Modify `updateAddRecipient()` in `tui/encrypt_dialog.go`:

Replace lines 424-512 with:

```go
// updateAddRecipient handles the add recipient state
func (m EncryptDialogModel) updateAddRecipient(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	// First, handle keys that should work regardless of input focus
	switch msg.String() {
	case "esc":
		m.state = StateAgeKeyConfirm
		m.recipientName.Blur()
		m.recipientKey.Blur()
		m.errorMsg = ""
		return m, nil

	case "tab", "down":
		if m.focusedInput == 0 {
			m.focusedInput = 1
			m.recipientName.Blur()
			m.recipientKey.Focus()
		} else {
			m.focusedInput = 0
			m.recipientKey.Blur()
			m.recipientName.Focus()
		}
		return m, textinput.Blink

	case "shift+tab", "up":
		if m.focusedInput == 1 {
			m.focusedInput = 0
			m.recipientKey.Blur()
			m.recipientName.Focus()
			return m, textinput.Blink
		}

	case "enter":
		name := strings.TrimSpace(m.recipientName.Value())
		key := strings.TrimSpace(m.recipientKey.Value())

		// Validate
		if name == "" {
			m.errorMsg = "Name is required"
			return m, nil
		}
		if key == "" {
			m.errorMsg = "Age public key is required"
			return m, nil
		}
		if !strings.HasPrefix(key, "age1") {
			m.errorMsg = "Invalid Age key (must start with 'age1')"
			return m, nil
		}

		// Add recipient to project
		if m.project != nil {
			m.project.Config.Recipients = append(m.project.Config.Recipients, project.Recipient{
				Name: name,
				Key:  key,
			})
			// Save config
			if err := m.project.Save(); err != nil {
				m.errorMsg = "Failed to save: " + err.Error()
				return m, nil
			}
		}

		m.successMsg = fmt.Sprintf("Added recipient: %s", name)
		m.state = StateAgeKeyConfirm
		m.recipientName.Blur()
		m.recipientKey.Blur()
		m.errorMsg = ""
		return m, nil

	case "ctrl+g":
		// Use Ctrl+G for Generate Key (allows 'g' to be typed in inputs)
		m.state = StateGenerateKey
		m.generatedPublicKey = ""
		m.generatedPrivateKey = ""
		m.generatingKey = false
		m.errorMsg = ""
		return m, nil
	}

	// Pass all other keys to the focused input (including 'g', 'G', etc.)
	var cmd tea.Cmd
	if m.focusedInput == 0 {
		m.recipientName, cmd = m.recipientName.Update(msg)
	} else {
		m.recipientKey, cmd = m.recipientKey.Update(msg)
	}
	m.errorMsg = "" // Clear error on typing
	return m, cmd
}
```

**Step 3: Update the help text in renderAddRecipient()**

Find the help text in `renderAddRecipient()` and update the keybinding hint from `[G]` to `[Ctrl+G]`.

Search for line containing `[G] Generate key` and replace with `[Ctrl+G] Generate key`.

**Step 4: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 5: Manual test**

1. Run `./podx` to open TUI
2. Go to Files tab (press 4)
3. Select a file and press 'e' to encrypt
4. Select Age encryption
5. If no recipients, the Add Recipient dialog appears
6. Type a name containing 'g' (e.g., "George")
7. Verify that 'g' is typed into the name field, NOT triggering Generate Key
8. Press Ctrl+G to verify Generate Key still works

**Step 6: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "fix(tui): allow typing 'g' in Add Recipient dialog inputs

Change Generate Key shortcut from 'g' to 'Ctrl+G' to prevent
conflict with typing in text input fields.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: Improve Files Tab Visual Appearance

**Files:**
- Modify: `tui/files.go:734-771` (getFileIcon function)
- Modify: `tui/files.go:973-1040` (renderFileLine function)

**Step 1: Update file icons to use emoji for better visual appeal**

Replace the `getFileIcon()` function at lines 734-771:

```go
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
```

**Step 2: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add tui/files.go
git commit -m "feat(tui): improve file icons with emoji for visual appeal

Replace ASCII file type indicators with emoji icons for
better visual distinction between file types.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: Improve Add Recipient Dialog Visual Design

**Files:**
- Modify: `tui/encrypt_dialog.go` (renderAddRecipient function)

**Step 1: Find and update renderAddRecipient()**

Locate the `renderAddRecipient()` function and update the styling:

```go
// renderAddRecipient renders the add recipient state
func (m EncryptDialogModel) renderAddRecipient() string {
	var s strings.Builder

	// Title with icon
	s.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		MarginBottom(1).
		Render("➕ Add Recipient"))
	s.WriteString("\n\n")

	// Description
	s.WriteString(MutedStyle.Render("Add a team member who can decrypt your files."))
	s.WriteString("\n\n")

	// Name field with better styling
	nameLabel := "Name:"
	if m.focusedInput == 0 {
		nameLabel = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("Name:")
	}
	s.WriteString(fmt.Sprintf("%s\t%s %s\n", nameLabel, m.renderInputCursor(0), m.recipientName.View()))

	// Age Key field
	keyLabel := "Age Key:"
	if m.focusedInput == 1 {
		keyLabel = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("Age Key:")
	}
	s.WriteString(fmt.Sprintf("%s\t%s %s\n", keyLabel, m.renderInputCursor(1), m.recipientKey.View()))

	s.WriteString("\n")

	// Success message
	if m.successMsg != "" {
		s.WriteString(SuccessStyle.Render("✓ " + m.successMsg))
		s.WriteString("\n")
	}

	// Error message
	if m.errorMsg != "" {
		s.WriteString(ErrorStyle.Render("✗ " + m.errorMsg))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Help text
	s.WriteString(MutedStyle.Render("Age public keys start with 'age1...'"))
	s.WriteString("\n")
	s.WriteString(MutedStyle.Render("Generate a key pair with: podx keygen -t age"))
	s.WriteString("\n\n")

	// Action hints - updated Ctrl+G
	s.WriteString(MutedStyle.Render("[Enter] Add  [Ctrl+G] Generate key  [Tab] Next  [Esc] Cancel"))

	return s.String()
}

// renderInputCursor returns a cursor indicator for the input field
func (m EncryptDialogModel) renderInputCursor(inputIndex int) string {
	if m.focusedInput == inputIndex {
		return lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(">")
	}
	return " "
}
```

**Step 2: Build and test**

Run: `go build -o podx .`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): improve Add Recipient dialog visual design

Better label styling, cursor indicators, and help text layout.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: Final Build, Test, and Install

**Step 1: Full rebuild**

Run: `go build -o podx .`

**Step 2: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 3: Manual integration test**

1. Run `./podx`
2. Navigate to Files tab (press 4)
3. Verify emoji icons display correctly
4. Select a file and press 'e' to encrypt
5. Select Age encryption method
6. In Add Recipient dialog:
   - Type "George" in Name field - verify 'g' works
   - Type "age1test" in Age Key field - verify 'g' works
   - Press Ctrl+G to verify Generate Key opens
   - Press Esc to cancel
7. Navigate back and verify UI looks good

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

1. **Bug Fix:** Changed Generate Key shortcut from 'g' to 'Ctrl+G' in Add Recipient dialog to allow typing 'g' in text inputs
2. **Visual:** Updated file icons from ASCII brackets to emoji for better visual distinction
3. **Visual:** Improved Add Recipient dialog styling with better labels, cursor indicators, and help text
