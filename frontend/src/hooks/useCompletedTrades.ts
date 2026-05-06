import { useQuery } from "@tanstack/react-query";
import { GetCompletedTrades } from "../lib/wails";

export function useCompletedTrades(accountId: number | null) {
  return useQuery({
    queryKey: ["completedTrades", accountId],
    queryFn: () => GetCompletedTrades(accountId!),
    staleTime: 2 * 60 * 1000,
    enabled: !!accountId,
  });
}
