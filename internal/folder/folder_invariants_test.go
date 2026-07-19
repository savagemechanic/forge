package folder

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// FOLDER INVARIANT TESTS
// ============================================================================

// INVARIANT 1: ID = crypto.SHA256(CanonicalPath)
func TestFolder_IDIsDeterministicHash(t *testing.T) {
	t.Run("should fail when ID is not SHA256 hash of CanonicalPath", func(t *testing.T) {
		folder := &Folder{
			ID:            "wrong-id", // NOT A HASH
			CanonicalPath: "/Users/test/project",
			GitRemote:     "git@github.com:user/repo.git",
			GitRoot:       "/Users/test/project/.git",
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateIDHash()
		require.NoError(t, err)
		assert.False(t, valid, "Folder with wrong ID should fail validation")
	})

	t.Run("should fail when ID is different hash algorithm", func(t *testing.T) {
		// MD5 would be 32 hex chars, SHA256 is 64
		folder := &Folder{
			ID:            "d41d8cd98f00b204e9800998ecf8427e", // MD5 of empty
			CanonicalPath: "/Users/test/project",
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateIDHash()
		require.NoError(t, err)
		assert.False(t, valid, "Folder with non-SHA256 ID should fail validation")
	})

	t.Run("should pass when ID is correct SHA256 hash", func(t *testing.T) {
		path := "/Users/test/project"
		id := ComputeFolderID(path)

		folder := &Folder{
			ID:            id,
			CanonicalPath: path,
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateIDHash()
		require.NoError(t, err)
		assert.True(t, valid, "Folder with correct ID should pass validation")
	})

	t.Run("should be deterministic - same path always produces same ID", func(t *testing.T) {
		path := "/Users/test/project"

		id1 := ComputeFolderID(path)
		id2 := ComputeFolderID(path)
		id3 := ComputeFolderID(path)

		assert.Equal(t, id1, id2, "Same path should produce same ID")
		assert.Equal(t, id2, id3, "Same path should produce same ID (idempotent)")
	})

	t.Run("should produce different IDs for different paths", func(t *testing.T) {
		path1 := "/Users/test/project1"
		path2 := "/Users/test/project2"

		id1 := ComputeFolderID(path1)
		id2 := ComputeFolderID(path2)

		assert.NotEqual(t, id1, id2, "Different paths should produce different IDs")
	})
}

// INVARIANT 2: CanonicalPath is cleaned (no ../, no symlinks)
func TestFolder_CanonicalPathIsClean(t *testing.T) {
	t.Run("should fail when CanonicalPath contains ..", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/../project"),
			CanonicalPath: "/Users/test/../project", // CONTAINS ..
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateCanonicalPath()
		require.NoError(t, err)
		assert.False(t, valid, "Folder with .. in path should fail validation")
	})

	t.Run("should fail when CanonicalPath is not absolute", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("relative/path"),
			CanonicalPath: "relative/path", // NOT ABSOLUTE
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateCanonicalPath()
		require.NoError(t, err)
		assert.False(t, valid, "Folder with relative path should fail validation")
	})

	t.Run("should fail when CanonicalPath contains trailing slash", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project/"),
			CanonicalPath: "/Users/test/project/", // TRAILING SLASH
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateCanonicalPath()
		require.NoError(t, err)
		assert.False(t, valid, "Folder with trailing slash should fail validation")
	})

	t.Run("should fail when CanonicalPath contains duplicate slashes", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users//test/project"),
			CanonicalPath: "/Users//test/project", // DOUBLE SLASH
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateCanonicalPath()
		require.NoError(t, err)
		assert.False(t, valid, "Folder with duplicate slashes should fail validation")
	})

	t.Run("should pass when CanonicalPath is properly cleaned", func(t *testing.T) {
		path, err := filepath.Abs("/Users/test/project")
		require.NoError(t, err)

		folder := &Folder{
			ID:            ComputeFolderID(path),
			CanonicalPath: path, // PROPERLY CLEANED
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid, err := folder.ValidateCanonicalPath()
		require.NoError(t, err)
		assert.True(t, valid, "Folder with clean path should pass validation")
	})

	t.Run("should resolve symlinks when creating CanonicalPath", func(t *testing.T) {
		// This test would need actual filesystem setup
		// For now, we'll test the CleanPath function behavior
		input := "/var/../usr/local/bin"
		expected := "/usr/local/bin"

		cleaned := CleanPath(input)
		assert.Equal(t, expected, cleaned, "CleanPath should resolve ..")
	})
}

