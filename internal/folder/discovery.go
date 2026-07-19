package folder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType describes the kind of project discovered in a folder.
type ProjectType string

const (
	ProjectTypeGo       ProjectType = "go"
	ProjectTypeGoModule ProjectType = "go-module"
	ProjectTypeGoWorkspace ProjectType = "go-workspace"
	ProjectTypeGit      ProjectType = "git"
	ProjectTypeUnknown  ProjectType = "unknown"
)

// ProjectInfo holds discovered facts about a folder.
type ProjectInfo struct {
	RootPath    string       // absolute path to project root (git root or dir with go.mod)
	GitRoot     string       // absolute path containing .git, "" if none
	Types       []ProjectType // detected project types
	GoModule    string       // module path from go.mod, "" if none
	HasGoWork   bool         // true if go.work exists
	Subpackages []string     // list of subpackages under root
}

// Discover inspects the given directory and returns project info.
// It walks up the tree to find the git root and go.mod/go.work files.
func Discover(startDir string) (*ProjectInfo, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path: %w", err)
	}

	info := &ProjectInfo{RootPath: abs}

	// Walk up to find git root and project markers
	current := abs
	for {
		// Check for .git
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			info.GitRoot = current
		}

		// Check for go.mod
		if goMod, err := readGoMod(filepath.Join(current, "go.mod")); err == nil {
			info.GoModule = goMod
			info.RootPath = current
			info.Types = append(info.Types, ProjectTypeGo, ProjectTypeGoModule)
		}

		// Check for go.work
		if _, err := os.Stat(filepath.Join(current, "go.work")); err == nil {
			info.HasGoWork = true
			info.RootPath = current
			if !containsType(info.Types, ProjectTypeGoWorkspace) {
				info.Types = append(info.Types, ProjectTypeGoWorkspace)
			}
		}

		// Stop at filesystem root
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	// If we found a git root but no go.mod at root, root is git root
	if info.GitRoot != "" && info.RootPath == abs && !containsType(info.Types, ProjectTypeGoModule) {
		info.RootPath = info.GitRoot
	}

	// Mark as git type
	if info.GitRoot != "" {
		if !containsType(info.Types, ProjectTypeGit) {
			info.Types = append(info.Types, ProjectTypeGit)
		}
	}

	// Default to unknown if nothing detected
	if len(info.Types) == 0 {
		info.Types = []ProjectType{ProjectTypeUnknown}
	}

	// Discover subpackages if it's a Go project
	if containsType(info.Types, ProjectTypeGo) {
		info.Subpackages = discoverGoPackages(info.RootPath)
	}

	return info, nil
}

// Summary returns a human-readable one-line description of the project.
func (p *ProjectInfo) Summary() string {
	if len(p.Types) == 0 || (len(p.Types) == 1 && p.Types[0] == ProjectTypeUnknown) {
		return "unknown project"
	}

	parts := []string{}
	if p.GoModule != "" {
		parts = append(parts, fmt.Sprintf("Go module %s", p.GoModule))
	}
	if p.HasGoWork {
		parts = append(parts, "go.work workspace")
	}
	if p.GitRoot != "" {
		rel, err := filepath.Rel(p.GitRoot, p.RootPath)
		if err != nil || rel == "." {
			parts = append(parts, "git root")
		} else {
			parts = append(parts, fmt.Sprintf("git (%s)", rel))
		}
	}
	if len(p.Subpackages) > 0 {
		parts = append(parts, fmt.Sprintf("%d packages", len(p.Subpackages)))
	}

	return strings.Join(parts, " · ")
}

// readGoMod reads the module path from a go.mod file.
func readGoMod(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("no module directive in %s", path)
}

// discoverGoPackages finds subdirectories containing .go files.
func discoverGoPackages(root string) []string {
	var pkgs []string
	skipDirs := map[string]bool{
		"vendor": true, ".git": true, "node_modules": true,
		"testdata": true, "testdata_": true, "_testdata": true,
	}

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				if name != filepath.Base(root) {
					return filepath.SkipDir
				}
			}
			// Check if dir contains .go files
			entries, _ := os.ReadDir(path)
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
					rel, _ := filepath.Rel(root, path)
					if rel != "." {
						pkgs = append(pkgs, rel)
					}
					break
				}
			}
		}
		return nil
	})

	return pkgs
}

func containsType(types []ProjectType, t ProjectType) bool {
	for _, x := range types {
		if x == t {
			return true
		}
	}
	return false
}
