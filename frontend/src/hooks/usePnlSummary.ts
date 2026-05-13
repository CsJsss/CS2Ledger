import { useQuery } from "@tanstack/react-query";
import { GetPnlSummary } from "../lib/wails";
import { useAccounts } from "./useAccounts";
import type { pnl } from "../lib/wails";

export function usePnlSummary(selectedAccountId: number | null) {
  const { data: accounts = [] } = useAccounts();
  return useQuery({
    queryKey: ["pnlSummary", selectedAccountId ?? "all"],
    queryFn: async () => {
      if (selectedAccountId != null) {
        return GetPnlSummary(selectedAccountId);
      }
      const zero = { totalTrades: 0, totalGrossPl: 0, totalFee: 0, totalNetPl: 0 } satisfies pnl.PnlSummaryView;
      if (accounts.length === 0) return zero;
      const results = await Promise.all(accounts.map((a) => GetPnlSummary(a.ID)));
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
