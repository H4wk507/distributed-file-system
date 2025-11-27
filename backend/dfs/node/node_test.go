package node

import (
	"dfs-backend/dfs/common"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

/* Check if lock request is removed from queue after timeout */
func TestLockTimeout(t *testing.T) {

	// Create a node without starting it (no network operations needed)
	node := CreateNodeWithBully("127.0.0.1", 9999, common.RoleStorage, 1)

	resourceID := "test-resource"

	// Add a lock request that is already expired (requested 31 seconds ago)
	expiredRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now().Add(-31 * time.Second), // older than LOCK_TIMEOUT (30s)
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = append(node.lockQueue[resourceID], expiredRequest)
	node.pendingAcks[resourceID] = make(map[uuid.UUID]bool)
	node.lockQueueMutex.Unlock()

	// Verify lock request is in queue before timeout check
	node.lockQueueMutex.RLock()
	beforeLen := len(node.lockQueue[resourceID])
	node.lockQueueMutex.RUnlock()

	if beforeLen != 1 {
		t.Fatalf("expected 1 lock request before timeout check, got %d", beforeLen)
	}

	// Run the timeout check
	node.checkLocksTimeouts()

	// Verify lock request was removed due to timeout
	node.lockQueueMutex.RLock()
	afterLen := len(node.lockQueue[resourceID])
	node.lockQueueMutex.RUnlock()

	if afterLen != 0 {
		t.Errorf("expected 0 lock requests after timeout check, got %d", afterLen)
	}
}

/* Check if lock request is not removed from queue if it is not expired */
func TestLockNotTimedOut(t *testing.T) {
	// Create a node without starting it
	node := CreateNodeWithBully("127.0.0.1", 9998, common.RoleStorage, 1)

	resourceID := "test-resource"

	// Add a fresh lock request (not expired)
	freshRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(), // just created, not expired
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = append(node.lockQueue[resourceID], freshRequest)
	node.pendingAcks[resourceID] = make(map[uuid.UUID]bool)
	node.lockQueueMutex.Unlock()

	// Run the timeout check
	node.checkLocksTimeouts()

	// Verify lock request is still in queue (not timed out)
	node.lockQueueMutex.RLock()
	afterLen := len(node.lockQueue[resourceID])
	node.lockQueueMutex.RUnlock()

	if afterLen != 1 {
		t.Errorf("expected 1 lock request to remain (not timed out), got %d", afterLen)
	}
}

/* Check if lock request is removed from queue when node dies */
func TestLockReleasedWhenNodeDies(t *testing.T) {
	// Create our local node
	node := CreateNodeWithBully("127.0.0.1", 9997, common.RoleStorage, 1)

	resourceID := "shared-resource"

	// Simulate a dead peer that was holding a lock
	deadPeerID := uuid.New()
	deadPeerInfo := &common.NodeInfo{
		ID:            deadPeerID,
		IP:            "127.0.0.1",
		Port:          8888,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      2,
		LastHeartbeat: time.Now().Add(-20 * time.Second), // last heartbeat 20s ago (will be marked dead)
	}

	// Add the peer to our node's peer list
	node.AddPeer(deadPeerInfo)

	// The dead peer had acquired a lock before dying (request is now expired)
	deadPeerLockRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      deadPeerID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now().Add(-31 * time.Second), // older than LOCK_TIMEOUT
	}

	// Our node also wants the lock (fresh request)
	ourLockRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 2,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	// Add both requests to the queue (dead peer's request is first due to lower logical time)
	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{deadPeerLockRequest, ourLockRequest}
	node.pendingAcks[resourceID] = make(map[uuid.UUID]bool)
	node.lockQueueMutex.Unlock()

	// Verify dead peer's lock is first in queue
	node.lockQueueMutex.RLock()
	if len(node.lockQueue[resourceID]) != 2 {
		t.Fatalf("expected 2 lock requests, got %d", len(node.lockQueue[resourceID]))
	}
	firstInQueue := node.lockQueue[resourceID][0]
	node.lockQueueMutex.RUnlock()

	if firstInQueue.NodeID != deadPeerID {
		t.Fatalf("expected dead peer to be first in queue")
	}

	// Run the timeout check - this should remove the dead peer's expired lock
	node.checkLocksTimeouts()

	// Verify dead peer's lock was removed, only our lock remains
	node.lockQueueMutex.RLock()
	remainingRequests := len(node.lockQueue[resourceID])
	node.lockQueueMutex.RUnlock()

	if remainingRequests != 1 {
		t.Errorf("expected 1 lock request after dead peer's lock timed out, got %d", remainingRequests)
	}

	// Verify our request is now first in queue
	node.lockQueueMutex.RLock()
	if len(node.lockQueue[resourceID]) > 0 {
		firstInQueue = node.lockQueue[resourceID][0]
		if firstInQueue.NodeID != node.ID {
			t.Errorf("expected our node to be first in queue after dead peer's lock expired")
		}
	}
	node.lockQueueMutex.RUnlock()
}

