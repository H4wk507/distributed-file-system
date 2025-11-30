package common

import (
	"testing"

	"github.com/google/uuid"
)

// Helper function to check if a slice contains a specific UUID
func contains(slice []uuid.UUID, id uuid.UUID) bool {
	for _, v := range slice {
		if v == id {
			return true
		}
	}
	return false
}

// Helper function to verify cycle path is valid
// A valid cycle path has the last element appearing earlier in the path
// e.g., [A, B, C, D, B] - B appears at index 1 and at the end
func isValidCyclePath(path []uuid.UUID) bool {
	if len(path) < 2 {
		return false
	}
	lastNode := path[len(path)-1]
	// Check if last node appears earlier in the path (forming a cycle)
	for i := 0; i < len(path)-1; i++ {
		if path[i] == lastNode {
			return true
		}
	}
	return false
}

// Helper function to verify path follows graph edges
func pathFollowsEdges(path []uuid.UUID, graph map[uuid.UUID]map[uuid.UUID]bool) bool {
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		if !graph[from][to] {
			return false
		}
	}
	return true
}

/* Check if empty graph has no cycle */
func TestHasCycle_EmptyGraph(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)

	hasCycle, path := HasCycle(graph)

	if hasCycle {
		t.Error("expected no cycle in empty graph")
	}
	if path != nil {
		t.Errorf("expected nil path, got %v", path)
	}
}

/* Check if single node without edges has no cycle */
func TestHasCycle_SingleNodeNoEdges(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()
	graph[nodeA] = make(map[uuid.UUID]bool)

	hasCycle, path := HasCycle(graph)

	if hasCycle {
		t.Error("expected no cycle for single node without edges")
	}
	if path != nil {
		t.Errorf("expected nil path, got %v", path)
	}
}

/* Check if simple two-node cycle is detected: A -> B -> A */
func TestHasCycle_SimpleCycle(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()
	nodeB := uuid.New()

	graph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	graph[nodeB] = map[uuid.UUID]bool{nodeA: true}

	hasCycle, path := HasCycle(graph)

	if !hasCycle {
		t.Fatal("expected cycle to be detected for A -> B -> A")
	}

	// Path should be [A, B, A] or [B, A, B] depending on iteration order
	if len(path) != 3 {
		t.Errorf("expected path length 3 for simple cycle, got %d: %v", len(path), path)
	}

	// First and last should be the same (cycle closes)
	if !isValidCyclePath(path) {
		t.Errorf("expected path to form a valid cycle (first == last), got %v", path)
	}

	// Path should contain both nodes
	if !contains(path, nodeA) || !contains(path, nodeB) {
		t.Errorf("expected path to contain both A and B, got %v", path)
	}

	// Verify path follows actual edges in graph
	if !pathFollowsEdges(path, graph) {
		t.Errorf("path does not follow graph edges: %v", path)
	}
}

/* Check if three-node cycle is detected: A -> B -> C -> A */
func TestHasCycle_ThreeNodeCycle(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()

	graph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	graph[nodeB] = map[uuid.UUID]bool{nodeC: true}
	graph[nodeC] = map[uuid.UUID]bool{nodeA: true}

	hasCycle, path := HasCycle(graph)

	if !hasCycle {
		t.Fatal("expected cycle to be detected for A -> B -> C -> A")
	}

	// Path should be [A, B, C, A] or permutation depending on start node
	if len(path) != 4 {
		t.Errorf("expected path length 4 for three-node cycle, got %d: %v", len(path), path)
	}

	// First and last should be the same (cycle closes)
	if !isValidCyclePath(path) {
		t.Errorf("expected path to form a valid cycle (first == last), got %v", path)
	}

	// Path should contain all three nodes
	if !contains(path, nodeA) || !contains(path, nodeB) || !contains(path, nodeC) {
		t.Errorf("expected path to contain A, B, and C, got %v", path)
	}

	// Verify path follows actual edges in graph
	if !pathFollowsEdges(path, graph) {
		t.Errorf("path does not follow graph edges: %v", path)
	}
}

/* Check if linear chain has no cycle: A -> B -> C */
func TestHasCycle_LinearChainNoCycle(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()

	graph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	graph[nodeB] = map[uuid.UUID]bool{nodeC: true}
	graph[nodeC] = make(map[uuid.UUID]bool) // C has no outgoing edges

	hasCycle, path := HasCycle(graph)

	if hasCycle {
		t.Errorf("expected no cycle in linear chain A -> B -> C, got path %v", path)
	}
	if path != nil {
		t.Errorf("expected nil path for no cycle, got %v", path)
	}
}

/* Check if self-loop is detected: A -> A */
func TestHasCycle_SelfLoop(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()

	graph[nodeA] = map[uuid.UUID]bool{nodeA: true}

	hasCycle, path := HasCycle(graph)

	if !hasCycle {
		t.Fatal("expected cycle to be detected for self-loop A -> A")
	}

	// Path should be [A, A]
	if len(path) != 2 {
		t.Errorf("expected path length 2 for self-loop, got %d: %v", len(path), path)
	}

	// Both elements should be A
	if path[0] != nodeA || path[1] != nodeA {
		t.Errorf("expected path [A, A], got %v", path)
	}

	// Verify it's a valid cycle
	if !isValidCyclePath(path) {
		t.Errorf("expected path to form a valid cycle, got %v", path)
	}
}

