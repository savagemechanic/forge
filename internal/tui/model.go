// Package tui is a DRIVING adapter — it consumes the application layer
// (runtime) and renders it via Bubble Tea. It depends on ports +
// application, never on concrete driven adapters directly (those are
// injected by the composition root in cmd/forge).
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cloudspacelab/forge/internal/goengine"
	"github.com/cloudspacelab/forge/internal/ports"
	"github.com/cloudspacelab/forge/internal/runtime"
	"github.com/cloudspacelab/forge/internal/skill"
)

// block is a single rendered line-group in the transcript.
type block struct {
	role    string // "user", "assistant", "tool", "result", "system", "error", "command"
	content string
}

// Model is the Bubble Tea model for the Forge TUI.
type Model struct {
	rt          *runtime.Runtime
	skillLoader *skill.Loader
	bus         interface {
		Channel() <-chan ports.Event
	}

	// UI components
	viewport viewport.Model
	input    textinput.Model

	// state
	blocks      []block
	width       int
	height      int
	processing  bool
	quit        bool
	showHelp    bool
	helpOverlay string

	// Go intelligence index (lazily built)
	goIndex *goengine.Index
}

// runtimeEventMsg wraps a runtime event for Bubble Tea.
type runtimeEventMsg struct{ event ports.Event }

// New creates the TUI model with the given runtime and event bus.
func New(rt *runtime.Runtime, loader *skill.Loader, bus interface{ Channel() <-chan ports.Event }) Model {
	vp := viewport.New(80, 20)
	vp.SetContent("")

	ti := textinput.New()
	ti.Placeholder = "Type a message or /help for commands..."
	ti.Prompt = "❯ "
	ti.PromptStyle = composerStyle
	ti.CharLimit = 0
	ti.Focus()

	return Model{
		rt:          rt,
		skillLoader: loader,
		bus:         bus,
		viewport:    vp,
		input:       ti,
	}
}

// Init starts the Bubble Tea program — we listen for runtime events.
func (m Model) Init() tea.Cmd {
	return waitForEvent(m.bus.Channel())
}

// waitForEvent is a tea.Cmd that blocks on the event channel and returns
// a runtimeEventMsg when one arrives. Re-issued in Update to keep listening.
func waitForEvent(ch <-chan ports.Event) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-ch
		if !ok {
			return nil
		}
		return runtimeEventMsg{event: e}
	}
}

// Update handles all messages: key presses, runtime events, window size.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeComponents()
		m.refreshView()
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quit = true
			return m, tea.Quit
		case tea.KeyCtrlL:
			m.blocks = nil
			m.refreshView()
			return m, nil
		case tea.KeyEsc:
			m.showHelp = false
			m.refreshView()
			return m, nil
		case tea.KeyEnter:
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			m.input.Reset()

			// Slash command?
			if strings.HasPrefix(text, "/") {
				result := m.handleSlashCommand(text)
				if result.quit {
					m.quit = true
					return m, tea.Quit
				}
				if result.showHelp {
					m.showHelp = true
					m.helpOverlay = result.output
				} else if result.output != "" {
					label := "command"
					if result.isError {
						label = "error"
					}
					m.blocks = append(m.blocks, block{role: "command", content: text})
					m.blocks = append(m.blocks, block{role: label, content: result.output})
				}
				m.refreshView()
				return m, nil
			}

			// Natural language → submit to runtime
			m.blocks = append(m.blocks, block{role: "user", content: text})
			m.processing = true
			m.refreshView()

			go m.rt.Submit(text)
			cmds = append(cmds, waitForEvent(m.bus.Channel()))
			return m, tea.Batch(cmds...)

		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case runtimeEventMsg:
		m.handleRuntimeEvent(msg.event)
		if msg.event.Type == ports.EventTurnEnd {
			m.processing = false
		}
		if !m.quit {
			cmds = append(cmds, waitForEvent(m.bus.Channel()))
		}
		m.refreshView()
		return m, tea.Batch(cmds...)
	}

	// Default: update viewport + input
	var vpCmd, tiCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.input, tiCmd = m.input.Update(msg)
	return m, tea.Batch(vpCmd, tiCmd)
}

