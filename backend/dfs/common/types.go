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

	MessageLockRequest MessageType = "lock_request"
	MessageLockAck     MessageType = "lock_ack"
	MessageLockRelease MessageType = "lock_release"
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

type LockRequest struct {
	ResourceID  string     `json:"resource_id"`
	NodeID      uuid.UUID  `json:"node_id"`
	LogicalTime int        `json:"logical_time"`
	Status      LockStatus `json:"lock_status"`
}
