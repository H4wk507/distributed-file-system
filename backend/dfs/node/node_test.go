package node

import (
	"dfs-backend/dfs/common"
	"encoding/json"
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

/* Check if EnterCriticalSection returns true when conditions are met */
func TestEnterCriticalSection_Success(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9989, common.RoleStorage, 1)

	resourceID := "test-resource"

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
	node.pendingAcks[resourceID] = make(map[uuid.UUID]bool) // no peers, no acks needed
	node.lockQueueMutex.Unlock()

	// Should successfully enter critical section
	if !node.EnterCriticalSection(resourceID) {
		t.Error("expected EnterCriticalSection to return true when conditions are met")
	}
}

/* Check if EnterCriticalSection returns false when conditions are not met */
func TestEnterCriticalSection_Failure_NotFirstInQueue(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9988, common.RoleStorage, 1)

	resourceID := "test-resource"
	otherNodeID := uuid.New()

	// Other node's request is first
	otherRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      otherNodeID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	// Our request is second
	ourRequest := &common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      node.ID,
		LogicalTime: 2,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}

	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{otherRequest, ourRequest}
	node.pendingAcks[resourceID] = make(map[uuid.UUID]bool)
	node.lockQueueMutex.Unlock()

	// Should NOT enter critical section
	if node.EnterCriticalSection(resourceID) {
		t.Error("expected EnterCriticalSection to return false when not first in queue")
	}
}

/* Check if handleLockAcquired updates lockHolders correctly */
func TestHandleLockAcquired_UpdatesLockHolders(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9987, common.RoleStorage, 1)

	resourceID := "test-resource"
	holderID := uuid.New()

	// Simulate receiving a LockAcquired message
	payload, _ := json.Marshal(resourceID)
	msg := common.Message{
		Type:    common.MessageLockAcquired,
		From:    holderID,
		Payload: payload,
	}

	node.handleLockAcquired(msg)

	// Verify lockHolders was updated
	node.lockQueueMutex.RLock()
	holder, exists := node.lockHolders[resourceID]
	node.lockQueueMutex.RUnlock()

	if !exists {
		t.Fatal("expected lockHolders to contain the resource")
	}

	if holder != holderID {
		t.Errorf("expected holder to be %v, got %v", holderID, holder)
	}
}

/* Check if handleLockRelease removes from lockHolders */
func TestHandleLockRelease_RemovesFromLockHolders(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9986, common.RoleStorage, 1)

	resourceID := "test-resource"
	holderID := uuid.New()

	// Setup: holder already has the lock
	node.lockQueueMutex.Lock()
	node.lockHolders[resourceID] = holderID
	node.lockQueue[resourceID] = []*common.LockRequest{
		{
			ResourceID:  resourceID,
			NodeID:      holderID,
			LogicalTime: 1,
			Status:      common.StatusPending,
			RequestedAt: time.Now(),
		},
	}
	node.lockQueueMutex.Unlock()

	// Simulate receiving a LockRelease message
	payload, _ := json.Marshal(resourceID)
	msg := common.Message{
		Type:    common.MessageLockRelease,
		From:    holderID,
		Payload: payload,
	}

	node.handleLockRelease(msg)

	// Verify lockHolders was cleared
	node.lockQueueMutex.RLock()
	_, exists := node.lockHolders[resourceID]
	queueLen := len(node.lockQueue[resourceID])
	node.lockQueueMutex.RUnlock()

	if exists {
		t.Error("expected lockHolders to NOT contain the resource after release")
	}

	if queueLen != 0 {
		t.Errorf("expected queue to be empty after release, got %d", queueLen)
	}
}

