import { useQuery } from '@tanstack/react-query';
import { GetUnmatchedSells } from '../lib/wails';
import { useAccounts } from './useAccounts';

export function useUnmatchedSells(selectedAccountId: number | null) {
  const { data: accounts = [] } = useAccounts();
  return useQuery({
    queryKey: ['unmatchedSells', selectedAccountId ?? 'all'],
    queryFn: async () => {
      if (selectedAccountId != null) {
        return GetUnmatchedSells(selectedAccountId);
      }
      if (accounts.length === 0) return [];
      const results = await Promise.all(accounts.map((a) => GetUnmatchedSells(a.ID)));
      return results.flat();
    },
    staleTime: 2 * 60 * 1000,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
