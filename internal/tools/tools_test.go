package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_HasBuiltInTools(t *testing.T) {
	r := NewRegistry()
	expected := []string{"read", "write", "edit", "list", "bash", "git_status"}
	for _, name := range expected {
		_, ok := r.Get(name)
		assert.True(t, ok, "tool %s should be registered", name)
	}
}

func TestRegistry_UnknownToolReturnsError(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: "."}, "nonexistent", nil)
	assert.Error(t, result.Err)
}

func TestFileReadTool_ReadsFile(t *testing.T) {
	tmpDir := t.TempDir()
	content := "hello world"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(content), 0644))

	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: tmpDir}, "read", map[string]string{"path": "test.txt"})
	require.NoError(t, result.Err)
	assert.Equal(t, content, result.Output)
}

func TestFileWriteTool_RequiresApproval(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRegistry()

	// Without approval
	result := r.Execute(&Context{WorkDir: tmpDir, Approved: false}, "write",
		map[string]string{"path": "out.txt", "content": "data"})
	assert.Error(t, result.Err, "write without approval should fail")
}

func TestFileWriteTool_WritesWhenApproved(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRegistry()

	result := r.Execute(&Context{WorkDir: tmpDir, Approved: true}, "write",
		map[string]string{"path": "out.txt", "content": "data"})
	require.NoError(t, result.Err)
	assert.Equal(t, []string{"out.txt"}, result.Changed)

	data, _ := os.ReadFile(filepath.Join(tmpDir, "out.txt"))
	assert.Equal(t, "data", string(data))
}

func TestFileEditTool_ReplacesUniqueText(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "edit.go")
	require.NoError(t, os.WriteFile(path, []byte("old text here"), 0644))

	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: tmpDir, Approved: true}, "edit",
		map[string]string{"path": "edit.go", "old": "old text", "new": "new text"})
	require.NoError(t, result.Err)

	data, _ := os.ReadFile(path)
	assert.Equal(t, "new text here", string(data))
}

func TestFileEditTool_FailsOnAmbiguousMatch(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "edit.go")
	require.NoError(t, os.WriteFile(path, []byte("dup dup"), 0644))

	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: tmpDir, Approved: true}, "edit",
		map[string]string{"path": "edit.go", "old": "dup", "new": "x"})
	assert.Error(t, result.Err)
}

func TestListTool_ListsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("x"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "sub"), 0755))

	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: tmpDir}, "list", map[string]string{"dir": "."})
	require.NoError(t, result.Err)
	assert.Contains(t, result.Output, "a.go")
	assert.Contains(t, result.Output, "sub/")
}

func TestBashTool_RunsCommand(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: "."}, "bash", map[string]string{"command": "echo hello"})
	require.NoError(t, result.Err)
	assert.Equal(t, "hello", result.Output)
}

func TestBashTool_DryRunDoesNotExecute(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: ".", DryRun: true}, "bash", map[string]string{"command": "echo should-not-run"})
	assert.Contains(t, result.Output, "[dry-run]")
}
