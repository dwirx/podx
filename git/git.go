package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RunGit executes a git command and returns stdout
func RunGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

// IsGitRepo checks if directory is a git repository
func IsGitRepo(dir string) bool {
	_, err := RunGit(dir, "rev-parse", "--git-dir")
	return err == nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(dir string) (string, error) {
	return RunGit(dir, "rev-parse", "--abbrev-ref", "HEAD")
}

// GetDefaultRemote returns the default remote (usually "origin")
func GetDefaultRemote(dir string) (string, error) {
	output, err := RunGit(dir, "remote")
	if err != nil {
		return "", err
	}

	remotes := strings.Split(output, "\n")
	if len(remotes) == 0 || remotes[0] == "" {
		return "", fmt.Errorf("no remote configured")
	}

	// Prefer "origin" if it exists
	for _, r := range remotes {
		if r == "origin" {
			return "origin", nil
		}
	}

	return remotes[0], nil
}

// HasRemote checks if a remote exists
func HasRemote(dir, remote string) bool {
	output, err := RunGit(dir, "remote")
	if err != nil {
		return false
	}

	for _, r := range strings.Split(output, "\n") {
		if r == remote {
			return true
		}
	}
	return false
}

// GetStatus returns the git status (short format)
func GetStatus(dir string) (string, error) {
	return RunGit(dir, "status", "--short")
}

// GetStagedFiles returns list of staged files
func GetStagedFiles(dir string) ([]string, error) {
	output, err := RunGit(dir, "diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetModifiedFiles returns list of modified (unstaged) files
func GetModifiedFiles(dir string) ([]string, error) {
	output, err := RunGit(dir, "diff", "--name-only")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// GetUntrackedFiles returns list of untracked files
func GetUntrackedFiles(dir string) ([]string, error) {
	output, err := RunGit(dir, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// Add stages files for commit
func Add(dir string, files ...string) error {
	args := append([]string{"add"}, files...)
	_, err := RunGit(dir, args...)
	return err
}

// AddAll stages all changes
func AddAll(dir string) error {
	_, err := RunGit(dir, "add", "-A")
	return err
}

// Commit creates a commit with the given message
func Commit(dir, message string) error {
	_, err := RunGit(dir, "commit", "-m", message)
	return err
}

// Push pushes to remote
func Push(dir, remote, branch string) error {
	_, err := RunGit(dir, "push", remote, branch)
	return err
}

// PushSetUpstream pushes and sets upstream tracking
func PushSetUpstream(dir, remote, branch string) error {
	_, err := RunGit(dir, "push", "-u", remote, branch)
	return err
}

// HasUpstream checks if current branch has an upstream
func HasUpstream(dir string) bool {
	_, err := RunGit(dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	return err == nil
}

// GetLastCommitMessage returns the last commit message
func GetLastCommitMessage(dir string) (string, error) {
	return RunGit(dir, "log", "-1", "--format=%s")
}

// GetRecentCommitMessages returns recent commit messages for style reference
func GetRecentCommitMessages(dir string, count int) ([]string, error) {
	output, err := RunGit(dir, "log", fmt.Sprintf("-%d", count), "--format=%s")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}
