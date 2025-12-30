package grpc

import (
	"context"
	"dfs-backend/dfs/common"
	pb "dfs-backend/proto"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NodeHandler defines the interface that the Node must implement
// for the gRPC server to delegate message handling
type NodeHandler interface {
	// Getters
	GetID() uuid.UUID
	GetNodeInfo() *common.NodeInfo
	GetPeers() []common.NodeInfo
	GetPeer(peerID uuid.UUID) (*common.NodeInfo, bool)
	GetRole() common.NodeRole
	GetStorage() StorageInterface

	// Peer management
	AddPeer(peer *common.NodeInfo)
	RemovePeer(peerID uuid.UUID)
	UpdatePeerRole(peerID uuid.UUID, role common.NodeRole)

	// Logical time
	IncrementAndGetLogicalTime() int
	UpdateAndGetLogicalTime(receivedTime int) int

	// Lock handling (internal state access)
	HandleLockRequest(req *common.LockRequest, fromID uuid.UUID)
	HandleLockAck(resourceID string, fromID uuid.UUID)
	HandleLockAcquired(resourceID string, fromID uuid.UUID)
	HandleLockRelease(resourceID string, fromID uuid.UUID)
	HandleLockAbort(fromID uuid.UUID)

	// File operations
	HandleFileDelete(hash string)
	HandleReplicateFile(req *common.ReplicateFileRequest)
	HandleFileStore(req *common.FileStoreRequest) (string, error)

	// Metadata
	HandleMetadataRequest(requestID string, fromID uuid.UUID) *common.MetadataResponse
	HandleMetadataResponse(resp *common.MetadataResponse)

	// Streaming
	HandleStreamUploadInit(req *common.StreamUploadInitRequest) *common.StreamUploadReadyResponse
	HandleStreamDownloadInit(req *common.StreamDownloadInitRequest) *common.StreamDownloadReadyResponse
	HandleStreamStoreAck(ack *common.StreamStoreAck)
	HandleStreamUploadReady(sessionID string, fileID uuid.UUID, filename, contentType string, size int64)

	// Election (via elector)
	HandleElection(ctx context.Context, fromID uuid.UUID, nodeInfo *common.NodeInfo) (bool, error)
	HandleCoordinator(fromID uuid.UUID)

	// Hash ring access (for master)
	GetHashRing() HashRingInterface
	GetGlobalFileIndex() map[string]*common.GlobalFileInfo
}

// StorageInterface abstracts the local storage
type StorageInterface interface {
	GetAllMetadata() []*common.NodeFileMetadata
	DeleteFile(hash string) error
	StoreFile(fileID uuid.UUID, filename, contentType string, size int64, data []byte) (string, error)
}

// HashRingInterface abstracts the hash ring
type HashRingInterface interface {
	FindNodesForFile(fileID uuid.UUID, count int) []common.NodeInfo
}

// NodeServer implements the gRPC NodeService
type NodeServer struct {
	pb.UnimplementedNodeServiceServer

	handler NodeHandler
	server  *grpc.Server
	port    int
	logger  *log.Logger

	// For tracking peers and their heartbeat status (duplicate of node's peers for quick access)
	peerLastHeartbeat      map[uuid.UUID]time.Time
	peerLastHeartbeatMutex sync.RWMutex
}

// NewNodeServer creates a new gRPC server for node communication
func NewNodeServer(port int, handler NodeHandler, logger *log.Logger) *NodeServer {
	return &NodeServer{
		handler:           handler,
		port:              port,
		logger:            logger,
		peerLastHeartbeat: make(map[uuid.UUID]time.Time),
	}
}

// Start starts the gRPC server
func (s *NodeServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.server = grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              10 * time.Second,
			Timeout:           3 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	pb.RegisterNodeServiceServer(s.server, s)

	go func() {
		s.logger.Printf("gRPC server starting on %s", addr)
		if err := s.server.Serve(listener); err != nil {
			s.logger.Printf("gRPC server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	return nil
}

// Stop gracefully stops the gRPC server
func (s *NodeServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
		s.logger.Println("gRPC server stopped")
	}
}

// ============================================================================
// Heartbeat
// ============================================================================

func (s *NodeServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	nodeInfo := ProtoToNodeInfo(req.NodeInfo)
	if nodeInfo == nil {
		return &pb.Empty{}, nil
	}

	// Update peer info
	nodeInfo.LastHeartbeat = time.Now()
	s.handler.AddPeer(nodeInfo)

	s.peerLastHeartbeatMutex.Lock()
	s.peerLastHeartbeat[nodeInfo.ID] = time.Now()
	s.peerLastHeartbeatMutex.Unlock()

	return &pb.Empty{}, nil
}

// ============================================================================
// Election (Bully Algorithm)
// ============================================================================

func (s *NodeServer) Election(ctx context.Context, req *pb.ElectionRequest) (*pb.ElectionResponse, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	fromID := ParseUUID(req.FromId)
	nodeInfo := ProtoToNodeInfo(req.NodeInfo)

	shouldRespond, _ := s.handler.HandleElection(ctx, fromID, nodeInfo)

	if shouldRespond {
		return &pb.ElectionResponse{
			FromId:      s.handler.GetID().String(),
			LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
		}, nil
	}

	// Return empty response if we don't respond (lower priority)
	return &pb.ElectionResponse{}, nil
}

func (s *NodeServer) Coordinator(ctx context.Context, req *pb.CoordinatorRequest) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	fromID := ParseUUID(req.FromId)
	s.handler.HandleCoordinator(fromID)

	return &pb.Empty{}, nil
}

// ============================================================================
// Discovery
// ============================================================================

func (s *NodeServer) Discovery(ctx context.Context, req *pb.DiscoveryRequest) (*pb.DiscoveryResponse, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	peers := s.handler.GetPeers()
	peerPtrs := make([]*common.NodeInfo, len(peers))
	for i := range peers {
		peerPtrs[i] = &peers[i]
	}

	// Add self to the list
	myInfo := s.handler.GetNodeInfo()
	peerPtrs = append(peerPtrs, myInfo)

	return &pb.DiscoveryResponse{
		Peers:       NodeInfoSliceToProto(peerPtrs),
		LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
	}, nil
}

// ============================================================================
// Node Join/Leave
// ============================================================================

func (s *NodeServer) NodeJoined(ctx context.Context, req *pb.NodeJoinedRequest) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	nodeInfo := ProtoToNodeInfo(req.NodeInfo)
	if nodeInfo != nil {
		s.handler.AddPeer(nodeInfo)
		s.logger.Printf("Node joined: %s (%s:%d)", nodeInfo.ID, nodeInfo.IP, nodeInfo.Port)
	}

	return &pb.Empty{}, nil
}

func (s *NodeServer) NodeLeft(ctx context.Context, req *pb.NodeLeftRequest) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	nodeID := ParseUUID(req.NodeId)
	s.handler.RemovePeer(nodeID)
	s.logger.Printf("Node left: %s", nodeID)

	return &pb.Empty{}, nil
}

