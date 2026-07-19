package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// SESSION INVARIANT TESTS
// ============================================================================

// INVARIANT 1: Messages are strictly ordered by creation timestamp
func TestSession_MessagesStrictlyOrderedByTimestamp(t *testing.T) {
	t.Run("should fail when messages are out of order", func(t *testing.T) {
		sess := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages: []Message{
				{ID: "msg-1", CreatedAt: time.Now().Add(-2 * time.Hour), Content: "first"},
				{ID: "msg-2", CreatedAt: time.Now().Add(-1 * time.Hour), Content: "second"},
				{ID: "msg-3", CreatedAt: time.Now().Add(-3 * time.Hour), Content: "third"}, // OUT OF ORDER
			},
		}

		// This should fail - messages not ordered
		assert.False(t, sess.ValidateMessageOrder(),
			"Session with out-of-order messages should fail validation")
	})

	t.Run("should fail when messages have same timestamp", func(t *testing.T) {
		now := time.Now()
		sess := &Session{
			ID:        "sess-2",
			FolderID:  "folder-1",
			CreatedAt: now,
			UpdatedAt: now,
			Messages: []Message{
				{ID: "msg-1", CreatedAt: now, Content: "first"},
				{ID: "msg-2", CreatedAt: now, Content: "second"}, // SAME TIMESTAMP
			},
		}

		assert.False(t, sess.ValidateMessageOrder(),
			"Session with duplicate timestamps should fail validation")
	})

	t.Run("should pass when messages are strictly ordered", func(t *testing.T) {
		now := time.Now()
		sess := &Session{
			ID:        "sess-3",
			FolderID:  "folder-1",
			CreatedAt: now,
			UpdatedAt: now,
			Messages: []Message{
				{ID: "msg-1", CreatedAt: now.Add(-2 * time.Hour), Content: "first"},
				{ID: "msg-2", CreatedAt: now.Add(-1 * time.Hour), Content: "second"},
				{ID: "msg-3", CreatedAt: now, Content: "third"},
			},
		}

		assert.True(t, sess.ValidateMessageOrder(),
			"Session with ordered messages should pass validation")
	})
}

