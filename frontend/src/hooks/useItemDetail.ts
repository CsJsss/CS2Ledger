import { useQuery } from "@tanstack/react-query";
import { GetItemDetail } from "../lib/wails";

export function useItemDetail(accountId: number | null, assetId: string | null) {
  return useQuery({
    queryKey: ["itemDetail", accountId, assetId],
    queryFn: () => GetItemDetail(accountId!, assetId!),
    staleTime: 5 * 60 * 1000,
    enabled: !!accountId && !!assetId,
  });
}
