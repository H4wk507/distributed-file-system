package node

import (
	"context"
	"dfs-backend/dfs/common"
	. "dfs-backend/dfs/common"
	"dfs-backend/dfs/election"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	ID       uuid.UUID
	IP       string
	Port     int
	Role     NodeRole
	Status   NodeStatus
	Priority int

	elector election.Elector

	Peers     map[uuid.UUID]*NodeInfo
	peerMutex sync.RWMutex

	listener net.Listener

	messageChan chan common.Message
	stopChan    chan struct{}

	logicalTime      int
	logicalTimeMutex sync.Mutex

	lockQueue      map[string][]*LockRequest // key = resourceID, value = request queue
	pendingAcks    map[string]map[uuid.UUID]bool
	lockQueueMutex sync.RWMutex

	logger *log.Logger
}

func CreateNodeWithBully(ip string, port int, role NodeRole, priority int) *Node {
	n := &Node{
		ID:          uuid.New(),
		IP:          ip,
		Port:        port,
		Role:        role,
		Status:      StatusStarting,
		Priority:    priority,
		Peers:       make(map[uuid.UUID]*NodeInfo),
		messageChan: make(chan common.Message, 100),
		stopChan:    make(chan struct{}),
		logicalTime: 0,
		lockQueue:   make(map[string][]*LockRequest),
		pendingAcks: make(map[string]map[uuid.UUID]bool),
		logger:      log.New(log.Writer(), fmt.Sprintf("[Node %s] ", role), log.LstdFlags),
	}

	n.elector = election.NewBullyElector(n)

	return n
}

func (n *Node) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", n.IP, n.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	n.listener = listener
	n.Status = StatusOnline
	n.logger.Printf("Node started on %s (ID: %s, Priority: %d)", addr, n.ID, n.Priority)

	go n.acceptConnections(ctx)
	go n.handleMessages(ctx)
	go n.startHeartbeat(ctx)
	go n.monitorPeers(ctx)
	if err := n.elector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start elector: %w", err)
	}

	return nil
}

func (n *Node) Stop() error {
	n.Status = StatusStopping
	close(n.stopChan)

	if n.listener != nil {
		if err := n.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	n.Status = StatusOffline
	n.logger.Println("Node stopped")
	return nil
}

func (n *Node) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		default:
			conn, err := n.listener.Accept()
			if err != nil {
				select {
				case <-n.stopChan:
					return
				default:
					n.logger.Printf("Failed to accept connection: %v", err)
					continue
				}
			}
			go n.handleConnection(conn)
		}
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var msgWithTime MessageWithTime
	if err := decoder.Decode(&msgWithTime); err != nil {
		n.logger.Printf("Failed to decode message: %v", err)
		return
	}

	n.UpdateAndGetLogicalTime(msgWithTime.LogicalTime)

	if msgWithTime.Message.Type == MessageDiscovery {
		n.handleDiscoveryWithConnection(msgWithTime.Message, conn)
		return
	}

	select {
	case n.messageChan <- msgWithTime.Message:
	default:
		n.logger.Println("Message channel full, dropping message")
	}
}

func (n *Node) handleMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		case msg := <-n.messageChan:
			n.processMessage(ctx, msg)
		}
	}
}

func (n *Node) processMessage(ctx context.Context, msg common.Message) {
	n.logger.Printf("Processing message of type %s from %s", msg.Type, msg.From)

	switch msg.Type {
	case MessageElection, MessageOK, MessageCoordinator:
		if err := n.elector.HandleMessage(ctx, msg); err != nil {
			n.logger.Printf("Failed to handle election message: %v", err)
		}
		return
	case MessageHeartbeat:
		n.handleHeartbeat(msg)
	case MessageDiscovery:
		n.handleDiscovery(msg)
	case MessageNodeJoined:
		n.handleNodeJoined(msg)
	case MessageNodeLeft:
		n.handleNodeLeft(msg)
	case MessageLockRequest:
		n.handleLockRequest(msg)
	case MessageLockAck:
		n.handleLockAck(msg)
	case MessageLockRelease:
		n.handleLockRelease(msg)
	default:
		n.logger.Printf("Received msg of type %s from %s", msg.Type, msg.From)
	}
}

