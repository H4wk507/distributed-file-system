package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID uuid.UUID `db:"id" json:"id"`

	Name        string    `db:"name" json:"name"`
	Size        int64     `db:"size" json:"size"`
	Hash        string    `db:"hash" json:"hash"`
	ContentType string    `db:"content_type" json:"content_type"`
	OwnerID     uuid.UUID `db:"owner_id" json:"owner_id"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
