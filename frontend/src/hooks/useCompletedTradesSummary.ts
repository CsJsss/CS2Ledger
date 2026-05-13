import { useQuery } from "@tanstack/react-query";
import { GetCompletedTradesSummary } from "../lib/wails";
import { useAccounts } from "./useAccounts";
import type { trade } from "../lib/wails";

export function useCompletedTradesSummary(selectedAccountId: number | null) {
  const { data: accounts = [] } = useAccounts();
  return useQuery({
    queryKey: ["completedTradesSummary", selectedAccountId ?? "all"],
    queryFn: async () => {
      if (selectedAccountId != null) {
        return GetCompletedTradesSummary(selectedAccountId);
      }
      const zero = { totalTrades: 0, totalGrossPl: 0, totalFee: 0, totalNetPl: 0 } satisfies trade.CompletedTradesSummary;
      if (accounts.length === 0) return zero;
      const results = await Promise.all(accounts.map((a) => GetCompletedTradesSummary(a.ID)));
      return results.reduce(
        (acc, r) => ({
          totalTrades: acc.totalTrades + r.totalTrades,
          totalGrossPl: acc.totalGrossPl + r.totalGrossPl,
          totalFee: acc.totalFee + r.totalFee,
          totalNetPl: acc.totalNetPl + r.totalNetPl,
        }),
        zero,
      );
    },
    staleTime: 2 * 60 * 1000,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
