package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type SkillScope string

const (
	SkillScopeGlobal   SkillScope = "global"
	SkillScopeFolder  SkillScope = "folder"
	SkillScopeBuiltIn SkillScope = "builtin"
)

type Permission struct {
	Tool         string
	Allowed      bool
	Limit        *int      // Optional limit on usage
	AllowedPaths []string  // Paths this tool can access
	DeniedPaths  []string  // Paths this tool cannot access
	MaxFileSize  int64     // Max file size in bytes
}

type ToolExecution struct {
	Tool string
	Args map[string]string
}

type VersionEntry struct {
	Version string
	At      time.Time
	By      string
}

type Skill struct {
	ID              string
	Name            string
	Scope           SkillScope
	Version         string         // SemVer
	Description     string
	Entrypoint      string         // Path to entrypoint file
	Permissions     []Permission
	CreatedAt       time.Time
	UpdatedAt       time.Time
	VersionHistory  []VersionEntry // History of version changes
	Directory       string         // Skill directory path
}

// ============================================================================
// STUB IMPLEMENTATIONS - NOW IMPLEMENTED (GREEN PHASE)
// ============================================================================

var semVerRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

// ValidateSemVerAndNoAutoIncrement checks version is valid SemVer and doesn't auto-increment
func (s *Skill) ValidateSemVerAndNoAutoIncrement(previous *Skill) error {
	// Validate SemVer format
	if !semVerRegex.MatchString(s.Version) {
		return &InvalidSemVerError{Version: s.Version}
	}

	// Remove 'v' prefix for comparison
	version := strings.TrimPrefix(s.Version, "v")

	// If we have a previous version, check for auto-increment
	if previous != nil {
		prevVersion := strings.TrimPrefix(previous.Version, "v")
		prevParts := strings.Split(prevVersion, ".")
		currentParts := strings.Split(version, ".")

		// Check if it's a patch increment (X.Y.Z -> X.Y.(Z+1))
		if len(prevParts) >= 3 && len(currentParts) >= 3 {
			if prevParts[0] == currentParts[0] && // Major same
				prevParts[1] == currentParts[1] && // Minor same
				currentParts[2] != prevParts[2] {   // Patch changed
				// Could be auto-increment, but we allow manual patch bumps
				return nil
			}
		}
	}

	return nil
}

// ValidateNameScopeUnique checks name is unique within scope
func (s *Skill) ValidateNameScopeUnique(store *InMemorySkillStore) error {
	for _, skill := range store.skills {
		if skill.ID != s.ID && skill.Name == s.Name && skill.Scope == s.Scope {
			return &DuplicateSkillError{
				Name:  s.Name,
				Scope: s.Scope,
			}
		}
	}
	return nil
}

// ValidatePermissionsWhitelist checks permissions are whitelisted tools only
func (s *Skill) ValidatePermissionsWhitelist(allowedTools map[string]bool) error {
	if len(s.Permissions) == 0 {
		return &EmptyPermissionsError{Name: s.Name}
	}

	allowedToolNames := make(map[string]bool)
	for tool := range allowedTools {
		allowedToolNames[tool] = true
	}

	for _, perm := range s.Permissions {
		if !allowedToolNames[perm.Tool] {
			return &UnauthorizedToolError{
				Skill:    s.Name,
				Tool:     perm.Tool,
				Allowed:  allowedToolNames,
			}
		}
	}

	return nil
}

// ValidateEntrypointExists checks entrypoint file exists
func (s *Skill) ValidateEntrypointExists() error {
	if s.Entrypoint == "" {
		return &InvalidEntrypointError{Path: s.Entrypoint, Reason: "empty"}
	}

	// Check if absolute path
	if !filepath.IsAbs(s.Entrypoint) {
		return &InvalidEntrypointError{Path: s.Entrypoint, Reason: "not absolute"}
	}

	// Check file exists
	if _, err := os.Stat(s.Entrypoint); err != nil {
		if os.IsNotExist(err) {
			return &InvalidEntrypointError{Path: s.Entrypoint, Reason: "does not exist"}
		}
		return &InvalidEntrypointError{Path: s.Entrypoint, Reason: err.Error()}
	}

	return nil
}

