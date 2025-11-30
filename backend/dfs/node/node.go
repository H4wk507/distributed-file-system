package node

import (
	"bytes"
	"context"
	"dfs-backend/dfs/common"
	. "dfs-backend/dfs/common"
	"dfs-backend/dfs/election"
	"dfs-backend/dfs/hashing"
	"dfs-backend/dfs/storage"
	"encoding/json"
	"fmt"
	"io"
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

	peers     map[uuid.UUID]*NodeInfo
	peerMutex sync.RWMutex

	listener net.Listener

	messageChan chan common.Message
	stopChan    chan struct{}

	logicalTime      int
	logicalTimeMutex sync.Mutex

	lockQueue      map[string][]*LockRequest // key = resourceID, value = request queue
	pendingAcks    map[string]map[uuid.UUID]bool
	lockHolders    map[string]uuid.UUID // key = resourceID, value = NodeID that holds the lock
	lockQueueMutex sync.RWMutex

	waitForGraph      map[uuid.UUID]map[uuid.UUID]bool // key = NodeID that waits, value = map of NodeIDs for which key waits
	waitForGraphMutex sync.RWMutex

	hashRing *hashing.HashRing // tylko dla master, nil dla storage nodes

	storage *storage.LocalStorage

	pendingUploads      map[uuid.UUID]*common.PendingUpload // tylko dla master, nil dla storage nodes
	pendingUploadsMutex sync.RWMutex

	logger *log.Logger
}

func CreateNodeWithBully(ip string, port int, role NodeRole, priority int) *Node {
	n := &Node{
		ID:           uuid.New(),
		IP:           ip,
		Port:         port,
		Role:         role,
		Status:       StatusStarting,
		Priority:     priority,
		peers:        make(map[uuid.UUID]*NodeInfo),
		messageChan:  make(chan common.Message, 100),
		stopChan:     make(chan struct{}),
		logicalTime:  0,
		lockQueue:    make(map[string][]*LockRequest),
		pendingAcks:  make(map[string]map[uuid.UUID]bool),
		lockHolders:  make(map[string]uuid.UUID),
		waitForGraph: make(map[uuid.UUID]map[uuid.UUID]bool),
		logger:       log.New(log.Writer(), fmt.Sprintf("[Node %s] ", role), log.LstdFlags),
	}

	n.storage = storage.NewLocalStorage(fmt.Sprintf("./data/node-%s", n.ID))
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
	n.SetRole(n.Role)
	n.storage.LoadIndex()
	n.logger.Printf("Node started on %s (ID: %s, Priority: %d)", addr, n.ID, n.Priority)

	go n.acceptConnections(ctx)
	go n.handleMessages(ctx)
	go n.startHeartbeat(ctx)
	go n.monitorPeers(ctx)
	go n.monitorLocks(ctx)
	go n.monitorDeadlocks(ctx)
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
	case MessageLockAcquired:
		n.handleLockAcquired(msg)
	case MessageLockRelease:
		n.handleLockRelease(msg)
	case MessageLockAbort:
		n.handleLockAbort(msg)
	case MessageFileStore:
		n.handleFileStore(msg)
	case MessageFileStoreAck:
		n.handleFileStoreAck(msg)
	case MessageFileRetrieve:
		n.handleFileRetrieve(msg)
	case MessageFileRetrieveResponse:
		n.handleFileRetrieveResponse(msg)
	case MessageFileDelete:
		n.handleFileDelete(msg)
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

	if peer, exists := n.peers[msg.From]; exists {
		peer.LastHeartbeat = time.Now()
		peer.Status = nodeInfo.Status
	} else {
		nodeInfo.LastHeartbeat = time.Now()
		n.peers[msg.From] = &nodeInfo
		n.logger.Printf("New peer discovered: %s (%s:%d)", msg.From, nodeInfo.IP, nodeInfo.Port)
	}
}

func (n *Node) handleDiscovery(msg Message) {
	n.logger.Printf("Discovery message received in regular handler from %s", msg.From)
}

