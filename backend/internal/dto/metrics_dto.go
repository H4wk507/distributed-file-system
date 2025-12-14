package dto

type SystemMetricsResponse struct {
	TotalFiles    int   `json:"total_files"`
	TotalSize     int64 `json:"total_size"`
	ActiveNodes   int   `json:"active_nodes"`
	TotalReplicas int   `json:"total_replicas"`
	UptimeSeconds int64 `json:"uptime_seconds"`
}
