package dto

import (
	"time"

	"github.com/google/uuid"
)

type FileUploadResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
}

type FileItem struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	Hash        string    `json:"hash"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FileListResponse struct {
	Files   []FileItem `json:"files"`
	Total   int        `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
}