func (n *Node) handleDiscoveryWithConnection(msg common.Message, conn net.Conn) {
	n.peerMutex.RLock()
	peers := make([]*NodeInfo, 0, len(n.peers))
	for _, peer := range n.peers {
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

	if n.hashRing != nil && nodeInfo.Role == common.RoleStorage {
		n.hashRing.AddNode(nodeInfo)
	}
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

	holder, exists := n.lockHolders[lockRequest.ResourceID]
	if exists && holder != lockRequest.NodeID {
		n.waitForGraphMutex.Lock()
		if n.waitForGraph[lockRequest.NodeID] == nil {
			n.waitForGraph[lockRequest.NodeID] = make(map[uuid.UUID]bool)
		}
		n.waitForGraph[lockRequest.NodeID][holder] = true
		n.waitForGraphMutex.Unlock()
	}

	n.lockQueueMutex.Unlock()

	n.peerMutex.RLock()
	p, exists := n.peers[msg.From]
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

	n.logger.Printf("Received ack for %s from %s (total: %d/%d)", resourceID, msg.From, len(n.pendingAcks[resourceID]), len(n.peers))
}

func (n *Node) handleLockAcquired(msg Message) {
	var resourceID string
	if err := json.Unmarshal(msg.Payload, &resourceID); err != nil {
		n.logger.Printf("Failed to unmarshal lock acquired payload: %v", err)
		return
	}

	n.lockQueueMutex.Lock()
	n.lockHolders[resourceID] = msg.From
	n.lockQueueMutex.Unlock()
}

func (n *Node) handleLockRelease(msg Message) {
	var resourceID string
	if err := json.Unmarshal(msg.Payload, &resourceID); err != nil {
		n.logger.Printf("Failed to unmarshal lock release payload: %v", err)
		return
	}

	n.lockQueueMutex.Lock()
	defer n.lockQueueMutex.Unlock()

	n.waitForGraphMutex.Lock()
	for waiter := range n.waitForGraph {
		for nodeID := range n.waitForGraph[waiter] {
			if nodeID == msg.From {
				delete(n.waitForGraph[waiter], nodeID)
			}
		}
	}
	n.waitForGraphMutex.Unlock()

	delete(n.lockHolders, resourceID)

	for i, req := range n.lockQueue[resourceID] {
		if req.NodeID == msg.From {
			n.lockQueue[resourceID] = append(n.lockQueue[resourceID][:i], n.lockQueue[resourceID][i+1:]...)
			n.logger.Printf("Lock on %s released by %s", resourceID, msg.From)
			break
		}
	}
}

func (n *Node) handleLockAbort(msg Message) {
	// To zwalnia locki na wszystkich resources dla danego node'a. Nie mamy tego rozróżnienia w grafie.
	// TODO: Dodać to rozróżnienie na przyszłość??

	n.logger.Printf("Received lock abort from %s", msg.From)

	n.lockQueueMutex.Lock()
	resourcesToRelease := []string{}
	for resourceID, queue := range n.lockQueue {
		for _, req := range queue {
			if req.NodeID == n.ID {
				resourcesToRelease = append(resourcesToRelease, resourceID)
				break
			}
		}
	}
	n.lockQueueMutex.Unlock()

	for _, resourceID := range resourcesToRelease {
		n.ReleaseLock(resourceID)
	}
}

// TODO: what about very big files?
func (n *Node) handleFileStore(msg Message) {
	var fileRequest common.FileStoreRequest
	if err := json.Unmarshal(msg.Payload, &fileRequest); err != nil {
		n.logger.Printf("Failed to unmarshal FileStoreRequest payload: %v", err)
		return
	}

	r := bytes.NewReader(fileRequest.Data)
	meta, err := n.storage.SaveFile(fileRequest.FileID, fileRequest.Filename, fileRequest.ContentType, r)
	if err != nil {
		n.logger.Printf("Failed to save file %s to storage: %v", fileRequest.Filename, err)
		return
	}

	peer, exists := n.GetPeer(msg.From)
	if !exists {
		n.logger.Printf("Peer %s not found", msg.From)
		return
	}

	fileStoreResponsePayload := common.FileStoreResponse{
		Hash:   meta.Hash,
		FileID: meta.FileID,
	}
	payload, err := json.Marshal(fileStoreResponsePayload)
	if err != nil {
		n.logger.Printf("Failed to marshal fileStoreResponsePayload: %v", err)
		return
	}

	ackMsg := Message{
		Type:    common.MessageFileStoreAck,
		From:    n.ID,
		To:      peer.ID,
		Payload: payload,
	}
	go n.SendMessage(peer.IP, peer.Port, ackMsg)
}

func (n *Node) handleFileStoreAck(msg Message) {
	var fileResponse common.FileStoreResponse
	if err := json.Unmarshal(msg.Payload, &fileResponse); err != nil {
		n.logger.Printf("Failed to unmarshal FileStoreResponse request: %v", err)
		return
	}

	n.pendingUploadsMutex.Lock()
	defer n.pendingUploadsMutex.Unlock()
	pending, exists := n.pendingUploads[fileResponse.FileID]
	if !exists {
		n.logger.Printf("FileID '%s' was not found in pendingUploads", fileResponse.FileID.String())
		return
	}

	pending.ReceivedAcks[msg.From] = fileResponse.Hash
	if len(pending.ReceivedAcks) == len(pending.ExpectedNodes) {
		var firstHash string
		for _, firstHash = range pending.ReceivedAcks {
			break
		}
		for _, hash := range pending.ReceivedAcks {
			if hash != firstHash {
				n.logger.Print("Hashes of the uploaded files differ!")
			}
		}

		delete(n.pendingUploads, fileResponse.FileID)
		n.logger.Printf("File %s uploaded successfully to %d nodes", pending.Filename, len(pending.ExpectedNodes))
	}
}

func (n *Node) handleFileRetrieve(msg Message) {
	var request common.FileRetrieveRequest
	if err := json.Unmarshal(msg.Payload, &request); err != nil {
		n.logger.Printf("failed to unmarshal FileRetrieveRequest payload: %v", err)
		return
	}
	reader, meta, err := n.storage.GetFile(request.Hash)
	if err != nil {
		n.logger.Printf("error while getting file with hash '%s': %v", request.Hash, err)
		// TODO: odeslij wiadomosc z bledem
		return
	}
	defer reader.Close()

	// TODO: what about big files?
	data, err := io.ReadAll(reader)
	if err != nil {
		n.logger.Printf("failed to read file: %v", err)
		return
	}

	peer, exists := n.GetPeer(msg.From)
	if !exists {
		n.logger.Printf("peer %s does not exist", msg.From)
		return
	}

	response := common.FileRetrieveResponse{
		FileID:   meta.FileID,
		Filename: meta.Filename,
		Hash:     meta.Hash,
		Data:     data,
	}
	payload, err := json.Marshal(response)
	if err != nil {
		n.logger.Printf("Failed to marshal FileRetrieveResponse: %v", err)
		return
	}

	responseMsg := common.Message{
		Type:    common.MessageFileRetrieveResponse,
		From:    n.ID,
		To:      peer.ID,
		Payload: payload,
	}
	go n.SendMessage(peer.IP, peer.Port, responseMsg)
}

func (n *Node) handleFileRetrieveResponse(msg Message) {
	var response common.FileRetrieveResponse
	if err := json.Unmarshal(msg.Payload, &response); err != nil {
		n.logger.Printf("Failed to unmarshal FileRetrieveResponse: %v", err)
		return
	}

	n.logger.Printf("Received file %s (%d bytes)", response.Filename, len(response.Data))
	// TODO: API HTTP
}

// TODO: response
func (n *Node) handleFileDelete(msg Message) {
	var hash string
	if err := json.Unmarshal(msg.Payload, &hash); err != nil {
		n.logger.Printf("failed to unmarshal MessageFileDelete: %v", err)
		return
	}

	err := n.storage.DeleteFile(hash)
	if err != nil {
		n.logger.Printf("Error while deleting file '%s': %v", hash, err)
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
	peers := make([]*NodeInfo, 0, len(n.peers))
	for _, peer := range n.peers {
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

func (n *Node) monitorLocks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.checkLocksTimeouts()
		}
	}
}

func (n *Node) monitorDeadlocks(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.checkDeadlocks()
		}
	}
}

