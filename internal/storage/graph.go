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
	Kind      string
	Name      string
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

// ValidateConfidenceRange returns true if confidence is in [0, 1]
func (e *Edge) ValidateConfidenceRange() bool {
	return e.Confidence >= 0 && e.Confidence <= 1
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

// UpsertNodes is an alias for SaveNodes
func (s *InMemoryGraphStore) UpsertNodes(nodes []Node) {
	for _, node := range nodes {
		s.nodes[node.ID] = &node
	}
}

// UpsertEdges is an alias for SaveEdges - stores all edges, including potential duplicates
func (s *InMemoryGraphStore) UpsertEdges(edges []Edge) {
	for _, edge := range edges {
		// Store with auto-generated ID if empty
		edgeID := edge.ID
		if edgeID == "" {
			edgeID = fmt.Sprintf("auto:%s:%s:%s:%d", edge.SourceID, edge.TargetID, edge.Kind, len(s.edges))
		}
		edgeCopy := edge
		edgeCopy.ID = edgeID
		s.edges[edgeID] = &edgeCopy
	}
}

// ValidateNoDuplicateEdgesGlobal checks all edges for duplicates
func (s *InMemoryGraphStore) ValidateNoDuplicateEdgesGlobal() bool {
	seen := make(map[string]bool)
	for _, edge := range s.edges {
		key := edge.SourceID + ":" + edge.TargetID + ":" + string(edge.Kind)
		if seen[key] {
			return false // Duplicate found
		}
		seen[key] = true
	}
	return true
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
// ValidateReferentialIntegrity returns true if edge endpoints exist
func (s *InMemoryGraphStore) ValidateReferentialIntegrity(edge Edge) bool {
	_, sourceOK := s.nodes[edge.SourceID]
	_, targetOK := s.nodes[edge.TargetID]
	return sourceOK && targetOK
}

// ValidateNoDuplicateEdges returns true if no duplicate (source, target, kind) exists
func (s *InMemoryGraphStore) ValidateNoDuplicateEdges(edge Edge) bool {
	for _, existing := range s.edges {
		if existing.SourceID == edge.SourceID &&
			existing.TargetID == edge.TargetID &&
			existing.Kind == edge.Kind {
			// Found duplicate (same source, target, kind)
			// Only allow if it's the exact same edge ID
			if edge.ID != "" && existing.ID == edge.ID {
				// It's the same edge being re-validated
				return true
			}
			// Either no ID set (meaning we're checking a new edge) or different IDs
			// In both cases, we have a duplicate
			return false
		}
	}
	return true
}

// ValidateNoCycles returns true if no cycles exist for the given edge kind
func (s *InMemoryGraphStore) ValidateNoCycles(kindStr string) bool {
	// Build adjacency list for this kind only
	adj := make(map[string][]string)
	for _, edge := range s.edges {
		edgeKindStr := string(edge.Kind)
		// Map test strings to lowercase constant values
		if (kindStr == "CONTAINS" && edgeKindStr == "contains") ||
			(kindStr == "DEPENDS_ON" && edgeKindStr == "depends_on") ||
			kindStr == edgeKindStr {
			adj[edge.SourceID] = append(adj[edge.SourceID], edge.TargetID)
		}
	}

	// Detect cycles using DFS with coloring
	color := make(map[string]int)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		color[node] = 1 // Gray (in progress)
		for _, neighbor := range adj[node] {
			if color[neighbor] == 0 {
				if dfs(neighbor) {
					return true // Cycle detected in subtree
				}
			} else if color[neighbor] == 1 {
				return true // Back edge found - cycle!
			}
		}
		color[node] = 2 // Black (done)
		return false // No cycle from this node
	}

	// Check all nodes
	for nodeID := range s.nodes {
		if color[nodeID] == 0 {
			if dfs(nodeID) {
				return false // Cycle detected
			}
		}
	}

	return true // No cycles found
}

// NewMemoryGraphStore is an alias for NewGraphStore
func NewMemoryGraphStore() *InMemoryGraphStore {
	return NewGraphStore()
}