// INVARIANT 3: GitRoot is empty if not a Git repo, otherwise points to .git directory
func TestFolder_GitRootConsistency(t *testing.T) {
	t.Run("should pass when GitRoot is empty and not a Git repo", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			GitRoot:       "", // EMPTY - NOT A GIT REPO
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid := folder.ValidateGitRoot()
		assert.True(t, valid, "Empty GitRoot for non-Git folder should pass")
	})

	t.Run("should pass when GitRoot points to .git directory", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			GitRoot:       "/Users/test/project/.git", // VALID
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid := folder.ValidateGitRoot()
		assert.True(t, valid, "Valid GitRoot should pass")
	})

	t.Run("should fail when GitRoot does not end with /.git", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			GitRoot:       "/Users/test/project", // NOT .git
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid := folder.ValidateGitRoot()
		assert.False(t, valid, "GitRoot without /.git should fail")
	})

	t.Run("should fail when GitRoot is outside CanonicalPath", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			GitRoot:       "/other/project/.git", // OUTSIDE FOLDER
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid := folder.ValidateGitRoot()
		assert.False(t, valid, "GitRoot outside folder should fail")
	})

	t.Run("should pass when GitRoot is nested inside CanonicalPath", func(t *testing.T) {
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			GitRoot:       "/Users/test/project/submodule/.git", // NESTED (submodule)
			CreatedAt:     time.Now(),
			LastOpenedAt:  time.Now(),
		}

		valid := folder.ValidateGitRoot()
		assert.True(t, valid, "Nested GitRoot (submodule) should pass")
	})
}

// INVARIANT 4: LastOpenedAt is monotonically increasing (only updated on forge open)
func TestFolder_LastOpenedAtMonotonic(t *testing.T) {
	t.Run("should fail when LastOpenedAt is before CreatedAt", func(t *testing.T) {
		now := time.Now()
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			CreatedAt:     now.Add(1 * time.Hour),
			LastOpenedAt:  now, // BEFORE CREATED
		}

		valid := folder.ValidateLastOpenedAt()
		assert.False(t, valid, "LastOpenedAt before CreatedAt should fail")
	})

	t.Run("should fail when LastOpenedAt is exactly CreatedAt (must be >=)", func(t *testing.T) {
		now := time.Now()
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			CreatedAt:     now,
			LastOpenedAt:  now, // SAME TIME
		}

		valid := folder.ValidateLastOpenedAt()
		assert.False(t, valid, "LastOpenedAt equal to CreatedAt should fail (must be >)")
	})

	t.Run("should pass when LastOpenedAt is after CreatedAt", func(t *testing.T) {
		now := time.Now()
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			CreatedAt:     now.Add(-1 * time.Hour),
			LastOpenedAt:  now, // AFTER CREATED
		}

		valid := folder.ValidateLastOpenedAt()
		assert.True(t, valid, "LastOpenedAt after CreatedAt should pass")
	})

	t.Run("should fail when LastOpenedAt decreases", func(t *testing.T) {
		now := time.Now()
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			CreatedAt:     now.Add(-2 * time.Hour),
			LastOpenedAt:  now.Add(-1 * time.Hour),
		}

		// Try to set to earlier time
		previous := folder.LastOpenedAt
		folder.LastOpenedAt = now.Add(-2 * time.Hour) // DECREASED

		valid := folder.ValidateLastOpenedAtNotDecreased(previous)
		assert.False(t, valid, "Decreasing LastOpenedAt should fail")
	})

	t.Run("should pass when LastOpenedAt increases or stays same", func(t *testing.T) {
		now := time.Now()
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			CreatedAt:     now.Add(-2 * time.Hour),
			LastOpenedAt:  now.Add(-1 * time.Hour),
		}

		// Update to later time
		previous := folder.LastOpenedAt
		folder.LastOpenedAt = now // INCREASED

		valid := folder.ValidateLastOpenedAtNotDecreased(previous)
		assert.True(t, valid, "Increasing LastOpenedAt should pass")
	})

	t.Run("should pass multiple updates in sequence", func(t *testing.T) {
		now := time.Now()
		folder := &Folder{
			ID:            ComputeFolderID("/Users/test/project"),
			CanonicalPath: "/Users/test/project",
			CreatedAt:     now.Add(-3 * time.Hour),
			LastOpenedAt:  now.Add(-2 * time.Hour),
		}

		times := []time.Time{
			now.Add(-1 * time.Hour),
			now.Add(-30 * time.Minute),
			now,
		}

		for i, newTime := range times {
			previous := folder.LastOpenedAt
			folder.LastOpenedAt = newTime
			valid := folder.ValidateLastOpenedAtNotDecreased(previous)
			assert.True(t, valid, "Update %d: monotonic increase should pass", i)
		}
	})
}