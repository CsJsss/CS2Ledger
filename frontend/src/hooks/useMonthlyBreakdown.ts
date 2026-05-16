import { useQuery } from "@tanstack/react-query";
import { GetMonthlyBreakdown } from "../lib/wails";
import { useAccounts } from "./useAccounts";

export function useMonthlyBreakdown(selectedAccountId: number | null, year: number) {
  const { data: accounts = [] } = useAccounts();
  return useQuery({
    queryKey: ["monthlyBreakdown", selectedAccountId ?? "all", year],
    queryFn: async () => {
      if (selectedAccountId != null) {
        return GetMonthlyBreakdown(selectedAccountId, year);
      }
      if (accounts.length === 0) return [];
      const results = await Promise.all(accounts.map((a) => GetMonthlyBreakdown(a.ID, year)));
      const monthMap = new Map<string, number>();
      for (const r of results) {
        for (const m of r) {
          monthMap.set(m.month, (monthMap.get(m.month) ?? 0) + m.netPl);
        }
      }
      return Array.from(monthMap, ([month, netPl]) => ({ month, netPl }));
    },
    staleTime: 2 * 60 * 1000,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
