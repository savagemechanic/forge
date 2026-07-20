package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: All() returns every registered tool; Register adds custom tools.

func TestRegistry_All(t *testing.T) {
	r := NewRegistry()
	all := r.All()
	assert.NotEmpty(t, all)
	// Should have at least the 6 built-in tools
	assert.GreaterOrEqual(t, len(all), 6)
}

func TestRegistry_RegisterCustomTool(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{})
	got, ok := r.Get("mock")
	require.True(t, ok)
	assert.Equal(t, "mock tool", got.Description())
}

func TestRegistry_GetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	assert.False(t, ok)
}

// INVARIANT: every tool's Description() returns a non-empty string.

func TestAllTools_HaveDescriptions(t *testing.T) {
	r := NewRegistry()
	for _, tool := range r.All() {
		assert.NotEmpty(t, tool.Description(), "tool %s has empty description", tool.Name())
		assert.NotEmpty(t, tool.Name(), "a tool has an empty name")
	}
}

// INVARIANT: FileReadTool returns error on missing file.

func TestFileReadTool_MissingFile(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: t.TempDir()}, "read", map[string]string{"path": "nope.go"})
	assert.Error(t, result.Err)
}

// INVARIANT: FileReadTool errors on missing path arg.

func TestFileReadTool_MissingArg(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: "."}, "read", map[string]string{})
	assert.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "path")
}

// INVARIANT: FileWriteTool dry-run doesn't write.

func TestFileWriteTool_DryRun(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: dir, Approved: true, DryRun: true}, "write",
		map[string]string{"path": "x.txt", "content": "data"})
	require.NoError(t, result.Err)
	assert.Contains(t, result.Output, "dry-run")
}

// INVARIANT: FileEditTool errors on missing old text.

func TestFileEditTool_MissingOldText(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: t.TempDir(), Approved: true}, "edit",
		map[string]string{"path": "x"})
	assert.Error(t, result.Err)
}

// INVARIANT: BashTool errors on missing command.

func TestBashTool_MissingCommand(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: "."}, "bash", map[string]string{})
	assert.Error(t, result.Err)
}

// INVARIANT: BashTool captures stderr.

func TestBashTool_CapturesStderr(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: "."}, "bash",
		map[string]string{"command": "echo error >&2"})
	// Should capture the stderr output
	assert.Contains(t, result.Output, "error")
}

// INVARIANT: ListTool defaults to current dir and works on a dir.

func TestListTool_DefaultDir(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: dir}, "list", map[string]string{})
	assert.NoError(t, result.Err)
}

// INVARIANT: GitStatusTool errors outside a git repo.

func TestGitStatusTool_NonGitDir(t *testing.T) {
	r := NewRegistry()
	result := r.Execute(&Context{WorkDir: t.TempDir()}, "git_status", map[string]string{})
	assert.Error(t, result.Err)
}

// mockTool for Register test.
type mockTool struct{}

func (m *mockTool) Name() string        { return "mock" }
func (m *mockTool) Description() string { return "mock tool" }
func (m *mockTool) Execute(ctx *Context, args map[string]string) Result {
	return Result{Tool: "mock", Output: "mocked"}
}
