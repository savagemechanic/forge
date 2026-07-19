package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// GRAPH INVARIANT TESTS
// ============================================================================

// INVARIANT 1: Every edge.source_id and edge.target_id must exist in nodes
func TestGraph_EdgeReferentialIntegrity(t *testing.T) {
	t.Run("should fail when edge source_id does not exist", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add node
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder", Name: "test"},
		})

		// Add edge with non-existent source
		edge := Edge{
			SourceID: "node-nonexistent", // INVALID
			TargetID: "node-1",
			Kind:     "CONTAINS",
		}

		valid := store.ValidateReferentialIntegrity(edge)
		assert.False(t, valid, "Edge with non-existent source should fail")
	})

	t.Run("should fail when edge target_id does not exist", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add node
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder", Name: "test"},
		})

		// Add edge with non-existent target
		edge := Edge{
			SourceID: "node-1",
			TargetID: "node-nonexistent", // INVALID
			Kind:     "CONTAINS",
		}

		valid := store.ValidateReferentialIntegrity(edge)
		assert.False(t, valid, "Edge with non-existent target should fail")
	})

	t.Run("should fail when both edge endpoints do not exist", func(t *testing.T) {
		store := NewMemoryGraphStore()

		edge := Edge{
			SourceID: "node-nonexistent-1",
			TargetID: "node-nonexistent-2",
			Kind:     "CONTAINS",
		}

		valid := store.ValidateReferentialIntegrity(edge)
		assert.False(t, valid, "Edge with both endpoints non-existent should fail")
	})

	t.Run("should pass when both endpoints exist", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder", Name: "folder1"},
			{ID: "node-2", Kind: "package", Name: "pkg"},
		})

		// Add valid edge
		edge := Edge{
			SourceID: "node-1",
			TargetID: "node-2",
			Kind:     "CONTAINS",
		}

		valid := store.ValidateReferentialIntegrity(edge)
		assert.True(t, valid, "Edge with existing endpoints should pass")
	})
}

// INVARIANT 2: (source_id, target_id, kind) is unique (no duplicate edges)
func TestGraph_NoDuplicateEdges(t *testing.T) {
	t.Run("should fail when duplicate edge exists", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder"},
			{ID: "node-2", Kind: "package"},
		})

		// Add first edge
		store.UpsertEdges([]Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "CONTAINS"},
		})

		// Try to add duplicate
		edge := Edge{
			SourceID: "node-1",
			TargetID: "node-2",
			Kind:     "CONTAINS", // DUPLICATE
		}

		valid := store.ValidateNoDuplicateEdges(edge)
		assert.False(t, valid, "Duplicate edge should fail validation")
	})

	t.Run("should pass when same endpoints have different kind", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder"},
			{ID: "node-2", Kind: "package"},
		})

		// Add first edge
		store.UpsertEdges([]Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "CONTAINS"},
		})

		// Add same endpoints with different kind
		edge := Edge{
			SourceID: "node-1",
			TargetID: "node-2",
			Kind:     "DEPENDS_ON", // DIFFERENT KIND
		}

		valid := store.ValidateNoDuplicateEdges(edge)
		assert.True(t, valid, "Same endpoints with different kind should pass")
	})

	t.Run("should pass when different endpoints", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder"},
			{ID: "node-2", Kind: "package"},
			{ID: "node-3", Kind: "package"},
		})

		// Add first edge
		store.UpsertEdges([]Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "CONTAINS"},
		})

		// Add edge with different endpoint
		edge := Edge{
			SourceID: "node-1",
			TargetID: "node-3", // DIFFERENT TARGET
			Kind:     "CONTAINS",
		}

		valid := store.ValidateNoDuplicateEdges(edge)
		assert.True(t, valid, "Edge with different endpoint should pass")
	})

	t.Run("should handle multiple duplicate checks", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder"},
			{ID: "node-2", Kind: "package"},
		})

		// Add multiple edges of same type
		edges := []Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "CONTAINS"},
			{SourceID: "node-1", TargetID: "node-2", Kind: "CONTAINS"}, // DUPLICATE
		}

		store.UpsertEdges(edges)

		valid := store.ValidateNoDuplicateEdgesGlobal()
		assert.False(t, valid, "Multiple duplicate edges should fail")
	})
}

