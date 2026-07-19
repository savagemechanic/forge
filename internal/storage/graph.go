package storage

import (
	"fmt"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type EdgeKind string

const (
	EdgeKindContains    EdgeKind = "contains"
	EdgeKindDependsOn   EdgeKind = "depends_on"
	EdgeKindReferences  EdgeKind = "references"
	EdgeKindImplies     EdgeKind = "implies"
	EdgeKindInfluences  EdgeKind = "influences"
)

type Node struct {
	ID        string
	Type      string
	CreatedAt time.Time
}

type Edge struct {
	ID         string
	SourceID   string
	TargetID   string
	Kind       EdgeKind
	Confidence float64
	CreatedAt  time.Time
}

// ============================================================================
// STUB IMPLEMENTATIONS - NOW IMPLEMENTED (GREEN PHASE)
// ============================================================================

// ValidateEdgeReferentialIntegrity checks edge endpoints exist
func (e *Edge) ValidateEdgeReferentialIntegrity(store *InMemoryGraphStore) error {
	// Check source node exists
	if _, err := store.GetNode(e.SourceID); err != nil {
		return &ReferentialIntegrityError{
			EdgeID:    e.ID,
			SourceID:  e.SourceID,
			TargetID:  e.TargetID,
			MissingID: e.SourceID,
		}
	}

	// Check target node exists
	if _, err := store.GetNode(e.TargetID); err != nil {
		return &ReferentialIntegrityError{
			EdgeID:    e.ID,
			SourceID:  e.SourceID,
			TargetID:  e.TargetID,
			MissingID: e.TargetID,
		}
	}

	return nil
}

// ValidateNoDuplicateEdges checks no duplicate (source_id, target_id, kind) edges
func (e *Edge) ValidateNoDuplicateEdges(store *InMemoryGraphStore) error {
	for _, edge := range store.edges {
		if edge.SourceID == e.SourceID &&
			edge.TargetID == e.TargetID &&
			edge.Kind == e.Kind &&
			edge.ID != e.ID {
			return &DuplicateEdgeError{
				EdgeID:   e.ID,
				SourceID: e.SourceID,
				TargetID: e.TargetID,
				Kind:     e.Kind,
			}
		}
	}
	return nil
}

// ValidateConfidenceRange checks confidence is in [0, 1]
func (e *Edge) ValidateConfidenceRange() error {
	if e.Confidence < 0 {
		return &ConfidenceRangeError{
			EdgeID:     e.ID,
			Confidence: e.Confidence,
			Reason:     "confidence is negative",
		}
	}
	if e.Confidence > 1 {
		return &ConfidenceRangeError{
			EdgeID:     e.ID,
			Confidence: e.Confidence,
			Reason:     "confidence > 1.0",
		}
	}
	return nil
}

// ============================================================================
// GRAPH VALIDATION FUNCTIONS
// ============================================================================

// ValidateDAG checks that CONTAINS and DEPENDS_ON edges form a DAG (no cycles)
func ValidateDAG(store *InMemoryGraphStore) error {
	// Build adjacency list for DAG-relevant edges
	adj := make(map[string][]string)
	for _, edge := range store.edges {
		if edge.Kind == EdgeKindContains || edge.Kind == EdgeKindDependsOn {
			adj[edge.SourceID] = append(adj[edge.SourceID], edge.TargetID)
		}
	}

	// Detect cycles using DFS with coloring
	// Colors: 0 = white (unvisited), 1 = gray (in progress), 2 = black (done)
	color := make(map[string]int)
	path := make(map[string]string)

	var hasCycle bool
	var cycleStart, cycleEnd string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		color[node] = 1 // Gray
		for _, neighbor := range adj[node] {
			if color[neighbor] == 0 {
				path[neighbor] = node
				if dfs(neighbor) {
					return true
				}
			} else if color[neighbor] == 1 {
				// Found cycle
				hasCycle = true
				cycleEnd = neighbor
				cycleStart = node
				return true
			}
		}
		color[node] = 2 // Black
		return false
	}

	// Check all nodes
	for _, node := range store.nodes {
		if color[node.ID] == 0 {
			path[node.ID] = ""
			if dfs(node.ID) {
				break
			}
		}
	}

	if hasCycle {
		// Build cycle path
		cycle := []string{cycleEnd}
		current := cycleStart
		for current != cycleEnd {
			cycle = append([]string{current}, cycle...)
			current = path[current]
		}
		return &CycleError{
			Cycle: cycle,
		}
	}

	return nil
}

