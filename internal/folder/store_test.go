package folder

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: FolderStore round-trips; Get errors on missing.

func TestFolderStore_SaveGet(t *testing.T) {
	s := NewFolderStore()
	now := time.Now()
	f := &Folder{ID: "f1", CanonicalPath: "/x", CreatedAt: now, LastOpenedAt: now}
	s.Save(f)

	got, err := s.Get("f1")
	require.NoError(t, err)
	assert.Equal(t, "/x", got.CanonicalPath)
}

func TestFolderStore_GetMissing(t *testing.T) {
	s := NewFolderStore()
	_, err := s.Get("nope")
	assert.Error(t, err)
}

// INVARIANT: ValidateIDIsDeterministicHash — ID is SHA256(canonical path).

func TestFolder_ValidateIDIsDeterministicHash(t *testing.T) {
	f := &Folder{CanonicalPath: "/home/project"}
	expected := generateFolderID("/home/project")
	f.ID = expected
	assert.True(t, f.ValidateIDIsDeterministicHash())

	f.ID = "wrong"
	assert.False(t, f.ValidateIDIsDeterministicHash())
}

// INVARIANT: ValidateCanonicalPathIsClean — no ".." segments, absolute.

func TestFolder_ValidateCanonicalPathIsClean(t *testing.T) {
	f := &Folder{CanonicalPath: "/clean/path"}
	assert.NoError(t, f.ValidateCanonicalPathIsClean())

	f.CanonicalPath = "/bad/../escape"
	assert.Error(t, f.ValidateCanonicalPathIsClean())

	f.CanonicalPath = "relative"
	assert.Error(t, f.ValidateCanonicalPathIsClean())

	f.CanonicalPath = "/trailing/"
	assert.Error(t, f.ValidateCanonicalPathIsClean())
}

// INVARIANT: UpdateLastOpenedAt enforces monotonic increase.

func TestFolder_UpdateLastOpenedAt(t *testing.T) {
	now := time.Now()
	f := &Folder{CreatedAt: now, LastOpenedAt: now}

	err := f.UpdateLastOpenedAt()
	assert.NoError(t, err)
	assert.True(t, f.LastOpenedAt.After(now) || f.LastOpenedAt.Equal(now))
}

// INVARIANT: ValidateLastOpenedAtMonotonic.

func TestFolder_ValidateLastOpenedAtMonotonic(t *testing.T) {
	now := time.Now()
	f := &Folder{CreatedAt: now, LastOpenedAt: now.Add(time.Hour)}
	assert.True(t, f.ValidateLastOpenedAtMonotonic(nil))

	f.LastOpenedAt = now.Add(-time.Hour)
	assert.False(t, f.ValidateLastOpenedAtMonotonic(nil))
}

// INVARIANT: ValidateLastOpenedAtNotDecreased.

func TestFolder_ValidateLastOpenedAtNotDecreased(t *testing.T) {
	prev := time.Now()
	f := &Folder{LastOpenedAt: prev.Add(time.Hour)}
	assert.True(t, f.ValidateLastOpenedAtNotDecreased(prev))

	f.LastOpenedAt = prev.Add(-time.Hour)
	assert.False(t, f.ValidateLastOpenedAtNotDecreased(prev))
}

// INVARIANT: ResolveToCanonicalPath resolves absolute paths (real dirs).
// On macOS, /var resolves to /private/var — so compare via EvalSymlinks.

func TestFolder_ResolveToCanonicalPath(t *testing.T) {
	root := t.TempDir()
	f := &Folder{CanonicalPath: root}
	otherDir := t.TempDir()
	p, err := f.ResolveToCanonicalPath(otherDir)
	require.NoError(t, err)
	// Both should resolve to the same canonical form
	expected, _ := filepath.EvalSymlinks(otherDir)
	assert.Equal(t, expected, p)
}

// INVARIANT: CleanPath removes trailing slash (package function).

func TestCleanPath(t *testing.T) {
	assert.Equal(t, "/a/b", CleanPath("/a/b/"))
	assert.Equal(t, "/a/b", CleanPath("/a/b"))
}

// INVARIANT: Summary handles unknown project (the 80% line).

func TestProjectInfo_Summary_Unknown(t *testing.T) {
	info := &ProjectInfo{RootPath: "/x", Types: []ProjectType{ProjectTypeUnknown}}
	s := info.Summary()
	assert.Equal(t, "unknown project", s)
}