func (n *Node) handleHeartbeat(msg Message) {
	var nodeInfo NodeInfo
	if err := json.Unmarshal(msg.Payload, &nodeInfo); err != nil {
		n.logger.Printf("Failed to unmarshal heartbeat payload: %v", err)
		return
	}

	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	if peer, exists := n.Peers[msg.From]; exists {
		peer.LastHeartbeat = time.Now()
		peer.Status = nodeInfo.Status
	} else {
		nodeInfo.LastHeartbeat = time.Now()
		n.Peers[msg.From] = &nodeInfo
		n.logger.Printf("New peer discovered: %s (%s:%d)", msg.From, nodeInfo.IP, nodeInfo.Port)
	}
}

func (n *Node) handleDiscovery(msg Message) {
	n.logger.Printf("Discovery message received in regular handler from %s", msg.From)
}

func (n *Node) handleDiscoveryWithConnection(msg common.Message, conn net.Conn) {
	n.peerMutex.RLock()
	peers := make([]*NodeInfo, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, peer)
	}
	n.peerMutex.RUnlock()

	myInfo := n.GetNodeInfo()
	peers = append(peers, myInfo)

	payload, _ := json.Marshal(peers)

	newTime := n.IncrementAndGetLogicalTime()
	response := MessageWithTime{
		Message: Message{
			Type:    MessageDiscovery,
			From:    n.ID,
			To:      msg.From,
			Payload: payload,
		},
		LogicalTime: newTime,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(response); err != nil {
		n.logger.Printf("Failed to send discovery response: %v", err)
		return
	}

	n.logger.Printf("Sent discovery response to %s with %d peers", msg.From, len(peers))
}

func (n *Node) handleNodeJoined(msg Message) {
	var nodeInfo NodeInfo
	if err := json.Unmarshal(msg.Payload, &nodeInfo); err != nil {
		n.logger.Printf("Failed to unmarshal node joined payload: %v", err)
		return
	}

	n.AddPeer(&nodeInfo)
	n.logger.Printf("Node joined: %s (%s:%d)", nodeInfo.ID, nodeInfo.IP, nodeInfo.Port)
}

func (n *Node) handleNodeLeft(msg Message) {
	n.RemovePeer(msg.From)
	n.logger.Printf("Node left: %s", msg.From)
}

func (n *Node) handleLockRequest(msg Message) {
	var lockRequest LockRequest
	if err := json.Unmarshal(msg.Payload, &lockRequest); err != nil {
		n.logger.Printf("Failed to unmarshal lock request payload: %v", err)
		return
	}

	n.lockQueueMutex.Lock()
	n.lockQueue[lockRequest.ResourceID] = append(n.lockQueue[lockRequest.ResourceID], &lockRequest)
	queue := n.lockQueue[lockRequest.ResourceID]
	sort.SliceStable(queue, func(i, j int) bool {
		return common.CompareLockRequests(queue[i], queue[j])
	})
	n.lockQueueMutex.Unlock()

	n.peerMutex.RLock()
	p, exists := n.Peers[msg.From]
	n.peerMutex.RUnlock()
	if !exists {
		n.logger.Printf("Unknown peer: %s", msg.From)
		return
	}

	payload, _ := json.Marshal(lockRequest.ResourceID)
	ackMsg := Message{
		Type:    common.MessageLockAck,
		From:    n.ID,
		To:      p.ID,
		Payload: payload,
	}
	go n.SendMessage(p.IP, p.Port, ackMsg)
}

func (n *Node) handleLockAck(msg Message) {
	var resourceID string
	if err := json.Unmarshal(msg.Payload, &resourceID); err != nil {
		n.logger.Printf("Failed to unmarshal lock ack payload: %v", err)
		return
	}

	n.lockQueueMutex.Lock()
	defer n.lockQueueMutex.Unlock()
	if _, exists := n.pendingAcks[resourceID]; !exists {
		n.pendingAcks[resourceID] = make(map[uuid.UUID]bool)
	}
	n.pendingAcks[resourceID][msg.From] = true

	n.logger.Printf("Received ack for %s from %s (total: %d/%d)", resourceID, msg.From, len(n.pendingAcks[resourceID]), len(n.Peers))
}

func (n *Node) handleLockRelease(msg Message) {
	var resourceID string
	if err := json.Unmarshal(msg.Payload, &resourceID); err != nil {
		n.logger.Printf("Failed to unmarshal lock release payload: %v", err)
		return
	}

	n.lockQueueMutex.Lock()
	defer n.lockQueueMutex.Unlock()

	for i, req := range n.lockQueue[resourceID] {
		if req.NodeID == msg.From {
			n.lockQueue[resourceID] = append(n.lockQueue[resourceID][:i], n.lockQueue[resourceID][i+1:]...)
			n.logger.Printf("Lock on %s released by %s", resourceID, msg.From)
			break
		}
	}

}

