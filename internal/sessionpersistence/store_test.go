package sessionpersistence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudspacelab/forge/internal/ports"
)

// INVARIANT: Save → Get round-trips a session record exactly.
// Structural invariant: the store never loses or corrupts a record.

func TestFileSessionStore_RoundTrip(t *testing.T) {
	store := NewFileSessionStore(t.TempDir())

	rec := &ports.SessionRecord{
		ID:        "sess-1",
		FolderID:  "folder-1",
		State:     "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []ports.SessionMessage{
			{ID: "m1", Role: "user", Content: "hello", CreatedAt: time.Now()},
			{ID: "m2", Role: "assistant", Content: "hi", CreatedAt: time.Now()},
		},
	}
	require.NoError(t, store.Save(rec))

	got, err := store.Get("sess-1")
	require.NoError(t, err)
	assert.Equal(t, rec.ID, got.ID)
	assert.Equal(t, rec.FolderID, got.FolderID)
	assert.Equal(t, rec.State, got.State)
	assert.Len(t, got.Messages, 2)
	assert.Equal(t, "hello", got.Messages[0].Content)
}

func TestFileSessionStore_GetMissing(t *testing.T) {
	store := NewFileSessionStore(t.TempDir())
	_, err := store.Get("nonexistent")
	assert.Error(t, err)
}

func TestFileSessionStore_ListByFolder(t *testing.T) {
	store := NewFileSessionStore(t.TempDir())

	now := time.Now()
	recs := []*ports.SessionRecord{
		{ID: "s1", FolderID: "f1", State: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "s2", FolderID: "f1", State: "active", CreatedAt: now, UpdatedAt: now.Add(time.Second)},
		{ID: "s3", FolderID: "f2", State: "active", CreatedAt: now, UpdatedAt: now},
	}
	for _, r := range recs {
		require.NoError(t, store.Save(r))
	}

	// Filter by folder f1
	list, err := store.ListByFolder("f1")
	require.NoError(t, err)
	assert.Len(t, list, 2)

	// Newest first (s2 has later UpdatedAt)
	assert.Equal(t, "s2", list[0].ID)

	// Empty filter returns all
	all, err := store.ListByFolder("")
	require.NoError(t, err)
	assert.Len(t, all, 3)
}

func TestFileSessionStore_EmptyDir(t *testing.T) {
	store := NewFileSessionStore(t.TempDir())
	list, err := store.ListByFolder("anything")
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRecordFromSession(t *testing.T) {
	now := time.Now()
	rec := RecordFromSession("s1", "f1", "active", []sessionMessage{
		{ID: "m1", Role: "user", Content: "hi", CreatedAt: now},
	})
	assert.Equal(t, "s1", rec.ID)
	assert.Len(t, rec.Messages, 1)
	assert.Equal(t, "user", rec.Messages[0].Role)
}