// ============================================================================
// STORE INTERFACE
// ============================================================================

type GraphStore interface {
	SaveNodes(nodes ...*Node)
	SaveEdges(edges ...*Edge)
	GetNode(id string) (*Node, error)
	GetEdge(id string) (*Edge, error)
	ListNodes() ([]*Node, error)
	ListEdges() ([]*Edge, error)
}

type InMemoryGraphStore struct {
	nodes map[string]*Node
	edges map[string]*Edge
}

func NewGraphStore() *InMemoryGraphStore {
	return &InMemoryGraphStore{
		nodes: make(map[string]*Node),
		edges: make(map[string]*Edge),
	}
}

func (s *InMemoryGraphStore) SaveNodes(nodes ...*Node) {
	for _, node := range nodes {
		s.nodes[node.ID] = node
	}
}

func (s *InMemoryGraphStore) SaveEdges(edges ...*Edge) {
	for _, edge := range edges {
		s.edges[edge.ID] = edge
	}
}

func (s *InMemoryGraphStore) GetNode(id string) (*Node, error) {
	node, ok := s.nodes[id]
	if !ok {
		return nil, &NodeNotFoundError{ID: id}
	}
	return node, nil
}

func (s *InMemoryGraphStore) GetEdge(id string) (*Edge, error) {
	edge, ok := s.edges[id]
	if !ok {
		return nil, &EdgeNotFoundError{ID: id}
	}
	return edge, nil
}

func (s *InMemoryGraphStore) ListNodes() ([]*Node, error) {
	result := make([]*Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		result = append(result, node)
	}
	return result, nil
}

func (s *InMemoryGraphStore) ListEdges() ([]*Edge, error) {
	result := make([]*Edge, 0, len(s.edges))
	for _, edge := range s.edges {
		result = append(result, edge)
	}
	return result, nil
}

// ============================================================================
// ERRORS
// ============================================================================

type ReferentialIntegrityError struct {
	EdgeID    string
	SourceID  string
	TargetID  string
	MissingID string
}

func (e *ReferentialIntegrityError) Error() string {
	return fmt.Sprintf("referential integrity violation: edge %s (%s -> %s) references non-existent node %s",
		e.EdgeID, e.SourceID, e.TargetID, e.MissingID)
}

type DuplicateEdgeError struct {
	EdgeID   string
	SourceID string
	TargetID string
	Kind     EdgeKind
}

func (e *DuplicateEdgeError) Error() string {
	return fmt.Sprintf("duplicate edge: (%s, %s, %s) already exists (edge ID: %s)",
		e.SourceID, e.TargetID, e.Kind, e.EdgeID)
}

type ConfidenceRangeError struct {
	EdgeID     string
	Confidence float64
	Reason     string
}

func (e *ConfidenceRangeError) Error() string {
	return fmt.Sprintf("confidence range error: edge %s has confidence %f (%s)",
		e.EdgeID, e.Confidence, e.Reason)
}

type CycleError struct {
	Cycle []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("cycle detected: %v", e.Cycle)
}

type NodeNotFoundError struct {
	ID string
}

func (e *NodeNotFoundError) Error() string {
	return "node not found: " + e.ID
}

type EdgeNotFoundError struct {
	ID string
}

func (e *EdgeNotFoundError) Error() string {
	return "edge not found: " + e.ID
}