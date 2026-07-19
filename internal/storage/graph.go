package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type Node struct {
	ID             string
	Kind           string
	Name           string
	QualifiedName  string
	Metadata       map[string]interface{}
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Edge struct {
	SourceID   string
	TargetID   string
	Kind       string
	Confidence float64
	Source     string // Where this edge came from
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type GraphStore interface {
	UpsertNodes(nodes []Node) error
	UpsertEdges(edges []Edge) error
	GetNode(id string) (*Node, error)
	GetEdge(sourceID, targetID, kind string) (*Edge, error)
	Neighbors(id string, query EdgeQuery) ([]Node, error)
	Traverse(seeds []string, policy TraversalPolicy) (*Subgraph, error)
	Search(query SearchQuery) ([]SearchResult, error)
}

type EdgeQuery struct {
	Kinds      []string
	MinConfidence float64
	Direction string // "in", "out", "both"
	Limit      int
}

type TraversalPolicy struct {
	MaxDepth    int
	Kinds       []string
	MinConfidence float64
}

type Subgraph struct {
	Nodes []Node
	Edges []Edge
}

type SearchQuery struct {
	Kind          string
	NamePattern   string
	Metadata      map[string]interface{}
	EdgeKinds     []string
	Limit         int
}

type SearchResult struct {
	Node      Node
	Edge      *Edge // nil if searching nodes only
	Score     float64
}

// ============================================================================
// IN-MEMORY GRAPH STORE (STUB - ALL FAIL TESTS)
// ============================================================================

type MemoryGraphStore struct {
	nodes map[string]Node
	edges map[string]Edge // key: "sourceID|targetID|kind"
	mu    sync.RWMutex
}

func NewMemoryGraphStore() *MemoryGraphStore {
	return &MemoryGraphStore{
		nodes: make(map[string]Node),
		edges: make(map[string]Edge),
	}
}

func (g *MemoryGraphStore) UpsertNodes(nodes []Node) error {
	// STUB: Always return error to fail tests
	return errors.New("not implemented")
}

func (g *MemoryGraphStore) UpsertEdges(edges []Edge) error {
	// STUB: Always return error to fail tests
	return errors.New("not implemented")
}

func (g *MemoryGraphStore) GetNode(id string) (*Node, error) {
	// STUB: Always return error to fail tests
	return nil, errors.New("not implemented")
}

func (g *MemoryGraphStore) GetEdge(sourceID, targetID, kind string) (*Edge, error) {
	// STUB: Always return error to fail tests
	return nil, errors.New("not implemented")
}

func (g *MemoryGraphStore) Neighbors(id string, query EdgeQuery) ([]Node, error) {
	// STUB: Always return error to fail tests
	return nil, errors.New("not implemented")
}

func (g *MemoryGraphStore) Traverse(seeds []string, policy TraversalPolicy) (*Subgraph, error) {
	// STUB: Always return error to fail tests
	return nil, errors.New("not implemented")
}

func (g *MemoryGraphStore) Search(query SearchQuery) ([]SearchResult, error) {
	// STUB: Always return error to fail tests
	return nil, errors.New("not implemented")
}

// ============================================================================
// VALIDATION METHODS (STUB - ALL FAIL TESTS)
// ============================================================================

// ValidateReferentialIntegrity checks that edge endpoints exist
// STUB: Always returns false to fail tests
func (g *MemoryGraphStore) ValidateReferentialIntegrity(edge Edge) bool {
	// TODO: Implement referential integrity check
	// - Check sourceID exists in nodes
	// - Check targetID exists in nodes
	// For now, always fail
	return false
}

// ValidateNoDuplicateEdges checks no duplicate (source, target, kind) edges exist
// STUB: Always returns false to fail tests
func (g *MemoryGraphStore) ValidateNoDuplicateEdges(edge Edge) bool {
	// TODO: Implement duplicate edge detection
	// For now, always fail
	return false
}

// ValidateNoDuplicateEdgesGlobal checks no duplicate edges exist globally
// STUB: Always returns false to fail tests
func (g *MemoryGraphStore) ValidateNoDuplicateEdgesGlobal() bool {
	// TODO: Implement global duplicate detection
	// For now, always fail
	return false
}

// ValidateNoCycles checks that graph is a DAG for given edge kind
// STUB: Always returns false to fail tests
func (g *MemoryGraphStore) ValidateNoCycles(edgeKind string) bool {
	// TODO: Implement cycle detection using DFS
	// - Build adjacency list for given edge kind
	// - Run DFS to detect back edges
	// For now, always fail
	return false
}

// ============================================================================
// EDGE VALIDATION (STUB - ALL FAIL TESTS)
// ============================================================================

// ValidateConfidenceRange checks confidence is in [0, 1]
// STUB: Always returns false to fail tests
func (e *Edge) ValidateConfidenceRange() bool {
	// TODO: Implement range validation
	// - Confidence must be >= 0 and <= 1
	// For now, always fail
	return false
}

// ============================================================================
// HELPERS
// ============================================================================

// edgeKey creates a unique key for an edge
func edgeKey(sourceID, targetID, kind string) string {
	return fmt.Sprintf("%s|%s|%s", sourceID, targetID, kind)
}

// computeNodeID creates a deterministic ID for a node
func computeNodeID(kind, qualifiedName string) string {
	data := kind + "|" + qualifiedName
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// hasCycle uses DFS to detect cycles in a directed graph
func hasCycle(adjacency map[string][]string, start string, visited, recursionStack map[string]bool) bool {
	visited[start] = true
	recursionStack[start] = true

	for _, neighbor := range adjacency[start] {
		if !visited[neighbor] {
			if hasCycle(adjacency, neighbor, visited, recursionStack) {
				return true
			}
		} else if recursionStack[neighbor] {
			return true
		}
	}

	recursionStack[start] = false
	return false
}

// topologicalSort returns nodes in topological order (for DAGs)
func topologicalSort(adjacency map[string][]string) ([]string, error) {
	// TODO: Implement Kahn's algorithm or DFS-based topological sort
	return nil, errors.New("not implemented")
}

// serializeNode converts node to JSON
func serializeNode(node Node) ([]byte, error) {
	return json.Marshal(node)
}

// deserializeNode converts JSON to node
func deserializeNode(data []byte) (Node, error) {
	var node Node
	err := json.Unmarshal(data, &node)
	return node, err
}

// serializeEdge converts edge to JSON
func serializeEdge(edge Edge) ([]byte, error) {
	return json.Marshal(edge)
}

// deserializeEdge converts JSON to edge
func deserializeEdge(data []byte) (Edge, error) {
	var edge Edge
	err := json.Unmarshal(data, &edge)
	return edge, err
}

// DAG edge kinds that must not have cycles
var dageKinds = map[string]bool{
	"CONTAINS":   true,
	"DEPENDS_ON": true,
}

// isDAGKind returns true if edge kind must form a DAG
func isDAGKind(kind string) bool {
	return dageKinds[kind]
}

// buildAdjacencyList builds adjacency list for given edge kind
func (g *MemoryGraphStore) buildAdjacencyList(kind string) map[string][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	adjacency := make(map[string][]string)

	for _, edge := range g.edges {
		if edge.Kind == kind {
			adjacency[edge.SourceID] = append(adjacency[edge.SourceID], edge.TargetID)
		}
	}

	return adjacency
}

// getAllNodeIDs returns all node IDs in the graph
func (g *MemoryGraphStore) getAllNodeIDs() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	ids := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	return ids
}