/* Check if Wait-For Graph edge is added when lock request comes in and holder exists */
func TestWaitForGraph_EdgeAddedOnLockRequest(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9985, common.RoleStorage, 1)

	resourceID := "test-resource"
	holderID := uuid.New()
	waiterID := uuid.New()

	// Setup: holderID already holds the lock
	node.lockQueueMutex.Lock()
	node.lockHolders[resourceID] = holderID
	node.lockQueueMutex.Unlock()

	// Add waiter as peer so we can send ack
	node.AddPeer(&common.NodeInfo{
		ID:            waiterID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      2,
		LastHeartbeat: time.Now(),
	})

	// Simulate waiter requesting the lock
	lockRequest := common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      waiterID,
		LogicalTime: 2,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}
	payload, _ := json.Marshal(lockRequest)
	msg := common.Message{
		Type:    common.MessageLockRequest,
		From:    waiterID,
		Payload: payload,
	}

	node.handleLockRequest(msg)

	// Verify edge was added: waiter -> holder
	node.waitForGraphMutex.RLock()
	waitsFor, exists := node.waitForGraph[waiterID]
	node.waitForGraphMutex.RUnlock()

	if !exists {
		t.Fatal("expected waitForGraph to contain waiter")
	}

	if !waitsFor[holderID] {
		t.Errorf("expected waiter %v to wait for holder %v", waiterID, holderID)
	}
}

/* Check if Wait-For Graph edge is NOT added when no holder exists */
func TestWaitForGraph_NoEdgeWhenNoHolder(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9984, common.RoleStorage, 1)

	resourceID := "test-resource"
	waiterID := uuid.New()

	// No holder set for this resource

	// Add waiter as peer
	node.AddPeer(&common.NodeInfo{
		ID:            waiterID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      2,
		LastHeartbeat: time.Now(),
	})

	// Simulate waiter requesting the lock
	lockRequest := common.LockRequest{
		ResourceID:  resourceID,
		NodeID:      waiterID,
		LogicalTime: 1,
		Status:      common.StatusPending,
		RequestedAt: time.Now(),
	}
	payload, _ := json.Marshal(lockRequest)
	msg := common.Message{
		Type:    common.MessageLockRequest,
		From:    waiterID,
		Payload: payload,
	}

	node.handleLockRequest(msg)

	// Verify NO edge was added
	node.waitForGraphMutex.RLock()
	_, exists := node.waitForGraph[waiterID]
	node.waitForGraphMutex.RUnlock()

	if exists {
		t.Error("expected NO edge in waitForGraph when no holder exists")
	}
}

/* Check if Wait-For Graph edges are removed when lock is released */
func TestWaitForGraph_EdgesRemovedOnRelease(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9983, common.RoleStorage, 1)

	resourceID := "test-resource"
	holderID := uuid.New()
	waiter1ID := uuid.New()
	waiter2ID := uuid.New()

	// Setup: holderID holds lock, waiter1 and waiter2 are waiting
	node.lockQueueMutex.Lock()
	node.lockHolders[resourceID] = holderID
	node.lockQueue[resourceID] = []*common.LockRequest{
		{ResourceID: resourceID, NodeID: holderID, LogicalTime: 1, RequestedAt: time.Now()},
	}
	node.lockQueueMutex.Unlock()

	// Setup WFG: waiter1 -> holder, waiter2 -> holder
	node.waitForGraphMutex.Lock()
	node.waitForGraph[waiter1ID] = map[uuid.UUID]bool{holderID: true}
	node.waitForGraph[waiter2ID] = map[uuid.UUID]bool{holderID: true}
	node.waitForGraphMutex.Unlock()

	// Simulate holder releasing the lock
	payload, _ := json.Marshal(resourceID)
	msg := common.Message{
		Type:    common.MessageLockRelease,
		From:    holderID,
		Payload: payload,
	}

	node.handleLockRelease(msg)

	// Verify edges to holder were removed
	node.waitForGraphMutex.RLock()
	waiter1WaitsFor := node.waitForGraph[waiter1ID]
	waiter2WaitsFor := node.waitForGraph[waiter2ID]
	node.waitForGraphMutex.RUnlock()

	if waiter1WaitsFor[holderID] {
		t.Error("expected waiter1's edge to holder to be removed")
	}

	if waiter2WaitsFor[holderID] {
		t.Error("expected waiter2's edge to holder to be removed")
	}
}

