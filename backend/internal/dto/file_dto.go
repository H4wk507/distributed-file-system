package dto

import (
    "github.com/google/uuid"
)

type FileResponse struct {
    ID           uuid.UUID `json:"id"`
    Name         string    `json:"name"`
    Size         int64     `json:"size"`
}

