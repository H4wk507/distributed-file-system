package handlers

import (
	"dfs-backend/internal/database"
	"dfs-backend/internal/services"
	"dfs-backend/utils/response"
	"net/http"

	"github.com/google/uuid"
)

type FileHandler struct {
	service *services.FileService
}

func NewFileHandler(db *database.DB) *FileHandler {
	return &FileHandler{
		service: services.NewFileService(db),
	}
}

func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid file ID format")
		return
	}

	file, err := h.service.GetByID(id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "File not found")
		return
	}

	response.Success(w, file, "")
}
