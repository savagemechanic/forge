package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Loader discovers, loads, and manages skills from the filesystem.
// Skills live in three locations:
//   - Built-in: bundled with Forge (skills/ directory)
//   - Global: ~/.forge/skills/
//   - Folder: .forge/skills/ in the project root
type Loader struct {
	builtinDir string
	globalDir  string
	folderDir  string
	skills     []*LoadedSkill
}

// LoadedSkill is a Skill with its loaded content and source location.
type LoadedSkill struct {
	Skill
	Body      string // raw content of the SKILL.md
	Directory string
	Source    string // "builtin", "global", "folder"
	Active    bool
}

// NewLoader creates a loader scanning the standard locations.
func NewLoader(forgeRoot, projectRoot string) *Loader {
	home, _ := os.UserHomeDir()
	return &Loader{
		builtinDir: filepath.Join(forgeRoot, "skills"),
		globalDir:  filepath.Join(home, ".forge", "skills"),
		folderDir:  filepath.Join(projectRoot, ".forge", "skills"),
	}
}

// Load scans all skill directories and loads every SKILL.md found.
func (l *Loader) Load() error {
	l.skills = nil

	// Built-in skills
	l.loadDir(l.builtinDir, "builtin")
	// Global skills
	l.loadDir(l.globalDir, "global")
	// Folder skills
	l.loadDir(l.folderDir, "folder")

	return nil
}

// loadDir loads all skills from a single directory.
func (l *Loader) loadDir(dir, source string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // directory doesn't exist, that's fine
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillDir := filepath.Join(dir, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}

		ls, err := loadSkillFile(skillFile, skillDir, source)
		if err != nil {
			continue
		}
		l.skills = append(l.skills, ls)
	}
}

// loadSkillFile parses a SKILL.md file into a LoadedSkill.
func loadSkillFile(path, dir, source string) (*LoadedSkill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	body := string(data)

	name := filepath.Base(dir)
	ls := &LoadedSkill{
		Skill: Skill{
			ID:         source + ":" + name,
			Name:       name,
			Scope:      scopeFromSource(source),
			Version:    "1.0.0",
			Entrypoint: "SKILL.md",
			Directory:  dir,
		},
		Body:      body,
		Directory: dir,
		Source:    source,
		Active:    true,
	}

	// Parse frontmatter / metadata from body
	parseSkillMetadata(ls, body)

	return ls, nil
}

// scopeFromSource maps a source string to a SkillScope.
func scopeFromSource(source string) SkillScope {
	switch source {
	case "builtin":
		return SkillScopeBuiltIn
	case "global":
		return SkillScopeGlobal
	case "folder":
		return SkillScopeFolder
	default:
		return SkillScopeFolder
	}
}

// parseSkillMetadata extracts a description from the SKILL.md body.
// It skips YAML frontmatter and headings, taking the first paragraph.
func parseSkillMetadata(ls *LoadedSkill, body string) {
	lines := strings.Split(body, "\n")

	// Skip YAML frontmatter (between --- delimiters at the start)
	start := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				start = i + 1
				break
			}
		}
	}

	// Also try to get description from frontmatter
	if start > 0 {
		for _, line := range lines[1:start] {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "description:") {
				ls.Description = strings.Trim(strings.TrimPrefix(line, "description:"), " \"")
			}
		}
	}

	// If no description from frontmatter, use first non-heading paragraph
	if ls.Description == "" {
		for _, line := range lines[start:] {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			ls.Description = line
			break
		}
	}
}

// All returns all loaded skills.
func (l *Loader) All() []*LoadedSkill { return l.skills }

// Active returns only active skills.
func (l *Loader) Active() []*LoadedSkill {
	var result []*LoadedSkill
	for _, s := range l.skills {
		if s.Active {
			result = append(result, s)
		}
	}
	return result
}

// Get returns a skill by name.
func (l *Loader) Get(name string) (*LoadedSkill, bool) {
	for _, s := range l.skills {
		if s.Name == name {
			return s, true
		}
	}
	return nil, false
}

// Activate enables a skill by name.
func (l *Loader) Activate(name string) error {
	for _, s := range l.skills {
		if s.Name == name {
			s.Active = true
			return nil
		}
	}
	return fmt.Errorf("skill not found: %s", name)
}

// Deactivate disables a skill by name.
func (l *Loader) Deactivate(name string) error {
	for _, s := range l.skills {
		if s.Name == name {
			s.Active = false
			return nil
		}
	}
	return fmt.Errorf("skill not found: %s", name)
}

// List returns a formatted string listing all skills with status.
func (l *Loader) List() string {
	if len(l.skills) == 0 {
		return "No skills installed."
	}
	var b strings.Builder
	for _, s := range l.skills {
		status := "○"
		if s.Active {
			status = "●"
		}
		fmt.Fprintf(&b, "  %s %-20s [%s] %s\n", status, s.Name, s.Source, s.Description)
	}
	return b.String()
}

// Install installs a skill directory to the global or folder location.
func (l *Loader) Install(name, source, body string, global bool) error {
	dir := l.folderDir
	if global {
		dir = l.globalDir
	}
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(body), 0644); err != nil {
		return err
	}
	// Reload
	return l.Load()
}
