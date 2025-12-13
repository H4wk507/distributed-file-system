package services

import (
	"dfs-backend/internal/database"
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

func (s *FileService) GetFileByID(fileID uuid.UUID) (*models.File, error) {
	var file models.File
	query := `select id, name, size, hash, content_type, owner_id from files where id = $1`
	if err := s.db.Get(&file, query, fileID); err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", fileID, err)
	}
	return &file, nil
}

func (s *FileService) CreateFile(file *models.File) error {
	query := `
		insert into files (id, name, size, hash, content_type, owner_id)
		values ($1, $2, $3, $4, $5, $6)
	 `
	_, err := s.db.Exec(query, file.ID, file.Name, file.Size, file.Hash, file.ContentType, file.OwnerID)
	return err
}

func (s *FileService) DeleteFileByID(fileID uuid.UUID) error {
	query := `delete from files where id = $1`
	_, err := s.db.Exec(query, fileID)
	return err
}
