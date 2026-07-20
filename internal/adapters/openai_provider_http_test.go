package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudspacelab/forge/internal/ports"
)

// INVARIANT: the OpenAI provider must produce the same parsed result
// regardless of which server supplies the bytes. The mock server lets us
// verify the TRANSPORT contract (request shape) and INTERPRETATION contract
// (response parsing) without a live model.

// mockMLXServer returns a canned OpenAI-format response.
func mockMLXServer(t *testing.T, content string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's hitting the chat completions endpoint
		assert.Contains(t, r.URL.Path, "chat/completions")

		// Verify the request body shape
		var req openaiChatRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "test-model", req.Model)
		assert.NotEmpty(t, req.Messages)

		resp := openaiChatResponse{}
		_ = json.Unmarshal([]byte(fmt.Sprintf(`{"choices":[{"message":{"role":"assistant","content":%q}}]}`, content)), &resp)

		json.NewEncoder(w).Encode(resp)
	}))
}

func TestOpenAIProvider_PlainTextResponse(t *testing.T) {
	server := mockMLXServer(t, "Hello from the model!")
	defer server.Close()

	p := NewOpenAIProvider(OpenAIProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	})

	resp := p.Generate(ports.ProviderRequest{
		SystemPrompt: "You are a test.",
		Messages:     []ports.ProviderMessage{{Role: "user", Content: "hi"}},
	})

	assert.Equal(t, "Hello from the model!", resp.Text)
	assert.True(t, resp.Stop, "no tool calls → stop=true")
	assert.Empty(t, resp.ToolCalls)
}

func TestOpenAIProvider_ParsesToolCalls(t *testing.T) {
	content := `Let me read that. <tool name="read" path="main.go" />`
	server := mockMLXServer(t, content)
	defer server.Close()

	p := NewOpenAIProvider(OpenAIProviderConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	})

	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{{Role: "user", Content: "read main.go"}},
	})

	assert.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "read", resp.ToolCalls[0].Name)
	assert.Equal(t, "main.go", resp.ToolCalls[0].Args["path"])
	assert.False(t, resp.Stop, "tool calls present → stop=false")
	// Text should have the markup removed
	assert.NotContains(t, resp.Text, "<tool")
}

func TestOpenAIProvider_ServerDown_GivesClearError(t *testing.T) {
	p := NewOpenAIProvider(OpenAIProviderConfig{
		BaseURL: "http://127.0.0.1:1", // nobody listening
		Model:   "test-model",
	})

	resp := p.Generate(ports.ProviderRequest{
		Messages: []ports.ProviderMessage{{Role: "user", Content: "hi"}},
	})

	assert.Contains(t, resp.Text, "Cannot reach model server")
	assert.Contains(t, resp.Text, "mlx_lm.server")
	assert.True(t, resp.Stop)
}

func TestOpenAIProvider_SendsSystemPrompt(t *testing.T) {
	var capturedReq openaiChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedReq)
		json.NewEncoder(w).Encode(openaiChatResponse{})
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIProviderConfig{BaseURL: server.URL, Model: "m"})
	p.Generate(ports.ProviderRequest{
		SystemPrompt: "You are Forge.",
		Tools:        []ports.ToolSpec{{Name: "read", Description: "read a file"}},
		Messages:     []ports.ProviderMessage{{Role: "user", Content: "x"}},
	})

	require.NotEmpty(t, capturedReq.Messages)
	assert.Equal(t, "system", capturedReq.Messages[0].Role)
	assert.Contains(t, capturedReq.Messages[0].Content, "You are Forge")
	assert.Contains(t, capturedReq.Messages[0].Content, "read")
}

func TestOpenAIProvider_Name(t *testing.T) {
	p := NewOpenAIProvider(OpenAIProviderConfig{Model: "llama-3.2"})
	assert.Contains(t, p.Name(), "llama-3.2")
}
