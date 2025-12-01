package handlers

import (
	"crypto/sha256"
	"dfs-backend/dfs/node"
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/models"
	"dfs-backend/internal/services"
	"dfs-backend/utils/response"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type FileHandler struct {
	service *services.FileService
	node    *node.Node
}

func NewFileHandler(db *database.DB, n *node.Node) *FileHandler {
	return &FileHandler{
		service: services.NewFileService(db),
		node:    n,
	}
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: support larger files, chunked upload protocol
	r.Body = http.MaxBytesReader(w, r.Body, 100*1024*1024) // 100MB limit

	if err := r.ParseMultipartForm(32 * 1024 * 1024); err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse multipart form: %v", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Failed to get file from form: %v", err))
		return
	}
	defer file.Close()

	// TODO: what about big files?
	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read file: %v", err))
		return
	}

	fileID := uuid.New()
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	contentHash := sha256.Sum256(data)
	contentHashStr := hex.EncodeToString(contentHash[:])

	fileModel := &models.File{
		ID:          fileID,
		Name:        header.Filename,
		Size:        header.Size,
		Hash:        contentHashStr,
		ContentType: contentType,
		OwnerID:     uuid.Nil, // TODO:
	}

	if err := h.service.CreateFile(fileModel); err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save file metadata: %v", err))
		return
	}

	h.node.StoreFile(fileID, header.Filename, contentType, data)

	// TODO: streaming response?
	resp := dto.FileUploadResponse{
		ID:       fileID,
		Filename: header.Filename,
		Size:     header.Size,
	}

	response.JSON(w, http.StatusAccepted, response.SuccessResponse{
		Success: true,
		Data:    resp,
		Message: "File upload initiated",
	})
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: move to unique identifier from DB, fileID, and also on hashring

	filename := r.PathValue("filename")
	if filename == "" {
		response.Error(w, http.StatusBadRequest, "'filename' not present in path")
		return
	}

	hash, err := h.service.GetFileHashByName(filename)
	if err != nil {
		response.Error(w, http.StatusNotFound, "File not found")
		return
	}

	fileResponse, err := h.node.RetrieveFile(filename, hash)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve file: %v", err))
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileResponse.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileResponse.Data)))

	w.WriteHeader(http.StatusOK)
	w.Write(fileResponse.Data)
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// TODO: move to unique identifier from DB, fileID, and also on hashring

	filename := r.PathValue("filename")
	if filename == "" {
		response.Error(w, http.StatusBadRequest, "'filename' not present in path")
		return
	}

	hash, err := h.service.GetFileHashByName(filename)
	if err != nil {
		response.Error(w, http.StatusNotFound, "File not found")
		return
	}

	if err := h.service.DeleteFileByName(filename); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete file from database")
		return
	}

	h.node.DeleteFile(filename, hash)

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    filename,
		Message: "File deleted successfully",
	})
}
