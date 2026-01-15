package git

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hades/podx/project"
	"github.com/hades/podx/security"
)

// SyncOptions holds options for the sync command
type SyncOptions struct {
	Message      string // Commit message (if empty, will prompt)
	Remote       string // Remote name (default: origin)
	Branch       string // Branch name (default: current branch)
	SecurityMode string // Security mode: default, strict, relaxed
}

// SyncResult holds the result of a sync operation
type SyncResult struct {
	FilesEncrypted int
	FilesStaged    int
	CommitMessage  string
	Remote         string
	Branch         string
	Pushed         bool
	Warnings       []string
}

// Sync performs the full sync workflow: encrypt -> check -> add -> commit -> push
func Sync(dir string, opts *SyncOptions) (*SyncResult, error) {
	result := &SyncResult{}

	// Check if it's a git repo
	if !IsGitRepo(dir) {
		return nil, fmt.Errorf("not a git repository")
	}

	// Step 1: Load project and auto-encrypt
	fmt.Println("\n[1/5] Encrypting secrets...")
	p, err := project.Load(dir)
	if err != nil {
		// Not a podx project, skip encryption
		fmt.Println("  (no .podx.yaml found, skipping encryption)")
	} else {
		count, err := p.EncryptAll()
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		result.FilesEncrypted = count
		if count > 0 {
			fmt.Printf("  Encrypted %d file(s)\n", count)
		} else {
			fmt.Println("  No files to encrypt")
		}
	}

	// Step 2: Security check
	fmt.Println("\n[2/5] Running security check...")
	securityMode := opts.SecurityMode
	if securityMode == "" {
		securityMode = "default"
	}

	checkResult := security.CheckProjectWithSeverity(dir, securityMode)
	if !checkResult.Passed {
		// Handle based on severity
		for _, issue := range checkResult.Issues {
			switch issue.Severity {
			case security.SeverityHigh:
				return nil, fmt.Errorf("HIGH security issue found: %s at %s:%d\n  %s\n  This must be fixed manually before sync",
					issue.Pattern, issue.File, issue.Line, issue.Content)
			case security.SeverityMedium:
				// Ask user
				fmt.Printf("  MEDIUM: %s at %s:%d\n", issue.Pattern, issue.File, issue.Line)
				fmt.Printf("    %s\n", issue.Content)
				fmt.Print("  Continue anyway? (y/N): ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					return nil, fmt.Errorf("sync cancelled by user")
				}
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s at %s:%d", issue.Pattern, issue.File, issue.Line))
			case security.SeverityLow:
				fmt.Printf("  LOW: %s at %s:%d (continuing)\n", issue.Pattern, issue.File, issue.Line)
				result.Warnings = append(result.Warnings, fmt.Sprintf("%s at %s:%d", issue.Pattern, issue.File, issue.Line))
			}
		}
	} else {
		fmt.Println("  No security issues found")
	}

	// Step 3: Git add
	fmt.Println("\n[3/5] Staging files...")
	if err := AddAll(dir); err != nil {
		return nil, fmt.Errorf("git add failed: %w", err)
	}

	staged, err := GetStagedFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	if len(staged) == 0 {
		fmt.Println("  No changes to commit")
		return result, nil
	}

	result.FilesStaged = len(staged)
	fmt.Printf("  Staged %d file(s):\n", len(staged))
	for _, f := range staged {
		if len(staged) <= 10 {
			fmt.Printf("    %s\n", f)
		}
	}
	if len(staged) > 10 {
		fmt.Printf("    ... and %d more\n", len(staged)-10)
	}

	// Step 4: Commit
	fmt.Println("\n[4/5] Creating commit...")
	message := opts.Message
	if message == "" {
		// Generate suggestion and prompt
		suggestion := GenerateCommitMessage(staged)
		fmt.Printf("  Suggested message: %s\n", suggestion)
		fmt.Print("  Press Enter to accept, or type new message: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			message = suggestion
		} else {
			message = input
		}
	}

	result.CommitMessage = message

	if err := Commit(dir, message); err != nil {
		return nil, fmt.Errorf("git commit failed: %w", err)
	}
	fmt.Printf("  Committed: %s\n", message)

	// Step 5: Push
	fmt.Println("\n[5/5] Pushing to remote...")

	remote := opts.Remote
	if remote == "" {
		remote, err = GetDefaultRemote(dir)
		if err != nil {
			return nil, fmt.Errorf("no remote configured. Add with: git remote add origin <url>")
		}
	}
	result.Remote = remote

	branch := opts.Branch
	if branch == "" {
		branch, err = GetCurrentBranch(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}
	}
	result.Branch = branch

	// Check if upstream is set
	if HasUpstream(dir) {
		if err := Push(dir, remote, branch); err != nil {
			return nil, fmt.Errorf("git push failed: %w", err)
		}
	} else {
		if err := PushSetUpstream(dir, remote, branch); err != nil {
			return nil, fmt.Errorf("git push failed: %w", err)
		}
	}

	result.Pushed = true
	fmt.Printf("  Pushed to %s/%s\n", remote, branch)

	return result, nil
}

// PrintSyncResult prints a summary of the sync operation
func PrintSyncResult(result *SyncResult) {
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Sync completed successfully!")
	fmt.Println(strings.Repeat("─", 50))

	if result.FilesEncrypted > 0 {
		fmt.Printf("  Encrypted: %d file(s)\n", result.FilesEncrypted)
	}
	fmt.Printf("  Committed: %d file(s)\n", result.FilesStaged)
	fmt.Printf("  Message:   %s\n", result.CommitMessage)
	fmt.Printf("  Pushed:    %s/%s\n", result.Remote, result.Branch)

	if len(result.Warnings) > 0 {
		fmt.Println("\n  Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("    - %s\n", w)
		}
	}
}