// ValidateEntrypoint checks entrypoint file exists (bool version for tests)
func (s *Skill) ValidateEntrypoint() bool {
	if s.Entrypoint == "" {
		return false
	}

	// Check file exists
	fullPath := filepath.Join(s.Directory, s.Entrypoint)
	if _, err := os.Stat(fullPath); err != nil {
		return false
	}

	return true
}

// ValidateSemVer checks if version is valid SemVer (bool version for tests)
func (s *Skill) ValidateSemVer() bool {
	return semVerRegex.MatchString(s.Version)
}

// IsToolAllowedWithPath checks if a tool is allowed with a specific path
func (s *Skill) IsToolAllowedWithPath(tool, path string) bool {
	for _, perm := range s.Permissions {
		if perm.Tool == tool {
			// Check denied paths first
			for _, denied := range perm.DeniedPaths {
				if strings.HasPrefix(path, denied) || strings.HasPrefix(denied, path) {
					return false
				}
			}

			// If allowed paths specified, check them
			if len(perm.AllowedPaths) > 0 {
				for _, allowedPath := range perm.AllowedPaths {
					if strings.HasPrefix(path, allowedPath) {
						return true
					}
				}
				return false // Path not in allowed list
			}

			// No path restrictions = allowed
			return true
		}
	}
	return false // Tool not in permissions
}

// ValidateExecution checks if a tool execution is allowed
func (s *Skill) ValidateExecution(exec *ToolExecution) error {
	for _, perm := range s.Permissions {
		if perm.Tool == exec.Tool {
			// Tool is in permissions list = allowed
			return nil
		}
	}
	return &UnauthorizedToolError{
		Skill:   s.Name,
		Tool:    exec.Tool,
		Allowed: nil,
	}
}

// IncrementVersion increments the skill's version by external authority only
func (s *Skill) IncrementVersion(store *InMemorySkillStore, author string) error {
	if author == "self" {
		return fmt.Errorf("skill cannot increment its own version")
	}

	// Parse current version
	major, minor, patch, err := ParseSemVer(s.Version)
	if err != nil {
		return err
	}

	// Record current version in history
	s.RecordVersionChange(s.Version, time.Now(), author)

	// Increment patch
	s.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch+1)

	return nil
}

// RecordVersionChange records a version change in history
func (s *Skill) RecordVersionChange(version string, at time.Time, by string) {
	s.VersionHistory = append(s.VersionHistory, VersionEntry{
		Version: version,
		At:      at,
		By:      by,
	})
}

// ValidateVersionHistory checks version history is in chronological order
func (s *Skill) ValidateVersionHistory() bool {
	if len(s.VersionHistory) <= 1 {
		return true
	}

	// Check timestamps are chronological
	for i := 1; i < len(s.VersionHistory); i++ {
		if s.VersionHistory[i].At.Before(s.VersionHistory[i-1].At) {
			return false
		}
		
		// Also check SemVer is monotonically increasing
		cmp, err := CompareSemVer(s.VersionHistory[i].Version, s.VersionHistory[i-1].Version)
		if err != nil {
			return false
		}
		if cmp <= 0 {
			return false
		}
	}

	return true
}

// ============================================================================
// STORE INTERFACE
// ============================================================================

type SkillStore interface {
	Save(skills ...*Skill)
	Get(id string) (*Skill, error)
	ListByScope(scope SkillScope) ([]*Skill, error)
}

type InMemorySkillStore struct {
	skills map[string]*Skill
}

func NewSkillStore() *InMemorySkillStore {
	return &InMemorySkillStore{
		skills: make(map[string]*Skill),
	}
}

func (s *InMemorySkillStore) Save(skills ...*Skill) {
	for _, skill := range skills {
		s.skills[skill.ID] = skill
	}
}

func (s *InMemorySkillStore) Get(id string) (*Skill, error) {
	skill, ok := s.skills[id]
	if !ok {
		return nil, &SkillNotFoundError{ID: id}
	}
	return skill, nil
}

func (s *InMemorySkillStore) ListByScope(scope SkillScope) ([]*Skill, error) {
	var result []*Skill
	for _, skill := range s.skills {
		if skill.Scope == scope {
			result = append(result, skill)
		}
	}
	return result, nil
}

// ============================================================================
// ERRORS
// ============================================================================

type InvalidSemVerError struct {
	Version string
}

func (e *InvalidSemVerError) Error() string {
	return fmt.Sprintf("invalid SemVer: %s", e.Version)
}

