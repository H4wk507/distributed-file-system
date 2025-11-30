package storage

import (
	"time"

	"github.com/google/uuid"
)

type FileMetadata struct {
	FileID      uuid.UUID `json:"file_id"`
	Filename    string    `json:"filename"`
	Hash        string    `json:"hash"`
	Size        int64     `json:"size"` // in bytes
	ContentType string    `json:"content_type"`
	StoredAt    time.Time `json:"stored_at"`
	Status      string    `json:"status"` // TODO: enum
}
