package grpc

import (
	"dfs-backend/dfs/common"
	pb "dfs-backend/proto"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NodeInfo conversions

func NodeInfoToProto(ni *common.NodeInfo) *pb.NodeInfo {
	if ni == nil {
		return nil
	}
	return &pb.NodeInfo{
		Id:            ni.ID.String(),
		Ip:            ni.IP,
		Port:          int32(ni.Port),
		Role:          NodeRoleToProto(ni.Role),
		Status:        NodeStatusToProto(ni.Status),
		Priority:      int32(ni.Priority),
		LastHeartbeat: timestamppb.New(ni.LastHeartbeat),
	}
}

func ProtoToNodeInfo(pni *pb.NodeInfo) *common.NodeInfo {
	if pni == nil {
		return nil
	}
	id, _ := uuid.Parse(pni.Id)
	var lastHeartbeat time.Time
	if pni.LastHeartbeat != nil {
		lastHeartbeat = pni.LastHeartbeat.AsTime()
	}
	return &common.NodeInfo{
		ID:            id,
		IP:            pni.Ip,
		Port:          int(pni.Port),
		Role:          ProtoToNodeRole(pni.Role),
		Status:        ProtoToNodeStatus(pni.Status),
		Priority:      int(pni.Priority),
		LastHeartbeat: lastHeartbeat,
	}
}

// NodeRole conversions

func NodeRoleToProto(role common.NodeRole) pb.NodeRole {
	switch role {
	case common.RoleMaster:
		return pb.NodeRole_NODE_ROLE_MASTER
	case common.RoleStorage:
		return pb.NodeRole_NODE_ROLE_STORAGE
	default:
		return pb.NodeRole_NODE_ROLE_UNSPECIFIED
	}
}

func ProtoToNodeRole(role pb.NodeRole) common.NodeRole {
	switch role {
	case pb.NodeRole_NODE_ROLE_MASTER:
		return common.RoleMaster
	case pb.NodeRole_NODE_ROLE_STORAGE:
		return common.RoleStorage
	default:
		return common.RoleStorage // default to storage
	}
}

// NodeStatus conversions

func NodeStatusToProto(status common.NodeStatus) pb.NodeStatus {
	switch status {
	case common.StatusOnline:
		return pb.NodeStatus_NODE_STATUS_ONLINE
	case common.StatusOffline:
		return pb.NodeStatus_NODE_STATUS_OFFLINE
	case common.StatusStarting:
		return pb.NodeStatus_NODE_STATUS_STARTING
	case common.StatusStopping:
		return pb.NodeStatus_NODE_STATUS_STOPPING
	default:
		return pb.NodeStatus_NODE_STATUS_UNSPECIFIED
	}
}

func ProtoToNodeStatus(status pb.NodeStatus) common.NodeStatus {
	switch status {
	case pb.NodeStatus_NODE_STATUS_ONLINE:
		return common.StatusOnline
	case pb.NodeStatus_NODE_STATUS_OFFLINE:
		return common.StatusOffline
	case pb.NodeStatus_NODE_STATUS_STARTING:
		return common.StatusStarting
	case pb.NodeStatus_NODE_STATUS_STOPPING:
		return common.StatusStopping
	default:
		return common.StatusOffline
	}
}

// LockStatus conversions

func LockStatusToProto(status common.LockStatus) pb.LockStatus {
	switch status {
	case common.StatusPending:
		return pb.LockStatus_LOCK_STATUS_PENDING
	case common.StatusGranted:
		return pb.LockStatus_LOCK_STATUS_GRANTED
	case common.StatusReleased:
		return pb.LockStatus_LOCK_STATUS_RELEASED
	default:
		return pb.LockStatus_LOCK_STATUS_UNSPECIFIED
	}
}

func ProtoToLockStatus(status pb.LockStatus) common.LockStatus {
	switch status {
	case pb.LockStatus_LOCK_STATUS_PENDING:
		return common.StatusPending
	case pb.LockStatus_LOCK_STATUS_GRANTED:
		return common.StatusGranted
	case pb.LockStatus_LOCK_STATUS_RELEASED:
		return common.StatusReleased
	default:
		return common.StatusPending
	}
}

// LockRequest conversions

func LockRequestToProto(lr *common.LockRequest) *pb.LockRequestMsg {
	if lr == nil {
		return nil
	}
	return &pb.LockRequestMsg{
		ResourceId:  lr.ResourceID,
		NodeId:      lr.NodeID.String(),
		LogicalTime: int32(lr.LogicalTime),
		Status:      LockStatusToProto(lr.Status),
		RequestedAt: timestamppb.New(lr.RequestedAt),
	}
}

func ProtoToLockRequest(plr *pb.LockRequestMsg) *common.LockRequest {
	if plr == nil {
		return nil
	}
	nodeID, _ := uuid.Parse(plr.NodeId)
	var requestedAt time.Time
	if plr.RequestedAt != nil {
		requestedAt = plr.RequestedAt.AsTime()
	}
	return &common.LockRequest{
		ResourceID:  plr.ResourceId,
		NodeID:      nodeID,
		LogicalTime: int(plr.LogicalTime),
		Status:      ProtoToLockStatus(plr.Status),
		RequestedAt: requestedAt,
	}
}

// NodeFileMetadata conversions

func NodeFileMetadataToProto(nfm *common.NodeFileMetadata) *pb.NodeFileMetadata {
	if nfm == nil {
		return nil
	}
	return &pb.NodeFileMetadata{
		FileId:      nfm.FileID.String(),
		Filename:    nfm.Filename,
		Hash:        nfm.Hash,
		Size:        nfm.Size,
		ContentType: nfm.ContentType,
	}
}

func ProtoToNodeFileMetadata(pnfm *pb.NodeFileMetadata) *common.NodeFileMetadata {
	if pnfm == nil {
		return nil
	}
	fileID, _ := uuid.Parse(pnfm.FileId)
	return &common.NodeFileMetadata{
		FileID:      fileID,
		Filename:    pnfm.Filename,
		Hash:        pnfm.Hash,
		Size:        pnfm.Size,
		ContentType: pnfm.ContentType,
	}
}

// StorageNodeAddr conversions

func StorageNodeAddrToProto(sna *common.StorageNodeAddr) *pb.StorageNodeAddr {
	if sna == nil {
		return nil
	}
	return &pb.StorageNodeAddr{
		NodeId: sna.NodeID.String(),
		Addr:   sna.Addr,
	}
}

func ProtoToStorageNodeAddr(psna *pb.StorageNodeAddr) *common.StorageNodeAddr {
	if psna == nil {
		return nil
	}
	nodeID, _ := uuid.Parse(psna.NodeId)
	return &common.StorageNodeAddr{
		NodeID: nodeID,
		Addr:   psna.Addr,
	}
}

// Slice conversions

func NodeInfoSliceToProto(nis []*common.NodeInfo) []*pb.NodeInfo {
	result := make([]*pb.NodeInfo, len(nis))
	for i, ni := range nis {
		result[i] = NodeInfoToProto(ni)
	}
	return result
}

func ProtoToNodeInfoSlice(pnis []*pb.NodeInfo) []*common.NodeInfo {
	result := make([]*common.NodeInfo, len(pnis))
	for i, pni := range pnis {
		result[i] = ProtoToNodeInfo(pni)
	}
	return result
}

func NodeFileMetadataSliceToProto(nfms []common.NodeFileMetadata) []*pb.NodeFileMetadata {
	result := make([]*pb.NodeFileMetadata, len(nfms))
	for i, nfm := range nfms {
		result[i] = NodeFileMetadataToProto(&nfm)
	}
	return result
}

func StorageNodeAddrSliceToProto(snas []common.StorageNodeAddr) []*pb.StorageNodeAddr {
	result := make([]*pb.StorageNodeAddr, len(snas))
	for i, sna := range snas {
		result[i] = StorageNodeAddrToProto(&sna)
	}
	return result
}

// Helper to parse UUID safely
func ParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
