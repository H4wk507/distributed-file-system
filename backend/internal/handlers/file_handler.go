package handlers

import (
	"dfs-backend/dfs/client"
	"dfs-backend/dfs/common"
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/middleware"
	"dfs-backend/internal/models"
	"dfs-backend/internal/services"
	"dfs-backend/utils/response"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type UploadSession struct {
	ID            string
	FileID        uuid.UUID
	UserID        uuid.UUID
	FileName      string
	FileSize      int64
	TotalChunks   int
	ContentType   string
	StreamSession *client.StreamUploadSession
	NextChunk     int
	CreatedAt     time.Time
	mu            sync.Mutex
}

type FileHandler struct {
	service  *services.FileService
	client   *client.MasterClient
	sessions map[string]*UploadSession
	sessMu   sync.RWMutex
}

func NewFileHandler(db *database.DB, c *client.MasterClient) *FileHandler {
	h := &FileHandler{
		service:  services.NewFileService(db),
		client:   c,
		sessions: make(map[string]*UploadSession),
	}
	go h.cleanupStaleSessions()
	return h
}

func (h *FileHandler) cleanupStaleSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		h.sessMu.Lock()
		for id, sess := range h.sessions {
			if time.Since(sess.CreatedAt) > 24*time.Hour {
				if sess.StreamSession != nil {
					sess.StreamSession.Close()
				}
				delete(h.sessions, id)
			}
		}
		h.sessMu.Unlock()
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
			ID:            f.ID,
			Filename:      f.Name,
			Size:          f.Size,
			ContentType:   f.ContentType,
			Hash:          f.Hash,
			OwnerID:       f.OwnerID,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
			ReplicasCount: f.ReplicasCount,
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

func (h *FileHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	nodes, err := h.client.GetStorageNodes()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get storage nodes: %v", err))
		return
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    nodes,
	})
}

func (h *FileHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	nodes, err := h.client.GetStorageNodes()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get storage nodes: %v", err))
		return
	}

	nodeId := r.PathValue("nodeID")
	if nodeId == "" {
		response.Error(w, http.StatusBadRequest, "'nodeID' not present in path")
		return
	}

	var foundNode *common.NodeInfo
	for _, node := range nodes {
		if node.ID.String() == nodeId {
			foundNode = node
			break
		}
	}

	if foundNode == nil {
		response.Error(w, http.StatusNotFound, "Node not found")
		return
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    foundNode,
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

func (h *FileHandler) InitChunkedUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req dto.ChunkedUploadInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FileName == "" || req.FileSize <= 0 || req.TotalChunks <= 0 {
		response.Error(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileID := uuid.New()

	// Initialize streaming session - opens TCP connections to storage nodes
	streamSession, err := h.client.InitStreamUpload(fileID, req.FileName, contentType, req.FileSize)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize upload: %v", err))
		return
	}

	sessionID := streamSession.SessionID

	session := &UploadSession{
		ID:            sessionID,
		FileID:        fileID,
		UserID:        claims.UserID,
		FileName:      req.FileName,
		FileSize:      req.FileSize,
		TotalChunks:   req.TotalChunks,
		ContentType:   contentType,
		StreamSession: streamSession,
		NextChunk:     0,
		CreatedAt:     time.Now(),
	}

	h.sessMu.Lock()
	h.sessions[sessionID] = session
	h.sessMu.Unlock()

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    dto.ChunkedUploadInitResponse{SessionID: sessionID},
	})
}

func (h *FileHandler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Parse with larger limit for bigger chunks (frontend uses 50MB)
	if err := r.ParseMultipartForm(64 * 1024 * 1024); err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse form: %v", err))
		return
	}

	sessionID := r.FormValue("sessionId")
	chunkIndexStr := r.FormValue("chunkIndex")

	if sessionID == "" || chunkIndexStr == "" {
		response.Error(w, http.StatusBadRequest, "Missing sessionId or chunkIndex")
		return
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid chunkIndex")
		return
	}

	h.sessMu.RLock()
	session, exists := h.sessions[sessionID]
	h.sessMu.RUnlock()

	if !exists {
		response.Error(w, http.StatusNotFound, "Upload session not found")
		return
	}

	if session.UserID != claims.UserID {
		response.Error(w, http.StatusForbidden, "Not authorized for this upload session")
		return
	}

	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		response.Error(w, http.StatusBadRequest, "Invalid chunk index")
		return
	}

	chunkFile, _, err := r.FormFile("chunk")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Missing chunk file")
		return
	}
	defer chunkFile.Close()

	session.mu.Lock()
	defer session.mu.Unlock()

	// Verify sequential order - chunks must arrive in order for streaming
	if chunkIndex != session.NextChunk {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Out of order chunk: expected %d, got %d", session.NextChunk, chunkIndex))
		return
	}

	// Read chunk data and stream directly to storage nodes
	data, err := io.ReadAll(chunkFile)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to read chunk")
		return
	}

	if err := session.StreamSession.WriteChunk(data); err != nil {
		session.StreamSession.Close()
		h.sessMu.Lock()
		delete(h.sessions, sessionID)
		h.sessMu.Unlock()
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stream chunk: %v", err))
		return
	}

	session.NextChunk++

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"chunkIndex":     chunkIndex,
			"receivedChunks": session.NextChunk,
			"totalChunks":    session.TotalChunks,
		},
	})
}

func (h *FileHandler) FinalizeChunkedUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req dto.ChunkedUploadFinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.sessMu.Lock()
	session, exists := h.sessions[req.SessionID]
	if exists {
		delete(h.sessions, req.SessionID)
	}
	h.sessMu.Unlock()

	if !exists {
		response.Error(w, http.StatusNotFound, "Upload session not found")
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.UserID != claims.UserID {
		session.StreamSession.Close()
		response.Error(w, http.StatusForbidden, "Not authorized for this upload session")
		return
	}

	if session.NextChunk != session.TotalChunks {
		session.StreamSession.Close()
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Not all chunks received: got %d/%d", session.NextChunk, session.TotalChunks))
		return
	}

	result, err := session.StreamSession.Finalize()
	if err != nil {
		session.StreamSession.Close()
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to finalize upload: %v", err))
		return
	}

	fileModel := &models.File{
		ID:          session.FileID,
		Name:        session.FileName,
		Size:        session.FileSize,
		Hash:        result.Hash,
		ContentType: session.ContentType,
		OwnerID:     session.UserID,
	}

	if err := h.service.CreateFile(fileModel); err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save file metadata: %v", err))
		return
	}

	if err := h.service.CreateReplicas(session.FileID, result.NodeIDs); err != nil {
		fmt.Printf("Warning: Failed to save replica info: %v\n", err)
	}

	response.JSON(w, http.StatusAccepted, response.SuccessResponse{
		Success: true,
		Data: dto.ChunkedUploadFinalizeResponse{
			ID:       session.FileID,
			Filename: session.FileName,
			Size:     session.FileSize,
		},
		Message: "File uploaded successfully",
	})
}
