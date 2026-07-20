package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: a worktree isolates changes from the parent tree.
// After Abort(), the parent working tree is unchanged.

// setupGitRepo creates a real git repo in a temp dir with an initial commit.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %s: %s", args, out)
	}

	run("init")
	run("config", "user.name", "test")
	run("config", "user.email", "test@test")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644))
	run("add", "-A")
	run("commit", "-m", "initial")

	return dir
}

func TestCreateWorktree_IsolatesChanges(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/test")
	require.NoError(t, err)
	defer wt.Abort()

	// Modify a file in the worktree (not the parent)
	wtPath := filepath.Join(wt.Path, "README.md")
	require.NoError(t, os.WriteFile(wtPath, []byte("# modified"), 0644))

	// Parent should still have the original
	parentContent, err := os.ReadFile(filepath.Join(repo, "README.md"))
	require.NoError(t, err)
	assert.Equal(t, "# test", string(parentContent),
		"parent tree must be unchanged while worktree holds the edit")
}

func TestWorktree_ChangedFiles(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/diff")
	require.NoError(t, err)
	defer wt.Abort()

	// Add a new file
	require.NoError(t, os.WriteFile(filepath.Join(wt.Path, "new.go"), []byte("package x"), 0644))

	files, err := wt.ChangedFiles()
	require.NoError(t, err)
	assert.Contains(t, files, "new.go")
}

func TestWorktree_Diff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/diff2")
	require.NoError(t, err)
	defer wt.Abort()

	require.NoError(t, os.WriteFile(filepath.Join(wt.Path, "new.go"), []byte("package x"), 0644))

	diff, err := wt.Diff()
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
}

func TestWorktree_Abort_CleansUp(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := setupGitRepo(t)
	wt, err := CreateWorktree(repo, "feature/abort")
	require.NoError(t, err)

	wtPath := wt.Path
	require.NoError(t, wt.Abort())

	_, statErr := os.Stat(wtPath)
	assert.True(t, os.IsNotExist(statErr), "worktree dir should be removed after Abort")
}

func TestCreateWorktree_NotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := CreateWorktree(dir, "branch")
	assert.Error(t, err)
}
