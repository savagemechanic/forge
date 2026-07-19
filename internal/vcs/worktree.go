// Package vcs provides git-based isolation for safe code modification.
// Changes are made in a git worktree, keeping the main working tree clean
// until the user approves and merges.
package vcs

import (
	"fmt"
	"os/exec"
	"strings"
)

// Worktree represents an isolated git worktree for safe modifications.
type Worktree struct {
	Path   string // the worktree directory
	Branch string // the branch checked out in the worktree
	Parent string // the original repo root
}

// CreateWorktree creates a new git worktree on a fresh branch.
// This isolates all file mutations from the main working tree.
func CreateWorktree(repoRoot, branchName string) (*Worktree, error) {
	// Verify it's a git repo
	if err := runGit(repoRoot, "rev-parse", "--git-dir"); err != nil {
		return nil, fmt.Errorf("not a git repo: %s", repoRoot)
	}

	// Create a temp dir for the worktree
	tmpDir, err := exec.Command("mktemp", "-d").Output()
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	wtPath := strings.TrimSpace(string(tmpDir))

	// Remove the temp dir (git worktree add wants to create it)
	runGit(repoRoot, "worktree", "add", "--detach", wtPath)

	// Create and checkout a branch
	if err := runGit(wtPath, "checkout", "-b", branchName); err != nil {
		// Cleanup on failure
		runGit(repoRoot, "worktree", "remove", "--force", wtPath)
		return nil, fmt.Errorf("create branch: %w", err)
	}

	return &Worktree{
		Path:   wtPath,
		Branch: branchName,
		Parent: repoRoot,
	}, nil
}

// Diff returns the git diff of changes in this worktree.
func (w *Worktree) Diff() (string, error) {
	out, err := exec.Command("git", "-C", w.Path, "diff").Output()
	if err != nil {
		return "", err
	}
	// Also include untracked (new) files
	untracked, _ := exec.Command("git", "-C", w.Path, "ls-files", "--others", "--exclude-standard").Output()
	if len(untracked) > 0 {
		return string(out) + "\n# Untracked files:\n" + string(untracked), nil
	}
	return string(out), nil
}

// ChangedFiles returns the list of changed files (modified + untracked).
func (w *Worktree) ChangedFiles() ([]string, error) {
	out, err := exec.Command("git", "-C", w.Path, "status", "--porcelain").Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 3 {
			continue
		}
		files = append(files, strings.TrimSpace(line[3:]))
	}
	return files, nil
}

// Commit stages all changes and commits them.
func (w *Worktree) Commit(message string) error {
	if err := runGit(w.Path, "add", "-A"); err != nil {
		return err
	}
	return runGit(w.Path, "commit", "-m", message)
}

// MergeBack merges the worktree branch into the parent repo's current branch.
func (w *Worktree) MergeBack() error {
	// Get current branch of parent
	out, err := exec.Command("git", "-C", w.Parent, "branch", "--show-current").Output()
	if err != nil {
		return fmt.Errorf("get parent branch: %w", err)
	}
	parentBranch := strings.TrimSpace(string(out))

	// First commit any uncommitted changes in worktree
	if err := w.Commit("merge-back checkpoint"); err != nil {
		// no changes is fine
	}

	// Merge into parent
	if err := runGit(w.Parent, "merge", w.Branch); err != nil {
		return fmt.Errorf("merge into %s: %w", parentBranch, err)
	}
	return nil
}

// Abort discards the worktree and all its changes.
func (w *Worktree) Abort() error {
	if err := runGit(w.Parent, "worktree", "remove", "--force", w.Path); err != nil {
		return err
	}
	return runGit(w.Parent, "branch", "-D", w.Branch)
}

// Cleanup removes the worktree directory (after successful merge).
func (w *Worktree) Cleanup() error {
	return runGit(w.Parent, "worktree", "remove", w.Path)
}

// runGit executes a git command in the given directory.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %s (%w)", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return nil
}