func (n *Node) checkPeersHealth() {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	now := time.Now()
	for id, peer := range n.peers {
		if now.Sub(peer.LastHeartbeat) > 15*time.Second {
			n.logger.Printf("Peer timeout: %s (%s:%d)", id, peer.IP, peer.Port)
			if peer.Role == RoleMaster {
				n.logger.Printf("Coordinator %s is unresponsive, notifying elector", id)
				n.elector.NotifyCoordinatorDead(id)
			}
			delete(n.peers, id)
		}
	}
}

func (n *Node) checkLocksTimeouts() {
	n.lockQueueMutex.Lock()

	var toRealease []string

	for resourceID, queue := range n.lockQueue {
		newQueue := make([]*LockRequest, 0)
		for _, req := range queue {
			if time.Since(req.RequestedAt) > common.LOCK_TIMEOUT {
				n.logger.Printf("Lock request timeout on resource %s by node %s", resourceID, req.NodeID)
				if req.NodeID == n.ID {
					delete(n.pendingAcks, resourceID)
					toRealease = append(toRealease, resourceID)
				}
				continue
			}
			newQueue = append(newQueue, req)
		}
		n.lockQueue[resourceID] = newQueue
	}
	n.lockQueueMutex.Unlock()

	for _, resourceID := range toRealease {
		n.logger.Printf("Releasing lock on %s due to timeout", resourceID)
		payload, _ := json.Marshal(resourceID)
		msg := Message{
			Type:    common.MessageLockRelease,
			From:    n.ID,
			Payload: payload,
		}
		n.BroadcastMessage(msg)
	}
}

