import { useQuery } from "@tanstack/react-query";
import { GetCompletedTradesSummary } from "../lib/wails";

export function useCompletedTradesSummary(accountId: number | null) {
  return useQuery({
    queryKey: ["completedTradesSummary", accountId],
    queryFn: () => GetCompletedTradesSummary(accountId!),
    staleTime: 2 * 60 * 1000,
    enabled: !!accountId,
  });
}
