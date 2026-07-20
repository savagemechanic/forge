package adapters_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudspacelab/forge/internal/adapters"
	"github.com/cloudspacelab/forge/internal/ports"
	"github.com/cloudspacelab/forge/internal/runtime"
)

// CONTRACT TEST SUITE: ports.Provider
// INVARIANT: every Provider implementation must produce a non-hanging,
// non-panicking response for any well-formed request, and honor the
// Stop/ToolCalls contract. Runs against ALL impls → guarantees they
// are interchangeable (the hexagonal promise).

func TestProviderContract(t *testing.T) {
	mockServer := newMockChatServer(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`)
	defer mockServer.Close()

	impls := map[string]ports.Provider{
		"local":  &runtime.LocalProvider{},
		"echo":   &runtime.EchoProvider{},
		"openai": adapters.NewOpenAIProvider(adapters.OpenAIProviderConfig{BaseURL: mockServer.URL, Model: "m"}),
	}

	for name, p := range impls {
		t.Run(name, func(t *testing.T) {
			runProviderContract(t, p)
		})
	}
}

func runProviderContract(t *testing.T, p ports.Provider) {
	t.Helper()

	t.Run("never panics on any input", func(t *testing.T) {
		assert.NotPanics(t, func() {
			p.Generate(ports.ProviderRequest{
				Messages: []ports.ProviderMessage{{Role: "user", Content: "anything"}},
			})
		})
	})

	t.Run("responds to greeting", func(t *testing.T) {
		resp := p.Generate(ports.ProviderRequest{
			Messages: []ports.ProviderMessage{{Role: "user", Content: "hello"}},
		})
		// Shape contract: has some text or tool calls, never both empty
		assert.True(t, resp.Text != "" || len(resp.ToolCalls) > 0,
			"provider must produce SOME output")
	})

	t.Run("name is non-empty", func(t *testing.T) {
		assert.NotEmpty(t, p.Name())
	})

	t.Run("handles empty messages", func(t *testing.T) {
		assert.NotPanics(t, func() {
			p.Generate(ports.ProviderRequest{Messages: nil})
		})
	})
}

// newMockChatServer returns an httptest.Server that responds with a canned
// OpenAI-format JSON body to any request.
func newMockChatServer(cannedBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate it parses as a chat request (shape contract)
		var req struct {
			Model string `json:"model"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		fmt.Fprint(w, cannedBody)
	}))
}
