package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: Commit stages and commits; MergeBack merges into parent;
// Cleanup removes the worktree dir.

func TestWorktree_Commit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/commit")
	require.NoError(t, err)
	defer wt.Abort()

	require.NoError(t, os.WriteFile(filepath.Join(wt.Path, "committed.go"), []byte("package x"), 0644))
	require.NoError(t, wt.Commit("add file"))

	// After commit, changed files should be empty (all committed)
	files, err := wt.ChangedFiles()
	require.NoError(t, err)
	assert.Empty(t, files, "no uncommitted changes after Commit")
}

func TestWorktree_MergeBack(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/merge")
	require.NoError(t, err)

	// Make a change and merge it back
	require.NoError(t, os.WriteFile(filepath.Join(wt.Path, "merged.go"), []byte("package merged"), 0644))
	require.NoError(t, wt.MergeBack())

	// The file should now exist in the parent repo
	_, err = os.Stat(filepath.Join(repo, "merged.go"))
	assert.NoError(t, err, "merged file should exist in parent after MergeBack")
}

func TestWorktree_Cleanup(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/cleanup")
	require.NoError(t, err)

	wtPath := wt.Path
	require.NoError(t, wt.Cleanup())

	_, statErr := os.Stat(wtPath)
	assert.True(t, os.IsNotExist(statErr), "worktree should be removed by Cleanup")
}

func TestWorktree_Diff_UntrackedOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/untracked")
	require.NoError(t, err)
	defer wt.Abort()

	require.NoError(t, os.WriteFile(filepath.Join(wt.Path, "brandnew.go"), []byte("package b"), 0644))

	diff, err := wt.Diff()
	require.NoError(t, err)
	assert.Contains(t, diff, "brandnew.go")
}
