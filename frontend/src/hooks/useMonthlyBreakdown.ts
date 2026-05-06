import { useQuery } from "@tanstack/react-query";
import { GetMonthlyBreakdown } from "../lib/wails";

export function useMonthlyBreakdown(accountId: number | null, year: number) {
  return useQuery({
    queryKey: ["monthlyBreakdown", accountId, year],
    queryFn: () => GetMonthlyBreakdown(accountId!, year),
    staleTime: 2 * 60 * 1000,
    enabled: !!accountId,
  });
}
