import type {
  ApiResponse,
  FileInfo,
  FileListResponse,
  UploadResponse,
} from "@/api/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAxios } from "./useAxios";

// implement those endpoints
export function useFiles(page: number, perPage: number) {
  const axios = useAxios();

  return useQuery({
    queryKey: ["files", page, perPage],
    queryFn: async () => {
      const { data } = await axios.get<ApiResponse<FileListResponse>>(
        "/files",
        {
          params: { page, per_page: perPage },
        },
      );
      return data.data;
    },
  });
}

// TODO: implement those endpoints
export function useFile(fileId: string) {
  const axios = useAxios();

  return useQuery({
    queryKey: ["file", fileId],
    queryFn: async () => {
      const { data } = await axios.get<ApiResponse<FileInfo>>(
        `/files/${fileId}`,
      );
      return data.data;
    },
    enabled: !!fileId,
  });
}

export function useUploadFile() {
  const axios = useAxios();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData();
      formData.append("file", file);

      const { data } = await axios.post<ApiResponse<UploadResponse>>(
        "/files/upload",
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        },
      );
      return data.data;
    },
    onSuccess: (data) => {
      toast.success("Plik przesłany", {
        description: `${data?.filename} został przesłany pomyślnie`,
      });
      queryClient.invalidateQueries({ queryKey: ["files"] });
      queryClient.invalidateQueries({ queryKey: ["metrics"] });
    },
  });
}

// TODO: implement those endpoints
export function useDeleteFile() {
  const axios = useAxios();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (fileId: string) => {
      await axios.delete(`/files/${fileId}`);
    },
    onSuccess: () => {
      toast.success("Plik usunięty");
      queryClient.invalidateQueries({ queryKey: ["files"] });
      queryClient.invalidateQueries({ queryKey: ["metrics"] });
    },
  });
}

// TODO: implement those endpoints
export function useDownloadFile() {
  const axios = useAxios();

  return useMutation({
    mutationFn: async ({
      fileId,
      filename,
    }: {
      fileId: string;
      filename: string;
    }) => {
      const { data } = await axios.get(`/files/${fileId}/download`, {
        responseType: "blob",
      });

      const url = window.URL.createObjectURL(new Blob([data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", filename);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    },
  });
}
