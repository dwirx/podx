# TUI Key Generation Improvements Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Complete the key generation integration in TUI by adding remaining features: "Use My Key" shortcut, Dashboard key status display, and Dashboard key generation action.

**Architecture:** Extend the existing implementation by adding the missing features from the original plan. The keygen package already has `LoadAgeRecipient()` and `GetAgeKeyInfo()` functions. We need to wire these into the Dashboard model and add the "M" shortcut in Age Key confirm state.

**Tech Stack:** Go, Bubbletea, existing `keygen` package

---

### Task 1: Add "Use My Key" Shortcut to Age Key Confirm

**Files:**
- Modify: `tui/encrypt_dialog.go:356-393`

**Step 1: Add import for project package if not present**

The project package is already imported. Verify at line 14.

**Step 2: Add 'M' key handler in updateAgeKeyConfirm**

In `updateAgeKeyConfirm` method (around line 356), add a new case after `case "a", "A":`:

```go
case "m", "M":
	// Load and use my own key as recipient
	pubKey, err := keygen.LoadAgeRecipient()
	if err != nil {
		m.errorMsg = "No key found. Press [A] then [G] to generate one."
		return m, nil
	}
	// Auto-add self as recipient
	if m.project != nil {
		// Check if already added
		for _, r := range m.project.Config.Recipients {
			if r.Key == pubKey {
				m.successMsg = "Your key is already added as recipient"
				return m, nil
			}
		}
		m.project.Config.Recipients = append(m.project.Config.Recipients, project.Recipient{
			Name: "Me (local)",
			Key:  pubKey,
		})
		if err := m.project.Save(); err != nil {
			m.errorMsg = "Failed to save: " + err.Error()
			return m, nil
		}
		m.successMsg = "Added your local key as recipient"
	}
	return m, nil
```

**Step 3: Update renderAgeKeyConfirm footer**

Change the footer line (around line 998) from:

```go
s.WriteString(MutedStyle.Render(fmt.Sprintf("[Enter] %s  [A] Add recipient  [Esc] Back", action)))
```

to:

```go
s.WriteString(MutedStyle.Render(fmt.Sprintf("[Enter] %s  [M] My key  [A] Add recipient  [Esc] Back", action)))
```

**Step 4: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 5: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): add 'use my key' shortcut for quick self-addition as recipient"
```

---

### Task 2: Add Key Info Field to Dashboard Model

**Files:**
- Modify: `tui/dashboard.go`

**Step 1: Add import for keygen package**

Add to imports (around line 3-14):

```go
"github.com/hades/podx/keygen"
```

**Step 2: Add keyInfo field to DashboardModel**

After `version string` field (around line 37), add:

```go
keyInfo keygen.KeyInfo
```

**Step 3: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 4: Commit**

```bash
git add tui/dashboard.go
git commit -m "feat(tui): add key info field to dashboard model"
```

---

### Task 3: Load Key Info on Project Load

**Files:**
- Modify: `tui/dashboard.go`

**Step 1: Update projectLoadedMsg struct**

Update the struct (around line 40-44) to include keyInfo:

```go
type projectLoadedMsg struct {
	project *project.Project
	err     error
	keyInfo keygen.KeyInfo
}
```

**Step 2: Update loadProject function**

Change the function (around line 97-101) from:

```go
func (m DashboardModel) loadProject() tea.Msg {
	proj, err := project.Load(m.cwd)
	return projectLoadedMsg{project: proj, err: err}
}
```

to:

```go
func (m DashboardModel) loadProject() tea.Msg {
	proj, err := project.Load(m.cwd)
	keyInfo := keygen.GetAgeKeyInfo()
	return projectLoadedMsg{project: proj, err: err, keyInfo: keyInfo}
}
```

**Step 3: Handle keyInfo in Update**

In the `projectLoadedMsg` handler (around line 154-162), add after `m.err = msg.err`:

```go
m.keyInfo = msg.keyInfo
```

**Step 4: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 5: Commit**

```bash
git add tui/dashboard.go
git commit -m "feat(tui): load key info on project load"
```

---

### Task 4: Display Key Status in Dashboard Security Card

**Files:**
- Modify: `tui/dashboard.go:346-410`

**Step 1: Add key status to renderSecurityStatusWithWidth**

After the hook status section (around line 403), before the style section, add:

```go
// Local key status
if m.keyInfo.HasKey {
	keyPreview := m.keyInfo.PublicKey
	if len(keyPreview) > 20 {
		keyPreview = keyPreview[:20] + "..."
	}
	lines = append(lines, SuccessStyle.Render("  [OK] Local Age key"))
	lines = append(lines, MutedStyle.Render(fmt.Sprintf("       %s", keyPreview)))
} else {
	lines = append(lines, WarningStyle.Render("  [--] No local Age key"))
	lines = append(lines, MutedStyle.Render("       Press [K] to generate"))
}
```

**Step 2: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 3: Commit**

```bash
git add tui/dashboard.go
git commit -m "feat(tui): display local Age key status in dashboard security card"
```

---

### Task 5: Add "Generate Key" Action to Dashboard

**Files:**
- Modify: `tui/dashboard.go`

**Step 1: Add new action constant**

After `ActionHookInstall` (around line 21), add:

```go
ActionKeygen
```

**Step 2: Update actions slice in NewDashboardModel**

Change the actions slice (around line 68-73) to:

```go
actions: []string{
	"Encrypt All",
	"Decrypt All",
	"Run Check",
	"Install Hook",
	"Generate Key",
},
```

**Step 3: Update actionIcons and actionDescs in renderQuickActionsWithWidth**

Update the slices (around line 425-431):

```go
actionIcons := []string{"E", "D", "C", "H", "K"}
actionDescs := []string{
	"Encrypt all secret files",
	"Decrypt all secret files",
	"Run security checks",
	"Install pre-commit hook",
	"Generate new Age key pair",
}
```

**Step 4: Add keygen case in executeAction**

In `executeAction` method (around line 146), before the closing `}` of the switch, add:

```go
case ActionKeygen:
	result, err := keygen.GenerateAge()
	if err != nil {
		return actionResultMsg{action: action, success: false, message: err.Error()}
	}
	// Show truncated public key
	pubKey := result.PublicKey
	if len(pubKey) > 30 {
		pubKey = pubKey[:30] + "..."
	}
	return actionResultMsg{action: action, success: true, message: fmt.Sprintf("Generated key: %s", pubKey)}
