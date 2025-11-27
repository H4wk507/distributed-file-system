package hashing

import (
	"dfs-backend/dfs/common"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"

	"github.com/google/uuid"
)

const DEFAULT_VIRTUAL_NODES = 150

type HashRing struct {
	ring         []uint32                      // posortowane pozycje na ringu
	nodeMap      map[uint32]uuid.UUID          // key = pozycja na ringu, value = node id
	nodes        map[uuid.UUID]common.NodeInfo // key = node id, value = node info
	virtualNodes int
	mutex        sync.RWMutex
}

func NewHashRing(virtualNodes int) *HashRing {
	if virtualNodes <= 0 {
		virtualNodes = DEFAULT_VIRTUAL_NODES
	}

	return &HashRing{
		ring:         make([]uint32, 0),
		nodeMap:      make(map[uint32]uuid.UUID),
		nodes:        make(map[uuid.UUID]common.NodeInfo),
		virtualNodes: virtualNodes,
	}
}

func HashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func generateVirtualNodeKey(nodeID uuid.UUID, i int) string {
	return fmt.Sprintf("%s#%d", nodeID.String(), i)
}

func (h *HashRing) AddNode(node common.NodeInfo) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.nodes[node.ID]; exists {
		return
	}

	h.nodes[node.ID] = node

	for i := 0; i < h.virtualNodes; i++ {
		key := generateVirtualNodeKey(node.ID, i)
		hash := HashKey(key)
		h.ring = append(h.ring, hash)
		h.nodeMap[hash] = node.ID
	}

	sort.SliceStable(h.ring, func(i, j int) bool {
		return h.ring[i] < h.ring[j]
	})
}

func (h *HashRing) RemoveNode(nodeID uuid.UUID) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.nodes[nodeID]; !exists {
		return
	}

	delete(h.nodes, nodeID)

	var newRing []uint32
	for _, hash := range h.ring {
		if h.nodeMap[hash] == nodeID {
			delete(h.nodeMap, hash)
		} else {
			newRing = append(newRing, hash)
		}
	}
	h.ring = newRing
}

func (h *HashRing) FindNodesForFile(filename string, n int) []common.NodeInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if len(h.ring) == 0 {
		return nil
	}

	hash := HashKey(filename)

	idx := sort.Search(len(h.ring), func(i int) bool {
		return h.ring[i] >= hash
	})

	if idx == len(h.ring) {
		idx = 0
	}

	seen := make(map[uuid.UUID]bool)
	result := make([]common.NodeInfo, 0, n)

	for i := 0; i < len(h.ring) && len(result) < n; i++ {
		pos := (idx + i) % len(h.ring)
		nodeID := h.nodeMap[h.ring[pos]]

		if !seen[nodeID] {
			seen[nodeID] = true
			if node, exists := h.nodes[nodeID]; exists {
				result = append(result, node)
			}
		}
	}

	return result
}

func (h *HashRing) GetNodeCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.nodes)
}

func (h *HashRing) GetAllNodes() []common.NodeInfo {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	nodes := make([]common.NodeInfo, 0, len(h.nodes))
	for _, node := range h.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}
