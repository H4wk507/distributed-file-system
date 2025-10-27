package services

import (
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/models"
	"fmt"

	"github.com/google/uuid"
)

type FileService struct {
	db *database.DB
}

func NewFileService(db *database.DB) *FileService {
	return &FileService{db: db}
}

func (s *FileService) GetByID(id uuid.UUID) (*dto.FileResponse, error) {
	var file models.File
	query := `SELECT id, name, size FROM files WHERE id = $1`
	err := s.db.Get(&file, query, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return &dto.FileResponse{
		ID:   file.ID,
		Name: file.Name,
		Size: file.Size,
	}, nil
}
