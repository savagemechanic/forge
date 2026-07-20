// Package runtime is the APPLICATION layer (a hexagonal use-case
// orchestrator). It depends ONLY on ports and domain — never on concrete
// adapters. Adapters (tools, providers, storage) are injected at the
// composition root (cmd/forge) via the port interfaces.
package runtime

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudspacelab/forge/internal/folder"
	"github.com/cloudspacelab/forge/internal/memory"
	"github.com/cloudspacelab/forge/internal/ports"
	"github.com/cloudspacelab/forge/internal/session"
	"github.com/cloudspacelab/forge/internal/skill"
)

// Config holds runtime configuration.
type Config struct {
	FolderPath   string
	AutoApprove  bool
	MaxToolCalls int
}

// Runtime is the core agent orchestrator. It wires together the domain
// (folder, session, memory, skill) with injected ports (provider,
// toolExecutor, events, approver). It knows nothing about how those
// ports are implemented.
type Runtime struct {
	cfg      Config
	project  *folder.ProjectInfo
	session  *session.Session
	provider ports.Provider
	toolExec ports.ToolExecutor
	events   ports.EventBus
	approver ports.Approver
	sessions ports.SessionRepository
	memStore *memory.InMemoryMemoryEntryStore
	skills   *skill.Loader

	// mu protects session.Messages and conv, which are written in the
	// Submit() goroutine and read from the driving adapter's goroutine.
	mu sync.RWMutex

	conv []ports.ProviderMessage
}

// New creates a Runtime. Domain objects are created internally; ports
// (provider, tools, events, approver, sessions) are injected afterwards
// via the Set* methods. This keeps the application layer adapter-free.
func New(cfg Config) (*Runtime, error) {
	project, err := folder.Discover(cfg.FolderPath)
	if err != nil {
		return nil, fmt.Errorf("discover folder: %w", err)
	}

	now := time.Now()
	sess := &session.Session{
		ID:           generateID(),
		FolderID:     folderID(project.RootPath),
		State:        session.SessionStateActive,
		Status:       session.SessionStateActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastActiveAt: now,
		Messages:     []session.Message{},
	}

	rt := &Runtime{
		cfg:      cfg,
		project:  project,
		session:  sess,
		memStore: memory.NewMemoryEntryStore(),
	}

	if rt.cfg.MaxToolCalls == 0 {
		rt.cfg.MaxToolCalls = 20
	}

	return rt, nil
}

// ---- PORT INJECTION (composition root calls these) ----

func (rt *Runtime) SetProvider(p ports.Provider)              { rt.provider = p }
func (rt *Runtime) SetToolExecutor(t ports.ToolExecutor)      { rt.toolExec = t }
func (rt *Runtime) SetEventBus(e ports.EventBus)              { rt.events = e }
func (rt *Runtime) SetApprover(a ports.Approver)              { rt.approver = a }
func (rt *Runtime) SetSessionRepo(s ports.SessionRepository)  { rt.sessions = s }
func (rt *Runtime) SetSkills(s *skill.Loader)                 { rt.skills = s }

// ---- DOMAIN ACCESSORS (read-only views) ----

func (rt *Runtime) Project() *folder.ProjectInfo { return rt.project }

// Session returns the active session. Callers must not read .Messages
// concurrently with Submit(); use MessageCount or SessionSnapshot instead.
func (rt *Runtime) Session() *session.Session { return rt.session }

// MessageCount returns the number of messages, thread-safe.
func (rt *Runtime) MessageCount() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return len(rt.session.Messages)
}

// SessionSnapshot returns a copy of the session's scalar fields and a
// snapshot of the message count, safe to inspect from any goroutine.
func (rt *Runtime) SessionSnapshot() (id, folderID, state string, msgCount int) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.session.ID, rt.session.FolderID, string(rt.session.State), len(rt.session.Messages)
}

// SystemPrompt builds the system prompt from domain context.
func (rt *Runtime) SystemPrompt() string {
	var b strings.Builder
	b.WriteString("You are Forge, a terminal-native coding agent.\n\n")
	b.WriteString("You operate inside the user's project folder. You can read files, ")
	b.WriteString("run commands, and propose changes (with approval).\n\n")

	b.WriteString("PROJECT CONTEXT:\n")
	fmt.Fprintf(&b, "  Root: %s\n", rt.project.RootPath)
	if rt.project.GoModule != "" {
		fmt.Fprintf(&b, "  Module: %s\n", rt.project.GoModule)
	}
	if rt.project.GitRoot != "" {
		b.WriteString("  Git: yes\n")
	}
	if len(rt.project.Subpackages) > 0 {
		limit := min(len(rt.project.Subpackages), 10)
		fmt.Fprintf(&b, "  Packages: %s\n", strings.Join(rt.project.Subpackages[:limit], ", "))
		if len(rt.project.Subpackages) > 10 {
			fmt.Fprintf(&b, "    ... and %d more\n", len(rt.project.Subpackages)-10)
		}
	}
	b.WriteString("\n")

	if rt.skills != nil {
		active := rt.skills.Active()
		if len(active) > 0 {
			b.WriteString("ACTIVE SKILLS:\n")
			for _, s := range active {
				fmt.Fprintf(&b, "  - %s: %s\n", s.Name, s.Description)
			}
			b.WriteString("\n")
		}
	}

	if rt.toolExec != nil {
		b.WriteString("AVAILABLE TOOLS:\n")
		for _, t := range rt.toolExec.List() {
			fmt.Fprintf(&b, "  - %s: %s\n", t.Name, t.Description)
		}
	}

	return b.String()
}

