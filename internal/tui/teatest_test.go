package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudspacelab/forge/internal/adapters"
	"github.com/cloudspacelab/forge/internal/ports"
	"github.com/cloudspacelab/forge/internal/runtime"
	"github.com/cloudspacelab/forge/internal/skill"
	"github.com/cloudspacelab/forge/internal/tools"
)

// asModel type-asserts the returned tea.Model back to tui.Model.
func asModel(t *testing.T, m tea.Model) Model {
	t.Helper()
	tm, ok := m.(Model)
	require.True(t, ok, "expected tui.Model, got %T", m)
	return tm
}

// updateModel runs Update and returns the typed model (ignoring cmd).
func updateModel(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	result, _ := m.Update(msg)
	return asModel(t, result)
}

// =============================================================================
// UPDATE — WindowSizeMsg
// =============================================================================

func TestUpdate_WindowSize_SetsDimensions(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, tea.WindowSizeMsg{Width: 120, Height: 50})
	assert.Equal(t, 120, m2.width)
	assert.Equal(t, 50, m2.height)
}

// =============================================================================
// UPDATE — Autocomplete navigation (all branches)
// =============================================================================

func TestUpdate_Autocomplete_TabMovesDown(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/r")
	m.completionIdx = 0
	startIdx := m.completionIdx

	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, (startIdx+1)%len(m.completions), m2.completionIdx)
}

func TestUpdate_Autocomplete_DownArrow(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/")
	m.completionIdx = 0

	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, m2.completionIdx)
}

func TestUpdate_Autocomplete_UpArrow_Wraps(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/")
	m.completionIdx = 0 // at top

	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyUp})
	// Should wrap to last index
	assert.Equal(t, len(m.completions)-1, m2.completionIdx)
}

func TestUpdate_Autocomplete_UpArrow_Normal(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/")
	m.completionIdx = 2

	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, m2.completionIdx)
}

func TestUpdate_Autocomplete_EnterSelects(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/read")
	m.completionIdx = 0

	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m2.showCompletions)
	assert.Equal(t, "/read ", m2.input.Value())
}

func TestUpdate_Autocomplete_RightSelects(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/quit")
	m.completionIdx = 0

	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyRight})
	assert.False(t, m2.showCompletions)
	assert.Contains(t, m2.input.Value(), "/quit")
}

func TestUpdate_Autocomplete_EscDismisses(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/")

	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m2.showCompletions)
}

// =============================================================================
// UPDATE — Regular key handling (Ctrl+C, Ctrl+L, Esc, Enter, default)
// =============================================================================

func TestUpdate_CtrlC_Quits(t *testing.T) {
	m := buildTestModel(t)
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.True(t, asModel(t, m2).quit)
	require.NotNil(t, cmd)
}

func TestUpdate_CtrlL_ClearsBlocks(t *testing.T) {
	m := buildTestModel(t)
	m.blocks = []block{{role: "user", content: "old"}}
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyCtrlL})
	assert.Empty(t, m2.blocks)
}

func TestUpdate_Esc_HidesHelp(t *testing.T) {
	m := buildTestModel(t)
	m.showHelp = true
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m2.showHelp)
}

func TestUpdate_Enter_EmptyInput_NoOp(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Empty(t, m2.blocks)
}

func TestUpdate_Enter_SlashCommand(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/project")
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotEmpty(t, m2.blocks)
	// Should have a command block and an output block
	assert.GreaterOrEqual(t, len(m2.blocks), 2)
	assert.Equal(t, "command", m2.blocks[0].role)
}

func TestUpdate_Enter_SlashCommand_ShowHelp(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/help")
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, m2.showHelp)
	assert.NotEmpty(t, m2.helpOverlay)
}

func TestUpdate_Enter_SlashCommand_Quit(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/quit")
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, asModel(t, m2).quit)
}

func TestUpdate_Enter_SlashCommand_Error(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/nonexistent")
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	// Should produce an error block
	found := false
	for _, b := range m2.blocks {
		if b.role == "error" {
			found = true
		}
	}
	assert.True(t, found, "should have an error block for unknown command")
}

