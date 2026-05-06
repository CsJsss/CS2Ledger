import { useQuery } from "@tanstack/react-query";
import { GetAccounts } from "../lib/wails";

export function useAccounts() {
  return useQuery({
    queryKey: ["accounts"],
    queryFn: GetAccounts,
    staleTime: Infinity,
  });
}