func (n *Node) checkDeadlocks() {
	if n.Role != common.RoleMaster {
		return
	}
	hasCycle, cyclePath := common.HasCycle(n.waitForGraph)
	if hasCycle {
		n.logger.Printf("Cycle detected in waitForGraph: %v", cyclePath)
		victim := n.selectVictim(cyclePath)
		n.sendAbort(victim)
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
	peers := make([]*NodeInfo, 0, len(n.peers))
	for _, peer := range n.peers {
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

func (n *Node) selectVictim(cyclePath []uuid.UUID) uuid.UUID {
	var victim uuid.UUID
	lowestPriority := int(^uint(0) >> 1)
	for _, nodeID := range cyclePath {
		var priority int
		if nodeID == n.ID {
			priority = n.Priority
		} else if peer, exists := n.GetPeer(nodeID); exists {
			priority = peer.Priority
		} else {
			continue
		}

		if priority < lowestPriority {
			lowestPriority = priority
			victim = nodeID
		}
	}
	return victim
}

func (n *Node) sendAbort(victim uuid.UUID) {
	msg := Message{
		Type: common.MessageLockAbort,
		From: n.ID,
	}

	if victim == n.ID {
		n.logger.Printf("Aborting lock on myself")
		n.handleLockAbort(msg)
		return
	}

	peer, exists := n.GetPeer(victim)
	if !exists {
		n.logger.Printf("Peer %s not found", victim)
		return
	}

	go n.SendMessage(peer.IP, peer.Port, msg)
}

func (n *Node) initMasterResources() {
	n.hashRing = hashing.NewHashRing(150)
	n.pendingUploads = make(map[uuid.UUID]*common.PendingUpload)

	for _, peer := range n.GetAllPeers() {
		if peer.Role == common.RoleStorage {
			n.hashRing.AddNode(*peer)
		}
	}
	n.logger.Printf("Initialized master resources with %d storage nodes", n.hashRing.GetNodeCount())
}

func (n *Node) AddPeer(peer *NodeInfo) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	peer.LastHeartbeat = time.Now()
	n.peers[peer.ID] = peer
}

func (n *Node) GetPeer(peerID uuid.UUID) (*NodeInfo, bool) {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	peer, exists := n.peers[peerID]
	return peer, exists
}

func (n *Node) GetAllPeers() []*NodeInfo {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	peers := make([]*NodeInfo, 0, len(n.peers))
	for _, peer := range n.peers {
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
	oldRole := n.Role
	n.Role = role

	if role == common.RoleMaster && oldRole != common.RoleMaster {
		n.initMasterResources()
	}

	if role != common.RoleMaster && oldRole == common.RoleMaster {
		n.hashRing = nil
		n.pendingUploads = nil
		n.logger.Printf("Cleaned up master resources")
	}
}

func (n *Node) GetPeers() []NodeInfo {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	peers := make([]NodeInfo, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, *peer)
	}
	return peers
}

func (n *Node) UpdatePeerRole(peerID uuid.UUID, role NodeRole) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()

	if peer, exists := n.peers[peerID]; exists {
		peer.Role = role
	}
}

func (n *Node) RemovePeer(peerID uuid.UUID) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	delete(n.peers, peerID)
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
		RequestedAt: time.Now(),
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
	for i, req := range n.lockQueue[resourceID] {
		if req.NodeID == n.ID {
			n.lockQueue[resourceID] = append(n.lockQueue[resourceID][:i], n.lockQueue[resourceID][i+1:]...)
			break
		}
	}
	delete(n.pendingAcks, resourceID)
	n.lockQueueMutex.Unlock()

	payload, _ := json.Marshal(resourceID)
	msg := Message{
		Type:    common.MessageLockRelease,
		From:    n.ID,
		Payload: payload,
	}
	n.BroadcastMessage(msg)

	n.logger.Printf("Released lock on %s", resourceID)
}