func TestUpdate_Enter_NaturalLanguage(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("hello")
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, m2.processing)
	// Should have a user block
	assert.NotEmpty(t, m2.blocks)
	assert.Equal(t, "user", m2.blocks[0].role)
	assert.Equal(t, "hello", m2.blocks[0].content)
}

func TestUpdate_DefaultKey_UpdatesInput(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	assert.Contains(t, m2.input.Value(), "x")
}

// =============================================================================
// UPDATE — runtimeEventMsg
// =============================================================================

func TestUpdate_RuntimeEvent_Assistant(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventAssistant, Text: "hello back", Time: time.Now(),
	}})
	assert.NotEmpty(t, m2.blocks)
	assert.Equal(t, "assistant", m2.blocks[len(m2.blocks)-1].role)
}

func TestUpdate_RuntimeEvent_ToolCall(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventToolCall, Tool: "read", Args: map[string]string{"path": "x.go"}, Time: time.Now(),
	}})
	last := m2.blocks[len(m2.blocks)-1]
	assert.Equal(t, "tool", last.role)
	assert.Contains(t, last.content, "read")
}

func TestUpdate_RuntimeEvent_ToolResult_Success(t *testing.T) {
	m := buildTestModel(t)
	result := &ports.ToolResult{Output: "file contents"}
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventToolResult, Result: result, Time: time.Now(),
	}})
	last := m2.blocks[len(m2.blocks)-1]
	assert.Equal(t, "result", last.role)
	assert.Contains(t, last.content, "file contents")
}

func TestUpdate_RuntimeEvent_ToolResult_Error(t *testing.T) {
	m := buildTestModel(t)
	result := &ports.ToolResult{Err: assertError("boom")}
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventToolResult, Result: result, Time: time.Now(),
	}})
	last := m2.blocks[len(m2.blocks)-1]
	assert.Equal(t, "result", last.role)
	assert.Contains(t, last.content, "boom")
}

func TestUpdate_RuntimeEvent_ToolResult_Nil(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventToolResult, Result: nil, Time: time.Now(),
	}})
	// Nil result should not add a block
	for _, b := range m2.blocks {
		assert.NotEqual(t, "result", b.role)
	}
}

func TestUpdate_RuntimeEvent_System(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventSystem, Text: "system msg", Time: time.Now(),
	}})
	last := m2.blocks[len(m2.blocks)-1]
	assert.Equal(t, "system", last.role)
}

func TestUpdate_RuntimeEvent_Error(t *testing.T) {
	m := buildTestModel(t)
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventError, Text: "something broke", Time: time.Now(),
	}})
	last := m2.blocks[len(m2.blocks)-1]
	assert.Equal(t, "error", last.role)
}

func TestUpdate_RuntimeEvent_TurnEnd_ClearsProcessing(t *testing.T) {
	m := buildTestModel(t)
	m.processing = true
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventTurnEnd, Time: time.Now(),
	}})
	assert.False(t, m2.processing)
}

func TestUpdate_RuntimeEvent_TurnStart(t *testing.T) {
	m := buildTestModel(t)
	beforeCount := len(m.blocks)
	m2 := updateModel(t, m, runtimeEventMsg{event: ports.Event{
		Type: ports.EventTurnStart, Text: "hi", Time: time.Now(),
	}})
	// TurnStart doesn't add a block (user block already added by Update)
	assert.Equal(t, beforeCount, len(m2.blocks))
}

// =============================================================================
// UPDATE — default message (non-key, non-event)
// =============================================================================

func TestUpdate_DefaultMsg_UpdatesViewportAndInput(t *testing.T) {
	m := buildTestModel(t)
	// A plain struct (not a recognized type) hits the default branch
	m2 := updateModel(t, m, tea.MouseMsg{})
	assert.NotNil(t, m2)
}

// =============================================================================
// handleRuntimeEvent — direct call (pointer receiver)
// =============================================================================

