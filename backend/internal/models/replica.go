package models

import (
	"time"

	"github.com/google/uuid"
)

type ReplicaStatus = string

const (
	Synced  ReplicaStatus = "synced"
	Syncing ReplicaStatus = "syncing"
	Failed  ReplicaStatus = "failed"
)

type Replica struct {
	ID uuid.UUID `db:"id" json:"id"`

	FileID uuid.UUID     `db:"file_id" json:"file_id"`
	NodeID uuid.UUID     `db:"node_id" json:"node_id"`
	Status ReplicaStatus `db:"status" json:"status"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
