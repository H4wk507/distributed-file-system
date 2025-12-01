package services

import (
	"dfs-backend/internal/database"
	"dfs-backend/internal/models"
	"fmt"
)

type FileService struct {
	db *database.DB
}

func NewFileService(db *database.DB) *FileService {
	return &FileService{db: db}
}

func (s *FileService) GetFileHashByName(filename string) (string, error) {
	var hash string
	query := `select hash from files where filename = $1`
	if err := s.db.Get(hash, query, filename); err != nil {
		return "", fmt.Errorf("failed to get hash for file %s: %w", filename, err)
	}

	return hash, nil
}

func (s *FileService) CreateFile(file *models.File) error {
	query := `
		insert into files (id, name, size, hash, content_type, owner_id)
		values ($1, $2, $3, $4, $5, $6)
	 `
	_, err := s.db.Exec(query, file.ID, file.Name, file.Size, file.Hash, file.ContentType, file.OwnerID)
	return err
}

func (s *FileService) DeleteFileByName(filename string) error {
	query := `delete from files where name = $1`
	_, err := s.db.Exec(query, filename)
	return err
}
