package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadsBuiltinSkills(t *testing.T) {
	// Use the actual forge skills directory
	loader := NewLoader("../..", ".")
	err := loader.Load()
	require.NoError(t, err)

	skills := loader.All()
	assert.NotEmpty(t, skills, "should load built-in skills")
}

func TestLoader_NerdSkillExists(t *testing.T) {
	loader := NewLoader("../..", ".")
	require.NoError(t, loader.Load())

	s, ok := loader.Get("nerd")
	require.True(t, ok, "nerd skill should be loaded")
	assert.NotEmpty(t, s.Description)
	assert.NotEmpty(t, s.Body)
	assert.Equal(t, "builtin", s.Source)
}

func TestLoader_ParsesDescriptionFromFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "skills", "test-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: test-skill
description: A test skill for testing.
version: 1.0.0
---

# Test Skill

Body content here.
`), 0644))

	loader := NewLoaderWithDirs(filepath.Join(tmpDir, "skills"), t.TempDir(), t.TempDir())
	require.NoError(t, loader.Load())

	s, ok := loader.Get("test-skill")
	require.True(t, ok)
	assert.Equal(t, "A test skill for testing.", s.Description)
}

func TestLoader_ActiveInactive(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "skills", "active-one")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: active-one\ndescription: Active.\n---\n# Body"), 0644))

	loader := NewLoaderWithDirs(filepath.Join(tmpDir, "skills"), t.TempDir(), t.TempDir())
	require.NoError(t, loader.Load())

	assert.Len(t, loader.Active(), 1)
	require.NoError(t, loader.Deactivate("active-one"))
	assert.Len(t, loader.Active(), 0)
	require.NoError(t, loader.Activate("active-one"))
	assert.Len(t, loader.Active(), 1)
}

func TestLoader_Install(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoaderWithDirs(t.TempDir(), filepath.Join(tmpDir, "skills"), t.TempDir())
	require.NoError(t, loader.Load())

	require.NoError(t, loader.Install("my-skill", "global", "---\nname: my-skill\ndescription: Installed.\n---\n# Body", true))

	// Reload from the install dir
	loader2 := NewLoaderWithDirs(t.TempDir(), filepath.Join(tmpDir, "skills"), t.TempDir())
	require.NoError(t, loader2.Load())
	s, ok := loader2.Get("my-skill")
	require.True(t, ok)
	assert.Equal(t, "Installed.", s.Description)
}

func TestLoader_List(t *testing.T) {
	loader := NewLoader("../..", ".")
	require.NoError(t, loader.Load())
	list := loader.List()
	assert.NotEmpty(t, list)
	assert.Contains(t, list, "nerd")
}