// ============================================================================
// Distributed Locking
// ============================================================================

func (s *NodeServer) RequestLock(ctx context.Context, req *pb.LockRequestMsg) (*pb.LockAck, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	lockReq := ProtoToLockRequest(req)
	fromID := ParseUUID(req.NodeId)

	s.handler.HandleLockRequest(lockReq, fromID)

	return &pb.LockAck{
		ResourceId:  req.ResourceId,
		FromId:      s.handler.GetID().String(),
		LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
	}, nil
}

func (s *NodeServer) AcquireLock(ctx context.Context, req *pb.LockAcquiredMsg) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	fromID := ParseUUID(req.FromId)
	s.handler.HandleLockAcquired(req.ResourceId, fromID)

	return &pb.Empty{}, nil
}

func (s *NodeServer) ReleaseLock(ctx context.Context, req *pb.LockReleaseMsg) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	fromID := ParseUUID(req.FromId)
	s.handler.HandleLockRelease(req.ResourceId, fromID)

	return &pb.Empty{}, nil
}

func (s *NodeServer) AbortLock(ctx context.Context, req *pb.LockAbortMsg) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	fromID := ParseUUID(req.FromId)
	s.handler.HandleLockAbort(fromID)

	return &pb.Empty{}, nil
}

// ============================================================================
// File Operations
// ============================================================================

func (s *NodeServer) DeleteFile(ctx context.Context, req *pb.FileDeleteRequest) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	s.handler.HandleFileDelete(req.Hash)

	return &pb.Empty{}, nil
}

func (s *NodeServer) StoreFile(ctx context.Context, req *pb.FileStoreRequest) (*pb.FileStoreResponse, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	storeReq := &common.FileStoreRequest{
		FileID:      ParseUUID(req.FileId),
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        req.Size,
		Data:        req.Data,
	}

	hash, err := s.handler.HandleFileStore(storeReq)
	if err != nil {
		return nil, err
	}

	return &pb.FileStoreResponse{
		Hash:   hash,
		FileId: req.FileId,
	}, nil
}

// ============================================================================
// Metadata Sync
// ============================================================================

func (s *NodeServer) GetMetadata(ctx context.Context, req *pb.MetadataReq) (*pb.MetadataResp, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	fromID := ParseUUID(req.FromId)
	resp := s.handler.HandleMetadataRequest(req.RequestId, fromID)

	if resp == nil {
		return &pb.MetadataResp{
			RequestId:   req.RequestId,
			NodeId:      s.handler.GetID().String(),
			Files:       []*pb.NodeFileMetadata{},
			LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
		}, nil
	}

	return &pb.MetadataResp{
		RequestId:   resp.RequestID,
		NodeId:      resp.NodeID.String(),
		Files:       NodeFileMetadataSliceToProto(resp.Files),
		LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
	}, nil
}

// ============================================================================
// File Replication
// ============================================================================

