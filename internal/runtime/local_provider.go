package runtime

import (
	"regexp"
	"strings"

	"github.com/cloudspacelab/forge/internal/ports"
)

// LocalProvider implements ports.Provider using a rule-based intent
// parser — no LLM required. It maps natural language to tool calls
// using pattern matching. This makes Forge fully functional offline.
type LocalProvider struct{}

func (p *LocalProvider) Name() string { return "local" }

func (p *LocalProvider) Generate(req ports.ProviderRequest) ports.ProviderResponse {
	last := ""
	if len(req.Messages) > 0 {
		last = req.Messages[len(req.Messages)-1].Content
	}
	return parseIntent(last)
}

// parseIntent maps a natural-language string to tool calls + response text.
func parseIntent(text string) ports.ProviderResponse {
	lower := strings.ToLower(strings.TrimSpace(text))

	// --- Greetings (only match standalone, not inside other text) ---
	if lower == "hello" || lower == "hi" || lower == "hey" ||
		strings.HasPrefix(lower, "hello ") || strings.HasPrefix(lower, "hi ") {
		return ports.ProviderResponse{
			Text: "Hello! I'm Forge. I can read files, run commands, build, test, and help you code.\n" +
				"Try: \"read main.go\", \"run tests\", \"list files\", or type / for slash commands.",
			Stop: true,
		}
	}

	// --- Read a file ---
	if file := extractFile(lower, "read", "show", "cat", "open", "display"); file != "" {
		return ports.ProviderResponse{
			Text:      "Reading " + file + "...",
			ToolCalls: []ports.ToolCall{{Name: "read", Args: map[string]string{"path": file}}},
		}
	}

	// --- List files ---
	if match(lower, "list files", "ls", "show files", "dir ", "list dir") {
		dir := "."
		if d := extractArg(lower, "in ", "under "); d != "" {
			dir = d
		}
		return ports.ProviderResponse{
			Text:      "Listing files in " + dir + "...",
			ToolCalls: []ports.ToolCall{{Name: "list", Args: map[string]string{"dir": dir}}},
		}
	}

	// --- Run tests ---
	if match(lower, "test", "run test", "go test", "run the test") {
		return ports.ProviderResponse{
			Text:      "Running tests...",
			ToolCalls: []ports.ToolCall{{Name: "bash", Args: map[string]string{"command": "go test ./... 2>&1"}}},
		}
	}

	// --- Build ---
	if match(lower, "build", "compile", "go build", "make build") {
		return ports.ProviderResponse{
			Text:      "Building...",
			ToolCalls: []ports.ToolCall{{Name: "bash", Args: map[string]string{"command": "go build ./... 2>&1"}}},
		}
	}

	// --- Git status ---
	if match(lower, "status", "git status", "what changed", "working tree") {
		return ports.ProviderResponse{
			Text:      "Checking git status...",
			ToolCalls: []ports.ToolCall{{Name: "git_status", Args: map[string]string{}}},
		}
	}

	// --- Run arbitrary command ---
	if cmd := extractArg(lower, "run ", "execute "); cmd != "" && !match(lower, "run test", "run the test") {
		return ports.ProviderResponse{
			Text:      "Running: " + cmd,
			ToolCalls: []ports.ToolCall{{Name: "bash", Args: map[string]string{"command": cmd}}},
		}
	}

	// --- Show help ---
	if match(lower, "help", "what can you do", "commands") {
		return ports.ProviderResponse{
			Text: "I can: read files (\"read main.go\"), list files (\"ls\"), run tests (\"test\"), " +
				"build (\"build\"), check git (\"status\"), run commands (\"run echo hi\").\n" +
				"Type / for slash commands.",
			Stop: true,
		}
	}

	// --- Fallback: try to interpret as a bash command ---
	if strings.Contains(text, " ") && !strings.Contains(lower, "what") && !strings.Contains(lower, "how") {
		return ports.ProviderResponse{
			Text:      "Trying as a command: " + text,
			ToolCalls: []ports.ToolCall{{Name: "bash", Args: map[string]string{"command": text}}},
		}
	}

	return ports.ProviderResponse{
		Text: "I'm not sure how to help with that. Try:\n" +
			"  • \"read <file>\" — read a file\n" +
			"  • \"ls\" — list files\n" +
			"  • \"test\" — run tests\n" +
			"  • \"build\" — build the project\n" +
			"  • \"status\" — git status\n" +
			"  • / for slash commands\n" +
			"\nOr configure a local model: forge --provider mlx --model mlx-community/Llama-3.2-3B-Instruct-4bit",
		Stop: true,
	}
}

// match returns true if the text contains any of the patterns.
func match(text string, patterns ...string) bool {
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// extractFile finds a filename mentioned after a verb like "read".
func extractFile(text string, verbs ...string) string {
	for _, verb := range verbs {
		re := regexp.MustCompile(regexp.QuoteMeta(verb) + `\s+(\S+)`)
		if m := re.FindStringSubmatch(text); m != nil {
			file := m[1]
			// Skip common non-file words
			skip := map[string]bool{"the": true, "a": true, "all": true, "files": true, "me": true, "this": true}
			if !skip[file] {
				return file
			}
		}
	}
	return ""
}

// extractArg finds text after a keyword (for "run <cmd>", "in <dir>").
func extractArg(text string, keywords ...string) string {
	original := text
	for _, kw := range keywords {
		idx := strings.Index(text, kw)
		if idx >= 0 {
			arg := strings.TrimSpace(text[idx+len(kw):])
			// Take up to end of line or next sentence
			if end := strings.IndexAny(arg, ".\n"); end >= 0 {
				arg = arg[:end]
			}
			arg = strings.TrimSpace(arg)
			if arg != "" {
				return arg
			}
		}
		text = original
	}
	return ""
}
