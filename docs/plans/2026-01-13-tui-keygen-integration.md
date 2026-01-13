# TUI Key Generation Integration Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add key generation capability directly in TUI so users can generate Age key pairs and automatically use them as recipients without leaving the interface.

**Architecture:** Extend the Add Recipient dialog in `encrypt_dialog.go` to include a "Generate New Key" option. Add a new state `StateGenerateKey` that handles key generation using the existing `keygen` package. Auto-populate the recipient fields after generation.

**Tech Stack:** Go, Bubbletea, existing `keygen` package

---

### Task 1: Add Generate Key State and Fields to EncryptDialogModel

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Add new dialog state constant**

After line 33 (after `StateAddRecipient`), add:

```go
StateGenerateKey  // State for generating a new key pair
```

**Step 2: Add generated key fields to EncryptDialogModel struct**

After line 57 (after `recipientKey textinput.Model`), add:

```go
// Generated key fields
generatedPublicKey  string
generatedPrivateKey string
generatingKey       bool
```

**Step 3: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 4: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): add key generation state to encrypt dialog model"
```

---

### Task 2: Add Key Generation Message Types

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Add keygen result message type**

After the `encryptCompleteMsg` struct (around line 65), add:

```go
// keygenCompleteMsg is sent when key generation completes
type keygenCompleteMsg struct {
	success    bool
	publicKey  string
	privateKey string
	keyFile    string
	errorMsg   string
}
```

**Step 2: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 3: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): add keygen message type for async key generation"
```

---

### Task 3: Implement Key Generation Command

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Add import for keygen package**

Add to imports:

```go
"github.com/hades/podx/keygen"
```

**Step 2: Add doGenerateKey method**

After `doDecryptAge` method (around line 687), add:

```go
// doGenerateKey generates a new Age key pair
func (m EncryptDialogModel) doGenerateKey() tea.Cmd {
	return func() tea.Msg {
		result, err := keygen.GenerateAge()
		if err != nil {
			return keygenCompleteMsg{
				success:  false,
				errorMsg: err.Error(),
			}
		}

		return keygenCompleteMsg{
			success:    true,
			publicKey:  result.PublicKey,
			privateKey: result.PrivateKey,
			keyFile:    result.KeyFile,
		}
	}
}
```

**Step 3: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 4: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): implement async key generation command"
```

---

### Task 4: Handle Key Generation Message in Update

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Add handler for keygenCompleteMsg**

In the `Update` method, after the `encryptCompleteMsg` case (around line 182), add:

```go
case keygenCompleteMsg:
	m.generatingKey = false
	if msg.success {
		m.generatedPublicKey = msg.publicKey
		m.generatedPrivateKey = msg.privateKey
		// Auto-fill the recipient key field
		m.recipientKey.SetValue(msg.publicKey)
		m.successMsg = "Key generated successfully!"
		m.state = StateAddRecipient
		m.focusedInput = 0 // Focus on name field
		m.recipientName.Focus()
		return m, textinput.Blink
	} else {
		m.errorMsg = msg.errorMsg
		m.state = StateAddRecipient
	}
	return m, nil
```

**Step 2: Add StateGenerateKey case in key handling**

In the switch statement for `m.state` (around line 195), add after `StateAddRecipient`:

```go
case StateGenerateKey:
	return m.updateGenerateKey(msg)
```

**Step 3: Verify it compiles**

Run: `go build ./tui`
Expected: May fail - updateGenerateKey not defined yet

**Step 4: Commit (if compiles)**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): handle key generation completion message"
```

---

### Task 5: Implement updateGenerateKey Handler

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Add updateGenerateKey method**

After `updateAddRecipient` method (around line 439), add:

```go
// updateGenerateKey handles the key generation state
func (m EncryptDialogModel) updateGenerateKey(msg tea.KeyMsg) (EncryptDialogModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.generatingKey {
			// Can't cancel during generation
			return m, nil
		}
		m.state = StateAddRecipient
		m.generatedPublicKey = ""
		m.generatedPrivateKey = ""
		return m, nil

	case "enter", "y", "Y":
		if !m.generatingKey {
			m.generatingKey = true
			m.errorMsg = ""
			return m, m.doGenerateKey()
		}

	case "n", "N":
		m.state = StateAddRecipient
		return m, nil
	}

	return m, nil
}
```