// INVARIANT 3: Graph is a DAG for CONTAINS/DEPENDS_ON relationships (no cycles)
func TestGraph_ContainsNoCycles(t *testing.T) {
	t.Run("should fail when CONTAINS edge creates a 2-node cycle", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder"},
			{ID: "node-2", Kind: "folder"},
		})

		// Add edges in both directions (cycle)
		edges := []Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "CONTAINS"},
			{SourceID: "node-2", TargetID: "node-1", Kind: "CONTAINS"}, // CYCLE
		}

		store.UpsertEdges(edges)

		valid := store.ValidateNoCycles("CONTAINS")
		assert.False(t, valid, "2-node cycle should fail validation")
	})

	t.Run("should fail when CONTAINS edge creates a deep cycle", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder"},
			{ID: "node-2", Kind: "folder"},
			{ID: "node-3", Kind: "folder"},
			{ID: "node-4", Kind: "folder"},
		})

		// Add edges forming a cycle: 1 → 2 → 3 → 4 → 1
		edges := []Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "CONTAINS"},
			{SourceID: "node-2", TargetID: "node-3", Kind: "CONTAINS"},
			{SourceID: "node-3", TargetID: "node-4", Kind: "CONTAINS"},
			{SourceID: "node-4", TargetID: "node-1", Kind: "CONTAINS"}, // CYCLE
		}

		store.UpsertEdges(edges)

		valid := store.ValidateNoCycles("CONTAINS")
		assert.False(t, valid, "Deep cycle should fail validation")
	})

	t.Run("should fail when DEPENDS_ON edge creates a cycle", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "package"},
			{ID: "node-2", Kind: "package"},
			{ID: "node-3", Kind: "package"},
		})

		// Add dependency edges forming a cycle
		edges := []Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "DEPENDS_ON"},
			{SourceID: "node-2", TargetID: "node-3", Kind: "DEPENDS_ON"},
			{SourceID: "node-3", TargetID: "node-1", Kind: "DEPENDS_ON"}, // CYCLE
		}

		store.UpsertEdges(edges)

		valid := store.ValidateNoCycles("DEPENDS_ON")
		assert.False(t, valid, "Dependency cycle should fail validation")
	})

	t.Run("should pass when CONTAINS edges form a tree", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "root", Kind: "folder"},
			{ID: "child1", Kind: "folder"},
			{ID: "child2", Kind: "folder"},
			{ID: "grandchild1", Kind: "package"},
			{ID: "grandchild2", Kind: "package"},
		})

		// Add tree edges
		edges := []Edge{
			{SourceID: "root", TargetID: "child1", Kind: "CONTAINS"},
			{SourceID: "root", TargetID: "child2", Kind: "CONTAINS"},
			{SourceID: "child1", TargetID: "grandchild1", Kind: "CONTAINS"},
			{SourceID: "child1", TargetID: "grandchild2", Kind: "CONTAINS"},
		}

		store.UpsertEdges(edges)

		valid := store.ValidateNoCycles("CONTAINS")
		assert.True(t, valid, "Tree structure should pass validation")
	})

	t.Run("should pass when DEPENDS_ON edges form a DAG", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "pkg-a", Kind: "package"},
			{ID: "pkg-b", Kind: "package"},
			{ID: "pkg-c", Kind: "package"},
			{ID: "pkg-d", Kind: "package"},
		})

		// Add DAG edges: c and d both depend on b, b depends on a
		edges := []Edge{
			{SourceID: "pkg-c", TargetID: "pkg-b", Kind: "DEPENDS_ON"},
			{SourceID: "pkg-d", TargetID: "pkg-b", Kind: "DEPENDS_ON"},
			{SourceID: "pkg-b", TargetID: "pkg-a", Kind: "DEPENDS_ON"},
		}

		store.UpsertEdges(edges)

		valid := store.ValidateNoCycles("DEPENDS_ON")
		assert.True(t, valid, "DAG structure should pass validation")
	})

	t.Run("should pass when edge kind is not constrained", func(t *testing.T) {
		store := NewMemoryGraphStore()

		// Add nodes
		store.UpsertNodes([]Node{
			{ID: "node-1", Kind: "folder"},
			{ID: "node-2", Kind: "folder"},
		})

		// Add cyclic edge with unconstrained kind
		edges := []Edge{
			{SourceID: "node-1", TargetID: "node-2", Kind: "RELATED"},
			{SourceID: "node-2", TargetID: "node-1", Kind: "RELATED"}, // CYCLE OK FOR THIS KIND
		}

		store.UpsertEdges(edges)

		// CONTAINS constraint should still pass (edges are different kind)
		valid := store.ValidateNoCycles("CONTAINS")
		assert.True(t, valid, "Cyclic RELATED edges should not affect CONTAINS constraint")
	})
}

// INVARIANT 4: Confidence is 0.0 to 1.0
func TestGraph_EdgeConfidenceRange(t *testing.T) {
	t.Run("should fail when confidence is negative", func(t *testing.T) {
		edge := Edge{
			SourceID:  "node-1",
			TargetID:  "node-2",
			Kind:      "CONTAINS",
			Confidence: -0.1, // INVALID
		}

		valid := edge.ValidateConfidenceRange()
		assert.False(t, valid, "Negative confidence should fail")
	})

	t.Run("should fail when confidence is > 1.0", func(t *testing.T) {
		edge := Edge{
			SourceID:  "node-1",
			TargetID:  "node-2",
			Kind:      "CONTAINS",
			Confidence: 1.1, // INVALID
		}

		valid := edge.ValidateConfidenceRange()
		assert.False(t, valid, "Confidence > 1.0 should fail")
	})

	t.Run("should pass when confidence is 0.0", func(t *testing.T) {
		edge := Edge{
			SourceID:  "node-1",
			TargetID:  "node-2",
			Kind:      "CONTAINS",
			Confidence: 0.0, // VALID
		}

		valid := edge.ValidateConfidenceRange()
		assert.True(t, valid, "Confidence 0.0 should pass")
	})

	t.Run("should pass when confidence is 1.0", func(t *testing.T) {
		edge := Edge{
			SourceID:  "node-1",
			TargetID:  "node-2",
			Kind:      "CONTAINS",
			Confidence: 1.0, // VALID
		}

		valid := edge.ValidateConfidenceRange()
		assert.True(t, valid, "Confidence 1.0 should pass")
	})

	t.Run("should pass when confidence is between 0 and 1", func(t *testing.T) {
		edge := Edge{
			SourceID:  "node-1",
			TargetID:  "node-2",
			Kind:      "CONTAINS",
			Confidence: 0.5, // VALID
		}

		valid := edge.ValidateConfidenceRange()
		assert.True(t, valid, "Confidence 0.5 should pass")
	})
}