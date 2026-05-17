import { useQuery } from "@tanstack/react-query";
import { GetBillRecords } from "../lib/wails";
import type { model } from "../lib/wails";

export function useBillRecords(accountId: number | null) {
  return useQuery({
    queryKey: ["billRecords", accountId ?? "all"],
    queryFn: () => GetBillRecords(accountId ?? 0),
    staleTime: 2 * 60 * 1000,
    enabled: accountId !== undefined,
  });
}