**Step 2: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 3: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): implement key generation state handler"
```

---

### Task 6: Add Generate Key Option to Add Recipient Dialog

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Update updateAddRecipient to handle 'G' key**

In `updateAddRecipient` method, add a new case before the `default` case:

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

**Step 2: Update renderAddRecipient footer**

Change the footer line (around line 958) from:

```go
s.WriteString(MutedStyle.Render("[Enter] Add  [Tab] Next field  [Esc] Cancel"))
```

to:

```go
s.WriteString(MutedStyle.Render("[Enter] Add  [G] Generate key  [Tab] Next  [Esc] Cancel"))
```

**Step 3: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 4: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): add generate key shortcut to add recipient dialog"
```

---

### Task 7: Implement renderGenerateKey View

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Add View case for StateGenerateKey**

In the `View` method switch statement (around line 704), add:

```go
case StateGenerateKey:
	content = m.renderGenerateKey()
```

**Step 2: Add renderGenerateKey method**

After `renderAddRecipient` method (around line 961), add:

```go
// renderGenerateKey renders the key generation view
func (m EncryptDialogModel) renderGenerateKey() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("Generate Age Key Pair"))
	s.WriteString("\n\n")

	if m.generatingKey {
		s.WriteString("Generating key pair...\n")
		s.WriteString(MutedStyle.Render("Please wait..."))
		return s.String()
	}

	if m.generatedPublicKey != "" {
		// Show generated key
		s.WriteString(SuccessStyle.Render("Key generated successfully!"))
		s.WriteString("\n\n")

		s.WriteString(CardTitleStyle.Render("Public Key:"))
		s.WriteString("\n")
		// Wrap long key
		pubKey := m.generatedPublicKey
		if len(pubKey) > 50 {
			s.WriteString("  " + pubKey[:50] + "\n")
			s.WriteString("  " + pubKey[50:])
		} else {
			s.WriteString("  " + pubKey)
		}
		s.WriteString("\n\n")

		s.WriteString(CardTitleStyle.Render("Private Key:"))
		s.WriteString("\n")
		privKey := m.generatedPrivateKey
		if len(privKey) > 50 {
			s.WriteString("  " + privKey[:50] + "\n")
			s.WriteString("  " + privKey[50:])
		} else {
			s.WriteString("  " + privKey)
		}
		s.WriteString("\n\n")

		s.WriteString(WarningStyle.Render("Keys saved to ~/.config/podx/"))
		s.WriteString("\n\n")

		s.WriteString(MutedStyle.Render("[Enter] Use this key  [Esc] Back"))
	} else {
		// Confirmation prompt
		s.WriteString("This will generate a new Age X25519 key pair.\n\n")

		s.WriteString("The keys will be saved to:\n")
		s.WriteString(MutedStyle.Render("  ~/.config/podx/age-keys.txt (private)\n"))
		s.WriteString(MutedStyle.Render("  ~/.config/podx/age-recipients/default.txt (public)\n"))
		s.WriteString("\n")

		if m.errorMsg != "" {
			s.WriteString(ErrorStyle.Render("Error: " + m.errorMsg))
			s.WriteString("\n\n")
		}

		s.WriteString("Generate new key pair?\n\n")
		s.WriteString(MutedStyle.Render("[Y/Enter] Generate  [N/Esc] Cancel"))
	}

	return s.String()
}
```