/* Check if disconnected graph with one cycle is detected */
func TestHasCycle_DisconnectedGraphWithCycle(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)

	// Disconnected component 1: A -> B (no cycle)
	nodeA := uuid.New()
	nodeB := uuid.New()
	graph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	graph[nodeB] = make(map[uuid.UUID]bool)

	// Disconnected component 2: C -> D -> C (cycle)
	nodeC := uuid.New()
	nodeD := uuid.New()
	graph[nodeC] = map[uuid.UUID]bool{nodeD: true}
	graph[nodeD] = map[uuid.UUID]bool{nodeC: true}

	hasCycle, path := HasCycle(graph)

	if !hasCycle {
		t.Fatal("expected cycle to be detected in disconnected graph")
	}

	// Path should be [C, D, C] or [D, C, D]
	if len(path) != 3 {
		t.Errorf("expected path length 3, got %d: %v", len(path), path)
	}

	// First and last should be the same
	if !isValidCyclePath(path) {
		t.Errorf("expected path to form a valid cycle, got %v", path)
	}

	// Path should contain C and D (the cycle nodes), not A and B
	if !contains(path, nodeC) || !contains(path, nodeD) {
		t.Errorf("expected path to contain C and D, got %v", path)
	}

	// Path should NOT contain A or B (they're not in the cycle)
	if contains(path, nodeA) || contains(path, nodeB) {
		t.Errorf("expected path to NOT contain A or B (no cycle there), got %v", path)
	}

	// Verify path follows actual edges
	if !pathFollowsEdges(path, graph) {
		t.Errorf("path does not follow graph edges: %v", path)
	}
}

/* Check if diamond graph without cycle is handled correctly: A -> B, A -> C, B -> D, C -> D */
func TestHasCycle_DiamondNoCycle(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()
	nodeD := uuid.New()

	graph[nodeA] = map[uuid.UUID]bool{nodeB: true, nodeC: true}
	graph[nodeB] = map[uuid.UUID]bool{nodeD: true}
	graph[nodeC] = map[uuid.UUID]bool{nodeD: true}
	graph[nodeD] = make(map[uuid.UUID]bool)

	hasCycle, path := HasCycle(graph)

	if hasCycle {
		t.Errorf("expected no cycle in diamond graph, got path %v", path)
	}
	if path != nil {
		t.Errorf("expected nil path for no cycle, got %v", path)
	}
}

/* Check if complex graph with cycle is detected */
func TestHasCycle_ComplexGraphWithCycle(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()
	nodeD := uuid.New()
	nodeE := uuid.New()

	// A -> B -> C -> D -> B (cycle B -> C -> D -> B)
	// A -> E (no cycle branch)
	graph[nodeA] = map[uuid.UUID]bool{nodeB: true, nodeE: true}
	graph[nodeB] = map[uuid.UUID]bool{nodeC: true}
	graph[nodeC] = map[uuid.UUID]bool{nodeD: true}
	graph[nodeD] = map[uuid.UUID]bool{nodeB: true} // back edge creating cycle
	graph[nodeE] = make(map[uuid.UUID]bool)

	hasCycle, path := HasCycle(graph)

	if !hasCycle {
		t.Fatal("expected cycle to be detected in complex graph")
	}

	// The cycle is B -> C -> D -> B, so path could be [A, B, C, D, B] or similar
	// The last element should appear earlier in the path (cycle detected)
	if !isValidCyclePath(path) {
		t.Errorf("expected path to form a valid cycle (last node appears earlier), got %v", path)
	}

	// Path should contain B, C, D (the cycle nodes)
	if !contains(path, nodeB) || !contains(path, nodeC) || !contains(path, nodeD) {
		t.Errorf("expected path to contain B, C, D (cycle nodes), got %v", path)
	}

	// E should not be in the path (it's not part of any cycle)
	if contains(path, nodeE) {
		t.Errorf("expected path to NOT contain E, got %v", path)
	}

	// Verify path follows actual edges
	if !pathFollowsEdges(path, graph) {
		t.Errorf("path does not follow graph edges: %v", path)
	}
}

/* Check that path starts from the node where cycle was detected */
func TestHasCycle_PathStructure(t *testing.T) {
	graph := make(map[uuid.UUID]map[uuid.UUID]bool)
	nodeA := uuid.New()
	nodeB := uuid.New()

	graph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	graph[nodeB] = map[uuid.UUID]bool{nodeA: true}

	hasCycle, path := HasCycle(graph)

	if !hasCycle {
		t.Fatal("expected cycle to be detected")
	}

	// The returned path should show the complete cycle
	// Starting from some node X, following edges, and returning to X
	if len(path) < 2 {
		t.Fatalf("path too short: %v", path)
	}

	startNode := path[0]
	endNode := path[len(path)-1]

	if startNode != endNode {
		t.Errorf("cycle path should start and end with same node, got start=%v end=%v", startNode, endNode)
	}

	// Verify all transitions in path are valid edges
	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]
		if !graph[from][to] {
			t.Errorf("invalid edge in path: %v -> %v (index %d)", from, to, i)
		}
	}
}
