package skill

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// SKILL INVARIANT TESTS
// ============================================================================

// INVARIANT 1: Name+Scope is unique (no two skills with same name in same scope)
func TestSkill_NameScopeUnique(t *testing.T) {
	t.Run("should fail when two skills have same name in same scope", func(t *testing.T) {
		skill1 := &Skill{
			ID:          "skill-1",
			Name:        "create-skill", // SAME NAME
			Scope:       SkillScopeGlobal, // SAME SCOPE
			Description: "Creates new skills",
			Version:     "1.0.0",
		}
		skill2 := &Skill{
			ID:          "skill-2",
			Name:        "create-skill", // SAME NAME
			Scope:       SkillScopeGlobal, // SAME SCOPE
			Description: "Another skill creator",
			Version:     "2.0.0",
		}

		store := NewMemorySkillStore()
		store.Save(skill1, skill2)

		valid, err := ValidateNameScopeUnique(store)
		require.NoError(t, err)
		assert.False(t, valid, "Duplicate name+scope should fail validation")
	})

	t.Run("should pass when skills have same name but different scope", func(t *testing.T) {
		skill1 := &Skill{
			ID:          "skill-1",
			Name:        "create-skill",
			Scope:       SkillScopeGlobal, // DIFFERENT SCOPE
			Description: "Global skill creator",
			Version:     "1.0.0",
		}
		skill2 := &Skill{
			ID:          "skill-2",
			Name:        "create-skill", // SAME NAME
			Scope:       SkillScopeFolder, // DIFFERENT SCOPE
			Description: "Folder skill creator",
			Version:     "1.0.0",
		}

		store := NewMemorySkillStore()
		store.Save(skill1, skill2)

		valid, err := ValidateNameScopeUnique(store)
		require.NoError(t, err)
		assert.True(t, valid, "Same name in different scopes should pass")
	})

	t.Run("should pass when skills have different names in same scope", func(t *testing.T) {
		skill1 := &Skill{
			ID:          "skill-1",
			Name:        "create-skill",
			Scope:       SkillScopeGlobal,
			Description: "Creates skills",
			Version:     "1.0.0",
		}
		skill2 := &Skill{
			ID:          "skill-2",
			Name:        "modify-skill", // DIFFERENT NAME
			Scope:       SkillScopeGlobal, // SAME SCOPE
			Description: "Modifies skills",
			Version:     "1.0.0",
		}

		store := NewMemorySkillStore()
		store.Save(skill1, skill2)

		valid, err := ValidateNameScopeUnique(store)
		require.NoError(t, err)
		assert.True(t, valid, "Different names in same scope should pass")
	})

	t.Run("should fail when three skills have same name in same scope", func(t *testing.T) {
		skills := []*Skill{
			{ID: "skill-1", Name: "test-skill", Scope: SkillScopeFolder, Version: "1.0.0"},
			{ID: "skill-2", Name: "test-skill", Scope: SkillScopeFolder, Version: "1.0.0"},
			{ID: "skill-3", Name: "test-skill", Scope: SkillScopeFolder, Version: "1.0.0"},
		}

		store := NewMemorySkillStore()
		store.Save(skills...)

		valid, err := ValidateNameScopeUnique(store)
		require.NoError(t, err)
		assert.False(t, valid, "Three duplicate names should fail validation")
	})
}

