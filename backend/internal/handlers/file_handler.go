package handlers

import (
	"dfs-backend/dfs/client"
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/middleware"
	"dfs-backend/internal/models"
	"dfs-backend/internal/services"
	"dfs-backend/utils/response"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type FileHandler struct {
	service *services.FileService
	client  *client.MasterClient
}

func NewFileHandler(db *database.DB, c *client.MasterClient) *FileHandler {
	return &FileHandler{
		service: services.NewFileService(db),
		client:  c,
	}
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	page := 1
	perPage := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	files, total, err := h.service.ListFilesPaginated(claims.UserID, page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list files: %v", err))
		return
	}

	fileItems := make([]dto.FileItem, len(files))
	for i, f := range files {
		fileItems[i] = dto.FileItem{
			ID:          f.ID,
			Filename:    f.Name,
			Size:        f.Size,
			ContentType: f.ContentType,
			Hash:        f.Hash,
			OwnerID:     f.OwnerID,
			CreatedAt:   f.CreatedAt,
			UpdatedAt:   f.UpdatedAt,
		}
	}

	resp := dto.FileListResponse{
		Files:   fileItems,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

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

	fileID := uuid.New()
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	hash, err := h.client.UploadFileStream(fileID, header.Filename, contentType, header.Size, file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to storage: %v", err))
		return
	}

	fileModel := &models.File{
		ID:          fileID,
		Name:        header.Filename,
		Size:        header.Size,
		Hash:        hash,
		ContentType: contentType,
		OwnerID:     claims.UserID,
	}

	if err := h.service.CreateFile(fileModel); err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save file metadata: %v", err))
		return
	}

	resp := dto.FileUploadResponse{
		ID:       fileID,
		Filename: header.Filename,
		Size:     header.Size,
	}

	response.JSON(w, http.StatusAccepted, response.SuccessResponse{
		Success: true,
		Data:    resp,
		Message: "File uploaded successfully",
	})
}

func (h *FileHandler) GetFileMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	fileIDStr := r.PathValue("fileID")
	if fileIDStr == "" {
		response.Error(w, http.StatusBadRequest, "'fileID' not present in path")
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid fileID format")
		return
	}

	file, err := h.service.GetFileByID(fileID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "File not found")
		return
	}

	if file.OwnerID != claims.UserID && claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "Forbidden")
		return
	}

	metadata := dto.FileItem{
		ID:          file.ID,
		Filename:    file.Name,
		Size:        file.Size,
		ContentType: file.ContentType,
		Hash:        file.Hash,
		OwnerID:     file.OwnerID,
		CreatedAt:   file.CreatedAt,
		UpdatedAt:   file.UpdatedAt,
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    metadata,
	})
}

func (h *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	fileIDStr := r.PathValue("fileID")
	if fileIDStr == "" {
		response.Error(w, http.StatusBadRequest, "'fileID' not present in path")
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid fileID format")
		return
	}

	file, err := h.service.GetFileByID(fileID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "File not found")
		return
	}

	if file.OwnerID != claims.UserID && claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "Forbidden")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	if err := h.client.DownloadFileStream(file.ID, file.Hash, w); err != nil {
		fmt.Printf("Error streaming file %s: %v\n", file.ID, err)
		return
	}
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	fileIDStr := r.PathValue("fileID")
	if fileIDStr == "" {
		response.Error(w, http.StatusBadRequest, "'fileID' not present in path")
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid fileID format")
		return
	}

	file, err := h.service.GetFileByID(fileID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "File not found")
		return
	}

	if file.OwnerID != claims.UserID && claims.Role != "admin" {
		response.Error(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := h.service.DeleteFileByID(file.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete file from database")
		return
	}

	deleteResp, err := h.client.DeleteFile(file.ID, file.Hash)
	if err != nil {
		fmt.Printf("Warning: Failed to delete file from storage: %v\n", err)
	} else if !deleteResp.Success {
		fmt.Printf("Warning: Storage delete error: %s\n", deleteResp.Error)
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    file.ID,
		Message: "File deleted successfully",
	})
}
