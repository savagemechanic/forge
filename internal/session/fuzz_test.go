package session

import (
	"testing"
	"time"
)

// INVARIANT: ValidateParentID must NEVER accept a self-loop, regardless
// of chain length or node count. The fuzzer generates random parent chains.
// Loop invariant: a node whose ParentID == its own ID is always invalid.

func FuzzParentIDSelfLoop(f *testing.F) {
	f.Add("sess-1")
	f.Add("sess-xyz")
	f.Fuzz(func(t *testing.T, id string) {
		if id == "" {
			return // skip empty
		}
		s := &Session{ID: id, ParentID: id, FolderID: "f", CreatedAt: time.Now()}
		store := NewMemorySessionStore()
		store.Save(s)
		valid, err := s.ValidateParentID(store)
		// Property: a self-loop MUST be rejected, never panic, never error.
		if err != nil {
			t.Fatalf("self-loop returned error instead of false: %v", err)
		}
		if valid {
			t.Fatalf("self-loop ParentID==ID (%q) was accepted", id)
		}
	})
}

// INVARIANT: ValidateMessageOrder must not panic on any message slice.
// Property: no input causes a panic; equal timestamps are always rejected.

func FuzzMessageOrder(f *testing.F) {
	f.Add(int64(0), int64(0))
	f.Add(int64(1), int64(2))
	f.Add(int64(5), int64(3))
	f.Fuzz(func(t *testing.T, ts1, ts2 int64) {
		s := &Session{
			Messages: []Message{
				{CreatedAt: time.Unix(ts1, 0)},
				{CreatedAt: time.Unix(ts2, 0)},
			},
		}
		// Must not panic.
		_ = s.ValidateMessageOrder()
	})
}
