// Package adapters contains the DRIVEN adapters — concrete implementations
// of the ports defined in internal/ports. These translate between the
// application's port interfaces and real infrastructure (filesystem,
// shell, channels, storage).
//
// Driving adapters (TUI, CLI) live in their own packages and consume
// the application + ports directly.
package adapters

import (
	"github.com/cloudspacelab/forge/internal/ports"
	"github.com/cloudspacelab/forge/internal/tools"
)

// ----------------------------------------------------------------------------
// TOOL EXECUTOR ADAPTER
// Wraps the concrete tools.Registry to satisfy ports.ToolExecutor.
// ----------------------------------------------------------------------------

// ToolExecutorAdapter adapts tools.Registry to the ports.ToolExecutor port.
type ToolExecutorAdapter struct {
	registry *tools.Registry
}

// NewToolExecutorAdapter creates an adapter around a tools.Registry.
func NewToolExecutorAdapter(r *tools.Registry) *ToolExecutorAdapter {
	return &ToolExecutorAdapter{registry: r}
}

// Execute runs a named tool, translating types between port and concrete.
func (a *ToolExecutorAdapter) Execute(ctx *ports.ToolContext, name string, args map[string]string) ports.ToolResult {
	toolCtx := &tools.Context{
		WorkDir:  ctx.WorkDir,
		Approved: ctx.Approved,
		DryRun:   ctx.DryRun,
	}
	r := a.registry.Execute(toolCtx, name, args)
	return ports.ToolResult{
		Tool:    r.Tool,
		Output:  r.Output,
		Err:     r.Err,
		Changed: r.Changed,
	}
}

// List returns tool specs for all registered tools.
func (a *ToolExecutorAdapter) List() []ports.ToolSpec {
	all := a.registry.All()
	specs := make([]ports.ToolSpec, len(all))
	for i, t := range all {
		specs[i] = ports.ToolSpec{Name: t.Name(), Description: t.Description()}
	}
	return specs
}

// ----------------------------------------------------------------------------
// CHANNEL EVENT BUS ADAPTER
// Implements ports.EventBus using a buffered channel. The driving
// adapter (TUI) reads from the channel via Messages().
// ----------------------------------------------------------------------------

// ChannelEventBus emits events to a channel that Bubble Tea can consume.
type ChannelEventBus struct {
	ch chan ports.Event
}

// NewChannelEventBus creates a bus with the given buffer size.
func NewChannelEventBus(bufSize int) *ChannelEventBus {
	return &ChannelEventBus{ch: make(chan ports.Event, bufSize)}
}

// Emit sends an event (non-blocking — drops on full buffer).
func (b *ChannelEventBus) Emit(e ports.Event) {
	select {
	case b.ch <- e:
	default:
	}
}

// Channel returns the underlying channel for the driving adapter to read.
func (b *ChannelEventBus) Channel() <-chan ports.Event {
	return b.ch
}
