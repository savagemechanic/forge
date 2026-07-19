package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/cloudspacelab/forge/internal/ports"
)

// OpenAIProviderConfig configures an OpenAI-compatible provider.
// Works with MLX servers (mlx-lm), LM Studio, Ollama (OpenAI mode),
// vLLM, or any server exposing /v1/chat/completions.
type OpenAIProviderConfig struct {
	BaseURL string // e.g. "http://localhost:8080/v1"
	Model   string // e.g. "mlx-community/Llama-3.2-3B-Instruct-4bit"
	APIKey  string // optional for local servers
	Timeout time.Duration
}

// OpenAIProvider implements ports.Provider via an OpenAI-compatible API.
// This is the adapter for MLX and other local model servers.
type OpenAIProvider struct {
	cfg    OpenAIProviderConfig
	client *http.Client
}

// NewOpenAIProvider creates an MLX/OpenAI-compatible provider.
func NewOpenAIProvider(cfg OpenAIProviderConfig) *OpenAIProvider {
	if cfg.Timeout == 0 {
		cfg.Timeout = 120 * time.Second
	}
	return &OpenAIProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

func (p *OpenAIProvider) Name() string { return "openai:" + p.cfg.Model }

func (p *OpenAIProvider) Generate(req ports.ProviderRequest) ports.ProviderResponse {
	// Build the OpenAI chat request
	messages := []openaiMessage{
		{Role: "system", Content: req.SystemPrompt + buildToolInstructions(req.Tools)},
	}
	for _, m := range req.Messages {
		role := m.Role
		if role == "tool" {
			role = "user" // feed tool results back as user context
		}
		content := m.Content
		if m.ToolName != "" {
			content = fmt.Sprintf("[tool:%s result]\n%s", m.ToolName, m.Content)
		}
		messages = append(messages, openaiMessage{Role: role, Content: content})
	}

	body := openaiChatRequest{
		Model:    p.cfg.Model,
		Messages: messages,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return ports.ProviderResponse{Text: fmt.Sprintf("error building request: %v", err), Stop: true}
	}

	url := strings.TrimSuffix(p.cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return ports.ProviderResponse{Text: fmt.Sprintf("error creating request: %v", err), Stop: true}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return ports.ProviderResponse{
			Text: fmt.Sprintf("⚠ Cannot reach model server at %s\n\n"+
				"Make sure your MLX server is running:\n"+
				"  mlx_lm.server --model %s --port %s\n\n"+
				"Or use LM Studio / Ollama with OpenAI compatibility.\n"+
				"Error: %v", p.cfg.BaseURL, p.cfg.Model, portFromURL(p.cfg.BaseURL), err),
			Stop: true,
		}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return ports.ProviderResponse{
			Text: fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(raw)),
			Stop: true,
		}
	}

	var chatResp openaiChatResponse
	if err := json.Unmarshal(raw, &chatResp); err != nil {
		return ports.ProviderResponse{Text: fmt.Sprintf("error parsing response: %v", err), Stop: true}
	}

	if len(chatResp.Choices) == 0 {
		return ports.ProviderResponse{Text: "(empty response)", Stop: true}
	}

	content := chatResp.Choices[0].Message.Content
	toolCalls := parseToolCallsFromText(content)

	// If tool calls found, clean the text to remove the markup
	cleanText := content
	if len(toolCalls) > 0 {
		cleanText = removeToolMarkup(content)
	}

	stop := len(toolCalls) == 0
	return ports.ProviderResponse{
		Text:      cleanText,
		ToolCalls: toolCalls,
		Stop:      stop,
	}
}

// buildToolInstructions appends tool-calling instructions to the system prompt.
// Local models often don't support native function calling, so we instruct
// them to emit tool calls in a parseable format.
func buildToolInstructions(tools []ports.ToolSpec) string {
	if len(tools) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\nYou can use tools by writing TOOL CALLS in this format:\n")
	b.WriteString("<tool name=\"toolname\" arg1=\"value1\" arg2=\"value2\" />\n\n")
	b.WriteString("Available tools:\n")
	for _, t := range tools {
		b.WriteString(fmt.Sprintf("  - %s: %s\n", t.Name, t.Description))
	}
	b.WriteString("\nTo read a file: <tool name=\"read\" path=\"main.go\" />\n")
	b.WriteString("To run a command: <tool name=\"bash\" command=\"go test ./...\" />\n")
	b.WriteString("You can call multiple tools. After tools complete, I'll give you the results.\n")
	b.WriteString("When you're done and don't need more tools, just respond normally.\n")
	return b.String()
}

// toolCallRegex matches <tool .../> tags. Uses non-greedy to allow /
// inside attribute values like command="go test ./...".
var toolCallRegex = regexp.MustCompile(`<tool\s+(.*?)/>`)

// parseToolCallsFromText extracts tool calls from model output.
func parseToolCallsFromText(text string) []ports.ToolCall {
	matches := toolCallRegex.FindAllStringSubmatch(text, -1)
	var calls []ports.ToolCall

	attrRegex := regexp.MustCompile(`(\w+)="([^"]*)"`)
	for _, m := range matches {
		attrs := attrRegex.FindAllStringSubmatch(m[1], -1)
		args := make(map[string]string)
		var name string
		for _, a := range attrs {
			if a[1] == "name" {
				name = a[2]
			} else {
				args[a[1]] = a[2]
			}
		}
		if name != "" {
			calls = append(calls, ports.ToolCall{Name: name, Args: args})
		}
	}
	return calls
}

// removeToolMarkup strips <tool .../> tags from the text.
func removeToolMarkup(text string) string {
	cleaned := toolCallRegex.ReplaceAllString(text, "")
	return strings.TrimSpace(cleaned)
}

func portFromURL(url string) string {
	// Find the port after the host (skip http:// or https://)
	stripped := url
	if idx := strings.Index(stripped, "://"); idx >= 0 {
		stripped = stripped[idx+3:]
	}
	// stripped is now "localhost:8080/v1" or "localhost"
	colon := strings.Index(stripped, ":")
	if colon < 0 {
		return "8080" // default
	}
	rest := stripped[colon+1:]
	if slash := strings.Index(rest, "/"); slash >= 0 {
		rest = rest[:slash]
	}
	if rest == "" {
		return "8080"
	}
	return rest
}

// ---- OpenAI API types ----

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
