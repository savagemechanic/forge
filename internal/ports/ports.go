// Package ports defines the application-level port interfaces (the
// "hexagon" boundary). The application layer (runtime) depends only on
// these interfaces; concrete adapters implement them.
//
// Domain-level ports already live inside each domain package
// (e.g. session.SessionStore, skill.SkillStore, memory.MemoryEntryStore,
// storage.GraphStore). This package adds the application-orchestration
// ports that span multiple domains.
package ports

import (
	"time"
)

// ----------------------------------------------------------------------------
// PROVIDER PORT — abstracts the language model
// ----------------------------------------------------------------------------

// ProviderRequest is the input to a language model provider.
type ProviderRequest struct {
	SystemPrompt string
	Messages     []ProviderMessage
	Tools        []ToolSpec
}

// ProviderMessage is a single message in the provider conversation.
type ProviderMessage struct {
	Role     string // "user", "assistant", "tool"
	Content  string
	ToolName string
}

// ProviderResponse is the output from a provider.
type ProviderResponse struct {
	Text      string
	ToolCalls []ToolCall
	Stop      bool
}

// ToolCall is a request from the provider to execute a tool.
type ToolCall struct {
	Name string
	Args map[string]string
}

// ToolSpec describes a tool to the provider.
type ToolSpec struct {
	Name        string
	Description string
}

// Provider is the port for language model adapters.
type Provider interface {
	Name() string
	Generate(req ProviderRequest) ProviderResponse
}

// ----------------------------------------------------------------------------
// TOOL EXECUTOR PORT — abstracts tool execution & approval
// ----------------------------------------------------------------------------

// ToolResult is the outcome of a tool execution.
type ToolResult struct {
	Tool    string
	Output  string
	Err     error
	Changed []string
}

// ToolContext carries execution-scoped state.
type ToolContext struct {
	WorkDir  string
	Approved bool
	DryRun   bool
}

// ToolExecutor is the port for the tool/policy layer.
// It runs a named tool with arguments and returns the result.
type ToolExecutor interface {
	Execute(ctx *ToolContext, name string, args map[string]string) ToolResult
	List() []ToolSpec
}

// ----------------------------------------------------------------------------
// EVENT SINK PORT — abstracts where runtime events go (TUI, log, etc.)
// ----------------------------------------------------------------------------

// EventType identifies a runtime event.
type EventType string

const (
	EventTurnStart   EventType = "turn_start"
	EventAssistant   EventType = "assistant"
	EventToolCall    EventType = "tool_call"
	EventToolResult  EventType = "tool_result"
	EventApprovalReq EventType = "approval_request"
	EventTurnEnd     EventType = "turn_end"
	EventError       EventType = "error"
	EventSystem      EventType = "system"
)

// Event is a single update streamed from the application to a driving adapter.
type Event struct {
	Type   EventType
	Text   string
	Tool   string
	Args   map[string]string
	Result *ToolResult
	Time   time.Time
}

// EventBus is the port for streaming events to driving adapters.
type EventBus interface {
	Emit(e Event)
}

// ----------------------------------------------------------------------------
// APPROVAL PORT — abstracts the human-in-the-loop approval gate
// ----------------------------------------------------------------------------

// ApprovalRequest describes a change that needs human approval.
type ApprovalRequest struct {
	Tool     string
	Summary  string
	Files    []string
	Diff     string
}

// ApprovalResponse is the human's decision.
type ApprovalResponse struct {
	Approved bool
	Reason   string
}

// Approver is the port for the approval gate (interactive or auto).
type Approver interface {
	RequestApproval(req ApprovalRequest) ApprovalResponse
}

// ----------------------------------------------------------------------------
// SESSION REPOSITORY PORT — abstracts session persistence
// ----------------------------------------------------------------------------

// SessionRecord is the persistable view of a session.
type SessionRecord struct {
	ID         string
	FolderID   string
	ParentID   string
	State      string
	Messages   []SessionMessage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SessionMessage is a persistable message.
type SessionMessage struct {
	ID        string
	Role      string
	Content   string
	CreatedAt time.Time
}

// SessionRepository is the port for session persistence adapters.
type SessionRepository interface {
	Save(s *SessionRecord) error
	Get(id string) (*SessionRecord, error)
	ListByFolder(folderID string) ([]*SessionRecord, error)
}
