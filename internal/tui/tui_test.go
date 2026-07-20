package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/cloudspacelab/forge/internal/adapters"
	"github.com/cloudspacelab/forge/internal/runtime"
	"github.com/cloudspacelab/forge/internal/skill"
	"github.com/cloudspacelab/forge/internal/tools"
)

// buildTestModel creates a Model wired to real adapters for headless testing.
func buildTestModel(t *testing.T) Model {
	t.Helper()
	rt, err := runtime.New(runtime.Config{FolderPath: "../..", AutoApprove: true})
	if err != nil {
		t.Fatal(err)
	}
	rt.SetToolExecutor(adapters.NewToolExecutorAdapter(tools.NewRegistry()))
	rt.SetProvider(&runtime.EchoProvider{})
	bus := adapters.NewChannelEventBus(8)
	rt.SetEventBus(bus)
	loader := skill.NewLoaderWithDirs(t.TempDir(), t.TempDir(), t.TempDir())
	loader.Load()
	return New(rt, loader, bus)
}

// INVARIANT: filterCommands returns all commands on empty prefix, filters
// on a prefix, and returns nothing for a non-matching prefix.

func TestFilterCommands_Empty(t *testing.T) {
	all := filterCommands("")
	assert.NotEmpty(t, all)
	assert.GreaterOrEqual(t, len(all), 15)
}

func TestFilterCommands_Prefix(t *testing.T) {
	got := filterCommands("/re")
	assert.NotEmpty(t, got)
	for _, c := range got {
		assert.True(t, len(c.Name) >= 3 && c.Name[:3] == "/re" || c.Shortcut[:3] == "/re")
	}
}

func TestFilterCommands_NoMatch(t *testing.T) {
	got := filterCommands("/zzzzz")
	assert.Empty(t, got)
}

func TestFilterCommands_ExactCommand(t *testing.T) {
	got := filterCommands("/read")
	assert.Len(t, got, 1)
	assert.Equal(t, "/read", got[0].Name)
}

// INVARIANT: command_registry has all expected commands.

func TestCommandRegistry_HasAllCommands(t *testing.T) {
	names := map[string]bool{}
	for _, c := range commands {
		names[c.Name] = true
	}
	required := []string{"/help", "/read", "/bash", "/test", "/build", "/status",
		"/skills", "/quit", "/clear", "/project", "/ls"}
	for _, r := range required {
		assert.True(t, names[r], "missing command: %s", r)
	}
}

// INVARIANT: handleSlashCommand dispatches to the right handler.

func TestHandleSlashCommand_Help(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/help")
	assert.True(t, r.showHelp)
	assert.Contains(t, r.output, "FORGE")
}

func TestHandleSlashCommand_Project(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/project")
	assert.NotEmpty(t, r.output)
	assert.Contains(t, r.output, "Root")
}

func TestHandleSlashCommand_Quit(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/quit")
	assert.True(t, r.quit)
}

func TestHandleSlashCommand_Unknown(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/nonexistent")
	assert.True(t, r.isError)
}

func TestHandleSlashCommand_Ls(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/ls")
	assert.False(t, r.isError)
}

func TestHandleSlashCommand_Version(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/version")
	assert.Contains(t, r.output, "Forge")
}

func TestHandleSlashCommand_Clear(t *testing.T) {
	m := buildTestModel(t)
	m.blocks = []block{{role: "user", content: "x"}}
	r := m.handleSlashCommand("/clear")
	assert.Empty(t, m.blocks)
	_ = r
}

func TestHandleSlashCommand_Skills(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/skills")
	assert.NotEmpty(t, r.output)
}

// INVARIANT: Bubble Tea model Update handles key events without panicking.

func TestModel_Update_WindowSize(t *testing.T) {
	m := buildTestModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	assert.NotNil(t, updated)
}

func TestModel_Update_CtrlC(t *testing.T) {
	m := buildTestModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	// Ctrl+C should produce a Quit command
	assert.NotNil(t, cmd)
}

func TestModel_View_Renders(t *testing.T) {
	m := buildTestModel(t)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := m.View()
	assert.NotEmpty(t, view)
}

// INVARIANT: renderBlock formats each role correctly.

func TestRenderBlock_User(t *testing.T) {
	out := renderBlock(block{role: "user", content: "hi"})
	assert.Contains(t, out, "hi")
}

func TestRenderBlock_Assistant(t *testing.T) {
	out := renderBlock(block{role: "assistant", content: "hello"})
	assert.Contains(t, out, "hello")
}

func TestRenderBlock_Error(t *testing.T) {
	out := renderBlock(block{role: "error", content: "bad"})
	assert.Contains(t, out, "bad")
}

func TestRenderBlock_System(t *testing.T) {
	out := renderBlock(block{role: "system", content: "note"})
	assert.Contains(t, out, "note")
}

// INVARIANT: helper functions work.

func TestPathBase(t *testing.T) {
	assert.Equal(t, "forge", pathBase("/Users/x/forge"))
	assert.Equal(t, "forge", pathBase("forge"))
	assert.Equal(t, "dir", pathBase("/a/b/dir/"))
}

func TestFormatArgs(t *testing.T) {
	assert.Equal(t, "", formatArgs(nil))
	out := formatArgs(map[string]string{"path": "x.go"})
	assert.Contains(t, out, "path")
	assert.Contains(t, out, "x.go")
}

func TestYesNo(t *testing.T) {
	assert.Equal(t, "yes", yesno(true))
	assert.Equal(t, "no", yesno(false))
}

func TestIndent(t *testing.T) {
	got := indent("a\nb", "  ")
	assert.Equal(t, "  a\n  b", got)
}