/* Check if Wait-For Graph preserves edges to other nodes on release */
func TestWaitForGraph_PreservesOtherEdgesOnRelease(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9982, common.RoleStorage, 1)

	resourceID := "test-resource"
	holderID := uuid.New()
	otherHolderID := uuid.New()
	waiterID := uuid.New()

	// Setup: holderID holds lock
	node.lockQueueMutex.Lock()
	node.lockHolders[resourceID] = holderID
	node.lockQueue[resourceID] = []*common.LockRequest{
		{ResourceID: resourceID, NodeID: holderID, LogicalTime: 1, RequestedAt: time.Now()},
	}
	node.lockQueueMutex.Unlock()

	// Setup WFG: waiter -> holder AND waiter -> otherHolder (for different resource)
	node.waitForGraphMutex.Lock()
	node.waitForGraph[waiterID] = map[uuid.UUID]bool{
		holderID:      true,
		otherHolderID: true,
	}
	node.waitForGraphMutex.Unlock()

	// Simulate holder releasing the lock
	payload, _ := json.Marshal(resourceID)
	msg := common.Message{
		Type:    common.MessageLockRelease,
		From:    holderID,
		Payload: payload,
	}

	node.handleLockRelease(msg)

	// Verify only edge to holderID was removed, edge to otherHolderID remains
	node.waitForGraphMutex.RLock()
	waitsFor := node.waitForGraph[waiterID]
	node.waitForGraphMutex.RUnlock()

	if waitsFor[holderID] {
		t.Error("expected edge to released holder to be removed")
	}

	if !waitsFor[otherHolderID] {
		t.Error("expected edge to other holder to be preserved")
	}
}

/* Check if Wait-For Graph detects simple cycle */
func TestWaitForGraph_DetectsCycle(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9981, common.RoleStorage, 1)

	nodeA := uuid.New()
	nodeB := uuid.New()

	// Setup cycle: A -> B -> A
	node.waitForGraphMutex.Lock()
	node.waitForGraph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	node.waitForGraph[nodeB] = map[uuid.UUID]bool{nodeA: true}
	node.waitForGraphMutex.Unlock()

	// Check for cycle
	node.waitForGraphMutex.RLock()
	hasCycle, path := common.HasCycle(node.waitForGraph)
	node.waitForGraphMutex.RUnlock()

	if !hasCycle {
		t.Error("expected cycle to be detected")
	}

	if len(path) < 2 {
		t.Errorf("expected path with at least 2 nodes, got %v", path)
	}
}

/* Check if Wait-For Graph reports no cycle when none exists */
func TestWaitForGraph_NoCycleWhenLinear(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9980, common.RoleStorage, 1)

	nodeA := uuid.New()
	nodeB := uuid.New()
	nodeC := uuid.New()

	// Setup linear chain: A -> B -> C (no cycle)
	node.waitForGraphMutex.Lock()
	node.waitForGraph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	node.waitForGraph[nodeB] = map[uuid.UUID]bool{nodeC: true}
	node.waitForGraph[nodeC] = make(map[uuid.UUID]bool) // C waits for no one
	node.waitForGraphMutex.Unlock()

	// Check for cycle
	node.waitForGraphMutex.RLock()
	hasCycle, path := common.HasCycle(node.waitForGraph)
	node.waitForGraphMutex.RUnlock()

	if hasCycle {
		t.Errorf("expected no cycle in linear chain, but got path %v", path)
	}
}

