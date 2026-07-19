package tui

import (
	"fmt"
	"strings"

	"github.com/cloudspacelab/forge/internal/goengine"
)

// commandResult is the output of a slash command.
type commandResult struct {
	output   string
	isError  bool
	quit     bool
	showHelp bool
}

// handleSlashCommand processes a command starting with "/".
// Returns the result to display and whether to quit.
func (m *Model) handleSlashCommand(input string) commandResult {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return commandResult{output: "", isError: false}
	}
	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "/help", "/h", "/?":
		return commandResult{output: helpText(), showHelp: true}

	case "/quit", "/q", "/exit":
		return commandResult{quit: true}

	case "/clear":
		m.blocks = nil
		return commandResult{output: ""}

	case "/skills":
		if m.rt != nil && m.skillLoader != nil {
			return commandResult{output: m.skillLoader.List()}
		}
		return commandResult{output: "No skills loaded."}

	case "/project", "/info":
		if m.rt != nil {
			p := m.rt.Project()
			var b strings.Builder
			fmt.Fprintf(&b, "  Root:    %s\n", p.RootPath)
			if p.GoModule != "" {
				fmt.Fprintf(&b, "  Module:  %s\n", p.GoModule)
			}
			fmt.Fprintf(&b, "  Git:     %s\n", yesno(p.GitRoot != ""))
			if p.HasGoWork {
				fmt.Fprintf(&b, "  Workspace: yes\n")
			}
			if len(p.Subpackages) > 0 {
				fmt.Fprintf(&b, "  Packages (%d):\n", len(p.Subpackages))
				for _, pkg := range p.Subpackages {
					fmt.Fprintf(&b, "    %s\n", pkg)
				}
			}
			return commandResult{output: b.String()}
		}
		return commandResult{output: "No project loaded."}

	case "/ls", "/list":
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		res := m.executeTool("list", map[string]string{"dir": dir})
		return commandResult{output: res.Output, isError: res.Err != nil}

	case "/cat", "/read":
		if len(args) == 0 {
			return commandResult{output: "Usage: /read <file>", isError: true}
		}
		res := m.executeTool("read", map[string]string{"path": args[0]})
		return commandResult{output: res.Output, isError: res.Err != nil}

	case "/bash", "/run":
		if len(args) == 0 {
			return commandResult{output: "Usage: /bash <command>", isError: true}
		}
		res := m.executeTool("bash", map[string]string{"command": strings.Join(args, " ")})
		return commandResult{output: res.Output, isError: res.Err != nil}

	case "/status":
		res := m.executeTool("git_status", map[string]string{})
		return commandResult{output: res.Output, isError: res.Err != nil}

	case "/tools":
		if m.rt != nil {
			var b strings.Builder
			b.WriteString("Available tools:\n")
			for _, t := range m.rt.ToolSpecs() {
				fmt.Fprintf(&b, "  %-12s %s\n", t.Name, t.Description)
			}
			return commandResult{output: b.String()}
		}
		return commandResult{output: "No tools loaded."}

	case "/test":
		pkg := "./..."
		if len(args) > 0 {
			pkg = args[0]
		}
		res := m.executeTool("bash", map[string]string{"command": "go test " + pkg + " 2>&1"})
		return commandResult{output: res.Output, isError: res.Err != nil}

	case "/build":
		res := m.executeTool("bash", map[string]string{"command": "go build ./... 2>&1"})
		return commandResult{output: res.Output, isError: res.Err != nil}

	case "/session":
		if m.rt != nil {
			s := m.rt.Session()
			var b strings.Builder
			fmt.Fprintf(&b, "  ID:        %s\n", s.ID)
			fmt.Fprintf(&b, "  Folder:    %s\n", s.FolderID)
			fmt.Fprintf(&b, "  State:     %s\n", s.State)
			fmt.Fprintf(&b, "  Messages:  %d\n", len(s.Messages))
			return commandResult{output: b.String()}
		}
		return commandResult{output: "No session."}

	case "/nerd":
		if m.skillLoader != nil {
			if s, ok := m.skillLoader.Get("nerd"); ok {
				return commandResult{output: fmt.Sprintf("Nerd skill loaded:\n\n%s", s.Body)}
			}
		}
		return commandResult{output: "Nerd skill not installed. Run: make install-skills", isError: true}

	case "/version", "/v":
		return commandResult{output: "Forge 0.2.0 (self-hosting build)"}

	case "/index":
		idx, err := goengine.Load(m.rt.Project().RootPath)
		if err != nil {
			return commandResult{output: fmt.Sprintf("index error: %v", err), isError: true}
		}
		m.goIndex = idx
		return commandResult{output: "Go index built: " + idx.Summary()}

	case "/packages":
		if m.goIndex == nil {
			return commandResult{output: "No index. Run /index first.", isError: true}
		}
		return commandResult{output: m.goIndex.PrintPackages()}

	case "/symbols":
		if m.goIndex == nil {
			return commandResult{output: "No index. Run /index first.", isError: true}
		}
		if len(args) == 0 {
			// List all exported symbols
			var b strings.Builder
			for _, s := range m.goIndex.ExportedSymbols() {
				fmt.Fprintf(&b, "  %-6s %-40s %s\n", s.Kind, s.Name, s.Pos)
			}
			return commandResult{output: b.String()}
		}
		syms := m.goIndex.FindSymbol(args[0])
		if len(syms) == 0 {
			return commandResult{output: fmt.Sprintf("No symbols matching '%s'", args[0])}
		}
		var b strings.Builder
		for _, s := range syms {
			fmt.Fprintf(&b, "  %-6s %-40s %s\n", s.Kind, s.Name, s.Pos)
		}
		return commandResult{output: b.String()}

	default:
		return commandResult{
			output:  fmt.Sprintf("Unknown command: %s\nType /help for available commands.", cmd),
			isError: true,
		}
	}
}

// helpText returns the full help screen.
func helpText() string {
	return `FORGE — Terminal Coding Agent

SLASH COMMANDS:
  /help, /?        Show this help
  /project         Show discovered project info
  /session         Show current session details
  /ls [dir]        List files in a directory
  /read <file>     Read a file's contents
  /bash <cmd>      Run a shell command
  /test [pkg]      Run go test (default: ./...)
  /build           Run go build ./...
  /status          Show git status
  /tools           List available tools
  /skills          List installed skills
  /nerd            Show the nerd skill (ASCII flowcharts)
  /index           Build the Go symbol index
  /packages        List all Go packages
  /symbols [name]  Find symbols by name
  /version         Show version
  /clear           Clear the transcript
  /quit, /q        Exit Forge

NATURAL LANGUAGE:
  Just type a message and press Enter. Forge processes it through
  the agent loop (provider + tools + skills).

  Examples:
    "list the files in internal/"
    "what does this project do?"
    "run the tests"

KEYBINDINGS:
  Enter          Submit input
  Ctrl+C         Quit
  Ctrl+L         Clear screen
  Up/Down        Scroll transcript
  PgUp/PgDn      Page up/down
  Tab            Autocomplete (commands)
`
}

func yesno(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