func (n *Node) CanEnterCriticalSection(resourceID string) bool {
	n.lockQueueMutex.Lock()
	defer n.lockQueueMutex.Unlock()

	if len(n.lockQueue[resourceID]) == 0 {
		return false
	}

	for len(n.lockQueue[resourceID]) > 0 {
		first := n.lockQueue[resourceID][0]
		if time.Since(first.RequestedAt) <= common.LOCK_TIMEOUT {
			break
		}
		n.lockQueue[resourceID] = n.lockQueue[resourceID][1:]
		if first.NodeID == n.ID {
			delete(n.pendingAcks, resourceID)
		}
	}

	if len(n.lockQueue[resourceID]) == 0 {
		return false
	}

	first := n.lockQueue[resourceID][0]
	if first.NodeID != n.ID {
		return false
	}

	n.peerMutex.RLock()
	expectedAcks := len(n.peers)
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

func (n *Node) EnterCriticalSection(resourceID string) bool {
	if !n.CanEnterCriticalSection(resourceID) {
		return false
	}

	payload, _ := json.Marshal(resourceID)
	msg := Message{
		Type:    common.MessageLockAcquired,
		From:    n.ID,
		Payload: payload,
	}
	n.BroadcastMessage(msg)

	return true
}

func (n *Node) StoreFile(fileID uuid.UUID, filename string, contentType string, data []byte) {
	nodes := n.hashRing.FindNodesForFile(filename, 3)

	expectedNodes := make([]uuid.UUID, len(nodes))
	for i, node := range nodes {
		expectedNodes[i] = node.ID
	}

	pendingUpload := &common.PendingUpload{
		FileID:        fileID,
		Filename:      filename,
		ExpectedNodes: expectedNodes,
		ReceivedAcks:  make(map[uuid.UUID]string),
	}

	n.pendingUploadsMutex.Lock()
	n.pendingUploads[fileID] = pendingUpload
	n.pendingUploadsMutex.Unlock()

	messageFileStorePayload := common.FileStoreRequest{
		FileID:      fileID,
		Filename:    filename,
		ContentType: contentType,
		Size:        int64(len(data)),
		Data:        data,
	}

	payload, err := json.Marshal(messageFileStorePayload)
	if err != nil {
		n.logger.Printf("Failed to marshal messageFileStorePayload: %v", err)
		return
	}

	msg := common.Message{
		Type:    common.MessageFileStore,
		From:    n.ID,
		Payload: payload,
	}

	for _, node := range nodes {
		go n.SendMessage(node.IP, node.Port, msg)
	}
}
