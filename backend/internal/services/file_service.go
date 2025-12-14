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

func (s *FileService) CreateReplicas(fileID uuid.UUID, nodeIDs []uuid.UUID) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	query := `INSERT INTO replicas (file_id, node_id, status) VALUES ($1, $2, 'synced')`
	for _, nodeID := range nodeIDs {
		if _, err := s.db.Exec(query, fileID, nodeID); err != nil {
			return fmt.Errorf("failed to create replica for node %s: %w", nodeID, err)
		}
	}
	return nil
}

func (s *FileService) ListFilesPaginated(ownerID uuid.UUID, page, perPage int) ([]models.File, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	var total int
	countQuery := `SELECT COUNT(*) FROM files WHERE owner_id = $1`
	if err := s.db.Get(&total, countQuery, ownerID); err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	var files []models.File
	query := `
		SELECT f.id, f.name, f.size, f.hash, f.content_type, f.owner_id, f.created_at, f.updated_at, count(r.id) as replicas_count
		FROM files f
		left join replicas r on f.id = r.file_id and r.status = 'synced'
		WHERE f.owner_id = $1 
		GROUP BY f.id, f.name, f.size, f.hash, f.content_type, f.owner_id, f.created_at, f.updated_at
		ORDER BY f.created_at DESC 
		LIMIT $2 OFFSET $3
	`
	if err := s.db.Select(&files, query, ownerID, perPage, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}

	return files, total, nil
}