// INVARIANT 2: Permissions is a whitelist; tool calls not listed are rejected
func TestSkill_PermissionsWhitelist(t *testing.T) {
	t.Run("should fail when calling tool not in permissions", func(t *testing.T) {
		skill := &Skill{
			ID:          "skill-1",
			Name:        "test-skill",
			Scope:       SkillScopeGlobal,
			Permissions: []Permission{
				{Tool: "filesystem.read"},
				{Tool: "filesystem.write"},
			},
		}

		// Try to call tool not in permissions
		allowed := skill.IsToolAllowed("git.commit") // NOT IN PERMISSIONS
		assert.False(t, allowed, "Tool not in permissions should be rejected")
	})

	t.Run("should pass when calling tool in permissions", func(t *testing.T) {
		skill := &Skill{
			ID:          "skill-1",
			Name:        "test-skill",
			Scope:       SkillScopeGlobal,
			Permissions: []Permission{
				{Tool: "filesystem.read"},
				{Tool: "filesystem.write"},
				{Tool: "go.test"},
			},
		}

		allowed := skill.IsToolAllowed("go.test") // IN PERMISSIONS
		assert.True(t, allowed, "Tool in permissions should be allowed")
	})

	t.Run("should fail when calling with partial tool match", func(t *testing.T) {
		skill := &Skill{
			ID:          "skill-1",
			Name:        "test-skill",
			Scope:       SkillScopeGlobal,
			Permissions: []Permission{
				{Tool: "filesystem.read"},
			},
		}

		// Partial match should not be enough
		allowed := skill.IsToolAllowed("filesystem") // NOT EXACT MATCH
		assert.False(t, allowed, "Partial tool match should be rejected")
	})

	t.Run("should respect resource constraints in permissions", func(t *testing.T) {
		skill := &Skill{
			ID:    "skill-1",
			Name:  "test-skill",
			Scope: SkillScopeGlobal,
			Permissions: []Permission{
				{
					Tool:           "filesystem.write",
					AllowedPaths:   []string{"/tmp", "/safe"},
					DeniedPaths:    []string{"/etc", "/root"},
					MaxFileSize:    1024 * 1024, // 1MB
				},
			},
		}

		// Test allowed path
		allowed := skill.IsToolAllowedWithPath("filesystem.write", "/safe/file.txt")
		assert.True(t, allowed, "Allowed path should pass")

		// Test denied path
		allowed = skill.IsToolAllowedWithPath("filesystem.write", "/etc/passwd")
		assert.False(t, allowed, "Denied path should be rejected")
	})

	t.Run("should fail when empty permissions list", func(t *testing.T) {
		skill := &Skill{
			ID:          "skill-1",
			Name:        "test-skill",
			Scope:       SkillScopeGlobal,
			Permissions: []Permission{}, // EMPTY
		}

		// No tools should be allowed
		allowed := skill.IsToolAllowed("filesystem.read")
		assert.False(t, allowed, "Empty permissions should reject all tools")
	})

	t.Run("should validate permission on execution", func(t *testing.T) {
		skill := &Skill{
			ID:    "skill-1",
			Name:  "test-skill",
			Scope: SkillScopeGlobal,
			Permissions: []Permission{
				{Tool: "go.test"},
			},
		}

		execution := ToolExecution{
			Tool: "filesystem.write", // NOT IN PERMISSIONS
			Args: map[string]string{"path": "/file.txt"},
		}

		err := skill.ValidateExecution(&execution)
		assert.Error(t, err, "Execution with unpermitted tool should fail")
	})

	t.Run("should pass valid execution", func(t *testing.T) {
		skill := &Skill{
			ID:    "skill-1",
			Name:  "test-skill",
			Scope: SkillScopeGlobal,
			Permissions: []Permission{
				{Tool: "go.test"},
			},
		}

		execution := ToolExecution{
			Tool: "go.test", // IN PERMISSIONS
			Args: map[string]string{"package": "./..."},
		}

		err := skill.ValidateExecution(&execution)
		assert.NoError(t, err, "Execution with permitted tool should pass")
	})
}

