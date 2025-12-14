package common

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type NodeRole string

const (
	RoleMaster  NodeRole = "master"
	RoleStorage NodeRole = "storage"
)

type NodeStatus string

const (
	StatusOnline   NodeStatus = "online"
	StatusOffline  NodeStatus = "offline"
	StatusStarting NodeStatus = "starting"
	StatusStopping NodeStatus = "stopping"
)

type NodeInfo struct {
	ID            uuid.UUID  `json:"id"`
	IP            string     `json:"ip"`
	Port          int        `json:"port"`
	Role          NodeRole   `json:"role"`
	Status        NodeStatus `json:"status"`
	Priority      int        `json:"priority"`
	LastHeartbeat time.Time  `json:"last_heartbeat"`
}

type MessageType string

const (
	MessageHeartbeat   MessageType = "heartbeat"
	MessageElection    MessageType = "election"
	MessageDiscovery   MessageType = "discovery"
	MessageNodeJoined  MessageType = "node_joined"
	MessageNodeLeft    MessageType = "node_left"
	MessageCoordinator MessageType = "coordinator"
	MessageOK          MessageType = "ok"

	MessageLockRequest  MessageType = "lock_request"
	MessageLockAck      MessageType = "lock_ack"
	MessageLockAcquired MessageType = "lock_acquired"
	MessageLockRelease  MessageType = "lock_release"
	MessageLockAbort    MessageType = "lock_abort"

	MessageFileStore            MessageType = "file_store"
	MessageFileStoreAck         MessageType = "file_storage_ack"
	MessageFileRetrieve         MessageType = "file_retrieve"
	MessageFileRetrieveResponse MessageType = "file_retrieve_response"
	MessageFileDelete           MessageType = "file_delete"

	MessageMetadataRequest  MessageType = "metadata_request"
	MessageMetadataResponse MessageType = "metadata_response"
	MessageReplicateFile    MessageType = "replicate_file"

	MessageAPIFileDelete MessageType = "api_file_delete"

	MessageStreamUploadInit    MessageType = "stream_upload_init"
	MessageStreamUploadReady   MessageType = "stream_upload_ready"
	MessageStreamDownloadInit  MessageType = "stream_download_init"
	MessageStreamDownloadReady MessageType = "stream_download_ready"
	MessageStreamStoreAck      MessageType = "stream_store_ack"
)

type Message struct {
	Type    MessageType     `json:"type"`
	From    uuid.UUID       `json:"from"`
	To      uuid.UUID       `json:"to"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type MessageWithTime struct {
	Message     Message `json:"message"`
	LogicalTime int     `json:"logical_time"`
}

type LockStatus string

const (
	StatusPending  LockStatus = "pending"
	StatusGranted  LockStatus = "granted"
	StatusReleased LockStatus = "released"
)

// TODO: Co jeśli upload pliku zajmie więcej niż 30 sekund? LOCK_HEARTBEAT bardziej elastyczne
const LOCK_TIMEOUT = 30 * time.Second

type LockRequest struct {
	ResourceID  string     `json:"resource_id"`
	NodeID      uuid.UUID  `json:"node_id"`
	LogicalTime int        `json:"logical_time"`
	Status      LockStatus `json:"lock_status"`
	RequestedAt time.Time  `json:"requested_at"`
}

type GraphNodeStatus string

const (
	GraphNodeStatusNotVisited GraphNodeStatus = "not_visited"
	GraphNodeStatusVisiting   GraphNodeStatus = "visiting"
	GraphNodeStatusVisited    GraphNodeStatus = "visited"
)

type PendingUpload struct {
	FileID        uuid.UUID
	Filename      string
	Hash          string
	ExpectedNodes []uuid.UUID
	ReceivedAcks  map[uuid.UUID]string
	CreatedAt     time.Time
}

type FileStoreRequest struct {
	FileID      uuid.UUID `json:"file_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Data        []byte    `json:"data"`
}

type FileStoreResponse struct {
	Hash   string    `json:"hash"`
	FileID uuid.UUID `json:"file_id"`
}

type FileRetrieveRequest struct {
	RequestID string `json:"request_id"`
	Hash      string `json:"hash"`
}

type FileRetrieveResponse struct {
	RequestID string    `json:"request_id"`
	FileID    uuid.UUID `json:"file_id"`
	Filename  string    `json:"filename"`
	Hash      string    `json:"hash"`
	Data      []byte    `json:"data"`
	Error     string    `json:"error,omitempty"`
}

type GlobalFileInfo struct {
	FileID      uuid.UUID
	Filename    string
	Hash        string
	Size        int64 // in bytes
	ContentType string
	Replicas    []uuid.UUID
}

type MetadataSyncState struct {
	RequestID     string
	ExpectedNodes int
	ReceivedNodes map[uuid.UUID]bool
	Done          chan struct{}
}

type MetadataRequest struct {
	RequestID string `json:"request_id"`
}

type NodeFileMetadata struct {
	FileID      uuid.UUID `json:"file_id"`
	Filename    string    `json:"filename"`
	Hash        string    `json:"hash"`
	Size        int64     `json:"size"` // in bytes
	ContentType string    `json:"content_type"`
}

type MetadataResponse struct {
	RequestID string             `json:"request_id"`
	NodeID    uuid.UUID          `json:"node_id"`
	Files     []NodeFileMetadata `json:"files"`
}

type ReplicateFileRequest struct {
	FileID      uuid.UUID   `json:"file_id"`
	Filename    string      `json:"filename"`
	Hash        string      `json:"hash"`
	SourceNode  uuid.UUID   `json:"source_node"`
	TargetNodes []uuid.UUID `json:"target_nodes"`
}

type APIFileDeleteRequest struct {
	RequestID string    `json:"request_id"`
	FileID    uuid.UUID `json:"file_id"`
	Hash      string    `json:"hash"`
}

type APIFileDeleteResponse struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type StreamUploadInitRequest struct {
	RequestID   string    `json:"request_id"`
	FileID      uuid.UUID `json:"file_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
}

type StorageNodeAddr struct {
	NodeID uuid.UUID `json:"node_id"`
	Addr   string    `json:"addr"` // ip:port
}

type StreamUploadReadyResponse struct {
	RequestID    string            `json:"request_id"`
	SessionID    string            `json:"session_id"`
	StorageNodes []StorageNodeAddr `json:"storage_nodes"`
	Success      bool              `json:"success"`
	Error        string            `json:"error,omitempty"`
}

type StreamDownloadInitRequest struct {
	RequestID string    `json:"request_id"`
	FileID    uuid.UUID `json:"file_id"`
	Hash      string    `json:"hash"`
}

type StreamDownloadReadyResponse struct {
	RequestID   string `json:"request_id"`
	StorageAddr string `json:"storage_addr"` // ip:port
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Hash        string `json:"hash"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

type StreamStoreAck struct {
	SessionID string    `json:"session_id"`
	FileID    uuid.UUID `json:"file_id"`
	Hash      string    `json:"hash"`
	NodeID    uuid.UUID `json:"node_id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

type StreamSession struct {
	SessionID     string
	FileID        uuid.UUID
	Filename      string
	ContentType   string
	Size          int64
	ExpectedNodes []uuid.UUID
	ReceivedAcks  map[uuid.UUID]string // nodeID -> hash
	CreatedAt     time.Time
	Done          chan struct{}
}