// INVARIANT 2: ParentID either empty or points to an existing Session (no cycles)
func TestSession_ParentIDNoCycles(t *testing.T) {
	t.Run("should fail when ParentID points to non-existent session", func(t *testing.T) {
		sess := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			ParentID:  "sess-nonexistent", // INVALID
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemorySessionStore()
		store.Save(sess)

		valid, err := sess.ValidateParentID(store)
		require.NoError(t, err)
		assert.False(t, valid, "Session with non-existent parent should fail validation")
	})

	t.Run("should fail when ParentID creates a cycle", func(t *testing.T) {
		sess1 := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		sess2 := &Session{
			ID:        "sess-2",
			FolderID:  "folder-1",
			ParentID:  "sess-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		sess1.ParentID = "sess-2" // CYCLE: sess1 → sess2 → sess1

		store := NewMemorySessionStore()
		store.Save(sess1)
		store.Save(sess2)

		valid, err := sess1.ValidateParentID(store)
		require.NoError(t, err)
		assert.False(t, valid, "Session with cycle should fail validation")
	})

	t.Run("should fail on deep cycle detection", func(t *testing.T) {
		sess1 := &Session{ID: "sess-1", FolderID: "folder-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		sess2 := &Session{ID: "sess-2", FolderID: "folder-1", ParentID: "sess-1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		sess3 := &Session{ID: "sess-3", FolderID: "folder-1", ParentID: "sess-2", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		sess4 := &Session{ID: "sess-4", FolderID: "folder-1", ParentID: "sess-3", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		sess1.ParentID = "sess-4" // CYCLE: sess1 → sess2 → sess3 → sess4 → sess1

		store := NewMemorySessionStore()
		store.Save(sess1, sess2, sess3, sess4)

		valid, err := sess1.ValidateParentID(store)
		require.NoError(t, err)
		assert.False(t, valid, "Session with deep cycle should fail validation")
	})

	t.Run("should pass when ParentID is empty", func(t *testing.T) {
		sess := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			ParentID:  "", // Empty is valid
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemorySessionStore()
		store.Save(sess)

		valid, err := sess.ValidateParentID(store)
		require.NoError(t, err)
		assert.True(t, valid, "Session with empty ParentID should pass validation")
	})

	t.Run("should pass when ParentID points to valid parent", func(t *testing.T) {
		parent := &Session{
			ID:        "sess-parent",
			FolderID:  "folder-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		child := &Session{
			ID:        "sess-child",
			FolderID:  "folder-1",
			ParentID:  "sess-parent",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemorySessionStore()
		store.Save(parent, child)

		valid, err := child.ValidateParentID(store)
		require.NoError(t, err)
		assert.True(t, valid, "Session with valid parent should pass validation")
	})
}

// INVARIANT 3: ActiveRun is either empty or points to a Run with SessionID == this.ID
func TestSession_ActiveRunSessionIDMatch(t *testing.T) {
	t.Run("should fail when ActiveRun points to non-existent run", func(t *testing.T) {
		sess := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			ActiveRun: "run-nonexistent", // INVALID
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemoryRunStore()
		valid, err := sess.ValidateActiveRun(store)
		require.NoError(t, err)
		assert.False(t, valid, "Session with non-existent ActiveRun should fail validation")
	})

	t.Run("should fail when ActiveRun.SessionID != Session.ID", func(t *testing.T) {
		sess1 := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			ActiveRun: "run-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-2", // DIFFERENT SESSION
			State:     RunStateExecuting,
		}

		store := NewMemoryRunStore()
		store.Save(run)

		valid, err := sess1.ValidateActiveRun(store)
		require.NoError(t, err)
		assert.False(t, valid, "Session with mismatched ActiveRun.SessionID should fail validation")
	})

	t.Run("should pass when ActiveRun is empty", func(t *testing.T) {
		sess := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			ActiveRun: "", // Empty is valid
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemoryRunStore()
		valid, err := sess.ValidateActiveRun(store)
		require.NoError(t, err)
		assert.True(t, valid, "Session with empty ActiveRun should pass validation")
	})

	t.Run("should pass when ActiveRun.SessionID matches", func(t *testing.T) {
		sess := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			ActiveRun: "run-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-1", // MATCHES
			State:     RunStateExecuting,
		}

		store := NewMemoryRunStore()
		store.Save(run)

		valid, err := sess.ValidateActiveRun(store)
		require.NoError(t, err)
		assert.True(t, valid, "Session with matching ActiveRun.SessionID should pass validation")
	})
}

// INVARIANT 4: Exactly one Session per folder can be Status=active
func TestSession_AtMostOneActivePerFolder(t *testing.T) {
	t.Run("should fail when two sessions are active in same folder", func(t *testing.T) {
		sess1 := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			Status:    SessionStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		sess2 := &Session{
			ID:        "sess-2",
			FolderID:  "folder-1", // SAME FOLDER
			Status:    SessionStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemorySessionStore()
		store.Save(sess1, sess2)

		valid, err := ValidateSingleActivePerFolder("folder-1", store)
		require.NoError(t, err)
		assert.False(t, valid, "Folder with multiple active sessions should fail validation")
	})

	t.Run("should fail when more than two sessions are active", func(t *testing.T) {
		sessions := []*Session{
			{ID: "sess-1", FolderID: "folder-1", Status: SessionStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "sess-2", FolderID: "folder-1", Status: SessionStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "sess-3", FolderID: "folder-1", Status: SessionStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		store := NewMemorySessionStore()
		store.Save(sessions...)

		valid, err := ValidateSingleActivePerFolder("folder-1", store)
		require.NoError(t, err)
		assert.False(t, valid, "Folder with >2 active sessions should fail validation")
	})

	t.Run("should pass when only one session is active", func(t *testing.T) {
		sess1 := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			Status:    SessionStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		sess2 := &Session{
			ID:        "sess-2",
			FolderID:  "folder-1",
			Status:    SessionStatusIdle, // NOT ACTIVE
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemorySessionStore()
		store.Save(sess1, sess2)

		valid, err := ValidateSingleActivePerFolder("folder-1", store)
		require.NoError(t, err)
		assert.True(t, valid, "Folder with exactly one active session should pass validation")
	})

	t.Run("should pass when no sessions are active", func(t *testing.T) {
		sess1 := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			Status:    SessionStatusIdle,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		sess2 := &Session{
			ID:        "sess-2",
			FolderID:  "folder-1",
			Status:    SessionStatusCompacted,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemorySessionStore()
		store.Save(sess1, sess2)

		valid, err := ValidateSingleActivePerFolder("folder-1", store)
		require.NoError(t, err)
		assert.True(t, valid, "Folder with no active sessions should pass validation")
	})

	t.Run("should pass when sessions are in different folders", func(t *testing.T) {
		sess1 := &Session{
			ID:        "sess-1",
			FolderID:  "folder-1",
			Status:    SessionStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}
		sess2 := &Session{
			ID:        "sess-2",
			FolderID:  "folder-2", // DIFFERENT FOLDER
			Status:    SessionStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Messages:  []Message{},
		}

		store := NewMemorySessionStore()
		store.Save(sess1, sess2)

		valid, err := ValidateSingleActivePerFolder("folder-1", store)
		require.NoError(t, err)
		assert.True(t, valid, "Each folder can have its own active session")
	})
}