func (s *NodeServer) ReplicateFile(ctx context.Context, req *pb.ReplicateFileReq) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	targetNodes := make([]uuid.UUID, len(req.TargetNodes))
	for i, t := range req.TargetNodes {
		targetNodes[i] = ParseUUID(t)
	}

	replicateReq := &common.ReplicateFileRequest{
		FileID:      ParseUUID(req.FileId),
		Filename:    req.Filename,
		Hash:        req.Hash,
		SourceNode:  ParseUUID(req.SourceNode),
		TargetNodes: targetNodes,
	}

	s.handler.HandleReplicateFile(replicateReq)

	return &pb.Empty{}, nil
}

// ============================================================================
// API File Delete
// ============================================================================

func (s *NodeServer) APIDeleteFile(ctx context.Context, req *pb.APIFileDeleteReq) (*pb.APIFileDeleteResp, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	// Only master handles this
	if s.handler.GetRole() != common.RoleMaster {
		return &pb.APIFileDeleteResp{
			RequestId:   req.RequestId,
			Success:     false,
			Error:       "node is not master",
			LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
		}, nil
	}

	fileID := ParseUUID(req.FileId)
	hashRing := s.handler.GetHashRing()
	if hashRing == nil {
		return &pb.APIFileDeleteResp{
			RequestId:   req.RequestId,
			Success:     false,
			Error:       "no hash ring available",
			LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
		}, nil
	}

	// Get nodes that should have this file
	nodes := hashRing.FindNodesForFile(fileID, 3)

	// Delete from all nodes (this will be done via the client connection manager)
	s.logger.Printf("API: Delete request for file %s, found %d nodes", req.FileId, len(nodes))

	// The actual delete will be triggered by the node's client
	// For now, just return success and let the caller handle the actual deletion
	return &pb.APIFileDeleteResp{
		RequestId:   req.RequestId,
		Success:     true,
		LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
	}, nil
}

// ============================================================================
// Streaming Session Coordination
// ============================================================================

func (s *NodeServer) InitStreamUpload(ctx context.Context, req *pb.StreamUploadInitReq) (*pb.StreamUploadReadyResp, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	initReq := &common.StreamUploadInitRequest{
		RequestID:   req.RequestId,
		FileID:      ParseUUID(req.FileId),
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        req.Size,
	}

	resp := s.handler.HandleStreamUploadInit(initReq)
	if resp == nil {
		return &pb.StreamUploadReadyResp{
			RequestId:   req.RequestId,
			Success:     false,
			Error:       "failed to initialize upload",
			LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
		}, nil
	}

	return &pb.StreamUploadReadyResp{
		RequestId:    resp.RequestID,
		SessionId:    resp.SessionID,
		StorageNodes: StorageNodeAddrSliceToProto(resp.StorageNodes),
		Success:      resp.Success,
		Error:        resp.Error,
		LogicalTime:  int32(s.handler.IncrementAndGetLogicalTime()),
	}, nil
}

func (s *NodeServer) InitStreamDownload(ctx context.Context, req *pb.StreamDownloadInitReq) (*pb.StreamDownloadReadyResp, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	initReq := &common.StreamDownloadInitRequest{
		RequestID: req.RequestId,
		FileID:    ParseUUID(req.FileId),
		Hash:      req.Hash,
	}

	resp := s.handler.HandleStreamDownloadInit(initReq)
	if resp == nil {
		return &pb.StreamDownloadReadyResp{
			RequestId:   req.RequestId,
			Success:     false,
			Error:       "failed to initialize download",
			LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
		}, nil
	}

	return &pb.StreamDownloadReadyResp{
		RequestId:   resp.RequestID,
		StorageAddr: resp.StorageAddr,
		Filename:    resp.Filename,
		Size:        resp.Size,
		Hash:        resp.Hash,
		Success:     resp.Success,
		Error:       resp.Error,
		LogicalTime: int32(s.handler.IncrementAndGetLogicalTime()),
	}, nil
}

func (s *NodeServer) StreamStoreAck(ctx context.Context, req *pb.StreamStoreAckMsg) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	ack := &common.StreamStoreAck{
		SessionID: req.SessionId,
		FileID:    ParseUUID(req.FileId),
		Hash:      req.Hash,
		NodeID:    ParseUUID(req.NodeId),
		Success:   req.Success,
		Error:     req.Error,
	}

	s.handler.HandleStreamStoreAck(ack)

	return &pb.Empty{}, nil
}

func (s *NodeServer) StreamUploadReady(ctx context.Context, req *pb.StreamUploadReadyMsg) (*pb.Empty, error) {
	s.handler.UpdateAndGetLogicalTime(int(req.LogicalTime))

	s.handler.HandleStreamUploadReady(
		req.SessionId,
		ParseUUID(req.FileId),
		req.Filename,
		req.ContentType,
		req.Size,
	)

	return &pb.Empty{}, nil
}

// ============================================================================
// Helper for sorting lock requests (used by node handler)
// ============================================================================

func CompareLockRequests(a, b *common.LockRequest) bool {
	if a.LogicalTime != b.LogicalTime {
		return a.LogicalTime < b.LogicalTime
	}
	return a.NodeID.String() < b.NodeID.String()
}

func SortLockQueue(queue []*common.LockRequest) {
	sort.SliceStable(queue, func(i, j int) bool {
		return CompareLockRequests(queue[i], queue[j])
	})
}