// handleRuntimeEvent appends the appropriate block for an event.
func (m *Model) handleRuntimeEvent(e ports.Event) {
	switch e.Type {
	case ports.EventTurnStart:
		// user block already added by Update
	case ports.EventAssistant:
		m.blocks = append(m.blocks, block{role: "assistant", content: e.Text})
	case ports.EventToolCall:
		var b strings.Builder
		fmt.Fprintf(&b, "%s(%s)", e.Tool, formatArgs(e.Args))
		m.blocks = append(m.blocks, block{role: "tool", content: b.String()})
	case ports.EventToolResult:
		if e.Result != nil {
			content := e.Result.Output
			if e.Result.Err != nil {
				content = e.Result.Err.Error()
			}
			m.blocks = append(m.blocks, block{role: "result", content: content})
		}
	case ports.EventSystem:
		m.blocks = append(m.blocks, block{role: "system", content: e.Text})
	case ports.EventError:
		m.blocks = append(m.blocks, block{role: "error", content: e.Text})
	}
}

// View renders the entire screen.
func (m Model) View() string {
	if m.showHelp && m.helpOverlay != "" {
		return boxStyle.Render(m.helpOverlay)
	}

	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Transcript
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Status line
	b.WriteString(m.renderStatusLine())
	b.WriteString("\n")

	// Composer
	b.WriteString(m.input.View())

	return b.String()
}

func (m Model) renderHeader() string {
	var parts []string
	if m.rt != nil {
		p := m.rt.Project()
		name := pathBase(p.RootPath)
		parts = append(parts, "🔥 "+name)
		if p.GoModule != "" {
			parts = append(parts, dimStyle.Render(p.GoModule))
		}
		parts = append(parts, dimStyle.Render(p.Summary()))
	}
	return headerStyle.Render(strings.Join(parts, "  "))
}

func (m Model) renderStatusLine() string {
	status := hintStyle.Render("  /help for commands · Ctrl+C quit · ↑↓ scroll")
	if m.processing {
		status = assistantLabelStyle.Render("  ⟳ working...") + hintStyle.Render("  /help · Ctrl+C")
	}
	return status
}

func (m *Model) resizeComponents() {
	headerH := 2
	statusH := 2
	composerH := 1
	vpHeight := m.height - headerH - statusH - composerH
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewport.Width = m.width
	m.viewport.Height = vpHeight
	m.input.Width = m.width - 4
}

func (m *Model) refreshView() {
	var b strings.Builder
	for _, blk := range m.blocks {
		b.WriteString(renderBlock(blk))
		b.WriteString("\n")
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

// renderBlock formats a single transcript block with role styling.
func renderBlock(blk block) string {
	switch blk.role {
	case "user":
		return userLabelStyle.Render("you") + " " + userContentStyle.Render(blk.content)
	case "assistant":
		return assistantLabelStyle.Render("forge") + "\n" + indent(assistantContentStyle.Render(blk.content), "  ")
	case "tool":
		return "  " + toolLabelStyle.Render("→ "+blk.content)
	case "result":
		return "  " + indent(resultContentStyle.Render(blk.content), "    ")
	case "system":
		return systemLabelStyle.Render("· ") + systemContentStyle.Render(blk.content)
	case "error":
		return errorLabelStyle.Render("✗ ") + errorContentStyle.Render(blk.content)
	case "command":
		return toolLabelStyle.Render("$ ") + toolContentStyle.Render(blk.content)
	default:
		return blk.content
	}
}

// formatArgs renders tool args compactly.
func formatArgs(args map[string]string) string {
	if len(args) == 0 {
		return ""
	}
	var parts []string
	for k, v := range args {
		val := v
		if len(val) > 40 {
			val = val[:40] + "…"
		}
		parts = append(parts, fmt.Sprintf("%s=%q", k, val))
	}
	return strings.Join(parts, ", ")
}

// indent prefixes each line with the given prefix.
func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

func pathBase(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[i+1:]
		}
	}
	return p
}

// executeTool runs a tool directly (for slash commands) via the runtime.
func (m *Model) executeTool(name string, args map[string]string) ports.ToolResult {
	if m.rt == nil {
		return ports.ToolResult{Err: fmt.Errorf("no runtime")}
	}
	return m.rt.ExecuteToolDirect(name, args)
}

// Run starts the Bubble Tea program.
func Run(m Model) error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
