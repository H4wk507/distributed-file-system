import type { ApiResponse, User } from "@/api/types";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { useAxios } from "./useAxios";

export function useUser() {
  const axios = useAxios();

  return useQuery<User | null>({
    queryKey: ["user"],
    queryFn: async () => {
      try {
        const resp = await axios.get<ApiResponse<User>>("/auth/me");
        return resp.data.data ?? null;
      } catch {
        return null;
      }
    },
    retry: false,
    staleTime: Infinity,
  });
}

export function useLogout() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const logout = () => {
    document.cookie = "token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
    queryClient.setQueryData(["user"], null);

    navigate("/login");
  };

  return logout;
}
