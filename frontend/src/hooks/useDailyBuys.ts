import { useQuery } from '@tanstack/react-query';
import { GetDailyBuys } from '../lib/wails';
import { useAccounts } from './useAccounts';

const STALE_TIME_MS = 2 * 60 * 1000;

export function useDailyBuys(selectedAccountId: number | null) {
  const { data: accounts = [] } = useAccounts();

  return useQuery({
    queryKey: ['dailyBuys', selectedAccountId ?? 0],
    queryFn: () => GetDailyBuys(selectedAccountId ?? 0),
    staleTime: STALE_TIME_MS,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
