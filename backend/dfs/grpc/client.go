package grpc

import (
	"context"
	"dfs-backend/dfs/common"
	pb "dfs-backend/proto"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// PeerConnectionManager manages persistent gRPC connections to peer nodes
type PeerConnectionManager struct {
	connections map[uuid.UUID]*grpc.ClientConn
	clients     map[uuid.UUID]pb.NodeServiceClient
	mu          sync.RWMutex

	selfID uuid.UUID
	logger *log.Logger

	// For getting logical time
	logicalTimeGetter func() int
}

// NewPeerConnectionManager creates a new connection manager
func NewPeerConnectionManager(selfID uuid.UUID, logicalTimeGetter func() int, logger *log.Logger) *PeerConnectionManager {
	return &PeerConnectionManager{
		connections:       make(map[uuid.UUID]*grpc.ClientConn),
		clients:           make(map[uuid.UUID]pb.NodeServiceClient),
		selfID:            selfID,
		logicalTimeGetter: logicalTimeGetter,
		logger:            logger,
	}
}

// Connect establishes a connection to a peer node
func (m *PeerConnectionManager) Connect(peer *common.NodeInfo) error {
	if peer.ID == m.selfID {
		return nil // Don't connect to self
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already connected
	if _, exists := m.clients[peer.ID]; exists {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", peer.IP, peer.Port)

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	m.connections[peer.ID] = conn
	m.clients[peer.ID] = pb.NewNodeServiceClient(conn)

	m.logger.Printf("Connected to peer %s at %s", peer.ID, addr)
	return nil
}

// Disconnect closes the connection to a peer
func (m *PeerConnectionManager) Disconnect(peerID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, exists := m.connections[peerID]; exists {
		conn.Close()
		delete(m.connections, peerID)
		delete(m.clients, peerID)
		m.logger.Printf("Disconnected from peer %s", peerID)
	}
}

// GetClient returns the gRPC client for a peer
func (m *PeerConnectionManager) GetClient(peerID uuid.UUID) (pb.NodeServiceClient, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[peerID]
	return client, exists
}

// GetAllClients returns all connected clients
func (m *PeerConnectionManager) GetAllClients() map[uuid.UUID]pb.NodeServiceClient {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[uuid.UUID]pb.NodeServiceClient, len(m.clients))
	for id, client := range m.clients {
		result[id] = client
	}
	return result
}

// Close closes all connections
func (m *PeerConnectionManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, conn := range m.connections {
		conn.Close()
		m.logger.Printf("Closed connection to peer %s", id)
	}

	m.connections = make(map[uuid.UUID]*grpc.ClientConn)
	m.clients = make(map[uuid.UUID]pb.NodeServiceClient)
}

// EnsureConnected ensures a connection exists to a peer, creating one if needed
func (m *PeerConnectionManager) EnsureConnected(peer *common.NodeInfo) (pb.NodeServiceClient, error) {
	m.mu.RLock()
	if client, exists := m.clients[peer.ID]; exists {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	if err := m.Connect(peer); err != nil {
		return nil, err
	}

	m.mu.RLock()
	client := m.clients[peer.ID]
	m.mu.RUnlock()

	return client, nil
}

// ============================================================================
// High-level RPC methods
// ============================================================================

// SendHeartbeat sends a heartbeat to a peer
func (m *PeerConnectionManager) SendHeartbeat(ctx context.Context, peer *common.NodeInfo, nodeInfo *common.NodeInfo) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.HeartbeatRequest{
		NodeInfo:    NodeInfoToProto(nodeInfo),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.Heartbeat(ctx, req)
	return err
}

// SendElection sends an election message and waits for OK response
func (m *PeerConnectionManager) SendElection(ctx context.Context, peer *common.NodeInfo, nodeInfo *common.NodeInfo) (*pb.ElectionResponse, error) {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return nil, err
	}

	req := &pb.ElectionRequest{
		FromId:      m.selfID.String(),
		NodeInfo:    NodeInfoToProto(nodeInfo),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	return client.Election(ctx, req)
}

// SendCoordinator announces this node as coordinator
func (m *PeerConnectionManager) SendCoordinator(ctx context.Context, peer *common.NodeInfo) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.CoordinatorRequest{
		FromId:      m.selfID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.Coordinator(ctx, req)
	return err
}

// SendDiscovery discovers peers from a seed node
func (m *PeerConnectionManager) SendDiscovery(ctx context.Context, peer *common.NodeInfo) ([]*common.NodeInfo, error) {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return nil, err
	}

	req := &pb.DiscoveryRequest{
		FromId:      m.selfID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	resp, err := client.Discovery(ctx, req)
	if err != nil {
		return nil, err
	}

	return ProtoToNodeInfoSlice(resp.Peers), nil
}

// SendNodeJoined announces this node has joined
func (m *PeerConnectionManager) SendNodeJoined(ctx context.Context, peer *common.NodeInfo, nodeInfo *common.NodeInfo) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.NodeJoinedRequest{
		NodeInfo:    NodeInfoToProto(nodeInfo),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.NodeJoined(ctx, req)
	return err
}

// SendNodeLeft announces a node has left
func (m *PeerConnectionManager) SendNodeLeft(ctx context.Context, peer *common.NodeInfo, leftNodeID uuid.UUID) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.NodeLeftRequest{
		NodeId:      leftNodeID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.NodeLeft(ctx, req)
	return err
}

// SendLockRequest requests a lock from a peer
func (m *PeerConnectionManager) SendLockRequest(ctx context.Context, peer *common.NodeInfo, lockReq *common.LockRequest) (*pb.LockAck, error) {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return nil, err
	}

	req := LockRequestToProto(lockReq)
	req.LogicalTime = int32(m.logicalTimeGetter())

	return client.RequestLock(ctx, req)
}

// SendLockAcquired notifies a peer that we acquired a lock
func (m *PeerConnectionManager) SendLockAcquired(ctx context.Context, peer *common.NodeInfo, resourceID string) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.LockAcquiredMsg{
		ResourceId:  resourceID,
		FromId:      m.selfID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.AcquireLock(ctx, req)
	return err
}

// SendLockRelease notifies a peer that we released a lock
func (m *PeerConnectionManager) SendLockRelease(ctx context.Context, peer *common.NodeInfo, resourceID string) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.LockReleaseMsg{
		ResourceId:  resourceID,
		FromId:      m.selfID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.ReleaseLock(ctx, req)
	return err
}

// SendLockAbort tells a peer to abort its locks
func (m *PeerConnectionManager) SendLockAbort(ctx context.Context, peer *common.NodeInfo) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.LockAbortMsg{
		FromId:      m.selfID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.AbortLock(ctx, req)
	return err
}

// SendFileDelete tells a peer to delete a file
func (m *PeerConnectionManager) SendFileDelete(ctx context.Context, peer *common.NodeInfo, hash string) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.FileDeleteRequest{
		Hash:        hash,
		FromId:      m.selfID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.DeleteFile(ctx, req)
	return err
}

// SendMetadataRequest requests metadata from a peer
func (m *PeerConnectionManager) SendMetadataRequest(ctx context.Context, peer *common.NodeInfo, requestID string) (*common.MetadataResponse, error) {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return nil, err
	}

	req := &pb.MetadataReq{
		RequestId:   requestID,
		FromId:      m.selfID.String(),
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	resp, err := client.GetMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	files := make([]common.NodeFileMetadata, len(resp.Files))
	for i, f := range resp.Files {
		files[i] = *ProtoToNodeFileMetadata(f)
	}

	return &common.MetadataResponse{
		RequestID: resp.RequestId,
		NodeID:    ParseUUID(resp.NodeId),
		Files:     files,
	}, nil
}

// SendReplicateFile tells a peer to replicate a file
func (m *PeerConnectionManager) SendReplicateFile(ctx context.Context, peer *common.NodeInfo, req *common.ReplicateFileRequest) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	targetNodes := make([]string, len(req.TargetNodes))
	for i, t := range req.TargetNodes {
		targetNodes[i] = t.String()
	}

	pbReq := &pb.ReplicateFileReq{
		FileId:      req.FileID.String(),
		Filename:    req.Filename,
		Hash:        req.Hash,
		SourceNode:  req.SourceNode.String(),
		TargetNodes: targetNodes,
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.ReplicateFile(ctx, pbReq)
	return err
}

// SendStreamStoreAck sends a stream store acknowledgment to the master
func (m *PeerConnectionManager) SendStreamStoreAck(ctx context.Context, peer *common.NodeInfo, ack *common.StreamStoreAck) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.StreamStoreAckMsg{
		SessionId:   ack.SessionID,
		FileId:      ack.FileID.String(),
		Hash:        ack.Hash,
		NodeId:      ack.NodeID.String(),
		Success:     ack.Success,
		Error:       ack.Error,
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.StreamStoreAck(ctx, req)
	return err
}

// SendStreamUploadReady tells a storage node to prepare for an upload session
func (m *PeerConnectionManager) SendStreamUploadReady(ctx context.Context, peer *common.NodeInfo, sessionID string, fileID uuid.UUID, filename, contentType string, size int64) error {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return err
	}

	req := &pb.StreamUploadReadyMsg{
		SessionId:   sessionID,
		FileId:      fileID.String(),
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	_, err = client.StreamUploadReady(ctx, req)
	return err
}

// SendFileStore sends a file to a peer for storage (used in replication)
func (m *PeerConnectionManager) SendFileStore(ctx context.Context, peer *common.NodeInfo, req *common.FileStoreRequest) (*common.FileStoreResponse, error) {
	client, err := m.EnsureConnected(peer)
	if err != nil {
		return nil, err
	}

	pbReq := &pb.FileStoreRequest{
		FileId:      req.FileID.String(),
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        req.Size,
		Data:        req.Data,
		LogicalTime: int32(m.logicalTimeGetter()),
	}

	resp, err := client.StoreFile(ctx, pbReq)
	if err != nil {
		return nil, err
	}

	return &common.FileStoreResponse{
		Hash:   resp.Hash,
		FileID: ParseUUID(resp.FileId),
	}, nil
}

// ============================================================================
// Broadcast methods
// ============================================================================

// BroadcastHeartbeat sends heartbeat to all peers
func (m *PeerConnectionManager) BroadcastHeartbeat(ctx context.Context, peers []*common.NodeInfo, nodeInfo *common.NodeInfo) {
	for _, peer := range peers {
		go func(p *common.NodeInfo) {
			if err := m.SendHeartbeat(ctx, p, nodeInfo); err != nil {
				m.logger.Printf("Failed to send heartbeat to %s: %v", p.ID, err)
			}
		}(peer)
	}
}

// BroadcastCoordinator announces coordinator to all peers
func (m *PeerConnectionManager) BroadcastCoordinator(ctx context.Context, peers []*common.NodeInfo) {
	for _, peer := range peers {
		go func(p *common.NodeInfo) {
			if err := m.SendCoordinator(ctx, p); err != nil {
				m.logger.Printf("Failed to send coordinator to %s: %v", p.ID, err)
			}
		}(peer)
	}
}

// BroadcastNodeJoined announces node join to all peers
func (m *PeerConnectionManager) BroadcastNodeJoined(ctx context.Context, peers []*common.NodeInfo, nodeInfo *common.NodeInfo) {
	for _, peer := range peers {
		go func(p *common.NodeInfo) {
			if err := m.SendNodeJoined(ctx, p, nodeInfo); err != nil {
				m.logger.Printf("Failed to send node joined to %s: %v", p.ID, err)
			}
		}(peer)
	}
}

// BroadcastLockAcquired announces lock acquisition to all peers
func (m *PeerConnectionManager) BroadcastLockAcquired(ctx context.Context, peers []*common.NodeInfo, resourceID string) {
	for _, peer := range peers {
		go func(p *common.NodeInfo) {
			if err := m.SendLockAcquired(ctx, p, resourceID); err != nil {
				m.logger.Printf("Failed to send lock acquired to %s: %v", p.ID, err)
			}
		}(peer)
	}
}

// BroadcastLockRelease announces lock release to all peers
func (m *PeerConnectionManager) BroadcastLockRelease(ctx context.Context, peers []*common.NodeInfo, resourceID string) {
	for _, peer := range peers {
		go func(p *common.NodeInfo) {
			if err := m.SendLockRelease(ctx, p, resourceID); err != nil {
				m.logger.Printf("Failed to send lock release to %s: %v", p.ID, err)
			}
		}(peer)
	}
}