/* Check if lock requests are sorted by logical time */
func TestLockQueueOrdering(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9996, common.RoleStorage, 1)

	resourceID := "test-resource"

	// Create requests with different logical times (out of order)
	node1ID := uuid.New()
	node2ID := uuid.New()
	node3ID := uuid.New()

	request1 := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node1ID,
		LogicalTime: 5, // middle
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}
	request2 := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node2ID,
		LogicalTime: 2, // lowest - should be first
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}
	request3 := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node3ID,
		LogicalTime: 8, // highest - should be last
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	// Add requests out of order and sort them (simulating handleLockRequest behavior)
	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{request1, request2, request3}
	queue := node.lockQueue[resourceID]
	// Sort using the same logic as in handleLockRequest
	sort.SliceStable(queue, func(i, j int) bool {
		return common.CompareLockRequests(queue[i], queue[j])
	})
	node.lockQueueMutex.Unlock()

	// Verify ordering: request2 (time=2), request1 (time=5), request3 (time=8)
	node.lockQueueMutex.RLock()
	defer node.lockQueueMutex.RUnlock()

	if len(node.lockQueue[resourceID]) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(node.lockQueue[resourceID]))
	}

	if node.lockQueue[resourceID][0].LogicalTime != 2 {
		t.Errorf("expected first request to have logical time 2, got %d", node.lockQueue[resourceID][0].LogicalTime)
	}
	if node.lockQueue[resourceID][1].LogicalTime != 5 {
		t.Errorf("expected second request to have logical time 5, got %d", node.lockQueue[resourceID][1].LogicalTime)
	}
	if node.lockQueue[resourceID][2].LogicalTime != 8 {
		t.Errorf("expected third request to have logical time 8, got %d", node.lockQueue[resourceID][2].LogicalTime)
	}
}

/* Check if node can enter critical section when it is first in queue with all acks */
func TestCanEnterCriticalSection_FirstInQueueWithAllAcks(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9995, common.RoleStorage, 1)

	resourceID := "test-resource"

	// Add two peers
	peer1ID := uuid.New()
	peer2ID := uuid.New()

	node.AddPeer(&common.NodeInfo{
		ID:            peer1ID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      2,
		LastHeartbeat: time.Now(),
	})
	node.AddPeer(&common.NodeInfo{
		ID:            peer2ID,
		IP:            "127.0.0.1",
		Port:          8002,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      3,
		LastHeartbeat: time.Now(),
	})

	// Our node's lock request (first in queue with lowest logical time)
	ourRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{ourRequest}
	// Simulate receiving acks from all peers
	node.pendingAcks[resourceID] = map[uuid.UUID]bool{
		peer1ID: true,
		peer2ID: true,
	}
	node.lockQueueMutex.Unlock()

	// Should be able to enter critical section
	if !node.CanEnterCriticalSection(resourceID) {
		t.Error("expected to be able to enter critical section when first in queue with all acks")
	}
}

/* Check if node cannot enter critical section when it is not first in queue */
func TestCanEnterCriticalSection_NotFirstInQueue(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9994, common.RoleStorage, 1)

	resourceID := "test-resource"

	otherNodeID := uuid.New()

	// Add peer
	node.AddPeer(&common.NodeInfo{
		ID:            otherNodeID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      2,
		LastHeartbeat: time.Now(),
	})

	// Other node's request is first (lower logical time)
	otherRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      otherNodeID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	// Our request is second (higher logical time)
	ourRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 2,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{otherRequest, ourRequest}
	node.pendingAcks[resourceID] = map[uuid.UUID]bool{
		otherNodeID: true, // even with ack, we can't enter because we're not first
	}
	node.lockQueueMutex.Unlock()

	// Should NOT be able to enter critical section
	if node.CanEnterCriticalSection(resourceID) {
		t.Error("expected NOT to be able to enter critical section when not first in queue")
	}
}

