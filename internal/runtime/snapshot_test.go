package runtime

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudspacelab/forge/internal/adapters"
	"github.com/cloudspacelab/forge/internal/tools"
)

// SNAPSHOT / GOLDEN TESTS
// INVARIANT: deterministic outputs must not change without intent.
// These capture the shape of SystemPrompt so a regression is loud.
// (Using inline expectations rather than golden files for portability.)

func TestSystemPrompt_HasStableStructure(t *testing.T) {
	rt, err := New(Config{FolderPath: "../..", AutoApprove: true})
	if err != nil {
		t.Fatal(err)
	}
	rt.SetToolExecutor(adapters.NewToolExecutorAdapter(tools.NewRegistry()))

	prompt := rt.SystemPrompt()

	// Structural invariants — these sections must always exist
	requiredSections := []string{
		"You are Forge",
		"PROJECT CONTEXT",
		"Root:",
		"AVAILABLE TOOLS",
	}
	for _, section := range requiredSections {
		assert.True(t, strings.Contains(prompt, section),
			"SystemPrompt missing required section: %q", section)
	}

	// Must mention all registered tools
	tools := []string{"read", "write", "edit", "list", "bash"}
	for _, tool := range tools {
		assert.Contains(t, prompt, tool, "SystemPrompt must list tool: %s", tool)
	}
}

func TestSystemPrompt_NoMutationOnRepeatedCall(t *testing.T) {
	rt, err := New(Config{FolderPath: "../.."})
	if err != nil {
		t.Fatal(err)
	}
	p1 := rt.SystemPrompt()
	p2 := rt.SystemPrompt()
	assert.Equal(t, p1, p2, "SystemPrompt must be deterministic")
}
