import type { ApiResponse, NodeInfo } from "@/api/types";
import { useQuery } from "@tanstack/react-query";
import { useAxios } from "./useAxios";

export function useNodes() {
  const axios = useAxios();

  return useQuery({
    queryKey: ["nodes"],
    queryFn: async () => {
      const { data } = await axios.get<ApiResponse<NodeInfo[]>>("/nodes");
      return data.data;
    },
    refetchInterval: 10000,
  });
}

export function useNode(nodeId: string) {
  const axios = useAxios();

  return useQuery({
    queryKey: ["node", nodeId],
    queryFn: async () => {
      const { data } = await axios.get<ApiResponse<NodeInfo>>(
        `/nodes/${nodeId}`,
      );
      return data.data;
    },
    enabled: !!nodeId,
    refetchInterval: 5000,
  });
}
