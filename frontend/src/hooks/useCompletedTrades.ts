import { useQuery } from "@tanstack/react-query";
import { GetCompletedTrades } from "../lib/wails";
import { useAccounts } from "./useAccounts";

interface UseCompletedTradesOpts {
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortDir?: string;
}

export function useCompletedTrades(selectedAccountId: number | null, opts?: UseCompletedTradesOpts) {
  const { data: accounts = [] } = useAccounts();
  const page = opts?.page ?? 1;
  const pageSize = opts?.pageSize ?? 50;
  const sortBy = opts?.sortBy ?? "itemName";
  const sortDir = opts?.sortDir ?? "asc";

  return useQuery({
    queryKey: ["completedTrades", selectedAccountId ?? 0, page, pageSize, sortBy, sortDir],
    queryFn: () => {
      // selectedAccountId=null → pass 0 to backend (all-accounts query)
      const accountId = selectedAccountId ?? 0;
      return GetCompletedTrades(accountId, page, pageSize, sortBy, sortDir);
    },
    staleTime: 2 * 60 * 1000,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
