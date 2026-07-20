package storage

import (
	"fmt"
	"testing"
)

// BENCHMARKS
// INVARIANT: cycle detection (DFS) must scale. These catch perf regressions
// on large dependency graphs. Run with: go test -bench=. ./internal/storage/

// buildDAG creates a DAG with n nodes in a chain (worst case for DFS —
// must traverse the full depth before confirming no cycle).
func buildDAG(b *testing.B, n int) *InMemoryGraphStore {
	store := NewMemoryGraphStore()
	nodes := make([]Node, n)
	for i := 0; i < n; i++ {
		nodes[i] = Node{ID: idN(i), Kind: "pkg"}
	}
	store.UpsertNodes(nodes)

	edges := make([]Edge, n-1)
	for i := 0; i < n-1; i++ {
		edges[i] = Edge{SourceID: idN(i), TargetID: idN(i + 1), Kind: EdgeKindContains}
	}
	store.UpsertEdges(edges)
	return store
}

// buildCycle creates a graph with a cycle for the worst-case detection.
func buildCycle(b *testing.B, n int) *InMemoryGraphStore {
	store := buildDAG(b, n)
	// close the loop: last → first
	store.UpsertEdges([]Edge{{SourceID: idN(n - 1), TargetID: idN(0), Kind: EdgeKindContains}})
	return store
}

func idN(i int) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	if i < len(digits) {
		return "n" + string(digits[i])
	}
	return "n" + string(digits[i%len(digits)]) + string(digits[i/len(digits)])
}

func BenchmarkValidateNoCycles_DAG(b *testing.B) {
	for _, n := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("dag_%d", n), func(b *testing.B) {
			store := buildDAG(b, n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.ValidateNoCycles("CONTAINS")
			}
		})
	}
}

func BenchmarkValidateNoCycles_Cycle(b *testing.B) {
	for _, n := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("cycle_%d", n), func(b *testing.B) {
			store := buildCycle(b, n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.ValidateNoCycles("CONTAINS")
			}
		})
	}
}

func BenchmarkValidateNoDuplicateEdges(b *testing.B) {
	store := NewMemoryGraphStore()
	store.UpsertNodes([]Node{{ID: "a"}, {ID: "b"}})
	store.UpsertEdges([]Edge{{SourceID: "a", TargetID: "b", Kind: EdgeKindContains}})
	edge := Edge{SourceID: "a", TargetID: "b", Kind: EdgeKindContains}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.ValidateNoDuplicateEdges(edge)
	}
}
