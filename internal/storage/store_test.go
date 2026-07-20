package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: the graph store round-trips nodes and edges; list returns
// all; Get errors on missing IDs. ValidateDAG detects cycles via DFS.

func TestGraphStore_SaveGetNode(t *testing.T) {
	s := NewGraphStore()
	n := &Node{ID: "n1", Kind: "folder", CreatedAt: time.Now()}
	s.SaveNodes(n)

	got, err := s.GetNode("n1")
	require.NoError(t, err)
	assert.Equal(t, "n1", got.ID)
}

func TestGraphStore_GetNodeMissing(t *testing.T) {
	s := NewGraphStore()
	_, err := s.GetNode("nope")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGraphStore_SaveGetEdge(t *testing.T) {
	s := NewGraphStore()
	e := &Edge{ID: "e1", SourceID: "a", TargetID: "b", Kind: EdgeKindContains}
	s.SaveEdges(e)

	got, err := s.GetEdge("e1")
	require.NoError(t, err)
	assert.Equal(t, "e1", got.ID)
}

func TestGraphStore_GetEdgeMissing(t *testing.T) {
	s := NewGraphStore()
	_, err := s.GetEdge("nope")
	assert.Error(t, err)
}

func TestGraphStore_ListNodes(t *testing.T) {
	s := NewGraphStore()
	s.SaveNodes(&Node{ID: "a"}, &Node{ID: "b"})
	list, err := s.ListNodes()
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestGraphStore_ListEdges(t *testing.T) {
	s := NewGraphStore()
	s.SaveEdges(&Edge{ID: "e1"}, &Edge{ID: "e2"})
	list, err := s.ListEdges()
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestGraphStore_EmptyList(t *testing.T) {
	s := NewGraphStore()
	nodes, _ := s.ListNodes()
	edges, _ := s.ListEdges()
	assert.Len(t, nodes, 0)
	assert.Len(t, edges, 0)
}

// INVARIANT: ValidateEdgeReferentialIntegrity (the error-returning variant)
// must name the missing endpoint in its error.

func TestValidateEdgeReferentialIntegrity_MissingSource(t *testing.T) {
	s := NewGraphStore()
	s.UpsertNodes([]Node{{ID: "n1"}})
	e := &Edge{ID: "e1", SourceID: "missing", TargetID: "n1"}
	err := e.ValidateEdgeReferentialIntegrity(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestValidateEdgeReferentialIntegrity_MissingTarget(t *testing.T) {
	s := NewGraphStore()
	s.UpsertNodes([]Node{{ID: "n1"}})
	e := &Edge{ID: "e1", SourceID: "n1", TargetID: "missing"}
	err := e.ValidateEdgeReferentialIntegrity(s)
	require.Error(t, err)
}

func TestValidateEdgeReferentialIntegrity_Valid(t *testing.T) {
	s := NewGraphStore()
	s.UpsertNodes([]Node{{ID: "n1"}, {ID: "n2"}})
	e := &Edge{ID: "e1", SourceID: "n1", TargetID: "n2"}
	assert.NoError(t, e.ValidateEdgeReferentialIntegrity(s))
}

// INVARIANT: ValidateNoDuplicateEdges (error variant) names the duplicate.

func TestValidateNoDuplicateEdges_Error_Duplicate(t *testing.T) {
	s := NewGraphStore()
	s.UpsertEdges([]Edge{{ID: "e1", SourceID: "a", TargetID: "b", Kind: EdgeKindContains}})
	e := &Edge{ID: "e2", SourceID: "a", TargetID: "b", Kind: EdgeKindContains}
	err := e.ValidateNoDuplicateEdges(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestValidateNoDuplicateEdges_Error_NoDup(t *testing.T) {
	s := NewGraphStore()
	e := &Edge{ID: "e1", SourceID: "a", TargetID: "b", Kind: EdgeKindContains}
	assert.NoError(t, e.ValidateNoDuplicateEdges(s))
}

// INVARIANT: ValidateDAG detects cycles via DFS coloring, returns the
// cycle path in the error.

func TestValidateDAG_CycleDetected(t *testing.T) {
	s := NewGraphStore()
	s.UpsertNodes([]Node{{ID: "a"}, {ID: "b"}})
	s.UpsertEdges([]Edge{
		{SourceID: "a", TargetID: "b", Kind: EdgeKindContains},
		{SourceID: "b", TargetID: "a", Kind: EdgeKindContains},
	})
	err := ValidateDAG(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

func TestValidateDAG_NoCycle(t *testing.T) {
	s := NewGraphStore()
	s.UpsertNodes([]Node{{ID: "a"}, {ID: "b"}, {ID: "c"}})
	s.UpsertEdges([]Edge{
		{SourceID: "a", TargetID: "b", Kind: EdgeKindContains},
		{SourceID: "b", TargetID: "c", Kind: EdgeKindContains},
	})
	assert.NoError(t, ValidateDAG(s))
}

// INVARIANT: error types render meaningful messages.

func TestErrorMessages(t *testing.T) {
	assert.Contains(t, (&ReferentialIntegrityError{MissingID: "x"}).Error(), "x")
	assert.Contains(t, (&DuplicateEdgeError{SourceID: "s"}).Error(), "s")
	assert.Contains(t, (&ConfidenceRangeError{Reason: "bad"}).Error(), "bad")
	assert.Contains(t, (&CycleError{Cycle: []string{"a", "b"}}).Error(), "a")
	assert.Contains(t, (&NodeNotFoundError{ID: "n"}).Error(), "n")
	assert.Contains(t, (&EdgeNotFoundError{ID: "e"}).Error(), "e")
}