func (n *Node) startHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.sendHeartbeat()
		}
	}
}

func (n *Node) sendHeartbeat() {
	nodeInfo := n.GetNodeInfo()
	payload, err := json.Marshal(nodeInfo)
	if err != nil {
		n.logger.Printf("Failed to marshal node info: %v", err)
		return
	}

	msg := Message{
		Type:    MessageHeartbeat,
		From:    n.ID,
		Payload: payload,
	}

	n.peerMutex.RLock()
	peers := make([]*NodeInfo, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, peer)
	}
	n.peerMutex.RUnlock()

	for _, peer := range peers {
		go n.SendMessage(peer.IP, peer.Port, msg)
	}
}

func (n *Node) monitorPeers(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.checkPeersHealth()
		}
	}
}

func (n *Node) checkPeersHealth() {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	now := time.Now()
	for id, peer := range n.Peers {
		if now.Sub(peer.LastHeartbeat) > 15*time.Second {
			n.logger.Printf("Peer timeout: %s (%s:%d)", id, peer.IP, peer.Port)
			if peer.Role == RoleMaster {
				n.logger.Printf("Coordinator %s is unresponsive, notifying elector", id)
				n.elector.NotifyCoordinatorDead(id)
			}
			delete(n.Peers, id)
		}
	}
}

func (n *Node) SendMessage(ip string, port int, msg common.Message) error {
	addr := fmt.Sprintf("%s:%d", ip, port)

	maxRetries := 3
	timeout := 2 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond) // Exponential backoff
			continue
		}
		defer conn.Close()

		newTime := n.IncrementAndGetLogicalTime()
		n.logger.Printf("Sending message of type %s to %s with logical time %d", msg.Type, addr, newTime)

		msgWithTime := MessageWithTime{
			Message:     msg,
			LogicalTime: newTime,
		}

		conn.SetWriteDeadline(time.Now().Add(timeout))

		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(msgWithTime); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to send message to %s after %d retries: %w", addr, maxRetries, lastErr)
}

func (n *Node) BroadcastMessage(msg common.Message) error {
	n.peerMutex.RLock()
	peers := make([]*NodeInfo, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, peer)
	}
	n.peerMutex.RUnlock()

	for _, peer := range peers {
		go func(p *NodeInfo) {
			// Każdy odbiorca dostanie inny logical time -- czy to zamierzone?
			if err := n.SendMessage(p.IP, p.Port, msg); err != nil {
				n.logger.Printf("Failed to send message to %s: %v", p.ID, err)
			}
		}(peer)
	}
	return nil
}

func (n *Node) DiscoverPeers(seedIP string, seedPort int) error {

	addr := fmt.Sprintf("%s:%d", seedIP, seedPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to seed node: %w", err)
	}
	defer conn.Close()

	newTime := n.IncrementAndGetLogicalTime()
	msg := MessageWithTime{
		Message: Message{
			Type: MessageDiscovery,
			From: n.ID,
		},
		LogicalTime: newTime,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send discovery request: %w", err)
	}

	decoder := json.NewDecoder(conn)
	var response MessageWithTime
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("failed to receive discovery response: %w", err)
	}

	n.UpdateAndGetLogicalTime(response.LogicalTime)

	var peers []*NodeInfo
	if err := json.Unmarshal(response.Message.Payload, &peers); err != nil {
		return fmt.Errorf("failed to unmarshal peer list: %w", err)
	}

	for _, peer := range peers {
		if peer.ID != n.ID {
			n.AddPeer(peer)
		}
	}

	n.logger.Printf("Discovered %d peers from seed node", len(peers))

	n.announceJoin()

	return nil
}

func (n *Node) announceJoin() {
	nodeInfo := n.GetNodeInfo()
	payload, _ := json.Marshal(nodeInfo)

	msg := Message{
		Type:    MessageNodeJoined,
		From:    n.ID,
		Payload: payload,
	}

	n.BroadcastMessage(msg)
}

func (n *Node) AddPeer(peer *NodeInfo) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	peer.LastHeartbeat = time.Now()
	n.Peers[peer.ID] = peer
}

