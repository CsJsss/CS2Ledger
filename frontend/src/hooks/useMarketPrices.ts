import { useQuery } from "@tanstack/react-query";
import { GetMarketPrices } from "../lib/wails";
import { useAccounts } from "./useAccounts";
import { PLATFORM_CSQAQ } from "../lib/constants";

export function useMarketPrices() {
  const { data: accounts = [] } = useAccounts();
  const csqaqAccount = accounts.find((a) => a.platform === PLATFORM_CSQAQ);

  return useQuery({
    queryKey: ["marketPrices", csqaqAccount?.ID],
    queryFn: () => GetMarketPrices(),
    staleTime: 5 * 60 * 1000,
    enabled: !!csqaqAccount,
  });
}
