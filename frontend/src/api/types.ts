export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
}

export interface FileInfo {
  id: string;
  filename: string;
  size: number;
  content_type: string;
  hash: string;
  owner_id: string;
  replicas?: string[];
  created_at: string;
  updated_at: string;
  replicas_count: number;
}

export interface FileListResponse {
  files: FileInfo[];
  total: number;
  page: number;
  per_page: number;
}

export interface UploadResponse {
  id: string;
  filename: string;
  size: number;
}

export interface ChunkedUploadInitResponse {
  sessionId: string;
}

export interface ChunkedUploadChunkResponse {
  chunkIndex: number;
  receivedChunks: number;
  totalChunks: number;
}

export interface NodeInfo {
  id: string;
  ip: string;
  port: number;
  role: "master" | "storage";
  status: "online" | "offline" | "unknown";
  storage_used: number;
  storage_total: number;
  files_count: number;
  last_heartbeat: string;
}

export interface SystemMetrics {
  total_files: number;
  total_size: number;
  active_nodes: number;
  total_replicas: number;
  uptime_seconds: number;
}