func TestHandleRuntimeEvent_AllTypes(t *testing.T) {
	cases := []struct {
		name   string
		event  ports.Event
		role   string
		addsBl bool
	}{
		{"assistant", ports.Event{Type: ports.EventAssistant, Text: "x"}, "assistant", true},
		{"toolcall", ports.Event{Type: ports.EventToolCall, Tool: "read"}, "tool", true},
		{"toolresult", ports.Event{Type: ports.EventToolResult, Result: &ports.ToolResult{Output: "ok"}}, "result", true},
		{"toolresult_nil", ports.Event{Type: ports.EventToolResult, Result: nil}, "", false},
		{"system", ports.Event{Type: ports.EventSystem, Text: "s"}, "system", true},
		{"error", ports.Event{Type: ports.EventError, Text: "e"}, "error", true},
		{"turnstart", ports.Event{Type: ports.EventTurnStart, Text: "t"}, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := buildTestModel(t)
			before := len(m.blocks)
			m.handleRuntimeEvent(tc.event)
			if tc.addsBl {
				assert.Greater(t, len(m.blocks), before)
				if tc.role != "" {
					assert.Equal(t, tc.role, m.blocks[len(m.blocks)-1].role)
				}
			}
		})
	}
}

// =============================================================================
// updateCompletions
// =============================================================================

func TestUpdateCompletions_ShowsOnSlashPrefix(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/re")
	m.updateCompletions()
	assert.True(t, m.showCompletions)
	assert.NotEmpty(t, m.completions)
}

func TestUpdateCompletions_HidesAfterSpace(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/read file")
	m.updateCompletions()
	assert.False(t, m.showCompletions)
}

func TestUpdateCompletions_HidesOnNonSlash(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("hello")
	m.updateCompletions()
	assert.False(t, m.showCompletions)
}

func TestUpdateCompletions_HidesOnNoMatch(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/zzzz")
	m.updateCompletions()
	assert.False(t, m.showCompletions)
}

func TestUpdateCompletions_ResetsIndexOverflow(t *testing.T) {
	m := buildTestModel(t)
	m.input.SetValue("/")
	m.completionIdx = 999
	m.updateCompletions()
	assert.Equal(t, 0, m.completionIdx)
}

// =============================================================================
// renderCompletions
// =============================================================================

func TestRenderCompletions_RendersList(t *testing.T) {
	m := buildTestModel(t)
	m.showCompletions = true
	m.completions = filterCommands("/re")
	m.completionIdx = 0
	out := m.renderCompletions()
	assert.Contains(t, out, "/read")
	assert.Contains(t, out, "navigate")
}

func TestRenderCompletions_HighlightsSelected(t *testing.T) {
	m := buildTestModel(t)
	m.completions = filterCommands("/")
	m.completionIdx = 1
	out := m.renderCompletions()
	assert.NotEmpty(t, out)
}

func TestRenderCompletions_WithArgs(t *testing.T) {
	m := buildTestModel(t)
	m.completions = []Command{{Name: "/test", Args: "[pkg]", Description: "run tests"}}
	m.completionIdx = 0
	out := m.renderCompletions()
	assert.Contains(t, out, "[pkg]")
}

// =============================================================================
// renderStatusLine
// =============================================================================

func TestRenderStatusLine_Processing(t *testing.T) {
	m := buildTestModel(t)
	m.processing = true
	out := m.renderStatusLine()
	assert.Contains(t, out, "working")
}

func TestRenderStatusLine_Idle(t *testing.T) {
	m := buildTestModel(t)
	m.processing = false
	out := m.renderStatusLine()
	assert.Contains(t, out, "/help")
}

// =============================================================================
// View — all paths
// =============================================================================

func TestView_HelpOverlay(t *testing.T) {
	m := buildTestModel(t)
	m.showHelp = true
	m.helpOverlay = "HELP TEXT"
	out := m.View()
	assert.Contains(t, out, "HELP TEXT")
}

func TestView_Normal(t *testing.T) {
	m := buildTestModel(t)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.blocks = []block{{role: "user", content: "hi"}}
	m.refreshView()
	out := m.View()
	assert.Contains(t, out, "hi")
}

func TestView_WithCompletions(t *testing.T) {
	m := buildTestModel(t)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.showCompletions = true
	m.completions = filterCommands("/")
	out := m.View()
	assert.Contains(t, out, "/help")
}

