package services

import (
	"dfs-backend/internal/database"
)

type FileService struct {
	db *database.DB
}

func NewFileService(db *database.DB) *FileService {
	return &FileService{db: db}
}
