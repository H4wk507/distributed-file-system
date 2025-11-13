package node

import (
	"context"
	. "dfs-backend/dfs/common"
	"dfs-backend/dfs/election"
	"encoding/json"
	"fmt"
	"log"
	"net"
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

	messageChan chan Message
	stopChan    chan struct{}

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
		messageChan: make(chan Message, 100),
		stopChan:    make(chan struct{}),
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
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		n.logger.Printf("Failed to decode message: %v", err)
		return
	}

	if msg.Type == MessageDiscovery {
		n.handleDiscoveryWithConnection(msg, conn)
		return
	}

	select {
	case n.messageChan <- msg:
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

func (n *Node) processMessage(ctx context.Context, msg Message) {
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
	default:
		n.logger.Printf("Received message of type %s from %s", msg.Type, msg.From)
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

func (n *Node) handleDiscoveryWithConnection(msg Message, conn net.Conn) {
	n.peerMutex.RLock()
	peers := make([]*NodeInfo, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, peer)
	}
	n.peerMutex.RUnlock()

	myInfo := n.GetNodeInfo()
	peers = append(peers, myInfo)

	payload, _ := json.Marshal(peers)
	response := Message{
		Type:      MessageDiscovery,
		From:      n.ID,
		To:        msg.From,
		Timestamp: time.Now(),
		Payload:   payload,
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
		Type:      MessageHeartbeat,
		From:      n.ID,
		Timestamp: time.Now(),
		Payload:   payload,
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

func (n *Node) SendMessage(ip string, port int, data interface{}) error {
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

		conn.SetWriteDeadline(time.Now().Add(timeout))

		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(data); err != nil {
			lastErr = err
			continue
		}
		conn.Close()

		return nil
	}

	return fmt.Errorf("failed to send message to %s after %d retries: %w", addr, maxRetries, lastErr)
}

func (n *Node) BroadcastMessage(data interface{}) error {
	n.peerMutex.RLock()
	peers := make([]*NodeInfo, 0, len(n.Peers))
	for _, peer := range n.Peers {
		peers = append(peers, peer)
	}
	n.peerMutex.RUnlock()

	for _, peer := range peers {
		go func(p *NodeInfo) {
			if err := n.SendMessage(p.IP, p.Port, data); err != nil {
				n.logger.Printf("Failed to send message to %s: %v", p.ID, err)
			}
		}(peer)
	}
	return nil
}

func (n *Node) DiscoverPeers(seedIP string, seedPort int) error {
	msg := Message{
		Type:      MessageDiscovery,
		From:      n.ID,
		Timestamp: time.Now(),
	}

	addr := fmt.Sprintf("%s:%d", seedIP, seedPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to seed node: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send discovery request: %w", err)
	}

	decoder := json.NewDecoder(conn)
	var response Message
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("failed to receive discovery response: %w", err)
	}

	var peers []*NodeInfo
	if err := json.Unmarshal(response.Payload, &peers); err != nil {
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
		Type:      MessageNodeJoined,
		From:      n.ID,
		Timestamp: time.Now(),
		Payload:   payload,
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