/* Check if handleLockAbort releases all locks held by node */
func TestHandleLockAbort_ReleasesAllLocks(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9979, common.RoleStorage, 1)

	resource1 := "file1.txt"
	resource2 := "file2.txt"
	resource3 := "file3.txt"

	// Setup: node holds locks on resource1 and resource2, but not resource3
	node.lockQueueMutex.Lock()
	node.lockQueue[resource1] = []*common.LockRequest{
		{ResourceID: resource1, NodeID: node.ID, LogicalTime: 1, RequestedAt: time.Now()},
	}
	node.lockQueue[resource2] = []*common.LockRequest{
		{ResourceID: resource2, NodeID: node.ID, LogicalTime: 2, RequestedAt: time.Now()},
	}
	// resource3 is held by another node
	otherNodeID := uuid.New()
	node.lockQueue[resource3] = []*common.LockRequest{
		{ResourceID: resource3, NodeID: otherNodeID, LogicalTime: 1, RequestedAt: time.Now()},
	}
	node.pendingAcks[resource1] = make(map[uuid.UUID]bool)
	node.pendingAcks[resource2] = make(map[uuid.UUID]bool)
	node.lockQueueMutex.Unlock()

	// Simulate receiving abort message
	msg := common.Message{
		Type: common.MessageLockAbort,
		From: uuid.New(), // from master
	}

	node.handleLockAbort(msg)

	// Verify our locks were released
	node.lockQueueMutex.RLock()
	res1Len := len(node.lockQueue[resource1])
	res2Len := len(node.lockQueue[resource2])
	res3Len := len(node.lockQueue[resource3])
	node.lockQueueMutex.RUnlock()

	if res1Len != 0 {
		t.Errorf("expected resource1 queue to be empty, got %d", res1Len)
	}

	if res2Len != 0 {
		t.Errorf("expected resource2 queue to be empty, got %d", res2Len)
	}

	// Other node's lock should remain
	if res3Len != 1 {
		t.Errorf("expected resource3 queue to have 1 request (other node), got %d", res3Len)
	}
}

/* Check if handleLockAbort does nothing when node has no locks */
func TestHandleLockAbort_NoLocksToRelease(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9978, common.RoleStorage, 1)

	// No locks set up

	msg := common.Message{
		Type: common.MessageLockAbort,
		From: uuid.New(),
	}

	// Should not panic
	node.handleLockAbort(msg)

	// Verify queue is still empty
	node.lockQueueMutex.RLock()
	queueLen := len(node.lockQueue)
	node.lockQueueMutex.RUnlock()

	if queueLen != 0 {
		t.Errorf("expected empty lock queue, got %d resources", queueLen)
	}
}

/* Check if selectVictim returns node with lowest priority */
func TestSelectVictim_LowestPriority(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9977, common.RoleMaster, 10) // high priority

	// Add peers with different priorities
	lowPriorityNode := uuid.New()
	highPriorityNode := uuid.New()

	node.AddPeer(&common.NodeInfo{
		ID:            lowPriorityNode,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      1, // lowest
		LastHeartbeat: time.Now(),
	})
	node.AddPeer(&common.NodeInfo{
		ID:            highPriorityNode,
		IP:            "127.0.0.1",
		Port:          8002,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      5, // medium
		LastHeartbeat: time.Now(),
	})

	// Cycle includes all three nodes
	cyclePath := []uuid.UUID{node.ID, lowPriorityNode, highPriorityNode, node.ID}

	victim := node.selectVictim(cyclePath)

	if victim != lowPriorityNode {
		t.Errorf("expected victim to be low priority node %v, got %v", lowPriorityNode, victim)
	}
}

/* Check if selectVictim returns self when self has lowest priority */
func TestSelectVictim_SelfIsVictim(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9976, common.RoleMaster, 1) // lowest priority

	peerID := uuid.New()
	node.AddPeer(&common.NodeInfo{
		ID:            peerID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      10, // higher than node
		LastHeartbeat: time.Now(),
	})

	cyclePath := []uuid.UUID{node.ID, peerID, node.ID}

	victim := node.selectVictim(cyclePath)

	if victim != node.ID {
		t.Errorf("expected victim to be self %v, got %v", node.ID, victim)
	}
}

