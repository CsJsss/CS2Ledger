import { useQuery } from '@tanstack/react-query';
import { GetInventory } from '../lib/wails';
import { useAccounts } from './useAccounts';

interface UseInventoryOpts {
  status?: string;
  weaponType?: string;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortDir?: string;
}

export function useInventory(selectedAccountId: number | null, opts?: UseInventoryOpts) {
  const { data: accounts = [] } = useAccounts();
  const status = opts?.status ?? '';
  const weaponType = opts?.weaponType ?? '';
  const page = opts?.page ?? 1;
  const pageSize = opts?.pageSize ?? 50;
  const sortBy = opts?.sortBy ?? 'itemName';
  const sortDir = opts?.sortDir ?? 'asc';

  return useQuery({
    queryKey: [
      'inventory',
      selectedAccountId ?? 0,
      status,
      weaponType,
      page,
      pageSize,
      sortBy,
      sortDir,
    ],
    queryFn: () => {
      // selectedAccountId=null → pass 0 to backend (all-accounts query)
      const accountId = selectedAccountId ?? 0;
      return GetInventory(accountId, status, weaponType, page, pageSize, sortBy, sortDir);
    },
    staleTime: 5 * 60 * 1000,
    enabled: selectedAccountId !== null || accounts.length > 0,
  });
}
