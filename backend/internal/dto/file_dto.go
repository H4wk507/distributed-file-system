package dto

import (
	"github.com/google/uuid"
)

type FileUploadResponse struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
}
