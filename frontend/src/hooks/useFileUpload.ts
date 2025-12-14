import type { ApiResponse, UploadResponse } from "@/api/types";
import { useQueryClient } from "@tanstack/react-query";
import type { CanceledError } from "axios";
import { useCallback, useRef, useState } from "react";
import { useAxios } from "./useAxios";

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

      setUploads((prev) => [
        ...prev,
        {
          id: uploadId,
          fileName: file.name,
          fileSize: file.size,
          progress: 0,
          status: "pending",
        },
      ]);

      try {
        const formData = new FormData();
        formData.append("file", file);

        const response = await axios.post<ApiResponse<UploadResponse>>(
          "/files/upload/",
          formData,
          {
            headers: {
              "Content-Type": "multipart/form-data",
            },
            signal: controller.signal,
            onUploadProgress: (progressEvent) => {
              if (progressEvent.total) {
                const progress = Math.round(
                  (progressEvent.loaded / progressEvent.total) * 100,
                );
                updateUpload(uploadId, { progress, status: "uploading" });
              }
            },
          },
        );

        updateUpload(uploadId, {
          progress: 100,
          status: "completed",
          response: response.data.data,
        });

        queryClient.invalidateQueries({ queryKey: ["files"] });
        queryClient.invalidateQueries({ queryKey: ["metrics"] });
      } catch (err) {
        const isCancelled =
          err instanceof Error &&
          (err.name === "CanceledError" ||
            (err as CanceledError<unknown>).code === "ERR_CANCELED");

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
