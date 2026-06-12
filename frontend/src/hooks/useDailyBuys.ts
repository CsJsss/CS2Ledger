import { useQuery } from '@tanstack/react-query';
import { GetDailyBuys } from '../lib/wails';
import { useAccounts } from './useAccounts';

const STALE_TIME_MS = 2 * 60 * 1000;
const DEFAULT_PAGE_SIZE = 30;

export function useDailyBuys(
  selectedAccountId: number | null,
  page: number,
  pageSize: number = DEFAULT_PAGE_SIZE,
) {
  const { data: accounts = [] } = useAccounts();

  return useQuery({
    queryKey: ['dailyBuys', selectedAccountId ?? 0, page, pageSize],
    queryFn: () => {
      const accountId = selectedAccountId ?? 0;
      return GetDailyBuys(accountId, page, pageSize);
    },
    staleTime: STALE_TIME_MS,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
