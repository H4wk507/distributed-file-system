import type {
  ApiResponse,
  ChunkedUploadChunkResponse,
  ChunkedUploadInitResponse,
  UploadResponse,
} from "@/api/types";
import { useQueryClient } from "@tanstack/react-query";
import type { CanceledError } from "axios";
import { useCallback, useRef, useState } from "react";
import { useAxios } from "./useAxios";

const CHUNK_SIZE = 50 * 1024 * 1024;

export interface UploadItem {
  id: string;
  fileName: string;
  fileSize: number;
  progress: number; // 0-100
  status: "pending" | "uploading" | "completed" | "error" | "cancelled";
  error?: string;
  response?: UploadResponse;
  uploadedChunks?: number;
  totalChunks?: number;
}

interface UseFileUploadReturn {
  uploads: UploadItem[];
  uploadFile: (file: File) => Promise<void>;
  cancelUpload: (uploadId: string) => void;
  clearUpload: (uploadId: string) => void;
  clearAll: () => void;
  clearCompleted: () => void;
}

export function useFileUpload(): UseFileUploadReturn {
  const [uploads, setUploads] = useState<UploadItem[]>([]);
  const abortControllers = useRef<Map<string, AbortController>>(new Map());
  const queryClient = useQueryClient();

  const axios = useAxios();

  const updateUpload = useCallback(
    (id: string, updates: Partial<UploadItem>) => {
      setUploads((prev) =>
        prev.map((u) => (u.id === id ? { ...u, ...updates } : u)),
      );
    },
    [],
  );

  const uploadFile = useCallback(
    async (file: File) => {
      const uploadId = crypto.randomUUID();
      const controller = new AbortController();
      abortControllers.current.set(uploadId, controller);

      const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

      setUploads((prev) => [
        ...prev,
        {
          id: uploadId,
          fileName: file.name,
          fileSize: file.size,
          progress: 0,
          status: "pending",
          uploadedChunks: 0,
          totalChunks,
        },
      ]);

      try {
        // Step 1: Initialize chunked upload session
        const initResponse = await axios.post<
          ApiResponse<ChunkedUploadInitResponse>
        >(
          "/files/upload/init",
          {
            fileName: file.name,
            fileSize: file.size,
            totalChunks,
            contentType: file.type || "application/octet-stream",
          },
          { signal: controller.signal },
        );

        const sessionId = initResponse.data.data?.sessionId;
        if (!sessionId) {
          throw new Error("Failed to initialize upload session");
        }

        updateUpload(uploadId, { status: "uploading" });

        for (let chunkIndex = 0; chunkIndex < totalChunks; chunkIndex++) {
          if (controller.signal.aborted) {
            throw new Error("Upload cancelled");
          }

          const start = chunkIndex * CHUNK_SIZE;
          const end = Math.min(start + CHUNK_SIZE, file.size);
          const chunk = file.slice(start, end);

          const formData = new FormData();
          formData.append("chunk", chunk);
          formData.append("chunkIndex", String(chunkIndex));
          formData.append("sessionId", sessionId);

          await axios.post<ApiResponse<ChunkedUploadChunkResponse>>(
            "/files/upload/chunk",
            formData,
            {
              headers: { "Content-Type": "multipart/form-data" },
              signal: controller.signal,
            },
          );

          const progress = Math.round(((chunkIndex + 1) / totalChunks) * 100);
          updateUpload(uploadId, {
            progress,
            uploadedChunks: chunkIndex + 1,
          });
        }

        const finalizeResponse = await axios.post<ApiResponse<UploadResponse>>(
          "/files/upload/finalize",
          { sessionId },
          { signal: controller.signal },
        );

        updateUpload(uploadId, {
          progress: 100,
          status: "completed",
          response: finalizeResponse.data.data,
        });

        queryClient.invalidateQueries({ queryKey: ["files"] });
        queryClient.invalidateQueries({ queryKey: ["metrics"] });
      } catch (err) {
        const isCancelled =
          err instanceof Error &&
          (err.name === "CanceledError" ||
            (err as CanceledError<unknown>).code === "ERR_CANCELED" ||
            err.message === "Upload cancelled");

        if (isCancelled) {
          updateUpload(uploadId, { status: "cancelled" });
        } else {
          updateUpload(uploadId, {
            status: "error",
            error: err instanceof Error ? err.message : "Upload failed",
          });
        }
      } finally {
        abortControllers.current.delete(uploadId);
      }
    },
    [axios, queryClient, updateUpload],
  );

  const cancelUpload = useCallback((uploadId: string) => {
    const controller = abortControllers.current.get(uploadId);
    if (controller) {
      controller.abort();
    }
  }, []);

  const clearUpload = useCallback(
    (uploadId: string) => {
      cancelUpload(uploadId);
      setUploads((prev) => prev.filter((u) => u.id !== uploadId));
    },
    [cancelUpload],
  );

  const clearAll = useCallback(() => {
    abortControllers.current.forEach((controller) => controller.abort());
    abortControllers.current.clear();
    setUploads([]);
  }, []);

  const clearCompleted = useCallback(() => {
    setUploads((prev) =>
      prev.filter(
        (u) =>
          u.status !== "completed" &&
          u.status !== "error" &&
          u.status !== "cancelled",
      ),
    );
  }, []);

  return {
    uploads,
    uploadFile,
    cancelUpload,
    clearUpload,
    clearAll,
    clearCompleted,
  };
}
