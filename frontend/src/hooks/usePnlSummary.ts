import { useQuery } from "@tanstack/react-query";
import { GetPnlSummary } from "../lib/wails";

export function usePnlSummary(accountId: number | null) {
  return useQuery({
    queryKey: ["pnlSummary", accountId],
    queryFn: () => GetPnlSummary(accountId!),
    staleTime: 2 * 60 * 1000,
    enabled: !!accountId,
  });
}
