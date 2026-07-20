package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: the session store round-trips, lists by folder, lists active.

func TestSessionStore_ListByFolder(t *testing.T) {
	s := NewMemorySessionStore()
	s.Save(&Session{ID: "s1", FolderID: "f1", CreatedAt: time.Now()})
	s.Save(&Session{ID: "s2", FolderID: "f2", CreatedAt: time.Now()})
	s.Save(&Session{ID: "s3", FolderID: "f1", CreatedAt: time.Now()})

	list, err := s.ListByFolder("f1")
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestSessionStore_ListByFolder_Empty(t *testing.T) {
	s := NewMemorySessionStore()
	list, err := s.ListByFolder("nothing")
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestSessionStore_ListActive(t *testing.T) {
	s := NewMemorySessionStore()
	now := time.Now()
	s.Save(&Session{ID: "s1", FolderID: "f", State: SessionStateActive, CreatedAt: now, LastActiveAt: now})
	s.Save(&Session{ID: "s2", FolderID: "f", State: SessionStateArchived, CreatedAt: now, LastActiveAt: now})

	list, err := s.ListActive()
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "s1", list[0].ID)
}

func TestSessionStore_GetMissing(t *testing.T) {
	s := NewMemorySessionStore()
	_, err := s.Get("nope")
	assert.Error(t, err)
}

// INVARIANT: ValidateIDIsDeterministicHash — ID is SHA256 of folder+time.

func TestSession_ValidateIDIsDeterministicHash(t *testing.T) {
	now := time.Now()
	s := &Session{
		ID:        generateSessionID("folder-1", now),
		FolderID:  "folder-1",
		CreatedAt: now,
	}
	assert.True(t, s.ValidateIDIsDeterministicHash())

	s.ID = "wrong"
	assert.False(t, s.ValidateIDIsDeterministicHash())
}

// INVARIANT: ValidateMessageTimestamps — strictly ordered by .Timestamp.

func TestSession_ValidateMessageTimestamps(t *testing.T) {
	base := time.Now()
	s := &Session{
		Messages: []Message{
			{Timestamp: base},
			{Timestamp: base.Add(time.Second)},
		},
	}
	assert.True(t, s.ValidateMessageTimestamps())

	// Out of order
	s.Messages[1].Timestamp = base.Add(-time.Second)
	assert.False(t, s.ValidateMessageTimestamps())

	// Duplicate
	s.Messages[1].Timestamp = base
	assert.False(t, s.ValidateMessageTimestamps())
}

// INVARIANT: ValidateActiveRunSessionIDMatch — run belongs to session.

func TestSession_ValidateActiveRunSessionIDMatch(t *testing.T) {
	rs := NewMemoryRunStore()
	rs.Save(&Run{ID: "r1", SessionID: "s1"})

	s := &Session{ID: "s1", ActiveRunID: "r1"}
	assert.True(t, s.ValidateActiveRunSessionIDMatch(rs))

	// Run belongs to different session
	rs.Save(&Run{ID: "r2", SessionID: "other"})
	s.ActiveRunID = "r2"
	assert.False(t, s.ValidateActiveRunSessionIDMatch(rs))

	// Nonexistent run
	s.ActiveRunID = "nope"
	assert.False(t, s.ValidateActiveRunSessionIDMatch(rs))
}

// INVARIANT: UpdateLastActiveAt enforces monotonicity.

func TestSession_UpdateLastActiveAt(t *testing.T) {
	now := time.Now()
	s := &Session{CreatedAt: now, LastActiveAt: now, UpdatedAt: now}

	// First update should succeed (now is >= LastActiveAt)
	err := s.UpdateLastActiveAt()
	assert.NoError(t, err)
}

// INVARIANT: ValidateStateTransitionsForwardOnly — uses canTransition lookup.

func TestRun_ValidateStateTransitionsForwardOnly(t *testing.T) {
	r := &Run{State: RunStateExecuting}
	// Planning → Executing is NOT direct (must go via Approved)
	assert.False(t, r.ValidateStateTransitionsForwardOnly(RunStatePlanning))

	// Approved → Executing IS allowed
	r2 := &Run{State: RunStateExecuting}
	assert.True(t, r2.ValidateStateTransitionsForwardOnly(RunStateApproved))

	// Executing → Validating IS allowed
	r3 := &Run{State: RunStateValidating}
	assert.True(t, r3.ValidateStateTransitionsForwardOnly(RunStateExecuting))

	// Backward: Executing → Planning NOT allowed
	r4 := &Run{State: RunStatePlanning}
	assert.False(t, r4.ValidateStateTransitionsForwardOnly(RunStateExecuting))
}

// INVARIANT: ValidateOperationsImmutableWithSnapshot + ValidateOperationsAgainstSnapshot.

func TestRun_ValidateOperationsImmutableWithSnapshot(t *testing.T) {
	r := &Run{Operations: []Operation{{ID: "o1"}}}
	snap := r.OperationsSnapshot()
	assert.True(t, r.ValidateOperationsImmutableWithSnapshot(snap))

	// Modify an operation — should detect
	r.Operations[0].ID = "changed"
	assert.False(t, r.ValidateOperationsImmutableWithSnapshot(snap))
}

// INVARIANT: ValidateValidationState — validation data only in valid states.

func TestRun_ValidateValidationState(t *testing.T) {
	now := time.Now()
	// Populated validation in DONE state — valid
	r := &Run{State: RunStateDone, Validation: Validation{PassedAt: &now}}
	assert.True(t, r.ValidateValidationState())

	// Populated validation in EXECUTING state — invalid
	r2 := &Run{State: RunStateExecuting, Validation: Validation{PassedAt: &now}}
	assert.False(t, r2.ValidateValidationState())

	// Empty validation in any state — valid
	r3 := &Run{State: RunStateExecuting}
	assert.True(t, r3.ValidateValidationState())
}

// INVARIANT: OperationsSnapshot returns a copy (not the same slice).

func TestRun_OperationsSnapshot_IsCopy(t *testing.T) {
	r := &Run{Operations: []Operation{{ID: "o1"}}}
	snap := r.OperationsSnapshot()
	snap[0].ID = "mutated"
	// Original must be unchanged
	assert.Equal(t, "o1", r.Operations[0].ID)
}

// INVARIANT: ValidateAtMostOneActivePerFolder — across whole store.

func TestValidateAtMostOneActivePerFolder(t *testing.T) {
	s := NewMemorySessionStore()
	now := time.Now()
	s.Save(&Session{ID: "a", FolderID: "f", State: SessionStateActive, CreatedAt: now})
	s.Save(&Session{ID: "b", FolderID: "f", State: SessionStateActive, CreatedAt: now})

	valid, err := ValidateAtMostOneActivePerFolder(s)
	require.NoError(t, err)
	assert.False(t, valid)
}
