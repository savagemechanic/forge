package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: store Get/ListByScopeAndKind/filter the entries correctly.

func TestMemoryEntryStore_Get(t *testing.T) {
	s := NewMemoryEntryStore()
	e := &MemoryEntry{ID: "e1", Content: "test", Scope: MemoryScopeFolder, Kind: MemoryKindPreference}
	s.Save(e)

	got, err := s.Get("e1")
	require.NoError(t, err)
	assert.Equal(t, "test", got.Content)
}

func TestMemoryEntryStore_GetMissing(t *testing.T) {
	s := NewMemoryEntryStore()
	_, err := s.Get("nope")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryEntryStore_ListByScopeAndKind(t *testing.T) {
	s := NewMemoryEntryStore()
	s.Save(&MemoryEntry{ID: "1", Scope: MemoryScopeFolder, Kind: MemoryKindPreference, Content: "a"})
	s.Save(&MemoryEntry{ID: "2", Scope: MemoryScopeGlobal, Kind: MemoryKindPreference, Content: "b"})
	s.Save(&MemoryEntry{ID: "3", Scope: MemoryScopeFolder, Kind: MemoryKindDecision, Content: "c"})

	list, err := s.ListByScopeAndKind(MemoryScopeFolder, MemoryKindPreference)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "1", list[0].ID)
}

func TestMemoryEntryStore_ListByScopeAndKind_Empty(t *testing.T) {
	s := NewMemoryEntryStore()
	list, err := s.ListByScopeAndKind(MemoryScopeFolder, MemoryKindDecision)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

// INVARIANT: ApplyTimeDecay decreases confidence exponentially by age.

func TestMemoryEntry_ApplyTimeDecay(t *testing.T) {
	now := time.Now()
	// Old entry — should decay significantly
	e := &MemoryEntry{
		Confidence: 0.8,
		Source:     MemorySourceRepoDerived,
		UpdatedAt:  now.Add(-720 * time.Hour), // 30 days ago
	}
	newConf := e.ApplyTimeDecay(now)
	assert.Less(t, newConf, 0.8, "confidence should decay")
	assert.GreaterOrEqual(t, newConf, 0.0, "should stay non-negative")

	// Recent entry — minimal decay
	e2 := &MemoryEntry{
		Confidence: 0.8,
		Source:     MemorySourceUserExplicit,
		UpdatedAt:  now.Add(-1 * time.Hour),
	}
	newConf2 := e2.ApplyTimeDecay(now)
	assert.Greater(t, newConf2, 0.7, "recent explicit entry should barely decay")
}

// INVARIANT: error messages render.

func TestMemoryErrors(t *testing.T) {
	assert.Contains(t, (&InvalidScopeError{Scope: "bad"}).Error(), "bad")
	assert.Contains(t, (&EntryNotFoundError{ID: "x"}).Error(), "x")
}

// INVARIANT: ValidateConfidenceRange.

func TestMemoryEntry_ValidateConfidenceRange(t *testing.T) {
	e := &MemoryEntry{Confidence: 0.5}
	assert.True(t, e.ValidateConfidenceRange())

	e.Confidence = 1.5
	assert.False(t, e.ValidateConfidenceRange())

	e.Confidence = -0.1
	assert.False(t, e.ValidateConfidenceRange())
}

// INVARIANT: GetStorageLocation.

func TestMemoryEntry_GetStorageLocation(t *testing.T) {
	e := &MemoryEntry{Scope: MemoryScopeFolder, FolderID: "f1"}
	loc, err := e.GetStorageLocation()
	require.NoError(t, err)
	assert.Contains(t, loc, "f1")

	e.Scope = MemoryScopeGlobal
	loc, err = e.GetStorageLocation()
	require.NoError(t, err)
	assert.NotEmpty(t, loc)
}
