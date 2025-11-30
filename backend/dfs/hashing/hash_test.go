package hashing

import (
	"dfs-backend/dfs/common"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func createTestNode(priority int) common.NodeInfo {
	return common.NodeInfo{
		ID:       uuid.New(),
		IP:       "127.0.0.1",
		Port:     8000 + priority,
		Role:     common.RoleStorage,
		Status:   common.StatusOnline,
		Priority: priority,
	}
}

func TestHashKey_Deterministic(t *testing.T) {
	key := "test_file.txt"

	hash1 := HashKey(key)
	hash2 := HashKey(key)

	if hash1 != hash2 {
		t.Errorf("HashKey should be deterministic: got %d and %d", hash1, hash2)
	}
}

func TestHashKey_DifferentInputs(t *testing.T) {
	hash1 := HashKey("file_a.txt")
	hash2 := HashKey("file_b.txt")

	if hash1 == hash2 {
		t.Error("Different inputs should (usually) produce different hashes")
	}
}

func TestNewHashRing_DefaultVirtualNodes(t *testing.T) {
	hr := NewHashRing(0)

	if hr.virtualNodes != DEFAULT_VIRTUAL_NODES {
		t.Errorf("expected default virtual nodes %d, got %d", DEFAULT_VIRTUAL_NODES, hr.virtualNodes)
	}
}

func TestNewHashRing_CustomVirtualNodes(t *testing.T) {
	hr := NewHashRing(50)

	if hr.virtualNodes != 50 {
		t.Errorf("expected 50 virtual nodes, got %d", hr.virtualNodes)
	}
}

func TestHashRing_AddNode(t *testing.T) {
	hr := NewHashRing(100)
	node := createTestNode(1)

	hr.AddNode(node)

	if hr.GetNodeCount() != 1 {
		t.Errorf("expected 1 node, got %d", hr.GetNodeCount())
	}

	// Sprawdź że ring ma 100 pozycji (virtual nodes)
	if len(hr.ring) != 100 {
		t.Errorf("expected 100 ring positions, got %d", len(hr.ring))
	}
}

func TestHashRing_AddNode_Duplicate(t *testing.T) {
	hr := NewHashRing(100)
	node := createTestNode(1)

	hr.AddNode(node)
	hr.AddNode(node) // duplikat

	if hr.GetNodeCount() != 1 {
		t.Errorf("duplicate node should not be added, got %d nodes", hr.GetNodeCount())
	}

	if len(hr.ring) != 100 {
		t.Errorf("expected 100 ring positions after duplicate add, got %d", len(hr.ring))
	}
}

func TestHashRing_AddMultipleNodes(t *testing.T) {
	hr := NewHashRing(100)

	for i := 0; i < 5; i++ {
		hr.AddNode(createTestNode(i))
	}

	if hr.GetNodeCount() != 5 {
		t.Errorf("expected 5 nodes, got %d", hr.GetNodeCount())
	}

	// 5 węzłów * 100 virtual nodes = 500 pozycji
	if len(hr.ring) != 500 {
		t.Errorf("expected 500 ring positions, got %d", len(hr.ring))
	}
}

func TestHashRing_RemoveNode(t *testing.T) {
	hr := NewHashRing(100)

	node1 := createTestNode(1)
	node2 := createTestNode(2)

	hr.AddNode(node1)
	hr.AddNode(node2)
	hr.RemoveNode(node1.ID)

	if hr.GetNodeCount() != 1 {
		t.Errorf("expected 1 node after removal, got %d", hr.GetNodeCount())
	}

	if len(hr.ring) != 100 {
		t.Errorf("expected 100 ring positions after removal, got %d", len(hr.ring))
	}
}

func TestHashRing_RemoveNonExistent(t *testing.T) {
	hr := NewHashRing(100)
	node := createTestNode(1)

	hr.AddNode(node)
	hr.RemoveNode(uuid.New()) // nieistniejący węzeł

	if hr.GetNodeCount() != 1 {
		t.Errorf("removing non-existent node should not affect count, got %d", hr.GetNodeCount())
	}
}

func TestHashRing_FindNodesForFile_EmptyRing(t *testing.T) {
	hr := NewHashRing(100)

	nodes := hr.FindNodesForFile("test.txt", 3)

	if nodes != nil {
		t.Errorf("expected nil for empty ring, got %v", nodes)
	}
}

func TestHashRing_FindNodesForFile_SingleNode(t *testing.T) {
	hr := NewHashRing(100)
	node := createTestNode(1)

	hr.AddNode(node)

	nodes := hr.FindNodesForFile("test.txt", 3)

	if len(nodes) != 1 {
		t.Errorf("expected 1 node (only one exists), got %d", len(nodes))
	}

	if nodes[0].ID != node.ID {
		t.Error("returned node should match added node")
	}
}

func TestHashRing_FindNodesForFile_MultipleNodes(t *testing.T) {
	hr := NewHashRing(100)

	node1 := createTestNode(1)
	node2 := createTestNode(2)
	node3 := createTestNode(3)

	hr.AddNode(node1)
	hr.AddNode(node2)
	hr.AddNode(node3)

	nodes := hr.FindNodesForFile("testfile.txt", 2)

	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}

	// Sprawdź że węzły są unikalne
	if nodes[0].ID == nodes[1].ID {
		t.Error("returned nodes should be unique")
	}
}

