import { useQuery } from "@tanstack/react-query";
import { GetCompletedTrades } from "../lib/wails";
import { useAccounts } from "./useAccounts";

export function useCompletedTrades(selectedAccountId: number | null) {
  const { data: accounts = [] } = useAccounts();
  return useQuery({
    queryKey: ["completedTrades", selectedAccountId ?? "all"],
    queryFn: async () => {
      if (selectedAccountId != null) {
        return GetCompletedTrades(selectedAccountId);
      }
      if (accounts.length === 0) return [];
      const results = await Promise.all(accounts.map((a) => GetCompletedTrades(a.ID)));
      return results.flat();
    },
    staleTime: 2 * 60 * 1000,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
