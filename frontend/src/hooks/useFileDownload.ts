import type { CanceledError } from "axios";
import { useCallback, useRef, useState } from "react";
import { useAxios } from "./useAxios";

interface DownloadItem {
  id: string;
  fileName: string;
  progress: number; // 0-100
  status: "pending" | "downloading" | "completed" | "error" | "cancelled";
  error?: string;
}

interface UseFileDownloadReturn {
  downloads: DownloadItem[];
  downloadFile: (fileId: string, fileName?: string) => Promise<void>;
  cancelDownload: (fileId: string) => void;
  clearDownload: (fileId: string) => void;
  clearAll: () => void;
}

export function useFileDownload(): UseFileDownloadReturn {
  const [downloads, setDownloads] = useState<DownloadItem[]>([]);
  const abortControllers = useRef<Map<string, AbortController>>(new Map());

  const axios = useAxios();

  const updateDownload = useCallback(
    (id: string, updates: Partial<DownloadItem>) => {
      setDownloads((prev) =>
        prev.map((d) => (d.id === id ? { ...d, ...updates } : d)),
      );
    },
    [],
  );

  const downloadFile = useCallback(
    async (fileId: string, fileName?: string) => {
      if (
        downloads.some((d) => d.id === fileId && d.status === "downloading")
      ) {
        return;
      }

      const controller = new AbortController();
      abortControllers.current.set(fileId, controller);

      setDownloads((prev) => {
        const existing = prev.find((d) => d.id === fileId);
        if (existing) {
          return prev.map((d) =>
            d.id === fileId
              ? {
                  ...d,
                  progress: 0,
                  status: "pending" as const,
                  error: undefined,
                }
              : d,
          );
        }
        return [
          ...prev,
          {
            id: fileId,
            fileName: fileName || fileId,
            progress: 0,
            status: "pending",
          },
        ];
      });

      try {
        const response = await axios.get<Blob>(`files/${fileId}`, {
          responseType: "blob",
          signal: controller.signal,
          onDownloadProgress: (progressEvent) => {
            if (progressEvent.total) {
              const progress = Math.round(
                (progressEvent.loaded / progressEvent.total) * 100,
              );
              updateDownload(fileId, { progress, status: "downloading" });
            }
          },
        });

        const disposition = response.headers["content-disposition"] as
          | string
          | undefined;
        const extractedName = disposition?.match(/filename="(.+)"/)?.[1];
        const finalFileName = fileName || extractedName || fileId;

        const blob = response.data;

        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = finalFileName;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        updateDownload(fileId, {
          fileName: finalFileName,
          progress: 100,
          status: "completed",
        });
      } catch (err) {
        const isCancelled =
          err instanceof Error &&
          (err.name === "CanceledError" ||
            (err as CanceledError<unknown>).code === "ERR_CANCELED");
        if (isCancelled) {
          updateDownload(fileId, { status: "cancelled" });
        } else {
          updateDownload(fileId, {
            status: "error",
            error: err instanceof Error ? err.message : "Download failed",
          });
        }
      } finally {
        abortControllers.current.delete(fileId);
      }
    },
    [axios, downloads, updateDownload],
  );

  const cancelDownload = useCallback((fileId: string) => {
    const controller = abortControllers.current.get(fileId);
    if (controller) {
      controller.abort();
    }
  }, []);

  const clearDownload = useCallback(
    (fileId: string) => {
      cancelDownload(fileId);
      setDownloads((prev) => prev.filter((d) => d.id !== fileId));
    },
    [cancelDownload],
  );

  const clearAll = useCallback(() => {
    abortControllers.current.forEach((controller) => controller.abort());
    abortControllers.current.clear();
    setDownloads([]);
  }, []);

  return {
    downloads,
    downloadFile,
    cancelDownload,
    clearDownload,
    clearAll,
  };
}