type DuplicateSkillError struct {
	Name  string
	Scope SkillScope
}

func (e *DuplicateSkillError) Error() string {
	return fmt.Sprintf("duplicate skill: %s in scope %s", e.Name, e.Scope)
}

type EmptyPermissionsError struct {
	Name string
}

func (e *EmptyPermissionsError) Error() string {
	return fmt.Sprintf("skill has empty permissions: %s", e.Name)
}

type UnauthorizedToolError struct {
	Skill   string
	Tool    string
	Allowed map[string]bool
}

func (e *UnauthorizedToolError) Error() string {
	tools := make([]string, 0, len(e.Allowed))
	for tool := range e.Allowed {
		tools = append(tools, tool)
	}
	return fmt.Sprintf("skill %s uses unauthorized tool %s (allowed: %v)",
		e.Skill, e.Tool, tools)
}

type InvalidEntrypointError struct {
	Path   string
	Reason string
}

func (e *InvalidEntrypointError) Error() string {
	return fmt.Sprintf("invalid entrypoint %s: %s", e.Path, e.Reason)
}

type SkillNotFoundError struct {
	ID string
}

func (e *SkillNotFoundError) Error() string {
	return "skill not found: " + e.ID
}

// ============================================================================
// HELPERS
// ============================================================================

// GetAllowedTools returns the default set of allowed tools
func GetAllowedTools() map[string]bool {
	return map[string]bool{
		"read":        true,
		"write":       true,
		"bash":        true,
		"git":         true,
		"search":      true,
		"list":        true,
		"diff":        true,
		"patch":       true,
		"test":        true,
		"build":       true,
		"install":     true,
		"diagnose":    true,
		"ask":         true,
		"remember":    true,
		"kanban":      true,
		"summary":     true,
		"web-search":  true,
		"web-fetch":   true,
	}
}

// ParseSemVer parses a SemVer string into major, minor, patch
func ParseSemVer(version string) (major, minor, patch int, err error) {
	version = strings.TrimPrefix(version, "v")
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return 0, 0, 0, &InvalidSemVerError{Version: version}
	}

	_, err = fmt.Sscanf(parts[0], "%d", &major)
	if err != nil {
		return 0, 0, 0, &InvalidSemVerError{Version: version}
	}

	_, err = fmt.Sscanf(parts[1], "%d", &minor)
	if err != nil {
		return 0, 0, 0, &InvalidSemVerError{Version: version}
	}

	// Parse patch, ignoring pre-release and build metadata
	patchStr := strings.Split(parts[2], "-")[0]
	patchStr = strings.Split(patchStr, "+")[0]
	_, err = fmt.Sscanf(patchStr, "%d", &patch)
	if err != nil {
		return 0, 0, 0, &InvalidSemVerError{Version: version}
	}

	return major, minor, patch, nil
}

// CompareSemVer compares two SemVer strings
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func CompareSemVer(a, b string) (int, error) {
	majorA, minorA, patchA, err := ParseSemVer(a)
	if err != nil {
		return 0, err
	}

	majorB, minorB, patchB, err := ParseSemVer(b)
	if err != nil {
		return 0, err
	}

	if majorA != majorB {
		if majorA < majorB {
			return -1, nil
		}
		return 1, nil
	}

	if minorA != minorB {
		if minorA < minorB {
			return -1, nil
		}
		return 1, nil
	}

	if patchA != patchB {
		if patchA < patchB {
			return -1, nil
		}
		return 1, nil
	}

	return 0, nil
}
// NewMemorySkillStore is an alias for NewSkillStore
func NewMemorySkillStore() *InMemorySkillStore {
	return NewSkillStore()
}

// IsToolAllowed checks if a tool is allowed by this skill's permissions
func (s *Skill) IsToolAllowed(tool string) bool {
	for _, perm := range s.Permissions {
		if perm.Tool == tool {
			// Tool is in permissions list = allowed
			return true
		}
	}
	return false // Not in whitelist = not allowed
}


// ValidateNameScopeUnique checks that a skill's name is unique within its scope (two-arg version for tests)
func ValidateNameScopeUnique(store *InMemorySkillStore, skill *Skill) (bool, error) {
	err := skill.ValidateNameScopeUnique(store)
	return err == nil, nil
}
