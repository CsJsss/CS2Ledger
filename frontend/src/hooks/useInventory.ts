import { useQuery } from "@tanstack/react-query";
import { GetInventory } from "../lib/wails";

export function useInventory(accountId: number | null, status?: string) {
  return useQuery({
    queryKey: ["inventory", accountId, status],
    queryFn: () => GetInventory(accountId!, status ?? ""),
    staleTime: 5 * 60 * 1000,
    enabled: !!accountId,
  });
}