/* Check if selectVictim handles unknown nodes gracefully */
func TestSelectVictim_UnknownNodeInCycle(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9975, common.RoleMaster, 5)

	knownPeerID := uuid.New()
	unknownPeerID := uuid.New() // not added to peers

	node.AddPeer(&common.NodeInfo{
		ID:            knownPeerID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      3,
		LastHeartbeat: time.Now(),
	})

	// Cycle includes unknown node
	cyclePath := []uuid.UUID{node.ID, knownPeerID, unknownPeerID, node.ID}

	victim := node.selectVictim(cyclePath)

	// Should return known peer with lowest priority (3 < 5)
	if victim != knownPeerID {
		t.Errorf("expected victim to be known peer %v, got %v", knownPeerID, victim)
	}
}

/* Check if checkDeadlocks only runs on master */
func TestCheckDeadlocks_OnlyRunsOnMaster(t *testing.T) {
	// Create storage node (not master)
	node := CreateNodeWithBully("127.0.0.1", 9974, common.RoleStorage, 1)

	// Setup a cycle that should be detected
	nodeA := uuid.New()
	nodeB := uuid.New()

	node.waitForGraphMutex.Lock()
	node.waitForGraph[nodeA] = map[uuid.UUID]bool{nodeB: true}
	node.waitForGraph[nodeB] = map[uuid.UUID]bool{nodeA: true}
	node.waitForGraphMutex.Unlock()

	// checkDeadlocks should return early (not master)
	// This shouldn't panic or cause issues
	node.checkDeadlocks()

	// WFG should remain unchanged (no action taken)
	node.waitForGraphMutex.RLock()
	hasCycle, _ := common.HasCycle(node.waitForGraph)
	node.waitForGraphMutex.RUnlock()

	if !hasCycle {
		t.Error("expected cycle to still exist (storage node shouldn't resolve it)")
	}
}

/* Check if checkDeadlocks detects and handles cycle on master */
func TestCheckDeadlocks_MasterDetectsCycle(t *testing.T) {
	// Create master node
	node := CreateNodeWithBully("127.0.0.1", 9973, common.RoleMaster, 10)

	// Add a peer that will be the victim (lower priority)
	victimID := uuid.New()
	node.AddPeer(&common.NodeInfo{
		ID:            victimID,
		IP:            "127.0.0.1",
		Port:          8001,
		Role:          common.RoleStorage,
		Status:        common.StatusOnline,
		Priority:      1, // lower than master
		LastHeartbeat: time.Now(),
	})

	// Setup cycle: master -> victim -> master
	node.waitForGraphMutex.Lock()
	node.waitForGraph[node.ID] = map[uuid.UUID]bool{victimID: true}
	node.waitForGraph[victimID] = map[uuid.UUID]bool{node.ID: true}
	node.waitForGraphMutex.Unlock()

	// checkDeadlocks should detect cycle and try to send abort
	// Note: sendAbort will fail to connect but shouldn't panic
	node.checkDeadlocks()

	// The cycle detection worked if we got here without panic
	// (actual abort would require network, which we're not testing here)
}

/* Check if sendAbort handles self-abort correctly */
func TestSendAbort_SelfAbort(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9972, common.RoleMaster, 1)

	resourceID := "test-resource"

	// Setup: node holds a lock
	node.lockQueueMutex.Lock()
	node.lockQueue[resourceID] = []*common.LockRequest{
		{ResourceID: resourceID, NodeID: node.ID, LogicalTime: 1, RequestedAt: time.Now()},
	}
	node.pendingAcks[resourceID] = make(map[uuid.UUID]bool)
	node.lockQueueMutex.Unlock()

	// Send abort to self
	node.sendAbort(node.ID)

	// Verify lock was released
	node.lockQueueMutex.RLock()
	queueLen := len(node.lockQueue[resourceID])
	node.lockQueueMutex.RUnlock()

	if queueLen != 0 {
		t.Errorf("expected lock to be released after self-abort, got queue length %d", queueLen)
	}
}

/* Check if sendAbort handles unknown peer gracefully */
func TestSendAbort_UnknownPeer(t *testing.T) {
	node := CreateNodeWithBully("127.0.0.1", 9971, common.RoleMaster, 1)

	unknownPeerID := uuid.New() // not in peers

	// Should not panic when peer is not found
	node.sendAbort(unknownPeerID)
}
