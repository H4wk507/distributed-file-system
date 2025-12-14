package dto

import (
	"time"

	"github.com/google/uuid"
)

type FileItem struct {
	ID            uuid.UUID `json:"id"`
	Filename      string    `json:"filename"`
	Size          int64     `json:"size"`
	ContentType   string    `json:"content_type"`
	Hash          string    `json:"hash"`
	OwnerID       uuid.UUID `json:"owner_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ReplicasCount int       `json:"replicas_count"`
}

type FileListResponse struct {
	Files   []FileItem `json:"files"`
	Total   int        `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
}

type ChunkedUploadInitRequest struct {
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	TotalChunks int    `json:"totalChunks"`
	ContentType string `json:"contentType"`
}

type ChunkedUploadInitResponse struct {
	SessionID string `json:"sessionId"`
}

type ChunkedUploadFinalizeRequest struct {
	SessionID string `json:"sessionId"`
}

type ChunkedUploadFinalizeResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
}