**Step 3: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 4: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): add key generation view with confirmation and result display"
```

---

### Task 8: Add Quick "Use My Key" Button to Age Key Confirm

**Files:**
- Modify: `tui/encrypt_dialog.go`

**Step 1: Update updateAgeKeyConfirm to handle 'M' for "My key"**

In `updateAgeKeyConfirm` method (around line 321), add a new case:

```go
case "m", "M":
	// Load and use my own key as recipient
	pubKey, err := keygen.LoadAgeRecipient()
	if err != nil {
		m.errorMsg = "No key found. Press [G] to generate one."
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

**Step 2: Update renderAgeKeyConfirm footer**

Change the footer (around line 903) to include the new option:

```go
s.WriteString(MutedStyle.Render(fmt.Sprintf("[Enter] %s  [M] Add my key  [A] Add recipient  [Esc] Back", action)))
```

**Step 3: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 4: Commit**

```bash
git add tui/encrypt_dialog.go
git commit -m "feat(tui): add 'use my key' shortcut for quick self-addition as recipient"
```

---

### Task 9: Add Key Info Display to Dashboard

**Files:**
- Modify: `tui/dashboard.go`

**Step 1: Add import for keygen package**

Add to imports:

```go
"github.com/hades/podx/keygen"
```

**Step 2: Add keyInfo field to DashboardModel**

After `version string` field (around line 37), add:

```go
keyInfo keygen.KeyInfo
```

**Step 3: Load key info in loadProject**

In `loadProject` method (around line 99), after loading project, add:

```go
func (m DashboardModel) loadProject() tea.Msg {
	proj, err := project.Load(m.cwd)
	return projectLoadedMsg{project: proj, err: err}
}
```

Update to:

```go
func (m DashboardModel) loadProject() tea.Msg {
	proj, err := project.Load(m.cwd)
	keyInfo := keygen.GetAgeKeyInfo()
	return projectLoadedMsg{project: proj, err: err, keyInfo: keyInfo}
}
```

**Step 4: Update projectLoadedMsg struct**

Update the struct (around line 42) to include keyInfo:

```go
type projectLoadedMsg struct {
	project *project.Project
	err     error
	keyInfo keygen.KeyInfo
}
```

**Step 5: Handle keyInfo in Update**

In the `projectLoadedMsg` handler (around line 154), add:

```go
m.keyInfo = msg.keyInfo
```

**Step 6: Verify it compiles**

Run: `go build ./tui`
Expected: Success with no errors

**Step 7: Commit**

```bash
git add tui/dashboard.go
git commit -m "feat(tui): load and store key info in dashboard model"
```

---

### Task 10: Display Key Status in Dashboard Security Card

**Files:**
- Modify: `tui/dashboard.go`

**Step 1: Add key status to renderSecurityStatusWithWidth**

After the hook status section (around line 402), add:

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
	lines = append(lines, MutedStyle.Render("       Run 'podx keygen -t age' to generate"))
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

### Task 11: Add "Generate Key" to Dashboard Quick Actions

**Files:**
- Modify: `tui/dashboard.go`

**Step 1: Add new action constant**

After `ActionHookInstall` (around line 22), add:

```go
ActionKeygen
```

**Step 2: Update actions slice in NewDashboardModel**

Change the actions slice (around line 68) to:

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

Update the slices (around line 425):

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

In `executeAction` method (around line 141), add before the closing `}`:

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

### Task 12: Final Build and Test

**Step 1: Build the full project**

Run: `make build`
Expected: Success

**Step 2: Run all tests**

Run: `go test ./...`
Expected: All tests pass

**Step 3: Install binary**

Run: `sudo cp podx /usr/local/bin/`

**Step 4: Manual test - Launch TUI**

Run: `podx`
Expected: TUI launches, dashboard shows key status

**Step 5: Manual test - Generate key from Dashboard**

Navigate to "Generate Key" action, press Enter
Expected: Key is generated and message shown

**Step 6: Manual test - Generate key from Add Recipient dialog**

Go to Files tab, select a file, press 'e', select Age Key, press 'A' to add recipient, press 'G'
Expected: Key generation dialog appears

**Step 7: Commit all changes**

```bash
git add .
git commit -m "feat(tui): complete key generation integration

- Add key generation directly in Add Recipient dialog
- Add 'My Key' shortcut to quickly add self as recipient
- Display local Age key status in Dashboard
- Add Generate Key to Dashboard quick actions
- Show generated key details with save location"
```
