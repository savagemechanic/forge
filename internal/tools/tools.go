// Package tools provides the tool interface and built-in implementations
// that the agent runtime can call. Each tool executes in a controlled
// context and is subject to policy approval.
package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result is the output of a tool execution.
type Result struct {
	Tool    string
	Output  string
	Err     error
	Changed []string // files created/modified
}

// Context carries execution-scoped state (cwd, approvals, session).
type Context struct {
	WorkDir    string
	Approved   bool // whether changes are pre-approved
	DryRun     bool // if true, don't actually modify files
}

// Tool is the interface every callable tool implements.
type Tool interface {
	Name() string
	Description() string
	Execute(ctx *Context, args map[string]string) Result
}

// Registry holds all registered tools, keyed by name.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a registry pre-loaded with all built-in tools.
func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]Tool)}
	r.Register(&FileReadTool{})
	r.Register(&FileWriteTool{})
	r.Register(&FileEditTool{})
	r.Register(&ListFilesTool{})
	r.Register(&BashTool{})
	r.Register(&GitStatusTool{})
	return r
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns all registered tools.
func (r *Registry) All() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// Execute runs a named tool with the given args.
func (r *Registry) Execute(ctx *Context, name string, args map[string]string) Result {
	tool, ok := r.tools[name]
	if !ok {
		return Result{Tool: name, Err: fmt.Errorf("unknown tool: %s", name)}
	}
	return tool.Execute(ctx, args)
}

// ============================================================================
// FILE READ
// ============================================================================

type FileReadTool struct{}

func (t *FileReadTool) Name() string        { return "read" }
func (t *FileReadTool) Description() string { return "Read the contents of a file" }

func (t *FileReadTool) Execute(ctx *Context, args map[string]string) Result {
	path := args["path"]
	if path == "" {
		return Result{Tool: t.Name(), Err: fmt.Errorf("missing 'path' argument")}
	}

	fullPath := resolvePath(ctx.WorkDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return Result{Tool: t.Name(), Err: err}
	}
	return Result{Tool: t.Name(), Output: string(data)}
}

// ============================================================================
// FILE WRITE
// ============================================================================

type FileWriteTool struct{}

func (t *FileWriteTool) Name() string        { return "write" }
func (t *FileWriteTool) Description() string { return "Write content to a file (creates or overwrites)" }

func (t *FileWriteTool) Execute(ctx *Context, args map[string]string) Result {
	path := args["path"]
	content := args["content"]
	if path == "" {
		return Result{Tool: t.Name(), Err: fmt.Errorf("missing 'path' argument")}
	}

	fullPath := resolvePath(ctx.WorkDir, path)

	if ctx.DryRun {
		return Result{Tool: t.Name(), Output: fmt.Sprintf("[dry-run] would write %d bytes to %s", len(content), path), Changed: []string{path}}
	}
	if !ctx.Approved {
		return Result{Tool: t.Name(), Err: fmt.Errorf("write requires approval (path: %s)", path)}
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return Result{Tool: t.Name(), Err: err}
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return Result{Tool: t.Name(), Err: err}
	}
	return Result{Tool: t.Name(), Output: fmt.Sprintf("wrote %d bytes to %s", len(content), path), Changed: []string{path}}
}

// ============================================================================
// FILE EDIT (find & replace)
// ============================================================================

type FileEditTool struct{}

func (t *FileEditTool) Name() string        { return "edit" }
func (t *FileEditTool) Description() string { return "Edit a file by replacing exact text" }

func (t *FileEditTool) Execute(ctx *Context, args map[string]string) Result {
	path := args["path"]
	oldText := args["old"]
	newText := args["new"]
	if path == "" || oldText == "" {
		return Result{Tool: t.Name(), Err: fmt.Errorf("missing 'path' or 'old' argument")}
	}

	fullPath := resolvePath(ctx.WorkDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return Result{Tool: t.Name(), Err: err}
	}

	content := string(data)
	count := strings.Count(content, oldText)
	if count == 0 {
		return Result{Tool: t.Name(), Err: fmt.Errorf("old text not found in %s", path)}
	}
	if count > 1 {
		return Result{Tool: t.Name(), Err: fmt.Errorf("old text found %d times in %s (must be unique)", count, path)}
	}

	updated := strings.Replace(content, oldText, newText, 1)

	if ctx.DryRun {
		return Result{Tool: t.Name(), Output: fmt.Sprintf("[dry-run] would replace %d chars in %s", len(oldText), path), Changed: []string{path}}
	}
	if !ctx.Approved {
		return Result{Tool: t.Name(), Err: fmt.Errorf("edit requires approval (path: %s)", path)}
	}

	if err := os.WriteFile(fullPath, []byte(updated), 0644); err != nil {
		return Result{Tool: t.Name(), Err: err}
	}
	return Result{Tool: t.Name(), Output: fmt.Sprintf("edited %s", path), Changed: []string{path}}
}

// ============================================================================
// LIST FILES
// ============================================================================

type ListFilesTool struct{}

func (t *ListFilesTool) Name() string        { return "list" }
func (t *ListFilesTool) Description() string { return "List files in a directory" }

func (t *ListFilesTool) Execute(ctx *Context, args map[string]string) Result {
	dir := args["dir"]
	if dir == "" {
		dir = "."
	}

	fullPath := resolvePath(ctx.WorkDir, dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return Result{Tool: t.Name(), Err: err}
	}

	var lines []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && name != ".git" {
			continue
		}
		if e.IsDir() {
			lines = append(lines, name+"/")
		} else {
			info, _ := e.Info()
			lines = append(lines, fmt.Sprintf("%-40s %8d", name, info.Size()))
		}
	}
	return Result{Tool: t.Name(), Output: strings.Join(lines, "\n")}
}

// ============================================================================
// BASH
// ============================================================================

type BashTool struct{}

func (t *BashTool) Name() string        { return "bash" }
func (t *BashTool) Description() string { return "Execute a shell command" }

func (t *BashTool) Execute(ctx *Context, args map[string]string) Result {
	cmdStr := args["command"]
	if cmdStr == "" {
		return Result{Tool: t.Name(), Err: fmt.Errorf("missing 'command' argument")}
	}

	if ctx.DryRun {
		return Result{Tool: t.Name(), Output: fmt.Sprintf("[dry-run] would run: %s", cmdStr)}
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = ctx.WorkDir
	out, err := cmd.CombinedOutput()
	return Result{
		Tool:   t.Name(),
		Output: strings.TrimRight(string(out), "\n"),
		Err:    err,
	}
}

// ============================================================================
// GIT STATUS
// ============================================================================

type GitStatusTool struct{}

func (t *GitStatusTool) Name() string        { return "git_status" }
func (t *GitStatusTool) Description() string { return "Show git working tree status" }

func (t *GitStatusTool) Execute(ctx *Context, args map[string]string) Result {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = ctx.WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Tool: t.Name(), Err: err}
	}
	return Result{Tool: t.Name(), Output: strings.TrimRight(string(out), "\n")}
}

// resolvePath makes a path absolute relative to the working directory.
func resolvePath(workDir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workDir, path)
}
