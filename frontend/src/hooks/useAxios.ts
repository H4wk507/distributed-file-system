import axios from "axios";

const API_BASE_URL = import.meta.env.VITE_API_URL as string;

export function useAxios() {
  return axios.create({
    baseURL: API_BASE_URL,
    headers: {
      "Content-Type": "application/json",
    },
    withCredentials: true,
  });
}
