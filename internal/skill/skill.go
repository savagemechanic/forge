package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type SkillScope string

const (
	SkillScopeBuiltIn SkillScope = "built-in"
	SkillScopeGlobal  SkillScope = "global"
	SkillScopeFolder  SkillScope = "folder"
)

type Trigger struct {
	Type     string // "keyword", "intent", "pattern"
	Pattern  string
	Weight   float64
}

type Precondition struct {
	Type  string // "go-version", "project-type", "file-exists"
	Value string
}

type Permission struct {
	Tool           string
	AllowedPaths   []string
	DeniedPaths    []string
	MaxFileSize    int64
	AllowedNetwork bool
}

type VersionEntry struct {
	Version string
	At      time.Time
	By      string
	Reason  string
}

type Skill struct {
	ID             string
	Name           string
	Description    string
	Scope          SkillScope
	Triggers       []Trigger
	Preconditions  []Precondition
	Permissions    []Permission
	Entrypoint     string
	Version        string
	VersionHistory []VersionEntry
	Directory      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ToolExecution struct {
	Tool string
	Args map[string]string
}

// ============================================================================
// STUB IMPLEMENTATIONS (RED PHASE - ALL WILL FAIL)
// ============================================================================

// IsToolAllowed checks if a tool is in the permissions whitelist
// STUB: Always returns false to fail tests
func (s *Skill) IsToolAllowed(tool string) bool {
	// TODO: Implement permission lookup
	// Check if tool exists in s.Permissions
	// For now, always fail
	return false
}

// IsToolAllowedWithPath checks if a tool+path combination is allowed
// STUB: Always returns false to fail tests
func (s *Skill) IsToolAllowedWithPath(tool, path string) bool {
	// TODO: Implement permission lookup with path constraints
	// Check tool permissions + allowed/denied paths
	// For now, always fail
	return false
}

// ValidateExecution checks if an execution is permitted
// STUB: Always returns error to fail tests
func (s *Skill) ValidateExecution(exec *ToolExecution) error {
	// TODO: Implement execution validation
	// - Check tool is in permissions
	// - Check path constraints
	// - Check network access
	// For now, always return error
	return fmt.Errorf("not implemented")
}

// ValidateSemVer checks that version follows SemVer
// STUB: Always returns false to fail tests
func (s *Skill) ValidateSemVer() bool {
	// TODO: Implement SemVer validation
	// Format: MAJOR.MINOR.PATCH (all numeric, no leading zeros)
	// For now, always fail
	return false
}

// IncrementVersion increments the skill version
// STUB: Always returns error to fail tests
func (s *Skill) IncrementVersion(store SkillStore, authority string) error {
	// TODO: Implement version increment with authorization check
	// - Authority cannot be "self"
	// - Increment PATCH by default
	// - Record in version history
	// For now, always return error
	return fmt.Errorf("self-increment not allowed")
}

// RecordVersionChange records a version change in history
// STUB: Does nothing to fail tests
func (s *Skill) RecordVersionChange(newVersion string, at time.Time, by string) {
	// TODO: Implement version history recording
	// For now, do nothing
}

// ValidateVersionHistory checks version history is ordered
// STUB: Always returns false to fail tests
func (s *Skill) ValidateVersionHistory() bool {
	// TODO: Implement version history validation
	// - Each entry must be in chronological order
	// - Versions must be monotonically increasing (by SemVer)
	// For now, always fail
	return false
}

// ValidateEntrypoint checks that entrypoint file exists
// STUB: Always returns false to fail tests
func (s *Skill) ValidateEntrypoint() bool {
	// TODO: Implement entrypoint validation
	// - Entrypoint must not be empty
	// - File must exist in skill directory
	// For now, always fail
	return false
}

// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// ValidateNameScopeUnique checks no two skills have same name+scope
// STUB: Always returns false to fail tests
func ValidateNameScopeUnique(store SkillStore) (bool, error) {
	// TODO: Implement duplicate detection
	// Group by (name, scope), ensure max 1 per group
	// For now, always fail
	return false, nil
}

// ============================================================================
// STORE INTERFACE
// ============================================================================

type SkillStore interface {
	Save(skills ...*Skill)
	Get(id string) (*Skill, error)
	GetByNameAndScope(name string, scope SkillScope) (*Skill, error)
	ListByScope(scope SkillScope) ([]*Skill, error)
}

type MemorySkillStore struct {
	skills map[string]*Skill
}

func NewMemorySkillStore() *MemorySkillStore {
	return &MemorySkillStore{
		skills: make(map[string]*Skill),
	}
}

func (s *MemorySkillStore) Save(skills ...*Skill) {
	for _, skill := range skills {
		s.skills[skill.ID] = skill
	}
}

func (s *MemorySkillStore) Get(id string) (*Skill, error) {
	skill, ok := s.skills[id]
	if !ok {
		return nil, &SkillNotFoundError{ID: id}
	}
	return skill, nil
}

func (s *MemorySkillStore) GetByNameAndScope(name string, scope SkillScope) (*Skill, error) {
	for _, skill := range s.skills {
		if skill.Name == name && skill.Scope == scope {
			return skill, nil
		}
	}
	return nil, &SkillNotFoundError{ID: fmt.Sprintf("%s@%s", name, scope)}
}

func (s *MemorySkillStore) ListByScope(scope SkillScope) ([]*Skill, error) {
	// STUB: Always return error
	return nil, &NotImplementedError{}
}

// ============================================================================
// ERRORS
// ============================================================================

type SkillNotFoundError struct {
	ID string
}

func (e *SkillNotFoundError) Error() string {
	return "skill not found: " + e.ID
}

type PermissionDeniedError struct {
	Tool string
	Reason string
}

func (e *PermissionDeniedError) Error() string {
	return fmt.Sprintf("permission denied for tool '%s': %s", e.Tool, e.Reason)
}

type NotImplementedError struct{}

func (e *NotImplementedError) Error() string {
	return "not implemented"
}

// ============================================================================
// HELPERS
// ============================================================================

// isValidSemVer checks if version string is valid SemVer
func isValidSemVer(version string) bool {
	// STUB: Always return false
	return false
}

// compareSemVer compares two SemVer strings
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareSemVer(a, b string) int {
	// STUB: Always return -1
	return -1
}

// incrementSemVer increments SemVer
func incrementSemVer(version string, part string) (string, error) {
	// STUB: Always return error
	return "", &NotImplementedError{}
}

// getSkillDirectory returns the directory for a skill based on scope
func getSkillDirectory(scope SkillScope, name string) string {
	homeDir, _ := os.UserHomeDir()

	switch scope {
	case SkillScopeBuiltIn:
		return filepath.Join("dist", "skills", name)
	case SkillScopeGlobal:
		return filepath.Join(homeDir, ".forge", "skills", name)
	case SkillScopeFolder:
		return filepath.Join(".forge", "skills", name)
	default:
		return ""
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// normalizeToolName normalizes a tool name
func normalizeToolName(tool string) string {
	return strings.ToLower(strings.TrimSpace(tool))
}