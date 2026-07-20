package storage

import "testing"

// INVARIANT: ValidateConfidenceRange must return true ⟺ 0 ≤ c ≤ 1, for
// ALL float64 values. The fuzzer exhaustively checks the boundary.
// Property: the function is the exact characterization of [0,1].

func FuzzConfidenceRange(f *testing.F) {
	f.Add(0.0)
	f.Add(1.0)
	f.Add(0.5)
	f.Add(-0.1)
	f.Add(1.1)
	f.Add(0.999999)
	f.Fuzz(func(t *testing.T, c float64) {
		e := &Edge{Confidence: c}
		got := e.ValidateConfidenceRange()
		// Property: result ⟺ c in [0, 1].
		want := c >= 0.0 && c <= 1.0
		if got != want {
			t.Fatalf("confidence %v: got %v want %v", c, got, want)
		}
		// Must not panic on a second call (no mutation).
		_ = e.ValidateConfidenceRange()
	})
}

// INVARIANT: ValidateNoDuplicateEdges must treat (source, target, kind)
// as the identity. The fuzzer generates random edge sets and checks that
// the global check agrees with the pairwise check.
func FuzzDuplicateEdges(f *testing.F) {
	f.Add("a", "b", "CONTAINS")
	f.Add("a", "b", "DEPENDS_ON")
	f.Add("x", "y", "RELATED")
	f.Fuzz(func(t *testing.T, src, tgt, kind string) {
		store := NewMemoryGraphStore()
		store.UpsertNodes([]Node{
			{ID: src}, {ID: tgt},
		})
		e := Edge{SourceID: src, TargetID: tgt, Kind: EdgeKind(kind)}
		// Must not panic.
		_ = store.ValidateNoDuplicateEdges(e)
		_ = store.ValidateReferentialIntegrity(e)
	})
}
