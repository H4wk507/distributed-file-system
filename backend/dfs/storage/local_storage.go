package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LocalStorage struct {
	basePath string
	index    map[string]FileMetadata
	mutex    sync.RWMutex
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		index:    make(map[string]FileMetadata),
	}
}

func (s *LocalStorage) SaveFile(fileID uuid.UUID, filename string, contentType string, data io.Reader) (*FileMetadata, error) {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return nil, err
	}
	tmpPath := filepath.Join(s.basePath, fmt.Sprintf("temp-%d", time.Now().UnixNano()))
	defer os.Remove(tmpPath)

	f, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}

	contentHash := sha256.New()
	buf := make([]byte, 1*1024*1024) // 1MB
	size, err := io.CopyBuffer(io.MultiWriter(f, contentHash), data, buf)
	f.Close()
	if err != nil {
		return nil, err
	}

	contentHashStr := hex.EncodeToString(contentHash.Sum(nil))
	finalPath := s.filePath(contentHashStr)

	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return nil, err
	}

	metadata := FileMetadata{
		FileID:      fileID,
		Filename:    filename,
		Hash:        contentHashStr,
		Size:        size,
		ContentType: contentType,
		StoredAt:    time.Now(),
		Status:      "complete",
	}

	s.mutex.Lock()
	s.index[contentHashStr] = metadata
	s.mutex.Unlock()

	if err := s.SaveIndex(); err != nil {
		return nil, fmt.Errorf("file saved but index update failed: %w", err)
	}
	return &metadata, nil
}

func (s *LocalStorage) DeleteFile(hash string) error {
	path := s.filePath(hash)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.mutex.Lock()
	delete(s.index, hash)
	s.mutex.Unlock()

	if err := s.SaveIndex(); err != nil {
		return fmt.Errorf("file deleted but index update failed: %w", err)
	}

	return nil
}

func (s *LocalStorage) GetFile(hash string) (io.ReadCloser, *FileMetadata, error) {
	s.mutex.RLock()
	metadata, exists := s.index[hash]
	s.mutex.RUnlock()

	if !exists {
		return nil, nil, fmt.Errorf("hash %s does not exist in the index", hash)
	}

	f, err := os.Open(s.filePath(hash))
	if err != nil {
		return nil, nil, err
	}
	return f, &metadata, nil
}

func (s *LocalStorage) LoadIndex() error {
	data, err := os.ReadFile(s.indexPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	return json.Unmarshal(data, &s.index)
}

func (s *LocalStorage) SaveIndex() error {
	s.mutex.RLock()
	jsonStr, err := json.Marshal(s.index)
	s.mutex.RUnlock()
	if err != nil {
		return err
	}

	tmpPath := s.indexPath() + ".tmp"
	if err := os.WriteFile(tmpPath, jsonStr, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.indexPath())
}

func (s *LocalStorage) GetAllMetadata() []FileMetadata {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]FileMetadata, 0, len(s.index))
	for _, meta := range s.index {
		result = append(result, meta)
	}
	return result
}

func (s *LocalStorage) filePath(hash string) string {
	return filepath.Join(s.basePath, "data", hash[:2], fmt.Sprintf("%s.dat", hash))
}

func (s *LocalStorage) indexPath() string {
	return filepath.Join(s.basePath, "index.json")
}
