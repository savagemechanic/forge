package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudspacelab/forge/internal/adapters"
	"github.com/cloudspacelab/forge/internal/ports"
	"github.com/cloudspacelab/forge/internal/tools"
)

func TestNew_CreatesRuntimeWithSession(t *testing.T) {
	rt, err := New(Config{FolderPath: "../..", AutoApprove: true})
	require.NoError(t, err)

	assert.NotNil(t, rt.Session())
	assert.NotNil(t, rt.Project())
	assert.Equal(t, "active", string(rt.Session().State))
}

func TestRuntime_SystemPromptIncludesContext(t *testing.T) {
	rt, err := New(Config{FolderPath: "../.."})
	require.NoError(t, err)

	prompt := rt.SystemPrompt()
	assert.Contains(t, prompt, "You are Forge")
	assert.Contains(t, prompt, "PROJECT CONTEXT")
}

func TestRuntime_InjectionOfPorts(t *testing.T) {
	rt, err := New(Config{FolderPath: "../.."})
	require.NoError(t, err)

	// Inject tool executor
	te := adapters.NewToolExecutorAdapter(tools.NewRegistry())
	rt.SetToolExecutor(te)

	// Inject provider
	rt.SetProvider(&EchoProvider{})

	// Inject event bus
	bus := adapters.NewChannelEventBus(8)
	rt.SetEventBus(bus)

	specs := rt.ToolSpecs()
	assert.NotEmpty(t, specs)
	assert.Contains(t, []string{"read", "write", "edit", "list", "bash"}, specs[0].Name)
}

func TestRuntime_ExecuteToolDirect(t *testing.T) {
	rt, err := New(Config{FolderPath: "../..", AutoApprove: true})
	require.NoError(t, err)
	rt.SetToolExecutor(adapters.NewToolExecutorAdapter(tools.NewRegistry()))

	result := rt.ExecuteToolDirect("list", map[string]string{"dir": "."})
	assert.NoError(t, result.Err)
	assert.Contains(t, result.Output, "cmd")
}

func TestRuntime_SubmitWithoutProvider(t *testing.T) {
	rt, err := New(Config{FolderPath: "../.."})
	require.NoError(t, err)

	bus := adapters.NewChannelEventBus(8)
	rt.SetEventBus(bus)

	rt.Submit("test message")

	// Should emit turn_start, system, turn_end
	var events []ports.Event
	for {
		select {
		case e := <-bus.Channel():
			events = append(events, e)
		default:
			goto done
		}
	}
done:
	assert.GreaterOrEqual(t, len(events), 2)
	assert.Equal(t, ports.EventTurnStart, events[0].Type)
}

func TestEchoProvider_GeneratesResponse(t *testing.T) {
	p := &EchoProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "hello"},
		},
	})
	assert.Contains(t, resp.Text, "Hello")
	assert.True(t, resp.Stop)
}

func TestEchoProvider_DetectsListIntent(t *testing.T) {
	p := &EchoProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "list the files"},
		},
	})
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "list", resp.ToolCalls[0].Name)
}
