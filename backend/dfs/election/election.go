package election

import (
	"context"
	"dfs-backend/dfs/common"

	"github.com/google/uuid"
)

type Elector interface {
	Start(ctx context.Context) error
	Stop() error
	StartElection(ctx context.Context) error
	HandleMessage(ctx context.Context, msg common.Message) error
	NotifyCoordinatorDead(coordinatorID uuid.UUID)
}

type MessageType string

const (
	MessageElection    MessageType = "election"
	MessageOK          MessageType = "ok"
	MessageCoordinator MessageType = "coordinator"
)
