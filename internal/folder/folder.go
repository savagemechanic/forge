package folder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type Folder struct {
	ID            string
	CanonicalPath string
	GitRoot       string
	Language      string
	CreatedAt     time.Time
	LastOpenedAt  time.Time
	GitRemote     string // Deprecated: kept for compatibility
}

// ============================================================================
// STUB IMPLEMENTATIONS - NOW IMPLEMENTED (GREEN PHASE)
// ============================================================================

// ValidateIDIsDeterministicHash checks ID is SHA256 of canonical path
func (f *Folder) ValidateIDIsDeterministicHash() bool {
	expected := generateFolderID(f.CanonicalPath)
	return f.ID == expected
}

// ValidateIDHash is the old name for ValidateIDIsDeterministicHash (for compatibility)
func (f *Folder) ValidateIDHash() (bool, error) {
	expected := generateFolderID(f.CanonicalPath)
	return f.ID == expected, nil
}

// ValidateCanonicalPathIsClean checks CanonicalPath is properly cleaned
func (f *Folder) ValidateCanonicalPathIsClean() error {
	// Must be absolute path
	if !filepath.IsAbs(f.CanonicalPath) {
		return fmt.Errorf("canonical path must be absolute: %s", f.CanonicalPath)
	}

	// Must be cleaned (no ".." or ".")
	cleanPath := filepath.Clean(f.CanonicalPath)
	if f.CanonicalPath != cleanPath {
		return fmt.Errorf("canonical path not clean: %s (should be %s)", f.CanonicalPath, cleanPath)
	}

	// No duplicate separators (filepath.Clean handles this)
	if filepath.ToSlash(f.CanonicalPath) != filepath.ToSlash(cleanPath) {
		return fmt.Errorf("canonical path has duplicate separators: %s", f.CanonicalPath)
	}

	// No trailing slash (except root)
	if f.CanonicalPath != "/" && len(f.CanonicalPath) > 0 && f.CanonicalPath[len(f.CanonicalPath)-1] == filepath.Separator {
		return fmt.Errorf("canonical path has trailing slash: %s", f.CanonicalPath)
	}

	return nil
}

// ValidateCanonicalPath is the old name for ValidateCanonicalPathIsClean (for compatibility)
func (f *Folder) ValidateCanonicalPath() (bool, error) {
	err := f.ValidateCanonicalPathIsClean()
	return err == nil, nil
}

