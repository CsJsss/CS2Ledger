import { useQuery } from '@tanstack/react-query';
import { GetDashboardSummary } from '../lib/wails';

export function useDashboard() {
  return useQuery({
    queryKey: ['dashboard'],
    queryFn: GetDashboardSummary,
    staleTime: 30 * 1000,
  });
}
