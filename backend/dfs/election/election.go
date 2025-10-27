package election

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ElectionStatus string

const (
	StatusIdle       ElectionStatus = "idle"
	StatusInProgress ElectionStatus = "in_progress"
	StatusCompleted  ElectionStatus = "completed"
)

type Elector interface {
	StartElection(ctx context.Context) error
	HandleElectionMessage(msg ElectionMessage) error
	GetCurrentLeader() (uuid.UUID, error)
	IsLeader() bool
}

type ElectionMessage struct {
	Type      MessageType `json:"type"`
	From      uuid.UUID   `json:"from"`
	To        uuid.UUID   `json:"to"`
	Priority  int         `json:"priority"`
	Timestamp time.Time   `json:"timestamp"`
}

type MessageType string

const (
	MessageElection    MessageType = "election"
	MessageOK          MessageType = "ok"
	MessageCoordinator MessageType = "coordinator"
)