func TestHashRing_FindNodesForFile_RequestMoreThanExists(t *testing.T) {
	hr := NewHashRing(100)

	hr.AddNode(createTestNode(1))
	hr.AddNode(createTestNode(2))

	nodes := hr.FindNodesForFile("test.txt", 10)

	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes (only 2 exist), got %d", len(nodes))
	}
}

func TestHashRing_Consistency(t *testing.T) {
	hr := NewHashRing(100)

	for i := 0; i < 5; i++ {
		hr.AddNode(createTestNode(i))
	}

	filename := "consistent_file.dat"

	// Wielokrotne zapytania powinny zwracać te same węzły
	nodes1 := hr.FindNodesForFile(filename, 3)
	nodes2 := hr.FindNodesForFile(filename, 3)

	if len(nodes1) != len(nodes2) {
		t.Fatal("consistent hashing should return same number of nodes")
	}

	for i := range nodes1 {
		if nodes1[i].ID != nodes2[i].ID {
			t.Errorf("consistent hashing should return same nodes for same file, position %d differs", i)
		}
	}
}

func TestHashRing_ConsistencyAfterNodeRemoval(t *testing.T) {
	hr := NewHashRing(100)

	node1 := createTestNode(1)
	node2 := createTestNode(2)
	node3 := createTestNode(3)

	hr.AddNode(node1)
	hr.AddNode(node2)
	hr.AddNode(node3)

	// Znajdź węzły dla pliku przed usunięciem
	nodesBefore := hr.FindNodesForFile("stable_file.txt", 1)
	targetNodeBefore := nodesBefore[0].ID

	// Usuń inny węzeł (nie ten który obsługuje plik)
	var nodeToRemove uuid.UUID
	for _, n := range []common.NodeInfo{node1, node2, node3} {
		if n.ID != targetNodeBefore {
			nodeToRemove = n.ID
			break
		}
	}

	hr.RemoveNode(nodeToRemove)

	// Plik powinien nadal być na tym samym węźle (o ile ten węzeł nie został usunięty)
	nodesAfter := hr.FindNodesForFile("stable_file.txt", 1)

	if nodesAfter[0].ID != targetNodeBefore {
		t.Error("file should remain on same node after removing different node")
	}
}

func TestHashRing_RingSorted(t *testing.T) {
	hr := NewHashRing(100)

	for i := 0; i < 10; i++ {
		hr.AddNode(createTestNode(i))
	}

	hr.mutex.RLock()
	defer hr.mutex.RUnlock()

	for i := 1; i < len(hr.ring); i++ {
		if hr.ring[i] < hr.ring[i-1] {
			t.Errorf("ring should be sorted, but position %d (%d) < position %d (%d)",
				i, hr.ring[i], i-1, hr.ring[i-1])
		}
	}
}

func TestHashRing_GetAllNodes(t *testing.T) {
	hr := NewHashRing(100)

	node1 := createTestNode(1)
	node2 := createTestNode(2)

	hr.AddNode(node1)
	hr.AddNode(node2)

	allNodes := hr.GetAllNodes()

	if len(allNodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(allNodes))
	}

	// Sprawdź że oba węzły są w wynikach
	foundNode1 := false
	foundNode2 := false
	for _, n := range allNodes {
		if n.ID == node1.ID {
			foundNode1 = true
		}
		if n.ID == node2.ID {
			foundNode2 = true
		}
	}

	if !foundNode1 || !foundNode2 {
		t.Error("GetAllNodes should return all added nodes")
	}
}

func TestHashRing_ConcurrentAccess(t *testing.T) {
	hr := NewHashRing(50)

	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hr.AddNode(createTestNode(idx))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			filename := fmt.Sprintf("file_%d.txt", idx)
			hr.FindNodesForFile(filename, 3)
		}(i)
	}

	wg.Wait()

	// Po zakończeniu powinniśmy mieć 10 węzłów
	if hr.GetNodeCount() != 10 {
		t.Errorf("expected 10 nodes after concurrent adds, got %d", hr.GetNodeCount())
	}
}

func TestHashRing_WrapAround(t *testing.T) {
	// Test że ring "owija się" - plik z dużym hashem trafia do pierwszego węzła
	hr := NewHashRing(1) // 1 virtual node dla prostoty

	node := createTestNode(1)
	hr.AddNode(node)

	// Powinien znaleźć węzeł niezależnie od wartości hasha
	nodes := hr.FindNodesForFile("any_file.txt", 1)

	if len(nodes) != 1 {
		t.Error("should find node even with wrap-around")
	}
}
