import { useQuery } from '@tanstack/react-query';
import { GetDailySells } from '../lib/wails';
import { useAccounts } from './useAccounts';

const STALE_TIME_MS = 2 * 60 * 1000;
const DEFAULT_PAGE_SIZE = 30;

export function useDailySells(
  selectedAccountId: number | null,
  year: number,
  month: number,
  page: number,
  pageSize: number = DEFAULT_PAGE_SIZE,
) {
  const { data: accounts = [] } = useAccounts();

  return useQuery({
    queryKey: ['dailySells', selectedAccountId ?? 0, year, month, page, pageSize],
    queryFn: () => {
      const accountId = selectedAccountId ?? 0;
      return GetDailySells(accountId, year, month, page, pageSize);
    },
    staleTime: STALE_TIME_MS,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