// Submit processes a user message through the agent loop, emitting events
// to the injected EventBus. The driving adapter (TUI) runs this in a
// goroutine and consumes events.
func (rt *Runtime) Submit(userText string) {
	rt.emit(ports.Event{Type: ports.EventTurnStart, Text: userText, Time: time.Now()})

	now := time.Now()
	rt.mu.Lock()
	rt.session.Messages = append(rt.session.Messages, session.Message{
		ID:        generateID(),
		Role:      "user",
		Content:   userText,
		CreatedAt: now,
	})
	rt.conv = append(rt.conv, ports.ProviderMessage{Role: "user", Content: userText})
	rt.mu.Unlock()

	if rt.provider == nil {
		rt.emit(ports.Event{
			Type: ports.EventSystem,
			Text: "No provider configured. Slash commands still work — type /help.",
			Time: time.Now(),
		})
		rt.emit(ports.Event{Type: ports.EventTurnEnd, Time: time.Now()})
		return
	}

	toolCallCount := 0
	for {
		if toolCallCount >= rt.cfg.MaxToolCalls {
			rt.emit(ports.Event{Type: ports.EventError, Text: "max tool calls reached", Time: time.Now()})
			break
		}

		resp := rt.provider.Generate(ports.ProviderRequest{
			SystemPrompt: rt.SystemPrompt(),
			Messages:     rt.conv,
			Tools:        rt.toolSpecs(),
		})

		if resp.Text != "" {
			rt.emit(ports.Event{Type: ports.EventAssistant, Text: resp.Text, Time: time.Now()})
			rt.conv = append(rt.conv, ports.ProviderMessage{Role: "assistant", Content: resp.Text})
		}

		if len(resp.ToolCalls) == 0 || resp.Stop {
			break
		}

		for _, tc := range resp.ToolCalls {
			toolCallCount++
			rt.emit(ports.Event{Type: ports.EventToolCall, Tool: tc.Name, Args: tc.Args, Time: time.Now()})

			result := rt.executeTool(tc)
			tr := ports.ToolResult{
				Tool: result.Tool, Output: result.Output, Err: result.Err, Changed: result.Changed,
			}
			rt.emit(ports.Event{Type: ports.EventToolResult, Tool: tc.Name, Result: &tr, Time: time.Now()})

			rt.conv = append(rt.conv, ports.ProviderMessage{
				Role: "tool", Content: result.Output, ToolName: tc.Name,
			})
		}
	}

	rt.emit(ports.Event{Type: ports.EventTurnEnd, Time: time.Now()})
}

// executeTool dispatches a tool call through the injected ToolExecutor
// port, applying the approval gate for state-changing tools.
func (rt *Runtime) executeTool(tc ports.ToolCall) ports.ToolResult {
	needsApproval := tc.Name == "write" || tc.Name == "edit"
	approved := rt.cfg.AutoApprove || !needsApproval

	if needsApproval && !rt.cfg.AutoApprove && rt.approver != nil {
		resp := rt.approver.RequestApproval(ports.ApprovalRequest{
			Tool:    tc.Name,
			Summary: fmt.Sprintf("%s %s", tc.Name, tc.Args["path"]),
			Files:   []string{tc.Args["path"]},
		})
		approved = resp.Approved
	}

	ctx := &ports.ToolContext{
		WorkDir:  rt.project.RootPath,
		Approved: approved,
	}
	return rt.toolExec.Execute(ctx, tc.Name, tc.Args)
}

// ExecuteToolDirect runs a tool without going through the provider loop.
// Used by slash commands in the driving adapter.
func (rt *Runtime) ExecuteToolDirect(name string, args map[string]string) ports.ToolResult {
	ctx := &ports.ToolContext{WorkDir: rt.project.RootPath, Approved: rt.cfg.AutoApprove || true}
	return rt.toolExec.Execute(ctx, name, args)
}

func (rt *Runtime) toolSpecs() []ports.ToolSpec {
	if rt.toolExec == nil {
		return nil
	}
	return rt.toolExec.List()
}

// ToolSpecs returns the available tool specs (for driving adapters).
func (rt *Runtime) ToolSpecs() []ports.ToolSpec {
	return rt.toolSpecs()
}

func (rt *Runtime) emit(e ports.Event) {
	if rt.events != nil {
		rt.events.Emit(e)
	}
}