```

**Step 5: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 6: Commit**

```bash
git add tui/dashboard.go
git commit -m "feat(tui): add key generation to dashboard quick actions"
```

---

### Task 6: Handle Enter Key After Key Generation Success

**Files:**
- Modify: `tui/encrypt_dialog.go:486-512`

**Step 1: Update updateGenerateKey to handle Enter after success**

In `updateGenerateKey` method (around line 486), update the `case "enter", "y", "Y":` block to also handle the case when a key has already been generated:

```go
case "enter", "y", "Y":
	if m.generatedPublicKey != "" {
		// Key already generated, use it
		m.recipientKey.SetValue(m.generatedPublicKey)
		m.state = StateAddRecipient
		m.focusedInput = 0
		m.recipientName.Focus()
		m.recipientName.SetValue("Me (local)")
		return m, textinput.Blink
	}
	if !m.generatingKey {
		m.generatingKey = true
		m.errorMsg = ""
		return m, m.doGenerateKey()
	}
```

**Step 2: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 3: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): auto-fill recipient name when using generated key"
```

---

### Task 7: Final Build and Test

**Step 1: Build the full project**

Run: `make build`
Expected: Success

**Step 2: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 3: Manual test - Launch TUI**

Run: `./podx`
Expected: TUI launches, dashboard shows key status in security card

**Step 4: Manual test - Generate key from Dashboard**

Navigate to "Generate Key" action (5th item), press Enter
Expected: Key is generated and message shown

**Step 5: Manual test - Use "My Key" shortcut**

Go to Files tab, select a file, press 'e', select Age Key, press 'M'
Expected: Your local key is added as recipient

**Step 6: Manual test - Generate key from Add Recipient dialog**

Go to Files tab, select a file, press 'e', select Age Key, press 'A' to add recipient, press 'G'
Expected: Key generation dialog appears, after generation pressing Enter pre-fills "Me (local)" as name

**Step 7: Commit all changes if any remaining**

```bash
git status
# If there are uncommitted changes:
git add .
git commit -m "feat(tui): complete key generation integration

- Add 'My Key' shortcut [M] to quickly add self as recipient
- Display local Age key status in Dashboard security card
- Add Generate Key action to Dashboard quick actions
- Auto-fill recipient name when using generated key"
```

