package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MEMORY ENTRY INVARIANT TESTS
// ============================================================================

// INVARIANT 1: Scope determines storage location
func TestMemoryEntry_ScopeStorageLocation(t *testing.T) {
	t.Run("session scope should store in-memory", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeSession,
			Kind:    MemoryKindFact,
			Content: "test fact",
		}

		storage, err := entry.GetStorageLocation()
		require.NoError(t, err)
		assert.Equal(t, "in-memory", storage, "Session scope should use in-memory storage")
	})

	t.Run("folder scope should store in .forge/memory/", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:        "mem-1",
			Scope:     MemoryScopeFolder,
			Kind:      MemoryKindFact,
			Content:   "test fact",
			FolderID:  "folder-1",
		}

		storage, err := entry.GetStorageLocation()
		require.NoError(t, err)
		assert.Contains(t, storage, ".forge/memory/", "Folder scope should use .forge/memory/ storage")
		assert.Contains(t, storage, "folder-1", "Folder storage should include folder ID")
	})

	t.Run("global scope should store in ~/.forge/global.db", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeGlobal,
			Kind:    MemoryKindPreference,
			Content: "test preference",
		}

		storage, err := entry.GetStorageLocation()
		require.NoError(t, err)
		assert.Contains(t, storage, ".forge/global.db", "Global scope should use global.db storage")
	})

	t.Run("should fail validation for invalid scope", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScope("invalid"), // INVALID
			Kind:    MemoryKindFact,
			Content: "test fact",
		}

		_, err := entry.GetStorageLocation()
		assert.Error(t, err, "Invalid scope should return error")
	})
}

// INVARIANT 2: Source=user-explicit entries are auto-Status=active; Source=model-inferred require Status=pending
func TestMemoryEntry_StatusBySource(t *testing.T) {
	t.Run("user-explicit source should auto-set status to active", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeFolder,
			Kind:    MemoryKindInstruction,
			Source:  MemorySourceUserExplicit,
			Content: "always use table-driven tests",
		}

		// Auto-set status on creation
		entry.SetInitialStatus()

		assert.Equal(t, MemoryStatusActive, entry.Status,
			"User-explicit entries should be auto-Status=active")
	})

	t.Run("repo-derived source should auto-set status to active", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeFolder,
			Kind:    MemoryKindFact,
			Source:  MemorySourceRepoDerived,
			Content: "project uses sqlc, not ORM",
		}

		entry.SetInitialStatus()

		assert.Equal(t, MemoryStatusActive, entry.Status,
			"Repo-derived entries should be auto-Status=active")
	})

	t.Run("tool-observed source should auto-set status to active", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeFolder,
			Kind:    MemoryKindWorkflow,
			Source:  MemorySourceToolObserved,
			Content: "tests run with make test-integration",
		}

		entry.SetInitialStatus()

		assert.Equal(t, MemoryStatusActive, entry.Status,
			"Tool-observed entries should be auto-Status=active")
	})

	t.Run("skill-derived source should auto-set status to active", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeFolder,
			Kind:    MemoryKindDecision,
			Source:  MemorySourceSkillDerived,
			Content: "use go test -race",
		}

		entry.SetInitialStatus()

		assert.Equal(t, MemoryStatusActive, entry.Status,
			"Skill-derived entries should be auto-Status=active")
	})

	t.Run("model-inferred source should require Status=pending", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeFolder,
			Kind:    MemoryKindFact,
			Source:  MemorySourceModelInferred,
			Content: "likely uses dependency injection",
		}

		entry.SetInitialStatus()

		assert.Equal(t, MemoryStatusPending, entry.Status,
			"Model-inferred entries should require Status=pending")
	})

	t.Run("should fail when model-inferred has Status=active without approval", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:      "mem-1",
			Scope:   MemoryScopeFolder,
			Kind:    MemoryKindFact,
			Source:  MemorySourceModelInferred,
			Content: "inferred fact",
			Status:  MemoryStatusActive, // INVALID - NOT APPROVED
		}

		valid := entry.ValidateStatusBySource()
		assert.False(t, valid, "Model-inferred with Status=active without approval should fail")
	})

	t.Run("should pass when model-inferred is approved to active", func(t *testing.T) {
		entry := &MemoryEntry{
			ID:          "mem-1",
			Scope:       MemoryScopeFolder,
			Kind:        MemoryKindFact,
			Source:      MemorySourceModelInferred,
			Content:     "inferred fact",
			Status:      MemoryStatusActive,
			ApprovedAt:  timePtr(time.Now()),
			ApprovedBy:  "user",
		}

		valid := entry.ValidateStatusBySource()
		assert.True(t, valid, "Model-inferred with approval should pass")
	})
}

