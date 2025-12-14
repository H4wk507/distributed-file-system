package handlers

import (
	"dfs-backend/dfs/client"
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/middleware"
	"dfs-backend/utils/response"
	"net/http"
	"time"
)

type MetricsHandler struct {
	db        *database.DB
	client    *client.MasterClient
	startedAt time.Time
}

func NewMetricsHandler(db *database.DB, c *client.MasterClient, startedAt time.Time) *MetricsHandler {
	return &MetricsHandler{db: db, client: c, startedAt: startedAt}
}

// SystemMetrics returns global system metrics for monitoring.
func (h *MetricsHandler) SystemMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var totalFiles int
	if err := h.db.Get(&totalFiles, `SELECT COUNT(*) FROM files`); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to compute total files")
		return
	}

	var totalSize int64
	if err := h.db.Get(&totalSize, `SELECT COALESCE(SUM(size), 0) FROM files`); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to compute total size")
		return
	}

	var totalReplicas int
	if err := h.db.Get(&totalReplicas, `SELECT COUNT(*) FROM replicas`); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to compute total replicas")
		return
	}

	activeNodes := 0
	if nodes, err := h.client.GetStorageNodes(); err == nil {
		activeNodes = len(nodes)
	}

	uptime := time.Since(h.startedAt).Seconds()

	resp := dto.SystemMetricsResponse{
		TotalFiles:    totalFiles,
		TotalSize:     totalSize,
		ActiveNodes:   activeNodes,
		TotalReplicas: totalReplicas,
		UptimeSeconds: int64(uptime),
	}

	response.JSON(w, http.StatusOK, response.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}
