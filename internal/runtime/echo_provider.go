package runtime

import (
	"strings"

	"github.com/cloudspacelab/forge/internal/ports"
)

// EchoProvider is a minimal OFFLINE adapter implementing ports.Provider.
// It recognises simple intents and demonstrates the loop without a real
// LLM. Real providers (ollama, openai) are separate adapters.
type EchoProvider struct{}

func (p *EchoProvider) Name() string { return "echo" }

func (p *EchoProvider) Generate(req ports.ProviderRequest) ports.ProviderResponse {
	last := ""
	if len(req.Messages) > 0 {
		last = req.Messages[len(req.Messages)-1].Content
	}
	lower := strings.ToLower(last)

	switch {
	case strings.Contains(lower, "hello") || strings.Contains(lower, "hi"):
		return ports.ProviderResponse{
			Text: "Hello! I'm Forge. I can read files, run commands, and help you build.\nType /help for slash commands.",
			Stop: true,
		}
	case strings.Contains(lower, "who are you") || strings.Contains(lower, "what are you"):
		return ports.ProviderResponse{
			Text: "I'm Forge, a terminal-native coding agent built on hexagonal architecture. I live in your project folder and remember what we do together.",
			Stop: true,
		}
	case strings.Contains(lower, "list") && (strings.Contains(lower, "file") || strings.Contains(lower, "dir")):
		return ports.ProviderResponse{
			Text:      "Let me list the files for you.",
			ToolCalls: []ports.ToolCall{{Name: "list", Args: map[string]string{"dir": "."}}},
		}
	case strings.Contains(lower, "status") || strings.Contains(lower, "git status"):
		return ports.ProviderResponse{
			Text:      "Checking git status...",
			ToolCalls: []ports.ToolCall{{Name: "git_status", Args: map[string]string{}}},
		}
	case strings.Contains(lower, "test") && !strings.Contains(lower, "latest"):
		return ports.ProviderResponse{
			Text:      "Running tests...",
			ToolCalls: []ports.ToolCall{{Name: "bash", Args: map[string]string{"command": "go test ./... 2>&1"}}},
		}
	case strings.Contains(lower, "build"):
		return ports.ProviderResponse{
			Text:      "Building...",
			ToolCalls: []ports.ToolCall{{Name: "bash", Args: map[string]string{"command": "go build ./... 2>&1"}}},
		}
	default:
		return ports.ProviderResponse{
			Text: "I heard: \"" + truncate(last, 80) + "\"\n\n(Echo mode — configure a real provider for full intelligence. Slash commands like /help, /ls, /skills still work.)",
			Stop: true,
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