// INVARIANT 3: Confidence is monotonic decreasing over time
func TestMemoryEntry_ConfidenceMonotonicDecreasing(t *testing.T) {
	t.Run("should fail when confidence increases", func(t *testing.T) {
		now := time.Now()
		entry := &MemoryEntry{
			ID:         "mem-1",
			Scope:      MemoryScopeFolder,
			Kind:       MemoryKindFact,
			Content:    "test fact",
			Confidence: 0.9,
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now.Add(-1 * time.Hour),
		}

		// Try to increase confidence
		previousConfidence := entry.Confidence
		entry.Confidence = 0.95 // INCREASED
		entry.UpdatedAt = now

		valid := entry.ValidateConfidenceNotIncreased(previousConfidence)
		assert.False(t, valid, "Increasing confidence should fail validation")
	})

	t.Run("should pass when confidence stays same", func(t *testing.T) {
		now := time.Now()
		entry := &MemoryEntry{
			ID:         "mem-1",
			Scope:      MemoryScopeFolder,
			Kind:       MemoryKindFact,
			Content:    "test fact",
			Confidence: 0.9,
			CreatedAt:  now.Add(-1 * time.Hour),
			UpdatedAt:  now,
		}

		previousConfidence := entry.Confidence
		// Confidence stays same, only time updates
		entry.UpdatedAt = time.Now()

		valid := entry.ValidateConfidenceNotIncreased(previousConfidence)
		assert.True(t, valid, "Same confidence should pass validation")
	})

	t.Run("should pass when confidence decreases", func(t *testing.T) {
		now := time.Now()
		entry := &MemoryEntry{
			ID:         "mem-1",
			Scope:      MemoryScopeFolder,
			Kind:       MemoryKindFact,
			Content:    "test fact",
			Confidence: 0.9,
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now.Add(-1 * time.Hour),
		}

		previousConfidence := entry.Confidence
		entry.Confidence = 0.85 // DECREASED
		entry.UpdatedAt = now

		valid := entry.ValidateConfidenceNotIncreased(previousConfidence)
		assert.True(t, valid, "Decreasing confidence should pass validation")
	})

	t.Run("should apply time decay function correctly", func(t *testing.T) {
		now := time.Now()
		entry := &MemoryEntry{
			ID:         "mem-1",
			Scope:      MemoryScopeFolder,
			Kind:       MemoryKindFact,
			Content:    "test fact",
			Confidence: 1.0,
			CreatedAt:  now.Add(-24 * time.Hour), // 1 day ago
			UpdatedAt:  now,
		}

		// Apply time decay
		previousConfidence := entry.Confidence
		decayConfidence := entry.ApplyTimeDecay(now.Add(24 * time.Hour)) // Another day passes

		assert.True(t, decayConfidence < previousConfidence,
			"Time decay should decrease confidence")
	})

	t.Run("should fail when confidence is out of range [0, 1]", func(t *testing.T) {
		testCases := []struct {
			confidence float64
			valid      bool
		}{
			{-0.1, false},
			{0.0, true},
			{0.5, true},
			{1.0, true},
			{1.1, false},
		}

		for _, tc := range testCases {
			entry := &MemoryEntry{
				ID:         "mem-1",
				Scope:      MemoryScopeFolder,
				Kind:       MemoryKindFact,
				Content:    "test fact",
				Confidence: tc.confidence,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			valid := entry.ValidateConfidenceRange()
			assert.Equal(t, tc.valid, valid,
				"Confidence %.2f should be %v", tc.confidence, tc.valid)
		}
	})
}

// INVARIANT 4: No two entries with same Scope+Kind+semantic_hash can both be Status=active (deduplication)
func TestMemoryEntry_NoDuplicateActiveEntries(t *testing.T) {
	t.Run("should fail when two active entries have same semantic hash", func(t *testing.T) {
		content := "always use table-driven Go tests"
		hash := computeSemanticHash(content)

		entry1 := &MemoryEntry{
			ID:         "mem-1",
			Scope:      MemoryScopeGlobal,
			Kind:       MemoryKindPreference,
			Content:    content,
			SemanticHash: hash,
			Status:     MemoryStatusActive,
			Confidence: 1.0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		entry2 := &MemoryEntry{
			ID:          "mem-2",
			Scope:       MemoryScopeGlobal,
			Kind:        MemoryKindPreference,
			Content:     content, // SAME CONTENT
			SemanticHash: hash,
			Status:      MemoryStatusActive, // BOTH ACTIVE - DUPLICATE
			Confidence:  0.9,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		store := NewMemoryEntryStore()
		store.Save(entry1, entry2)

		valid, err := ValidateNoDuplicateActiveEntries(store)
		require.NoError(t, err)
		assert.False(t, valid, "Duplicate active entries should fail validation")
	})

	t.Run("should pass when entries have same hash but different scope", func(t *testing.T) {
		content := "use slog for logging"
		hash := computeSemanticHash(content)

		entry1 := &MemoryEntry{
			ID:           "mem-1",
			Scope:        MemoryScopeGlobal,
			Kind:         MemoryKindPreference,
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		entry2 := &MemoryEntry{
			ID:           "mem-2",
			Scope:        MemoryScopeFolder, // DIFFERENT SCOPE
			Kind:         MemoryKindPreference,
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		store := NewMemoryEntryStore()
		store.Save(entry1, entry2)

		valid, err := ValidateNoDuplicateActiveEntries(store)
		require.NoError(t, err)
		assert.True(t, valid, "Same hash in different scopes should pass")
	})

	t.Run("should pass when entries have same hash but different kind", func(t *testing.T) {
		content := "run tests with make test"
		hash := computeSemanticHash(content)

		entry1 := &MemoryEntry{
			ID:           "mem-1",
			Scope:        MemoryScopeFolder,
			Kind:         MemoryKindWorkflow, // DIFFERENT KIND
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		entry2 := &MemoryEntry{
			ID:           "mem-2",
			Scope:        MemoryScopeFolder,
			Kind:         MemoryKindFact, // DIFFERENT KIND
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		store := NewMemoryEntryStore()
		store.Save(entry1, entry2)

		valid, err := ValidateNoDuplicateActiveEntries(store)
		require.NoError(t, err)
		assert.True(t, valid, "Same hash with different kind should pass")
	})

	t.Run("should pass when one entry is rejected", func(t *testing.T) {
		content := "prefer interfaces over structs"
		hash := computeSemanticHash(content)

		entry1 := &MemoryEntry{
			ID:           "mem-1",
			Scope:        MemoryScopeGlobal,
			Kind:         MemoryKindPreference,
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		entry2 := &MemoryEntry{
			ID:           "mem-2",
			Scope:        MemoryScopeGlobal,
			Kind:         MemoryKindPreference,
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusRejected, // NOT ACTIVE
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		store := NewMemoryEntryStore()
		store.Save(entry1, entry2)

		valid, err := ValidateNoDuplicateActiveEntries(store)
		require.NoError(t, err)
		assert.True(t, valid, "One active + one rejected should pass")
	})

	t.Run("should pass when one entry is pending", func(t *testing.T) {
		content := "use context for cancellation"
		hash := computeSemanticHash(content)

		entry1 := &MemoryEntry{
			ID:           "mem-1",
			Scope:        MemoryScopeGlobal,
			Kind:         MemoryKindPreference,
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		entry2 := &MemoryEntry{
			ID:           "mem-2",
			Scope:        MemoryScopeGlobal,
			Kind:         MemoryKindPreference,
			Content:      content,
			SemanticHash: hash,
			Status:       MemoryStatusPending, // NOT ACTIVE
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		store := NewMemoryEntryStore()
		store.Save(entry1, entry2)

		valid, err := ValidateNoDuplicateActiveEntries(store)
		require.NoError(t, err)
		assert.True(t, valid, "One active + one pending should pass")
	})

	t.Run("should fail when three active entries have same hash", func(t *testing.T) {
		content := "never ignore errors"
		hash := computeSemanticHash(content)

		entries := []*MemoryEntry{
			{ID: "mem-1", Scope: MemoryScopeGlobal, Kind: MemoryKindPreference,
				Content: content, SemanticHash: hash, Status: MemoryStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "mem-2", Scope: MemoryScopeGlobal, Kind: MemoryKindPreference,
				Content: content, SemanticHash: hash, Status: MemoryStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "mem-3", Scope: MemoryScopeGlobal, Kind: MemoryKindPreference,
				Content: content, SemanticHash: hash, Status: MemoryStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		store := NewMemoryEntryStore()
		store.Save(entries...)

		valid, err := ValidateNoDuplicateActiveEntries(store)
		require.NoError(t, err)
		assert.False(t, valid, "Three duplicate active entries should fail validation")
	})

	t.Run("should compute semantic hash deterministically", func(t *testing.T) {
		content := "use table-driven tests"
		hash1 := computeSemanticHash(content)
		hash2 := computeSemanticHash(content)
		hash3 := computeSemanticHash(content)

		assert.Equal(t, hash1, hash2, "Same content should produce same hash")
		assert.Equal(t, hash2, hash3, "Hash should be deterministic")
	})

	t.Run("should produce different hashes for different content", func(t *testing.T) {
		hash1 := computeSemanticHash("use table-driven tests")
		hash2 := computeSemanticHash("use unit tests")

		assert.NotEqual(t, hash1, hash2, "Different content should produce different hashes")
	})
}