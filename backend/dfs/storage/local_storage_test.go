package storage

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func setupTestStorage(t *testing.T) (*LocalStorage, string) {
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	storage := NewLocalStorage(tmpDir)
	return storage, tmpDir
}

func cleanupTestStorage(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestNewLocalStorage(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	if storage.basePath != tmpDir {
		t.Errorf("expected basePath %s, got %s", tmpDir, storage.basePath)
	}

	if storage.index == nil {
		t.Error("index should be initialized")
	}
}

func TestSaveFile(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	fileID := uuid.New()
	filename := "test.txt"
	contentType := "text/plain"
	data := []byte("Hello, World!")

	meta, err := storage.SaveFile(fileID, filename, contentType, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	// Check metadata
	if meta.FileID != fileID {
		t.Errorf("expected FileID %s, got %s", fileID, meta.FileID)
	}
	if meta.Filename != filename {
		t.Errorf("expected Filename %s, got %s", filename, meta.Filename)
	}
	if meta.ContentType != contentType {
		t.Errorf("expected ContentType %s, got %s", contentType, meta.ContentType)
	}
	if meta.Size != int64(len(data)) {
		t.Errorf("expected Size %d, got %d", len(data), meta.Size)
	}
	if meta.Status != "complete" {
		t.Errorf("expected Status 'complete', got %s", meta.Status)
	}
	if meta.Hash == "" {
		t.Error("Hash should not be empty")
	}

	// Check file exists on disk
	filePath := storage.filePath(meta.Hash)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file should exist on disk")
	}

	// Check if metadata file exists on disk
	metadataFilePath := storage.metadataFilePath(meta.Hash)
	log.Print(metadataFilePath)
	if _, err := os.Stat(metadataFilePath); os.IsNotExist(err) {
		t.Error("metadata file should exist on disk")
	}

	// Check file is in index
	if _, exists := storage.index[meta.Hash]; !exists {
		t.Error("file should be in index")
	}
}

func TestGetFile(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// First save a file
	fileID := uuid.New()
	data := []byte("Test content for GetFile")
	meta, err := storage.SaveFile(fileID, "test.txt", "text/plain", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	// Now get the file
	reader, retrievedMeta, err := storage.GetFile(meta.Hash)
	if err != nil {
		t.Fatalf("GetFile failed: %v", err)
	}
	defer reader.Close()

	// Check metadata matches
	if retrievedMeta.FileID != fileID {
		t.Errorf("expected FileID %s, got %s", fileID, retrievedMeta.FileID)
	}

	// Check content
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read file content: %v", err)
	}
	if !bytes.Equal(content, data) {
		t.Errorf("content mismatch: expected %s, got %s", data, content)
	}
}

func TestGetFile_NotExists(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	_, _, err := storage.GetFile("nonexistent-hash")
	if err == nil {
		t.Error("GetFile should return error for nonexistent file")
	}
}

func TestDeleteFile(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// First save a file
	fileID := uuid.New()
	data := []byte("File to delete")
	meta, err := storage.SaveFile(fileID, "delete-me.txt", "text/plain", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	// Verify file exists
	filePath := storage.filePath(meta.Hash)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("file should exist before delete")
	}

	// Delete the file
	err = storage.DeleteFile(meta.Hash)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	// Verify file is gone from disk
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("file should not exist after delete")
	}

	// Verify metadata file is gone from disk
	metadataFilePath := storage.filePath(fmt.Sprintf("%s.metadata.json", meta.Hash))
	if _, err := os.Stat(metadataFilePath); !os.IsNotExist(err) {
		t.Error("metadata file should not exist on disk")
	}

	// Verify file is gone from index
	if _, exists := storage.index[meta.Hash]; exists {
		t.Error("file should not be in index after delete")
	}
}

func TestDeleteFile_NotExists(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// Delete nonexistent file should not error (idempotent)
	err := storage.DeleteFile("nonexistent-hash")
	if err != nil {
		t.Errorf("DeleteFile for nonexistent file should not error, got: %v", err)
	}
}

func TestDeleteFile_Idempotent(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// Save a file
	data := []byte("File to delete twice")
	meta, _ := storage.SaveFile(uuid.New(), "delete-twice.txt", "text/plain", bytes.NewReader(data))

	// Delete twice - should not error
	err := storage.DeleteFile(meta.Hash)
	if err != nil {
		t.Fatalf("first DeleteFile failed: %v", err)
	}

	err = storage.DeleteFile(meta.Hash)
	if err != nil {
		t.Errorf("second DeleteFile should not error (idempotent), got: %v", err)
	}
}

func TestSaveAndLoadIndex(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// Save some files
	meta1, _ := storage.SaveFile(uuid.New(), "file1.txt", "text/plain", bytes.NewReader([]byte("content1")))
	meta2, _ := storage.SaveFile(uuid.New(), "file2.txt", "text/plain", bytes.NewReader([]byte("content2")))

	// Create new storage instance and load index
	storage2 := NewLocalStorage(tmpDir)
	err := storage2.LoadIndex()
	if err != nil {
		t.Fatalf("LoadIndex failed: %v", err)
	}

	// Verify index was loaded
	if len(storage2.index) != 2 {
		t.Errorf("expected 2 files in loaded index, got %d", len(storage2.index))
	}

	if _, exists := storage2.index[meta1.Hash]; !exists {
		t.Error("file1 should be in loaded index")
	}
	if _, exists := storage2.index[meta2.Hash]; !exists {
		t.Error("file2 should be in loaded index")
	}
}

func TestLoadIndex_EmptyStorage(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	// LoadIndex on empty storage should not error
	err := storage.LoadIndex()
	if err != nil {
		t.Errorf("LoadIndex on empty storage should not error, got: %v", err)
	}
}

func TestSaveFile_DuplicateContent(t *testing.T) {
	storage, tmpDir := setupTestStorage(t)
	defer cleanupTestStorage(tmpDir)

	data := []byte("Same content")

	// Save same content with different file IDs
	meta1, _ := storage.SaveFile(uuid.New(), "file1.txt", "text/plain", bytes.NewReader(data))
	meta2, _ := storage.SaveFile(uuid.New(), "file2.txt", "text/plain", bytes.NewReader(data))

	// Same content = same hash
	if meta1.Hash != meta2.Hash {
		t.Error("same content should produce same hash")
	}

	// But index should have the latest metadata
	indexMeta := storage.index[meta1.Hash]
	if indexMeta.Filename != "file2.txt" {
		t.Errorf("index should have latest filename, got %s", indexMeta.Filename)
	}
}

func TestFilePath(t *testing.T) {
	storage := NewLocalStorage("/base/path")

	hash := "abcdef1234567890"
	expected := filepath.Join("/base/path", "data", "ab", "abcdef1234567890.dat")
	actual := storage.filePath(hash)

	if actual != expected {
		t.Errorf("expected path %s, got %s", expected, actual)
	}
}

func TestIndexPath(t *testing.T) {
	storage := NewLocalStorage("/base/path")

	expected := filepath.Join("/base/path", "index.json")
	actual := storage.indexPath()

	if actual != expected {
		t.Errorf("expected path %s, got %s", expected, actual)
	}
}
