import type { ApiResponse, FileInfo, FileListResponse } from "@/api/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAxios } from "./useAxios";

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