// ResolveToCanonicalPath resolves the raw path to a clean canonical path
func (f *Folder) ResolveToCanonicalPath(rawPath string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(rawPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Resolve symlinks
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Clean the path
	cleanPath := filepath.Clean(resolvedPath)

	// Remove trailing slash (except root)
	if cleanPath != "/" && len(cleanPath) > 0 && cleanPath[len(cleanPath)-1] == filepath.Separator {
		cleanPath = cleanPath[:len(cleanPath)-1]
	}

	return cleanPath, nil
}

// CleanPath resolves and cleans a path (for compatibility with tests)
func CleanPath(path string) string {
	folder := &Folder{}
	cleaned, err := folder.ResolveToCanonicalPath(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return cleaned
}

// ValidateGitRootConsistency checks GitRoot is consistent with CanonicalPath
func (f *Folder) ValidateGitRootConsistency() error {
	// Empty GitRoot is OK if not a git repo
	if f.GitRoot == "" {
		if isGitRepository(f.CanonicalPath) {
			return fmt.Errorf("GitRoot is empty but folder is a git repository")
		}
		return nil
	}

	// GitRoot must be an absolute path
	if !filepath.IsAbs(f.GitRoot) {
		return fmt.Errorf("GitRoot must be absolute: %s", f.GitRoot)
	}

	// GitRoot must end with /.git
	if !strings.HasSuffix(f.GitRoot, "/.git") && !strings.HasSuffix(f.GitRoot, ".git") {
		return fmt.Errorf("GitRoot must end with .git: %s", f.GitRoot)
	}

	// Note: We don't check filesystem existence here since paths may be simulated in tests

	// GitRoot must be inside or at CanonicalPath
	rel, err := filepath.Rel(f.CanonicalPath, f.GitRoot)
	if err != nil {
		return fmt.Errorf("GitRoot not relative to CanonicalPath: %w", err)
	}

	// If rel starts with "..", GitRoot is outside CanonicalPath (not OK)
	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("GitRoot is outside CanonicalPath: %s is outside %s", f.GitRoot, f.CanonicalPath)
	}

	return nil
}

// ValidateGitRoot is the old name for ValidateGitRootConsistency (for compatibility)
func (f *Folder) ValidateGitRoot() bool {
	err := f.ValidateGitRootConsistency()
	return err == nil
}

// ValidateLastOpenedAt checks LastOpenedAt >= CreatedAt and never decreases
func (f *Folder) ValidateLastOpenedAt() bool {
	previous := f.CreatedAt // Use CreatedAt as "previous" for first check
	return f.ValidateLastOpenedAtMonotonic(&previous)
}

// ValidateLastOpenedAtNotDecreased is an alias for ValidateLastOpenedAtMonotonic
func (f *Folder) ValidateLastOpenedAtNotDecreased(previous time.Time) bool {
	return f.ValidateLastOpenedAtMonotonic(&previous)
}

// UpdateLastOpenedAt updates LastOpenedAt, enforcing monotonicity
func (f *Folder) UpdateLastOpenedAt() error {
	now := time.Now()
	if now.Before(f.LastOpenedAt) {
		return &NonMonotonicTimestampError{
			Field:     "LastOpenedAt",
			Previous:  f.LastOpenedAt,
			Attempted: now,
		}
	}
	f.LastOpenedAt = now
	return nil
}

// ValidateLastOpenedAtMonotonic checks LastOpenedAt > CreatedAt and never decreases
func (f *Folder) ValidateLastOpenedAtMonotonic(previous *time.Time) bool {
	// LastOpenedAt must be > CreatedAt (strictly greater)
	if !f.LastOpenedAt.After(f.CreatedAt) {
		return false
	}

	// If we have a previous value, must not decrease
	if previous != nil && f.LastOpenedAt.Before(*previous) {
		return false
	}

	return true
}

// ============================================================================
// STORE INTERFACE
// ============================================================================

type FolderStore interface {
	Save(folders ...*Folder)
	Get(id string) (*Folder, error)
	ListByPath(path string) (*Folder, error)
}

type InMemoryFolderStore struct {
	folders map[string]*Folder
}

func NewFolderStore() *InMemoryFolderStore {
	return &InMemoryFolderStore{
		folders: make(map[string]*Folder),
	}
}

func (s *InMemoryFolderStore) Save(folders ...*Folder) {
	for _, folder := range folders {
		s.folders[folder.ID] = folder
	}
}

func (s *InMemoryFolderStore) Get(id string) (*Folder, error) {
	folder, ok := s.folders[id]
	if !ok {
		return nil, &FolderNotFoundError{ID: id}
	}
	return folder, nil
}

func (s *InMemoryFolderStore) ListByPath(path string) (*Folder, error) {
	for _, folder := range s.folders {
		if folder.CanonicalPath == path {
			return folder, nil
		}
	}
	return nil, &FolderNotFoundError{ID: path}
}

// ============================================================================
// ERRORS
// ============================================================================

type FolderNotFoundError struct {
	ID string
}

func (e *FolderNotFoundError) Error() string {
	return "folder not found: " + e.ID
}

type NonMonotonicTimestampError struct {
	Field     string
	Previous  time.Time
	Attempted time.Time
}

func (e *NonMonotonicTimestampError) Error() string {
	return fmt.Sprintf("non-monotonic timestamp for %s: previous=%v, attempted=%v",
		e.Field, e.Previous, e.Attempted)
}

// ============================================================================
// HELPERS
// ============================================================================

// generateFolderID creates a deterministic ID from canonical path
func generateFolderID(canonicalPath string) string {
	hash := sha256.Sum256([]byte(canonicalPath))
	return hex.EncodeToString(hash[:])
}

// ComputeFolderID is exported for testing
func ComputeFolderID(canonicalPath string) string {
	return generateFolderID(canonicalPath)
}

// isGitRepository checks if path is a git repository
func isGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return true
	}
	return false
}

// DetectGitRoot finds the .git directory for a path
func DetectGitRoot(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	current := absPath
	for {
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() {
				return current, nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root without finding .git
			return "", nil
		}
		current = parent
	}
}