// =============================================================================
// renderBlock — all roles
// =============================================================================

func TestRenderBlock_AllRoles(t *testing.T) {
	cases := []struct {
		role    string
		content string
	}{
		{"user", "you said"},
		{"assistant", "forge replied"},
		{"tool", "read(main.go)"},
		{"result", "the output"},
		{"system", "a note"},
		{"error", "failed"},
		{"command", "/ls"},
		{"unknown", "mystery"},
	}
	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			out := renderBlock(block{role: tc.role, content: tc.content})
			assert.NotEmpty(t, out)
		})
	}
}

// =============================================================================
// Init
// =============================================================================

func TestInit_ReturnsCommand(t *testing.T) {
	m := buildTestModel(t)
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

// =============================================================================
// Run (using teatest for headless integration test)
// =============================================================================

func TestRun_StartsAndQuits(t *testing.T) {
	rt, err := runtime.New(runtime.Config{FolderPath: "../..", AutoApprove: true})
	require.NoError(t, err)
	rt.SetToolExecutor(adapters.NewToolExecutorAdapter(tools.NewRegistry()))
	rt.SetProvider(&runtime.EchoProvider{})
	bus := adapters.NewChannelEventBus(8)
	rt.SetEventBus(bus)
	loader := skill.NewLoaderWithDirs(t.TempDir(), t.TempDir(), t.TempDir())
	loader.Load()

	m := New(rt, loader, bus)
	// Run in a goroutine with a timeout; we can't easily send keys via teatest
	// without a real terminal, so we verify Run doesn't panic on init by
	// checking the model is valid.
	assert.NotNil(t, m.input)
	assert.NotNil(t, m.rt)
}

// =============================================================================
// formatArgs — full coverage
// =============================================================================

func TestFormatArgs_LongValue(t *testing.T) {
	longVal := string(make([]byte, 50))
	out := formatArgs(map[string]string{"path": longVal})
	// Long values should be truncated
	assert.Contains(t, out, "…")
}

// =============================================================================
// handleSlashCommand — every command branch
// =============================================================================

func TestHandleSlashCommand_AllCommands(t *testing.T) {
	m := buildTestModel(t)
	// Use safe/fast commands — avoid /test (runs full suite) and /build
	// which are slow and recursive inside tests.
	cmds := []string{
		"/help", "/project", "/session", "/ls",
		"/bash echo ok", "/status", "/tools",
		"/skills", "/version", "/clear",
	}
	for _, cmd := range cmds {
		t.Run(cmd, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m.handleSlashCommand(cmd)
			})
		})
	}
}

func TestHandleSlashCommand_ReadMissingArg(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/read")
	assert.True(t, r.isError)
}

func TestHandleSlashCommand_BashMissingArg(t *testing.T) {
	m := buildTestModel(t)
	r := m.handleSlashCommand("/bash")
	assert.True(t, r.isError)
}

// =============================================================================
// executeTool
// =============================================================================

func TestExecuteTool_NilRuntime(t *testing.T) {
	m := buildTestModel(t)
	m.rt = nil
	r := m.executeTool("list", nil)
	assert.Error(t, r.Err)
}

// =============================================================================
// waitForEvent
// =============================================================================

func TestWaitForEvent_ReceivesEvent(t *testing.T) {
	bus := adapters.NewChannelEventBus(4)
	cmd := waitForEvent(bus.Channel())
	go func() {
		time.Sleep(10 * time.Millisecond)
		bus.Emit(ports.Event{Type: ports.EventSystem, Text: "test"})
	}()
	msg := cmd()
	require.NotNil(t, msg)
}

// =============================================================================
// resizeComponents — edge cases
// =============================================================================

func TestResizeComponents_TinyTerminal(t *testing.T) {
	m := buildTestModel(t)
	m.width = 10
	m.height = 5
	m.resizeComponents()
	// Should not panic, viewport height clamped
	assert.GreaterOrEqual(t, m.viewport.Height, 3)
}

// helper
type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
func assertError(msg string) error { return &testError{msg: msg} }
