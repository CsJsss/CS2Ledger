import { useQuery } from "@tanstack/react-query";
import { GetInventory } from "../lib/wails";
import { useAccounts } from "./useAccounts";

export function useInventory(selectedAccountId: number | null, status?: string) {
  const { data: accounts = [] } = useAccounts();
  return useQuery({
    queryKey: ["inventory", selectedAccountId ?? "all", status],
    queryFn: async () => {
      if (selectedAccountId != null) {
        return GetInventory(selectedAccountId, status ?? "");
      }
      if (accounts.length === 0) return [];
      const results = await Promise.all(accounts.map((a) => GetInventory(a.ID, status ?? "")));
      return results.flat();
    },
    staleTime: 5 * 60 * 1000,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
