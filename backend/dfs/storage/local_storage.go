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
	tmpPath := filepath.Join(s.basePath, fmt.Sprintf("temp-%d", time.Now().UnixNano()))
	defer os.Remove(tmpPath)

	f, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	buf := make([]byte, 1*1024*1024) // 1MB
	size, err := io.CopyBuffer(io.MultiWriter(f, hash), data, buf)
	f.Close()
	if err != nil {
		return nil, err
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))
	finalPath := s.filePath(hashStr)

	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return nil, err
	}

	metadata := FileMetadata{
		FileID:      fileID,
		Filename:    filename,
		Hash:        hashStr,
		Size:        size,
		ContentType: contentType,
		StoredAt:    time.Now(),
		Status:      "complete",
	}

	s.mutex.Lock()
	s.index[hashStr] = metadata
	s.mutex.Unlock()

	if err := s.SaveIndex(); err != nil {
		return nil, fmt.Errorf("file saved buy index update failed: %w", err)
	}
	return &metadata, nil
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

func (s *LocalStorage) filePath(hash string) string {
	return filepath.Join(s.basePath, "data", hash[:2], fmt.Sprintf("%s.dat", hash))
}

func (s *LocalStorage) indexPath() string {
	return filepath.Join(s.basePath, "index.json")
}
