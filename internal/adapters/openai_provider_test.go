package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToolCallsFromText_SingleCall(t *testing.T) {
	text := `Let me read that file. <tool name="read" path="main.go" />`
	calls := parseToolCallsFromText(text)
	assert.Len(t, calls, 1)
	assert.Equal(t, "read", calls[0].Name)
	assert.Equal(t, "main.go", calls[0].Args["path"])
}

func TestParseToolCallsFromText_MultipleArgs(t *testing.T) {
	text := `<tool name="bash" command="go test ./..." />`
	calls := parseToolCallsFromText(text)
	assert.Len(t, calls, 1)
	assert.Equal(t, "bash", calls[0].Name)
	assert.Equal(t, "go test ./...", calls[0].Args["command"])
}

func TestParseToolCallsFromText_MultipleCalls(t *testing.T) {
	text := `<tool name="read" path="a.go" /> and then <tool name="read" path="b.go" />`
	calls := parseToolCallsFromText(text)
	assert.Len(t, calls, 2)
}

func TestParseToolCallsFromText_NoCalls(t *testing.T) {
	text := "Just a normal response with no tool calls."
	calls := parseToolCallsFromText(text)
	assert.Len(t, calls, 0)
}

func TestRemoveToolMarkup(t *testing.T) {
	text := "Here is what I found. <tool name=\"read\" path=\"x.go\" /> Done."
	clean := removeToolMarkup(text)
	assert.NotContains(t, clean, "<tool")
	assert.Contains(t, clean, "Here is what I found")
}

func TestPortFromURL(t *testing.T) {
	assert.Equal(t, "8080", portFromURL("http://localhost:8080/v1"))
	assert.Equal(t, "1234", portFromURL("http://localhost:1234/v1"))
	assert.Equal(t, "8080", portFromURL("http://localhost"))
}
