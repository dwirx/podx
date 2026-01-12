# PODX TUI Design

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add interactive TUI dashboard using Bubbletea with vim-style navigation (h,j,k,l).

**Architecture:** New `tui/` package with multi-tab interface: Dashboard, Commands, Security. Launches when `podx` runs without arguments.

**Tech Stack:** Bubbletea (TUI framework), Lipgloss (styling), Bubbles (components)

---

## Features Overview

### Entry Point
- `podx` (no args) → Opens TUI
- `podx <command>` → Runs command normally (unchanged)

### Three Tabs

| Tab | Key | Purpose |
|-----|-----|---------|
| Dashboard | `1` | Project status, encryption status, recipients |
| Commands | `2` | Interactive menu to run podx commands |
| Security | `3` | Security check results with findings |

### Navigation

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch tabs |
| `1`, `2`, `3` | Jump to tab |
| `h` / `←` | Left / Back |
| `j` / `↓` | Down |
| `k` / `↑` | Up |
| `l` / `→` | Right / Enter |
| `Enter` | Select / Execute |
| `q` / `Esc` | Quit |
| `?` | Help overlay |
| `r` | Refresh data |

---

## Visual Layout

```
╭─────────────────────────────────────────────────────────────╮
│  PODX v1.0.0                              [1] [2] [3]  ? q  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─ Project ──────────────┐  ┌─ Encryption ───────────────┐ │
│  │ 📁 /home/user/myapp    │  │ ✅ All secrets encrypted   │ │
│  │ 🔐 Backend: age        │  │ 📄 .env.podx (encrypted)   │ │
│  │ 👥 2 recipients        │  │                            │ │
│  └────────────────────────┘  └────────────────────────────┘ │
│                                                             │
│  ┌─ Recipients ───────────────────────────────────────────┐ │
│  │ • Owner (age1abc...)                                   │ │
│  │ • Team (age1def...)                                    │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌─ Quick Actions ────────────────────────────────────────┐ │
│  │ [E] Encrypt All  [D] Decrypt All  [C] Check  [H] Hook  │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  ↑↓/jk: navigate  Enter/l: select  Tab: switch  q: quit    │
╰─────────────────────────────────────────────────────────────╯
```

---

## File Structure

```
tui/
├── tui.go           # Main TUI entry point, model, update loop
├── dashboard.go     # Dashboard tab component
├── commands.go      # Commands tab component
├── security.go      # Security tab component
├── styles.go        # Lipgloss styles (colors, borders)
├── keys.go          # Key bindings
└── help.go          # Help overlay
```

---

## Color Scheme

| Element | Color |
|---------|-------|
| Header | Cyan bold |
| Success (✅) | Green |
| Warning (⚠️) | Yellow |
| Error (❌) | Red |
| Selected | Cyan background |
| Muted text | Gray |
| Border | Dim white |

---

## Commands Tab Menu

```
┌─ Commands ─────────────────────────────────────────┐
│                                                    │
│  PROJECT                                           │
│  > init           Initialize PODX project         │
│    add-recipient  Add team member                 │
│    encrypt-all    Encrypt all secrets             │
│    decrypt-all    Decrypt all secrets             │
│    status         Show project status             │
│                                                    │
│  SECURITY                                          │
│    check          Run security checks             │
│    hook install   Install pre-commit hook         │
│    hook uninstall Remove pre-commit hook          │
│                                                    │
│  OTHER                                             │
│    keygen         Generate key pair               │
│    update         Update PODX                     │
│                                                    │
└────────────────────────────────────────────────────┘
```

---

## Security Tab

```
┌─ Security Check ───────────────────────────────────┐
│                                                    │
│  ✅ Encryption Status    All secrets encrypted    │
│  ✅ Gitignore            Properly configured      │
│  ✅ Pattern Scan         No secrets detected      │
│                                                    │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Result: PASSED                                    │
│                                                    │
│  [R] Refresh  [F] Fix Issues  [I] Install Hook    │
│                                                    │
└────────────────────────────────────────────────────┘
```

---

## Implementation Notes

1. **State management**: Use Bubbletea's Elm-style architecture
2. **Async operations**: Use tea.Cmd for encrypt/decrypt operations
3. **Error handling**: Show errors in status bar, don't crash
4. **Responsive**: Adapt layout to terminal size
5. **Graceful degradation**: If not a PODX project, show init prompt