/* Check if node cannot enter critical section when it is missing acks */
func TestCanEnterCriticalSection_MissingAcks(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9993, common.RoleStorage, 1)

	resourceID := "test-resource"

	// Add two peers
	peer1ID := uuid.New()
	peer2ID := uuid.New()

	node.AddPeer(&common.NodeInfo{
		ID:            peer1ID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      2,
		LastHeartbeat: time.Now(),
	})
	node.AddPeer(&common.NodeInfo{
		ID:            peer2ID,
		IP:            "127.0.0.1",
		Port:          8002,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      3,
		LastHeartbeat: time.Now(),
	})

	// Our node's lock request (first in queue)
	ourRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{ourRequest}
	// Only one ack received, missing one
	node.pendingAcks[resourceID] = map[uuid.UUID]bool{
		peer1ID: true,
		// peer2ID ack is missing
	}
	node.lockQueueMutex.Unlock()

	// Should NOT be able to enter critical section (missing ack from peer2)
	if node.CanEnterCriticalSection(resourceID) {
		t.Error("expected NOT to be able to enter critical section when missing acks")
	}
}

/* Check if node can enter critical section when no peers exist */
func TestCanEnterCriticalSection_NoPeers(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9992, common.RoleStorage, 1)

	resourceID := "test-resource"

	// No peers added - node is alone in the cluster

	// Our node's lock request
	ourRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{ourRequest}
	node.pendingAcks[resourceID] = make(map[uuid.UUID]bool) // no acks needed
	node.lockQueueMutex.Unlock()

	// Should be able to enter critical section immediately (no peers to wait for)
	if !node.CanEnterCriticalSection(resourceID) {
		t.Error("expected to be able to enter critical section when no peers exist")
	}
}

/* Check if lock request is removed from queue when lock is released */
func TestReleaseLock_RemovesFromQueue(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9991, common.RoleStorage, 1)

	resourceID := "test-resource"

	// Add our lock request to the queue
	ourRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{ourRequest}
	node.pendingAcks[resourceID] = map[uuid.UUID]bool{}
	node.lockQueueMutex.Unlock()

	// Verify request is in queue
	node.lockQueueMutex.RLock()
	beforeLen := len(node.lockQueue[resourceID])
	node.lockQueueMutex.RUnlock()

	if beforeLen != 1 {
		t.Fatalf("expected 1 request before release, got %d", beforeLen)
	}

	// Release the lock (note: this will try to broadcast, but no peers exist so it's safe)
	node.ReleaseLock(resourceID)

	// Verify request was removed from queue
	node.lockQueueMutex.RLock()
	afterLen := len(node.lockQueue[resourceID])
	_, pendingAcksExist := node.pendingAcks[resourceID]
	node.lockQueueMutex.RUnlock()

	if afterLen != 0 {
		t.Errorf("expected 0 requests after release, got %d", afterLen)
	}

	if pendingAcksExist {
		t.Error("expected pendingAcks to be cleaned up after release")
	}
}

/* Check if node can enter critical section for multiple resources */
func TestMultipleResourceLocks(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9990, common.RoleStorage, 1)

	resource1 := "file1.txt"
	resource2 := "file2.txt"

	otherNodeID := uuid.New()

	// Other node holds lock on resource1
	otherRequest := &common.LockRequest{
		ResourceID:  resource1,
		NodeID:      otherNodeID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	// Our node holds lock on resource2
	ourRequest := &common.LockRequest{
		ResourceID:  resource2,
		NodeID:      node.ID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resource1] = []*common.LockRequest{otherRequest}
	node.lockQueue[resource2] = []*common.LockRequest{ourRequest}
	node.pendingAcks[resource1] = make(map[uuid.UUID]bool)
	node.pendingAcks[resource2] = make(map[uuid.UUID]bool)
	node.lockQueueMutex.Unlock()

	// We should NOT be able to enter CS for resource1 (other node is first)
	if node.CanEnterCriticalSection(resource1) {
		t.Error("expected NOT to be able to enter CS for resource1 (other node holds it)")
	}

	// We SHOULD be able to enter CS for resource2 (we are first, no peers)
	if !node.CanEnterCriticalSection(resource2) {
		t.Error("expected to be able to enter CS for resource2 (we hold it)")
	}
}