func (n *Node) GetPeer(peerID uuid.UUID) (*NodeInfo, bool) {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	peer, exists := n.Peers[peerID]
	return peer, exists
}

func (n *Node) GetAllPeers() []*NodeInfo {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	peers := make([]*NodeInfo, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, peer)
	}
	return peers
}

func (n *Node) GetNodeInfo() *NodeInfo {
	return &NodeInfo{
		ID:            n.ID,
		IP:            n.IP,
		Port:          n.Port,
		Role:          n.Role,
		Status:        n.Status,
		Priority:      n.Priority,
		LastHeartbeat: time.Now(),
	}
}

func (n *Node) GetID() uuid.UUID {
	return n.ID
}

func (n *Node) GetPriority() int {
	return n.Priority
}

func (n *Node) GetRole() NodeRole {
	return n.Role
}

func (n *Node) SetRole(role NodeRole) {
	n.Role = role
}

func (n *Node) GetPeers() []NodeInfo {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	peers := make([]NodeInfo, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, *peer)
	}
	return peers
}

func (n *Node) UpdatePeerRole(peerID uuid.UUID, role NodeRole) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	if peer, exists := n.Peers[peerID]; exists {
		peer.Role = role
	}
}

func (n *Node) RemovePeer(peerID uuid.UUID) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	delete(n.Peers, peerID)
}

func (n *Node) GetMessageChan() <-chan Message {
	return n.messageChan
}

func (n *Node) IncrementAndGetLogicalTime() int {
	n.logicalTimeMutex.Lock()
	defer n.logicalTimeMutex.Unlock()
	n.logicalTime++
	return n.logicalTime
}

func (n *Node) UpdateAndGetLogicalTime(receivedTime int) int {
	n.logicalTimeMutex.Lock()
	defer n.logicalTimeMutex.Unlock()
	n.logicalTime = max(receivedTime, n.logicalTime) + 1
	return n.logicalTime
}

func (n *Node) RequestLock(resourceID string) {
	newTime := n.IncrementAndGetLogicalTime()
	lockRequest := LockRequest{
		ResourceID:  resourceID,
		NodeID:      n.ID,
		LogicalTime: newTime,
		Status:      common.StatusPending,
	}

	n.lockQueueMutex.Lock()
	n.lockQueue[lockRequest.ResourceID] = append(n.lockQueue[lockRequest.ResourceID], &lockRequest)
	queue := n.lockQueue[lockRequest.ResourceID]
	sort.SliceStable(queue, func(i, j int) bool {
		return common.CompareLockRequests(queue[i], queue[j])
	})
	n.pendingAcks[resourceID] = make(map[uuid.UUID]bool)
	n.lockQueueMutex.Unlock()

	payload, _ := json.Marshal(lockRequest)
	msg := Message{
		Type:    common.MessageLockRequest,
		From:    n.ID,
		Payload: payload,
	}
	n.BroadcastMessage(msg)
	n.logger.Printf("Requested lock on %s with timestamp %d", resourceID, newTime)
}

func (n *Node) ReleaseLock(resourceID string) {
	n.lockQueueMutex.Lock()
	defer n.lockQueueMutex.Unlock()
	for i, req := range n.lockQueue[resourceID] {
		if req.NodeID == n.ID {
			n.lockQueue[resourceID] = append(n.lockQueue[resourceID][:i], n.lockQueue[resourceID][i+1:]...)
			break
		}
	}

	payload, _ := json.Marshal(resourceID)
	msg := Message{
		Type:    common.MessageLockRelease,
		From:    n.ID,
		Payload: payload,
	}
	n.BroadcastMessage(msg)

	delete(n.pendingAcks, resourceID)

	n.logger.Printf("Released lock on %s", resourceID)
}

func (n *Node) CanEnterCriticalSection(resourceID string) bool {
	n.lockQueueMutex.RLock()
	defer n.lockQueueMutex.RUnlock()

	if len(n.lockQueue[resourceID]) == 0 {
		return false
	}

	if n.lockQueue[resourceID][0].NodeID != n.ID {
		return false
	}

	n.peerMutex.RLock()
	expectedAcks := len(n.Peers)
	n.peerMutex.RUnlock()

	if expectedAcks == 0 {
		return true
	}

	receivedAcks := 0
	for _, acked := range n.pendingAcks[resourceID] {
		if acked {
			receivedAcks++
		}
	}

	return receivedAcks >= expectedAcks
}
