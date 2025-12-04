import type { ApiResponse, SystemMetrics } from "@/api/types";
import { useQuery } from "@tanstack/react-query";
import { useAxios } from "./useAxios";

// TODO: implement those endpoints
export function useMetrics() {
  const axios = useAxios();

  return useQuery({
    queryKey: ["metrics"],
    queryFn: async () => {
      const { data } =
        await axios.get<ApiResponse<SystemMetrics>>("/metrics/system");
      return data.data;
    },
    refetchInterval: 30_000,
  });
}
