package folder

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
	GitRemote     string
	GitRoot       string
	CreatedAt     time.Time
	LastOpenedAt  time.Time
}

// ============================================================================
// STUB IMPLEMENTATIONS (RED PHASE - ALL WILL FAIL)
// ============================================================================

// ValidateIDHash checks that ID is SHA256 hash of CanonicalPath
// STUB: Always returns false to fail tests
func (f *Folder) ValidateIDHash() (bool, error) {
	// TODO: Implement SHA256 hash validation
	// ID = hex.Encode(sha256(CanonicalPath))
	// For now, always fail
	return false, nil
}

// ValidateCanonicalPath checks that path is cleaned and absolute
// STUB: Always returns false to fail tests
func (f *Folder) ValidateCanonicalPath() (bool, error) {
	// TODO: Implement path validation
	// - Must be absolute
	// - No .. segments
	// - No trailing slash
	// - No duplicate slashes
	// - Symlinks resolved
	// For now, always fail
	return false, nil
}

// ValidateGitRoot checks GitRoot consistency
// STUB: Always returns false to fail tests
func (f *Folder) ValidateGitRoot() bool {
	// TODO: Implement GitRoot validation
	// - If empty, must not be a Git repo
	// - If set, must end with /.git
	// - Must be inside or equal to CanonicalPath
	// For now, always fail
	return false
}

// ValidateLastOpenedAt checks LastOpenedAt >= CreatedAt
// STUB: Always returns false to fail tests
func (f *Folder) ValidateLastOpenedAt() bool {
	// TODO: Implement monotonic validation
	// LastOpenedAt must be > CreatedAt (not equal)
	// For now, always fail
	return false
}

// ValidateLastOpenedAtNotDecreased checks that LastOpenedAt hasn't decreased
// STUB: Always returns false to fail tests
func (f *Folder) ValidateLastOpenedAtNotDecreased(previous time.Time) bool {
	// TODO: Implement monotonic increase check
	// LastOpenedAt must be >= previous
	// For now, always fail
	return false
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// ComputeFolderID generates deterministic ID from path
// STUB: Returns wrong value to fail tests
func ComputeFolderID(path string) string {
	// TODO: Implement proper SHA256 hashing
	// For now, return wrong value
	return "stub-id-not-a-hash"
}

// CleanPath cleans a path to canonical form
// STUB: Returns wrong value to fail tests
func CleanPath(path string) string {
	// TODO: Implement proper path cleaning
	// - filepath.Abs
	// - filepath.Clean
	// - EvalSymlinks
	// - Remove trailing slash
	// For now, return unmodified
	return path
}

// DiscoverFolder discovers folder identity from current directory
// Resolution order: Git root → go.work → go.mod → cwd
func DiscoverFolder(cwd string) (*Folder, error) {
	// Try Git root first
	if gitRoot := FindGitRoot(cwd); gitRoot != "" {
		folderPath := filepath.Dir(gitRoot)
		return CreateFolder(folderPath, gitRoot)
	}

	// Try go.work
	if workRoot := FindGoWork(cwd); workRoot != "" {
		return CreateFolder(workRoot, "")
	}

	// Try go.mod
	if modRoot := FindGoMod(cwd); modRoot != "" {
		return CreateFolder(modRoot, "")
	}

	// Use current directory
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		return nil, err
	}

	return CreateFolder(absPath, "")
}

// CreateFolder creates a new folder with canonical path
func CreateFolder(canonicalPath string, gitRoot string) (*Folder, error) {
	now := time.Now()
	folder := &Folder{
		ID:            ComputeFolderID(canonicalPath),
		CanonicalPath: canonicalPath,
		GitRoot:       gitRoot,
		CreatedAt:     now,
		LastOpenedAt:  now,
	}

	// Get Git remote if available
	if gitRoot != "" {
		if remote, err := GetGitRemote(gitRoot); err == nil {
			folder.GitRemote = remote
		}
	}

	return folder, nil
}

// ============================================================================
// DISCOVERY HELPERS
// ============================================================================

// FindGitRoot walks up from path looking for .git directory
func FindGitRoot(path string) string {
	for {
		gitPath := filepath.Join(path, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return gitPath
		}

		parent := filepath.Dir(path)
		if parent == path {
			return "" // Reached root
		}
		path = parent
	}
}

// FindGoWork walks up from path looking for go.work
func FindGoWork(path string) string {
	for {
		workPath := filepath.Join(path, "go.work")
		if _, err := os.Stat(workPath); err == nil {
			return path
		}

		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}

// FindGoMod walks up from path looking for go.mod
func FindGoMod(path string) string {
	for {
		modPath := filepath.Join(path, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return path
		}

		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}

// GetGitRemote extracts the primary Git remote URL
func GetGitRemote(gitRoot string) (string, error) {
	// TODO: Implement by reading .git/config
	return "", errors.New("not implemented")
}

// ============================================================================
// VALIDATION HELPERS (FOR PROPER IMPLEMENTATION)
// ============================================================================

// computeSHA256 computes SHA256 hash of string
func computeSHA256(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// isAbsolutePath checks if path is absolute
func isAbsolutePath(path string) bool {
	return strings.HasPrefix(path, "/") || strings.Contains(path, ":\\")
}

// hasDotDot checks if path contains .. segments
func hasDotDot(path string) bool {
	segments := strings.Split(path, string(filepath.Separator))
	for _, seg := range segments {
		if seg == ".." {
			return true
		}
	}
	return false
}

// endsWithGit checks if path ends with .git
func endsWithGit(path string) bool {
	return filepath.Base(path) == ".git"
}

// isInsideOrEqual checks if child is inside or equal to parent
func isInsideOrEqual(child, parent string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || !strings.HasPrefix(rel, "..")
}