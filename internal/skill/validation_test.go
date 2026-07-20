package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: ValidateSemVerAndNoAutoIncrement validates format and detects
// suspicious patch-only increments (possible auto-increment).

func TestValidateSemVerAndNoAutoIncrement(t *testing.T) {
	// Valid format, no previous
	s := &Skill{Version: "1.0.0"}
	assert.NoError(t, s.ValidateSemVerAndNoAutoIncrement(nil))

	// Invalid format
	s = &Skill{Version: "not-semver"}
	assert.Error(t, s.ValidateSemVerAndNoAutoIncrement(nil))

	// Previous with patch bump (allowed)
	prev := &Skill{Version: "1.0.0"}
	s = &Skill{Version: "1.0.1"}
	assert.NoError(t, s.ValidateSemVerAndNoAutoIncrement(prev))
}

// INVARIANT: ValidatePermissionsWhitelist — all tools must be in allowed set.

func TestValidatePermissionsWhitelist(t *testing.T) {
	allowed := map[string]bool{"read": true, "write": true}

	// All permitted
	s := &Skill{
		Name:        "ok",
		Permissions: []Permission{{Tool: "read"}, {Tool: "write"}},
	}
	assert.NoError(t, s.ValidatePermissionsWhitelist(allowed))

	// Empty permissions error
	s = &Skill{Name: "empty", Permissions: []Permission{}}
	assert.Error(t, s.ValidatePermissionsWhitelist(allowed))

	// Unauthorized tool
	s = &Skill{
		Name:        "bad",
		Permissions: []Permission{{Tool: "dangerous"}},
	}
	err := s.ValidatePermissionsWhitelist(allowed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dangerous")
}

// INVARIANT: ValidateEntrypointExists — empty or non-absolute or missing errors.

func TestValidateEntrypointExists(t *testing.T) {
	// Empty
	s := &Skill{Entrypoint: ""}
	err := s.ValidateEntrypointExists()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")

	// Non-absolute
	s = &Skill{Entrypoint: "relative/path"}
	err = s.ValidateEntrypointExists()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "absolute")

	// Absolute but missing
	s = &Skill{Entrypoint: "/nonexistent/file.md"}
	err = s.ValidateEntrypointExists()
	require.Error(t, err)

	// Absolute and exists
	tmpFile := filepath.Join(t.TempDir(), "SKILL.md")
	require.NoError(t, os.WriteFile(tmpFile, []byte("skill"), 0644))
	s = &Skill{Entrypoint: tmpFile}
	assert.NoError(t, s.ValidateEntrypointExists())
}

// INVARIANT: GetAllowedTools returns the standard tool set.

func TestGetAllowedTools(t *testing.T) {
	tools := GetAllowedTools()
	assert.Contains(t, tools, "read")
	assert.Contains(t, tools, "write")
	assert.Contains(t, tools, "bash")
	assert.Greater(t, len(tools), 5)
}

// INVARIANT: ParseSemVer / CompareSemVer parse and order versions.

func TestParseSemVer(t *testing.T) {
	maj, min, patch, err := ParseSemVer("1.2.3")
	require.NoError(t, err)
	assert.Equal(t, 1, maj)
	assert.Equal(t, 2, min)
	assert.Equal(t, 3, patch)

	_, _, _, err = ParseSemVer("bad")
	assert.Error(t, err)
}

func TestCompareSemVer(t *testing.T) {
	cmp, err := CompareSemVer("1.0.0", "2.0.0")
	require.NoError(t, err)
	assert.Equal(t, -1, cmp)

	cmp, err = CompareSemVer("2.0.0", "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, 1, cmp)

	cmp, err = CompareSemVer("1.0.0", "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, 0, cmp)
}

// INVARIANT: error messages render.

func TestSkillErrors(t *testing.T) {
	assert.Contains(t, (&InvalidSemVerError{Version: "x"}).Error(), "x")
	assert.Contains(t, (&DuplicateSkillError{Name: "n"}).Error(), "n")
	assert.Contains(t, (&EmptyPermissionsError{Name: "n"}).Error(), "n")
	assert.Contains(t, (&UnauthorizedToolError{Skill: "s", Tool: "t"}).Error(), "t")
	assert.Contains(t, (&InvalidEntrypointError{Path: "p"}).Error(), "p")
	assert.Contains(t, (&SkillNotFoundError{ID: "x"}).Error(), "x")
}

// INVARIANT: ScopeFromSource default fallback.

func TestLoader_scopeFromSource(t *testing.T) {
	assert.Equal(t, SkillScopeFolder, scopeFromSource("unknown"))
}

// INVARIANT: Activate/Deactivate on missing skill error.

func TestLoader_ActivateDeactivateMissing(t *testing.T) {
	l := NewLoaderWithDirs(t.TempDir(), t.TempDir(), t.TempDir())
	require.NoError(t, l.Load())
	assert.Error(t, l.Activate("nope"))
	assert.Error(t, l.Deactivate("nope"))
}

// INVARIANT: Get returns false for missing skill.

func TestLoader_GetMissing(t *testing.T) {
	l := NewLoaderWithDirs(t.TempDir(), t.TempDir(), t.TempDir())
	require.NoError(t, l.Load())
	_, ok := l.Get("nonexistent")
	assert.False(t, ok)
}

// INVARIANT: IncrementVersion with bad version errors.

func TestIncrementVersion_BadVersion(t *testing.T) {
	s := &Skill{Version: "notsemver"}
	l := NewMemorySkillStore()
	err := s.IncrementVersion(l, "user")
	assert.Error(t, err)
}
