package models

import (
    "time"
    
    "github.com/google/uuid"
)

type File struct {
    ID        uuid.UUID `db:"id" json:"id"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
    Name      string    `db:"name" json:"name"`
    Size      int64     `db:"size" json:"size"`
}
