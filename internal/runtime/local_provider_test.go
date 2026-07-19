package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudspacelab/forge/internal/ports"
)

func TestLocalProvider_ReadFile(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "read main.go"},
		},
	})
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "read", resp.ToolCalls[0].Name)
	assert.Equal(t, "main.go", resp.ToolCalls[0].Args["path"])
}

func TestLocalProvider_ListFiles(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "ls"},
		},
	})
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "list", resp.ToolCalls[0].Name)
}

func TestLocalProvider_RunTests(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "run the tests"},
		},
	})
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "bash", resp.ToolCalls[0].Name)
	assert.Contains(t, resp.ToolCalls[0].Args["command"], "go test")
}

func TestLocalProvider_Build(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "build the project"},
		},
	})
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "bash", resp.ToolCalls[0].Name)
	assert.Contains(t, resp.ToolCalls[0].Args["command"], "go build")
}

func TestLocalProvider_GitStatus(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "what's my git status?"},
		},
	})
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "git_status", resp.ToolCalls[0].Name)
}

func TestLocalProvider_RunCommand(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "run echo hello"},
		},
	})
	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "bash", resp.ToolCalls[0].Name)
	assert.Contains(t, resp.ToolCalls[0].Args["command"], "echo hello")
}

func TestLocalProvider_Help(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "help"},
		},
	})
	assert.Contains(t, resp.Text, "read files")
	assert.True(t, resp.Stop)
}

func TestLocalProvider_Greeting(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "hello"},
		},
	})
	assert.Contains(t, resp.Text, "Hello")
	assert.True(t, resp.Stop)
}

func TestLocalProvider_Fallback(t *testing.T) {
	p := &LocalProvider{}
	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{
			{Role: "user", Content: "asdfjkl"},
		},
	})
	// Should either try as a command or show help
	assert.True(t, resp.Stop || len(resp.ToolCalls) > 0)
}
