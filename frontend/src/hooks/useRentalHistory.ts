import { useQuery } from "@tanstack/react-query";
import { GetRentalHistory } from "../lib/wails";

export function useRentalHistory(accountId: number | null, assetId: string | null) {
  return useQuery({
    queryKey: ["rentalHistory", accountId, assetId],
    queryFn: () => GetRentalHistory(accountId!, assetId!),
    staleTime: 5 * 60 * 1000,
    enabled: !!accountId && !!assetId,
  });
}