// INVARIANT 3: Version follows SemVer; skill cannot auto-increment its own version
func TestSkill_SemVerAndNoAutoIncrement(t *testing.T) {
	t.Run("should fail when version is not SemVer", func(t *testing.T) {
		invalidVersions := []string{
			"1",           // Missing minor and patch
			"1.2",         // Missing patch
			"v1.2.3",      // Has v prefix
			"1.2.3.4",     // Too many parts
			"1.x.3",       // Non-numeric
			"1.2.",        // Missing patch
			".2.3",        // Missing major
			"1.2.3-beta",  // Prerelease (for now, reject)
			"1.2.3+meta",  // Build metadata (for now, reject)
			"01.2.3",      // Leading zeros
			"1.02.3",      // Leading zeros
		}

		for _, version := range invalidVersions {
			skill := &Skill{
				ID:      "skill-1",
				Name:    "test-skill",
				Scope:   SkillScopeGlobal,
				Version: version,
			}

			valid := skill.ValidateSemVer()
			assert.False(t, valid, "Invalid version '%s' should fail", version)
		}
	})

	t.Run("should pass valid SemVer", func(t *testing.T) {
		validVersions := []string{
			"0.0.1",
			"0.1.0",
			"1.0.0",
			"1.2.3",
			"10.20.30",
			"255.255.255",
		}

		for _, version := range validVersions {
			skill := &Skill{
				ID:      "skill-1",
				Name:    "test-skill",
				Scope:   SkillScopeGlobal,
				Version: version,
			}

			valid := skill.ValidateSemVer()
			assert.True(t, valid, "Valid version '%s' should pass", version)
		}
	})

	t.Run("should fail when skill tries to increment its own version", func(t *testing.T) {
		skill := &Skill{
			ID:      "skill-1",
			Name:    "test-skill",
			Scope:   SkillScopeGlobal,
			Version: "1.0.0",
		}

		// Store original
		store := NewMemorySkillStore()
		store.Save(skill)

		// Try to auto-increment (should fail)
		originalVersion := skill.Version
		err := skill.IncrementVersion(store, "self") // SELF-INCREMENT
		assert.Error(t, err, "Self-increment should fail")
		assert.Equal(t, originalVersion, skill.Version, "Version should not change")
	})

	t.Run("should allow version increment by external authority", func(t *testing.T) {
		skill := &Skill{
			ID:      "skill-1",
			Name:    "test-skill",
			Scope:   SkillScopeGlobal,
			Version: "1.0.0",
		}

		store := NewMemorySkillStore()
		store.Save(skill)

		// Increment by user (should pass)
		err := skill.IncrementVersion(store, "user") // EXTERNAL AUTHORITY
		assert.NoError(t, err, "External increment should pass")
		assert.Equal(t, "1.0.1", skill.Version, "Version should increment patch")
	})

	t.Run("should track version history", func(t *testing.T) {
		skill := &Skill{
			ID:      "skill-1",
			Name:    "test-skill",
			Scope:   SkillScopeGlobal,
			Version: "1.0.0",
			VersionHistory: []VersionEntry{
				{Version: "1.0.0", At: time.Now().Add(-1 * time.Hour), By: "user"},
			},
		}

		store := NewMemorySkillStore()
		store.Save(skill)

		// Record version change
		now := time.Now()
		skill.RecordVersionChange("1.1.0", now, "user")

		assert.Len(t, skill.VersionHistory, 2, "Should have 2 version entries")
		assert.Equal(t, "1.1.0", skill.VersionHistory[1].Version)
		assert.Equal(t, "user", skill.VersionHistory[1].By)
	})

	t.Run("should enforce version ordering in history", func(t *testing.T) {
		now := time.Now()
		history := []VersionEntry{
			{Version: "1.0.0", At: now.Add(-3 * time.Hour), By: "user"},
			{Version: "1.1.0", At: now.Add(-2 * time.Hour), By: "user"},
			{Version: "1.2.0", At: now.Add(-1 * time.Hour), By: "user"},
		}

		skill := &Skill{
			ID:             "skill-1",
			Name:           "test-skill",
			Scope:          SkillScopeGlobal,
			Version:        "1.2.0",
			VersionHistory: history,
		}

		valid := skill.ValidateVersionHistory()
		assert.True(t, valid, "Ordered version history should pass")
	})

	t.Run("should fail when version history is out of order", func(t *testing.T) {
		now := time.Now()
		history := []VersionEntry{
			{Version: "1.0.0", At: now.Add(-3 * time.Hour), By: "user"},
			{Version: "1.2.0", At: now.Add(-2 * time.Hour), By: "user"}, // SKIP
			{Version: "1.1.0", At: now.Add(-1 * time.Hour), By: "user"}, // OUT OF ORDER
		}

		skill := &Skill{
			ID:             "skill-1",
			Name:           "test-skill",
			Scope:          SkillScopeGlobal,
			Version:        "1.1.0",
			VersionHistory: history,
		}

		valid := skill.ValidateVersionHistory()
		assert.False(t, valid, "Out of order version history should fail")
	})
}

// INVARIANT 4: Entrypoint must exist in skill directory
func TestSkill_EntrypointExists(t *testing.T) {
	t.Run("should fail when entrypoint file does not exist", func(t *testing.T) {
		skill := &Skill{
			ID:         "skill-1",
			Name:       "test-skill",
			Scope:      SkillScopeGlobal,
			Entrypoint: "SKILL.md",
			Directory:  "/nonexistent/skill", // NONEXISTENT
		}

		exists := skill.ValidateEntrypoint()
		assert.False(t, exists, "Nonexistent entrypoint should fail")
	})

	t.Run("should fail when entrypoint is empty", func(t *testing.T) {
		skill := &Skill{
			ID:         "skill-1",
			Name:       "test-skill",
			Scope:      SkillScopeGlobal,
			Entrypoint: "", // EMPTY
		}

		exists := skill.ValidateEntrypoint()
		assert.False(t, exists, "Empty entrypoint should fail")
	})

	t.Run("should validate entrypoint exists for built-in skills", func(t *testing.T) {
		skill := &Skill{
			ID:         "skill-1",
			Name:       "create-skill",
			Scope:      SkillScopeBuiltIn,
			Entrypoint: "SKILL.md",
			Directory:  "skills/create-skill",
		}

		// For built-in skills, check dist/ directory
		valid := skill.ValidateEntrypoint()
		// This will fail in stub phase
		assert.False(t, valid, "Built-in skill should validate entrypoint")
	})

	t.Run("should validate entrypoint exists for global skills", func(t *testing.T) {
		skill := &Skill{
			ID:         "skill-1",
			Name:       "test-skill",
			Scope:      SkillScopeGlobal,
			Entrypoint: "SKILL.md",
			Directory:  "~/.forge/skills/test-skill",
		}

		valid := skill.ValidateEntrypoint()
		// This will fail in stub phase
		assert.False(t, valid, "Global skill should validate entrypoint")
	})

	t.Run("should validate entrypoint exists for folder skills", func(t *testing.T) {
		skill := &Skill{
			ID:         "skill-1",
			Name:       "test-skill",
			Scope:      SkillScopeFolder,
			Entrypoint: "SKILL.md",
			Directory:  ".forge/skills/test-skill",
		}

		valid := skill.ValidateEntrypoint()
		// This will fail in stub phase
		assert.False(t, valid, "Folder skill should validate entrypoint")
	})
}