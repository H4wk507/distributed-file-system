package election

import (
	"context"
	"dfs-backend/dfs/common"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Elector interface {
	Start(ctx context.Context) error
	Stop() error
	StartElection(ctx context.Context) error
	HandleMessage(ctx context.Context, msg common.Message) error
	NotifyCoordinatorDead(coordinatorID uuid.UUID)
}

type ElectionMessage struct {
	Type      MessageType     `json:"type"`
	From      uuid.UUID       `json:"from"`
	To        uuid.UUID       `json:"to"`
	Priority  int             `json:"priority"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type MessageType string

const (
	MessageElection    MessageType = "election"
	MessageOK          MessageType = "ok"
	MessageCoordinator MessageType = "coordinator"
)
