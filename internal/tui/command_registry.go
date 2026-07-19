package tui

// Command describes a single slash command for autocomplete + help.
type Command struct {
	Name        string // "/read"
	Shortcut    string // "/r" (optional)
	Description string
	Usage       string // "/read <file>"
	Args        string // "<file>" or ""
}

// commands is the master list of all slash commands, used by both
// autocomplete and the help screen.
var commands = []Command{
	{Name: "/help", Shortcut: "/?", Description: "Show all commands", Usage: "/help", Args: ""},
	{Name: "/project", Shortcut: "/info", Description: "Show discovered project info", Usage: "/project", Args: ""},
	{Name: "/session", Shortcut: "", Description: "Show current session details", Usage: "/session", Args: ""},
	{Name: "/ls", Shortcut: "/list", Description: "List files in a directory", Usage: "/ls [dir]", Args: "[dir]"},
	{Name: "/read", Shortcut: "/cat", Description: "Read a file's contents", Usage: "/read <file>", Args: "<file>"},
	{Name: "/bash", Shortcut: "/run", Description: "Run a shell command", Usage: "/bash <cmd>", Args: "<cmd>"},
	{Name: "/test", Shortcut: "", Description: "Run go test (default ./...)", Usage: "/test [pkg]", Args: "[pkg]"},
	{Name: "/build", Shortcut: "", Description: "Run go build ./...", Usage: "/build", Args: ""},
	{Name: "/status", Shortcut: "", Description: "Show git status", Usage: "/status", Args: ""},
	{Name: "/tools", Shortcut: "", Description: "List available tools", Usage: "/tools", Args: ""},
	{Name: "/skills", Shortcut: "", Description: "List installed skills", Usage: "/skills", Args: ""},
	{Name: "/nerd", Shortcut: "", Description: "Show the nerd skill (ASCII flowcharts)", Usage: "/nerd", Args: ""},
	{Name: "/index", Shortcut: "", Description: "Build the Go symbol index", Usage: "/index", Args: ""},
	{Name: "/packages", Shortcut: "", Description: "List all Go packages", Usage: "/packages", Args: ""},
	{Name: "/symbols", Shortcut: "", Description: "Find symbols by name", Usage: "/symbols [name]", Args: "[name]"},
	{Name: "/version", Shortcut: "/v", Description: "Show version", Usage: "/version", Args: ""},
	{Name: "/clear", Shortcut: "", Description: "Clear the transcript", Usage: "/clear", Args: ""},
	{Name: "/quit", Shortcut: "/q", Description: "Exit Forge", Usage: "/quit", Args: ""},
}

// filterCommands returns commands whose name or shortcut starts with prefix.
func filterCommands(prefix string) []Command {
	if prefix == "" {
		return commands
	}
	var result []Command
	for _, c := range commands {
		if startsWith(c.Name, prefix) || (c.Shortcut != "" && startsWith(c.Shortcut, prefix)) {
			result = append(result, c)
		}
	}
	return result
}

func startsWith(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	return s[:len(prefix)] == prefix
}
