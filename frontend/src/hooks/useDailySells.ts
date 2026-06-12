import { useQuery } from '@tanstack/react-query';
import { GetDailySells } from '../lib/wails';
import { useAccounts } from './useAccounts';

const STALE_TIME_MS = 2 * 60 * 1000;

export function useDailySells(
  selectedAccountId: number | null,
  year: number,
  month: number,
) {
  const { data: accounts = [] } = useAccounts();

  return useQuery({
    queryKey: ['dailySells', selectedAccountId ?? 0, year, month],
    queryFn: () => GetDailySells(selectedAccountId ?? 0, year, month),
    staleTime: STALE_TIME_MS,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
